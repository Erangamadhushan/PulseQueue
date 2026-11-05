
import { X } from "lucide-react";
import { NavLink } from "react-router";
import { Button } from "@/components/ui/button";

interface MobileSidebarProps {
  open: boolean;
  onClose: () => void;
}

const links = [
  ["Dashboard", "/"],
  ["Jobs", "/jobs"],
  ["Queues", "/queues"],
  ["Workers", "/workers"],
  ["Settings", "/settings"],
];

export function MobileSidebar({ open, onClose }: MobileSidebarProps) {
  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 lg:hidden">
      <button
        className="absolute inset-0 bg-black/30"
        onClick={onClose}
        aria-label="Close navigation"
      />
      <aside className="relative h-full w-72 border-r bg-background p-4 shadow-xl">
        <div className="mb-6 flex items-center justify-between">
          <p className="font-semibold">PulseQueue</p>
          <Button variant="ghost" size="icon" onClick={onClose} aria-label="Close navigation">
            <X className="size-4" />
          </Button>
        </div>
        <nav className="space-y-1">
          {links.map(([label, href]) => (
            <NavLink
              key={href}
              to={href}
              end={href === "/"}
              onClick={onClose}
              className={({ isActive }) =>
                `block rounded-md px-3 py-2 text-sm ${
                  isActive ? "bg-accent font-medium" : "text-muted-foreground hover:bg-accent"
                }`
              }
            >
              {label}
            </NavLink>
          ))}
        </nav>
      </aside>
    </div>
  );
}
