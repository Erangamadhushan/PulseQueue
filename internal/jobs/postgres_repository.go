package jobs

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

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
var _ ClaimRepository = (*PostgresRepository)(nil)

const jobColumns = `
	id,
	queue,
	type,
	payload,
	idempotency_key,
	status,
	attempts,
	max_attempts,
	run_at,
	last_error,
	created_at,
	updated_at,
	started_at,
	completed_at,
	worker_id,
	locked_at,
	lease_expires_at,
	claim_token
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
		&job.IdempotencyKey,
		&job.Status,
		&job.Attempts,
		&job.MaxAttempts,
		&job.RunAt,
		&job.LastError,
		&job.CreatedAt,
		&job.UpdatedAt,
		&job.StartedAt,
		&job.CompletedAt,
		&job.WorkerID,
		&job.LockedAt,
		&job.LeaseExpiresAt,
		&job.ClaimToken,
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
			idempotency_key,
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
			$7,
			$8
		)
		ON CONFLICT (idempotency_key) WHERE idempotency_key IS NOT NULL
		DO UPDATE SET idempotency_key = EXCLUDED.idempotency_key
		RETURNING
			id,
			queue,
			type,
			payload,
			idempotency_key,
			status,
			attempts,
			max_attempts,
			run_at,
			last_error,
			created_at,
			updated_at,
			started_at,
			completed_at,
			worker_id,
			locked_at,
			lease_expires_at,
			claim_token
	`

	createdJob, err := scanJob(
		r.pool.QueryRow(
			ctx,
			query,
			job.Queue,
			job.Type,
			job.Payload,
			job.IdempotencyKey,
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

func (r *PostgresRepository) Cancel(
	ctx context.Context,
	id uuid.UUID,
) (Job, error) {
	query := `
		UPDATE jobs
		SET status = $2, updated_at = NOW()
		WHERE id = $1 AND status = $3
		RETURNING ` + jobColumns

	job, err := scanJob(r.pool.QueryRow(
		ctx,
		query,
		id,
		JobStatusCancelled,
		JobStatusPending,
	))
	if err == nil {
		return job, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Job{}, fmt.Errorf("cancel job: %w", err)
	}

	current, err := r.GetByID(ctx, id)
	if err != nil {
		return Job{}, err
	}
	if current.Status != JobStatusPending {
		return Job{}, ErrJobNotCancellable
	}

	return Job{}, fmt.Errorf("cancel job: %w", ErrJobNotCancellable)
}

func (r *PostgresRepository) Stats(
	ctx context.Context,
) (JobStats, error) {
	const query = `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'pending'),
			COUNT(*) FILTER (WHERE status = 'running'),
			COUNT(*) FILTER (WHERE status = 'completed'),
			COUNT(*) FILTER (WHERE status = 'failed'),
			COUNT(*) FILTER (WHERE status = 'cancelled')
		FROM jobs
	`

	var stats JobStats
	if err := r.pool.QueryRow(ctx, query).Scan(
		&stats.Total,
		&stats.Pending,
		&stats.Running,
		&stats.Completed,
		&stats.Failed,
		&stats.Cancelled,
	); err != nil {
		return JobStats{}, fmt.Errorf("get job stats: %w", err)
	}
	return stats, nil
}

func (r *PostgresRepository) ClaimNext(
	ctx context.Context,
	input ClaimJobInput,
) (*ClaimResult, error) {
	queues, err := validateClaimInput(input)
	if err != nil {
		return nil, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin claim transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	claimToken := uuid.New()
	const query = `
			SELECT ` + jobColumns + `
			FROM jobs
			WHERE status = $1
				AND run_at <= NOW()
				AND attempts < max_attempts
				AND queue = ANY($2)
			ORDER BY run_at ASC, created_at ASC, id ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		`
	job, err := scanJob(tx.QueryRow(ctx, query, JobStatusPending, queues))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("select job to claim: %w", err)
	}

	claimed, err := scanJob(tx.QueryRow(ctx, `
			UPDATE jobs
			SET status = $2,
				attempts = attempts + 1,
				worker_id = $3,
				locked_at = NOW(),
				lease_expires_at = NOW() + ($4 * INTERVAL '1 second'),
				claim_token = $5,
				started_at = COALESCE(started_at, NOW()),
				updated_at = NOW()
			WHERE id = $1
			RETURNING `+jobColumns,
		job.ID,
		JobStatusRunning,
		strings.TrimSpace(input.WorkerID),
		input.LeaseDuration.Seconds(),
		claimToken,
	))
	if err != nil {
		return nil, fmt.Errorf("mark job running: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit claim transaction: %w", err)
	}
	return &ClaimResult{Job: claimed, ClaimToken: claimToken}, nil
}

func validateClaimInput(input ClaimJobInput) ([]string, error) {
	if strings.TrimSpace(input.WorkerID) == "" || input.LeaseDuration <= 0 {
		return nil, ErrInvalidClaimInput
	}
	queues := make([]string, 0, len(input.Queues))
	seen := make(map[string]struct{}, len(input.Queues))
	for _, queue := range input.Queues {
		queue = strings.TrimSpace(queue)
		if queue == "" {
			return nil, ErrInvalidClaimInput
		}
		if _, exists := seen[queue]; !exists {
			seen[queue] = struct{}{}
			queues = append(queues, queue)
		}
	}
	if len(queues) == 0 {
		queues = []string{"default"}
	}
	return queues, nil
}

func (r *PostgresRepository) Complete(
	ctx context.Context,
	id uuid.UUID,
	workerID string,
	claimToken uuid.UUID,
) error {
	return r.finish(ctx, id, workerID, claimToken, JobStatusCompleted, nil)
}

func (r *PostgresRepository) Fail(
	ctx context.Context,
	id uuid.UUID,
	workerID string,
	claimToken uuid.UUID,
	failure error,
) error {
	if failure == nil {
		return errors.New("failure is required")
	}
	return r.finish(ctx, id, workerID, claimToken, JobStatusFailed, &failure)
}

func (r *PostgresRepository) RenewLease(
	ctx context.Context,
	id uuid.UUID,
	workerID string,
	claimToken uuid.UUID,
	leaseDuration time.Duration,
) error {
	if leaseDuration <= 0 {
		return ErrInvalidClaimInput
	}

	tag, err := r.pool.Exec(ctx, `
		UPDATE jobs
		SET lease_expires_at = NOW() + ($5 * INTERVAL '1 second'),
			updated_at = NOW()
		WHERE id = $1
		  AND worker_id = $2
		  AND status = $3
		  AND claim_token = $4
	`, id, workerID, JobStatusRunning, claimToken, leaseDuration.Seconds())
	if err != nil {
		return fmt.Errorf("renew job lease: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrJobOwnershipLost
	}
	return nil
}

func (r *PostgresRepository) finish(
	ctx context.Context,
	id uuid.UUID,
	workerID string,
	claimToken uuid.UUID,
	status JobStatus,
	failure *error,
) error {
	var lastError *string
	if failure != nil {
		message := (*failure).Error()
		lastError = &message
	}

	tag, err := r.pool.Exec(ctx, `
			UPDATE jobs
			SET status = $3,
				last_error = $4,
				completed_at = CASE
					WHEN $3 = $7 THEN NOW()
					ELSE completed_at
				END,
				updated_at = NOW(),
				worker_id = NULL,
				locked_at = NULL,
				lease_expires_at = NULL,
				claim_token = NULL
			WHERE id = $1 AND worker_id = $2 AND status = $5
			AND claim_token = $6
		`, id, workerID, status, lastError, JobStatusRunning, claimToken, JobStatusCompleted)
	if err != nil {
		return fmt.Errorf("finish job: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrJobOwnershipLost
	}
	return nil
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
	if filter.Search != nil {
		args = append(args, "%"+*filter.Search+"%")
		conditions = append(conditions, fmt.Sprintf("(id::text ILIKE $%d OR type ILIKE $%d)", len(args), len(args)))
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

func (s *Service) GetStats(ctx context.Context) (JobStats, error) {
	stats, err := s.repository.Stats(ctx)
	if err != nil {
		return JobStats{}, fmt.Errorf("get job stats: %w", err)
	}

	return stats, nil
}
