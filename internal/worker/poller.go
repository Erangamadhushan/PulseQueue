package worker

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/Erangamadhushan/pulsequeue/internal/jobs"
	"github.com/google/uuid"
)

type Repository interface {
	ClaimNext(context.Context, jobs.ClaimJobInput) (*jobs.ClaimResult, error)
	Complete(context.Context, uuid.UUID, string, uuid.UUID) error
	Fail(context.Context, uuid.UUID, string, uuid.UUID, error) error
}

type Worker struct {
	repository    Repository
	executor      *Executor
	workerID      string
	pollInterval  time.Duration
	leaseDuration time.Duration
	queues        []string
	logger        *slog.Logger
}

func New(
	repository Repository,
	registry *Registry,
	workerID string,
	pollInterval time.Duration,
	leaseDuration time.Duration,
	logger *slog.Logger,
) (*Worker, error) {
	return NewWithQueues(repository, registry, workerID, pollInterval, leaseDuration, []string{"default"}, logger)
}

func NewWithQueues(
	repository Repository,
	registry *Registry,
	workerID string,
	pollInterval time.Duration,
	leaseDuration time.Duration,
	queues []string,
	logger *slog.Logger,
) (*Worker, error) {
	if repository == nil || registry == nil {
		return nil, errors.New("repository and registry are required")
	}
	if workerID == "" || pollInterval <= 0 || leaseDuration <= 0 || len(queues) == 0 {
		return nil, errors.New("worker ID, poll interval, and lease duration must be positive")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Worker{
		repository: repository, executor: NewExecutor(registry),
		workerID: workerID, pollInterval: pollInterval,
		leaseDuration: leaseDuration, queues: append([]string(nil), queues...), logger: logger,
	}, nil
}

func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		if err := w.processOne(ctx); err != nil {
			w.logger.Error("worker failed to process job", "error", err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (w *Worker) processOne(ctx context.Context) error {
	result, err := w.repository.ClaimNext(ctx, jobs.ClaimJobInput{
		WorkerID: w.workerID, LeaseDuration: w.leaseDuration, Queues: w.queues,
	})
	if err != nil {
		return err
	}
	if result == nil {
		return nil
	}
	job := result.Job
	if err := w.executor.Execute(ctx, job); err != nil {
		if persistErr := w.repository.Fail(ctx, job.ID, w.workerID, result.ClaimToken, err); persistErr != nil {
			return errors.Join(err, persistErr)
		}
		return err
	}
	return w.repository.Complete(ctx, job.ID, w.workerID, result.ClaimToken)
}
