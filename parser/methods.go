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
	Date  *time.Time
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
		date := getDate(item)
		if date == nil || date.Before(since) {
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
			Date:  date,
		}
		topics = append(topics, topic)
	}

	return topics, nil
}

// GetLast returns topic with latest publish date
func GetLast(topics []Topic) *Topic {
	if len(topics) == 0 {
		return nil
	}

	max := &topics[0]
	if len(topics) > 1 {
		for _, topic := range topics {
			if topic.Date.After(*max.Date) {
				max = &topic
			}
		}
	}

	return max
}

func getDate(item *gofeed.Item) *time.Time {
	tm := item.PublishedParsed
	if tm == nil {
		tm = item.UpdatedParsed // not all feeds has publish/update values, will ignore these feeds for now
	}

	return tm
}
