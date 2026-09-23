import { redirect } from "next/navigation";
import { getAppServerSession } from "@/lib/server-auth";
import type { UserSession } from "@/components/user/types";
import { UserAccountSecurityForm } from "@/components/user/UserAccountSecurityForm";
import { UserBottomNav } from "@/components/user/UserBottomNav";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

export default async function UserAccountSecurityPage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  if (!session?.backendToken) redirect("/login");

  return (
    <main className="min-h-screen bg-[#f0f8f8] px-3 pb-24 pt-3">
      <div className="mx-auto w-full max-w-md">
        <UserAccountSecurityForm authToken={session.backendToken} />
      </div>
      <UserBottomNav />
    </main>
  );
}
