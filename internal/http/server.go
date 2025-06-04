package http

import (
	"context"
	"log"
	"net"
	"net/http"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	_shutdownPeriod      = 15 * time.Second
	_shutdownHardPeriod  = 3 * time.Second
	_readinessDrainDelay = 5 * time.Second
)

var isShuttingDown atomic.Bool

func ListenAndServe() {
	// Setup signal context
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Ensure in-flight requests aren't cancelled immediately on SIGTERM
	ongoingCtx, stopOngoingGracefully := context.WithCancel(context.Background())
	s := &http.Server{
		Addr:    ":4000",
		Handler: router(),
		BaseContext: func(_ net.Listener) context.Context {
			return ongoingCtx
		},
	}

	go func() {
		log.Println("server starting at :4000")
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// Wait for signal
	<-rootCtx.Done()
	stop()
	isShuttingDown.Store(true)
	log.Println("Received shutdown signal, shutting down.")

	// Give time for readiness check to propagate
	time.Sleep(_readinessDrainDelay)
	log.Println("Readiness check propagated, now waiting for ongoing requests to finish.")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), _shutdownPeriod)
	defer cancel()
	err := s.Shutdown(shutdownCtx)
	stopOngoingGracefully()
	if err != nil {
		log.Println("Failed to wait for ongoing requests to finish, waiting for forced cancellation.")
		time.Sleep(_shutdownHardPeriod)
	}

	log.Println("Server shut down gracefully.")
}

func router() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /foo", func(w http.ResponseWriter, r *http.Request) {
		log.Println("hello")
	})
	mux.HandleFunc("GET /foo/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/foo/"):]
		log.Println("ID is:", id)
	})
	mux.HandleFunc("GET /bar", func(w http.ResponseWriter, r *http.Request) {
		log.Println("hello bar")
	})
	mux.HandleFunc("POST /foo", func(w http.ResponseWriter, r *http.Request) {
		log.Println("hello post foo")
	})
	return mux
}
