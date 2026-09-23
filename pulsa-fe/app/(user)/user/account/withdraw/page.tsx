import { getAppServerSession } from "@/lib/server-auth";
import { redirect } from "next/navigation";
import type { UserSession } from "@/components/user/types";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import { RetailWithdrawClient } from "@/components/user/RetailWithdrawClient";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

export default async function UserAccountWithdrawPage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  if (!session?.backendToken) redirect("/login");
  const role = String(session?.user?.role || "").trim().toLowerCase();
  if (role !== "agent" && role !== "master") redirect("/user/account");

  return (
    <main className="min-h-screen bg-sky-50 px-4 pb-24 pt-5">
      <div className="mx-auto w-full max-w-md space-y-4">
        <RetailWithdrawClient authToken={session.backendToken} />
      </div>
      <UserBottomNav />
    </main>
  );
}
