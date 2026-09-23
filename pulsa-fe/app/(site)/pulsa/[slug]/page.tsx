import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/nextauth";
import { getBrandsByKategori, getCategories, getProductsByBrand } from "@/lib/api.products";
import type { UserCategoryItem, UserSession } from "@/components/user/types";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { GuestPulsaQuickOrder } from "@/components/guest/GuestPulsaQuickOrder";
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
  const pulsaCategory = pickCategory(categories, "pulsa");
  if (!pulsaCategory) {
    return buildPageMetadata({
      title: "Pulsa Operator | Isiloka",
      description: "Isi pulsa operator favorit Anda di Isiloka.",
      path: "/pulsa",
    });
  }

  const brands = await getBrandsByKategori(String(pulsaCategory.id));
  const { slug } = await params;
  const brand = findBrandByDedicatedSlug(String(pulsaCategory.id), brands, slug);
  if (!brand) {
    return buildPageMetadata({
      title: "Pulsa Operator | Isiloka",
      description: "Isi pulsa operator favorit Anda di Isiloka.",
      path: "/pulsa",
    });
  }

  const products = await getProductsByBrand(String(pulsaCategory.id), String(brand.id));
  const lowestPrice = getLowestPrice(products);
  const brandTitle = toTitleCase(brand.nama);

  return buildPageMetadata({
    title: `Isi Pulsa ${brandTitle} Online | Isiloka`,
    description: lowestPrice
      ? `Isi pulsa ${brandTitle} online di Isiloka dengan nominal lengkap. Harga mulai Rp ${lowestPrice.toLocaleString("id-ID")} dan transaksi cepat.`
      : `Isi pulsa ${brandTitle} online di Isiloka dengan nominal lengkap dan transaksi cepat.`,
    path: `/pulsa/${slug}`,
    keywords: [`pulsa ${brand.nama.toLowerCase()}`, `isi pulsa ${brand.nama.toLowerCase()}`, "pulsakilat"],
    imageUrl: getRichProductImageUrl({ brandName: brandTitle, categoryName: "Pulsa", items: products }),
  });
}

export default async function GuestPulsaBrandPage({ params }: PageProps) {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const categories = (await getCategories()) as UserCategoryItem[];
  const pulsaCategory = pickCategory(categories, "pulsa");
  if (!pulsaCategory) notFound();

  const brands = await getBrandsByKategori(String(pulsaCategory.id));
  const { slug } = await params;
  const brand = findBrandByDedicatedSlug(String(pulsaCategory.id), brands, slug);
  if (!brand) notFound();
  const products = await getProductsByBrand(String(pulsaCategory.id), String(brand.id));
  const brandTitle = toTitleCase(brand.nama);
  const collectionJsonLd = buildCollectionJsonLd({
    title: `Isi Pulsa ${brandTitle} Online | Isiloka`,
    description: `Isi pulsa ${brandTitle} online di Isiloka dengan nominal lengkap dan transaksi cepat.`,
    path: `/pulsa/${slug}`,
    itemNames: products.slice(0, 12).map((item) => item.nama),
  });
  const productJsonLd = buildProductItemListJsonLd({
    title: `Isi Pulsa ${brandTitle} Online | Isiloka`,
    path: `/pulsa/${slug}`,
    brandName: brandTitle,
    categoryName: "Pulsa",
    items: products.slice(0, 24),
  });
  const faqJsonLd = buildFaqJsonLd([
    {
      question: `Apakah nominal pulsa ${brandTitle} di Isiloka lengkap?`,
      answer: `Ya. Isiloka menampilkan pilihan nominal pulsa ${brandTitle} yang aktif agar pembeli bisa memilih sesuai kebutuhan.`,
    },
    {
      question: `Bagaimana cara beli pulsa ${brandTitle} di Isiloka?`,
      answer: `Pilih nominal pulsa ${brandTitle}, masukkan nomor tujuan, lalu lanjutkan pembayaran sesuai metode yang tersedia.`,
    },
    {
      question: `Apakah pembelian pulsa ${brandTitle} di Isiloka bisa untuk calon member dan member?`,
      answer: `Bisa. Halaman ini bisa dipakai pembeli umum, dan member Isiloka juga bisa memanfaatkan pilihan produk yang sama untuk transaksi harian.`,
    },
  ]);
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "Pulsa", path: "/pulsa" },
    { name: brandTitle, path: `/pulsa/${slug}` },
  ]);
  const relatedLinks = [
    { label: `Paket Data ${brandTitle}`, href: `/paket-data/${slug}` },
    { label: `Paket Telepon ${brandTitle}`, href: `/paket-telepon/${slug}` },
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
        <GuestPulsaQuickOrder
          kategoriId={String(pulsaCategory.id)}
          brands={brands}
          authToken={session?.backendToken}
          buyerRole={session?.user?.role}
          forcedBrand={brand}
        />
        <section className="rounded-md border border-slate-200 bg-white p-4 shadow-[0_16px_36px_rgba(15,23,42,0.06)]">
          <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-sky-700">Layanan Terkait</p>
          <h2 className="mt-2 text-lg font-black tracking-tight text-slate-950">Kebutuhan lain yang sering dibeli bersama</h2>
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
