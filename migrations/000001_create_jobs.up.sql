CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    queue TEXT NOT NULL DEFAULT 'default',
    type TEXT NOT NULL,

    payload JSONB NOT NULL,

    status TEXT NOT NULL DEFAULT 'pending',

    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,

    run_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    last_error TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,

    CONSTRAINT jobs_status_check
        CHECK (
            status IN (
                'pending',
                'running',
                'completed',
                'failed',
                'cancelled'
            )
        ),

    CONSTRAINT jobs_attempts_check
        CHECK (attempts >= 0),

    CONSTRAINT jobs_max_attempts_check
        CHECK (max_attempts > 0),

    CONSTRAINT jobs_attempts_max_check
        CHECK (attempts <= max_attempts)
);

CREATE INDEX idx_jobs_queue_status_run_at
ON jobs (queue, status, run_at);