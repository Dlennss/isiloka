import { getAppServerSession } from "@/lib/server-auth";
import { getBrandsByKategori, getProductsByBrand } from "@/lib/api.products";
import type { UserBrandItem, UserSession } from "@/components/user/types";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import { UserAuthClientSync } from "@/components/user/UserAuthClientSync";
import { RetailBillingEntryFlow } from "@/components/shared/RetailBillingEntryFlow";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

function pickPLNBrand(brands: UserBrandItem[]) {
  return brands.find((item) => item.aktif && String(item.nama || "").trim().toLowerCase() === "pln") ?? null;
}

export default async function UserListrikTagihanPage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  const backendToken = session?.backendToken;
  const user = session?.user ?? null;
  const buyerRole = String(user?.role || "").trim().toLowerCase();
  const brands = await getBrandsByKategori("11");
  const plnBrand = pickPLNBrand(brands);
  const products = plnBrand ? await getProductsByBrand("11", String(plnBrand.id)) : [];

  return (
    <main className="bg-sky-50">
      {backendToken ? <UserAuthClientSync backendToken={backendToken} /> : null}
      <div className="space-y-4 px-4 pt-4">
        <RetailBillingEntryFlow
          title="Tagihan Listrik"
          description=""
          placeholder="Masukkan nomor meter pelanggan"
          items={products}
          mode="user"
          authToken={backendToken}
          buyerRole={buyerRole}
        />
      </div>

      <UserBottomNav />
    </main>
  );
}
