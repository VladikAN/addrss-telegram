package parser

import (
	"fmt"
	"io"
	"net/http"

	"github.com/mmcdole/gofeed"
)

const feedUserAgent = "Gofeed/1.0"

// fetchFeed downloads a feed and parses it.
// DNS failures, TLS failures, and any non-2xx HTTP status (including 400, 401, and 403)
// are returned as errors so the reader can mark the feed unhealthy.
func fetchFeed(uri string) (*gofeed.Feed, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", feedUserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("http status %d", resp.StatusCode)
	}

	feed, err := gofeed.NewParser().Parse(resp.Body)
	if err != nil {
		return nil, err
	}
	if feed == nil {
		return nil, fmt.Errorf("empty feed")
	}
	return feed, nil
}
