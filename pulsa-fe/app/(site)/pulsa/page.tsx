import type { Metadata } from "next";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/nextauth";
import { getBrandsByKategori, getCategories } from "@/lib/api.products";
import type { UserCategoryItem, UserSession } from "@/components/user/types";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { GuestPulsaQuickOrder } from "@/components/guest/GuestPulsaQuickOrder";
import { buildBreadcrumbJsonLd, buildCollectionJsonLd, buildPageMetadata } from "@/lib/site-search";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

function pickCategory(categories: UserCategoryItem[], keyword: string) {
  return categories.find((item) => item.aktif && item.nama.toLowerCase().includes(keyword));
}

export const metadata: Metadata = buildPageMetadata({
  title: "Pulsa Semua Operator | Isiloka",
  description: "Isi pulsa Telkomsel, Indosat, XL, Axis, Smartfren, dan operator lain di Isiloka dengan pilihan nominal yang jelas dan transaksi cepat.",
  path: "/pulsa",
  keywords: ["pulsa semua operator", "isi pulsa online", "pulsa telkomsel", "pulsa indosat", "pulsakilat"],
});

export default async function GuestPulsaPage() {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const categories = (await getCategories()) as UserCategoryItem[];
  const pulsaCategory = pickCategory(categories, "pulsa");
  const brands = pulsaCategory ? await getBrandsByKategori(String(pulsaCategory.id)) : [];
  const collectionJsonLd = buildCollectionJsonLd({
    title: "Pulsa Semua Operator | Isiloka",
    description: "Isi pulsa Telkomsel, Indosat, XL, Axis, Smartfren, dan operator lain di Isiloka dengan pilihan nominal yang jelas dan transaksi cepat.",
    path: "/pulsa",
    itemNames: brands.map((brand) => brand.nama),
  });
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "Pulsa", path: "/pulsa" },
  ]);

  return (
    <main className="bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4 pt-4">
        {!pulsaCategory ? (
          <section className="grid min-h-40 place-items-center rounded-md border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_10px_28px_rgba(15,23,42,0.08)]">
            Kategori pulsa belum tersedia.
          </section>
        ) : brands.length === 0 ? (
          <section className="grid min-h-40 place-items-center rounded-md border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_10px_28px_rgba(15,23,42,0.08)]">
            Brand pulsa belum tersedia.
          </section>
        ) : (
          <GuestPulsaQuickOrder
            kategoriId={String(pulsaCategory.id)}
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
