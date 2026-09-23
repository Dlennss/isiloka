import { getAppServerSession } from "@/lib/server-auth";
import { redirect } from "next/navigation";
import { getUserProfile } from "@/lib/api.auth";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import { UserSaldoPageContent } from "@/components/user/UserSaldoPageContent";
import type { UserSession } from "@/components/user/types";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

export default async function UserSaldoPage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  if (!session?.backendToken) redirect("/login");

  const profile = await getUserProfile(session.backendToken);
  const userCode = profile?.id ? `USR-${String(profile.id).padStart(3, "0")}` : "USR-001";
  const role = String(profile?.role || session.user?.role || "").trim().toLowerCase();

  return (
    <main className="min-h-screen bg-[#f0f8f8] px-3 pb-24 pt-3">
      <div className="mx-auto w-full max-w-md">
        <UserSaldoPageContent
          saldo={Number(profile?.saldo || 0)}
          userCode={userCode}
          showCredit={false}
        />
      </div>
      <UserBottomNav />
    </main>
  );
}
