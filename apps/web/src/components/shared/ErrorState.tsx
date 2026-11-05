import { AlertCircle } from "lucide-react";

export function ErrorState({ message = "Something went wrong.", onRetry }: { message?: string; onRetry?: () => void }) {
  return <div className="rounded-xl border border-destructive/30 bg-destructive/5 p-6 text-center"><AlertCircle className="mx-auto size-5 text-destructive" /><p className="mt-2 text-sm">{message}</p>{onRetry ? <button className="mt-3 text-sm font-medium underline" onClick={onRetry}>Try again</button> : null}</div>;
}