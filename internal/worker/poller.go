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
	return w.run(ctx, 0)
}

// RunGraceful stops polling on cancellation and gives the active job a grace
// period before cancelling its execution context.
func (w *Worker) RunGraceful(ctx context.Context, shutdownTimeout time.Duration) error {
	if shutdownTimeout <= 0 {
		return errors.New("shutdown timeout must be positive")
	}
	return w.run(ctx, shutdownTimeout)
}

func (w *Worker) run(ctx context.Context, shutdownTimeout time.Duration) error {
	backoff := time.Second
	for {
		var claimed bool
		var err error
		if shutdownTimeout > 0 {
			claimed, err = w.processOneGracefulResult(ctx, shutdownTimeout)
		} else {
			claimed, err = w.processOneResult(ctx)
		}
		if err != nil {
			if errors.Is(err, ErrWorkerShutdown) {
				return nil
			}
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
	return w.processOneWithContexts(ctx, ctx, 0)
}

func (w *Worker) processOneGracefulResult(ctx context.Context, shutdownTimeout time.Duration) (bool, error) {
	return w.processOneWithContexts(ctx, context.WithoutCancel(ctx), shutdownTimeout)
}

func (w *Worker) processOneWithContexts(
	pollCtx context.Context,
	executionParent context.Context,
	shutdownTimeout time.Duration,
) (bool, error) {
	if err := pollCtx.Err(); err != nil {
		return false, err
	}
	result, err := w.repository.ClaimNext(pollCtx, jobs.ClaimJobInput{
		WorkerID: w.workerID, LeaseDuration: w.leaseDuration, Queues: w.queues,
	})
	if err != nil {
		return false, &infrastructureError{err: err}
	}
	if result == nil {
		return false, nil
	}

	executionCtx, cancelExecution := context.WithCancel(executionParent)
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
	shutdownTimedOut := false
	handlerFinished := false
	select {
	case executionErr = <-handlerResult:
		handlerFinished = true
	case renewalErr := <-renewalErrors:
		cancelExecution()
		renewalFailed = true
		executionErr = renewalErr
	case <-pollCtx.Done():
		if shutdownTimeout <= 0 {
			cancelExecution()
			executionErr = pollCtx.Err()
			break
		}
		timer := time.NewTimer(shutdownTimeout)
		select {
		case executionErr = <-handlerResult:
			handlerFinished = true
			timer.Stop()
		case renewalErr := <-renewalErrors:
			timer.Stop()
			cancelExecution()
			renewalFailed = true
			executionErr = renewalErr
		case <-timer.C:
			shutdownTimedOut = true
			cancelExecution()
			executionErr = ErrWorkerShutdown
		}
	}

	stopRenewal()
	<-renewalDone

	if renewalFailed || (pollCtx.Err() != nil && !shutdownTimedOut && shutdownTimeout <= 0) {
		handlerErr := <-handlerResult
		if executionErr == nil {
			executionErr = handlerErr
		}
	} else {
		select {
		case handlerErr := <-handlerResult:
			handlerFinished = true
			if executionErr == nil {
				executionErr = handlerErr
			}
		default:
		}
	}

	if renewalFailed {
		return true, &infrastructureError{err: executionErr}
	}
	if shutdownTimedOut {
		if !handlerFinished {
			<-handlerResult
		}
		return true, ErrWorkerShutdown
	}
	if pollCtx.Err() != nil && shutdownTimeout <= 0 {
		return true, pollCtx.Err()
	}
	if executionErr != nil {
		persistenceCtx := pollCtx
		if shutdownTimeout > 0 {
			persistenceCtx = context.WithoutCancel(pollCtx)
		}
		if persistErr := w.repository.Fail(persistenceCtx, result.Job.ID, w.workerID, result.ClaimToken, executionErr); persistErr != nil {
			return true, &infrastructureError{err: errors.Join(executionErr, persistErr)}
		}
		return true, &jobExecutionError{err: executionErr}
	}
	persistenceCtx := pollCtx
	if shutdownTimeout > 0 {
		persistenceCtx = context.WithoutCancel(pollCtx)
	}
	if err := w.repository.Complete(persistenceCtx, result.Job.ID, w.workerID, result.ClaimToken); err != nil {
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
