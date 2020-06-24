package server

import (
	log "github.com/go-pkgz/lgr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/vladikan/feedreader-telegrambot/parser"
	"github.com/vladikan/feedreader-telegrambot/templates"
)

// Command is used for propper method execution
type Command struct {
	UserID int64
	Args   []string
}

var emptyText string

func runCommand(msg *tgbotapi.Message) string {
	var response string
	var err error

	if cmd := msg.CommandWithAt(); len(cmd) > 0 {
		command := &Command{UserID: msg.Chat.ID} // Use chat ID as unique user
		args := msg.CommandArguments()

		switch cmd {
		case "start":
			response, err = command.start()
		case "add":
			command.Args = splitURI(args)
			response, err = command.add()
		case "remove":
			command.Args = splitNonEmpty(args)
			response, err = command.remove()
		case "list":
			response, err = command.list()
		}
	}

	if err != nil {
		log.Printf("ERROR user %d command '%s' completed with error: '%s'", msg.From.ID, msg.Text, err)
		response, _ = templates.ToText("cmd-error")
	}

	if len(response) == 0 {
		log.Printf("WARN command '%s' is unknown", msg.Text)
		response, _ = templates.ToText("cmd-unknown")
	}

	return response
}

func (cmd *Command) start() (string, error) {
	return templates.ToText("start-success")
}

func (cmd *Command) add() (string, error) {
	if len(cmd.Args) == 0 {
		return templates.ToText("add-validation")
	}

	uri := cmd.Args[0] // will use only one for now
	if userFeed, err := db.GetUserURIFeed(cmd.UserID, uri); err != nil {
		return emptyText, err
	} else if userFeed != nil {
		return templates.ToTextW("add-exists", userFeed)
	}

	feed, err := db.GetFeed(uri)
	if err != nil {
		return emptyText, err
	}

	if feed == nil {
		title, err := parser.GetTitle(uri)
		if err != nil {
			return emptyText, err
		}

		feed, err = db.AddFeed(title, normalize(title), uri)
		if err != nil {
			return emptyText, err
		}
	}

	err = db.Subscribe(cmd.UserID, feed.ID)
	if err != nil {
		return emptyText, err
	}

	return templates.ToTextW("add-success", feed)
}

func (cmd *Command) remove() (string, error) {
	if len(cmd.Args) == 0 {
		return templates.ToText("remove-validation")
	}

	name := cmd.Args[0] // will use only one for now
	feed, err := db.GetUserNormalizedFeed(cmd.UserID, name)

	if err != nil {
		return emptyText, err
	}

	if feed == nil {
		return templates.ToText("remove-no-rows")
	}

	err = db.Unsubscribe(cmd.UserID, feed.ID)
	if err != nil {
		return emptyText, err
	}

	return templates.ToTextW("remove-success", feed)
}

func (cmd *Command) list() (string, error) {
	feeds, err := db.GetUserFeeds(cmd.UserID)

	if err != nil {
		return emptyText, err
	}

	if len(feeds) == 0 {
		return templates.ToText("list-empty")
	}

	return templates.ToTextW("list-result", feeds)
}
