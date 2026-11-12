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

type fakeRepository struct {
	job       jobs.Job
	completed uuid.UUID
	failed    uuid.UUID
}

func (r *fakeRepository) ClaimNext(context.Context, jobs.ClaimJobInput) (*jobs.ClaimResult, error) {
	token := uuid.New()
	r.job.ClaimToken = &token
	return &jobs.ClaimResult{Job: r.job, ClaimToken: token}, nil
}

func (r *fakeRepository) Complete(_ context.Context, id uuid.UUID, _ string, _ uuid.UUID) error {
	r.completed = id
	return nil
}

func (r *fakeRepository) Fail(_ context.Context, id uuid.UUID, _ string, _ uuid.UUID, _ error) error {
	r.failed = id
	return nil
}
