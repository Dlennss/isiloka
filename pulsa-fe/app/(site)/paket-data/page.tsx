import type { Metadata } from "next";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/nextauth";
import { getBrandsByKategori, getCategories } from "@/lib/api.products";
import type { UserCategoryItem, UserSession } from "@/components/user/types";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { GuestPaketDataQuickOrder } from "@/components/guest/GuestPaketDataQuickOrder";
import { buildBreadcrumbJsonLd, buildCollectionJsonLd, buildPageMetadata } from "@/lib/site-search";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

function pickCategory(categories: UserCategoryItem[], keyword: string) {
  return categories.find((item) => item.aktif && item.nama.toLowerCase().includes(keyword));
}

export const metadata: Metadata = buildPageMetadata({
  title: "Paket Data Semua Operator | Isiloka",
  description: "Beli paket data harian, mingguan, dan bulanan semua operator di Isiloka dengan pilihan grup produk yang lebih jelas.",
  path: "/paket-data",
  keywords: ["paket data", "kuota internet", "paket data telkomsel", "paket data indosat", "pulsakilat"],
});

export default async function GuestPaketDataPage() {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const categories = (await getCategories()) as UserCategoryItem[];
  const dataCategory = pickCategory(categories, "data");
  const brands = dataCategory ? await getBrandsByKategori(String(dataCategory.id)) : [];
  const collectionJsonLd = buildCollectionJsonLd({
    title: "Paket Data Semua Operator | Isiloka",
    description: "Beli paket data harian, mingguan, dan bulanan semua operator di Isiloka dengan pilihan grup produk yang lebih jelas.",
    path: "/paket-data",
    itemNames: brands.map((brand) => brand.nama),
  });
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "Paket Data", path: "/paket-data" },
  ]);

  return (
    <main className="bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4 pt-4">
        {!dataCategory ? (
          <section className="grid min-h-40 place-items-center rounded-md border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_10px_28px_rgba(15,23,42,0.08)]">
            Kategori paket data belum tersedia.
          </section>
        ) : brands.length === 0 ? (
          <section className="grid min-h-40 place-items-center rounded-md border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_10px_28px_rgba(15,23,42,0.08)]">
            Brand paket data belum tersedia.
          </section>
        ) : (
          <GuestPaketDataQuickOrder
            kategoriId={String(dataCategory.id)}
            brands={brands}
            authToken={session?.backendToken}
            buyerRole={session?.user?.role}
          />
        )}
      </div>

      <GuestBottomNav isLoggedIn={!!session?.backendToken} />
    </main>
  );
}
