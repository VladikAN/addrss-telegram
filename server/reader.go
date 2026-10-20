package server

import (
	"time"

	log "github.com/go-pkgz/lgr"

	"github.com/vladikan/addrss-telegram/database"
	"github.com/vladikan/addrss-telegram/metrics"
	"github.com/vladikan/addrss-telegram/parser"
	"github.com/vladikan/addrss-telegram/templates"
	"github.com/vladikan/addrss-telegram/youtube"
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
				log.Printf("ERROR reader fault: %s", err)
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
	started := time.Now()

	// Read feeds from db
	feeds, err := rd.DB.GetFeeds(rd.Feeds)
	if err != nil {
		return err
	}

	log.Printf("INFO Reader batch started. %d feeds", len(feeds))
	defer func() {
		elapsed := time.Since(started)
		metrics.ObserveBatch(elapsed.Seconds())
		log.Printf("INFO Reader batch completed. %d feeds in %s", len(feeds), elapsed.Round(time.Millisecond))
	}()

	stats := struct {
		updated    int
		notified   int
		feeds      int
		duplicates int
	}{}

	// Read feeds from servers
	for _, feed := range feeds {
		readStarted := time.Now()
		updates, err := parser.GetUpdates(feedReadURI(feed), *feed.LastPub)
		metrics.ObserveFeedRead(feed.SourceType, time.Since(readStarted).Seconds())
		if err != nil {
			metrics.FeedRead(metrics.ResultError, feed.SourceType)
			log.Printf("ERROR Feed '%s' unable get updates: %s", feed.Normalized, err)
			if markErr := rd.DB.SetFeedBroken(feed.ID); markErr != nil {
				log.Printf("ERROR Feed '%s' unable mark as broken: %s", feed.Normalized, markErr)
			}
			continue
		}

		if len(updates) > 0 {
			// Filter out already processed articles by URI
			var newUpdates []parser.Topic
			for _, update := range updates {
				// Skip if this is the same URI as the last processed post
				if feed.LastPubURI != "" && feed.LastPubURI == update.URI {
					stats.duplicates++
					continue
				}
				newUpdates = append(newUpdates, update)
			}

			if len(newUpdates) > 0 {
				metrics.FeedRead(metrics.ResultUpdated, feed.SourceType)
				metrics.PostsFound(feed.SourceType, len(newUpdates))

				users, err := rd.DB.GetFeedUsers(feed.ID)
				if err != nil {
					log.Printf("ERROR Feed '%s' unable get subscriptions: %s", feed.Normalized, err)
					continue
				}

				stats.updated += len(newUpdates)
				stats.notified += len(users)
				stats.feeds++
				if len(users) > 0 {
					rd.sendUpdates(newUpdates, users)
					metrics.NotificationsSent(feed.SourceType, len(newUpdates)*len(users))
				}

				// Update last publication date and URI to the latest processed article
				last := parser.GetLast(newUpdates)
				err = rd.DB.SetFeedLastPub(feed.ID, *last.Date, last.URI)
				if err != nil {
					log.Printf("ERROR Feed '%s' unable update last pub date and URI: %s", feed.Normalized, err)
					continue
				}
			} else {
				metrics.FeedRead(metrics.ResultOK, feed.SourceType)
			}

			continue
		}

		metrics.FeedRead(metrics.ResultOK, feed.SourceType)
		err = rd.DB.SetFeedUpdated(feed.ID)
		if err != nil {
			log.Printf("ERROR Feed '%s' unable mark as updated: %s", feed.Normalized, err)
			continue
		}
	}

	if stats.updated > 0 {
		log.Printf("DEBUG Reader found %d new post(s) for %d feed(s) and notified %d subscription(s) (skipped %d duplicates)", stats.updated, stats.feeds, stats.notified, stats.duplicates)
	}

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

func feedReadURI(feed database.Feed) string {
	if feed.SourceType == database.SourceYouTube {
		channelID := feed.ExternalID
		if channelID == "" {
			channelID = youtube.ExtractChannelID(feed.URI)
		}
		if channelID != "" {
			return youtube.AtomFeedURL(channelID)
		}
	}
	return feed.URI
}
