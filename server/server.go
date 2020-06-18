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
	log.Printf("Authorized on account %s", bot.Self.UserName)

	cfg := tgbotapi.NewUpdate(0)
	cfg.Timeout = 60

	// Hook for system signal
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Print("WARN System interrupt signal")
		cancel()
	}()

	// Read commands from users
	log.Print("INFO Start updates processing")
	updates, err := bot.GetUpdatesChan(cfg)
	go func() {
		for update := range updates {
			handleRequest(update)
		}
	}()

	// Stop bot operations and close db connection
	<-ctx.Done()
	log.Print("INFO Stop updates processing")
	bot.StopReceivingUpdates()
}

func handleRequest(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	msg := update.Message
	txt, err := runCommand(msg)
	if err == nil {
		rsp := tgbotapi.NewMessage(msg.Chat.ID, txt)
		bot.Send(rsp)

		return
	}

	log.Printf("ERROR command '%s' completed with error: %s", msg.Text, err)
	rsp := tgbotapi.NewMessage(msg.Chat.ID, "Sorry we have a error while processing your request")
	bot.Send(rsp)
}
