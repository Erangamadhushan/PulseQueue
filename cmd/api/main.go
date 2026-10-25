package main

import (
	"context" //
	"errors"
	"fmt"
	"log/slog" // go's standard structured logging package.
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Erangamadhushan/pulsequeue/internal/config"
	"github.com/Erangamadhushan/pulsequeue/internal/health"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		slog.Error("PulseQueue API stopped with an error",
			"error", err,
		)
		os.Exit(1)
	}

}

func run() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health.HealthHandler)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer stop()

	serverErrors := make(chan error, 1)

	go func() {
		slog.Info("PulseQueue API starting",
			"address", server.Addr,
		)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			slog.Info("PulseQueue API stopped gracefully")
			return nil
		}

		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		slog.Info("PulseQueue API shutting down gracefully")
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		_ = server.Close()
		return fmt.Errorf("failed to shutdown server gracefully: %w", err)
	}

	slog.Info("PulseQueue API stopped gracefully")
	return nil
}

/*

HTTP timeouts
ReadHeaderTimeout, ReadTimeout, WriteTimeout, and IdleTimeout limit how long different parts of an HTTP connection may take. These help protect the server from slow or stalled clients.


Goroutines and channels
go func() starts the server in a separate goroutine. The buffered serverErrors channel lets that goroutine report when the server stops, while the main goroutine waits for either a server error or a shutdown signal.


Graceful shutdown
signal.NotifyContext detects termination signals. server.Shutdown stops accepting new connections and waits for active requests to finish, subject to the deadline.

Structured logging and error wrapping
slog produces structured JSON logs. fmt.Errorf with %w preserves the underlying error so that callers can inspect it using errors.Is or errors.As.

*/
