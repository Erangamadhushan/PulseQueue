export function getJobsOffset(page: number, limit: number) {
  return (page - 1) * limit;
}