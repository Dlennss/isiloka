import { getAppServerSession } from "@/lib/server-auth";
import { redirect } from "next/navigation";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import { UserTransferBankPageContent } from "@/components/user/UserTransferBankPageContent";
import type { UserSession } from "@/components/user/types";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

export default async function UserTransferBankPage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  if (!session?.backendToken) redirect("/login");

  return (
    <main className="min-h-screen bg-[#f0f8f8] pb-24">
      <UserTransferBankPageContent />
      <UserBottomNav />
    </main>
  );
}
