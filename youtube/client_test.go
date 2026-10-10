package youtube

import "testing"

func TestResolve_WithChannelIDAndTitle_NoAPI(t *testing.T) {
	ch, err := Resolve("https://www.youtube.com/feeds/videos.xml?channel_id=UCHnyfMqiRRG1u-2MsSQLbXA", "Veritasium", nil)
	if err != nil {
		t.Fatalf("resolve without api: %v", err)
	}
	if ch.ID != "UCHnyfMqiRRG1u-2MsSQLbXA" || ch.Title != "Veritasium" {
		t.Fatalf("unexpected channel: %+v", ch)
	}
}

func TestResolve_RequiresAPIKeyForHandle(t *testing.T) {
	_, err := Resolve("https://www.youtube.com/@veritasium", "", nil)
	if err == nil || err.Error() != "youtube api key is not configured" {
		t.Fatalf("expected missing key error, got %v", err)
	}
}
