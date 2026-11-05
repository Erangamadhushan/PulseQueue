import { CheckCircle2, Clock3, Layers, XCircle } from "lucide-react";

import { StatCard } from "@/components/dashboard/StatCard";
import { SystemHealth } from "@/components/dashboard/SystemHealth";
import { QuickActions } from "@/components/dashboard/QuickActions";

import {
  useAPIHealth,
  useDatabaseReadiness,
  useJobStats,
} from "@/features/dashboard/hooks/use-dashboard";

export function DashboardPage() {
  const apiHealth = useAPIHealth();
  const database = useDatabaseReadiness();
  const stats = useJobStats();

  return (
    <div className="space-y-8 p-3">
      <div>
        <h1 className="text-2xl font-semibold">Overview</h1>

        <p className="mt-1 text-sm text-zinc-400">
          Monitor your background job infrastructure.
        </p>
      </div>

      <QuickActions />

      <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard
          title="Total Jobs"
          value={stats.data?.total ?? "—"}
          description={stats.isLoading ? "Loading statistics" : "All jobs"}
          icon={Layers}
        />

        <StatCard
          title="Pending Jobs"
          value={stats.data?.pending ?? "—"}
          description={stats.isLoading ? "Loading statistics" : "Waiting to run"}
          icon={Clock3}
        />

        <StatCard
          title="Completed Jobs"
          value={stats.data?.completed ?? "—"}
          description={stats.isLoading ? "Loading statistics" : "Successfully completed"}
          icon={CheckCircle2}
        />

        <StatCard
          title="Failed Jobs"
          value={stats.data?.failed ?? "—"}
          description={stats.isLoading ? "Loading statistics" : "Require attention"}
          icon={XCircle}
        />
      </section>

      <section className="space-y-4">
        <div>
          <h2 className="text-lg font-semibold text-white">System Health</h2>

          <p className="text-sm text-zinc-500">
            Current API and database availability.
          </p>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <SystemHealth
            title="Go API"
            isLoading={apiHealth.isLoading}
            isHealthy={apiHealth.isSuccess}
          />

          <SystemHealth
            title="PostgreSQL"
            isLoading={database.isLoading}
            isHealthy={database.isSuccess}
          />
        </div>
      </section>
    </div>
  );
}
