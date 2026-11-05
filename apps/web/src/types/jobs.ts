export type JobStatus =
  | "pending"
  | "running"
  | "completed"
  | "failed"
  | "cancelled";

export interface Job {
  id: string;
  queue: string;
  type: string;
  payload: Record<string, unknown>;
  status: JobStatus;
  attempts: number;
  max_attempts: number;
  run_at: string;
  last_error?: string;
  created_at: string;
  updated_at: string;
  started_at?: string;
  completed_at?: string;
}

export interface ListJobsParams {
  limit?: number;
  offset?: number;
  status?: JobStatus;
  queue?: string;
}

export interface CreateJobRequest {
  type: string;
  queue?: string;
  payload: Record<string, unknown>;
  max_attempts?: number;
  run_at?: string;
}