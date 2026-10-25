package config

import "testing"

func TestLoadDefaultsToPort8080(t *testing.T) {
	t.Setenv("PULSEQUEUE_PORT", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("expected port 8080, got %q", cfg.Port)
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	t.Setenv("PULSEQUEUE_PORT", "70000")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected an error for an invalid port")
	}
}

/*

Run the complete test suite:

go test ./...
go vet ./...

go vet checks for suspicious code patterns that the compiler may not report. It complements your tests rather than replacing them.

*/
