package parser

import (
	"github.com/mmcdole/gofeed"
)

// GetFeed parses uri with RSS/ATOM parser and returns feed name
func GetFeed(uri string) (string, error) {
	fp := gofeed.NewParser()

	feed, err := fp.ParseURL(uri)
	if err != nil {
		return "", err
	}

	return feed.Title, nil
}
