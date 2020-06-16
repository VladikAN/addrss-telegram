package server

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/vladikan/feedreader-telegrambot/commands"
)

func handleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	msg := update.Message
	cmd := commands.Build(msg)
	if cmd == nil {
		rsp := tgbotapi.NewMessage(msg.Chat.ID, "Sorry, command is unknown")
		rsp.ReplyToMessageID = msg.MessageID
		bot.Send(rsp)

		return
	}

	cmd.Execute()
}
