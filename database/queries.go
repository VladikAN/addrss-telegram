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
	UserID int
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
func (db *Database) Subscribe(userID int, feedID int) error {
	query := `INSERT INTO userfeeds (user_id, feed_id) VALUES ($1, $2) ON CONFLICT (user_id, feed_id) DO NOTHING`
	_, err := db.Pool.Exec(db.Context, query, userID, feedID)
	return err
}

// Unsubscribe unbind relation between user and feed
func (db *Database) Unsubscribe(userID int, feedID int) error {
	query := `DELETE FROM userfeeds WHERE user_id = $1 AND feed_id = $2`
	_, err := db.Pool.Exec(db.Context, query, userID, feedID)
	return err
}

// GetUserFeeds gets user subscriptions
func (db *Database) GetUserFeeds(userID int) ([]Feed, error) {
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

	for rows.Next() {
		feed, err := toFeed(rows)
		if err != nil {
			return feeds, err
		}

		feeds = append(feeds, *feed)
	}

	return feeds, nil
}

// GetUserURIFeed get user subscription by its uri (unique)
func (db *Database) GetUserURIFeed(userID int, uri string) (*Feed, error) {
	query := `SELECT f.id, f.name, f.normalized, f.uri, f.updated, f.healthy FROM userfeeds uf
	INNER JOIN feeds f ON f.id = uf.feed_id
	WHERE uf.user_id = $1 AND f.uri = $2
	LIMIT 1`

	rows, err := db.Pool.Query(db.Context, query, userID, uri)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	return toFeed(rows)
}

// GetUserNormalizedFeed get user subscription by its normalized name
func (db *Database) GetUserNormalizedFeed(userID int, normalized string) (*Feed, error) {
	query := `SELECT f.id, f.name, f.normalized, f.uri, f.updated, f.healthy FROM userfeeds uf
	INNER JOIN feeds f ON f.id = uf.feed_id
	WHERE uf.user_id = $1 AND f.normalized = $2
	LIMIT 1`

	rows, err := db.Pool.Query(db.Context, query, userID, normalized)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	if rows.Next() {
		return toFeed(rows)
	}

	return nil, nil
}

// GetFeed get feed record by its uri (unique)
func (db *Database) GetFeed(uri string) (*Feed, error) {
	query := `SELECT id, name, normalized, uri, updated, healthy
	FROM feeds
	WHERE uri = $1
	LIMIT 1`

	rows, err := db.Pool.Query(db.Context, query, uri)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	return toFeed(rows)
}

func toFeed(rows pgx.Rows) (*Feed, error) {
	var id int
	var name string
	var normalized string
	var uri string
	var updated *time.Time
	var healthy bool

	if err := rows.Scan(&id, &name, &normalized, &uri, &updated, &healthy); err == nil {
		return &Feed{
			ID:         id,
			Name:       name,
			Normalized: normalized,
			URI:        uri,
			Updated:    updated,
			Healthy:    healthy,
		}, err
	} else if err != nil {
		return nil, err
	}

	return nil, nil
}
