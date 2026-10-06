package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	log "github.com/go-pkgz/lgr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/vladikan/addrss-telegram/database"
	"github.com/vladikan/addrss-telegram/templates"
)

// Options holds all necessary settings for the app
type Options struct {
	Token          string
	APIToken       string
	Connection     string
	Debug          bool
	ReaderInterval int
	ReaderFeeds    int
	BotAdmin       int64
	HTTPPort       int
}

// Reply is a message to be sent to user/chat
type Reply struct {
	ChatID int64
	Text   string
}

var bot *tgbotapi.BotAPI
var db database.Database

// Start will call for bot instance and process update messages
func Start(options Options) {
	bt, err := tgbotapi.NewBotAPI(options.Token)
	if err != nil {
		log.Printf("PANIC Error while creating bot instance: %s", err)
	}

	bot = bt
	bot.Debug = options.Debug
	log.Printf("INFO Authorized on account %s", bot.Self.UserName)

	cfg := tgbotapi.NewUpdate(0)
	cfg.Timeout = 60

	// Hook for system terminate signal
	ctx, cancel := context.WithCancel(context.Background())
	go handleTerminate(cancel)

	templates.SetTemplateOutput()

	// Set db connection settings and use pool
	db, err = database.Open(ctx, options.Connection)
	if err != nil {
		log.Printf("PANIC Error while connecting to the database: %s", err)
	}
	defer db.Close()

	// Init messages channel
	replyQueue := make(chan Reply)
	go handleReply(replyQueue)
	defer close(replyQueue)

	// Start reader
	reader := &Reader{Interval: options.ReaderInterval, Feeds: options.ReaderFeeds, DB: db, Outbox: replyQueue}
	reader.Start()
	defer reader.Stop()

	// Read commands from users
	updates, _ := bot.GetUpdatesChan(cfg)
	go handleRequests(updates, replyQueue, &options)
	defer bot.StopReceivingUpdates()

	apiServer := startValidationAPIServer(&options)
	if apiServer != nil {
		defer apiServer.Close()
	}

	// Stop bot operations and close all connections
	<-ctx.Done()

	log.Print("INFO Stoping updates processing")
}

func startValidationAPIServer(options *Options) *http.Server {
	if options.APIToken == "" {
		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})
	registerValidationAPI(mux, options)

	addr := fmt.Sprintf(":%d", options.HTTPPort)
	srv := &http.Server{Addr: addr, Handler: mux}
	go func() {
		log.Printf("INFO Validation API listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("ERROR Validation API server error: %s", err)
		}
	}()
	return srv
}

func handleTerminate(cancel context.CancelFunc) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	log.Print("WARN System interrupt or terminate signal")
	cancel()
}

func handleRequests(updates tgbotapi.UpdatesChannel, replyQueue chan Reply, opt *Options) {
	log.Print("INFO Start updates processing")
	for update := range updates {
		msg := update.Message
		if msg == nil {
			continue
		}

		cmd := newCommand(msg, opt, replyQueue)
		replies := cmd.run()
		for _, reply := range replies {
			replyQueue <- reply
		}
	}

	log.Print("INFO Updates channel was closed")
}

func handleReply(queue chan Reply) {
	for msg := range queue {
		rsp := tgbotapi.NewMessage(msg.ChatID, msg.Text)
		rsp.ParseMode = "HTML"

		if _, err := bot.Send(rsp); err != nil {
			if isDeliveryForbidden(err) {
				if markErr := db.SetUserBlocked(msg.ChatID); markErr != nil {
					log.Printf("ERROR failed to mark user %d as blocked: %s", msg.ChatID, markErr)
				} else {
					log.Printf("WARN user %d blocked the bot (or account deactivated); delivery stopped", msg.ChatID)
				}
				continue
			}

			log.Printf("ERROR %T Problem while replying on %d chat: %s", err, msg.ChatID, err)
		}
	}

	log.Print("INFO Reply queue channel was closed")
}

// isDeliveryForbidden detects Telegram errors that mean we must stop messaging the user.
// Official Bot API returns HTTP 403 with descriptions such as:
//   - "Forbidden: bot was blocked by the user"
//   - "Forbidden: user is deactivated"
// There is no API to query block status ahead of time; the failed send is the signal.
// See https://core.telegram.org/bots/api#making-requests and Update.my_chat_member.
func isDeliveryForbidden(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "blocked by the user") ||
		strings.Contains(msg, "user is deactivated") ||
		strings.Contains(msg, "forbidden")
}
