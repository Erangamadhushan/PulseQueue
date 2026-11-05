package jobs

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCreateJobHandlerSuccess(t *testing.T) {
	runAt := time.Date(
		2026,
		time.November,
		3,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	repository := &fakeRepository{}
	service := NewService(repository)
	service.now = func() time.Time {
		return runAt
	}

	handler := NewHandler(service)
	request := httptest.NewRequest(
		http.MethodPost,
		"/jobs",
		strings.NewReader(
			`{"queue":"emails","type":"send_email","payload":{"to":"test@example.com"},"max_attempts":5}`,
		),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusCreated,
		)
	}

	if got := recorder.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want application/json; charset=utf-8", got)
	}

	var response JobResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Queue != "emails" {
		t.Errorf("Queue = %q, want emails", response.Queue)
	}
	if response.Type != "send_email" {
		t.Errorf("Type = %q, want send_email", response.Type)
	}
	if string(response.Payload) != `{"to":"test@example.com"}` {
		t.Errorf("Payload = %s, want %s", response.Payload, `{"to":"test@example.com"}`)
	}
	if response.Status != JobStatusPending {
		t.Errorf("Status = %q, want %q", response.Status, JobStatusPending)
	}
	if response.MaxAttempts != 5 {
		t.Errorf("MaxAttempts = %d, want 5", response.MaxAttempts)
	}
	if !response.RunAt.Equal(runAt) {
		t.Errorf("RunAt = %v, want %v", response.RunAt, runAt)
	}

	if repository.createdJob.Type != "send_email" {
		t.Errorf("repository received Type = %q, want send_email", repository.createdJob.Type)
	}
}

func TestCreateJobHandlerRejectsUnsupportedMethod(t *testing.T) {
	handler := NewHandler(nil)
	request := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusMethodNotAllowed,
		)
	}
	if got := recorder.Header().Get("Allow"); got != http.MethodPost {
		t.Errorf("Allow = %q, want %q", got, http.MethodPost)
	}

	assertErrorResponse(
		t,
		recorder,
		"method_not_allowed",
		"Only POST requests are supported.",
	)
}

func TestCreateJobHandlerRejectsInvalidJSON(t *testing.T) {
	handler := NewHandler(nil)
	request := newCreateJobRequest(`{"type":`)
	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	assertErrorResponse(
		t,
		recorder,
		"invalid_request",
		"The request body is invalid.",
	)
}

