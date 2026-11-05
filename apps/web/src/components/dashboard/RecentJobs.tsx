import { Link } from "react-router";
import { EmptyState } from "@/components/shared/EmptyState";
import { Button } from "@/components/ui/button";
import type { Job } from "@/types/jobs";

export function RecentJobs({ jobs }: { jobs: Job[] }) {
  if (!jobs.length) return <EmptyState title="No jobs yet" description="Create a job to see activity here." />;
  return <div className="divide-y rounded-xl border bg-card">{jobs.slice(0, 5).map((job) => <div key={job.id} className="flex items-center gap-3 p-4"><div className="min-w-0 flex-1"><p className="truncate text-sm font-medium">{job.type}</p><p className="text-xs text-muted-foreground">{job.queue}</p></div><span className="text-xs capitalize text-muted-foreground">{job.status}</span><Button variant="ghost" size="sm" render={<Link to={`/jobs/${job.id}`} />}>View</Button></div>)}</div>;
}