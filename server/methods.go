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
		db.Open()
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
	return "", nil
}

func add(it int, uris []string) (string, error) {
	if len(uris) == 0 {
		return toText("add-validation", nil), nil
	}

	trg := uris[0] // will use only one for now

	// TODO

	return toText("add-success", nil), nil
}

func remove(id int, uris []string) (string, error) {
	if len(uris) == 0 {
		return toText("remove-validation", nil), nil
	}

	trg := uris[0] // will use only one for now
	query := `DELETE FROM userfeeds
	WHERE user_id = $1 AND feed_id = (SELECT TOP(1) id FROM feeds WHERE uri = $2)`
	_, err := db.Connection.Exec(db.Context, query, id, trg)
	if err != nil {
		return "", err
	}

	return toText("remove-success", nil), nil
}

func list(id int) (string, error) {
	query := `SELECT f.name, f.uri FROM userfeeds uf
	INNER JOIN feeds f ON f.id = uf.feed_id
	WHERE uf.user_id = $1
	ORDER BY uf.added`

	rows, err := db.Connection.Query(db.Context, query, id)
	if err != nil {
		return "", err
	}

	var feeds []Feed
	err = rows.Scan(&feeds)
	if err != nil {
		return "", err
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
