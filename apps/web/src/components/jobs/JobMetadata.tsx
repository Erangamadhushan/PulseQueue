import type { Job } from "@/types/jobs";

function formatDate(value?: string | null) {
  if (!value) return "Not recorded";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "Not recorded" : date.toLocaleString();
}

export function JobMetadata({ job }: { job: Job }) {
  return (
    <section className="rounded-xl border bg-card p-5 shadow-sm">
      <h2 className="font-semibold">Metadata</h2>
      <dl className="mt-4 space-y-3 text-sm">
        <div className="flex justify-between gap-4"><dt className="text-muted-foreground">Created</dt><dd>{formatDate(job.created_at)}</dd></div>
        <div className="flex justify-between gap-4"><dt className="text-muted-foreground">Updated</dt><dd>{formatDate(job.updated_at)}</dd></div>
      </dl>
    </section>
  );
}
