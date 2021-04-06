package server

import (
	"fmt"
	"time"

	log "github.com/go-pkgz/lgr"

	"github.com/vladikan/addrss-telegram/database"
	"github.com/vladikan/addrss-telegram/parser"
	"github.com/vladikan/addrss-telegram/templates"
)

// Reader holds reader settings
type Reader struct {
	Interval int
	Feeds    int
	DB       database.Database
	Outbox   chan Reply

	stop chan interface{}
}

// Start will look for feed updates
func (rd *Reader) Start() {
	rd.stop = make(chan interface{})

	duration := time.Duration(rd.Interval) * time.Second
	tick := time.NewTicker(duration)

	go func() {
		read := func() {
			err := rd.readFeeds()
			if err != nil {
				log.Printf("ERROR Reader job completed with error: %s", err)
			}
		}

		read() // Force first read on start

		for {
			select {
			case <-rd.stop:
				tick.Stop()
				return
			case <-tick.C:
				read()
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
	log.Printf("DEBUG Reader job started. %d feeds to be readed", rd.Feeds)
	duration := time.Duration(rd.Interval) * time.Second

	// Read feeds from db
	feeds, err := rd.DB.GetForUpdate(rd.Feeds)
	if err != nil {
		return err
	}

	stats := struct {
		updated  int
		notified int
		feeds    int
	}{}

	// Read feeds from servers
	for _, feed := range feeds {
		updates, err := parser.GetUpdates(feed.URI, *feed.LastPub)
		if err != nil {
			rd.DB.SetFeedBroken(feed.ID)
			return fmt.Errorf("Feed '%s' unable to get updates: %s", feed.Normalized, err)
		}

		if len(updates) > 0 {
			users, err := rd.DB.GetFeedUsers(feed.ID)
			if err != nil {
				return fmt.Errorf("Feed '%s' unable to get subscriptions: %s", feed.Normalized, err)
			}

			stats.updated += len(updates)
			stats.notified += len(users)
			stats.feeds++
			if len(users) > 0 {
				rd.sendUpdates(updates, users)
			}

			last := parser.GetLast(updates)
			err = rd.DB.SetFeedLastPub(feed.ID, *last.Date)
			if err != nil {
				return fmt.Errorf("Feed '%s' unable to mark as updated with last publish date: %s", feed.Normalized, err)
			}

			continue
		}

		err = rd.DB.SetFeedUpdated(feed.ID)
		if err != nil {
			return fmt.Errorf("Feed '%s' unable to mark as updated: %s", feed.Normalized, err)
		}
	}

	if stats.updated > 0 {
		log.Printf("INFO Reader found %d new post(s) for %d feed(s) and notified %d subscription(s)", stats.updated, stats.feeds, stats.notified)
	}

	log.Printf("DEBUG Reader job completed. %d feeds updated. Next call in %s", len(feeds), time.Now().Add(duration))
	return nil
}

func (rd *Reader) sendUpdates(updates []parser.Topic, users []database.UserFeed) {
	for _, upd := range updates {
		txt, _ := templates.ToTextW("en", "topic", upd)

		for _, usr := range users {
			rd.Outbox <- Reply{ChatID: usr.UserID, Text: txt}
		}
	}
}
