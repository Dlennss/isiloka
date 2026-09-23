import { type ReactNode } from "react";

export default function DashboardAdminLayout({ children }: { children: ReactNode }) {
  return <div className="admin-dashboard-surface light">{children}</div>;
}
