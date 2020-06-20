package server

import (
	"bytes"
	"fmt"
	"text/template"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jackc/pgx/v4"
)

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
			return remove(msg.From.ID, splitURI(args))
		case "list":
			return list(msg.From.ID)
		case "read":
			return read(msg.From.ID, splitNonEmpty(args))
		}
	}

	return "Sorry, command is unknown", nil
}

func start(userID int) (string, error) {
	return toText("start-success", nil), nil
}

func add(userID int, uris []string) (string, error) {
	if len(uris) == 0 {
		return toText("add-validation", nil), nil
	}

	trg := uris[0] // will use only one for now
	if userFeed, err := getUserFeed(userID, trg); err != nil {
		return "", err
	} else if userFeed != nil {
		return toText("add-exists", userFeed), err
	}

	feed, err := getFeed(trg)
	if err != nil {
		return "", err
	}

	if feed == nil {
		//TODO parse name

		query := `INSERT INTO feeds (name, uri) VALUES ($1, $2) ON CONFLICT (uri) DO NOTHING`
		_, err := db.Pool.Exec(db.Context, query, "TODO name", trg)
		if err != nil {
			return "", err
		}

		feed, err = getFeed(trg)
		if err != nil {
			return "", err
		}
	}

	query := `INSERT INTO userfeeds (user_id, feed_id) VALUES ($1, $2) ON CONFLICT (user_id, feed_id) DO NOTHING`
	_, err = db.Pool.Exec(db.Context, query, userID, feed.ID)
	if err != nil {
		return "", err
	}

	return toText("add-success", feed), nil
}

func remove(userID int, uris []string) (string, error) {
	if len(uris) == 0 {
		return toText("remove-validation", nil), nil
	}

	trg := uris[0] // will use only one for now
	feed, err := getUserFeed(userID, trg)

	if err != nil {
		return "", err
	}

	if feed == nil {
		return toText("remove-no-rows", nil), nil
	}

	query := `DELETE FROM userfeeds WHERE user_id = $1 AND feed_id = $2`
	_, err = db.Pool.Exec(db.Context, query, userID, feed.ID)
	if err != nil {
		return "", err
	}

	return toText("remove-success", feed), nil
}

func list(userID int) (string, error) {
	query := `SELECT f.name, f.uri FROM userfeeds uf
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
		var name, uri string
		err = rows.Scan(&name, &uri)
		if err != nil {
			return "", err
		}
		feeds = append(feeds, Feed{Name: name, URI: uri})
	}

	if len(feeds) == 0 {
		return toText("list-empty", nil), nil
	}

	return toText("list-result", feeds), nil
}

func read(userID int, names []string) (string, error) {
	return "", nil
}

func getUserFeed(userID int, uri string) (*Feed, error) {
	query := `SELECT f.id, f.name, f.uri FROM userfeeds uf
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

func getFeed(uri string) (*Feed, error) {
	query := `SELECT id, name, uri
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
	var uri string

	if err := rows.Scan(&feedID, &name, &uri); err == nil {
		return &Feed{ID: feedID, Name: name, URI: uri}, err
	} else if err != nil {
		return nil, err
	}

	return nil, nil
}

func toText(name string, data interface{}) string {
	var tpl bytes.Buffer
	help, _ := template.ParseFiles(fmt.Sprintf("templates/%s.txt", name))
	help.Execute(&tpl, data)

	return tpl.String()
}
