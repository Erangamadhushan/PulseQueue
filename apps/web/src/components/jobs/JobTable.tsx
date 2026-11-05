import {
  Table,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { EmptyState } from "@/components/shared/EmptyState";
import { JobTableRow } from "./JobTableRow";
import type { Job } from "@/types/jobs";

export function JobTable({ jobs }: { jobs: Job[] }) {
  if (!jobs.length) {
    return <EmptyState title="No jobs found" description="Try changing your filters or create a new job." />;
  }

  return (
    <div className="overflow-hidden rounded-xl border bg-card">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Job</TableHead>
            <TableHead>Queue</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Attempts</TableHead>
            <TableHead>Scheduled</TableHead>
            <TableHead>Created</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {jobs.map((job) => (
            <JobTableRow key={job.id} job={job} />
          ))}
        </TableBody>
      </Table>
    </div>
  );
}