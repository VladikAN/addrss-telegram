package parser

import (
	"sort"
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
		return "", err
	}

	return feed.Title, nil
}

// GetUpdates load artiales since specified date
func GetUpdates(uri string, since time.Time) ([]Topic, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(uri)
	if err != nil {
		return nil, err
	}

	sort.Sort(feed) // sort from old to new

	var topics []Topic
	for _, item := range feed.Items {
		if item.UpdatedParsed.Before(since) {
			continue
		}

		text := html2text.HTML2Text(item.Content)
		topic := Topic{
			Title: item.Title,
			Text:  text,
			URI:   item.Link,
		}
		topics = append(topics, topic)
	}

	return topics, nil
}
