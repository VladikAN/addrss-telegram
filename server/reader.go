package server

import (
	"time"

	log "github.com/go-pkgz/lgr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/vladikan/feedreader-telegrambot/database"
	"github.com/vladikan/feedreader-telegrambot/parser"
	"github.com/vladikan/feedreader-telegrambot/templates"
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
	duration := time.Duration(rd.Interval) * time.Second

	// Read feeds from db
	feeds, err := db.GetForUpdate(rd.Feeds)
	if err != nil {
		return err
	}

	// Read feeds from web
	for _, feed := range feeds {
		updates, err := parser.GetUpdates(feed.URI, *feed.Updated)
		if err != nil {
			return err
		}

		if len(updates) > 0 {
			users, err := db.GetFeedUsers(feed.ID)
			if err != nil {
				return err
			}

			rd.sendUpdates(updates, users)
		}

		err = db.SetFeedUpdated(feed.ID)
		if err != nil {
			return err
		}
	}

	log.Printf("INFO Reader job completed. %d feeds updated. Next call in %s", len(feeds), time.Now().Add(duration))
	return nil
}

func (rd *Reader) sendUpdates(updates []parser.Topic, users []database.UserFeed) {
	for _, upd := range updates {
		txt, _ := templates.ToTextW("topic", upd)

		for _, usr := range users {
			msg := tgbotapi.NewMessage(usr.UserID, txt)
			msg.ParseMode = "HTML"
			bot.Send(msg)
		}
	}
}
