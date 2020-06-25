package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	log "github.com/go-pkgz/lgr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/vladikan/addrss-telegram/database"
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
var db *database.Database

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

	// Set db connection settings and use pool
	db = &database.Database{Connection: options.Connection}
	err = db.Open(ctx)
	if err != nil {
		log.Printf("PANIC Error while connecting to the database: %s", err)
	}

	// Init messages channel
	outbox := make(chan Reply)
	go handleOutbox(outbox)

	// Start reader
	reader := &Reader{Interval: options.ReaderInterval, Feeds: options.ReaderFeeds, DB: db, Outbox: outbox}
	reader.Start()

	// Read commands from users
	updates, err := bot.GetUpdatesChan(cfg)
	go handleRequests(updates, outbox)

	// Stop bot operations and close all connections
	<-ctx.Done()

	reader.Stop()
	bot.StopReceivingUpdates()
	db.Close()
	close(outbox)

	log.Print("INFO Stopped updates processing")
}

func handleRequests(updates tgbotapi.UpdatesChannel, outbox chan Reply) {
	log.Print("INFO Start updates processing")
	for update := range updates {
		if update.Message == nil {
			return
		}

		msg := update.Message
		txt := runCommand(msg)
		outbox <- Reply{ChatID: msg.Chat.ID, Text: txt}
	}
}

func handleOutbox(outbox chan Reply) {
	for msg := range outbox {
		rsp := tgbotapi.NewMessage(msg.ChatID, msg.Text)
		rsp.ParseMode = "HTML"

		if _, err := bot.Send(rsp); err != nil {
			log.Printf("ERROR Problem while replying on %d chat: %s", msg.ChatID, err)
		}
	}
}
