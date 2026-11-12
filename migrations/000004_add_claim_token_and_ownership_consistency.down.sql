ALTER TABLE jobs
DROP CONSTRAINT IF EXISTS jobs_ownership_consistency_check;

ALTER TABLE jobs
DROP COLUMN IF EXISTS claim_token;
