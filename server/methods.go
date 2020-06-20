package server

import (
	"bytes"
	"fmt"
	"text/template"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
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

func start(id int) (string, error) {
	return toText("start-success", nil), nil
}

func add(id int, uris []string) (string, error) {
	if len(uris) == 0 {
		return toText("add-validation", nil), nil
	}

	trg := uris[0] // will use only one for now
	query := `INSERT INTO feeds (name, uri) VALUES ($1, $2) ON CONFLICT (uri) DO NOTHING`
	_, err := db.Pool.Exec(db.Context, query, "TODO name", trg)
	if err != nil {
		return "", err
	}

	var feedID int
	var name string
	query = `SELECT id, name FROM feeds WHERE uri = $1`
	err = db.Pool.QueryRow(db.Context, query, trg).Scan(&feedID, &name)
	if err != nil {
		return "", err
	}

	query = `INSERT INTO userfeeds (user_id, feed_id) VALUES ($1, $2) ON CONFLICT (user_id, feed_id) DO NOTHING`
	_, err = db.Pool.Exec(db.Context, query, id, feedID)
	if err != nil {
		return "", err
	}

	return toText("add-success", name), nil
}

func remove(id int, uris []string) (string, error) {
	if len(uris) == 0 {
		return toText("remove-validation", nil), nil
	}

	trg := uris[0] // will use only one for now

	query := `SELECT f.id, f.name FROM userfeeds uf
	INNER JOIN feeds f ON f.id = uf.feed_id
	WHERE uf.user_id = $1 AND f.uri = $2`
	rows, err := db.Pool.Query(db.Context, query, id, trg)
	defer rows.Close()

	if err != nil {
		return "", err
	}

	if !rows.Next() {
		return toText("remove-no-rows", nil), nil
	}

	var feedID int
	var name string
	if err := rows.Scan(&feedID, &name); err != nil {
		return "", err
	}

	query = `DELETE FROM userfeeds WHERE user_id = $1 AND feed_id = $2`
	_, err = db.Pool.Exec(db.Context, query, id, feedID)
	if err != nil {
		return "", err
	}

	return toText("remove-success", name), nil
}

func list(id int) (string, error) {
	query := `SELECT f.name, f.uri FROM userfeeds uf
	INNER JOIN feeds f ON f.id = uf.feed_id
	WHERE uf.user_id = $1
	ORDER BY uf.added`

	rows, err := db.Pool.Query(db.Context, query, id)
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

func read(id int, names []string) (string, error) {
	return "", nil
}

func toText(name string, data interface{}) string {
	var tpl bytes.Buffer
	help, _ := template.ParseFiles(fmt.Sprintf("templates/%s.txt", name))
	help.Execute(&tpl, data)

	return tpl.String()
}