func TestCreateJobHandlerRejectsInvalidEncoding(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "wrong content type",
			body: `{"type":"send_email","payload":{}}`,
		},
		{
			name: "unknown field",
			body: `{"type":"send_email","payload":{},"unexpected":true}`,
		},
		{
			name: "trailing JSON",
			body: `{"type":"send_email","payload":{}} {}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(nil)
			request := newCreateJobRequest(tt.body)
			if tt.name == "wrong content type" {
				request.Header.Set("Content-Type", "text/plain")
			}
			recorder := httptest.NewRecorder()

			handler.Create(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf(
					"status = %d, want %d",
					recorder.Code,
					http.StatusBadRequest,
				)
			}
			assertErrorResponse(
				t,
				recorder,
				"invalid_request",
				"The request body is invalid.",
			)
		})
	}
}

func TestCreateJobHandlerRejectsInvalidJob(t *testing.T) {
	tests := []struct {
		name string
		body string
		code string
	}{
		{
			name: "missing type",
			body: `{"payload":{}}`,
			code: "invalid_job_type",
		},
		{
			name: "invalid payload",
			body: `{"type":"send_email"}`,
			code: "invalid_job_payload",
		},
		{
			name: "negative max attempts",
			body: `{"type":"send_email","payload":{},"max_attempts":-1}`,
			code: "invalid_max_attempts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(NewService(&fakeRepository{}))
			request := newCreateJobRequest(tt.body)
			recorder := httptest.NewRecorder()

			handler.Create(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf(
					"status = %d, want %d",
					recorder.Code,
					http.StatusBadRequest,
				)
			}
			assertErrorCode(t, recorder, tt.code)
		})
	}
}

func TestCreateJobHandlerReturnsRepositoryError(t *testing.T) {
	repositoryError := errors.New("database unavailable")
	handler := NewHandler(NewService(&fakeRepository{
		createErr: repositoryError,
	}))
	request := newCreateJobRequest(
		`{"type":"send_email","payload":{}}`,
	)
	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusInternalServerError,
		)
	}
	assertErrorResponse(
		t,
		recorder,
		"internal_server_error",
		"internal server error",
	)
}

func TestGetJobHandlerRejectsInvalidUUID(t *testing.T) {
	handler := NewHandler(nil)
	request := httptest.NewRequest(http.MethodGet, "/jobs/not-a-uuid", nil)
	recorder := httptest.NewRecorder()

	handler.GetJob(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	assertErrorCode(t, recorder, "invalid_id")
}

func TestCancelJobHandlerSuccess(t *testing.T) {
	id := uuid.New()
	repository := &fakeRepository{
		job:       Job{ID: id, Status: JobStatusPending},
		cancelJob: Job{ID: id, Status: JobStatusCancelled},
	}
	handler := NewHandler(NewService(repository))
	request := httptest.NewRequest(http.MethodPost, "/jobs/"+id.String()+"/cancel", nil)
	recorder := httptest.NewRecorder()

	handler.CancelJob(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var response JobResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Status != JobStatusCancelled {
		t.Errorf("status = %q, want %q", response.Status, JobStatusCancelled)
	}
}

func TestCancelJobHandlerErrors(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		repository *fakeRepository
		wantStatus int
		wantCode   string
	}{
		{
			name:       "invalid UUID",
			target:     "/jobs/not-a-uuid/cancel",
			repository: &fakeRepository{},
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_id",
		},
		{
			name:       "missing job",
			target:     "/jobs/" + uuid.NewString() + "/cancel",
			repository: &fakeRepository{},
			wantStatus: http.StatusNotFound,
			wantCode:   "job_not_found",
		},
		{
			name:   "invalid state",
			target: "/jobs/" + uuid.NewString() + "/cancel",
			repository: &fakeRepository{
				job: Job{Status: JobStatusRunning},
			},
			wantStatus: http.StatusConflict,
			wantCode:   "job_not_cancellable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(NewService(tt.repository))
			request := httptest.NewRequest(http.MethodPost, tt.target, nil)
			recorder := httptest.NewRecorder()

			handler.CancelJob(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			assertErrorCode(t, recorder, tt.wantCode)
		})
	}
}

func TestListJobsHandlerParsesQuery(t *testing.T) {
	repository := &fakeRepository{
		listedJobs: []Job{{
			ID:      uuid.New(),
			Queue:   "emails",
			Type:    "send_email",
			Payload: json.RawMessage(`{}`),
			Status:  JobStatusPending,
			RunAt:   time.Date(2026, time.November, 3, 12, 0, 0, 0, time.UTC),
		}},
	}
	handler := NewHandler(NewService(repository))
	request := httptest.NewRequest(
		http.MethodGet,
		"/jobs?limit=2&offset=1&status=pending&queue=%20emails%20",
		nil,
	)
	recorder := httptest.NewRecorder()

	handler.ListJobs(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if repository.listFilter.Limit != 2 || repository.listFilter.Offset != 1 {
		t.Fatalf("filter pagination = %+v, want limit 2 offset 1", repository.listFilter)
	}
	if repository.listFilter.Status == nil || *repository.listFilter.Status != JobStatusPending {
		t.Fatalf("status filter = %v, want pending", repository.listFilter.Status)
	}
	if repository.listFilter.Queue == nil || *repository.listFilter.Queue != "emails" {
		t.Fatalf("queue filter = %v, want emails", repository.listFilter.Queue)
	}

	var response []JobResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response) != 1 || response[0].ID != repository.listedJobs[0].ID.String() {
		t.Fatalf("response = %+v, want one mapped job", response)
	}
}

func TestListJobsHandlerRejectsInvalidQuery(t *testing.T) {
	tests := []string{
		"/jobs?limit=not-a-number",
		"/jobs?offset=-1",
		"/jobs?status=unknown",
		"/jobs?queue=%20",
	}

	for _, target := range tests {
		t.Run(target, func(t *testing.T) {
			handler := NewHandler(NewService(&fakeRepository{}))
			request := httptest.NewRequest(http.MethodGet, target, nil)
			recorder := httptest.NewRecorder()

			handler.ListJobs(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
			}
			assertErrorCode(t, recorder, "invalid_filter")
		})
	}
}

func newCreateJobRequest(body string) *http.Request {
	request := httptest.NewRequest(
		http.MethodPost,
		"/jobs",
		strings.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")
	return request
}

func assertErrorResponse(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
	wantCode string,
	wantMessage string,
) {
	t.Helper()

	var response ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}

	if response.Error.Code != wantCode {
		t.Errorf(
			"error code = %q, want %q",
			response.Error.Code,
			wantCode,
		)
	}
	if response.Error.Message != wantMessage {
		t.Errorf(
			"error message = %q, want %q",
			response.Error.Message,
			wantMessage,
		)
	}
}

func assertErrorCode(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
	wantCode string,
) {
	t.Helper()

	var response ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}

	if response.Error.Code != wantCode {
		t.Errorf(
			"error code = %q, want %q",
			response.Error.Code,
			wantCode,
		)
	}
}
