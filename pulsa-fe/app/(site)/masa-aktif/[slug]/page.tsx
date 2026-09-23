import type { Metadata } from "next";
import Link from "next/link";
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
  const brands = await getBrandsByKategori("9");
  const { slug } = await params;
  const brand = findBrandByDedicatedSlug("9", brands, slug);
  if (!brand) {
    return buildPageMetadata({
      title: "Masa Aktif Semua Operator | Isiloka",
      description: "Perpanjang masa aktif nomor Anda di Isiloka.",
      path: "/masa-aktif",
    });
  }

  const products = await getProductsByBrand("9", String(brand.id));
  const lowestPrice = getLowestPrice(products);
  const brandTitle = toTitleCase(brand.nama);

  return buildPageMetadata({
    title: `Perpanjang Masa Aktif ${brandTitle} | Isiloka`,
    description: lowestPrice
      ? `Perpanjang masa aktif ${brandTitle} di Isiloka. Produk aktif tersedia mulai Rp ${lowestPrice.toLocaleString("id-ID")}.`
      : `Perpanjang masa aktif ${brandTitle} di Isiloka dengan pilihan produk yang jelas.`,
    path: `/masa-aktif/${slug}`,
    keywords: [`masa aktif ${brand.nama.toLowerCase()}`, `perpanjang masa aktif ${brand.nama.toLowerCase()}`, "pulsakilat"],
    imageUrl: getRichProductImageUrl({ brandName: brandTitle, categoryName: "Masa Aktif", items: products }),
  });
}

export default async function MasaAktifBrandPage({ params }: PageProps) {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const brands = await getBrandsByKategori("9");
  const { slug } = await params;
  const brand = findBrandByDedicatedSlug("9", brands, slug);
  if (!brand) {
    notFound();
  }
  const products = await getProductsByBrand("9", String(brand.id));
  const brandTitle = toTitleCase(brand.nama);
  const collectionJsonLd = buildCollectionJsonLd({
    title: `Perpanjang Masa Aktif ${brandTitle} | Isiloka`,
    description: `Perpanjang masa aktif ${brandTitle} di Isiloka dengan pilihan produk yang jelas.`,
    path: `/masa-aktif/${slug}`,
    itemNames: products.slice(0, 12).map((item) => item.nama),
  });
  const productJsonLd = buildProductItemListJsonLd({
    title: `Perpanjang Masa Aktif ${brandTitle} | Isiloka`,
    path: `/masa-aktif/${slug}`,
    brandName: brandTitle,
    categoryName: "Masa Aktif",
    items: products.slice(0, 24),
  });
  const faqJsonLd = buildFaqJsonLd([
    {
      question: `Apakah masa aktif ${brandTitle} di Isiloka tersedia dalam beberapa pilihan?`,
      answer: `Ya. Isiloka menampilkan produk masa aktif ${brandTitle} yang aktif agar pembeli bisa memilih sesuai kebutuhan.`,
    },
    {
      question: `Bagaimana cara memperpanjang masa aktif ${brandTitle} di Isiloka?`,
      answer: `Pilih produk masa aktif ${brandTitle}, masukkan nomor tujuan, lalu lanjutkan pembayaran sesuai metode yang tersedia.`,
    },
    {
      question: `Apakah produk masa aktif ${brandTitle} ini bisa dipakai pembeli umum dan member?`,
      answer: `Bisa. Halaman ini bisa dipakai pembeli umum maupun member Isiloka untuk kebutuhan pribadi dan penjualan ulang.`,
    },
  ]);
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "Masa Aktif", path: "/masa-aktif" },
    { name: brandTitle, path: `/masa-aktif/${slug}` },
  ]);
  const relatedLinks = [
    { label: `Pulsa ${brandTitle}`, href: `/pulsa/${slug}` },
    { label: `Paket Data ${brandTitle}`, href: `/paket-data/${slug}` },
    { label: `Paket Telepon ${brandTitle}`, href: `/paket-telepon/${slug}` },
    { label: "Top Up Dana", href: "/ewallet/dana" },
  ];

  return (
    <main className="bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(productJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(faqJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4 pt-4">
        <GuestPaketDataQuickOrder
          kategoriId="9"
          brands={brands}
          authToken={session?.backendToken}
          buyerRole={session?.user?.role}
          forcedBrand={brand}
          title="Masa Aktif"
          productLabel="masa aktif"
        />
        <section className="rounded-md border border-slate-200 bg-white p-4 shadow-[0_16px_36px_rgba(15,23,42,0.06)]">
          <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-sky-700">Layanan Terkait</p>
          <h2 className="mt-2 text-lg font-black tracking-tight text-slate-950">Layanan operator yang sering dicari bersama masa aktif</h2>
          <div className="mt-4 grid gap-3">
            {relatedLinks.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className="rounded-2xl border border-slate-200 bg-slate-50/80 px-4 py-3 text-sm font-semibold text-slate-800 transition hover:border-sky-200 hover:text-sky-700"
              >
                {item.label}
              </Link>
            ))}
          </div>
        </section>
      </div>

      <GuestBottomNav isLoggedIn={!!session?.backendToken} />
    </main>
  );
}
