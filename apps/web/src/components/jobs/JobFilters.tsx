import type { JobStatus } from "@/types/jobs";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";

interface JobFiltersProps {
  status?: JobStatus;
  queue?: string;
  search?: string;
  onStatusChange: (status?: JobStatus) => void;
  onQueueChange: (queue?: string) => void;
  onSearchChange: (search?: string) => void;
}

export function JobFilters({
  status,
  queue,
  search,
  onStatusChange,
  onQueueChange,
  onSearchChange,
}: JobFiltersProps) {
  const hasFilters = Boolean(status || queue || search);

  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-end">
      <label className="flex-1 space-y-1.5 text-sm font-medium">
        <span>Search</span>
        <Input
          value={search ?? ""}
          onChange={(event) => onSearchChange(event.target.value || undefined)}
          placeholder="Job type or UUID"
          aria-label="Search by job type or UUID"
        />
      </label>
      <label className="w-full space-y-1.5 text-sm font-medium sm:max-w-52">
        <span>Queue</span>
        <Input
          value={queue ?? ""}
          onChange={(event) => onQueueChange(event.target.value || undefined)}
          placeholder="Queue name"
        />
      </label>
      <label className="w-full space-y-1.5 text-sm font-medium sm:max-w-44">
        <span>Status</span>
        <Select
          value={status ?? ""}
          onChange={(event) =>
            onStatusChange(
              (event.target.value || undefined) as JobStatus | undefined,
            )
          }
        >
          <option value="">All</option>
          <option value="pending">Pending</option>
          <option value="running">Running</option>
          <option value="completed">Completed</option>
          <option value="failed">Failed</option>
          <option value="cancelled">Cancelled</option>
        </Select>
      </label>
      {hasFilters && (
        <Button
          type="button"
          variant="outline"
          onClick={() => {
            onStatusChange(undefined);
            onQueueChange(undefined);
            onSearchChange(undefined);
          }}
        >
          Clear
        </Button>
      )}
    </div>
  );
}
