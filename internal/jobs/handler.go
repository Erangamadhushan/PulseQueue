package jobs

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Erangamadhushan/pulsequeue/internal/httpx"
	"github.com/google/uuid"
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

func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)

		WriteJSONError(
			w,
			http.StatusMethodNotAllowed,
			"method_not_allowed",
			"Only GET requests are supported.",
		)
		return
	}

	id, err := uuid.Parse(strings.TrimPrefix(r.URL.Path, "/jobs/"))
	if err != nil {
		WriteJSONError(
			w,
			http.StatusBadRequest,
			"invalid_id",
			"The job ID is invalid.",
		)
		return
	}

	job, err := h.service.GetJob(r.Context(), id)
	if err != nil {
		h.handleCreateError(w, err)
		return
	}

	if err := httpx.WriteJSON(
		w,
		http.StatusOK,
		ToJobResponse(job),
	); err != nil {
		// Log response writing failure.
		return
	}
}

func (h *Handler) ListJobs(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)

		WriteJSONError(
			w,
			http.StatusMethodNotAllowed,
			"method_not_allowed",
			"Only GET requests are supported.",
		)
		return
	}

	filter, err := parseListJobsFilter(r)
	if err != nil {
		h.handleCreateError(w, err)
		return
	}

	jobs, err := h.service.ListJobs(r.Context(), filter)
	if err != nil {
		h.handleCreateError(w, err)
		return
	}

	if err := httpx.WriteJSON(
		w,
		http.StatusOK,
		toJobResponses(jobs),
	); err != nil {
		// Log response writing failure.
		return
	}
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		WriteJSONError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET requests are supported.")
		return
	}

	stats, err := h.service.GetStats(r.Context())
	if err != nil {
		h.handleCreateError(w, err)
		return
	}

	if err := httpx.WriteJSON(w, http.StatusOK, stats); err != nil {
		return
	}
}

func (h *Handler) CancelJob(
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

	id, err := uuid.Parse(strings.TrimSuffix(
		strings.TrimPrefix(r.URL.Path, "/jobs/"),
		"/cancel",
	))
	if err != nil {
		WriteJSONError(
			w,
			http.StatusBadRequest,
			"invalid_id",
			"The job ID is invalid.",
		)
		return
	}

	job, err := h.service.CancelJob(r.Context(), id)
	if err != nil {
		h.handleCreateError(w, err)
		return
	}

	if err := httpx.WriteJSON(
		w,
		http.StatusOK,
		ToJobResponse(job),
	); err != nil {
		// Log response writing failure.
		return
	}
}

func parseListJobsFilter(r *http.Request) (ListJobsFilter, error) {
	filter := ListJobsFilter{
		Limit: DefaultListLimit,
	}
	query := r.URL.Query()

	if value := query.Get("limit"); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil {
			return ListJobsFilter{}, ErrInvalidListLimit
		}
		filter.Limit = limit
	}
	if value := query.Get("offset"); value != "" {
		offset, err := strconv.Atoi(value)
		if err != nil {
			return ListJobsFilter{}, ErrInvalidListOffset
		}
		filter.Offset = offset
	}
	if value := query.Get("status"); value != "" {
		status := JobStatus(value)
		filter.Status = &status
	}
	if value := query.Get("queue"); value != "" {
		queue := strings.TrimSpace(value)
		filter.Queue = &queue
	}

	return filter, nil
}

func toJobResponses(jobs []Job) []JobResponse {
	responses := make([]JobResponse, 0, len(jobs))
	for _, job := range jobs {
		responses = append(responses, ToJobResponse(job))
	}
	return responses
}
