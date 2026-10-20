package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	SourceRSS     = "rss"
	SourceYouTube = "youtube"

	ResultOK      = "ok"
	ResultUpdated = "updated"
	ResultError   = "error"
)

var feedReadBuckets = []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10, 30}

var batchDurationBuckets = []float64{1, 5, 15, 30, 60, 120, 300, 600, 1200}

var (
	feedReads = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "addrss_feed_reads_total",
		Help: "Feed read attempts by the reader.",
	}, []string{"result", "source"})

	feedReadDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "addrss_feed_read_duration_seconds",
		Help:    "Time to download and parse one feed.",
		Buckets: feedReadBuckets,
	}, []string{"source"})

	batchDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "addrss_reader_batch_duration_seconds",
		Help:    "Wall time of one reader batch, from feed selection through the last feed.",
		Buckets: batchDurationBuckets,
	})

	usersCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "addrss_users_created_total",
		Help: "Users inserted into the users table.",
	})

	feedsCreated = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "addrss_feeds_created_total",
		Help: "Feeds inserted into the catalog.",
	}, []string{"source"})

	subscriptionsRemoved = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "addrss_subscriptions_removed_total",
		Help: "User subscriptions removed.",
	}, []string{"source"})

	postsFound = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "addrss_posts_found_total",
		Help: "New posts that passed the duplicate URI filter.",
	}, []string{"source"})

	notificationsSent = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "addrss_notifications_sent_total",
		Help: "Reader notifications enqueued for delivery.",
	}, []string{"source"})
)

// Handler serves the Prometheus text format on /metrics.
func Handler() http.Handler {
	return promhttp.Handler()
}

// ObserveFeedRead records one feed fetch and parse, including failed attempts.
func ObserveFeedRead(source string, seconds float64) {
	feedReadDuration.WithLabelValues(normalizeSource(source)).Observe(seconds)
}

// FeedRead records the outcome of one feed read.
func FeedRead(result, source string) {
	feedReads.WithLabelValues(result, normalizeSource(source)).Inc()
}

// ObserveBatch records how long one reader batch took.
func ObserveBatch(seconds float64) {
	batchDuration.Observe(seconds)
}

// UserCreated records a newly inserted user.
func UserCreated() {
	usersCreated.Inc()
}

// FeedCreated records a newly inserted catalog feed.
func FeedCreated(source string) {
	feedsCreated.WithLabelValues(normalizeSource(source)).Inc()
}

// SubscriptionRemoved records a removed user-to-feed subscription.
func SubscriptionRemoved(source string) {
	subscriptionsRemoved.WithLabelValues(normalizeSource(source)).Inc()
}

// PostsFound adds new posts discovered in a feed.
func PostsFound(source string, n int) {
	if n <= 0 {
		return
	}
	postsFound.WithLabelValues(normalizeSource(source)).Add(float64(n))
}

// NotificationsSent adds messages the reader placed on the outbox.
func NotificationsSent(source string, n int) {
	if n <= 0 {
		return
	}
	notificationsSent.WithLabelValues(normalizeSource(source)).Add(float64(n))
}

func normalizeSource(source string) string {
	if source == SourceYouTube {
		return SourceYouTube
	}
	return SourceRSS
}
