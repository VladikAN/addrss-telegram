package parser

import (
	"strings"
	"testing"
)

func TestDecode(t *testing.T) {
	xml := `<?xml version="1.0" encoding="utf-8"?>
	<opml version="1.0">
	<head><title>Test Opml file</title></head>
	<body>
		<outline type="rss" text="feed1" xmlUrl="feed1XMLUri" htmlUrl=""/>
		<outline type="rss" text="feed2" xmlUrl="feed2XMLUri" htmlUrl=""/>
		<outline type="prefs"/>
	</body></opml>`

	rst, err := decode(strings.NewReader(xml))
	if err != nil {
		t.Errorf("Error not expected, but was: %s", err)
	}

	if len(rst) != 2 {
		t.Errorf("Expected array of length 2, but was %d", len(rst))
	}

	if rst[0].Title != "feed1" || rst[0].URL != "feed1XMLUri" {
		t.Errorf("Expected 'feed1-feed1XMLUri', but was '%s-%s'", rst[0].Title, rst[0].URL)
	}

	if rst[1].Title != "feed2" || rst[1].URL != "feed2XMLUri" {
		t.Errorf("Expected 'feed2-feed2XMLUri', but was '%s-%s'", rst[1].Title, rst[1].URL)
	}
}

func TestDecode_NestedAndYouTube(t *testing.T) {
	xml := `<?xml version="1.0" encoding="utf-8"?>
	<opml version="1.0">
	<body>
		<outline text="Folder">
			<outline type="rss" text="Blog" xmlUrl="https://example.com/rss"/>
			<outline text="Channel" title="Channel Title" xmlUrl="https://www.youtube.com/feeds/videos.xml?channel_id=UCHnyfMqiRRG1u-2MsSQLbXA"/>
		</outline>
	</body></opml>`

	rst, err := decode(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(rst) != 2 {
		t.Fatalf("expected 2 items, got %d", len(rst))
	}
	if rst[0].URL != "https://example.com/rss" {
		t.Errorf("unexpected rss url: %s", rst[0].URL)
	}
	if !strings.Contains(rst[1].URL, "channel_id=") {
		t.Errorf("expected youtube feed url, got %s", rst[1].URL)
	}
	if rst[1].Title != "Channel" {
		t.Errorf("expected title Channel, got %q", rst[1].Title)
	}
}
