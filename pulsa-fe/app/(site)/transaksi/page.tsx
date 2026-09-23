import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/nextauth";
import type { UserSession } from "@/components/user/types";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { GuestTransactionHistory } from "@/components/guest/GuestTransactionHistory";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

export default async function GuestTransactionsPage() {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const isLoggedIn = Boolean(session?.backendToken);

  return (
    <main className="min-h-screen bg-sky-50 pb-24">
      <div className="px-4 pt-5">
        <GuestTransactionHistory />
      </div>
      <GuestBottomNav isLoggedIn={isLoggedIn} />
    </main>
  );
}
