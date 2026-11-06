import { getAppServerSession } from "@/lib/server-auth";
import { CatalogCategory } from "@/components/shared/CatalogCategory";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";

export default async function GuestKategoriPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const session = await getAppServerSession();
  return (
    <main className="min-h-screen bg-sky-50">
      <CatalogCategory id={id} />
      <GuestBottomNav isLoggedIn={!!session?.backendToken} />
    </main>
  );
}
