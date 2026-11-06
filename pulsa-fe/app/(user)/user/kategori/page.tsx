import { getAppServerSession } from "@/lib/server-auth";
import { getCategories } from "@/lib/api.products";
import type { UserCategoryItem } from "@/components/user/types";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import { UserAuthClientSync } from "@/components/user/UserAuthClientSync";
import { ServiceDirectory } from "@/components/shared/ServiceDirectory";

export default async function UserAllCategoryPage() {
  const session = await getAppServerSession();
  const categories = await getCategories() as UserCategoryItem[];
  return (
    <main className="min-h-screen bg-sky-50">
      {session?.backendToken ? <UserAuthClientSync backendToken={session.backendToken} /> : null}
      <div className="px-4 pt-4"><ServiceDirectory categories={categories} mode="user" /></div>
      <UserBottomNav />
    </main>
  );
}
