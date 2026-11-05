import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { useNavigate } from "react-router";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useCreateJob } from "@/features/jobs/hooks/use-jobs";
import {
  createJobSchema,
  type CreateJobFormValues,
  type CreateJobFormOutput,
} from "@/features/jobs/schemas/job-schema";

const defaultPayload = '{\n  "message": "Hello PulseQueue"\n}';

export function CreateJobForm() {
  const navigate = useNavigate();
  const mutation = useCreateJob();
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<CreateJobFormValues, undefined, CreateJobFormOutput>({
    resolver: zodResolver(createJobSchema),
    defaultValues: {
      queue: "default",
      payload: defaultPayload,
      max_attempts: 3,
      run_at: "",
    },
  });

  const onSubmit = async (values: CreateJobFormOutput) => {
    let payload: Record<string, unknown>;

    try {
      payload = JSON.parse(values.payload) as Record<string, unknown>;
    } catch {
      setError("payload", { message: "Payload must be valid JSON" });
      return;
    }

    try {
      const job = await mutation.mutateAsync({
        type: values.type,
        queue: values.queue,
        payload,
        max_attempts: values.max_attempts,
        run_at: values.run_at
          ? new Date(values.run_at).toISOString()
          : undefined,
      });

      toast.success("Job created successfully");
      navigate(`/jobs/${job.id}`);
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "Unable to create job",
      );
    }
  };

  const submitting = isSubmitting || mutation.isPending;

  return (
    <form
      className="max-w-2xl space-y-5 rounded-xl border bg-card p-5 shadow-sm"
      onSubmit={(event) => {
        if (submitting) {
          event.preventDefault();
          return;
        }
        void handleSubmit(onSubmit)(event);
      }}
      noValidate
    >
      <label className="block space-y-2 text-sm font-medium">
        <span>Queue</span>
        <Input {...register("queue")} aria-invalid={Boolean(errors.queue)} />
        {errors.queue && (
          <p className="text-sm font-normal text-destructive">{errors.queue.message}</p>
        )}
      </label>

      <label className="block space-y-2 text-sm font-medium">
        <span>Job type</span>
        <Input
          {...register("type")}
          placeholder="send_email"
          aria-invalid={Boolean(errors.type)}
        />
        {errors.type && (
          <p className="text-sm font-normal text-destructive">{errors.type.message}</p>
        )}
      </label>

      <label className="block space-y-2 text-sm font-medium">
        <span>Payload</span>
        <textarea
          {...register("payload")}
          rows={7}
          className="w-full rounded-md border bg-background p-3 font-mono text-sm font-normal outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50"
          aria-invalid={Boolean(errors.payload)}
        />
        {errors.payload && (
          <p className="text-sm font-normal text-destructive">{errors.payload.message}</p>
        )}
      </label>

      <label className="block space-y-2 text-sm font-medium">
        <span>Max attempts</span>
        <Input
          type="number"
          min={1}
          max={100}
          {...register("max_attempts", { valueAsNumber: true })}
          aria-invalid={Boolean(errors.max_attempts)}
        />
        {errors.max_attempts && (
          <p className="text-sm font-normal text-destructive">
            {errors.max_attempts.message}
          </p>
        )}
      </label>

      <label className="block space-y-2 text-sm font-medium">
        <span>Schedule (optional)</span>
        <Input
          type="datetime-local"
          {...register("run_at")}
          aria-invalid={Boolean(errors.run_at)}
        />
        {errors.run_at && (
          <p className="text-sm font-normal text-destructive">{errors.run_at.message}</p>
        )}
      </label>

      <div className="flex justify-end">
        <Button type="submit" disabled={submitting}>
          {submitting ? "Queueing..." : "Queue job"}
        </Button>
      </div>
    </form>
  );
}
