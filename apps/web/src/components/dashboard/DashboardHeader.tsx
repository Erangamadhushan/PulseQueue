import { Link } from "react-router";
import { Button } from "@/components/ui/button";

export function DashboardHeader() {
  return <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"><div><h1 className="text-2xl font-semibold">Dashboard</h1><p className="text-sm text-muted-foreground">Monitor your background job system.</p></div><Button render={<Link to="/jobs/new" />}>Create job</Button></div>;
}