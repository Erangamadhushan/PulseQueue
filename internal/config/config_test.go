package config

import "testing"

func TestLoadDefaultsToPort8080(t *testing.T) {
	t.Setenv("PULSEQUEUE_PORT", "")
	t.Setenv("PULSEQUEUE_HOST", "")
	t.Setenv("PULSEQUEUE_API_TOKEN", "")
	t.Setenv("PULSEQUEUE_FRONTEND_ORIGIN", "")
	t.Setenv("PULSEQUEUE_DATABASE_URL", "postgres://localhost/pulsequeue")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("expected port 8080, got %q", cfg.Port)
	}
}

func TestLoadRejectsNonLoopbackHostWithoutAPIToken(t *testing.T) {
	t.Setenv("PULSEQUEUE_HOST", "0.0.0.0")
	t.Setenv("PULSEQUEUE_API_TOKEN", "")
	t.Setenv("PULSEQUEUE_DATABASE_URL", "postgres://localhost/pulsequeue")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected an error when a non-loopback host has no API token")
	}
}

func TestLoadAcceptsConfiguredFrontendOriginAndAPIToken(t *testing.T) {
	t.Setenv("PULSEQUEUE_HOST", "0.0.0.0")
	t.Setenv("PULSEQUEUE_API_TOKEN", "test-token")
	t.Setenv("PULSEQUEUE_FRONTEND_ORIGIN", "https://dashboard.example.com")
	t.Setenv("PULSEQUEUE_DATABASE_URL", "postgres://localhost/pulsequeue")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.APIToken != "test-token" || cfg.FrontendOrigin != "https://dashboard.example.com" {
		t.Fatalf("unexpected security configuration: %+v", cfg)
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
