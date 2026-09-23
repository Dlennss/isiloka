import { getAppServerSession } from "@/lib/server-auth";
import { redirect } from "next/navigation";
import { getUserOrders } from "@/lib/api.transactions";
import type { UserAppOrder, UserSession } from "@/components/user/types";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import { UserTransactionPageContent } from "@/components/user/UserTransactionPageContent";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

type PageProps = {
  searchParams?: Promise<{ status?: string }>;
};

const PAGE_SIZE = 10;

const VALID_STATUS = new Set([
  "pending_payment",
  "paid",
  "processing_provider",
  "success",
  "failed",
  "expired",
  "cancelled",
  "refunded",
]);

export default async function UserTransaksiPage({ searchParams }: PageProps) {
  const session = (await getAppServerSession()) as SessionShape | null;
  if (!session?.backendToken) {
    redirect("/login");
  }

  const resolvedSearch = (await searchParams) || {};
  const status = VALID_STATUS.has((resolvedSearch.status || "").trim()) ? (resolvedSearch.status || "").trim() : "";
  const fetchedItems = await getUserOrders(session.backendToken, status || undefined, PAGE_SIZE + 1, 0);
  const items = (fetchedItems as UserAppOrder[]) || [];
  const hasNextPage = items.length > PAGE_SIZE;
  const displayItems = hasNextPage ? items.slice(0, PAGE_SIZE) : items;

  return (
    <main className="min-h-screen bg-[#f6f8fb] px-4 pb-28 pt-5">
      <div className="mx-auto w-full max-w-md">
        <UserTransactionPageContent
          initialItems={displayItems}
          initialHasNextPage={hasNextPage}
          status={status}
          authToken={session.backendToken}
        />
      </div>
      <UserBottomNav />
    </main>
  );
}
