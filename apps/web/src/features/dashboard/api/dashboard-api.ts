import { apiRequest } from "@/lib/api-client";
import type { DashboardStats } from "@/types/dashboard";

export interface HealthResponse {
  status?: string;
  message?: string;
}

export async function getAPIHealth(): Promise<HealthResponse> {
  return apiRequest<HealthResponse>("/healthz");
}

export async function getDatabaseReadiness(): Promise<HealthResponse> {
  return apiRequest<HealthResponse>("/readyz");
}

export async function getJobStats(): Promise<DashboardStats> {
  return apiRequest<DashboardStats>("/stats");
}
