import type { Job } from "@/types/jobs";

export function JobPayloadViewer({ payload }: { payload: Job["payload"] }) {
  return (
    <section className="rounded-xl border bg-card p-5 shadow-sm">
      <h2 className="font-semibold">Payload</h2>
      <pre className="mt-3 overflow-x-auto rounded-lg bg-muted p-4 text-sm">
        {JSON.stringify(payload, null, 2)}
      </pre>
    </section>
  );
}
