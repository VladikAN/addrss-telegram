package parser

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const sampleRSS = `<?xml version="1.0"?>
<rss version="2.0">
  <channel>
    <title>Example</title>
    <item>
      <title>Post</title>
      <link>https://example.com/p</link>
      <pubDate>Mon, 02 Jan 2006 15:04:05 MST</pubDate>
    </item>
  </channel>
</rss>`

func TestGetUpdates_HTTPClientErrors(t *testing.T) {
	for _, code := range []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(code)
			_, _ = w.Write([]byte(sampleRSS))
		}))
		_, err := GetUpdates(srv.URL, time.Now().Add(-time.Hour))
		srv.Close()
		if err == nil {
			t.Errorf("status %d: expected error, got nil", code)
			continue
		}
		if !strings.Contains(err.Error(), "http status") {
			t.Errorf("status %d: expected http status error, got %v", code, err)
		}
	}
}

func TestGetUpdates_TLSError(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sampleRSS))
	}))
	defer srv.Close()

	_, err := GetUpdates(srv.URL, time.Now().Add(-time.Hour))
	if err == nil {
		t.Fatal("expected TLS error")
	}
}

func TestGetUpdates_DNSError(t *testing.T) {
	_, err := GetUpdates("https://this-host-should-not-resolve.invalid/rss", time.Now())
	if err == nil {
		t.Fatal("expected DNS error")
	}
}

func TestGetUpdates_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(sampleRSS))
	}))
	defer srv.Close()

	topics, err := GetUpdates(srv.URL, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(topics) != 1 || topics[0].Title != "Post" {
		t.Fatalf("unexpected topics: %+v", topics)
	}
}
