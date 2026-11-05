DROP INDEX IF EXISTS jobs_idempotency_key_unique;
ALTER TABLE jobs DROP COLUMN IF EXISTS idempotency_key;
