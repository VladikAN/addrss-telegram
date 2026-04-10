package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	log "github.com/go-pkgz/lgr"
	"github.com/vladikan/addrss-telegram/database"
	"github.com/vladikan/addrss-telegram/templates"
)

type httpRequest struct {
	UserID int64  `json:"user_id"`
	Text   string `json:"text"`
}

type httpReply struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

// StartLocal runs the bot in local HTTP mode without requiring the Telegram API.
func StartLocal(options Options) {
	ctx, cancel := context.WithCancel(context.Background())
	go handleTerminate(cancel)

	templates.SetTemplateOutput()

	var err error
	db, err = database.Open(ctx, options.Connection)
	if err != nil {
		log.Printf("PANIC Error while connecting to the database: %s", err)
	}
	defer db.Close()

	replyQueue := make(chan Reply, 256)
	go drainLocalReplies(replyQueue)
	defer close(replyQueue)

	reader := &Reader{Interval: options.ReaderInterval, Feeds: options.ReaderFeeds, DB: db, Outbox: replyQueue}
	reader.Start()
	defer reader.Stop()

	mux := http.NewServeMux()
	mux.HandleFunc("/command", makeCommandHandler(&options))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	addr := fmt.Sprintf(":%d", options.HTTPPort)
	srv := &http.Server{Addr: addr, Handler: mux}

	go func() {
		log.Printf("INFO Local HTTP server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("PANIC HTTP server error: %s", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Print("INFO Shutting down local HTTP server")
	srv.Close()
}

func makeCommandHandler(opt *Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req httpRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		if req.UserID == 0 {
			req.UserID = 1
		}

		cmd := newLocalCommand(req.UserID, req.Text, opt)
		replies := cmd.run()

		out := make([]httpReply, len(replies))
		for i, rpl := range replies {
			out[i] = httpReply{ChatID: rpl.ChatID, Text: rpl.Text}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(out)
	}
}

func drainLocalReplies(queue chan Reply) {
	for msg := range queue {
		log.Printf("INFO [local] reply to %d: %s", msg.ChatID, msg.Text)
	}
}

func handleTerminateLocal(cancel context.CancelFunc) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Print("WARN System interrupt or terminate signal")
	cancel()
}
