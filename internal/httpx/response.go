package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

func WriteError(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
) {
	err := WriteJSON(w, status, ErrorResponse{
		Error: APIError{
			Code:    code,
			Message: message,
		},
	})

	if err != nil {
		slog.Error("failed to write API error response",
			"error", err,
			"code", code,
			"status", status,
		)
	}
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	WriteError(
		w,
		http.StatusNotFound,
		"not_found",
		"The requested resource was not found.",
	)
}

/*

The any type is Go's alias for interface{}. It allows WriteJSON to serialize different response structs without requiring a separate function for each type.
The APIError struct separates the machine-readable error code from the human-readable message. Your frontend can use the code to decide how to handle an error without parsing its message.

One limitation: a response-write failure may occur after headers have already been sent, so logging it is more reliable than trying to send another response.

*/
