import { getAppServerSession } from "@/lib/server-auth";
import { redirect } from "next/navigation";
import type { UserSession } from "@/components/user/types";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import { RetailCommissionClient } from "@/components/user/RetailCommissionClient";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

export default async function UserAccountKomisiPage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  if (!session?.backendToken) redirect("/login");

  return (
    <main className="min-h-screen bg-sky-50 px-4 pb-24 pt-5">
      <div className="mx-auto w-full max-w-md space-y-4">
        <RetailCommissionClient authToken={session.backendToken} />
      </div>
      <UserBottomNav />
    </main>
  );
}
