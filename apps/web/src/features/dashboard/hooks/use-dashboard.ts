import { useQuery } from "@tanstack/react-query";

import { getAPIHealth, getDatabaseReadiness, getJobStats } from "../api/dashboard-api";

export function useAPIHealth() {
  return useQuery({
    queryKey: ["system", "api-health"],
    queryFn: getAPIHealth,
    refetchInterval: 15_000,
    retry: 1,
  });
}

export function useDatabaseReadiness() {
  return useQuery({
    queryKey: ["system", "database-readiness"],
    queryFn: getDatabaseReadiness,
    refetchInterval: 15_000,
    retry: 1,
  });
}

export function useJobStats() {
  return useQuery({
    queryKey: ["dashboard", "job-stats"],
    queryFn: getJobStats,
    refetchInterval: 15_000,
    retry: 1,
  });
}
