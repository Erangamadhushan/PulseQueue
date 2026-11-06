package worker

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Erangamadhushan/pulsequeue/internal/config"
	"github.com/Erangamadhushan/pulsequeue/internal/database"
	"github.com/Erangamadhushan/pulsequeue/internal/jobs"
	"github.com/Erangamadhushan/pulsequeue/internal/worker"
)

func main() {
	if err := run(); err != nil {
		log.Printf("worker stopped with error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	db, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect PostgreSQL: %w", err)
	}
	defer db.Close()

	registry := worker.NewRegistry()
	repository := jobs.NewPostgresRepository(db)
	engine, err := worker.New(
		repository,
		registry,
		cfg.WorkerID,
		cfg.WorkerPollInterval,
		cfg.WorkerLeaseDuration,
		logger,
	)
	if err != nil {
		return fmt.Errorf("create worker: %w", err)
	}
	logger.Info("PulseQueue worker started", "worker_id", cfg.WorkerID)
	if err := engine.Run(ctx); err != nil {
		return fmt.Errorf("run worker: %w", err)
	}
	logger.Info("PulseQueue worker stopped")
	return nil
}
