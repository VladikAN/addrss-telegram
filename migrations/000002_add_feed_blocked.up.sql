-- Add blocked flag for feed moderation / validation API
ALTER TABLE feeds
	ADD COLUMN IF NOT EXISTS blocked BOOLEAN NOT NULL DEFAULT FALSE;
