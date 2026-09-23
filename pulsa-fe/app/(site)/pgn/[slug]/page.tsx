import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/nextauth";
import { getBrandsByKategori, getProductsByBrand } from "@/lib/api.products";
import type { UserSession } from "@/components/user/types";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { RetailBillingEntryFlow } from "@/components/shared/RetailBillingEntryFlow";
import { findBrandByDedicatedSlug } from "@/lib/dedicated-category-brand-routes";
import { buildBreadcrumbJsonLd, buildCollectionJsonLd, buildPageMetadata } from "@/lib/site-search";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

type PageProps = {
  params: Promise<{ slug: string }>;
};

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const brands = await getBrandsByKategori("20");
  const { slug } = await params;
  const brand = findBrandByDedicatedSlug("20", brands, slug);
  if (!brand) {
    return buildPageMetadata({
      title: "Tagihan PGN | Isiloka",
      description: "Cek dan bayar tagihan gas PGN di Isiloka.",
      path: "/pgn",
    });
  }

  return buildPageMetadata({
    title: `${brand.nama} | Tagihan PGN Isiloka`,
    description: `Cek dan bayar tagihan ${brand.nama} di Isiloka dengan proses yang ringkas.`,
    path: `/pgn/${slug}`,
    keywords: [`tagihan ${brand.nama.toLowerCase()}`, `bayar ${brand.nama.toLowerCase()}`, "pgn online", "pulsakilat"],
  });
}

export default async function PGNBrandPage({ params }: PageProps) {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const brands = await getBrandsByKategori("20");
  const { slug } = await params;
  const brand = findBrandByDedicatedSlug("20", brands, slug);
  if (!brand) notFound();

  const items = await getProductsByBrand("20", String(brand.id));
  const collectionJsonLd = buildCollectionJsonLd({
    title: `${brand.nama} | Tagihan PGN Isiloka`,
    description: `Cek dan bayar tagihan ${brand.nama} di Isiloka dengan proses yang ringkas.`,
    path: `/pgn/${slug}`,
    itemNames: items.slice(0, 8).map((item) => String(item.nama || "").trim()).filter(Boolean),
  });
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "PGN", path: "/pgn" },
    { name: brand.nama, path: `/pgn/${slug}` },
  ]);

  return (
    <main className="bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4 pt-4">
        <RetailBillingEntryFlow
          title="Tagihan PGN"
          description=""
          placeholder="Masukkan nomor pelanggan"
          items={items}
          mode="guest"
          logoSrc="/images/gas/Logo_PGN.png"
          logoAlt="PGN"
        />
      </div>

      <GuestBottomNav isLoggedIn={!!session?.backendToken} />
    </main>
  );
}
