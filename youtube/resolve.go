package youtube

import (
	"fmt"
	"strings"
)

// Resolve turns a YouTube channel URL into a Channel.
// If the URL already contains a channel id and titleHint is set, client may be nil.
func Resolve(raw string, titleHint string, client *Client) (*Channel, error) {
	u, err := parseURL(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid youtube url: %w", err)
	}

	ref := ParseRef(u)
	if ref == nil {
		return nil, fmt.Errorf("unsupported youtube channel url")
	}

	titleHint = strings.TrimSpace(titleHint)
	if ref.Kind == "id" && titleHint != "" {
		return &Channel{ID: ref.Value, Title: titleHint}, nil
	}

	if client == nil || client.APIKey == "" {
		return nil, fmt.Errorf("youtube api key is not configured")
	}

	return client.ResolveChannel(raw)
}
