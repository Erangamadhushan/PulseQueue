import type { JobStatus } from "@/types/jobs";

const colors: Record<JobStatus, string> = { pending: "bg-amber-500", running: "bg-blue-500", completed: "bg-emerald-500", failed: "bg-red-500", cancelled: "bg-slate-400" };

export function JobStatusChart({ values }: { values: Partial<Record<JobStatus, number>> }) {
  const total = Object.values(values).reduce((sum, value) => sum + (value ?? 0), 0);
  return <div className="space-y-3">{(Object.keys(colors) as JobStatus[]).map((status) => { const value = values[status] ?? 0; const width = total ? (value / total) * 100 : 0; return <div key={status}><div className="mb-1 flex justify-between text-xs"><span className="capitalize text-muted-foreground">{status}</span><span>{value}</span></div><div className="h-2 rounded-full bg-muted"><div className={`h-2 rounded-full ${colors[status]}`} style={{ width: `${width}%` }} /></div></div>; })}</div>;
}