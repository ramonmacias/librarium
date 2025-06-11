package app

import (
	"context"
	"librarium/internal/http"
	"log"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	_shutdownHardPeriod  = 3 * time.Second
	_readinessDrainDelay = 5 * time.Second
)

type Application struct {
	server *http.Server

	isShuttingDown atomic.Bool
}

func NewLibrariumApplication() (*Application, error) {
	return nil, nil
}

func (a *Application) Run() {
	// Setup signal context
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a.server.ListenAndServe()
	// Wait for signal
	<-rootCtx.Done()
	stop()
	a.isShuttingDown.Store(true)
	log.Println("Received shutdown signal, shutting down.")

	// TODO: We should close the db connection in here as well

	// Give time for readiness check to propagate
	time.Sleep(_readinessDrainDelay)
	log.Println("Readiness check propagated, now waiting for ongoing requests to finish.")

	err := a.server.Shutdown()
	if err != nil {
		log.Println("Failed to wait for ongoing requests to finish, waiting for forced cancellation.")
		time.Sleep(_shutdownHardPeriod)
	}

	log.Println("Application shut down gracefully.")
}
