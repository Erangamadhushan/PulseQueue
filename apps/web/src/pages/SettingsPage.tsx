import { useState } from "react";
import { PageContainer } from "@/components/layout/PageContainer";
import { Button } from "@/components/ui/button";
import { useAPIHealth } from "@/features/dashboard/hooks/use-dashboard";

export function SettingsPage() {
  const apiHealth = useAPIHealth();
  const [failedJobAlerts, setFailedJobAlerts] = useState(
    () => localStorage.getItem("pulsequeue.failed-job-alerts") !== "false",
  );
  const [saved, setSaved] = useState(false);

  return (
    <PageContainer title="Settings" description="Manage the PulseQueue dashboard and API connection.">
      <div className="max-w-2xl space-y-6">
        <section className="rounded-xl border bg-card p-5 shadow-sm">
          <h2 className="font-semibold">API connection</h2>
          <p className="mt-1 text-sm text-muted-foreground">Live API connection status.</p>
          <div className="mt-5 flex items-center justify-between rounded-lg border p-4">
            <div><p className="text-sm font-medium">API endpoint</p><p className="text-sm text-muted-foreground">/api</p></div>
            <span className={apiHealth.isSuccess ? "text-sm text-emerald-600" : "text-sm text-destructive"}>
              {apiHealth.isLoading ? "Checking..." : apiHealth.isSuccess ? "Connected" : "Unavailable"}
            </span>
          </div>
        </section>
        <section className="rounded-xl border bg-card p-5 shadow-sm">
          <h2 className="font-semibold">Notifications</h2>
          <p className="mt-1 text-sm text-muted-foreground">Choose which operational events to surface.</p>
          <label className="mt-5 flex items-center justify-between text-sm">
            Failed job alerts
            <input
              type="checkbox"
              checked={failedJobAlerts}
              onChange={(event) => {
                setFailedJobAlerts(event.target.checked);
                setSaved(false);
              }}
              className="size-4 accent-primary"
            />
          </label>
          <Button
            className="mt-5"
            onClick={() => {
              localStorage.setItem("pulsequeue.failed-job-alerts", String(failedJobAlerts));
              setSaved(true);
            }}
          >
            Save preferences
          </Button>
          {saved ? <p className="mt-2 text-sm text-emerald-600">Preferences saved.</p> : null}
        </section>
      </div>
    </PageContainer>
  );
}