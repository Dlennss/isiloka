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
  title: "HP Pascabayar | Isiloka",
  description: "Cek dan bayar tagihan HP pascabayar di Isiloka dengan pilihan operator yang jelas.",
  path: "/hp-pascabayar",
  keywords: ["hp pascabayar", "bayar tagihan hp", "tagihan telepon seluler", "pulsakilat"],
});

export default async function HPPascabayarPage() {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const brands = await getBrandsByKategori("18");
  const collectionJsonLd = buildCollectionJsonLd({
    title: "HP Pascabayar | Isiloka",
    description: "Cek dan bayar tagihan HP pascabayar di Isiloka dengan pilihan operator yang jelas.",
    path: "/hp-pascabayar",
    itemNames: brands.map((brand) => brand.nama),
  });
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "HP Pascabayar", path: "/hp-pascabayar" },
  ]);

  return (
    <main className="bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4 pt-4">
        {brands.length === 0 ? (
          <section className="grid min-h-40 place-items-center rounded-md border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_10px_28px_rgba(15,23,42,0.08)]">
            Brand HP pascabayar belum tersedia.
          </section>
        ) : (
          <ProviderBrandPicker
            items={brands.map((brand) => ({
              brand,
              href: getDedicatedGuestBrandPath("18", brand) || `/kategori/18/brand/${brand.id}?name=${encodeURIComponent(brand.nama)}`,
            }))}
          />
        )}
      </div>

      <GuestBottomNav isLoggedIn={!!session?.backendToken} />
    </main>
  );
}
