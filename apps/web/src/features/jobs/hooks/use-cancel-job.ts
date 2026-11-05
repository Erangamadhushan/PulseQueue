import { useMutation, useQueryClient } from "@tanstack/react-query";

import { cancelJob } from "../api/jobs-api";

export function useCancelJob() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: cancelJob,

    onSuccess: async (job) => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ["jobs"],
        }),
        queryClient.invalidateQueries({
          queryKey: ["dashboard"],
        }),
      ]);

      queryClient.setQueryData(["jobs", job.id], job);
    },
  });
}
