import { apiRequest } from "@/lib/api-client";
import type { CreateJobRequest, Job, ListJobsParams } from "@/types/jobs";

function toQuery(params: ListJobsParams = {}) {
  const query = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== "") query.set(key, String(value));
  }
  const result = query.toString();
  return result ? `?${result}` : "";
}

export function listJobs(params?: ListJobsParams) {
  return apiRequest<Job[]>(`/jobs${toQuery(params)}`);
}

export function getJob(id: string) {
  return apiRequest<Job>(`/jobs/${encodeURIComponent(id)}`);
}

export async function createJob(data: CreateJobRequest): Promise<Job> {
  return apiRequest<Job>("/jobs", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export function cancelJob(id: string) {
  return apiRequest<Job>(`/jobs/${encodeURIComponent(id)}/cancel`, {
    method: "POST",
  });
}
