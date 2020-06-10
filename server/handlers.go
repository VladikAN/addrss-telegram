package server

import (
	log "github.com/go-pkgz/lgr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func handleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	log.Printf("DEBUG [%s] %s", update.Message.From.ID, update.Message.Text)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	msg.ReplyToMessageID = update.Message.MessageID

	bot.Send(msg)
}
