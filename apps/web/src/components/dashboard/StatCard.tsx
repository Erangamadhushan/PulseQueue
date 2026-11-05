import type { LucideIcon } from "lucide-react";

interface StatCardProps {
  title: string;
  value: number | string;
  description: string;
  icon: LucideIcon;
}

export function StatCard({
  title,
  value,
  description,
  icon: Icon,
}: StatCardProps) {
  return (
    <div className="rounded-xl border border-white/10 bg-[#101116] p-5">
      <div className="flex items-center justify-between">
        <p className="text-sm text-zinc-400">{title}</p>

        <Icon className="size-5 text-zinc-500" />
      </div>

      <p className="mt-4 text-3xl font-semibold tracking-tight text-white">
        {value}
      </p>

      <p className="mt-2 text-xs text-zinc-500">{description}</p>
    </div>
  );
}
