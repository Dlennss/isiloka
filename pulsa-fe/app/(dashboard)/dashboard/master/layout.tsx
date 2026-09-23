import type { ReactNode } from "react";

export const dynamic = "force-dynamic";
export const revalidate = 0;

export default function MasterDashboardLayout({ children }: { children: ReactNode }) {
  return children;
}
