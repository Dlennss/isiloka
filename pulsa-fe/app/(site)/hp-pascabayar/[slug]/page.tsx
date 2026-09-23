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
  const brands = await getBrandsByKategori("18");
  const { slug } = await params;
  const brand = findBrandByDedicatedSlug("18", brands, slug);
  if (!brand) {
    return buildPageMetadata({
      title: "HP Pascabayar | Isiloka",
      description: "Cek dan bayar tagihan HP pascabayar di Isiloka.",
      path: "/hp-pascabayar",
    });
  }

  return buildPageMetadata({
    title: `${toTitleCase(brand.nama)} | HP Pascabayar Isiloka`,
    description: `Cek dan bayar tagihan ${toTitleCase(brand.nama)} di Isiloka dengan proses yang ringkas.`,
    path: `/hp-pascabayar/${slug}`,
    keywords: [`${String(brand.nama).toLowerCase()} pascabayar`, `bayar ${String(brand.nama).toLowerCase()}`, "hp pascabayar", "pulsakilat"],
  });
}

export default async function HPPascabayarSlugPage({ params }: PageProps) {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const isLoggedIn = Boolean(session?.backendToken);
  const backendToken = session?.backendToken;
  const buyerRole = String(session?.user?.role || "").trim().toLowerCase();
  const brands = await getBrandsByKategori("18");
  const { slug } = await params;
  const brand = findBrandByDedicatedSlug("18", brands, slug);
  if (!brand) notFound();

  const items = await getProductsByBrand("18", String(brand.id));
  const brandTitle = toTitleCase(brand.nama);
  const collectionJsonLd = buildCollectionJsonLd({
    title: `${brandTitle} | HP Pascabayar Isiloka`,
    description: `Cek dan bayar tagihan ${brandTitle} di Isiloka dengan proses yang ringkas.`,
    path: `/hp-pascabayar/${slug}`,
    itemNames: items.slice(0, 8).map((item) => String(item.nama || "").trim()).filter(Boolean),
  });
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "HP Pascabayar", path: "/hp-pascabayar" },
    { name: brandTitle, path: `/hp-pascabayar/${slug}` },
  ]);

  return (
    <main className="bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4 pt-4">
        <RetailBillingEntryFlow
          title={brandTitle}
          description=""
          placeholder="Masukkan nomor HP pascabayar"
          items={items}
          mode={isLoggedIn ? "user" : "guest"}
          authToken={backendToken}
          buyerRole={buyerRole}
        />
      </div>

      <GuestBottomNav isLoggedIn={!!session?.backendToken} />
    </main>
  );
}
