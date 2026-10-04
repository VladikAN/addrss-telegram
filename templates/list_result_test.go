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
		Name       string
		Normalized string
		Healthy    bool
		Blocked    bool
		Updated    *time.Time
		LastPub    *time.Time
	}

	feeds := []feedView{
		{Name: "Blocked Feed", Normalized: "blocked-feed", Healthy: true, Blocked: true, Updated: &now, LastPub: &now},
		{Name: "Healthy Feed", Normalized: "healthy-feed", Healthy: true, Blocked: false, Updated: &now, LastPub: &now},
		{Name: "Broken Feed", Normalized: "broken-feed", Healthy: false, Blocked: false, Updated: &now, LastPub: &now},
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
}
