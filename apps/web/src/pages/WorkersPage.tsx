import { Cpu, MoreHorizontal } from "lucide-react";
import { PageContainer } from "@/components/layout/PageContainer";
import { Button } from "@/components/ui/button";

export function WorkersPage() {
  return (
    <PageContainer title="Workers" description="Monitor worker processes consuming your queues.">
      <div className="rounded-xl border bg-card shadow-sm">
        <div className="border-b p-5">
          <h2 className="font-semibold">Worker processes</h2>
          <p className="text-sm text-muted-foreground">8 workers are currently online.</p>
        </div>
        <div className="divide-y">
          {["worker-eu-01", "worker-eu-02", "worker-us-01", "worker-us-02"].map((worker) => (
            <div key={worker} className="flex items-center gap-4 p-5">
              <div className="flex size-9 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-600">
                <Cpu className="size-4" />
              </div>
              <div className="min-w-0 flex-1">
                <p className="font-medium">{worker}</p>
                <p className="text-sm text-muted-foreground">default queue · 2 jobs/min</p>
              </div>
              <span className="hidden text-sm text-emerald-600 sm:block">Online</span>
              <Button variant="ghost" size="icon" aria-label={`Actions for ${worker}`}>
                <MoreHorizontal className="size-4" />
              </Button>
            </div>
          ))}
        </div>
      </div>
    </PageContainer>
  );
}