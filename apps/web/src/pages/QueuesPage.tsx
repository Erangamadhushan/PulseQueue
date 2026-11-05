import { PageContainer } from "@/components/layout/PageContainer";

export function QueuesPage() {
  return (
    <PageContainer title="Queues" description="Inspect throughput and backlog for every queue.">
      <div className="rounded-xl border bg-card p-6 text-sm text-muted-foreground">
        Live queue metrics are not available yet. This view will populate when queue
        and worker telemetry is exposed by the API.
      </div>
    </PageContainer>
  );
}