import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/nextauth";
import { getBrandsByKategori, getProductsByBrand } from "@/lib/api.products";
import type { UserProductItem, UserSession } from "@/components/user/types";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { GuestProductGrid } from "@/components/guest/GuestProductGrid";
import { findBrandByDedicatedSlug } from "@/lib/dedicated-category-brand-routes";
import { getRichProductImageUrl } from "@/lib/product-rich-images";
import { RetailBillingEntryFlow } from "@/components/shared/RetailBillingEntryFlow";
import { buildBreadcrumbJsonLd, buildCollectionJsonLd, buildPageMetadata, buildProductItemListJsonLd } from "@/lib/site-search";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

type PageProps = {
  params: Promise<{ slug: string }>;
};

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const brands = await getBrandsByKategori("7");
  const { slug } = await params;
  const brand = findBrandByDedicatedSlug("7", brands, slug);
  if (!brand) {
    return buildPageMetadata({
      title: "TV Berlangganan | Isiloka",
      description: "Beli paket atau bayar layanan TV berlangganan di Isiloka.",
      path: "/tv",
    });
  }

  return buildPageMetadata({
    title: `${brand.nama} | TV Berlangganan Isiloka`,
    description: `Beli paket atau bayar layanan ${brand.nama} di Isiloka dengan proses yang ringkas.`,
    path: `/tv/${slug}`,
    keywords: [`${brand.nama.toLowerCase()}`, `bayar ${brand.nama.toLowerCase()}`, "tv berlangganan", "pulsakilat"],
    imageUrl: getRichProductImageUrl({ brandName: brand.nama, categoryName: "TV Berlangganan" }),
  });
}

function isTvCheckProduct(item: UserProductItem) {
  const sku = String(item.sku || "").toUpperCase().trim();
  const name = String(item.nama || "").toUpperCase().trim();
  return sku.startsWith("CEK") || sku.endsWith("C") || name.includes("CEK ");
}

function isTvPurchaseBrand(items: UserProductItem[]) {
  return items.some((item) => {
    const name = String(item.nama || "").toUpperCase();
    return name.includes("PAKET") || name.includes("SALDO") || name.includes("VOUCHER");
  });
}

function shouldUseTvBillingFlow(items: UserProductItem[]) {
  if (items.length === 0) return false;
  if (isTvPurchaseBrand(items)) return false;
  return items.some(isTvCheckProduct);
}

export default async function TVBrandPage({ params }: PageProps) {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const isLoggedIn = Boolean(session?.backendToken);
  const backendToken = session?.backendToken;
  const buyerRole = String(session?.user?.role || "").trim().toLowerCase();
  const brands = await getBrandsByKategori("7");
  const { slug } = await params;
  const brand = findBrandByDedicatedSlug("7", brands, slug);
  if (!brand) notFound();

  const products = await getProductsByBrand("7", String(brand.id));
  const useBillingFlow = shouldUseTvBillingFlow(products);
  const collectionJsonLd = buildCollectionJsonLd({
    title: `${brand.nama} | TV Berlangganan Isiloka`,
    description: `Beli paket atau bayar layanan ${brand.nama} di Isiloka dengan proses yang ringkas.`,
    path: `/tv/${slug}`,
    itemNames: products.slice(0, 12).map((item) => String(item.nama || "").trim()).filter(Boolean),
  });
  const productJsonLd = buildProductItemListJsonLd({
    title: `${brand.nama} | TV Berlangganan Isiloka`,
    path: `/tv/${slug}`,
    brandName: brand.nama,
    categoryName: "TV Berlangganan",
    items: products.slice(0, 24),
  });
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "TV", path: "/tv" },
    { name: brand.nama, path: `/tv/${slug}` },
  ]);

  return (
    <main className="min-h-screen bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      {!useBillingFlow ? (
        <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(productJsonLd) }} />
      ) : null}
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4">
        <section>
          {products.length === 0 ? (
            <div className="grid min-h-40 place-items-center rounded-3xl border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_8px_22px_rgba(15,23,42,0.13)]">
              Belum ada produk aktif untuk brand ini.
            </div>
          ) : useBillingFlow ? (
            <RetailBillingEntryFlow
              title={brand.nama}
              description={`Masukkan nomor pelanggan ${brand.nama} terlebih dulu sebelum melanjutkan pembayaran.`}
              placeholder="Masukkan nomor pelanggan"
              items={products}
              mode={isLoggedIn ? "user" : "guest"}
              authToken={backendToken}
              buyerRole={buyerRole}
            />
          ) : (
            <GuestProductGrid
              items={products}
              isLoggedIn={isLoggedIn}
              authToken={backendToken}
              buyerRole={buyerRole}
              destLabel="Nomor Pelanggan"
              destPlaceholder="Masukkan nomor pelanggan"
            />
          )}
        </section>
      </div>

      <GuestBottomNav isLoggedIn={isLoggedIn} />
    </main>
  );
}
