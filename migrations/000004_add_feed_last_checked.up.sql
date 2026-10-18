-- Last moderation check time. NULL means the feed has never been blocked/unblocked.
ALTER TABLE feeds
	ADD COLUMN IF NOT EXISTS last_checked TIMESTAMPTZ;
