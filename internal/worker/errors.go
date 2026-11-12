package worker

import "errors"

var ErrWorkerStopped = errors.New("worker stopped")

var ErrWorkerShutdown = errors.New("worker shutdown interrupted execution")
