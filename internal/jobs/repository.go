package jobs

import (
	"context"

	"github.com/google/uuid"
)

type ListJobsFilter struct {
	Limit  int
	Offset int
	Status *JobStatus
	Queue  *string
}

type JobStats struct {
	Total     int `json:"total"`
	Pending   int `json:"pending"`
	Running   int `json:"running"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
	Cancelled int `json:"cancelled"`
}

type Repository interface {
	Create(ctx context.Context, job Job) (Job, error)
	GetByID(ctx context.Context, id uuid.UUID) (Job, error)
	List(ctx context.Context, filter ListJobsFilter) ([]Job, error)
	Cancel(ctx context.Context, id uuid.UUID) (Job, error)
	Stats(ctx context.Context) (JobStats, error)
}
