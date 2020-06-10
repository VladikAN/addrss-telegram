package main

import (
	"flag"

	log "github.com/go-pkgz/lgr"
	"github.com/vladikan/feedreader-telegrambot/server"
)

func main() {
	log.Setup(log.Msec, log.LevelBraces)

	token := flag.String("token", "", "bot secret token")
	debug := flag.Bool("Debug", false, "debug flag")
	flag.Parse()

	if len(*token) == 0 {
		log.Print("PANIC bot token is not defined")
	}

	server.Start(*token, *debug)
}
