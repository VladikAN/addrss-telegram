package server

import (
	"time"

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

func runCommand(msg *tgbotapi.Message) (string, error) {
	if cmd := msg.CommandWithAt(); len(cmd) > 0 {
		args := msg.CommandArguments()

		// Use db connections pools
		err := db.Open()
		if err != nil {
			return "", err
		}
		defer db.Close()

		// Run command itself
		switch cmd {
		case "start":
			return start(msg.From.ID)
		case "add":
			return add(msg.From.ID, splitURI(args))
		case "remove":
			return remove(msg.From.ID, splitNonEmpty(args))
		case "list":
			return list(msg.From.ID)
		case "read":
			return read(msg.From.ID, splitNonEmpty(args))
		}
	}

	return "Sorry, command is unknown", nil
}

func start(userID int) (string, error) {
	return templates.ToText("start-success"), nil
}

func add(userID int, uris []string) (string, error) {
	if len(uris) == 0 {
		return templates.ToText("add-validation"), nil
	}

	uri := uris[0] // will use only one for now
	if userFeed, err := getUserFeed(userID, uri); err != nil {
		return "", err
	} else if userFeed != nil {
		return templates.ToTextW("add-exists", userFeed), nil
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

	updated := time.Now().AddDate(0, 0, -1)
	query := `INSERT INTO userfeeds (user_id, feed_id, updated) VALUES ($1, $2, $3) ON CONFLICT (user_id, feed_id) DO NOTHING`
	_, err = db.Pool.Exec(db.Context, query, userID, feed.ID, updated)
	if err != nil {
		return "", err
	}

	return templates.ToTextW("add-success", feed), nil
}

func remove(userID int, names []string) (string, error) {
	if len(names) == 0 {
		return templates.ToText("remove-validation"), nil
	}

	name := names[0] // will use only one for now
	feed, err := getUserNormalizedFeed(userID, name)

	if err != nil {
		return "", err
	}

	if feed == nil {
		return templates.ToText("remove-no-rows"), nil
	}

	query := `DELETE FROM userfeeds WHERE user_id = $1 AND feed_id = $2`
	_, err = db.Pool.Exec(db.Context, query, userID, feed.ID)
	if err != nil {
		return "", err
	}

	return templates.ToTextW("remove-success", feed), nil
}

func list(userID int) (string, error) {
	query := `SELECT f.name, f.normalized, f.uri FROM userfeeds uf
	INNER JOIN feeds f ON f.id = uf.feed_id
	WHERE uf.user_id = $1
	ORDER BY uf.added`

	rows, err := db.Pool.Query(db.Context, query, userID)
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
		return templates.ToText("list-empty"), nil
	}

	return templates.ToTextW("list-result", feeds), nil
}

func read(userID int, names []string) (string, error) {
	return "", nil
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
