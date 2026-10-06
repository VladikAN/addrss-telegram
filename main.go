package main

import (
	"embed"
	"fmt"

	log "github.com/go-pkgz/lgr"
	"github.com/umputun/go-flags"
	"github.com/vladikan/addrss-telegram/database"
	"github.com/vladikan/addrss-telegram/server"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type opts struct {
	Token          string `long:"token" env:"AR_TOKEN" description:"telegram bot secret token"`
	Connection     string `long:"db" env:"AR_DATABASE" default:"postgres://admin:admin@localhost:5432/feed" description:"postgres database connection string"`
	Debug          bool   `long:"debug" env:"AR_DEBUG" description:"turn on-off debug messages"`
	ReaderInterval int    `long:"reader-interval" env:"AR_READER_INTERVAL" default:"600" description:"Interval in seconds to read subscriptions for updates"`
	ReaderFeeds    int    `long:"reader-feeds" env:"AR_READER_FEEDS" default:"100" description:"How many feeds to read between intervals"`
	BotAdmin       int64  `long:"bot-admin" env:"AR_BOT_ADMIN" default:"0" description:"Bot admin user id for extra features"`
	Local          bool   `long:"local" env:"AR_LOCAL" description:"run in local HTTP mode without Telegram API"`
	HTTPPort       int    `long:"http-port" env:"AR_HTTP_PORT" default:"8080" description:"HTTP port for local development mode"`
}

func main() {
	op := opts{}
	if _, err := flags.Parse(&op); err != nil {
		panic(fmt.Sprintf("PANIC error while reading input options: %s", err))
	}

	if !op.Local && len(op.Token) == 0 {
		panic("PANIC bot token is missed")
	}

	logOpt := []log.Option{log.Msec, log.LevelBraces}
	if op.Debug {
		logOpt = append(logOpt, log.Debug)
	}
	log.Setup(logOpt...)

	// Run database migrations before starting the server
	log.Printf("INFO Running database migrations...")
	if err := database.RunMigrations(op.Connection, migrationsFS); err != nil {
		log.Printf("WARN Failed to run migrations: %s (continuing anyway)", err)
	} else {
		log.Printf("INFO Database migrations completed successfully")
	}

	opt := server.Options{
		Token:          op.Token,
		Connection:     op.Connection,
		Debug:          op.Debug,
		ReaderInterval: op.ReaderInterval,
		ReaderFeeds:    op.ReaderFeeds,
		BotAdmin:       op.BotAdmin,
		HTTPPort:       op.HTTPPort,
	}

	if op.Local {
		server.StartLocal(opt)
	} else {
		server.Start(opt)
	}
}
