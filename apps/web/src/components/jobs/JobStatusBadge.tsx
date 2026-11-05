import { Badge } from "@/components/ui/badge";
import type { JobStatus } from "@/types/jobs";

const statusStyles: Record<JobStatus, string> = {
  pending: "border-amber-500/20 bg-amber-500/10 text-amber-400",
  running: "border-blue-500/20 bg-blue-500/10 text-blue-400",
  completed: "border-emerald-500/20 bg-emerald-500/10 text-emerald-400",
  failed: "border-red-500/20 bg-red-500/10 text-red-400",
  cancelled: "border-zinc-500/20 bg-zinc-500/10 text-zinc-400",
};

interface JobStatusBadgeProps {
  status: JobStatus;
}

export function JobStatusBadge({ status }: JobStatusBadgeProps) {
  return (
    <Badge
      variant="outline"
      className={statusStyles[status]}
    >
      {status}
    </Badge>
  );
}