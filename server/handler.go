package server

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func handleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	msg := update.Message

	rsp := tgbotapi.NewMessage(msg.Chat.ID, msg.Text)
	rsp.ReplyToMessageID = msg.MessageID

	bot.Send(rsp)
}
