import { Button } from "@/components/ui/button";
import { Select } from "@/components/ui/select";
import { getJobsOffset } from "@/utils/jobs";

const DEFAULT_PAGE_SIZE = 20;
const MAX_PAGE_SIZE = 100;

interface JobPaginationProps {
  page: number;
  limit?: number;
  total?: number;
  hasNext?: boolean;
  onPageChange: (page: number, offset: number) => void;
  onLimitChange: (limit: number) => void;
}

export function JobPagination({
  page,
  limit = DEFAULT_PAGE_SIZE,
  total,
  hasNext = false,
  onPageChange,
  onLimitChange,
}: JobPaginationProps) {
  const pageSize = Math.min(Math.max(limit, 1), MAX_PAGE_SIZE);
  const currentPage = Math.max(page, 1);
  const totalPages = total === undefined
    ? currentPage + (hasNext ? 1 : 0)
    : Math.max(Math.ceil(total / pageSize), 1);

  return (
    <div className="flex flex-col gap-3 border-t p-4 text-sm sm:flex-row sm:items-center sm:justify-between">
      <label className="flex items-center gap-2 text-muted-foreground">
        <span>Rows per page</span>
        <Select
          value={String(pageSize)}
          onChange={(event) => onLimitChange(Math.min(Number(event.target.value), MAX_PAGE_SIZE))}
          className="w-24"
          aria-label="Rows per page"
        >
          <option value="20">20</option>
          <option value="50">50</option>
          <option value="100">100</option>
        </Select>
      </label>
      <div className="flex items-center justify-between gap-4 sm:justify-end">
        <span className="text-muted-foreground">
          Page {Math.min(currentPage, totalPages)}{total !== undefined ? ` of ${totalPages}` : ""}
        </span>
        <div className="flex gap-2">
          <Button
            type="button"
            variant="outline"
            disabled={currentPage <= 1}
            onClick={() =>
              onPageChange(currentPage - 1, getJobsOffset(currentPage - 1, pageSize))
            }
          >
            Previous
          </Button>
          <Button
            type="button"
            variant="outline"
            disabled={total !== undefined ? currentPage >= totalPages : !hasNext}
            onClick={() =>
              onPageChange(currentPage + 1, getJobsOffset(currentPage + 1, pageSize))
            }
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  );
}
