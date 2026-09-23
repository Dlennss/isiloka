import { getAppServerSession } from "@/lib/server-auth";
import { redirect } from "next/navigation";
import { getMyAgentCreditApplications, getUserProfile } from "@/lib/api.auth";
import { UserAgentCreditPageContent } from "@/components/user/UserAgentCreditPageContent";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import type { UserSession } from "@/components/user/types";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

function isAgentRole(role?: string | null) {
  const normalized = String(role || "").trim().toLowerCase();
  return normalized === "agent" || normalized === "retail_agent" || normalized === "agent_retail";
}

export default async function UserSaldoKreditAgentPage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  if (!session?.backendToken) redirect("/login");

  const profile = await getUserProfile(session.backendToken);
  const role = String(profile?.role || session.user?.role || "").trim().toLowerCase();
  if (!isAgentRole(role)) redirect("/user/saldo");

  const profileWithPhone = profile as typeof profile & { phone?: string; no_hp?: string; nomor_hp?: string; telepon?: string };
  const name = profile?.nama || session.user?.name || "User";
  const email = profile?.email || session.user?.email || "-";
  const phone = profileWithPhone?.phone || profileWithPhone?.no_hp || profileWithPhone?.nomor_hp || profileWithPhone?.telepon || "-";
  const applications = await getMyAgentCreditApplications(session.backendToken);

  return (
    <main className="min-h-screen bg-[#f0f8f8]">
      <UserAgentCreditPageContent
        name={name}
        email={email}
        phone={phone}
        storeName={profile?.store_name || ""}
        mainBalance={Number(profile?.saldo || 0)}
        initialApplications={applications}
      />
      <UserBottomNav />
    </main>
  );
}
