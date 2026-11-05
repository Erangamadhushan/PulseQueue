import type { Job } from "./jobs";

export interface DashboardStats {
  total: number;
  pending: number;
  running: number;
  completed: number;
  failed: number;
  cancelled: number;
}

export interface SystemHealth {
  status: "healthy" | "unhealthy" | "unknown";
  message?: string;
}

export interface DashboardData {
  stats: DashboardStats;
  recentJobs: Job[];
}