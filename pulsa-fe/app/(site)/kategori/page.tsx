import { getAppServerSession } from "@/lib/server-auth";
import { getCategories } from "@/lib/api.products";
import type { UserCategoryItem } from "@/components/user/types";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { ServiceDirectory } from "@/components/shared/ServiceDirectory";

export default async function GuestKategoriPage() {
  const session = await getAppServerSession();
  const categories = await getCategories() as UserCategoryItem[];
  return (
    <main className="min-h-screen bg-sky-50">
      <div className="px-4 pt-4"><ServiceDirectory categories={categories} /></div>
      <GuestBottomNav isLoggedIn={!!session?.backendToken} />
    </main>
  );
}
