package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/vladikan/addrss-telegram/database"
)

const metricsRSS = `<?xml version="1.0"?>
<rss version="2.0">
  <channel>
    <title>Example</title>
    <item>
      <title>One</title>
      <link>https://example.com/p1</link>
      <pubDate>Thu, 02 Jan 2020 15:04:05 MST</pubDate>
    </item>
    <item>
      <title>Two</title>
      <link>https://example.com/p2</link>
      <pubDate>Fri, 03 Jan 2020 15:04:05 MST</pubDate>
    </item>
  </channel>
</rss>`

func TestReadFeeds_RecordsMetrics(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(metricsRSS))
	}))
	defer srv.Close()

	lastPub := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	outbox := make(chan Reply, 8)
	rd := &Reader{
		DB: &dbMock{
			getFeedsMock: func() ([]database.Feed, error) {
				return []database.Feed{{
					ID:         7,
					Normalized: "example",
					URI:        srv.URL,
					SourceType: database.SourceRSS,
					LastPub:    &lastPub,
				}}, nil
			},
			getFeedUsersMock: func() ([]database.UserFeed, error) {
				return []database.UserFeed{{UserID: 1}, {UserID: 2}}, nil
			},
			setFeedLastPubMock: func() error { return nil },
		},
		Outbox: outbox,
	}

	beforeReads := metricValue(t, "addrss_feed_reads_total", map[string]string{"result": "updated", "source": "rss"})
	beforePosts := metricValue(t, "addrss_posts_found_total", map[string]string{"source": "rss"})
	beforeNotes := metricValue(t, "addrss_notifications_sent_total", map[string]string{"source": "rss"})
	beforeBatch := metricValue(t, "addrss_reader_batch_duration_seconds", nil)

	if err := rd.readFeeds(); err != nil {
		t.Fatal(err)
	}

	if got := metricValue(t, "addrss_feed_reads_total", map[string]string{"result": "updated", "source": "rss"}); got-beforeReads != 1 {
		t.Fatalf("feed reads delta = %v, want 1", got-beforeReads)
	}
	if got := metricValue(t, "addrss_posts_found_total", map[string]string{"source": "rss"}); got-beforePosts != 2 {
		t.Fatalf("posts delta = %v, want 2", got-beforePosts)
	}
	if got := metricValue(t, "addrss_notifications_sent_total", map[string]string{"source": "rss"}); got-beforeNotes != 4 {
		t.Fatalf("notifications delta = %v, want 4", got-beforeNotes)
	}
	if got := metricValue(t, "addrss_reader_batch_duration_seconds", nil); got-beforeBatch != 1 {
		t.Fatalf("batch observations delta = %v, want 1", got-beforeBatch)
	}
	if len(outbox) != 4 {
		t.Fatalf("outbox len = %d, want 4", len(outbox))
	}
}

func TestRemove_RecordsSubscriptionRemoved(t *testing.T) {
	templatesReady(t)
	db = &dbMock{
		getUserNormalizedFeedMock: func() (*database.Feed, error) {
			return &database.Feed{ID: 3, SourceType: database.SourceYouTube}, nil
		},
		unsubscribeMock: func() error { return nil },
	}

	before := metricValue(t, "addrss_subscriptions_removed_total", map[string]string{"source": "youtube"})
	if _, err := (&Command{args: "name", lang: "en"}).remove(); err != nil {
		t.Fatal(err)
	}
	if got := metricValue(t, "addrss_subscriptions_removed_total", map[string]string{"source": "youtube"}); got-before != 1 {
		t.Fatalf("subscriptions removed delta = %v, want 1", got-before)
	}
}

func templatesReady(t *testing.T) {
	t.Helper()
	setup()
}

func metricValue(t *testing.T, name string, labels map[string]string) float64 {
	t.Helper()
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range families {
		if family.GetName() != name {
			continue
		}
		for _, metric := range family.GetMetric() {
			if !labelsMatch(metric.GetLabel(), labels) {
				continue
			}
			if counter := metric.GetCounter(); counter != nil {
				return counter.GetValue()
			}
			if histogram := metric.GetHistogram(); histogram != nil {
				return float64(histogram.GetSampleCount())
			}
		}
	}
	return 0
}

func labelsMatch(got []*dto.LabelPair, want map[string]string) bool {
	if len(want) == 0 {
		return len(got) == 0
	}
	if len(got) != len(want) {
		return false
	}
	for _, label := range got {
		if want[label.GetName()] != label.GetValue() {
			return false
		}
	}
	return true
}
