package health

import (
	"context"
	"net/http"
	"time"

	"github.com/Erangamadhushan/pulsequeue/internal/httpx"
)

type Pinger interface {
	Ping(context.Context) error
}

func ReadinessHandler(db Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)

			httpx.WriteError(
				w,
				http.StatusMethodNotAllowed,
				"method_not_allowed",
				"Only GET requests are supported.",
			)
			return
		}

		ctx, cancel := context.WithTimeout(
			r.Context(),
			2*time.Second,
		)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			httpx.WriteError(
				w,
				http.StatusServiceUnavailable,
				"service_unavailable",
				"Service dependencies are unavailable.",
			)
			return
		}

		_ = httpx.WriteJSON(
			w,
			http.StatusOK,
			Response{Status: "ok"},
		)
	}
}
