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
import { toTitleCase } from "@/lib/text";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

type PageProps = {
  params: Promise<{ slug: string }>;
};

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const brands = await getBrandsByKategori("3");
  const { slug } = await params;
  const brand = findBrandByDedicatedSlug("3", brands, slug);
  if (!brand) {
    return buildPageMetadata({
      title: "Internet Pascabayar | Isiloka",
      description: "Cek dan bayar tagihan internet pascabayar di Isiloka.",
      path: "/internet-pascabayar",
    });
  }

  return buildPageMetadata({
    title: `${toTitleCase(brand.nama)} | Internet Pascabayar Isiloka`,
    description: `Cek dan bayar tagihan ${toTitleCase(brand.nama)} di Isiloka dengan proses yang ringkas.`,
    path: `/internet-pascabayar/${slug}`,
    keywords: [`${String(brand.nama).toLowerCase()} pascabayar`, `bayar ${String(brand.nama).toLowerCase()}`, "internet pascabayar", "pulsakilat"],
  });
}

export default async function InternetPascabayarSlugPage({ params }: PageProps) {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const isLoggedIn = Boolean(session?.backendToken);
  const backendToken = session?.backendToken;
  const buyerRole = String(session?.user?.role || "").trim().toLowerCase();
  const brands = await getBrandsByKategori("3");
  const { slug } = await params;
  const brand = findBrandByDedicatedSlug("3", brands, slug);
  if (!brand) notFound();

  const items = await getProductsByBrand("3", String(brand.id));
  const brandTitle = toTitleCase(brand.nama);
  const collectionJsonLd = buildCollectionJsonLd({
    title: `${brandTitle} | Internet Pascabayar Isiloka`,
    description: `Cek dan bayar tagihan ${brandTitle} di Isiloka dengan proses yang ringkas.`,
    path: `/internet-pascabayar/${slug}`,
    itemNames: items.slice(0, 8).map((item) => String(item.nama || "").trim()).filter(Boolean),
  });
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "Internet Pascabayar", path: "/internet-pascabayar" },
    { name: brandTitle, path: `/internet-pascabayar/${slug}` },
  ]);

  return (
    <main className="bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4 pt-4">
        <RetailBillingEntryFlow
          title={brandTitle}
          description=""
          placeholder="Masukkan ID pelanggan"
          items={items}
          mode={isLoggedIn ? "user" : "guest"}
          authToken={backendToken}
          buyerRole={buyerRole}
          logoSrc="/images/internet/logo_iconet.png"
          logoAlt={brand.nama}
        />
      </div>

      <GuestBottomNav isLoggedIn={!!session?.backendToken} />
    </main>
  );
}
