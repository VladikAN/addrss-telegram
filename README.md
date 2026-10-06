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

The application now includes automatic database migrations that run on startup. This means:

- **No manual SQL scripts needed** - migrations are applied automatically
- **Safe to restart** - migrations are only applied once using version tracking
- **Embedded in binary** - all migrations are compiled into the application

Database schema is initialized automatically when the application starts. The `deploy/01-init-database.sql` file is kept for reference, but the actual schema is managed through migrations in the `migrations/` directory.

For more information about creating and managing migrations, see [migrations/README.md](migrations/README.md).