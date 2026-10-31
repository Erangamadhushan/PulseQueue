package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type exampleRequest struct {
	Name string `json:"name"`
}

func TestDecodeJSONValidRequest(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/example",
		strings.NewReader(`{"name":"PulseQueue"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	var body exampleRequest
	if err := DecodeJSON(rec, req, &body); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}

	if body.Name != "PulseQueue" {
		t.Errorf("expected PulseQueue, got %q", body.Name)
	}
}

func TestDecodeJSONRejectsUnknownFields(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/example",
		strings.NewReader(`{"name":"PulseQueue","extra":true}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	var body exampleRequest
	if err := DecodeJSON(rec, req, &body); err == nil {
		t.Fatal("expected unknown field to be rejected")
	}
}

func TestDecodeJSONRejectsTrailingValues(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/example",
		strings.NewReader(`{"name":"PulseQueue"} {"name":"another"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	var body exampleRequest
	if err := DecodeJSON(rec, req, &body); err == nil {
		t.Fatal("expected trailing JSON value to be rejected")
	}
}

func TestDecodeJSONRejectsWrongContentType(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/example",
		strings.NewReader(`{"name":"PulseQueue"}`),
	)
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	var body exampleRequest
	if err := DecodeJSON(rec, req, &body); err == nil {
		t.Fatal("expected unsupported content type to be rejected")
	}
}

func TestDecodeJSONRejectsOversizedBody(t *testing.T) {
	payload := `{"name":"` +
		strings.Repeat("a", int(MaxJSONBodyBytes)) +
		`"}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/example",
		strings.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	var body exampleRequest
	if err := DecodeJSON(rec, req, &body); err == nil {
		t.Fatal("expected oversized body to be rejected")
	}
}
