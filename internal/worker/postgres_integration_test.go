package worker

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Erangamadhushan/pulsequeue/internal/database"
	"github.com/Erangamadhushan/pulsequeue/internal/jobs"
	"github.com/google/uuid"
	"log/slog"
)

func TestPostgresWorkersProcessJobsConcurrently(t *testing.T) {
	databaseURL := os.Getenv("PULSEQUEUE_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("PULSEQUEUE_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := database.NewPostgresPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create database pool: %v", err)
	}
	defer pool.Close()

	repository := jobs.NewPostgresRepository(pool)
	queue := "worker-integration-" + uuid.NewString()
	const jobCount = 6
	var handled atomic.Int32

	registry := NewRegistry()
	if err := registry.Register("integration", func(context.Context, jobs.Job) error {
		handled.Add(1)
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	created := make([]jobs.Job, 0, jobCount)
	for i := 0; i < jobCount; i++ {
		job, err := repository.Create(ctx, jobs.Job{
			Queue:       queue,
			Type:        "integration",
			Payload:     json.RawMessage(`{}`),
			Status:      jobs.JobStatusPending,
			MaxAttempts: 1,
			RunAt:       time.Now().UTC(),
		})
		if err != nil {
			t.Fatalf("create job %d: %v", i, err)
		}
		created = append(created, job)
	}

	workerContext, cancel := context.WithCancel(ctx)
	defer cancel()
	workerOne, err := NewWithQueues(
		repository,
		registry,
		"postgres-worker-1-"+uuid.NewString(),
		5*time.Millisecond,
		time.Second,
		[]string{queue},
		slog.Default(),
	)
	if err != nil {
		t.Fatal(err)
	}
	workerTwo, err := NewWithQueues(
		repository,
		registry,
		"postgres-worker-2-"+uuid.NewString(),
		5*time.Millisecond,
		time.Second,
		[]string{queue},
		slog.Default(),
	)
	if err != nil {
		t.Fatal(err)
	}

	var workers sync.WaitGroup
	workers.Add(2)
	go func() {
		defer workers.Done()
		if err := workerOne.Run(workerContext); err != nil {
			t.Errorf("worker one run: %v", err)
		}
	}()
	go func() {
		defer workers.Done()
		if err := workerTwo.Run(workerContext); err != nil {
			t.Errorf("worker two run: %v", err)
		}
	}()

	deadline := time.Now().Add(10 * time.Second)
	for {
		allCompleted := true
		for _, createdJob := range created {
			job, err := repository.GetByID(ctx, createdJob.ID)
			if err != nil {
				t.Fatalf("get job %s: %v", createdJob.ID, err)
			}
			if job.Status != jobs.JobStatusCompleted {
				allCompleted = false
				break
			}
		}
		if allCompleted {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("workers did not complete all jobs; handled=%d", handled.Load())
		}
		time.Sleep(25 * time.Millisecond)
	}

	cancel()
	workers.Wait()
	if handled.Load() != jobCount {
		t.Fatalf("handlers called = %d, want %d", handled.Load(), jobCount)
	}
}
