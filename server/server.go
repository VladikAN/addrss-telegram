package server

import (
	"context"
	"os"
	"os/signal"
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
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Print("WARN System interrupt signal")
		cancel()
	}()

	templates.SetTemplateOutput()

	// Set db connection settings and use pool
	db, err = database.Open(ctx, options.Connection)
	if err != nil {
		log.Printf("PANIC Error while connecting to the database: %s", err)
	}

	// Init messages channel
	replyQueue := make(chan Reply)
	go handleReply(replyQueue)

	// Start reader
	reader := &Reader{Interval: options.ReaderInterval, Feeds: options.ReaderFeeds, DB: db, Outbox: replyQueue}
	reader.Start()

	// Read commands from users
	updates, err := bot.GetUpdatesChan(cfg)
	go handleRequests(updates, replyQueue)

	// Stop bot operations and close all connections
	<-ctx.Done()

	reader.Stop()
	bot.StopReceivingUpdates()
	db.Close()
	close(replyQueue)

	log.Print("INFO Stopped updates processing")
}

func handleRequests(updates tgbotapi.UpdatesChannel, replyQueue chan Reply) {
	log.Print("INFO Start updates processing")
	for update := range updates {
		if update.Message == nil {
			return
		}

		msg := update.Message
		txt := runCommand(msg)
		replyQueue <- Reply{ChatID: msg.Chat.ID, Text: txt}
	}

	log.Print("DEBUG telegram updates channel was closed")
}

func handleReply(queue chan Reply) {
	for msg := range queue {
		rsp := tgbotapi.NewMessage(msg.ChatID, msg.Text)
		rsp.ParseMode = "HTML"

		if _, err := bot.Send(rsp); err != nil {
			log.Printf("ERROR Problem while replying on %d chat: %s", msg.ChatID, err)
		}
	}

	log.Print("DEBUG reply queue channel was closed")
}
