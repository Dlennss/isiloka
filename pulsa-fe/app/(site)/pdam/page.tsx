import type { Metadata } from "next";
import { getBrandsByKategori } from "@/lib/api.products";
import { PdamBrandSearch } from "@/components/guest/PdamBrandSearch";
import { buildBreadcrumbJsonLd, buildCollectionJsonLd, buildPageMetadata } from "@/lib/site-search";

export const dynamic = "force-dynamic";

export const metadata: Metadata = buildPageMetadata({
  title: "Tagihan PDAM | Isiloka",
  description: "Cek dan bayar tagihan PDAM berbagai daerah di Isiloka dengan langkah yang ringkas dan pilihan wilayah yang jelas.",
  path: "/pdam",
  keywords: ["bayar pdam online", "tagihan pdam", "cek tagihan air", "pdam pulsakilat"],
});

export default async function PDAMPage() {
  const brands = await getBrandsByKategori("17");
  const collectionJsonLd = buildCollectionJsonLd({
    title: "Tagihan PDAM | Isiloka",
    description: "Cek dan bayar tagihan PDAM berbagai daerah di Isiloka dengan langkah yang ringkas dan pilihan wilayah yang jelas.",
    path: "/pdam",
    itemNames: brands.map((brand) => brand.nama),
  });
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "PDAM", path: "/pdam" },
  ]);

  return (
    <main className="bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4 pt-4">
        <PdamBrandSearch brands={brands} />
      </div>
    </main>
  );
}
