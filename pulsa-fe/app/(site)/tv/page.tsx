import type { Metadata } from "next";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/nextauth";
import { getBrandsByKategori } from "@/lib/api.products";
import type { UserSession } from "@/components/user/types";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { ProviderBrandPicker } from "@/components/guest/ProviderBrandPicker";
import { getDedicatedGuestBrandPath } from "@/lib/dedicated-category-brand-routes";
import { buildBreadcrumbJsonLd, buildCollectionJsonLd, buildPageMetadata } from "@/lib/site-search";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

export const metadata: Metadata = buildPageMetadata({
  title: "TV Berlangganan | Isiloka",
  description: "Beli paket atau bayar layanan TV berlangganan di Isiloka dengan pilihan brand yang jelas.",
  path: "/tv",
  keywords: ["tv berlangganan", "bayar tv online", "voucher tv", "pulsakilat"],
});

export default async function TVPage() {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const brands = await getBrandsByKategori("7");
  const collectionJsonLd = buildCollectionJsonLd({
    title: "TV Berlangganan | Isiloka",
    description: "Beli paket atau bayar layanan TV berlangganan di Isiloka dengan pilihan brand yang jelas.",
    path: "/tv",
    itemNames: brands.map((brand) => brand.nama),
  });
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "TV", path: "/tv" },
  ]);

  return (
    <main className="bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4 pt-4">
        {brands.length === 0 ? (
          <section className="grid min-h-40 place-items-center rounded-md border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_10px_28px_rgba(15,23,42,0.08)]">
            Brand TV belum tersedia.
          </section>
        ) : (
          <ProviderBrandPicker
            items={brands.map((brand) => ({
              brand,
              href: getDedicatedGuestBrandPath("7", brand) || `/kategori/7/brand/${brand.id}?name=${encodeURIComponent(brand.nama)}`,
            }))}
          />
        )}
      </div>

      <GuestBottomNav isLoggedIn={!!session?.backendToken} />
    </main>
  );
}
