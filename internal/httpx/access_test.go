package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIMiddlewareRequiresTokenForProtectedRoutes(t *testing.T) {
	handler := APIMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), "test-token", "https://dashboard.example.com")

	request := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}

	request.Header.Set("Authorization", "Bearer test-token")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", response.Code)
	}
}

func TestAPIMiddlewareHandlesAllowedPreflight(t *testing.T) {
	handler := APIMiddleware(http.NotFoundHandler(), "test-token", "https://dashboard.example.com")
	request := httptest.NewRequest(http.MethodOptions, "/jobs", nil)
	request.Header.Set("Origin", "https://dashboard.example.com")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	request.Header.Set("Access-Control-Request-Headers", "authorization, content-type")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", response.Code)
	}
	if response.Header().Get("Access-Control-Allow-Origin") != "https://dashboard.example.com" {
		t.Fatal("expected the configured origin to be allowed")
	}
	if response.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Fatal("expected allowed headers to be present")
	}
}
