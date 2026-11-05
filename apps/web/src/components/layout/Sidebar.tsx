import { Activity, Boxes, BriefcaseBusiness, Cog, LayoutDashboard, ListChecks } from "lucide-react";
import { NavLink } from "react-router";

const navigation = [
  ["Dashboard", "/", LayoutDashboard],
  ["Jobs", "/jobs", BriefcaseBusiness],
  ["Queues", "/queues", ListChecks],
  ["Workers", "/workers", Boxes],
  ["Settings", "/settings", Cog],
] as const;

export function Sidebar() {
  return (
    <aside className="hidden w-64 shrink-0 border-r bg-sidebar lg:block">
      <div className="flex min-h-screen flex-col">
        <div className="flex h-16 items-center gap-3 border-b px-6">
          <div className="flex size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground"><Activity className="size-4" /></div>
          <div><p className="font-semibold">PulseQueue</p><p className="text-xs text-muted-foreground">Job control center</p></div>
        </div>
        <nav className="flex-1 space-y-1 p-4" aria-label="Main navigation">
          {navigation.map(([label, href, Icon]) => (
            <NavLink key={href} to={href} end={href === "/"} className={({ isActive }) => `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium ${isActive ? "bg-sidebar-accent text-sidebar-accent-foreground" : "text-sidebar-foreground/70 hover:bg-sidebar-accent/70"}`}>
              <Icon className="size-4" />{label}
            </NavLink>
          ))}
        </nav>
        <div className="border-t p-4 text-xs text-muted-foreground"><p>API status</p><p className="mt-1 flex items-center gap-2 text-foreground"><span className="size-2 rounded-full bg-emerald-500" />Operational</p></div>
      </div>
    </aside>
  );
}
