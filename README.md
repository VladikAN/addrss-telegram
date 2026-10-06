AddRss is an opensource RSS and ATOM feeds reader.

Get start with [AddRss](https://telegram.me/addrssbot) telegram bot, define your subscriptions and get updates through telegram bot messages.

![version - 1](/pics/preview.png)

# How to start

Follow [oficial guidelines](https://core.telegram.org/bots) to register new bot an obtain secret bot token.

Update `docker-compose.yaml` by changing the next keys:
* POSTGRES_USER - new secret database username.
* POSTGRES_PASSWORD - new secret database password.
* AR_TOKEN - secret bot token from telegram API.
* AR_DATABASE - database connection string with POSTGRES_USER and POSTGRES_PASSWORD defined earlier.

Type `docker-compose.exe -f .\docker-compose.yaml up -d` to start bot containers in detached mode.

Type `docker-compose.exe -f .\docker-compose.yaml down` to stop bot containers.

# Database Migrations

The application includes automatic database migrations that run on startup:

- **Automatic initialization** - database schema is created on first run
- **Version tracking** - migrations are applied only once via `schema_migrations` table  
- **Idempotent** - safe to restart, uses `IF NOT EXISTS` checks
- **Embedded** - all SQL migrations are compiled into the binary

Migration files are located in the `migrations/` directory. To create a new migration, add a file with the naming pattern `000XXX_description.up.sql` with appropriate `IF NOT EXISTS` clauses.