package main

import (
	log "github.com/go-pkgz/lgr"
	"github.com/umputun/go-flags"
	"github.com/vladikan/addrss-telegram/server"
)

type opts struct {
	Token          string `long:"token" env:"AR_TOKEN" description:"telegram bot secret token"`
	Conneciton     string `long:"db" env:"AR_DATABASE" default:"postgres://admin:admin@localhost:5432/feed" description:"postgres database connection string"`
	Debug          bool   `long:"debug" env:"AR_DEBUG" description:"turn on-off debug messages"`
	ReaderInterval int    `long:"reader-interval" env:"AR_READER_INTERVAL" default:"600" description:"Interval in seconds to read subscriptions for updates"`
	ReaderFeeds    int    `long:"reader-feeds" env:"AR_READER_FEEDS" default:"100" description:"How many feeds to read between intervals"`
}

func main() {
	log.Setup(log.Msec, log.LevelBraces)

	op := opts{}
	if _, err := flags.Parse(&op); err != nil {
		log.Printf("PANIC error while reading input options: %s", err)
	}

	if len(op.Token) == 0 {
		log.Print("PANIC bot token is not defined")
	}

	opt := server.Options{
		Token:          op.Token,
		Connection:     op.Conneciton,
		Debug:          op.Debug,
		ReaderInterval: op.ReaderInterval,
		ReaderFeeds:    op.ReaderFeeds,
	}
	server.Start(opt)
}
