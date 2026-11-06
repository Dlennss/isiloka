import { getAppServerSession } from "@/lib/server-auth";
import { CatalogCategory } from "@/components/shared/CatalogCategory";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import { UserAuthClientSync } from "@/components/user/UserAuthClientSync";

export default async function UserKategoriPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const session = await getAppServerSession();
  return (
    <main className="min-h-screen bg-sky-50">
      {session?.backendToken ? <UserAuthClientSync backendToken={session.backendToken} /> : null}
      <CatalogCategory id={id} mode="user" />
      <UserBottomNav />
    </main>
  );
}
