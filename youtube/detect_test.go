package youtube

import "testing"

func TestIsYouTubeURL(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"https://www.youtube.com/@veritasium", true},
		{"https://www.youtube.com/@полигон98", true},
		{"https://www.youtube.com/@%D0%BF%D0%BE%D0%BB%D0%B8%D0%B3%D0%BE%D0%BD98", true},
		{"https://youtube.com/channel/UCHnyfMqiRRG1u-2MsSQLbXA", true},
		{"https://www.youtube.com/c/Veritasium", true},
		{"https://www.youtube.com/user/1veritasium", true},
		{"https://www.youtube.com/feeds/videos.xml?channel_id=UCHnyfMqiRRG1u-2MsSQLbXA", true},
		{"https://m.youtube.com/@veritasium", true},
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", false},
		{"https://youtu.be/dQw4w9WgXcQ", false},
		{"https://example.com/feed.rss", false},
		{"not a url", false},
	}

	for _, tc := range cases {
		if got := IsYouTubeURL(tc.in); got != tc.want {
			t.Errorf("IsYouTubeURL(%q)=%v want %v", tc.in, got, tc.want)
		}
	}
}

func TestParseRef_CyrillicHandle(t *testing.T) {
	u, err := parseURL("https://www.youtube.com/@полигон98")
	if err != nil {
		t.Fatal(err)
	}
	ref := ParseRef(u)
	if ref == nil || ref.Kind != "handle" || ref.Value != "полигон98" {
		t.Fatalf("unexpected ref: %+v", ref)
	}
}

func TestIsYouTubeHostURL(t *testing.T) {
	if !IsYouTubeHostURL("https://www.youtube.com/watch?v=dQw4w9WgXcQ") {
		t.Fatal("expected youtube host url")
	}
	if IsYouTubeHostURL("https://example.com/feed.rss") {
		t.Fatal("expected non-youtube url")
	}
}

func TestExtractChannelID(t *testing.T) {
	id := "UCHnyfMqiRRG1u-2MsSQLbXA"
	got := ExtractChannelID("https://www.youtube.com/feeds/videos.xml?channel_id=" + id)
	if got != id {
		t.Fatalf("got %q want %q", got, id)
	}
	if ExtractChannelID("https://www.youtube.com/@veritasium") != "" {
		t.Fatal("expected empty id for handle url")
	}
}

func TestCanonicalAndAtom(t *testing.T) {
	id := "UCHnyfMqiRRG1u-2MsSQLbXA"
	if CanonicalURL(id) != "https://www.youtube.com/channel/"+id {
		t.Fatal("bad canonical url")
	}
	if AtomFeedURL(id) != "https://www.youtube.com/feeds/videos.xml?channel_id="+id {
		t.Fatal("bad atom url")
	}
}
