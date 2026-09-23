import { getAppServerSession } from "@/lib/server-auth";
import { redirect } from "next/navigation";
import type { UserSession } from "@/components/user/types";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import { RetailTopupClient } from "@/components/user/RetailTopupClient";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

type PageProps = {
  searchParams?: Promise<{ amount?: string }>;
};

export default async function UserAccountTopupPage({ searchParams }: PageProps) {
  const session = (await getAppServerSession()) as SessionShape | null;
  if (!session?.backendToken) redirect("/login");
  const requestedAmount = Number.parseInt(String((await searchParams)?.amount || ""), 10);
  const initialAmount = Number.isFinite(requestedAmount) && requestedAmount > 0 ? requestedAmount : 0;

  return (
    <main className="min-h-screen bg-[#f0f8f8] px-4 pb-24 pt-5">
      <div className="mx-auto w-full max-w-md space-y-4">
        <RetailTopupClient authToken={session.backendToken} initialAmount={initialAmount} />
      </div>
      <UserBottomNav />
    </main>
  );
}
