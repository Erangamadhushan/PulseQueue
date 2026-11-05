import { z } from "zod";

export const createJobSchema = z.object({
  type: z.string()
    .trim()
    .min(1, "Job type is required"),

  queue: z.string()
    .trim()
    .default("default"),

  payload: z.string()
    .min(1, "Payload is required")
    .refine((value) => {
      try {
        const parsed: unknown = JSON.parse(value);

        return (
          typeof parsed === "object" &&
          parsed !== null &&
          !Array.isArray(parsed)
        );
      } catch {
        return false;
      }
    }, "Payload must be a valid JSON object"),

  max_attempts: z.number()
    .int()
    .min(1, "Minimum attempts is 1")
    .max(100, "Maximum attempts is 100")
    .default(3),

  run_at: z.string().optional(),
});

export type CreateJobFormValues = z.input<typeof createJobSchema>;
export type CreateJobFormOutput = z.output<typeof createJobSchema>;