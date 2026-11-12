package worker

import (
	"errors"
	"strings"
	"sync"
)

var (
	ErrEmptyJobType     = errors.New("job type cannot be empty")
	ErrEmptyHandlerName = ErrEmptyJobType
	ErrNilHandler       = errors.New("handler cannot be nil")
	ErrDuplicateHandler = errors.New("handler is already registered")
	ErrHandlerNotFound  = errors.New("handler is not registered")
)

type Registry struct {
	mu       sync.RWMutex
	handlers map[string]Handler
}

func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]Handler)}
}

func (r *Registry) Register(name string, handler Handler) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyHandlerName
	}
	if handler == nil {
		return ErrNilHandler
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.handlers[name]; exists {
		return ErrDuplicateHandler
	}
	r.handlers[name] = handler
	return nil
}

func (r *Registry) Resolve(name string) (Handler, error) {
	handler, exists := r.Get(name)
	if !exists {
		return nil, ErrHandlerNotFound
	}
	return handler, nil
}

func (r *Registry) Get(jobType string) (Handler, bool) {
	jobType = strings.TrimSpace(jobType)
	r.mu.RLock()
	handler, exists := r.handlers[jobType]
	r.mu.RUnlock()
	return handler, exists
}

func (r *Registry) Has(jobType string) bool {
	_, exists := r.Get(jobType)
	return exists
}
