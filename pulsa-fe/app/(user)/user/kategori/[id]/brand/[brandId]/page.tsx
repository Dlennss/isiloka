import { getAppServerSession } from "@/lib/server-auth";
import { getProductsByBrand } from "@/lib/api.products";
import type { UserSession } from "@/components/user/types";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import { UserAuthClientSync } from "@/components/user/UserAuthClientSync";
import { UserProductGrid } from "@/components/user/UserProductGrid";
import { BPJSBrandFlow } from "@/components/shared/BPJSBrandFlow";
import { EMoneyBrandFlow } from "@/components/shared/EMoneyBrandFlow";
import { GuestPulsaQuickOrder } from "@/components/guest/GuestPulsaQuickOrder";
import { GuestPaketDataQuickOrder } from "@/components/guest/GuestPaketDataQuickOrder";
import { RetailBillingEntryFlow } from "@/components/shared/RetailBillingEntryFlow";

export const dynamic = "force-dynamic";
export const revalidate = 0;

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

type PageProps = {
  params: Promise<{ id: string; brandId: string }>;
};

export default async function UserBrandProductsPage({ params }: PageProps) {
  const session = (await getAppServerSession()) as SessionShape | null;
  const user = session?.user ?? null;
  const backendToken = session?.backendToken;
  const isLoggedIn = Boolean(backendToken);
  const buyerRole = String(user?.role || "").trim().toLowerCase();

  const [{ id, brandId }] = await Promise.all([params]);
  const products = await getProductsByBrand(id, brandId);
  const brand = products[0]?.brand_nama || "Brand";
  const categoryName = products[0]?.kategori_nama || "";
  const isDataCategory = String(categoryName).toUpperCase().includes("DATA");
  const normalizedCategoryName = String(categoryName).toUpperCase();
  const isWalletCategory = normalizedCategoryName.includes("E-WALLET") || normalizedCategoryName.includes("E-MONEY");
  const isBillingCategory = !isWalletCategory && ["7", "11", "17", "18", "20"].includes(String(id));
  const billingPlaceholder = id === "11"
    ? "Masukkan ID pelanggan / nomor meter"
    : id === "17"
      ? "Masukkan nomor pelanggan PDAM"
      : id === "18"
        ? "Masukkan nomor HP pascabayar"
        : id === "20"
          ? "Masukkan ID pelanggan gas"
          : "Masukkan ID pelanggan";
  const billingDescription = `Masukkan data pelanggan ${brand} terlebih dulu sebelum memilih produk pembayaran.`;

  return (
    <main className="min-h-screen bg-sky-50">
      {backendToken ? <UserAuthClientSync backendToken={backendToken} /> : null}
      <div className="space-y-4 px-4">
        <section>
          {products.length === 0 ? (
            <div className="grid min-h-40 place-items-center rounded-3xl border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_8px_22px_rgba(15,23,42,0.13)]">
              Belum ada produk aktif untuk brand ini.
            </div>
          ) : id === "19" && brandId === "171" ? (
            <BPJSBrandFlow items={products} authToken={backendToken} buyerRole={buyerRole} />
          ) : isDataCategory ? (
            <GuestPaketDataQuickOrder
              kategoriId={String(id)}
              brands={[]}
              authToken={backendToken}
              buyerRole={buyerRole}
              forcedBrand={{ id: Number(brandId), nama: brand, aktif: true }}
              brandHrefPrefix="/user/kategori"
            />
          ) : id === "2" || isWalletCategory ? (
            <EMoneyBrandFlow items={products} isLoggedIn={isLoggedIn} authToken={backendToken} mode="user" buyerRole={buyerRole} />
          ) : id === "1" ? (
            <GuestPulsaQuickOrder
              kategoriId={String(id)}
              brands={[]}
              authToken={backendToken}
              buyerRole={buyerRole}
              forcedBrand={{ id: Number(brandId), nama: brand, aktif: true }}
            />
          ) : isBillingCategory ? (
            <RetailBillingEntryFlow
              title={categoryName || brand}
              description={billingDescription}
              placeholder={billingPlaceholder}
              items={products}
              mode="user"
              authToken={backendToken}
              buyerRole={buyerRole}
            />
          ) : (
            <UserProductGrid items={products} isLoggedIn={isLoggedIn} authToken={backendToken} buyerRole={buyerRole} />
          )}
        </section>
      </div>

      <UserBottomNav />
    </main>
  );
}
