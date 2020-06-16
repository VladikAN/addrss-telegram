package main

import (
	"flag"

	log "github.com/go-pkgz/lgr"
	"github.com/vladikan/feedreader-telegrambot/server"
)

func main() {
	log.Setup(log.Msec, log.LevelBraces)

	token := flag.String("token", "", "bot secret token")
	dbConnection := flag.String("db", "", "database connection string")
	debug := flag.Bool("Debug", false, "debug flag")
	flag.Parse()

	opt := server.Options{
		Token:        *token,
		DbConnection: *dbConnection,
		Debug:        *debug,
	}

	if len(*token) == 0 {
		log.Print("PANIC bot token is not defined")
	}

	server.Start(opt)
}
