import { CheckCircle2, XCircle, LoaderCircle } from "lucide-react";

interface SystemHealthProps {
  title: string;
  isLoading: boolean;
  isHealthy: boolean;
}

export function SystemHealth({
  title,
  isLoading,
  isHealthy,
}: SystemHealthProps) {
  const Icon = isLoading ? LoaderCircle : isHealthy ? CheckCircle2 : XCircle;

  return (
    <div className="flex items-center justify-between rounded-lg border border-white/10 bg-[#101116] p-4">
      <span className="text-sm text-zinc-300">{title}</span>

      <div className="flex items-center gap-2">
        <Icon
          className={`size-4 ${
            isLoading
              ? "animate-spin text-zinc-400"
              : isHealthy
                ? "text-emerald-400"
                : "text-red-400"
          }`}
        />

        <span className="text-sm text-zinc-400">
          {isLoading ? "Checking" : isHealthy ? "Operational" : "Unavailable"}
        </span>
      </div>
    </div>
  );
}
