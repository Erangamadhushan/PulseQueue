import { ListChecks } from "lucide-react";
import { PageContainer } from "@/components/layout/PageContainer";

export function QueuesPage() {
  return (
    <PageContainer title="Queues" description="Inspect throughput and backlog for every queue.">
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        {[
          ["default", "12", "1,020", "2"],
          ["emails", "8", "242", "1"],
          ["webhooks", "4", "22", "0"],
        ].map(([name, pending, processed, workers]) => (
          <div key={name} className="rounded-xl border bg-card p-5 shadow-sm">
            <div className="flex items-center gap-3">
              <div className="flex size-9 items-center justify-center rounded-lg bg-primary/10 text-primary">
                <ListChecks className="size-4" />
              </div>
              <h2 className="font-semibold">{name}</h2>
            </div>
            <div className="mt-6 grid grid-cols-3 gap-3 text-center">
              <div><p className="text-xl font-semibold">{pending}</p><p className="text-xs text-muted-foreground">Pending</p></div>
              <div><p className="text-xl font-semibold">{processed}</p><p className="text-xs text-muted-foreground">Processed</p></div>
              <div><p className="text-xl font-semibold">{workers}</p><p className="text-xs text-muted-foreground">Workers</p></div>
            </div>
          </div>
        ))}
      </div>
    </PageContainer>
  );
}