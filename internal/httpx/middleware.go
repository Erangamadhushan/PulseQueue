package httpx

import (
	"log/slog"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(status int) {
	slog.Debug("Writing header", "status", status)
	if !r.wroteHeader {
		r.status = status
		r.wroteHeader = true
		r.ResponseWriter.WriteHeader(status)
	} else {
		slog.Warn("Attempted to write header multiple times", "status", status)
		return
	}
}

func (r *statusRecorder) Write(data []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}

	return r.ResponseWriter.Write(data)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		slog.Debug("Request completed", "method", r.Method, "url", r.URL, "status", recorder.status, "duration", time.Since(start))

	})
}

/*

Middleware wraps an HTTP handler and adds behavior before or after it runs. We can use the same pattern later for authentication, rate limiting, request IDs, and metrics.

The statusRecorder captures the HTTP status code so that the logger can record what actually happened. If a handler writes a response body without explicitly setting a status, the default status is 200 OK.

A production-grade middleware must also consider optional http.ResponseWriter interfaces such as http.Flusher and http.Hijacker. Our recorder is sufficient for these initial ordinary JSON endpoints, but we'll revisit interface preservation if PulseQueue needs streaming or upgraded connections.

*/
