import { getAppServerSession } from "@/lib/server-auth";
import { redirect } from "next/navigation";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import { RetailDownlineManager } from "@/components/user/RetailDownlineManager";
import type { UserSession } from "@/components/user/types";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

export default async function MarketingAgentsPage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  if (!session?.backendToken) redirect("/login");

  return (
    <main className="min-h-screen bg-[#f0f8f8] px-3 pb-24 pt-3">
      <div className="mx-auto w-full max-w-md">
        <RetailDownlineManager authToken={session.backendToken} role="marketing" allowCreate={false} />
      </div>
      <UserBottomNav />
    </main>
  );
}
