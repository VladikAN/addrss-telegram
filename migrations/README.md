# Database Migrations

This directory contains database migration files for the AddRss Telegram bot.

## Overview

Migrations are managed using [golang-migrate/migrate](https://github.com/golang-migrate/migrate) library and are automatically applied on application startup.

## Migration Files

Migration files follow the naming convention: `{version}_{description}.up.sql`

- `version`: Sequential number with leading zeros (e.g., 000001, 000002)
- `description`: Brief description of the migration (lowercase with underscores)
- `.up.sql`: Migration to apply changes

**Note:** Down migrations are not currently used in this project.

## Safety

All migrations should be idempotent and safe to run multiple times. Use SQL constructs like:

- `CREATE TABLE IF NOT EXISTS`
- `ALTER TABLE ... ADD COLUMN IF NOT EXISTS` (PostgreSQL 9.6+)
- `CREATE INDEX IF NOT EXISTS`

## Creating New Migrations

1. Create a new file with the next sequential version number:
   ```
   migrations/000002_add_user_preferences.up.sql
   ```

2. Write your migration with safe SQL:
   ```sql
   -- Add user preferences table
   CREATE TABLE IF NOT EXISTS user_preferences(
       user_id BIGINT PRIMARY KEY,
       language VARCHAR(10) DEFAULT 'en',
       timezone VARCHAR(50) DEFAULT 'UTC',
       created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
       
       CONSTRAINT user_preferences_user_fk FOREIGN KEY (user_id)
           REFERENCES users (user_id) MATCH SIMPLE
           ON UPDATE NO ACTION ON DELETE CASCADE
   );
   ```

3. Test the migration:
   ```bash
   go run . --local --debug
   ```

4. Verify it's safe to run multiple times by restarting the application.

## How It Works

1. On startup, `main.go` calls `database.RunMigrations()`
2. Migration files are embedded into the binary using Go's `embed` package
3. The migrate library creates a `schema_migrations` table to track applied migrations
4. Only new migrations (not yet applied) are executed
5. If migrations fail, the error is logged but the application continues

## Manual Migration Check

To check which migrations have been applied:

```sql
SELECT * FROM schema_migrations;
```

To view the current schema version:

```sql
SELECT version, dirty FROM schema_migrations;
```

## Troubleshooting

If a migration fails:

1. Check the application logs for the error message
2. Connect to the database and check the `schema_migrations` table
3. If `dirty = true`, the migration failed mid-execution
4. Fix the migration SQL and restart the application

## Best Practices

- Always test migrations on a development database first
- Keep migrations small and focused on one change
- Use transactions implicitly (each migration runs in a transaction)
- Never modify existing migration files after they've been deployed
- Always use IF NOT EXISTS / IF EXISTS clauses for safety
