package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	log "github.com/go-pkgz/lgr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/vladikan/feedreader-telegrambot/database"
)

// Options holds all necessary settings for the app
type Options struct {
	Token          string
	Connection     string
	Debug          bool
	ReaderInterval int
	ReaderFeeds    int
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
	defer db.Close()

	// Start reader
	reader := &Reader{Interval: options.ReaderInterval, Feeds: options.ReaderFeeds, DB: db}
	reader.Start()
	defer reader.Stop()

	// Read commands from users
	updates, err := bot.GetUpdatesChan(cfg)
	defer bot.StopReceivingUpdates()

	go func() {
		log.Print("INFO Start updates processing")
		for update := range updates {
			handleRequest(update)
		}
	}()

	// Stop bot operations and close db connection
	<-ctx.Done()
	log.Print("INFO Stopped updates processing")
}

func handleRequest(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	msg := update.Message
	txt := runCommand(msg)

	rsp := tgbotapi.NewMessage(msg.Chat.ID, txt)
	rsp.ParseMode = "HTML"
	bot.Send(rsp)
}
