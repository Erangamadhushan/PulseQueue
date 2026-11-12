package worker

import (
	"context"

	"github.com/Erangamadhushan/pulsequeue/internal/jobs"
)

type Handler func(context.Context, jobs.Job) error
