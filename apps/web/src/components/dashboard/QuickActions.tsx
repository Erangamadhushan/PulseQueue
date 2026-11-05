import { Link } from "react-router";
import { ArrowRight, Plus } from "lucide-react";

import { Button } from "@/components/ui/button";

export function QuickActions() {
  return (
    <div className="flex gap-3">
      <Button variant="default" render={<Link to="/jobs/new" />}>
          <Plus className="mr-2 size-4" />
          Create Job
      </Button>

      <Button variant="outline" render={<Link to="/jobs" />}>
          View All Jobs
          <ArrowRight className="ml-2 size-4" />
      </Button>
    </div>
  );
}
