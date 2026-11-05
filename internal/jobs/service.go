package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultQueue       = "default"
	DefaultMaxAttempts = 3
	DefaultListLimit   = 10
	MaxListLimit       = 100
)

type CreateJobInput struct {
	Queue       string
	Type        string
	Payload     json.RawMessage
	MaxAttempts int
	RunAt       *time.Time
}

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
		now:        time.Now,
	}
}

func (s *Service) CreateJob(
	ctx context.Context,
	input CreateJobInput,
) (Job, error) {
	queue := strings.TrimSpace(input.Queue)
	if queue == "" {
		queue = DefaultQueue
	}

	jobType := strings.TrimSpace(input.Type)
	if jobType == "" {
		return Job{}, ErrInvalidJobType
	}

	if !json.Valid(input.Payload) {
		return Job{}, ErrInvalidJobPayload
	}

	maxAttempts := input.MaxAttempts
	if maxAttempts == 0 {
		maxAttempts = DefaultMaxAttempts
	}

	if maxAttempts < 1 {
		return Job{}, ErrInvalidMaxAttempts
	}

	runAt := s.now().UTC()

	if input.RunAt != nil {
		runAt = input.RunAt.UTC()
	}

	job := Job{
		Queue:       queue,
		Type:        jobType,
		Payload:     input.Payload,
		Status:      JobStatusPending,
		Attempts:    0,
		MaxAttempts: maxAttempts,
		RunAt:       runAt,
	}

	createdJob, err := s.repository.Create(
		ctx,
		job,
	)
	if err != nil {
		return Job{}, fmt.Errorf(
			"create job: %w",
			err,
		)
	}

	return createdJob, nil
}


func (s *Service) CancelJob(
    ctx context.Context,
    id uuid.UUID,
) (Job, error) {

	job, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return Job{}, fmt.Errorf(
			"get job by id: %w",
			err,
		)
	}

	if job.Status != JobStatusPending {
		return Job{}, ErrJobNotPending
	}

	cancelledJob, err := s.repository.Cancel(ctx, id)
	if err != nil {
		return Job{}, fmt.Errorf(
			"cancel job: %w",
			err,
		)
	}

	return cancelledJob, nil
}
