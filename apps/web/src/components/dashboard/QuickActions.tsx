import { Link } from "react-router";
import { ArrowRight, Plus } from "lucide-react";

import { Button } from "@/components/ui/button";

export function QuickActions() {
  return (
    <div className="flex gap-3">
      <Button variant="default">
        <Link to="/jobs/new" className="flex items-center">
          <Plus className="mr-2 size-4" />
          Create Job
        </Link>
      </Button>

      <Button variant="outline" >
        <Link to="/jobs" className="flex items-center">
          View All Jobs
          <ArrowRight className="ml-2 size-4" />
        </Link>
      </Button>
    </div>
  );
}
