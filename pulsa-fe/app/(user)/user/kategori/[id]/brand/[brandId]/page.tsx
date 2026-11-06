import { getAppServerSession } from "@/lib/server-auth";
import { CatalogBrand } from "@/components/shared/CatalogBrand";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import { UserAuthClientSync } from "@/components/user/UserAuthClientSync";

export const dynamic = "force-dynamic";

export default async function BrandProductsPage({ params }: { params: Promise<{ id: string; brandId: string }> }) {
  const { id, brandId } = await params;
  const session = await getAppServerSession();
  return <main className="min-h-screen bg-sky-50">
    {session?.backendToken ? <UserAuthClientSync backendToken={session.backendToken} /> : null}
    <CatalogBrand id={id} brandId={brandId} mode="user" authToken={session?.backendToken} buyerRole={session?.user?.role} />
    <UserBottomNav />
  </main>;
}
