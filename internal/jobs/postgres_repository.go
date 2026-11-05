package jobs

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

type rowScanner interface {
	Scan(dest ...any) error
}

var _ Repository = (*PostgresRepository)(nil)

const jobColumns = `
	id,
	queue,
	type,
	payload,
	status,
	attempts,
	max_attempts,
	run_at,
	last_error,
	created_at,
	updated_at,
	started_at,
	completed_at
`

func NewPostgresRepository(
	pool *pgxpool.Pool,
) *PostgresRepository {
	return &PostgresRepository{
		pool: pool,
	}
}

func scanJob(
	row rowScanner,
) (Job, error) {
	var job Job

	err := row.Scan(
		&job.ID,
		&job.Queue,
		&job.Type,
		&job.Payload,
		&job.Status,
		&job.Attempts,
		&job.MaxAttempts,
		&job.RunAt,
		&job.LastError,
		&job.CreatedAt,
		&job.UpdatedAt,
		&job.StartedAt,
		&job.CompletedAt,
	)

	if err != nil {
		return Job{}, err
	}

	return job, nil
}

func (r *PostgresRepository) Create(
	ctx context.Context,
	job Job,
) (Job, error) {
	const query = `
		INSERT INTO jobs (
			queue,
			type,
			payload,
			status,
			attempts,
			max_attempts,
			run_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7
		)
		RETURNING
			id,
			queue,
			type,
			payload,
			status,
			attempts,
			max_attempts,
			run_at,
			last_error,
			created_at,
			updated_at,
			started_at,
			completed_at
	`

	createdJob, err := scanJob(
		r.pool.QueryRow(
			ctx,
			query,
			job.Queue,
			job.Type,
			job.Payload,
			job.Status,
			job.Attempts,
			job.MaxAttempts,
			job.RunAt,
		),
	)

	if err != nil {
		return Job{}, fmt.Errorf(
			"create job: %w",
			err,
		)
	}

	return createdJob, nil
}

func (r *PostgresRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (Job, error) {
	query := `
		SELECT ` + jobColumns + `
		FROM jobs
		WHERE id = $1
	`

	job, err := scanJob(
		r.pool.QueryRow(
			ctx,
			query,
			id,
		),
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Job{}, ErrJobNotFound
		}

		return Job{}, fmt.Errorf(
			"get job by id: %w",
			err,
		)
	}

	return job, nil
}

func (r *PostgresRepository) List(
	ctx context.Context,
	filter ListJobsFilter,
) ([]Job, error) {
	var query strings.Builder
	query.WriteString("SELECT ")
	query.WriteString(jobColumns)
	query.WriteString(" FROM jobs")

	conditions := make([]string, 0, 2)
	args := make([]any, 0, 4)
	if filter.Status != nil {
		args = append(args, *filter.Status)
		conditions = append(conditions, fmt.Sprintf("status = $%d", len(args)))
	}
	if filter.Queue != nil {
		args = append(args, *filter.Queue)
		conditions = append(conditions, fmt.Sprintf("queue = $%d", len(args)))
	}
	if len(conditions) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(conditions, " AND "))
	}
	query.WriteString(" ORDER BY created_at DESC, id DESC")
	args = append(args, filter.Limit, filter.Offset)
	query.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args)))

	rows, err := r.pool.Query(
		ctx,
		query.String(),
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list jobs: %w",
			err,
		)
	}
	defer rows.Close()

	jobs := make([]Job, 0)

	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, fmt.Errorf(
				"scan job: %w",
				err,
			)
		}

		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate jobs: %w",
			err,
		)
	}

	return jobs, nil
}

func (s *Service) GetJob(
	ctx context.Context,
	id uuid.UUID,
) (Job, error) {

	job, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return Job{}, fmt.Errorf(
			"get job: %w",
			err,
		)
	}

	return job, nil
}

func (s *Service) ListJobs(
	ctx context.Context,
	filter ListJobsFilter,
) ([]Job, error) {
	if filter.Limit == 0 {
		filter.Limit = DefaultListLimit
	}
	if filter.Limit < 1 || filter.Limit > MaxListLimit {
		return nil, ErrInvalidListLimit
	}
	if filter.Offset < 0 {
		return nil, ErrInvalidListOffset
	}
	if filter.Status != nil && !filter.Status.IsValid() {
		return nil, ErrInvalidJobStatus
	}
	if filter.Queue != nil {
		queue := strings.TrimSpace(*filter.Queue)
		if queue == "" {
			return nil, ErrInvalidJobQueue
		}
		filter.Queue = &queue
	}

	jobs, err := s.repository.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf(
			"list jobs: %w",
			err,
		)
	}

	return jobs, nil
}
