package commands

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

// Build constructs actual command based on incomming message
func Build(msg *tgbotapi.Message) Command {
	if cmd := msg.CommandWithAt(); len(cmd) > 0 {
		arg := msg.CommandArguments()
		base := baseCommand{MessageID: msg.MessageID, UserID: msg.From.ID, Command: cmd, Args: arg}

		switch cmd {
		case "start":
			return StartCommand{Base: base}
		case "add":
			return AddCommand{Base: base}
		case "remove":
			return RemoveCommand{Base: base}
		case "list":
			return ListCommand{Base: base}
		case "read":
			return ReadCommand{Base: base}
		}
	}

	return nil
}

// Execute will run Start command
func (cmd StartCommand) Execute() {}

// Execute will run Add command
func (cmd AddCommand) Execute() {}

// Execute will run Remove command
func (cmd RemoveCommand) Execute() {}

// Execute will run List command
func (cmd ListCommand) Execute() {}

// Execute will run Read command
func (cmd ReadCommand) Execute() {}
