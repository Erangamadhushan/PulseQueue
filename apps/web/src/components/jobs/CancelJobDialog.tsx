import { useState } from "react";
import { toast } from "sonner";
import { ApiError } from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { useCancelJob } from "@/features/jobs/hooks/use-cancel-job";

interface CancelJobDialogProps {
  jobId: string;
  disabled?: boolean;
  onSettled?: () => void;
}

export function CancelJobDialog({
  jobId,
  disabled,
  onSettled,
}: CancelJobDialogProps) {
  const mutation = useCancelJob();
  const [open, setOpen] = useState(false);

  const cancel = async () => {
    try {
      await mutation.mutateAsync(jobId);
      toast.success("Job cancelled");
      setOpen(false);
    } catch (error) {
      if (error instanceof ApiError && error.status === 409) {
        toast.error(
          "This job is no longer pending and could not be cancelled.",
        );
      } else {
        toast.error(
          error instanceof Error ? error.message : "Unable to cancel job",
        );
      }
    } finally {
      onSettled?.();
    }
  };

  return (
    <>
      <Button
        variant="outline"
        disabled={disabled || mutation.isPending}
        onClick={() => setOpen(true)}
      >
        Cancel job
      </Button>
      {open && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
          role="presentation"
          onMouseDown={(event) => {
            if (event.target === event.currentTarget && !mutation.isPending) {
              setOpen(false);
            }
          }}
        >
          <div
            className="w-full max-w-md rounded-xl border bg-card p-6 shadow-lg"
            role="dialog"
            aria-modal="true"
            aria-labelledby="cancel-job-title"
            aria-describedby="cancel-job-description"
          >
            <h2 id="cancel-job-title" className="text-lg font-semibold">
              Cancel job?
            </h2>
            <p
              id="cancel-job-description"
              className="mt-2 text-sm text-muted-foreground"
            >
              This will cancel the pending job. Cancellation cannot be undone in
              V1.
            </p>
            <div className="mt-6 flex justify-end gap-2">
              <Button
                variant="ghost"
                onClick={() => setOpen(false)}
                disabled={mutation.isPending}
              >
                Keep job
              </Button>
              <Button
                variant="destructive"
                onClick={() => void cancel()}
                disabled={mutation.isPending}
              >
                {mutation.isPending ? "Cancelling..." : "Confirm cancellation"}
              </Button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
