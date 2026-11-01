package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakePinger struct {
	err error
}

func (f fakePinger) Ping(context.Context) error {
	return f.err
}

func TestReadinessHandlerReturnsOK(t *testing.T) {
	handler := ReadinessHandler(
		fakePinger{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/readyz",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestReadinessHandlerReturnsUnavailable(t *testing.T) {
	handler := ReadinessHandler(
		fakePinger{
			err: errors.New("database unavailable"),
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/readyz",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusServiceUnavailable,
			rec.Code,
		)
	}
}
