# AGENTS.md

## Cursor Cloud specific instructions

### Overview
AddRss is a Telegram bot for RSS/ATOM feed reading, written in Go. For local development, use `--local` mode (HTTP server, no Telegram token needed). Production mode requires `AR_TOKEN`.

### Go version
This project requires **Go 1.26**. The binary is installed at `/usr/local/go/bin/go`. Ensure `PATH` includes `/usr/local/go/bin` (the update script handles this).

### Database
PostgreSQL 16 runs locally. The database `feed` with user `admin`/`admin` is pre-configured. Schema lives in `deploy/01-init-database.sql`. Start PostgreSQL if not running:
```
sudo pg_ctlcluster 16 main start
```

### Running tests
```
go test ./...
```
Tests cover `parser`, `server`, and `templates` packages. No tests require the database or Telegram token.

### Lint / vet
```
go vet ./...
```

### Building
```
go build -o addrss-telegram .
```

### Running the application

**Local HTTP mode** (no Telegram token required):
```
go run . --local --debug
```
or with explicit database:
```
AR_DATABASE="postgres://admin:admin@localhost:5432/feed" go run . --local --debug --http-port 8080
```
Send commands via HTTP:
```
curl -X POST http://localhost:8080/command -d '{"user_id": 1, "text": "/help"}'
```

**Telegram mode** (production, requires `AR_TOKEN`):
```
AR_TOKEN=<token> AR_DATABASE="postgres://admin:admin@localhost:5432/feed" go run .
```
Without a valid `AR_TOKEN`, Telegram mode will panic at startup.

### Key environment variables
See `README.md` for the full list. Defaults connect to `postgres://admin:admin@localhost:5432/feed`.

### Startup order
1. Start PostgreSQL: `sudo pg_ctlcluster 16 main start`
2. Run the app: `go run . --local --debug`

PostgreSQL must be running before the app starts. The app will panic if it cannot connect to the database.

### Docker
Docker is not installed in the Cloud Agent VM. Use `go run .` or the built binary directly with the local PostgreSQL instance for development.
