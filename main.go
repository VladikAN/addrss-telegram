package main

import (
	"flag"

	log "github.com/go-pkgz/lgr"
	"github.com/vladikan/feedreader-telegrambot/server"
)

func main() {
	log.Setup(log.Msec, log.LevelBraces)

	token := flag.String("token", "", "bot secret token")
	connection := flag.String("db", "postgres://admin:admin@localhost:5432/feed", "database connection string")
	debug := flag.Bool("debug", false, "debug flag")
	readerInterval := flag.Int("reader-interval", 5, "Interval in seconds to read subscriptions for updates")
	readerFeeds := flag.Int("reader-feeds", 10, "How many feeds to read between intervals")
	flag.Parse()

	if len(*token) == 0 {
		log.Print("PANIC bot token is not defined")
	}

	opt := server.Options{
		Token:          *token,
		Connection:     *connection,
		Debug:          *debug,
		ReaderInterval: *readerInterval,
		ReaderFeeds:    *readerFeeds,
	}
	server.Start(opt)
}
