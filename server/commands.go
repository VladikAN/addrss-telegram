package server

import (
	"fmt"

	log "github.com/go-pkgz/lgr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/vladikan/addrss-telegram/database"
	"github.com/vladikan/addrss-telegram/parser"
	"github.com/vladikan/addrss-telegram/templates"
)

// Command is used for propper method execution
type Command struct {
	UserID int64
	Args   []string
	Lang   string
}

var emptyText string

func runCommand(msg *tgbotapi.Message) string {
	log.Printf("DEBUG Received message: %s", msg.Text)

	var response string
	var err error

	command := &Command{UserID: msg.Chat.ID, Lang: msg.From.LanguageCode}
	if msg.Document != nil {
		command.Args = []string{msg.Document.FileID}
		response, err = command.importOpml()
	}

	if cmd := msg.CommandWithAt(); len(cmd) > 0 {
		args := msg.CommandArguments()

		switch cmd {
		case "start":
			response, err = command.start()
		case "help":
			response, err = command.help()
		case "add":
			command.Args = splitURI(args)
			response, err = command.add()
		case "import":
			response, err = command.importOpml() // simply call for validation message
		case "remove":
			command.Args = splitNonEmpty(args)
			response, err = command.remove()
		case "list":
			response, err = command.list()
		}
	}

	if err != nil {
		log.Printf("ERROR user %d command '%s' completed with error: '%s'", msg.From.ID, msg.Text, err)
		response, _ = templates.ToText(command.Lang, "cmd-error")
	}

	if len(response) == 0 {
		log.Printf("WARN command '%s' is unknown", msg.Text)
		response, _ = templates.ToText(command.Lang, "cmd-unknown")
	}

	return response
}

func (cmd *Command) start() (string, error) {
	return templates.ToText(cmd.Lang, "start-success")
}

func (cmd *Command) help() (string, error) {
	return templates.ToText(cmd.Lang, "help-success")
}

func (cmd *Command) add() (string, error) {
	if len(cmd.Args) == 0 {
		return templates.ToText(cmd.Lang, "add-validation")
	}

	uri := cmd.Args[0] // will use only one for now
	if userFeed, err := db.GetUserURIFeed(cmd.UserID, uri); err != nil {
		return emptyText, err
	} else if userFeed != nil {
		return templates.ToTextW(cmd.Lang, "add-exists", userFeed)
	}

	feed, err := addFeed(cmd.UserID, uri, "")
	if err != nil {
		return emptyText, err
	}

	return templates.ToTextW(cmd.Lang, "add-success", feed)
}

func (cmd *Command) importOpml() (string, error) {
	if len(cmd.Args) == 0 {
		return templates.ToText(cmd.Lang, "import-validation")
	}

	fl, err := bot.GetFileDirectURL(cmd.Args[0])
	if err != nil {
		return emptyText, err
	}

	items, err := parser.ReadOmpl(fl)
	if err != nil {
		return emptyText, fmt.Errorf("Error while parsing OMPL file, %s", err)
	}

	result := struct {
		Added  int
		Errors int
	}{}

	for _, item := range items {
		_, err = addFeed(cmd.UserID, item.URL, item.Title)
		if err != nil {
			log.Printf("ERROR Feed '%s' was not imported with error: '%s'", item.URL, err)
			result.Errors++
		} else {
			result.Added++
		}
	}

	return templates.ToTextW(cmd.Lang, "import-success", result)
}

func (cmd *Command) remove() (string, error) {
	if len(cmd.Args) == 0 {
		return templates.ToText(cmd.Lang, "remove-validation")
	}

	name := cmd.Args[0] // will use only one for now
	feed, err := db.GetUserNormalizedFeed(cmd.UserID, name)

	if err != nil {
		return emptyText, err
	}

	if feed == nil {
		return templates.ToText(cmd.Lang, "remove-no-rows")
	}

	err = db.Unsubscribe(cmd.UserID, feed.ID)
	if err != nil {
		return emptyText, err
	}

	return templates.ToTextW(cmd.Lang, "remove-success", feed)
}

func (cmd *Command) list() (string, error) {
	feeds, err := db.GetUserFeeds(cmd.UserID)

	if err != nil {
		return emptyText, err
	}

	if len(feeds) == 0 {
		return templates.ToText(cmd.Lang, "list-empty")
	}

	return templates.ToTextW(cmd.Lang, "list-result", feeds)
}

func addFeed(userID int64, uri string, title string) (*database.Feed, error) {
	feed, err := db.GetFeed(uri)
	if err != nil {
		return nil, err
	}

	if feed == nil {
		if len(title) == 0 {
			title, err = parser.GetTitle(uri)
			if err != nil {
				return nil, err
			}
		}

		feed, err = db.AddFeed(title, normalize(title), uri)
		if err != nil {
			return nil, err
		}
	} else {
		_ = db.ResetFeed(feed.ID)
	}

	err = db.Subscribe(userID, feed.ID)
	if err != nil {
		return nil, err
	}

	return feed, nil
}
