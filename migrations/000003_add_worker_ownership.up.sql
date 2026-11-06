ALTER TABLE jobs
ADD COLUMN worker_id TEXT,
ADD COLUMN locked_at TIMESTAMPTZ,
ADD COLUMN lease_expires_at TIMESTAMPTZ;

CREATE INDEX idx_jobs_claimable
ON jobs (status, run_at, created_at, id)
WHERE status = 'pending';
