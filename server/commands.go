package server

import (
	log "github.com/go-pkgz/lgr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jackc/pgx/v4"
	"github.com/vladikan/feedreader-telegrambot/parser"
	"github.com/vladikan/feedreader-telegrambot/templates"
)

// Feed represents feed db table structure
type Feed struct {
	ID         int
	Name       string
	Normalized string
	URI        string
}

// Command is used for propper method execution
type Command struct {
	UserID int
	Args   []string
}

func runCommand(msg *tgbotapi.Message) string {
	var response string
	var err error

	if cmd := msg.CommandWithAt(); len(cmd) > 0 {
		command := &Command{UserID: msg.From.ID}
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
	if userFeed, err := getUserFeed(cmd.UserID, uri); err != nil {
		return "", err
	} else if userFeed != nil {
		return templates.ToTextW("add-exists", userFeed)
	}

	feed, err := getFeed(uri)
	if err != nil {
		return "", err
	}

	if feed == nil {
		title, err := parser.GetFeed(uri)
		if err != nil {
			return "", err
		}

		query := `INSERT INTO feeds (name, normalized, uri) VALUES ($1, $2, $3) ON CONFLICT (uri) DO NOTHING`
		_, err = db.Pool.Exec(db.Context, query, title, normalize(title), uri)
		if err != nil {
			return "", err
		}

		feed, err = getFeed(uri)
		if err != nil {
			return "", err
		}
	}

	query := `INSERT INTO userfeeds (user_id, feed_id) VALUES ($1, $2) ON CONFLICT (user_id, feed_id) DO NOTHING`
	_, err = db.Pool.Exec(db.Context, query, cmd.UserID, feed.ID)
	if err != nil {
		return "", err
	}

	return templates.ToTextW("add-success", feed)
}

func (cmd *Command) remove() (string, error) {
	if len(cmd.Args) == 0 {
		return templates.ToText("remove-validation")
	}

	name := cmd.Args[0] // will use only one for now
	feed, err := getUserNormalizedFeed(cmd.UserID, name)

	if err != nil {
		return "", err
	}

	if feed == nil {
		return templates.ToText("remove-no-rows")
	}

	query := `DELETE FROM userfeeds WHERE user_id = $1 AND feed_id = $2`
	_, err = db.Pool.Exec(db.Context, query, cmd.UserID, feed.ID)
	if err != nil {
		return "", err
	}

	return templates.ToTextW("remove-success", feed)
}

func (cmd *Command) list() (string, error) {
	query := `SELECT f.name, f.normalized, f.uri FROM userfeeds uf
	INNER JOIN feeds f ON f.id = uf.feed_id
	WHERE uf.user_id = $1
	ORDER BY uf.added`

	rows, err := db.Pool.Query(db.Context, query, cmd.UserID)
	defer rows.Close()
	if err != nil {
		return "", err
	}

	var feeds []Feed
	for rows.Next() {
		var name, normalized, uri string
		err = rows.Scan(&name, &normalized, &uri)
		if err != nil {
			return "", err
		}
		feeds = append(feeds, Feed{Name: name, Normalized: normalized, URI: uri})
	}

	if len(feeds) == 0 {
		return templates.ToText("list-empty")
	}

	return templates.ToTextW("list-result", feeds)
}

func getUserFeed(userID int, uri string) (*Feed, error) {
	query := `SELECT f.id, f.name, f.normalized, f.uri FROM userfeeds uf
	INNER JOIN feeds f ON f.id = uf.feed_id
	WHERE uf.user_id = $1 AND f.uri = $2
	LIMIT 1`

	rows, err := db.Pool.Query(db.Context, query, userID, uri)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	return rowToFeed(rows)
}

func getUserNormalizedFeed(userID int, normalized string) (*Feed, error) {
	query := `SELECT f.id, f.name, f.normalized, f.uri FROM userfeeds uf
	INNER JOIN feeds f ON f.id = uf.feed_id
	WHERE uf.user_id = $1 AND f.normalized = $2
	LIMIT 1`

	rows, err := db.Pool.Query(db.Context, query, userID, normalized)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	return rowToFeed(rows)
}

func getFeed(uri string) (*Feed, error) {
	query := `SELECT id, name, normalized, uri
	FROM feeds
	WHERE uri = $1
	LIMIT 1`

	rows, err := db.Pool.Query(db.Context, query, uri)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	return rowToFeed(rows)
}

func rowToFeed(rows pgx.Rows) (*Feed, error) {
	if !rows.Next() {
		return nil, nil
	}

	var feedID int
	var name string
	var normalized string
	var uri string

	if err := rows.Scan(&feedID, &name, &normalized, &uri); err == nil {
		return &Feed{ID: feedID, Name: name, Normalized: normalized, URI: uri}, err
	} else if err != nil {
		return nil, err
	}

	return nil, nil
}
