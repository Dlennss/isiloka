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
  title: "Paket Telepon Semua Operator | Isiloka",
  description: "Pilih paket telepon semua operator di Isiloka dengan pilihan produk yang jelas dan ringkas.",
  path: "/paket-telepon",
  keywords: ["paket telepon", "nelpon semua operator", "paket telepon operator", "pulsakilat"],
});

export default async function GuestPaketTeleponPage() {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const categories = (await getCategories()) as UserCategoryItem[];
  const phoneCategory = pickCategory(categories, "telepon");
  const brands = phoneCategory ? await getBrandsByKategori(String(phoneCategory.id)) : [];
  const collectionJsonLd = buildCollectionJsonLd({
    title: "Paket Telepon Semua Operator | Isiloka",
    description: "Pilih paket telepon semua operator di Isiloka dengan pilihan produk yang jelas dan ringkas.",
    path: "/paket-telepon",
    itemNames: brands.map((brand) => brand.nama),
  });
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "Paket Telepon", path: "/paket-telepon" },
  ]);

  return (
    <main className="bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4 pt-4">
        {!phoneCategory ? (
          <section className="grid min-h-40 place-items-center rounded-md border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_10px_28px_rgba(15,23,42,0.08)]">
            Kategori paket telepon belum tersedia.
          </section>
        ) : brands.length === 0 ? (
          <section className="grid min-h-40 place-items-center rounded-md border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_10px_28px_rgba(15,23,42,0.08)]">
            Brand paket telepon belum tersedia.
          </section>
        ) : (
          <GuestPaketDataQuickOrder
            kategoriId={String(phoneCategory.id)}
            brands={brands}
            authToken={session?.backendToken}
            buyerRole={session?.user?.role}
            title="Paket Telepon"
            productLabel="paket telepon"
            showGroupTabs={false}
          />
        )}
      </div>

      <GuestBottomNav isLoggedIn={!!session?.backendToken} />
    </main>
  );
}
