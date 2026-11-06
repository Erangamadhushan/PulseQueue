package worker

import (
	"context"
	"fmt"

	"github.com/Erangamadhushan/pulsequeue/internal/jobs"
)

type Executor struct {
	registry *Registry
}

func NewExecutor(registry *Registry) *Executor {
	return &Executor{registry: registry}
}

func (e *Executor) Execute(ctx context.Context, job jobs.Job) (err error) {
	handler, err := e.registry.Resolve(job.Type)
	if err != nil {
		return err
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("handler panic: %v", recovered)
		}
	}()
	return handler(ctx, job)
}
