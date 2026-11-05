ALTER TABLE jobs
ADD COLUMN idempotency_key TEXT;

CREATE UNIQUE INDEX jobs_idempotency_key_unique
ON jobs (idempotency_key)
WHERE idempotency_key IS NOT NULL;
