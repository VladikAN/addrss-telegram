package parser

import (
	"fmt"
	"time"

	"github.com/k3a/html2text"
	"github.com/mmcdole/gofeed"
)

// Topic is a lightweight representation of the parsed article
type Topic struct {
	Title string
	Text  string
	URI   string
}

// GetTitle parses uri with RSS/ATOM parser and returns feed name
func GetTitle(uri string) (string, error) {
	fp := gofeed.NewParser()

	feed, err := fp.ParseURL(uri)
	if err != nil {
		return "", fmt.Errorf("Unable to read '%s': %s", uri, err)
	}

	return feed.Title, nil
}

// GetUpdates load artiales since specified date
func GetUpdates(uri string, since time.Time) ([]Topic, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(uri)
	if err != nil {
		return nil, fmt.Errorf("Unable to read '%s': %s", uri, err)
	}

	if feed == nil {
		return nil, nil
	}

	var topics []Topic
	for _, item := range feed.Items {
		tm := item.PublishedParsed
		if tm == nil {
			tm = item.UpdatedParsed // not all feeds has publish/update values, will ignore these feeds for now
		}

		if tm == nil || tm.Before(since) {
			continue
		}

		text := html2text.HTML2Text(item.Description)
		if len(text) > 512 {
			text = text[:512] + "..."
		}

		topic := Topic{
			Title: item.Title,
			Text:  text,
			URI:   item.Link,
		}
		topics = append(topics, topic)
	}

	return topics, nil
}
