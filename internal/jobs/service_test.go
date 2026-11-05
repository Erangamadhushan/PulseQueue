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
	listedJobs []Job
	listErr    error
	listFilter ListJobsFilter
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
	filter ListJobsFilter,
) ([]Job, error) {
	f.listFilter = filter
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listedJobs, nil
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

func TestServiceListJobsAppliesDefaultLimit(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	_, err := service.ListJobs(context.Background(), ListJobsFilter{})
	if err != nil {
		t.Fatalf("ListJobs() error = %v", err)
	}
	if repository.listFilter.Limit != DefaultListLimit {
		t.Errorf("Limit = %d, want %d", repository.listFilter.Limit, DefaultListLimit)
	}
	if repository.listFilter.Offset != 0 {
		t.Errorf("Offset = %d, want 0", repository.listFilter.Offset)
	}
}

func TestServiceListJobsRejectsInvalidFilters(t *testing.T) {
	status := JobStatus("unknown")
	queue := " "
	tests := []struct {
		name   string
		filter ListJobsFilter
		want   error
	}{
		{"zero limit is defaulted", ListJobsFilter{Limit: 0}, nil},
		{"negative limit", ListJobsFilter{Limit: -1}, ErrInvalidListLimit},
		{"limit above maximum", ListJobsFilter{Limit: MaxListLimit + 1}, ErrInvalidListLimit},
		{"negative offset", ListJobsFilter{Offset: -1}, ErrInvalidListOffset},
		{"invalid status", ListJobsFilter{Status: &status}, ErrInvalidJobStatus},
		{"empty queue", ListJobsFilter{Queue: &queue}, ErrInvalidJobQueue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &fakeRepository{}
			_, err := NewService(repository).ListJobs(context.Background(), tt.filter)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestServiceListJobsReturnsRepositoryError(t *testing.T) {
	repositoryError := errors.New("database unavailable")
	repository := &fakeRepository{listErr: repositoryError}

	_, err := NewService(repository).ListJobs(
		context.Background(),
		ListJobsFilter{Limit: 1},
	)
	if !errors.Is(err, repositoryError) {
		t.Fatalf("expected repository error, got %v", err)
	}
}
