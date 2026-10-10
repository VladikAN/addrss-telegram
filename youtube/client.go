package youtube

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const apiBase = "https://www.googleapis.com/youtube/v3/channels"

// Channel is a resolved YouTube channel.
type Channel struct {
	ID    string
	Title string
}

// Client calls YouTube Data API v3.
type Client struct {
	APIKey string
	HTTP   *http.Client
}

// NewClient creates a YouTube API client.
func NewClient(apiKey string) *Client {
	return &Client{
		APIKey: apiKey,
		HTTP:   &http.Client{Timeout: 15 * time.Second},
	}
}

// ResolveChannel resolves a YouTube channel URL into id + title via the Data API.
func (c *Client) ResolveChannel(raw string) (*Channel, error) {
	if c == nil || c.APIKey == "" {
		return nil, fmt.Errorf("youtube api key is not configured")
	}

	u, err := parseURL(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid youtube url: %w", err)
	}

	ref := ParseRef(u)
	if ref == nil {
		return nil, fmt.Errorf("unsupported youtube channel url")
	}

	switch ref.Kind {
	case "id":
		return c.lookup(url.Values{"id": {ref.Value}})
	case "handle":
		return c.lookup(url.Values{"forHandle": {ref.Value}})
	case "username":
		return c.lookup(url.Values{"forUsername": {ref.Value}})
	case "custom":
		ch, err := c.lookup(url.Values{"forHandle": {ref.Value}})
		if err == nil {
			return ch, nil
		}
		return c.lookup(url.Values{"forUsername": {ref.Value}})
	default:
		return nil, fmt.Errorf("unsupported youtube channel reference")
	}
}

type channelsResponse struct {
	Items []struct {
		ID      string `json:"id"`
		Snippet struct {
			Title string `json:"title"`
		} `json:"snippet"`
	} `json:"items"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) lookup(params url.Values) (*Channel, error) {
	params.Set("part", "snippet")
	params.Set("key", c.APIKey)
	params.Set("maxResults", "1")

	reqURL := apiBase + "?" + params.Encode()
	resp, err := c.HTTP.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("youtube api request failed: %w", err)
	}
	defer resp.Body.Close()

	var body channelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("youtube api decode failed: %w", err)
	}

	if body.Error != nil && body.Error.Message != "" {
		return nil, fmt.Errorf("youtube api error: %s", body.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("youtube api status %d", resp.StatusCode)
	}
	if len(body.Items) == 0 {
		return nil, fmt.Errorf("youtube channel not found")
	}

	item := body.Items[0]
	title := strings.TrimSpace(item.Snippet.Title)
	if title == "" {
		title = item.ID
	}
	return &Channel{ID: item.ID, Title: title}, nil
}
