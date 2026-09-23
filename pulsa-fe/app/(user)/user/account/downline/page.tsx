import { getAppServerSession } from "@/lib/server-auth";
import { redirect } from "next/navigation";
import { getUserProfile } from "@/lib/api.auth";
import type { UserSession } from "@/components/user/types";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import { RetailDownlineManager } from "@/components/user/RetailDownlineManager";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

export default async function UserAccountDownlinePage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  if (!session?.backendToken) redirect("/login");
  const profile = await getUserProfile(session.backendToken);
  const role = String(profile?.role || session.user?.role || "").toLowerCase();

  return (
    <main className="min-h-screen bg-[#f0f8f8] px-3 pb-24 pt-3">
      <div className="mx-auto w-full max-w-md space-y-4">
        <RetailDownlineManager authToken={session.backendToken} role={role} />
      </div>
      <UserBottomNav />
    </main>
  );
}
