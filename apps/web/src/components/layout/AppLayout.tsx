import { useState } from "react";
import { Outlet } from "react-router";
import { Header } from "./Header";
import { MobileSidebar } from "./MobileSidebar";
import { Sidebar } from "./Sidebar";

export function AppLayout() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  return (
    <div className="flex min-h-screen bg-muted/30">
      <Sidebar />
      <div className="flex min-w-0 flex-1 flex-col">
        <Header onMenuClick={() => setMobileMenuOpen(true)} />
        <div className="flex-1"><Outlet /></div>
      </div>
      <MobileSidebar open={mobileMenuOpen} onClose={() => setMobileMenuOpen(false)} />
    </div>
  );
}
