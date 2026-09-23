import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/nextauth";
import { getBrandsByKategori, getProductsByBrand } from "@/lib/api.products";
import type { UserSession } from "@/components/user/types";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { GuestPaketDataQuickOrder } from "@/components/guest/GuestPaketDataQuickOrder";
import { findBrandByDedicatedSlug } from "@/lib/dedicated-category-brand-routes";
import { getRichProductImageUrl } from "@/lib/product-rich-images";
import { buildBreadcrumbJsonLd, buildCollectionJsonLd, buildFaqJsonLd, buildPageMetadata, buildProductItemListJsonLd } from "@/lib/site-search";
import { toTitleCase } from "@/lib/text";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

type PageProps = {
  params: Promise<{ slug: string }>;
};

function getLowestPrice(products: Awaited<ReturnType<typeof getProductsByBrand>>) {
  const values = products
    .map((item) => Number(item.harga_guest_final ?? item.harga_dasar_app ?? 0))
    .filter((value) => Number.isFinite(value) && value > 0);

  return values.length > 0 ? Math.min(...values) : null;
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const brands = await getBrandsByKategori("8");
  const { slug } = await params;
  const brand = findBrandByDedicatedSlug("8", brands, slug);
  if (!brand) {
    return buildPageMetadata({
      title: "Aktivasi Perdana | Isiloka",
      description: "Pilih produk aktivasi perdana operator di Isiloka.",
      path: "/aktivasi-perdana",
    });
  }

  const products = await getProductsByBrand("8", String(brand.id));
  const lowestPrice = getLowestPrice(products);
  const brandTitle = toTitleCase(brand.nama);

  return buildPageMetadata({
    title: `Aktivasi Perdana ${brandTitle} | Isiloka`,
    description: lowestPrice
      ? `Pilih produk aktivasi perdana ${brandTitle} di Isiloka. Produk aktif tersedia mulai Rp ${lowestPrice.toLocaleString("id-ID")}.`
      : `Pilih produk aktivasi perdana ${brandTitle} di Isiloka dengan pilihan yang jelas.`,
    path: `/aktivasi-perdana/${slug}`,
    keywords: [`aktivasi perdana ${brand.nama.toLowerCase()}`, `${brand.nama.toLowerCase()} perdana`, "pulsakilat"],
    imageUrl: getRichProductImageUrl({ brandName: brandTitle, categoryName: "Aktivasi Perdana", items: products }),
  });
}

export default async function AktivasiPerdanaBrandPage({ params }: PageProps) {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const brands = await getBrandsByKategori("8");
  const { slug } = await params;
  const brand = findBrandByDedicatedSlug("8", brands, slug);
  if (!brand) notFound();
  const products = await getProductsByBrand("8", String(brand.id));
  const brandTitle = toTitleCase(brand.nama);
  const collectionJsonLd = buildCollectionJsonLd({
    title: `Aktivasi Perdana ${brandTitle} | Isiloka`,
    description: `Pilih produk aktivasi perdana ${brandTitle} di Isiloka dengan pilihan yang jelas.`,
    path: `/aktivasi-perdana/${slug}`,
    itemNames: products.slice(0, 12).map((item) => item.nama),
  });
  const productJsonLd = buildProductItemListJsonLd({
    title: `Aktivasi Perdana ${brandTitle} | Isiloka`,
    path: `/aktivasi-perdana/${slug}`,
    brandName: brandTitle,
    categoryName: "Aktivasi Perdana",
    items: products.slice(0, 24),
  });
  const faqJsonLd = buildFaqJsonLd([
    {
      question: `Apakah produk aktivasi perdana ${brandTitle} di Isiloka tersedia lengkap?`,
      answer: `Ya. Isiloka menampilkan produk aktivasi perdana ${brandTitle} yang aktif agar pembeli bisa memilih sesuai kebutuhan.`,
    },
    {
      question: `Bagaimana cara membeli aktivasi perdana ${brandTitle} di Isiloka?`,
      answer: `Pilih produk aktivasi perdana ${brandTitle}, masukkan nomor tujuan bila diperlukan, lalu lanjutkan pembayaran.`,
    },
    {
      question: `Apakah halaman aktivasi perdana ${brandTitle} ini cocok untuk pembeli umum dan member?`,
      answer: `Bisa. Halaman ini dapat dipakai pembeli umum maupun member Isiloka untuk transaksi pribadi dan penjualan ulang.`,
    },
  ]);
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "Aktivasi Perdana", path: "/aktivasi-perdana" },
    { name: brandTitle, path: `/aktivasi-perdana/${slug}` },
  ]);

  return (
    <main className="bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(productJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(faqJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4 pt-4">
        <GuestPaketDataQuickOrder
          kategoriId="8"
          brands={brands}
          authToken={session?.backendToken}
          buyerRole={session?.user?.role}
          forcedBrand={brand}
          title="Aktivasi Perdana"
          productLabel="aktivasi perdana"
        />
      </div>

      <GuestBottomNav isLoggedIn={!!session?.backendToken} />
    </main>
  );
}
