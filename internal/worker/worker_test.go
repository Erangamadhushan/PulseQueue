package worker

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/Erangamadhushan/pulsequeue/internal/jobs"
	"github.com/google/uuid"
)

func TestRegistryRejectsInvalidAndDuplicateHandlers(t *testing.T) {
	registry := NewRegistry()
	handler := func(context.Context, jobs.Job) error { return nil }

	for _, test := range []struct {
		name string
		call func() error
		want error
	}{
		{"empty name", func() error { return registry.Register("", handler) }, ErrEmptyHandlerName},
		{"nil handler", func() error { return registry.Register("job", nil) }, ErrNilHandler},
		{"first registration", func() error { return registry.Register("job", handler) }, nil},
		{"duplicate registration", func() error { return registry.Register("job", handler) }, ErrDuplicateHandler},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := test.call(); !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestRegistryGetAndHas(t *testing.T) {
	registry := NewRegistry()
	handler := func(context.Context, jobs.Job) error { return nil }
	if err := registry.Register(" send_email ", handler); err != nil {
		t.Fatal(err)
	}

	got, ok := registry.Get("send_email")
	if !ok || got == nil || !registry.Has("send_email") {
		t.Fatal("expected registered handler to be available")
	}
	if _, ok := registry.Get("missing"); ok || registry.Has("missing") {
		t.Fatal("did not expect unknown handler")
	}
}

func TestExecutorConvertsPanicsToErrors(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Register("panic", func(context.Context, jobs.Job) error {
		panic("boom")
	}); err != nil {
		t.Fatal(err)
	}

	err := NewExecutor(registry).Execute(context.Background(), jobs.Job{Type: "panic"})
	if err == nil || err.Error() != "handler panic: boom" {
		t.Fatalf("error = %v, want panic error", err)
	}
}

func TestWorkerPersistsSuccessAndFailure(t *testing.T) {
	repository := &fakeRepository{}
	repository.renewed = make(chan struct{}, 1)
	registry := NewRegistry()
	if err := registry.Register("success", func(context.Context, jobs.Job) error { return nil }); err != nil {
		t.Fatal(err)
	}

	if err := registry.Register("failure", func(context.Context, jobs.Job) error { return errors.New("failed") }); err != nil {
		t.Fatal(err)
	}

	engine, err := New(repository, registry, "worker-1", time.Millisecond, time.Second, slog.Default())
	if err != nil {
		t.Fatal(err)
	}

	repository.job = jobs.Job{ID: uuid.New(), Type: "success"}
	if err := engine.processOne(context.Background()); err != nil {
		t.Fatal(err)
	}
	if repository.completed != repository.job.ID {
		t.Fatalf("completed ID = %v, want %v", repository.completed, repository.job.ID)
	}

	repository.job = jobs.Job{ID: uuid.New(), Type: "failure"}
	if err := engine.processOne(context.Background()); err == nil {
		t.Fatal("expected handler failure")
	}
	if repository.failed != repository.job.ID {
		t.Fatalf("failed ID = %v, want %v", repository.failed, repository.job.ID)
	}
}

func TestWorkerFailsUnknownHandler(t *testing.T) {
	repository := &fakeRepository{}
	engine, err := New(repository, NewRegistry(), "worker-1", time.Hour, time.Minute, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	repository.job = jobs.Job{ID: uuid.New(), Type: "missing"}

	err = engine.processOne(context.Background())
	if err == nil {
		t.Fatal("expected unknown handler error")
	}
	if repository.failed != repository.job.ID {
		t.Fatalf("failed ID = %v, want %v", repository.failed, repository.job.ID)
	}
}

func TestWorkerPersistsPanicAsFailure(t *testing.T) {
	repository := &fakeRepository{}
	registry := NewRegistry()
	if err := registry.Register("panic", func(context.Context, jobs.Job) error {
		panic("boom")
	}); err != nil {
		t.Fatal(err)
	}
	engine, err := New(repository, registry, "worker-1", time.Hour, time.Minute, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	repository.job = jobs.Job{ID: uuid.New(), Type: "panic"}

	err = engine.processOne(context.Background())
	if err == nil || err.Error() != "handler panic: boom" {
		t.Fatalf("error = %v, want handler panic", err)
	}
	if repository.failed != repository.job.ID {
		t.Fatalf("failed ID = %v, want %v", repository.failed, repository.job.ID)
	}
}

func TestWorkerRenewsLeaseDuringExecution(t *testing.T) {
	repository := &fakeRepository{}
	repository.renewed = make(chan struct{}, 1)
	registry := NewRegistry()
	handlerStarted := make(chan struct{})
	releaseHandler := make(chan struct{})
	if err := registry.Register("slow", func(context.Context, jobs.Job) error {
		close(handlerStarted)
		<-releaseHandler
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	engine, err := New(repository, registry, "worker-1", time.Hour, 30*time.Millisecond, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	repository.job = jobs.Job{ID: uuid.New(), Type: "slow"}

	result := make(chan error, 1)
	go func() {
		result <- engine.processOne(context.Background())
	}()
	<-handlerStarted
	select {
	case <-repository.renewed:
	case <-time.After(time.Second):
		t.Fatal("lease was not renewed")
	}
	close(releaseHandler)
	if err := <-result; err != nil {
		t.Fatalf("processOne() error = %v", err)
	}
}

func TestWorkerStopsOnLeaseOwnershipLoss(t *testing.T) {
	repository := &fakeRepository{renewErr: jobs.ErrJobOwnershipLost}
	registry := NewRegistry()
	handlerCancelled := make(chan struct{})
	if err := registry.Register("slow", func(ctx context.Context, _ jobs.Job) error {
		<-ctx.Done()
		close(handlerCancelled)
		return ctx.Err()
	}); err != nil {
		t.Fatal(err)
	}

	engine, err := New(repository, registry, "worker-1", time.Hour, 30*time.Millisecond, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	repository.job = jobs.Job{ID: uuid.New(), Type: "slow"}

	err = engine.processOne(context.Background())
	if !errors.Is(err, jobs.ErrJobOwnershipLost) {
		t.Fatalf("processOne() error = %v, want ownership loss", err)
	}
	select {
	case <-handlerCancelled:
	case <-time.After(time.Second):
		t.Fatal("handler was not cancelled")
	}
	if repository.completed != uuid.Nil || repository.failed != uuid.Nil {
		t.Fatal("worker persisted a result after ownership loss")
	}
}

func TestWaitForNextPollHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := waitForNextPoll(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatalf("waitForNextPoll() error = %v, want context.Canceled", err)
	}
}

func TestWorkerPassesConfiguredQueuesToClaim(t *testing.T) {
	repository := &fakeRepository{}
	registry := NewRegistry()
	if err := registry.Register("queued", func(context.Context, jobs.Job) error { return nil }); err != nil {
		t.Fatal(err)
	}
	engine, err := NewWithQueues(
		repository,
		registry,
		"worker-queues",
		time.Hour,
		time.Minute,
		[]string{"emails", "critical"},
		slog.Default(),
	)
	if err != nil {
		t.Fatal(err)
	}
	repository.job = jobs.Job{ID: uuid.New(), Type: "queued"}

	if err := engine.processOne(context.Background()); err != nil {
		t.Fatalf("processOne() error = %v", err)
	}
	if repository.claimInput.WorkerID != "worker-queues" {
		t.Fatalf("worker ID = %q, want worker-queues", repository.claimInput.WorkerID)
	}
	if repository.claimInput.LeaseDuration != time.Minute {
		t.Fatalf("lease duration = %v, want 1m", repository.claimInput.LeaseDuration)
	}
	if len(repository.claimInput.Queues) != 2 ||
		repository.claimInput.Queues[0] != "emails" ||
		repository.claimInput.Queues[1] != "critical" {
		t.Fatalf("queues = %v, want [emails critical]", repository.claimInput.Queues)
	}
}

func TestWorkerJoinsHandlerAfterCancellation(t *testing.T) {
	repository := &fakeRepository{}
	registry := NewRegistry()
	handlerStarted := make(chan struct{})
	handlerFinished := make(chan struct{})
	if err := registry.Register("cancelled", func(ctx context.Context, _ jobs.Job) error {
		close(handlerStarted)
		<-ctx.Done()
		close(handlerFinished)
		return ctx.Err()
	}); err != nil {
		t.Fatal(err)
	}
	engine, err := New(repository, registry, "worker-1", time.Hour, time.Minute, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	repository.job = jobs.Job{ID: uuid.New(), Type: "cancelled"}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		result <- engine.processOne(ctx)
	}()
	select {
	case <-handlerStarted:
	case <-time.After(time.Second):
		t.Fatal("handler did not start")
	}
	cancel()
	select {
	case <-handlerFinished:
	case <-time.After(time.Second):
		t.Fatal("handler was not joined after cancellation")
	}
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Fatalf("processOne() error = %v, want context.Canceled", err)
	}
}

func TestWorkerReturnsInfrastructureErrorWhenCompletionFails(t *testing.T) {
	repository := &fakeRepository{completeErr: errors.New("database unavailable")}
	registry := NewRegistry()
	if err := registry.Register("success", func(context.Context, jobs.Job) error { return nil }); err != nil {
		t.Fatal(err)
	}
	engine, err := New(repository, registry, "worker-1", time.Hour, time.Minute, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	repository.job = jobs.Job{ID: uuid.New(), Type: "success"}

	err = engine.processOne(context.Background())
	if !errors.Is(err, repository.completeErr) {
		t.Fatalf("processOne() error = %v, want completion error", err)
	}
	var infrastructureErr *infrastructureError
	if !errors.As(err, &infrastructureErr) {
		t.Fatalf("processOne() error = %T, want infrastructureError", err)
	}
}

type fakeRepository struct {
	job         jobs.Job
	claimInput  jobs.ClaimJobInput
	completed   uuid.UUID
	failed      uuid.UUID
	renewErr    error
	renewed     chan struct{}
	completeErr error
}

func (r *fakeRepository) ClaimNext(_ context.Context, input jobs.ClaimJobInput) (*jobs.ClaimResult, error) {
	r.claimInput = input
	token := uuid.New()
	r.job.ClaimToken = &token
	return &jobs.ClaimResult{Job: r.job, ClaimToken: token}, nil
}

func (r *fakeRepository) Complete(_ context.Context, id uuid.UUID, _ string, _ uuid.UUID) error {
	r.completed = id
	return r.completeErr
}

func (r *fakeRepository) Fail(_ context.Context, id uuid.UUID, _ string, _ uuid.UUID, _ error) error {
	r.failed = id
	return nil
}

func (r *fakeRepository) RenewLease(
	_ context.Context,
	_ uuid.UUID,
	_ string,
	_ uuid.UUID,
	_ time.Duration,
) error {
	if r.renewed != nil {
		select {
		case r.renewed <- struct{}{}:
		default:
		}
	}
	if r.renewErr != nil {
		return r.renewErr
	}
	return nil
}
