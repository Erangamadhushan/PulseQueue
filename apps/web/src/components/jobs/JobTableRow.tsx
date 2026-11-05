import { Link, useNavigate } from "react-router";
import { Eye } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  TableCell,
  TableRow,
} from "@/components/ui/table";
import { JobStatusBadge } from "./JobStatusBadge";
import type { Job } from "@/types/jobs";

function formatDate(value: string) {
  return new Date(value).toLocaleString();
}

export function JobTableRow({ job }: { job: Job }) {
  const navigate = useNavigate();
  const detailsPath = `/jobs/${job.id}`;

  return (
    <TableRow
      className="cursor-pointer"
      onClick={() => navigate(detailsPath)}
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          navigate(detailsPath);
        }
      }}
      role="link"
      tabIndex={0}
    >
      <TableCell>
        <p className="font-medium">{job.type}</p>
        <p className="font-mono text-xs text-muted-foreground">{job.id}</p>
      </TableCell>
      <TableCell className="text-muted-foreground">{job.queue}</TableCell>
      <TableCell><JobStatusBadge status={job.status} /></TableCell>
      <TableCell>{job.attempts} / {job.max_attempts}</TableCell>
      <TableCell className="whitespace-nowrap text-muted-foreground">{formatDate(job.run_at)}</TableCell>
      <TableCell className="whitespace-nowrap text-muted-foreground">{formatDate(job.created_at)}</TableCell>
      <TableCell className="text-right">
        <Button
          variant="ghost"
          size="icon"
          render={<Link to={detailsPath} />}
          aria-label={`View ${job.id}`}
          onClick={(event) => event.stopPropagation()}
        >
          <Eye className="size-4" />
        </Button>
      </TableCell>
    </TableRow>
  );
}
