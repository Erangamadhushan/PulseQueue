package worker

import (
	"strings"
	"testing"
	"time"
)

func clearWorkerEnv(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"PULSEQUEUE_WORKER_ID",
		"PULSEQUEUE_WORKER_QUEUES",
		"PULSEQUEUE_WORKER_POLL_INTERVAL",
		"PULSEQUEUE_WORKER_LEASE_DURATION",
		"PULSEQUEUE_WORKER_SHUTDOWN_TIMEOUT",
	} {
		t.Setenv(name, "")
	}
}

func TestLoadWorkerConfigDefaults(t *testing.T) {
	clearWorkerEnv(t)
	cfg, err := LoadWorkerConfig()
	if err != nil {
		t.Fatalf("LoadWorkerConfig() error = %v", err)
	}
	if !strings.HasPrefix(cfg.ID, "worker-") {
		t.Fatalf("ID = %q, want generated worker ID", cfg.ID)
	}
	if cfg.PollInterval != 2*time.Second || cfg.LeaseDuration != time.Minute ||
		cfg.ShutdownTimeout != 30*time.Second {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if len(cfg.Queues) != 1 || cfg.Queues[0] != "default" {
		t.Fatalf("Queues = %v, want [default]", cfg.Queues)
	}
}

func TestLoadWorkerConfigParsesAndNormalizesValues(t *testing.T) {
	clearWorkerEnv(t)
	t.Setenv("PULSEQUEUE_WORKER_ID", " worker-01 ")
	t.Setenv("PULSEQUEUE_WORKER_QUEUES", " default, emails,default ")
	t.Setenv("PULSEQUEUE_WORKER_POLL_INTERVAL", "500ms")
	t.Setenv("PULSEQUEUE_WORKER_LEASE_DURATION", "2m")
	t.Setenv("PULSEQUEUE_WORKER_SHUTDOWN_TIMEOUT", "45s")

	cfg, err := LoadWorkerConfig()
	if err != nil {
		t.Fatalf("LoadWorkerConfig() error = %v", err)
	}
	if cfg.ID != "worker-01" || cfg.PollInterval != 500*time.Millisecond ||
		cfg.LeaseDuration != 2*time.Minute || cfg.ShutdownTimeout != 45*time.Second {
		t.Fatalf("unexpected configuration: %+v", cfg)
	}
	if len(cfg.Queues) != 2 || cfg.Queues[0] != "default" || cfg.Queues[1] != "emails" {
		t.Fatalf("Queues = %v, want [default emails]", cfg.Queues)
	}
}

func TestLoadWorkerConfigRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		env   string
		value string
	}{
		{"poll interval", "PULSEQUEUE_WORKER_POLL_INTERVAL", "not-a-duration"},
		{"lease duration", "PULSEQUEUE_WORKER_LEASE_DURATION", "0s"},
		{"shutdown timeout", "PULSEQUEUE_WORKER_SHUTDOWN_TIMEOUT", "-1s"},
		{"queue name", "PULSEQUEUE_WORKER_QUEUES", "default,,emails"},
		{"worker ID", "PULSEQUEUE_WORKER_ID", "   "},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clearWorkerEnv(t)
			t.Setenv(test.env, test.value)
			if _, err := LoadWorkerConfig(); err == nil {
				t.Fatalf("LoadWorkerConfig() error = nil, want validation error")
			}
		})
	}
}
