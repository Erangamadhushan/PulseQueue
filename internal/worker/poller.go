package worker

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/Erangamadhushan/pulsequeue/internal/jobs"
	"github.com/google/uuid"
)

const maxPollErrorBackoff = 10 * time.Second

type jobExecutionError struct {
	err error
}

func (e *jobExecutionError) Error() string { return e.err.Error() }
func (e *jobExecutionError) Unwrap() error { return e.err }

type infrastructureError struct {
	err error
}

func (e *infrastructureError) Error() string { return e.err.Error() }
func (e *infrastructureError) Unwrap() error { return e.err }

type Repository interface {
	ClaimNext(context.Context, jobs.ClaimJobInput) (*jobs.ClaimResult, error)
	Complete(context.Context, uuid.UUID, string, uuid.UUID) error
	Fail(context.Context, uuid.UUID, string, uuid.UUID, error) error
	RenewLease(context.Context, uuid.UUID, string, uuid.UUID, time.Duration) error
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
	backoff := time.Second
	for {
		claimed, err := w.processOneResult(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			var executionErr *jobExecutionError
			if errors.As(err, &executionErr) {
				w.logger.Error("job execution failed", "error", err)
				if err := waitForNextPoll(ctx, w.pollInterval); err != nil {
					return nil
				}
				continue
			}
			w.logger.Error("worker infrastructure error", "error", err)
			if err := waitForNextPoll(ctx, backoff); err != nil {
				return nil
			}
			backoff *= 2
			if backoff > maxPollErrorBackoff {
				backoff = maxPollErrorBackoff
			}
			continue
		}
		backoff = time.Second
		if !claimed {
			if err := waitForNextPoll(ctx, w.pollInterval); err != nil {
				return nil
			}
		}
	}
}

func (w *Worker) processOne(ctx context.Context) error {
	_, err := w.processOneResult(ctx)
	return err
}

func (w *Worker) processOneResult(ctx context.Context) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	result, err := w.repository.ClaimNext(ctx, jobs.ClaimJobInput{
		WorkerID: w.workerID, LeaseDuration: w.leaseDuration, Queues: w.queues,
	})
	if err != nil {
		return false, &infrastructureError{err: err}
	}
	if result == nil {
		return false, nil
	}

	executionCtx, cancelExecution := context.WithCancel(ctx)
	defer cancelExecution()

	renewalCtx, stopRenewal := context.WithCancel(executionCtx)
	renewalErrors := make(chan error, 1)
	renewalDone := make(chan struct{})
	go w.renewLease(
		renewalCtx,
		renewalDone,
		renewalErrors,
		result.Job.ID,
		result.ClaimToken,
	)

	handlerResult := make(chan error, 1)
	go func() {
		handlerResult <- w.executor.Execute(executionCtx, result.Job)
	}()

	var executionErr error
	renewalFailed := false
	select {
	case executionErr = <-handlerResult:
	case renewalErr := <-renewalErrors:
		cancelExecution()
		renewalFailed = true
		executionErr = renewalErr
	case <-ctx.Done():
		cancelExecution()
		executionErr = ctx.Err()
	}

	stopRenewal()
	<-renewalDone

	if renewalFailed || ctx.Err() != nil {
		handlerErr := <-handlerResult
		if executionErr == nil {
			executionErr = handlerErr
		}
	} else {
		select {
		case handlerErr := <-handlerResult:
			if executionErr == nil {
				executionErr = handlerErr
			}
		default:
		}
	}

	if renewalFailed {
		return true, &infrastructureError{err: executionErr}
	}
	if ctx.Err() != nil {
		return true, ctx.Err()
	}
	if executionErr != nil {
		if persistErr := w.repository.Fail(ctx, result.Job.ID, w.workerID, result.ClaimToken, executionErr); persistErr != nil {
			return true, &infrastructureError{err: errors.Join(executionErr, persistErr)}
		}
		return true, &jobExecutionError{err: executionErr}
	}
	if err := w.repository.Complete(ctx, result.Job.ID, w.workerID, result.ClaimToken); err != nil {
		return true, &infrastructureError{err: err}
	}
	return true, nil
}

func waitForNextPoll(ctx context.Context, interval time.Duration) error {
	timer := time.NewTimer(interval)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (w *Worker) renewLease(
	ctx context.Context,
	done chan<- struct{},
	errs chan<- error,
	jobID uuid.UUID,
	claimToken uuid.UUID,
) {
	defer close(done)

	interval := w.leaseDuration / 3
	if interval <= 0 {
		interval = time.Nanosecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.repository.RenewLease(
				ctx,
				jobID,
				w.workerID,
				claimToken,
				w.leaseDuration,
			); err != nil {
				errs <- err
				return
			}
		}
	}
}
