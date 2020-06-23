package server

import (
	"time"

	log "github.com/go-pkgz/lgr"
	"github.com/vladikan/feedreader-telegrambot/database"
)

// Reader holds reader settings
type Reader struct {
	Interval int
	Feeds    int
	DB       *database.Database

	stop chan interface{}
}

type feed struct {
	id  int
	uri string
}

// Start will look for feed updates
func (rd *Reader) Start() {
	rd.stop = make(chan interface{})

	duration := time.Duration(rd.Interval) * time.Second
	tick := time.NewTicker(duration)

	go func() {
		for {
			select {
			case <-rd.stop:
				tick.Stop()
				return
			case <-tick.C:
				err := rd.readFeeds()
				if err != nil {
					log.Printf("ERROR Reader job completed with error: %s", err)
				}
			}
		}
	}()
}

// Stop stops all reader activities
func (rd *Reader) Stop() {
	log.Print("INFO Reader jobs terminated")
	close(rd.stop)
}

func (rd *Reader) readFeeds() error {
	log.Printf("INFO Reader job started. %d feeds to be readed", rd.Feeds)

	now := time.Now()
	duration := time.Duration(rd.Interval) * time.Second

	// Read feeds from db
	query := `SELECT id, uri FROM feeds WHERE healthy = TRUE ORDER BY updated LIMIT $1`
	rows, err := rd.DB.Pool.Query(db.Context, query, rd.Feeds)
	if err != nil {
		return err
	}

	// Map db records
	var feeds []feed
	for rows.Next() {
		var id int
		var uri string
		err = rows.Scan(&id, &uri)
		if err != nil {
			return err
		}

		feeds = append(feeds, feed{id: id, uri: uri})
	}

	// Read feeds from web
	for _, feed := range feeds {

		query = `UPDATE feeds SET updated = $1 WHERE id = $2`
		_, err = rd.DB.Pool.Exec(db.Context, query, &now, &feed.id)
		if err != nil {
			return err
		}
	}

	log.Printf("INFO Reader job completed. %d feeds updated. Next call in %s", len(feeds), time.Now().Add(duration))
	return nil
}
