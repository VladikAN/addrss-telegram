package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestListResult_ShowsBlackCircleForBlocked(t *testing.T) {
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolve project root: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir to project root: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(filepath.Join(root, "templates"))
	})

	SetTemplateOutput()

	now := time.Now()
	type feedView struct {
		Name        string
		Normalized  string
		SourceLabel string
		Healthy     bool
		Blocked     bool
		Updated     *time.Time
		LastPub     *time.Time
	}

	feeds := []feedView{
		{Name: "Blocked Feed", Normalized: "blocked-feed", SourceLabel: "RSS", Healthy: true, Blocked: true, Updated: &now, LastPub: &now},
		{Name: "Healthy Feed", Normalized: "healthy-feed", SourceLabel: "RSS", Healthy: true, Blocked: false, Updated: &now, LastPub: &now},
		{Name: "Broken Feed", Normalized: "broken-feed", SourceLabel: "YT", Healthy: false, Blocked: false, Updated: &now, LastPub: &now},
	}

	text, err := ToTextW("en", "list-result", feeds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(text, "⚫") {
		t.Error("expected black circle for blocked feed")
	}
	if !strings.Contains(text, "🟢") {
		t.Error("expected green circle for healthy feed")
	}
	if !strings.Contains(text, "🔴") {
		t.Error("expected red circle for unhealthy feed")
	}
	if !strings.Contains(text, "[RSS]") {
		t.Error("expected [RSS] source label")
	}
	if !strings.Contains(text, "[YT]") {
		t.Error("expected [YT] source label")
	}
}

func TestStatsSuccess_IncludesBlockedCounts(t *testing.T) {
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolve project root: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir to project root: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(filepath.Join(root, "templates"))
	})

	SetTemplateOutput()

	stats := struct {
		Users        int
		Blocked      int
		Feeds        int
		BlockedFeeds int
	}{Users: 5, Blocked: 1, Feeds: 10, BlockedFeeds: 2}

	en, err := ToTextW("en", "stats-success", stats)
	if err != nil {
		t.Fatalf("en template error: %v", err)
	}
	if !strings.Contains(en, "1 - Blocked users.") {
		t.Errorf("expected blocked users in en stats, got: %s", en)
	}
	if !strings.Contains(en, "2 - Blocked feeds.") {
		t.Errorf("expected blocked feeds in en stats, got: %s", en)
	}

	ru, err := ToTextW("ru", "stats-success", stats)
	if err != nil {
		t.Fatalf("ru template error: %v", err)
	}
	if !strings.Contains(ru, "1 - Заблокировали бота.") {
		t.Errorf("expected blocked users in ru stats, got: %s", ru)
	}
	if !strings.Contains(ru, "2 - Заблокированных лент.") {
		t.Errorf("expected blocked feeds in ru stats, got: %s", ru)
	}
}
