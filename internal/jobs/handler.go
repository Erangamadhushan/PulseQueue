package jobs

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Erangamadhushan/pulsequeue/internal/httpx"
)

type CreateJobRequest struct {
	Queue       string          `json:"queue"`
	Type        string          `json:"type"`
	Payload     json.RawMessage `json:"payload"`
	MaxAttempts int             `json:"max_attempts"`
	RunAt       *time.Time      `json:"run_at"`
}

type JobResponse struct {
	ID          string          `json:"id"`
	Queue       string          `json:"queue"`
	Type        string          `json:"type"`
	Payload     json.RawMessage `json:"payload"`
	Status      JobStatus       `json:"status"`
	Attempts    int             `json:"attempts"`
	MaxAttempts int             `json:"max_attempts"`
	RunAt       time.Time       `json:"run_at"`

	LastError *string `json:"last_error,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)

		WriteJSONError(
			w,
			http.StatusMethodNotAllowed,
			"method_not_allowed",
			"Only POST requests are supported.",
		)
		return
	}

	var request CreateJobRequest

	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		WriteJSONError(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"The request body is invalid.",
		)
		return
	}

	job, err := h.service.CreateJob(
		r.Context(),
		CreateJobInput{
			Queue:       request.Queue,
			Type:        request.Type,
			Payload:     request.Payload,
			MaxAttempts: request.MaxAttempts,
			RunAt:       request.RunAt,
		},
	)

	if err != nil {
		h.handleCreateError(w, err)
		return
	}

	if err := httpx.WriteJSON(
		w,
		http.StatusCreated,
		ToJobResponse(job),
	); err != nil {
		// Log response writing failure.
		return
	}
}
