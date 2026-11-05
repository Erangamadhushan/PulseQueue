import { ArrowLeft, RefreshCw } from "lucide-react";
import { Link } from "react-router";
import type { ReactNode } from "react";
import { Button } from "@/components/ui/button";
import { JobStatusBadge } from "./JobStatusBadge";
import type { Job } from "@/types/jobs";

interface JobDetailsHeaderProps {
  job: Job;
  onRefresh?: () => void;
  isRefreshing?: boolean;
  actions?: ReactNode;
}

export function JobDetailsHeader({
  job,
  onRefresh,
  isRefreshing = false,
  actions,
}: JobDetailsHeaderProps) {
  return (
    <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
      <div className="flex items-start gap-3">
        <Button variant="ghost" size="icon" render={<Link to="/jobs" />} aria-label="Back to jobs">
          <ArrowLeft className="size-4" />
        </Button>
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">{job.type}</h1>
          <p className="mt-1 font-mono text-xs text-muted-foreground">{job.id}</p>
        </div>
        <JobStatusBadge status={job.status} />
      </div>
      <div className="flex items-center gap-2">
        {onRefresh && (
          <Button variant="outline" onClick={onRefresh} disabled={isRefreshing}>
            <RefreshCw className="size-4" /> {isRefreshing ? "Refreshing..." : "Refresh"}
          </Button>
        )}
        {actions}
      </div>
    </div>
  );
}
