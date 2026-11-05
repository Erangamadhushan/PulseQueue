import { useState } from "react";
import { Link } from "react-router";
import { Plus } from "lucide-react";
import { JobFilters } from "@/components/jobs/JobFilters";
import { JobPagination } from "@/components/jobs/JobPagination";
import { JobTable } from "@/components/jobs/JobTable";
import { ErrorState } from "@/components/shared/ErrorState";
import { EmptyState } from "@/components/shared/EmptyState";
import { Button } from "@/components/ui/button";
import { PageContainer } from "@/components/layout/PageContainer";
import { useJobs } from "@/features/jobs/hooks/use-jobs";
import type { JobStatus } from "@/types/jobs";

const DEFAULT_PAGE_SIZE = 20;

function JobsTableSkeleton() {
  return (
    <div className="space-y-3 rounded-xl border bg-card p-5" aria-label="Loading jobs">
      {Array.from({ length: 5 }, (_, index) => (
        <div key={index} className="h-12 animate-pulse rounded-md bg-muted" />
      ))}
    </div>
  );
}

export function JobsPage() {
  const [status, setStatus] = useState<JobStatus>();
  const [queue, setQueue] = useState<string>();
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(DEFAULT_PAGE_SIZE);

  const jobsQuery = useJobs({
    limit,
    offset: (page - 1) * limit,
    status,
    queue,
    search: search.trim() || undefined,
  });
  const jobs = jobsQuery.data ?? [];
  const filteredJobs = jobs;
  const hasNext = jobs.length === limit;

  const updateStatus = (nextStatus?: JobStatus) => {
    setStatus(nextStatus);
    setPage(1);
  };
  const updateQueue = (nextQueue?: string) => {
    setQueue(nextQueue);
    setPage(1);
  };

  return (
    <PageContainer
      title="Jobs"
      description="Search, inspect, and manage background jobs."
      actions={
        <Button render={<Link to="/jobs/new" />}>
          <Plus className="size-4" /> Create job
        </Button>
      }
    >
      <div className="space-y-4">
        <JobFilters
          status={status}
          queue={queue}
          search={search}
          onStatusChange={updateStatus}
          onQueueChange={updateQueue}
          onSearchChange={(nextSearch) => setSearch(nextSearch ?? "")}
        />

        {jobsQuery.isPending ? (
          <JobsTableSkeleton />
        ) : jobsQuery.isError ? (
          <ErrorState
            message="Unable to load jobs."
            onRetry={() => void jobsQuery.refetch()}
          />
        ) : filteredJobs.length === 0 ? (
          <EmptyState
            title="No jobs found"
            description="Try changing your filters or create a new job."
          />
        ) : (
          <JobTable jobs={filteredJobs} />
        )}

        {!jobsQuery.isPending && !jobsQuery.isError && (jobs.length > 0 || page > 1) && (
          <JobPagination
            page={page}
            limit={limit}
            hasNext={hasNext}
            onPageChange={(nextPage) => setPage(nextPage)}
            onLimitChange={(nextLimit) => {
              setLimit(nextLimit);
              setPage(1);
            }}
          />
        )}
      </div>
    </PageContainer>
  );
}
