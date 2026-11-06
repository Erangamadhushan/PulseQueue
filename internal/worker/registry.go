package worker

import (
	"errors"
	"strings"
	"sync"
)

var (
	ErrEmptyHandlerName = errors.New("handler name is required")
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
	r.mu.RLock()
	handler, exists := r.handlers[name]
	r.mu.RUnlock()
	if !exists {
		return nil, ErrHandlerNotFound
	}
	return handler, nil
}
