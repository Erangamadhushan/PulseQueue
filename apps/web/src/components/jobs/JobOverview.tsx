import type { Job } from "@/types/jobs";

function formatDate(value?: string | null) {
  if (!value) return "Not scheduled";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "Not scheduled" : date.toLocaleString();
}

export function JobOverview({ job }: { job: Job }) {
  return (
    <section className="rounded-xl border bg-card p-5 shadow-sm">
      <h2 className="font-semibold">Overview</h2>
      <dl className="mt-4 grid gap-4 text-sm sm:grid-cols-3">
        <div><dt className="text-muted-foreground">Queue</dt><dd className="mt-1 font-medium">{job.queue}</dd></div>
        <div><dt className="text-muted-foreground">Attempts</dt><dd className="mt-1 font-medium">{job.attempts} / {job.max_attempts}</dd></div>
        <div><dt className="text-muted-foreground">Scheduled</dt><dd className="mt-1 font-medium">{formatDate(job.run_at)}</dd></div>
      </dl>
    </section>
  );
}
