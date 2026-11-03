package jobs

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type JobStatus string


type Job struct {
	ID uuid.UUID

	Queue string
	Type  string

	Payload json.RawMessage

	Status JobStatus

	Attempts    int
	MaxAttempts int

	RunAt time.Time

	LastError *string

	CreatedAt time.Time
	UpdatedAt time.Time

	StartedAt   *time.Time
	CompletedAt *time.Time
}

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

func (s JobStatus) IsValid() bool {
	switch s {
	case JobStatusPending,
		JobStatusRunning,
		JobStatusCompleted,
		JobStatusFailed,
		JobStatusCancelled:
		return true
	default:
		return false
	}
}

func (s JobStatus) IsTerminal() bool {
	switch s {
	case JobStatusCompleted,
		JobStatusFailed,
		JobStatusCancelled:
		return true
	default:
		return false
	}
}

func (s JobStatus) CanTransitionTo(next JobStatus) bool {
	switch s {
	case JobStatusPending:
		return next == JobStatusRunning ||
			next == JobStatusCancelled

	case JobStatusRunning:
		return next == JobStatusCompleted ||
			next == JobStatusPending ||
			next == JobStatusFailed

	default:
		return false
	}
}

func ToJobResponse(job Job) JobResponse {
	return JobResponse{
		ID:          job.ID.String(),
		Queue:       job.Queue,
		Type:        job.Type,
		Payload:     job.Payload,
		Status:      job.Status,
		Attempts:    job.Attempts,
		MaxAttempts: job.MaxAttempts,
		RunAt:       job.RunAt,

		LastError: job.LastError,

		CreatedAt: job.CreatedAt,
		UpdatedAt: job.UpdatedAt,

		StartedAt:   job.StartedAt,
		CompletedAt: job.CompletedAt,
	}
}	

