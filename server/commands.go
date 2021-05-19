package server

import (
	"fmt"

	log "github.com/go-pkgz/lgr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/vladikan/addrss-telegram/database"
	"github.com/vladikan/addrss-telegram/parser"
	"github.com/vladikan/addrss-telegram/templates"
)

// Command is to aggregate message information and execute user command
type Command struct {
	userID int64
	admin  bool
	verb   string
	args   string
	fileId string
	lang   string
	raw    *tgbotapi.Message
}

var emptyText string

func newCommand(msg *tgbotapi.Message, opt *Options) *Command {
	cmd := &Command{
		userID: msg.Chat.ID,
		admin:  msg.Chat.ID == opt.BotAdmin,
		verb:   msg.CommandWithAt(),
		args:   msg.CommandArguments(),
		lang:   msg.From.LanguageCode,
		raw:    msg,
	}

	if msg.Document != nil {
		cmd.fileId = msg.Document.FileID
	}

	return cmd
}

func (cmd *Command) run() string {
	log.Printf("DEBUG Received message: %s", cmd.raw.Text)

	var response string
	var err error

	if len(cmd.fileId) != 0 {
		response, err = cmd.importOpml()
	}

	if len(cmd.verb) > 0 {
		switch cmd.verb {
		case "start":
			response, err = cmd.start()
		case "help":
			response, err = cmd.help()
		case "add":
			response, err = cmd.add()
		case "import":
			response, err = cmd.importOpml() // simply call for validation message
		case "remove":
			response, err = cmd.remove()
		case "list":
			response, err = cmd.list()
		}
	}

	if err != nil {
		log.Printf("ERROR user %d command '%s' completed with error: '%s'", cmd.userID, cmd.raw.Text, err)
		response, _ = templates.ToText(cmd.lang, "cmd-error")
	}

	if len(response) == 0 {
		log.Printf("WARN command '%s' is unknown", cmd.raw.Text)
		response, _ = templates.ToText(cmd.lang, "cmd-unknown")
	}

	return response
}

func (cmd *Command) stats() (string, error) {
	if !cmd.admin {
		return "", nil
	}

	return "", nil
}

func (cmd *Command) start() (string, error) {
	return templates.ToText(cmd.lang, "start-success")
}

func (cmd *Command) help() (string, error) {
	return templates.ToText(cmd.lang, "help-success")
}

func (cmd *Command) add() (string, error) {
	if len(cmd.args) == 0 {
		return templates.ToText(cmd.lang, "add-validation")
	}

	if userFeed, err := db.GetUserURIFeed(cmd.userID, cmd.args); err != nil {
		return emptyText, err
	} else if userFeed != nil {
		return templates.ToTextW(cmd.lang, "add-exists", userFeed)
	}

	feed, err := addFeed(cmd.userID, cmd.args, "")
	if err != nil {
		return emptyText, err
	}

	return templates.ToTextW(cmd.lang, "add-success", feed)
}

func (cmd *Command) importOpml() (string, error) {
	if len(cmd.fileId) == 0 {
		return templates.ToText(cmd.lang, "import-validation")
	}

	fl, err := bot.GetFileDirectURL(cmd.fileId)
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
		_, err = addFeed(cmd.userID, item.URL, item.Title)
		if err != nil {
			log.Printf("ERROR Feed '%s' was not imported with error: '%s'", item.URL, err)
			result.Errors++
		} else {
			result.Added++
		}
	}

	return templates.ToTextW(cmd.lang, "import-success", result)
}

func (cmd *Command) remove() (string, error) {
	if len(cmd.args) == 0 {
		return templates.ToText(cmd.lang, "remove-validation")
	}

	feed, err := db.GetUserNormalizedFeed(cmd.userID, cmd.args)
	if err != nil {
		return emptyText, err
	}

	if feed == nil {
		return templates.ToText(cmd.lang, "remove-no-rows")
	}

	err = db.Unsubscribe(cmd.userID, feed.ID)
	if err != nil {
		return emptyText, err
	}

	return templates.ToTextW(cmd.lang, "remove-success", feed)
}

func (cmd *Command) list() (string, error) {
	feeds, err := db.GetUserFeeds(cmd.userID)

	if err != nil {
		return emptyText, err
	}

	if len(feeds) == 0 {
		return templates.ToText(cmd.lang, "list-empty")
	}

	return templates.ToTextW(cmd.lang, "list-result", feeds)
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
