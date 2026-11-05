import { useParams } from "react-router";
import { CancelJobDialog } from "@/components/jobs/CancelJobDialog";
import { JobDetailsHeader } from "@/components/jobs/JobDetailsHeader";
import { JobErrorDetails } from "@/components/jobs/JobErrorDetails";
import { JobLifecycle } from "@/components/jobs/JobLifecycle";
import { JobMetadata } from "@/components/jobs/JobMetadata";
import { JobOverview } from "@/components/jobs/JobOverview";
import { JobPayloadViewer } from "@/components/jobs/JobPayloadViewer";
import { ErrorState } from "@/components/shared/ErrorState";
import { LoadingState } from "@/components/shared/LoadingState";
import { PageContainer } from "@/components/layout/PageContainer";
import { useJob } from "@/features/jobs/hooks/use-jobs";
import { ApiError } from "@/lib/api-client";

function getJobErrorMessage(error: unknown) {
  if (error instanceof ApiError) {
    if (error.status === 400) return "The job ID is invalid.";
    if (error.status === 404) return "This job could not be found.";
  }

  return "Unable to load job details.";
}

export function JobDetailsPage() {
  const { jobId } = useParams();
  const query = useJob(jobId);

  if (query.isPending) {
    return <PageContainer title="Job details"><LoadingState /></PageContainer>;
  }
  if (query.isError || !query.data) {
    return (
      <PageContainer title="Job details">
        <ErrorState
          message={getJobErrorMessage(query.error)}
          onRetry={() => void query.refetch()}
        />
      </PageContainer>
    );
  }

  const job = query.data;
  return (
    <PageContainer title="Job details" description="Inspect execution details for this job.">
      <div className="space-y-6">
        <JobDetailsHeader job={job} onRefresh={() => void query.refetch()} isRefreshing={query.isFetching} />
        <JobOverview job={job} />
        <div className="grid gap-6 lg:grid-cols-2">
          <JobMetadata job={job} />
          <JobErrorDetails job={job} />
        </div>
        <JobPayloadViewer payload={job.payload} />
        <JobLifecycle job={job} />
        {job.status === "pending" && (
          <CancelJobDialog jobId={job.id} onSettled={() => void query.refetch()} />
        )}
      </div>
    </PageContainer>
  );
}
