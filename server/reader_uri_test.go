package server

import (
	"testing"

	"github.com/vladikan/addrss-telegram/database"
)

func TestFeedReadURI_YouTube(t *testing.T) {
	feed := database.Feed{
		SourceType: database.SourceYouTube,
		ExternalID: "UCHnyfMqiRRG1u-2MsSQLbXA",
		URI:        "https://www.youtube.com/channel/UCHnyfMqiRRG1u-2MsSQLbXA",
	}
	got := feedReadURI(feed)
	want := "https://www.youtube.com/feeds/videos.xml?channel_id=UCHnyfMqiRRG1u-2MsSQLbXA"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFeedReadURI_RSS(t *testing.T) {
	feed := database.Feed{
		SourceType: database.SourceRSS,
		URI:        "https://example.com/rss",
	}
	if feedReadURI(feed) != feed.URI {
		t.Fatalf("rss uri should be unchanged")
	}
}
