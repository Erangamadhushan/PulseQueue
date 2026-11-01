package jobs

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, job Job) (Job, error)

	GetByID(
		ctx context.Context,
		id uuid.UUID,
	) (Job, error)

	List(
		ctx context.Context,
	) ([]Job, error)
}
