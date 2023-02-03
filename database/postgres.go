package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Postgres is concrete implementation for the PostgreSql
type Postgres struct {
	Connection string
	Pool       *pgxpool.Pool
	Context    context.Context

	cancel context.CancelFunc
}

// Database is an db operations proxy
type Database interface {
	// Close will termintae current connection, Should be called after all operations
	Close()

	// GetStats gets total number of users and feeds
	GetStats() (*Stats, error)

	// AddFeed inserts new feed to feeds postgres table
	AddFeed(name string, normalized string, uri string) (*Feed, error)

	// Subscribe bind relation between user and feed
	Subscribe(userID int64, feedID int) error

	// Unsubscribe unbind relation between user and feed
	Unsubscribe(userID int64, feedID int) error

	// DeleteUser will delete all user records
	DeleteUser(userID int64) error

	// GetUserFeeds gets user subscriptions
	GetUserFeeds(userID int64) ([]Feed, error)

	// GetUserURIFeed get user subscription by its uri (unique)
	GetUserURIFeed(userID int64, uri string) (*Feed, error)

	// GetUserNormalizedFeed get user subscription by its normalized name
	GetUserNormalizedFeed(userID int64, normalized string) (*Feed, error)

	// GetFeed get feed record by its uri (unique)
	GetFeed(uri string) (*Feed, error)

	// GetFeeds read specified count for update
	GetFeeds(count int) ([]Feed, error)

	// GetFeedUsers returns active feed subscriptions
	GetFeedUsers(feedID int) ([]UserFeed, error)

	// ResetFeed updates feed dates to prevent spam to first subscription after some time
	ResetFeed(feedID int) error

	// SetFeedUpdated update feed by new timespan and set healthy to true
	SetFeedUpdated(id int) error

	// SetFeedLastPub update feed by new timespan, set healthy to true and set last publication date
	SetFeedLastPub(id int, lastPub time.Time) error

	// SetFeedBroken update feed by setting healthy to false
	SetFeedBroken(id int) error
}

// Open will start database connection. Should be called first
func Open(ctx context.Context, connection string) (*Postgres, error) {
	pctx, cancel := context.WithCancel(ctx)
	pool, err := pgxpool.Connect(pctx, connection)
	if err != nil {
		cancel()
		return nil, err
	}

	return &Postgres{Pool: pool, Context: pctx, cancel: cancel}, nil
}

// Close will drop psql connections
func (db *Postgres) Close() {
	db.cancel()
	db.Pool.Close()
	db.Context = nil
}
