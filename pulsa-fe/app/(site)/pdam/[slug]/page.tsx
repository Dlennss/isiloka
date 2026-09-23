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
  const brands = await getBrandsByKategori("17");
  const { slug } = await params;
  const brand = findBrandByDedicatedSlug("17", brands, slug);
  if (!brand) {
    return buildPageMetadata({
      title: "Tagihan PDAM | Isiloka",
      description: "Cek dan bayar tagihan PDAM di Isiloka.",
      path: "/pdam",
    });
  }

  return buildPageMetadata({
    title: `${brand.nama} | Tagihan PDAM Isiloka`,
    description: `Cek dan bayar tagihan ${brand.nama} di Isiloka dengan proses cepat dan ringkas.`,
    path: `/pdam/${slug}`,
    keywords: [`tagihan ${brand.nama.toLowerCase()}`, `bayar ${brand.nama.toLowerCase()}`, "pdam online", "pulsakilat"],
  });
}

export default async function PdamSlugPage({ params }: PageProps) {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const brands = await getBrandsByKategori("17");
  const { slug } = await params;
  const brand = findBrandByDedicatedSlug("17", brands, slug);
  if (!brand) notFound();

  const items = await getProductsByBrand("17", String(brand.id));
  const collectionJsonLd = buildCollectionJsonLd({
    title: `${brand.nama} | Tagihan PDAM Isiloka`,
    description: `Cek dan bayar tagihan ${brand.nama} di Isiloka dengan proses cepat dan ringkas.`,
    path: `/pdam/${slug}`,
    itemNames: items.slice(0, 8).map((item) => String(item.nama || "").trim()).filter(Boolean),
  });
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "PDAM", path: "/pdam" },
    { name: brand.nama, path: `/pdam/${slug}` },
  ]);

  return (
    <main className="bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4 pt-4">
        <RetailBillingEntryFlow
          title={brand.nama}
          description={`Masukkan nomor pelanggan ${brand.nama} untuk cek tagihan.`}
          placeholder="Masukkan nomor pelanggan PDAM"
          items={items}
          mode="guest"
          minDestLength={6}
          maxDestLength={14}
        />
      </div>

      <GuestBottomNav isLoggedIn={!!session?.backendToken} />
    </main>
  );
}
