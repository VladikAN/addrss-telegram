package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var srv *http.Server

func Start() {
	// Hook for terminate signal
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		cancel()
	}()

	// Close context and call for http shutdown
	go func() {
		<-ctx.Done()
		shutdown()
	}()

	// Configure and start http server
	srv = &http.Server{Handler: newRouter()}
	srv.ListenAndServe()
}

func newRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Timeout(5 * time.Second))

	r.Get("/", GetHandler)

	return r
}

func shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		//		logger.Warnf("Server shutdown error, %s", err)
	}
}
