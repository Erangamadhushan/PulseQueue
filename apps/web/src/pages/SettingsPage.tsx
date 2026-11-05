import { PageContainer } from "@/components/layout/PageContainer";
import { Button } from "@/components/ui/button";

export function SettingsPage() {
  return (
    <PageContainer title="Settings" description="Manage the PulseQueue dashboard and API connection.">
      <div className="max-w-2xl space-y-6">
        <section className="rounded-xl border bg-card p-5 shadow-sm">
          <h2 className="font-semibold">API connection</h2>
          <p className="mt-1 text-sm text-muted-foreground">The dashboard is connected to the local API.</p>
          <div className="mt-5 flex items-center justify-between rounded-lg border p-4">
            <div><p className="text-sm font-medium">API endpoint</p><p className="text-sm text-muted-foreground">/api</p></div>
            <span className="text-sm text-emerald-600">Connected</span>
          </div>
        </section>
        <section className="rounded-xl border bg-card p-5 shadow-sm">
          <h2 className="font-semibold">Notifications</h2>
          <p className="mt-1 text-sm text-muted-foreground">Choose which operational events to surface.</p>
          <label className="mt-5 flex items-center justify-between text-sm">
            Failed job alerts
            <input type="checkbox" defaultChecked className="size-4 accent-primary" />
          </label>
          <Button className="mt-5">Save preferences</Button>
        </section>
      </div>
    </PageContainer>
  );
}