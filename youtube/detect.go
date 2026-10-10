package youtube

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	channelIDRe = regexp.MustCompile(`^UC[\w-]{22}$`)
	handleRe    = regexp.MustCompile(`^@?[\w.-]{3,30}$`)
)

// Ref describes a YouTube channel reference extracted from a URL.
type Ref struct {
	Kind  string // id | handle | username | custom
	Value string
}

// IsYouTubeURL reports whether raw looks like a YouTube channel or channel feed URL.
func IsYouTubeURL(raw string) bool {
	u, err := parseURL(raw)
	if err != nil {
		return false
	}
	return isYouTubeHost(u.Host) && ParseRef(u) != nil
}

// IsYouTubeHostURL reports whether raw points at a YouTube host (any path).
func IsYouTubeHostURL(raw string) bool {
	u, err := parseURL(raw)
	if err != nil {
		return false
	}
	return isYouTubeHost(u.Host)
}

// ParseRef extracts a channel reference from a parsed YouTube URL.
func ParseRef(u *url.URL) *Ref {
	if u == nil || !isYouTubeHost(u.Host) {
		return nil
	}

	path := strings.Trim(u.Path, "/")
	parts := splitPath(path)

	// https://www.youtube.com/feeds/videos.xml?channel_id=UC...
	if path == "feeds/videos.xml" {
		id := u.Query().Get("channel_id")
		if channelIDRe.MatchString(id) {
			return &Ref{Kind: "id", Value: id}
		}
		return nil
	}

	if len(parts) == 0 {
		return nil
	}

	switch {
	case strings.HasPrefix(parts[0], "@"):
		handle := strings.TrimPrefix(parts[0], "@")
		if handleRe.MatchString(handle) {
			return &Ref{Kind: "handle", Value: handle}
		}
	case parts[0] == "channel" && len(parts) >= 2 && channelIDRe.MatchString(parts[1]):
		return &Ref{Kind: "id", Value: parts[1]}
	case parts[0] == "user" && len(parts) >= 2 && parts[1] != "":
		return &Ref{Kind: "username", Value: parts[1]}
	case parts[0] == "c" && len(parts) >= 2 && parts[1] != "":
		return &Ref{Kind: "custom", Value: parts[1]}
	}

	return nil
}

// ExtractChannelID returns a channel id when the URL already contains one.
func ExtractChannelID(raw string) string {
	u, err := parseURL(raw)
	if err != nil {
		return ""
	}
	ref := ParseRef(u)
	if ref != nil && ref.Kind == "id" {
		return ref.Value
	}
	return ""
}

// CanonicalURL returns the stable channel URL stored in the database.
func CanonicalURL(channelID string) string {
	return "https://www.youtube.com/channel/" + channelID
}

// AtomFeedURL returns the official YouTube Atom feed for a channel.
func AtomFeedURL(channelID string) string {
	return "https://www.youtube.com/feeds/videos.xml?channel_id=" + channelID
}

func parseURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("empty url")
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	return url.Parse(raw)
}

func isYouTubeHost(host string) bool {
	host = strings.ToLower(host)
	host = strings.TrimPrefix(host, "www.")
	host = strings.TrimPrefix(host, "m.")
	return host == "youtube.com" || host == "youtube-nocookie.com" || host == "youtu.be"
}

func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	var parts []string
	for _, p := range strings.Split(path, "/") {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}
