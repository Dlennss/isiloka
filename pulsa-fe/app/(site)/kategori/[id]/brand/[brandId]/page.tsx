import type { Metadata } from "next";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/nextauth";
import { getProductsByBrand } from "@/lib/api.products";
import type { UserSession } from "@/components/user/types";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { GuestProductGrid } from "@/components/guest/GuestProductGrid";
import { BPJSBrandFlow } from "@/components/shared/BPJSBrandFlow";
import { GuestPulsaQuickOrder } from "@/components/guest/GuestPulsaQuickOrder";
import { GuestPaketDataQuickOrder } from "@/components/guest/GuestPaketDataQuickOrder";
import { EMoneyBrandFlow } from "@/components/shared/EMoneyBrandFlow";
import { RetailBillingEntryFlow } from "@/components/shared/RetailBillingEntryFlow";
import { getRichProductImageUrl } from "@/lib/product-rich-images";
import { buildBreadcrumbJsonLd, buildCollectionJsonLd, buildPageMetadata, buildProductItemListJsonLd } from "@/lib/site-search";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

type PageProps = {
  params: Promise<{ id: string; brandId: string }>;
};

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { id, brandId } = await params;
  const products = await getProductsByBrand(id, brandId);
  const brand = products[0]?.brand_nama || "Brand";
  const categoryName = products[0]?.kategori_nama || "Produk";

  return buildPageMetadata({
    title: `${brand} | ${categoryName} Isiloka`,
    description: `Lihat produk ${brand} pada kategori ${categoryName} di Isiloka.`,
    path: `/kategori/${id}/brand/${brandId}`,
    keywords: [String(brand).toLowerCase(), String(categoryName).toLowerCase(), "pulsakilat"],
    imageUrl: getRichProductImageUrl({ brandName: String(brand), categoryName: String(categoryName), items: products }),
  });
}

export default async function GuestBrandProductsPage({ params }: PageProps) {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const backendToken = session?.backendToken;
  const buyerRole = String(session?.user?.role || "").trim().toLowerCase();
  const isLoggedIn = Boolean(backendToken);

  const [{ id, brandId }] = await Promise.all([params]);
  const products = await getProductsByBrand(id, brandId);
  const brand = products[0]?.brand_nama || "Brand";
  const categoryName = products[0]?.kategori_nama || "";
  const collectionJsonLd = buildCollectionJsonLd({
    title: `${brand} | ${categoryName || "Produk"} Isiloka`,
    description: `Lihat produk ${brand} pada kategori ${categoryName || "produk"} di Isiloka.`,
    path: `/kategori/${id}/brand/${brandId}`,
    itemNames: products.slice(0, 12).map((item) => String(item.nama || "").trim()).filter(Boolean),
  });
  const productJsonLd = buildProductItemListJsonLd({
    title: `${brand} | ${categoryName || "Produk"} Isiloka`,
    path: `/kategori/${id}/brand/${brandId}`,
    brandName: brand,
    categoryName: categoryName || "Produk",
    items: products.slice(0, 24),
  });
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "Kategori", path: "/kategori" },
    { name: categoryName || "Produk", path: `/kategori/${id}` },
    { name: brand, path: `/kategori/${id}/brand/${brandId}` },
  ]);
  const isDataCategory = String(categoryName).toUpperCase().includes("DATA");
  const isBillingCategory = ["3", "7", "11", "17", "18", "20"].includes(String(id));
  const billingPlaceholder = id === "11"
    ? "Masukkan ID pelanggan / nomor meter"
    : id === "17"
      ? "Masukkan nomor pelanggan PDAM"
      : id === "18"
        ? "Masukkan nomor HP pascabayar"
        : id === "20"
          ? "Masukkan ID pelanggan gas"
          : "Masukkan ID pelanggan";
  const billingDescription = `Masukkan data pelanggan ${brand} terlebih dulu sebelum memilih produk pembayaran.`;

  return (
    <main className="min-h-screen bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      {!isBillingCategory ? (
        <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(productJsonLd) }} />
      ) : null}
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4">
        <section>
          {products.length === 0 ? (
            <div className="grid min-h-40 place-items-center rounded-3xl border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_8px_22px_rgba(15,23,42,0.13)]">
              Belum ada produk aktif untuk brand ini.
            </div>
          ) : id === "1" ? (
            <GuestPulsaQuickOrder
              kategoriId={id}
              brands={[{ id: Number(brandId), nama: brand, aktif: true }]}
              authToken={backendToken}
              buyerRole={buyerRole}
              forcedBrand={{ id: Number(brandId), nama: brand, aktif: true }}
            />
          ) : id === "19" && brandId === "171" ? (
            <BPJSBrandFlow items={products} authToken={backendToken} buyerRole={buyerRole} />
          ) : isDataCategory ? (
            <GuestPaketDataQuickOrder
              kategoriId={String(id)}
              brands={[]}
              authToken={backendToken}
              buyerRole={buyerRole}
              forcedBrand={{ id: Number(brandId), nama: brand, aktif: true }}
            />
          ) : id === "2" ? (
            <EMoneyBrandFlow items={products} isLoggedIn={isLoggedIn} authToken={backendToken} mode={isLoggedIn ? "user" : "guest"} buyerRole={buyerRole} />
          ) : isBillingCategory ? (
            <RetailBillingEntryFlow
              title={categoryName || brand}
              description={billingDescription}
              placeholder={billingPlaceholder}
              items={products}
              mode={isLoggedIn ? "user" : "guest"}
              authToken={backendToken}
              buyerRole={buyerRole}
            />
          ) : (
            <GuestProductGrid items={products} isLoggedIn={isLoggedIn} authToken={backendToken} buyerRole={buyerRole} />
          )}
        </section>
      </div>

      <GuestBottomNav isLoggedIn={isLoggedIn} />
    </main>
  );
}
