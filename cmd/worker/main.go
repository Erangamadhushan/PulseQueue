package main

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)
	go func() {
		<-signals
		cancel()
		signal.Stop(signals)
	}()

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	workerCfg, err := worker.LoadWorkerConfig()
	if err != nil {
		return fmt.Errorf("load worker configuration: %w", err)
	}
	db, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect PostgreSQL: %w", err)
	}
	defer db.Close()

	registry := worker.NewRegistry()
	repository := jobs.NewPostgresRepository(db)
	engine, err := worker.NewWithQueues(
		repository,
		registry,
		workerCfg.ID,
		workerCfg.PollInterval,
		workerCfg.LeaseDuration,
		workerCfg.Queues,
		logger,
	)
	if err != nil {
		return fmt.Errorf("create worker: %w", err)
	}
	logger.Info("PulseQueue worker started", "worker_id", workerCfg.ID)
	if err := engine.RunGraceful(ctx, workerCfg.ShutdownTimeout); err != nil {
		return fmt.Errorf("run worker: %w", err)
	}
	logger.Info("PulseQueue worker stopped")
	return nil
}
