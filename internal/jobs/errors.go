package jobs

import (
	"errors"
	"net/http"

	"github.com/Erangamadhushan/pulsequeue/internal/httpx"
)

var (
	ErrJobNotFound = errors.New("job not found")

	// Validation errors
	ErrInvalidJobType     = errors.New("invalid job type")
	ErrInvalidJobPayload  = errors.New("invalid job payload")
	ErrInvalidMaxAttempts = errors.New("invalid max attempts")
	ErrInvalidListLimit   = errors.New("invalid list limit")
	ErrInvalidListOffset  = errors.New("invalid list offset")
	ErrInvalidJobStatus   = errors.New("invalid job status")
	ErrInvalidJobQueue    = errors.New("invalid job queue")
	ErrJobNotPending      = errors.New("job is not pending")

	// Job cancellation errors
	ErrJobNotCancellable = errors.New("job cannot be cancelled")
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteJSONError(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
) {
	response := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}

	_ = httpx.WriteJSON(w, status, response)
}

func (h *Handler) handleCreateError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, ErrInvalidJobType):
		WriteJSONError(
			w,
			http.StatusBadRequest,
			"invalid_job_type",
			"job type is required",
		)

	case errors.Is(err, ErrInvalidJobPayload):
		WriteJSONError(
			w,
			http.StatusBadRequest,
			"invalid_job_payload",
			"payload must be a JSON object",
		)

	case errors.Is(err, ErrInvalidMaxAttempts):
		WriteJSONError(
			w,
			http.StatusBadRequest,
			"invalid_max_attempts",
			"max_attempts must be greater than zero",
		)

	case errors.Is(err, ErrInvalidListLimit),
		errors.Is(err, ErrInvalidListOffset),
		errors.Is(err, ErrInvalidJobStatus),
		errors.Is(err, ErrInvalidJobQueue):
		WriteJSONError(
			w,
			http.StatusBadRequest,
			"invalid_filter",
			"the job list filter is invalid",
		)

	case errors.Is(err, ErrJobNotFound):
		WriteJSONError(
			w,
			http.StatusNotFound,
			"job_not_found",
			"job not found",
		)

	case errors.Is(err, ErrJobNotPending),
		errors.Is(err, ErrJobNotCancellable):
		WriteJSONError(
			w,
			http.StatusConflict,
			"job_not_cancellable",
			"job cannot be cancelled",
		)

	default:
		WriteJSONError(
			w,
			http.StatusInternalServerError,
			"internal_server_error",
			"internal server error",
		)
	}
}
