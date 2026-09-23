import { Suspense } from "react";
import { getAppServerSession } from "@/lib/server-auth";
import { getCategories } from "@/lib/api.products";
import type { UserCategoryItem, UserSession } from "@/components/user/types";
import { UserCategoryGrid } from "@/components/user/UserCategoryGrid";
import { UserFavoriteTransactions, UserMonthlyBills, UserRecentActivity } from "@/components/user/UserMainSections";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import { UserAuthClientSync } from "@/components/user/UserAuthClientSync";
import { GuestAdsSection } from "@/components/guest/GuestAdsSection";
import { GuestAdsCarouselSkeleton } from "@/components/guest/GuestAdsCarouselSkeleton";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

export default async function UserAppHomePage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  const categories = (await getCategories()) as UserCategoryItem[];
  const role = String(session?.user?.role || "").trim().toLowerCase();
  const isAgent = role === "agent";

  return (
    <main className="bg-sky-50">
      {session?.backendToken ? <UserAuthClientSync backendToken={session.backendToken} /> : null}
      <div className="space-y-4 px-4 pt-4">
        <UserCategoryGrid items={categories} />
        <Suspense fallback={<GuestAdsCarouselSkeleton />}>
          <GuestAdsSection />
        </Suspense>
        <UserRecentActivity href="/user/kategori" />
        <UserFavoriteTransactions href="/user/kategori" />
        <UserMonthlyBills
          href="/user/listrik/tagihan"
          variant={isAgent ? "agent" : "user"}
          agentBills={[]}
        />
      </div>

      <UserBottomNav />
    </main>
  );
}
