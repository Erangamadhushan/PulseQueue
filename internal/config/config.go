package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Config struct {
	Host                  string
	Port                  string
	DatabaseURL           string
	APIToken              string
	FrontendOrigin        string
	WorkerID              string
	WorkerPollInterval    time.Duration
	WorkerLeaseDuration   time.Duration
	WorkerShutdownTimeout time.Duration
	WorkerQueues          []string
}

func LoadConfig() (Config, error) {
	port := strings.TrimSpace(os.Getenv("PULSEQUEUE_PORT"))

	if port == "" {
		port = "8080"
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return Config{}, fmt.Errorf(
			"invalid PULSEQUEUE_PORT %q: must be between 1 and 65535",
			port,
		)
	}

	host := strings.TrimSpace(os.Getenv("PULSEQUEUE_HOST"))
	if host == "" {
		host = "127.0.0.1"
	}

	apiToken := strings.TrimSpace(os.Getenv("PULSEQUEUE_API_TOKEN"))
	if !isLoopbackHost(host) && apiToken == "" {
		return Config{}, fmt.Errorf(
			"PULSEQUEUE_API_TOKEN is required when PULSEQUEUE_HOST is not loopback",
		)
	}

	frontendOrigin := strings.TrimSpace(os.Getenv("PULSEQUEUE_FRONTEND_ORIGIN"))
	if frontendOrigin != "" {
		origin, err := url.Parse(frontendOrigin)
		if err != nil ||
			(origin.Scheme != "http" && origin.Scheme != "https") ||
			origin.Host == "" ||
			origin.Path != "" ||
			origin.RawQuery != "" ||
			origin.Fragment != "" ||
			origin.User != nil {
			return Config{}, fmt.Errorf(
				"invalid PULSEQUEUE_FRONTEND_ORIGIN %q: must be an HTTP(S) origin",
				frontendOrigin,
			)
		}
	}

	databaseURL := strings.TrimSpace(
		os.Getenv("PULSEQUEUE_DATABASE_URL"),
	)

	if databaseURL == "" {
		return Config{}, fmt.Errorf(
			"PULSEQUEUE_DATABASE_URL is required",
		)
	}

	workerID := strings.TrimSpace(os.Getenv("PULSEQUEUE_WORKER_ID"))
	if workerID == "" {
		workerID = "worker-" + uuid.NewString()
	}
	pollInterval, err := durationEnv("PULSEQUEUE_WORKER_POLL_INTERVAL", time.Second)
	if err != nil {
		return Config{}, err
	}
	leaseDuration, err := durationEnv("PULSEQUEUE_WORKER_LEASE_DURATION", time.Minute)
	if err != nil {
		return Config{}, err
	}
	shutdownTimeout, err := durationEnv("PULSEQUEUE_WORKER_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	if pollInterval <= 0 || leaseDuration <= 0 || shutdownTimeout <= 0 {
		return Config{}, fmt.Errorf("worker durations must be positive")
	}
	queues := parseQueues(os.Getenv("PULSEQUEUE_WORKER_QUEUES"))

	return Config{
		Host:                  host,
		Port:                  port,
		DatabaseURL:           databaseURL,
		APIToken:              apiToken,
		FrontendOrigin:        frontendOrigin,
		WorkerID:              workerID,
		WorkerPollInterval:    pollInterval,
		WorkerLeaseDuration:   leaseDuration,
		WorkerShutdownTimeout: shutdownTimeout,
		WorkerQueues:          queues,
	}, nil
}

func parseQueues(value string) []string {
	queues := make([]string, 0)
	for _, queue := range strings.Split(value, ",") {
		queue = strings.TrimSpace(queue)
		if queue != "" {
			queues = append(queues, queue)
		}
	}
	if len(queues) == 0 {
		return []string{"default"}
	}
	return queues
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
	return duration, nil
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}

	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// os.Getenv reads an environment variable.
// strings.TrimSpace removes accidental spaces.
// strconv.Atoi converts the port string to an integer.
// Returning (Config, error) lets the caller handle invalid configuration explicitly.
// Config{} is returned when configuration is invalid, avoiding a partially configured result.
