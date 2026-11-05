package jobs_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/Erangamadhushan/pulsequeue/internal/database"
	"github.com/Erangamadhushan/pulsequeue/internal/jobs"
	"github.com/google/uuid"
)

func setupRepository(t *testing.T) *jobs.PostgresRepository {
	t.Helper()

	databaseURL := os.Getenv("PULSEQUEUE_TEST_DATABASE_URL")

	if databaseURL == "" {
		t.Skip(
			"PULSEQUEUE_TEST_DATABASE_URL is not set",
		)
	}

	ctx := context.Background()

	pool, err := database.NewPostgresPool(
		ctx,
		databaseURL,
	)
	if err != nil {
		t.Fatalf(
			"create database pool: %v",
			err,
		)
	}

	t.Cleanup(pool.Close)

	return jobs.NewPostgresRepository(pool)
}

func TestPostgresRepositoryCreateAndGetByID(
	t *testing.T,
) {
	repo := setupRepository(t)
	ctx := context.Background()

	payload := json.RawMessage(
		`{"message":"Hello PulseQueue"}`,
	)

	job := jobs.Job{
		Queue:       "default",
		Type:        "test_job",
		Payload:     payload,
		Status:      jobs.JobStatusPending,
		Attempts:    0,
		MaxAttempts: 3,
		RunAt:       time.Now().UTC(),
	}

	created, err := repo.Create(
		ctx,
		job,
	)
	if err != nil {
		t.Fatalf(
			"create job: %v",
			err,
		)
	}

	if created.ID.String() == "" {
		t.Error(
			"expected generated job ID",
		)
	}

	if created.Type != job.Type {
		t.Errorf(
			"type = %q, want %q",
			created.Type,
			job.Type,
		)
	}

	if created.Status != jobs.JobStatusPending {
		t.Errorf(
			"status = %q, want pending",
			created.Status,
		)
	}

	got, err := repo.GetByID(
		ctx,
		created.ID,
	)
	if err != nil {
		t.Fatalf(
			"get job: %v",
			err,
		)
	}

	if got.ID != created.ID {
		t.Errorf(
			"ID = %v, want %v",
			got.ID,
			created.ID,
		)
	}
}

func TestPostgresRepositoryGetByIDNotFound(
	t *testing.T,
) {
	repo := setupRepository(t)
	ctx := context.Background()

	_, err := repo.GetByID(
		ctx,
		uuid.New(),
	)

	if !errors.Is(
		err,
		jobs.ErrJobNotFound,
	) {
		t.Fatalf(
			"expected ErrJobNotFound, got %v",
			err,
		)
	}
}

func TestPostgresRepositoryCancel(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, jobs.Job{
		Queue:       "cancel-" + uuid.NewString(),
		Type:        "cancel_test",
		Payload:     json.RawMessage(`{}`),
		Status:      jobs.JobStatusPending,
		MaxAttempts: 3,
		RunAt:       time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}

	cancelled, err := repo.Cancel(ctx, created.ID)
	if err != nil {
		t.Fatalf("cancel job: %v", err)
	}
	if cancelled.Status != jobs.JobStatusCancelled {
		t.Fatalf("status = %q, want %q", cancelled.Status, jobs.JobStatusCancelled)
	}
}

func TestPostgresRepositoryCancelNotFound(t *testing.T) {
	repo := setupRepository(t)

	_, err := repo.Cancel(context.Background(), uuid.New())
	if !errors.Is(err, jobs.ErrJobNotFound) {
		t.Fatalf("expected ErrJobNotFound, got %v", err)
	}
}

func TestPostgresRepositoryStats(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()
	queue := "stats-" + uuid.NewString()

	for _, status := range []jobs.JobStatus{
		jobs.JobStatusPending,
		jobs.JobStatusRunning,
		jobs.JobStatusCompleted,
		jobs.JobStatusFailed,
		jobs.JobStatusCancelled,
	} {
		_, err := repo.Create(ctx, jobs.Job{
			Queue: queue, Type: "stats_test", Payload: json.RawMessage(`{}`),
			Status: status, MaxAttempts: 3, RunAt: time.Now().UTC(),
		})
		if err != nil {
			t.Fatalf("create %s job: %v", status, err)
		}
	}

	stats, err := repo.Stats(ctx)
	if err != nil {
		t.Fatalf("get stats: %v", err)
	}
	if stats.Total < 5 || stats.Pending < 1 || stats.Running < 1 ||
		stats.Completed < 1 || stats.Failed < 1 || stats.Cancelled < 1 {
		t.Fatalf("stats = %+v, want all seeded statuses represented", stats)
	}
}

func TestPostgresRepositoryCancelRunningJob(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, jobs.Job{
		Queue:       "cancel-" + uuid.NewString(),
		Type:        "cancel_test",
		Payload:     json.RawMessage(`{}`),
		Status:      jobs.JobStatusRunning,
		MaxAttempts: 3,
		RunAt:       time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}

	_, err = repo.Cancel(ctx, created.ID)
	if !errors.Is(err, jobs.ErrJobNotCancellable) {
		t.Fatalf("expected ErrJobNotCancellable, got %v", err)
	}
}

func TestPostgresRepositoryConcurrentCancel(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()
	created, err := repo.Create(ctx, jobs.Job{
		Queue:       "cancel-" + uuid.NewString(),
		Type:        "cancel_test",
		Payload:     json.RawMessage(`{}`),
		Status:      jobs.JobStatusPending,
		MaxAttempts: 3,
		RunAt:       time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}

	results := make(chan error, 2)
	for range 2 {
		go func() {
			_, cancelErr := repo.Cancel(ctx, created.ID)
			results <- cancelErr
		}()
	}

	var success, rejected int
	for range 2 {
		err := <-results
		switch err {
		case nil:
			success++
		case jobs.ErrJobNotCancellable:
			rejected++
		default:
			t.Errorf("unexpected cancellation error: %v", err)
		}
	}
	if success != 1 || rejected != 1 {
		t.Fatalf("successes = %d, rejections = %d, want 1 and 1", success, rejected)
	}

	final, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get final job: %v", err)
	}
	if final.Status != jobs.JobStatusCancelled {
		t.Fatalf("final status = %q, want %q", final.Status, jobs.JobStatusCancelled)
	}
}

func TestPostgresRepositoryListFiltersOrdersAndPaginates(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()
	queue := "integration-" + uuid.NewString()

	create := func(status jobs.JobStatus) jobs.Job {
		job, err := repo.Create(ctx, jobs.Job{
			Queue:       queue,
			Type:        "list_test",
			Payload:     json.RawMessage(`{}`),
			Status:      status,
			MaxAttempts: 3,
			RunAt:       time.Now().UTC(),
		})
		if err != nil {
			t.Fatalf("create job: %v", err)
		}
		return job
	}

	first := create(jobs.JobStatusPending)
	second := create(jobs.JobStatusPending)
	create(jobs.JobStatusCompleted)

	status := jobs.JobStatusPending
	queueFilter := queue
	got, err := repo.List(ctx, jobs.ListJobsFilter{
		Limit:  1,
		Offset: 1,
		Status: &status,
		Queue:  &queueFilter,
	})
	if err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d jobs, want 1", len(got))
	}
	if got[0].ID != first.ID {
		t.Fatalf("job ID = %v, want %v after descending pagination", got[0].ID, first.ID)
	}
	if second.CreatedAt.Before(first.CreatedAt) {
		t.Fatalf("created timestamps are not ordered for test: %v before %v", second.CreatedAt, first.CreatedAt)
	}
}
