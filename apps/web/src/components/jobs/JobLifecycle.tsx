import type { Job } from "@/types/jobs";

function formatDate(value?: string) {
  if (!value) return "Not recorded";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "Not recorded" : date.toLocaleString();
}

export function JobLifecycle({ job }: { job: Job }) {
  const events = [
    { label: "Created", timestamp: job.created_at },
    ...(job.started_at ? [{ label: "Started", timestamp: job.started_at }] : []),
    ...(job.completed_at ? [{ label: "Completed", timestamp: job.completed_at }] : []),
  ];

  return (
    <section className="rounded-xl border bg-card p-5 shadow-sm">
      <h2 className="font-semibold">Execution lifecycle</h2>
      <ol className="mt-4 space-y-3 border-l pl-4 text-sm">
        {events.map((event) => (
          <li key={event.label}>
            <span className="font-medium">{event.label}</span>
            <p className="text-muted-foreground">{formatDate(event.timestamp)}</p>
          </li>
        ))}
      </ol>
      <p className="mt-4 text-xs text-muted-foreground">
        Detailed attempt history will be available when execution history is persisted.
      </p>
    </section>
  );
}
