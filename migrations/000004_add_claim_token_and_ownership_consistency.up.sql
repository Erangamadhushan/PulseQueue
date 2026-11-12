ALTER TABLE jobs
ADD COLUMN claim_token UUID;

ALTER TABLE jobs
ADD CONSTRAINT jobs_ownership_consistency_check
CHECK (
    (
        status = 'running'
        AND worker_id IS NOT NULL
        AND locked_at IS NOT NULL
        AND lease_expires_at IS NOT NULL
        AND claim_token IS NOT NULL
    )
    OR
    (
        status <> 'running'
        AND worker_id IS NULL
        AND locked_at IS NULL
        AND lease_expires_at IS NULL
        AND claim_token IS NULL
    )
);
