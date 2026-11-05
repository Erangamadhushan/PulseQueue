import type { Job } from "@/types/jobs";

export function JobErrorDetails({ job }: { job: Job }) {
  return (
    <section className="rounded-xl border bg-card p-5 shadow-sm">
      <h2 className="font-semibold">Attempts and errors</h2>
      <p className="mt-3 text-sm">Attempts: {job.attempts} / {job.max_attempts}</p>
      {job.last_error ? (
        <pre className="mt-3 overflow-x-auto rounded-lg bg-destructive/10 p-4 text-sm text-destructive">
          {job.last_error}
        </pre>
      ) : (
        <p className="mt-3 text-sm text-muted-foreground">No execution error recorded.</p>
      )}
    </section>
  );
}
