package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeRepository struct {
	createdJob Job
	createErr  error
}

func (f *fakeRepository) Create(
	_ context.Context,
	job Job,
) (Job, error) {
	if f.createErr != nil {
		return Job{}, f.createErr
	}

	f.createdJob = job
	return job, nil
}

func (f *fakeRepository) GetByID(
	_ context.Context,
	_ uuid.UUID,
) (Job, error) {
	return Job{}, ErrJobNotFound
}

func (f *fakeRepository) List(
	_ context.Context,
) ([]Job, error) {
	return nil, nil
}

func TestServiceCreateJobAppliesDefaults(
	t *testing.T,
) {
	repository := &fakeRepository{}

	service := NewService(repository)

	fixedNow := time.Date(
		2026,
		time.November,
		2,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	service.now = func() time.Time {
		return fixedNow
	}

	input := CreateJobInput{
		Type: "send_email",
		Payload: json.RawMessage(
			`{"email":"test@example.com"}`,
		),
	}

	job, err := service.CreateJob(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	if job.Queue != DefaultQueue {
		t.Errorf(
			"Queue = %q, want %q",
			job.Queue,
			DefaultQueue,
		)
	}

	if job.Status != JobStatusPending {
		t.Errorf(
			"Status = %q, want %q",
			job.Status,
			JobStatusPending,
		)
	}

	if job.Attempts != 0 {
		t.Errorf(
			"Attempts = %d, want 0",
			job.Attempts,
		)
	}

	if job.MaxAttempts != DefaultMaxAttempts {
		t.Errorf(
			"MaxAttempts = %d, want %d",
			job.MaxAttempts,
			DefaultMaxAttempts,
		)
	}

	if !job.RunAt.Equal(fixedNow) {
		t.Errorf(
			"RunAt = %v, want %v",
			job.RunAt,
			fixedNow,
		)
	}
}

func TestServiceCreateJobRejectsEmptyType(
	t *testing.T,
) {
	service := NewService(&fakeRepository{})

	_, err := service.CreateJob(
		context.Background(),
		CreateJobInput{
			Type:    "   ",
			Payload: json.RawMessage(`{}`),
		},
	)

	if !errors.Is(err, ErrInvalidJobType) {
		t.Fatalf(
			"expected ErrInvalidJobType, got %v",
			err,
		)
	}
}

func TestServiceCreateJobRejectsInvalidPayload(
	t *testing.T,
) {
	service := NewService(&fakeRepository{})

	_, err := service.CreateJob(
		context.Background(),
		CreateJobInput{
			Type: "test",
			Payload: json.RawMessage(
				`{"invalid":`,
			),
		},
	)

	if !errors.Is(err, ErrInvalidJobPayload) {
		t.Fatalf(
			"expected ErrInvalidJobPayload, got %v",
			err,
		)
	}
}

func TestServiceCreateJobRejectsNegativeMaxAttempts(
	t *testing.T,
) {
	service := NewService(&fakeRepository{})

	_, err := service.CreateJob(
		context.Background(),
		CreateJobInput{
			Type:        "test",
			Payload:     json.RawMessage(`{}`),
			MaxAttempts: -1,
		},
	)

	if !errors.Is(err, ErrInvalidMaxAttempts) {
		t.Fatalf(
			"expected ErrInvalidMaxAttempts, got %v",
			err,
		)
	}
}

func TestServiceCreateJobReturnsRepositoryError(
	t *testing.T,
) {
	repositoryError := errors.New("database unavailable")

	service := NewService(
		&fakeRepository{
			createErr: repositoryError,
		},
	)

	_, err := service.CreateJob(
		context.Background(),
		CreateJobInput{
			Type:    "test",
			Payload: json.RawMessage(`{}`),
		},
	)

	if !errors.Is(err, repositoryError) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}
