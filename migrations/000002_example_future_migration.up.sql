-- Example migration (commented out)
-- This file demonstrates how to create future migrations
-- Uncomment and modify when needed

-- Example 1: Adding a new column to an existing table
-- ALTER TABLE feeds ADD COLUMN IF NOT EXISTS description TEXT DEFAULT '';

-- Example 2: Creating a new table
-- CREATE TABLE IF NOT EXISTS user_preferences(
--     user_id BIGINT PRIMARY KEY,
--     language VARCHAR(10) DEFAULT 'en',
--     timezone VARCHAR(50) DEFAULT 'UTC',
--     notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE,
--     created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
--     
--     CONSTRAINT user_preferences_user_fk FOREIGN KEY (user_id)
--         REFERENCES users (user_id) MATCH SIMPLE
--         ON UPDATE NO ACTION ON DELETE CASCADE
-- );

-- Example 3: Creating an index
-- CREATE INDEX IF NOT EXISTS idx_feeds_updated ON feeds(updated);

-- Example 4: Adding a unique constraint
-- ALTER TABLE feeds ADD CONSTRAINT IF NOT EXISTS feeds_normalized_key UNIQUE (normalized);
