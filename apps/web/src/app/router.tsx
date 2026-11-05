import {
  BrowserRouter,
  Navigate,
  Route,
  Routes,
} from "react-router";

import { AppLayout } from "@/components/layout/AppLayout";

import { DashboardPage } from "@/pages/DashboardPage";
import { JobsPage } from "@/pages/jobs/JobsPage";
import { CreateJobPage } from "@/pages/CreateJobPage";
import { JobDetailsPage } from "@/pages/JobDetailsPage";
import { WorkersPage } from "@/pages/WorkersPage";
import { QueuesPage } from "@/pages/QueuesPage";
import { SettingsPage } from "@/pages/SettingsPage";

export function AppRouter() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<AppLayout />}>
          <Route index element={<DashboardPage />} />
          <Route path="jobs" element={<JobsPage />} />
          <Route path="jobs/new" element={<CreateJobPage />} />
          <Route path="jobs/:jobId" element={<JobDetailsPage />} />
          <Route path="workers" element={<WorkersPage />} />
          <Route path="queues" element={<QueuesPage />} />
          <Route path="settings" element={<SettingsPage />} />
        </Route>

        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}