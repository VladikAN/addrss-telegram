package server

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func runCommand(msg *tgbotapi.Message) (string, error) {
	if cmd := msg.CommandWithAt(); len(cmd) > 0 {
		args := msg.CommandArguments()

		switch cmd {
		case "start":
			return start(msg.From.ID)
		case "add":
			return add(msg.From.ID, splitURI(args))
		case "remove":
			return remove(msg.From.ID, splitURI(args))
		case "list":
			return list(msg.From.ID)
		case "read":
			return read(msg.From.ID, splitNonEmpty(args))
		}
	}

	return "Sorry, command is unknown", nil
}

func start(id int) (string, error) {
	return "", nil
}

func add(it int, uris []string) (string, error) {
	if len(uris) == 0 {
		return "TODO: help text", nil
	}

	return "", nil
}

func remove(id int, uris []string) (string, error) {
	if len(uris) == 0 {
		return "TODO: help text", nil
	}

	return "", nil
}

func list(id int) (string, error) {
	return "", nil
}

func read(id int, names []string) (string, error) {
	return "", nil
}
