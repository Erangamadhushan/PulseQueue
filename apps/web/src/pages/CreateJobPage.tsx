import { ArrowLeft } from "lucide-react";
import { Link, useNavigate } from "react-router";
import { PageContainer } from "@/components/layout/PageContainer";
import { Button } from "@/components/ui/button";
import { CreateJobForm } from "@/components/jobs/CreateJobForm";

export function CreateJobPage() {
  const navigate = useNavigate();

  return (
    <PageContainer
      title="Create job"
      description="Add a job to the processing queue."
      actions={
        <Button variant="ghost" render={<Link to="/jobs" />}>
          <ArrowLeft className="size-4" /> Back to Jobs
        </Button>
      }
    >
      <div className="space-y-3">
        <CreateJobForm />
        <div className="flex max-w-2xl justify-end">
          <Button
            type="button"
            variant="outline"
            onClick={() => navigate("/jobs")}
          >
            Cancel
          </Button>
        </div>
      </div>
    </PageContainer>
  );
}