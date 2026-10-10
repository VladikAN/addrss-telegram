-- Source type for RSS vs YouTube subscriptions
ALTER TABLE feeds
	ADD COLUMN IF NOT EXISTS source_type VARCHAR(16) NOT NULL DEFAULT 'rss';

ALTER TABLE feeds
	ADD COLUMN IF NOT EXISTS external_id VARCHAR(128) NOT NULL DEFAULT '';
