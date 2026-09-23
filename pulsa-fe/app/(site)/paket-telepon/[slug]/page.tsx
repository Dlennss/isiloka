import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/nextauth";
import { getBrandsByKategori, getCategories, getProductsByBrand } from "@/lib/api.products";
import type { UserCategoryItem, UserSession } from "@/components/user/types";
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

function pickCategory(categories: UserCategoryItem[], keyword: string) {
  return categories.find((item) => item.aktif && item.nama.toLowerCase().includes(keyword));
}

function getLowestPrice(products: Awaited<ReturnType<typeof getProductsByBrand>>) {
  const values = products
    .map((item) => Number(item.harga_guest_final ?? item.harga_dasar_app ?? 0))
    .filter((value) => Number.isFinite(value) && value > 0);

  return values.length > 0 ? Math.min(...values) : null;
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const categories = (await getCategories()) as UserCategoryItem[];
  const phoneCategory = pickCategory(categories, "telepon");
  if (!phoneCategory) {
    return buildPageMetadata({
      title: "Paket Telepon Semua Operator | Isiloka",
      description: "Pilih paket telepon operator di Isiloka.",
      path: "/paket-telepon",
    });
  }

  const brands = await getBrandsByKategori(String(phoneCategory.id));
  const { slug } = await params;
  const brand = findBrandByDedicatedSlug(String(phoneCategory.id), brands, slug);
  if (!brand) {
    return buildPageMetadata({
      title: "Paket Telepon Semua Operator | Isiloka",
      description: "Pilih paket telepon operator di Isiloka.",
      path: "/paket-telepon",
    });
  }

  const products = await getProductsByBrand(String(phoneCategory.id), String(brand.id));
  const lowestPrice = getLowestPrice(products);
  const brandTitle = toTitleCase(brand.nama);

  return buildPageMetadata({
    title: `Paket Telepon ${brandTitle} | Isiloka`,
    description: lowestPrice
      ? `Pilih paket telepon ${brandTitle} di Isiloka. Produk aktif tersedia mulai Rp ${lowestPrice.toLocaleString("id-ID")}.`
      : `Pilih paket telepon ${brandTitle} di Isiloka dengan pilihan produk yang jelas.`,
    path: `/paket-telepon/${slug}`,
    keywords: [`paket telepon ${brand.nama.toLowerCase()}`, `nelpon ${brand.nama.toLowerCase()}`, "pulsakilat"],
    imageUrl: getRichProductImageUrl({ brandName: brandTitle, categoryName: "Paket Telepon", items: products }),
  });
}

export default async function GuestPaketTeleponBrandPage({ params }: PageProps) {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const categories = (await getCategories()) as UserCategoryItem[];
  const phoneCategory = pickCategory(categories, "telepon");
  if (!phoneCategory) notFound();

  const brands = await getBrandsByKategori(String(phoneCategory.id));
  const { slug } = await params;
  const brand = findBrandByDedicatedSlug(String(phoneCategory.id), brands, slug);
  if (!brand) notFound();
  const products = await getProductsByBrand(String(phoneCategory.id), String(brand.id));
  const brandTitle = toTitleCase(brand.nama);
  const collectionJsonLd = buildCollectionJsonLd({
    title: `Paket Telepon ${brandTitle} | Isiloka`,
    description: `Pilih paket telepon ${brandTitle} di Isiloka dengan pilihan produk yang jelas.`,
    path: `/paket-telepon/${slug}`,
    itemNames: products.slice(0, 12).map((item) => item.nama),
  });
  const productJsonLd = buildProductItemListJsonLd({
    title: `Paket Telepon ${brandTitle} | Isiloka`,
    path: `/paket-telepon/${slug}`,
    brandName: brandTitle,
    categoryName: "Paket Telepon",
    items: products.slice(0, 24),
  });
  const faqJsonLd = buildFaqJsonLd([
    {
      question: `Apakah paket telepon ${brandTitle} di Isiloka tersedia dalam banyak pilihan?`,
      answer: `Ya. Isiloka menampilkan paket telepon ${brandTitle} yang aktif agar pembeli bisa memilih sesuai kebutuhan.`,
    },
    {
      question: `Bagaimana cara membeli paket telepon ${brandTitle} di Isiloka?`,
      answer: `Pilih produk paket telepon ${brandTitle}, masukkan nomor tujuan, lalu lanjutkan pembayaran.`,
    },
    {
      question: `Apakah halaman paket telepon ${brandTitle} ini cocok untuk pembeli umum dan member?`,
      answer: `Bisa. Halaman ini dapat dipakai pembeli umum maupun member Isiloka untuk transaksi pribadi dan penjualan ulang.`,
    },
  ]);
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "Paket Telepon", path: "/paket-telepon" },
    { name: brandTitle, path: `/paket-telepon/${slug}` },
  ]);
  const relatedLinks = [
    { label: `Pulsa ${brandTitle}`, href: `/pulsa/${slug}` },
    { label: `Paket Data ${brandTitle}`, href: `/paket-data/${slug}` },
    { label: `Masa Aktif ${brandTitle}`, href: `/masa-aktif/${slug}` },
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
          kategoriId={String(phoneCategory.id)}
          brands={brands}
          authToken={session?.backendToken}
          buyerRole={session?.user?.role}
          forcedBrand={brand}
          title="Paket Telepon"
          productLabel="paket telepon"
          showGroupTabs={false}
        />
        <section className="rounded-md border border-slate-200 bg-white p-4 shadow-[0_16px_36px_rgba(15,23,42,0.06)]">
          <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-sky-700">Layanan Terkait</p>
          <h2 className="mt-2 text-lg font-black tracking-tight text-slate-950">Kategori operator yang sering dibeli setelah paket telepon</h2>
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
