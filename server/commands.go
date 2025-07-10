package server

import (
	"fmt"
	"strings"

	log "github.com/go-pkgz/lgr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/vladikan/addrss-telegram/database"
	"github.com/vladikan/addrss-telegram/parser"
	"github.com/vladikan/addrss-telegram/templates"
)

// Command is to aggregate message information and execute user command
type Command struct {
	userID     int64
	admin      bool
	adminID    int64
	verb       string
	args       string
	fileId     string
	lang       string
	raw        *tgbotapi.Message
	replyQueue chan Reply
}

var emptyText string

func newCommand(msg *tgbotapi.Message, opt *Options, replyQueue chan Reply) *Command {
	cmd := &Command{
		userID:     msg.Chat.ID,
		admin:      msg.Chat.ID == opt.BotAdmin,
		adminID:    opt.BotAdmin,
		verb:       msg.CommandWithAt(),
		args:       msg.CommandArguments(),
		lang:       msg.From.LanguageCode,
		raw:        msg,
		replyQueue: replyQueue,
	}

	if msg.Document != nil {
		cmd.fileId = msg.Document.FileID
	}

	return cmd
}

func (cmd *Command) run() []Reply {
	log.Printf("DEBUG request: %s", cmd.raw.Text)

	var replies []Reply
	var response string
	var err error

	if len(cmd.fileId) != 0 {
		response, err = cmd.importOpml()
	}

	if len(cmd.verb) > 0 {
		switch cmd.verb {
		case "stats":
			response, err = cmd.stats()
		case "notify":
			replies = cmd.notifyMulti()
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
		case "feedback":
			replies = cmd.feedbackMulti()
		}

		log.Printf("INFO User %d call '%s'", cmd.userID, cmd.verb)
	}

	if len(replies) > 0 {
		return replies
	}

	if err != nil {
		log.Printf("ERROR user %d command '%s' completed with error: '%s'", cmd.userID, cmd.raw.Text, err)
		response, _ = templates.ToText(cmd.lang, "cmd-error")
	}

	if len(response) == 0 {
		log.Printf("WARN command '%s' is unknown", cmd.raw.Text)
		response, _ = templates.ToText(cmd.lang, "cmd-unknown")
	}

	return []Reply{{ChatID: cmd.userID, Text: response}}
}

func (cmd *Command) stats() (string, error) {
	if !cmd.admin {
		log.Printf("WARN Someone is calling /stats with no admin rights, user %d", cmd.userID)
		return templates.ToText(cmd.lang, "cmd-unknown")
	}

	if stats, err := db.GetStats(); err != nil {
		return "", err
	} else {
		return templates.ToTextW(cmd.lang, "stats-success", stats)
	}
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
		return emptyText, fmt.Errorf("error while parsing OMPL file, %s", err)
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

func (cmd *Command) feedbackMulti() []Reply {
	const maxFeedbackLength = 1000
	var replies []Reply

	if len(strings.TrimSpace(cmd.args)) == 0 {
		text, _ := templates.ToText(cmd.lang, "feedback-validation")
		replies = append(replies, Reply{ChatID: cmd.userID, Text: text})
		return replies
	}

	if len(cmd.args) > maxFeedbackLength {
		text, _ := templates.ToTextW(cmd.lang, "feedback-too-long", maxFeedbackLength)
		replies = append(replies, Reply{ChatID: cmd.userID, Text: text})
		return replies
	}

	feedbackText, _ := templates.ToTextW(cmd.lang, "feedback-message", struct {
		UserID  int64
		Message string
	}{cmd.userID, cmd.args})
	replies = append(replies, Reply{ChatID: cmd.adminID, Text: feedbackText})
	text, _ := templates.ToText(cmd.lang, "feedback-success")
	replies = append(replies, Reply{ChatID: cmd.userID, Text: text})
	return replies
}

func (cmd *Command) notifyMulti() []Reply {
	const maxNotifyLength = 2000
	var replies []Reply

	if !cmd.admin {
		text, _ := templates.ToText(cmd.lang, "cmd-unknown")
		replies = append(replies, Reply{ChatID: cmd.userID, Text: text})
		return replies
	}

	if len(strings.TrimSpace(cmd.args)) == 0 {
		text, _ := templates.ToText(cmd.lang, "notify-validation")
		replies = append(replies, Reply{ChatID: cmd.userID, Text: text})
		return replies
	}

	if len(cmd.args) > maxNotifyLength {
		text, _ := templates.ToTextW(cmd.lang, "notify-too-long", maxNotifyLength)
		replies = append(replies, Reply{ChatID: cmd.userID, Text: text})
		return replies
	}

	userIDs, err := db.GetAllUsers()
	if err != nil {
		text, _ := templates.ToText(cmd.lang, "notify-error")
		replies = append(replies, Reply{ChatID: cmd.userID, Text: text})
		return replies
	}

	notificationText, _ := templates.ToTextW(cmd.lang, "notify-message", struct {
		Message string
	}{cmd.args})
	for _, userID := range userIDs {
		replies = append(replies, Reply{ChatID: userID, Text: notificationText})
	}

	summary, _ := templates.ToTextW(cmd.lang, "notify-success", struct {
		Total int
	}{len(userIDs)})
	replies = append(replies, Reply{ChatID: cmd.userID, Text: summary})
	return replies
}
