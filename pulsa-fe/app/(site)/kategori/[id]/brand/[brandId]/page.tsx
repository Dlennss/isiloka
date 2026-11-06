import { getAppServerSession } from "@/lib/server-auth";
import { CatalogBrand } from "@/components/shared/CatalogBrand";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { UserAuthClientSync } from "@/components/user/UserAuthClientSync";

export const dynamic = "force-dynamic";

export default async function BrandProductsPage({ params }: { params: Promise<{ id: string; brandId: string }> }) {
  const { id, brandId } = await params;
  const session = await getAppServerSession();
  return <main className="min-h-screen bg-sky-50">
    {session?.backendToken ? <UserAuthClientSync backendToken={session.backendToken} /> : null}
    <CatalogBrand id={id} brandId={brandId} mode="guest" authToken={session?.backendToken} buyerRole={session?.user?.role} />
    <GuestBottomNav isLoggedIn={!!session?.backendToken} />
  </main>;
}
