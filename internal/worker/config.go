package worker

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type WorkerConfig struct {
	ID              string
	Queues          []string
	PollInterval    time.Duration
	LeaseDuration   time.Duration
	ShutdownTimeout time.Duration
}

func LoadWorkerConfig() (WorkerConfig, error) {
	rawID := os.Getenv("PULSEQUEUE_WORKER_ID")
	id := strings.TrimSpace(rawID)
	if rawID == "" {
		id = "worker-" + uuid.NewString()
	} else if id == "" {
		return WorkerConfig{}, fmt.Errorf("invalid PULSEQUEUE_WORKER_ID: worker ID cannot be empty")
	}

	queues, err := parseQueues(os.Getenv("PULSEQUEUE_WORKER_QUEUES"))
	if err != nil {
		return WorkerConfig{}, err
	}
	pollInterval, err := durationEnv("PULSEQUEUE_WORKER_POLL_INTERVAL", 2*time.Second)
	if err != nil {
		return WorkerConfig{}, err
	}
	leaseDuration, err := durationEnv("PULSEQUEUE_WORKER_LEASE_DURATION", time.Minute)
	if err != nil {
		return WorkerConfig{}, err
	}
	shutdownTimeout, err := durationEnv("PULSEQUEUE_WORKER_SHUTDOWN_TIMEOUT", 30*time.Second)
	if err != nil {
		return WorkerConfig{}, err
	}
	if pollInterval <= 0 || leaseDuration <= 0 || shutdownTimeout <= 0 {
		return WorkerConfig{}, fmt.Errorf("worker durations must be positive")
	}

	return WorkerConfig{
		ID:              id,
		Queues:          queues,
		PollInterval:    pollInterval,
		LeaseDuration:   leaseDuration,
		ShutdownTimeout: shutdownTimeout,
	}, nil
}

func parseQueues(value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return []string{"default"}, nil
	}

	queues := make([]string, 0)
	seen := make(map[string]struct{})
	for _, rawQueue := range strings.Split(value, ",") {
		queue := strings.TrimSpace(rawQueue)
		if queue == "" {
			return nil, fmt.Errorf("invalid PULSEQUEUE_WORKER_QUEUES: queue names cannot be empty")
		}
		if _, exists := seen[queue]; exists {
			continue
		}
		seen[queue] = struct{}{}
		queues = append(queues, queue)
	}
	return queues, nil
}

func durationEnv(name string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", name, value, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("invalid %s %q: duration must be positive", name, value)
	}
	return duration, nil
}
