-- Migration for existing deployments: track users who blocked the bot.
CREATE TABLE IF NOT EXISTS users(
    user_id BIGINT PRIMARY KEY,
    blocked BOOLEAN NOT NULL DEFAULT FALSE,
    blocked_at TIMESTAMPTZ
);

INSERT INTO users (user_id, blocked)
SELECT DISTINCT user_id, FALSE
FROM userfeeds
ON CONFLICT (user_id) DO NOTHING;
