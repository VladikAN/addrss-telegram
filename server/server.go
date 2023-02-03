package server

import (
	"context"
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
	Connection     string
	Debug          bool
	ReaderInterval int
	ReaderFeeds    int
	BotAdmin       int64
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

	// Stop bot operations and close all connections
	<-ctx.Done()

	log.Print("INFO Stoping updates processing")
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

		cmd := newCommand(msg, opt)
		txt := cmd.run()
		replyQueue <- Reply{ChatID: msg.Chat.ID, Text: txt}
	}

	log.Print("INFO Updates channel was closed")
}

func handleReply(queue chan Reply) {
	for msg := range queue {
		rsp := tgbotapi.NewMessage(msg.ChatID, msg.Text)
		rsp.ParseMode = "HTML"

		if _, err := bot.Send(rsp); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "forbidden") {
				db.DeleteUser(msg.ChatID)
				log.Printf("WARN user %d is blocked the bot and now deleted", msg.ChatID)
				continue
			}

			log.Printf("ERROR %T Problem while replying on %d chat: %s", err, msg.ChatID, err)
		}
	}

	log.Print("INFO Reply queue channel was closed")
}
