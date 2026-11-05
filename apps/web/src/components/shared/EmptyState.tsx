import type { ReactNode } from "react";

export function EmptyState({ title, description, action }: { title: string; description: string; action?: ReactNode }) {
  return <div className="rounded-xl border border-dashed bg-card p-10 text-center"><h2 className="font-semibold">{title}</h2><p className="mt-1 text-sm text-muted-foreground">{description}</p>{action ? <div className="mt-4">{action}</div> : null}</div>;
}