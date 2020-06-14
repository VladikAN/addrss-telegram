package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	log "github.com/go-pkgz/lgr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var bot *tgbotapi.BotAPI

// Start will call for bot instance and process update messages
func Start(token string, debug bool) {
	bt, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Printf("PANIC Error while creating bot instance: %s", err)
	}

	bot = bt
	bot.Debug = debug
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
			handleUpdate(update)
		}
	}()

	<-ctx.Done()
	log.Print("INFO Stop updates processing")
	bot.StopReceivingUpdates()
}
