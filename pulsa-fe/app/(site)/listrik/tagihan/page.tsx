import type { Metadata } from "next";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/nextauth";
import { getBrandsByKategori, getCategories, getProductsByBrand } from "@/lib/api.products";
import type { UserBrandItem, UserCategoryItem, UserSession } from "@/components/user/types";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { RetailBillingEntryFlow } from "@/components/shared/RetailBillingEntryFlow";
import { buildBreadcrumbJsonLd, buildCollectionJsonLd, buildPageMetadata } from "@/lib/site-search";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

function pickPLNBrand(brands: UserBrandItem[]) {
  return brands.find((item) => item.aktif && String(item.nama || "").trim().toLowerCase() === "pln") ?? null;
}

function normalizeName(value: string) {
  return value.trim().toLowerCase();
}

function pickCategory(categories: UserCategoryItem[], keyword: string) {
  return categories.find((item) => item.aktif && normalizeName(item.nama).includes(keyword));
}

function pickPLNCategory(categories: UserCategoryItem[]) {
  return pickCategory(categories, "pln") ?? pickCategory(categories, "listrik");
}

export const metadata: Metadata = buildPageMetadata({
  title: "Tagihan Listrik PLN | Isiloka",
  description: "Bayar tagihan listrik PLN di Isiloka dengan alur yang ringkas untuk pelanggan rumah tangga dan agen.",
  path: "/listrik/tagihan",
  keywords: ["tagihan listrik", "bayar listrik", "pln pascabayar", "pulsakilat listrik"],
});

export default async function GuestListrikTagihanPage() {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const isLoggedIn = Boolean(session?.backendToken);
  const backendToken = session?.backendToken;
  const buyerRole = String(session?.user?.role || "").trim().toLowerCase();
  const categories = (await getCategories()) as UserCategoryItem[];
  const listrikCategory = pickPLNCategory(categories);
  const brands = listrikCategory ? await getBrandsByKategori(String(listrikCategory.id)) : [];
  const plnBrand = pickPLNBrand(brands);
  const products = listrikCategory && plnBrand ? await getProductsByBrand(String(listrikCategory.id), String(plnBrand.id)) : [];
  const collectionJsonLd = buildCollectionJsonLd({
    title: "Tagihan Listrik PLN | Isiloka",
    description: "Bayar tagihan listrik PLN di Isiloka dengan alur yang ringkas untuk pelanggan rumah tangga dan agen.",
    path: "/listrik/tagihan",
    itemNames: products.slice(0, 5).map((item) => item.nama),
  });
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "Listrik", path: "/listrik" },
    { name: "Tagihan Listrik", path: "/listrik/tagihan" },
  ]);

  return (
    <main className="bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4 pt-4">
        <RetailBillingEntryFlow
          title="Tagihan Listrik"
          description=""
          placeholder="Masukkan nomor meter pelanggan"
          items={products}
          mode={isLoggedIn ? "user" : "guest"}
          authToken={backendToken}
          buyerRole={buyerRole}
        />
      </div>

      <GuestBottomNav isLoggedIn={!!session?.backendToken} />
    </main>
  );
}
