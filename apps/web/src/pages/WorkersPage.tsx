import { PageContainer } from "@/components/layout/PageContainer";

export function WorkersPage() {
  return (
    <PageContainer title="Workers" description="Monitor worker processes consuming your queues.">
      <div className="rounded-xl border bg-card p-6 text-sm text-muted-foreground">
        Live worker telemetry is not available yet. Worker status will appear here
        when the API exposes registered workers.
      </div>
    </PageContainer>
  );
}