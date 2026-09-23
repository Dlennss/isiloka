import { getAppServerSession } from "@/lib/server-auth";
import { redirect } from "next/navigation";
import { getMyAgentCreditApplications, getUserProfile } from "@/lib/api.auth";
import { UserAgentLevelPageContent } from "@/components/user/UserAgentLevelPageContent";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import type { UserSession } from "@/components/user/types";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

export default async function UserAgentLevelPage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  if (!session?.backendToken) redirect("/login");

  const profile = await getUserProfile(session.backendToken);
  const role = String(profile?.role || session.user?.role || "").trim().toLowerCase();
  if (role !== "agent") redirect("/user/saldo");
  const applications = await getMyAgentCreditApplications(session.backendToken);
  const initialLevelCode = String(applications[0]?.credit_level_code || "start").trim().toLowerCase();

  return (
    <main className="min-h-screen bg-[#f0f8f8]">
      <UserAgentLevelPageContent initialLevelCode={initialLevelCode} />
      <UserBottomNav />
    </main>
  );
}
