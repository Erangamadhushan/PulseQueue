package health

import (
	"github.com/Erangamadhushan/pulsequeue/internal/httpx"
	"net/http"
)

type Response struct {
	Status string `json:"status"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
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

	if err := httpx.WriteJSON(
		w,
		http.StatusOK,
		Response{Status: "ok"},
	); err != nil {
		// A response-write failure cannot reliably be repaired
		// once the HTTP response has started.
		return
	}
}
