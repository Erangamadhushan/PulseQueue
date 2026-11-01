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
