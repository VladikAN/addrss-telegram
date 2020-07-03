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
