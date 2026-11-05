import { Bell, Menu, Plus } from "lucide-react";
import { Link } from "react-router";
import { Button } from "@/components/ui/button";

export function Header({ onMenuClick }: { onMenuClick: () => void }) {
  return (
    <header className="flex h-16 items-center justify-between border-b bg-background px-4 sm:px-6">
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="icon" className="lg:hidden" onClick={onMenuClick} aria-label="Open navigation"><Menu className="size-5" /></Button>
        <p className="hidden text-sm text-muted-foreground lg:block">Background job operations</p>
        <p className="font-semibold lg:hidden">PulseQueue</p>
      </div>
      <div className="flex items-center gap-2">
        <Button variant="ghost" size="icon" aria-label="Notifications"><Bell className="size-4" /></Button>
        <Button render={<Link to="/jobs/new" />} size="sm"><Plus className="size-4" /><span className="hidden sm:inline">Create job</span></Button>
      </div>
    </header>
  );
}
