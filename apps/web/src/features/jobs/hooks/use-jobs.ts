import { useQuery } from "@tanstack/react-query";
import { getJob, listJobs } from "../api/jobs-api";
import { useCreateJob } from "./use-create-job";
import { useCancelJob } from "./use-cancel-job";
import type { ListJobsParams } from "@/types/jobs";

export function useJobs(params?: ListJobsParams) {
  return useQuery({
    queryKey: ["jobs", params],
    queryFn: () => listJobs(params),
  });
}

export function useJob(id: string | undefined) {
  return useQuery({
    queryKey: ["jobs", id],
    queryFn: () => getJob(id as string),
    enabled: Boolean(id),
  });
}

export { useCreateJob, useCancelJob };
