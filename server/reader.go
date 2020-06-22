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
				rd.readFeeds()
				return
			}
		}
	}()
}

// Stop stops all reader activities
func (rd *Reader) Stop() {
	log.Print("INFO Reader jobs terminated")
	close(rd.stop)
}

func (rd *Reader) readFeeds() {
	duration := time.Duration(rd.Interval) * time.Second
	log.Printf("INFO Reader job started. %d feeds to be readed", rd.Feeds)

	log.Printf("INFO Reader job completed. Next call in %s", time.Now().Add(duration))
}
