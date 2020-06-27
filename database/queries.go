package database

import (
	"time"

	"github.com/jackc/pgx/v4"
)

// Feed represents feed db table structure
type Feed struct {
	ID         int
	Name       string
	Normalized string
	URI        string
	Updated    *time.Time
	Healthy    bool
}

// UserFeed represents user subscription to the feed
type UserFeed struct {
	UserID int64
	FeedID int
	Added  *time.Time
}

// AddFeed inserts new feed to feeds postgres table
func (db *Database) AddFeed(name string, normalized string, uri string) (*Feed, error) {
	query := `INSERT INTO feeds (name, normalized, uri) VALUES ($1, $2, $3) ON CONFLICT (uri) DO NOTHING`
	_, err := db.Pool.Exec(db.Context, query, name, normalized, uri)
	if err != nil {
		return nil, err
	}

	return db.GetFeed(uri)
}

// Subscribe bind relation between user and feed
func (db *Database) Subscribe(userID int64, feedID int) error {
	query := `INSERT INTO userfeeds (user_id, feed_id) VALUES ($1, $2) ON CONFLICT (user_id, feed_id) DO NOTHING`
	_, err := db.Pool.Exec(db.Context, query, userID, feedID)
	return err
}

// Unsubscribe unbind relation between user and feed
func (db *Database) Unsubscribe(userID int64, feedID int) error {
	query := `DELETE FROM userfeeds WHERE user_id = $1 AND feed_id = $2`
	_, err := db.Pool.Exec(db.Context, query, userID, feedID)
	return err
}

// GetUserFeeds gets user subscriptions
func (db *Database) GetUserFeeds(userID int64) ([]Feed, error) {
	var feeds []Feed

	query := `SELECT f.id, f.name, f.normalized, f.uri, f.updated, f.healthy FROM userfeeds uf
	INNER JOIN feeds f ON f.id = uf.feed_id
	WHERE uf.user_id = $1
	ORDER BY uf.added`

	rows, err := db.Pool.Query(db.Context, query, userID)
	defer rows.Close()
	if err != nil {
		return feeds, err
	}

	return toFeeds(rows)
}

// GetUserURIFeed get user subscription by its uri (unique)
func (db *Database) GetUserURIFeed(userID int64, uri string) (*Feed, error) {
	query := `SELECT f.id, f.name, f.normalized, f.uri, f.updated, f.healthy FROM userfeeds uf
	INNER JOIN feeds f ON f.id = uf.feed_id
	WHERE uf.user_id = $1 AND f.uri = $2
	LIMIT 1`

	row := db.Pool.QueryRow(db.Context, query, userID, uri)
	return toFeed(row)
}

// GetUserNormalizedFeed get user subscription by its normalized name
func (db *Database) GetUserNormalizedFeed(userID int64, normalized string) (*Feed, error) {
	query := `SELECT f.id, f.name, f.normalized, f.uri, f.updated, f.healthy FROM userfeeds uf
	INNER JOIN feeds f ON f.id = uf.feed_id
	WHERE uf.user_id = $1 AND f.normalized = $2
	LIMIT 1`

	row := db.Pool.QueryRow(db.Context, query, userID, normalized)
	return toFeed(row)
}

// GetFeed get feed record by its uri (unique)
func (db *Database) GetFeed(uri string) (*Feed, error) {
	query := `SELECT id, name, normalized, uri, updated, healthy
	FROM feeds
	WHERE uri = $1
	LIMIT 1`

	row := db.Pool.QueryRow(db.Context, query, uri)
	return toFeed(row)
}

// GetForUpdate read specified count for update
func (db *Database) GetForUpdate(count int) ([]Feed, error) {
	var feeds []Feed

	// Get healthy or unhealthy for last day
	query := `SELECT id, name, normalized, uri, updated, healthy
	FROM feeds
	WHERE healthy = TRUE OR updated < current_date
	ORDER BY updated
	LIMIT $1`

	rows, err := db.Pool.Query(db.Context, query, count)
	defer rows.Close()
	if err != nil {
		return feeds, err
	}

	return toFeeds(rows)
}

// GetFeedUsers returns active feed subscriptions
func (db *Database) GetFeedUsers(feedID int) ([]UserFeed, error) {
	query := `SELECT user_id, added FROM userfeeds WHERE feed_id = $1`
	rows, err := db.Pool.Query(db.Context, query, &feedID)
	if err != nil {
		return nil, err
	}

	var subs []UserFeed
	for rows.Next() {
		item := UserFeed{FeedID: feedID}
		err = rows.Scan(&item.UserID, &item.Added)
		if err != nil {
			return subs, err
		}

		subs = append(subs, item)
	}

	return subs, nil
}

// SetFeedUpdated update feed by new timespan and set healthy to true
func (db *Database) SetFeedUpdated(id int) error {
	query := `UPDATE feeds
	SET updated = $1,
	healthy = TRUE
	WHERE id = $2`

	_, err := db.Pool.Exec(db.Context, query, time.Now(), id)
	return err
}

// SetFeedBroken update feed by setting healthy to false
func (db *Database) SetFeedBroken(id int) error {
	query := `UPDATE feeds
	SET updated = $1,
	healthy = FALSE
	WHERE id = $2`

	_, err := db.Pool.Exec(db.Context, query, time.Now(), id)
	return err
}

func toFeed(row pgx.Row) (*Feed, error) {
	var id int
	var name string
	var normalized string
	var uri string
	var updated *time.Time
	var healthy bool

	if err := row.Scan(&id, &name, &normalized, &uri, &updated, &healthy); err == nil {
		return &Feed{
			ID:         id,
			Name:       name,
			Normalized: normalized,
			URI:        uri,
			Updated:    updated,
			Healthy:    healthy,
		}, err
	} else if err == pgx.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return nil, nil
}

func toFeeds(rows pgx.Rows) ([]Feed, error) {
	var feeds []Feed
	for rows.Next() {
		feed, err := toFeed(rows)
		if err != nil {
			return feeds, err
		}

		feeds = append(feeds, *feed)
	}

	return feeds, nil
}
