import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/nextauth";
import { getBrandsByKategori, getCategories, getProductsByBrand } from "@/lib/api.products";
import type { UserCategoryItem, UserSession } from "@/components/user/types";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { EMoneyBrandFlow } from "@/components/shared/EMoneyBrandFlow";
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
  searchParams?: Promise<{ dest?: string }>;
};

function normalizeBrandFamily(value: string) {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "");
}

function normalizeName(value: string) {
  return value.trim().toLowerCase();
}

async function getEwalletKategoriId() {
  const categories = (await getCategories()) as UserCategoryItem[];
  const category = categories.find((item) => item.aktif && normalizeName(item.nama).includes("e-wallet"))
    ?? categories.find((item) => item.aktif && normalizeName(item.nama).includes("e-money"));
  return String(category?.id ?? 3);
}

function getLowestPrice(products: Awaited<ReturnType<typeof getProductsByBrand>>) {
  const values = products
    .map((item) => Number(item.harga_guest_final ?? item.harga_dasar_app ?? 0))
    .filter((value) => Number.isFinite(value) && value > 0);

  return values.length > 0 ? Math.min(...values) : null;
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const kategoriId = await getEwalletKategoriId();
  const brands = await getBrandsByKategori(kategoriId);
  const { slug } = await params;
  const brand = findBrandByDedicatedSlug(kategoriId, brands, slug);
  if (!brand) {
    return buildPageMetadata({
      title: "Top Up E-Wallet | Isiloka",
      description: "Top up e-wallet favorit Anda di Isiloka.",
      path: "/ewallet",
    });
  }

  const familyKey = normalizeBrandFamily(brand.nama);
  const familyBrands = brands.filter((item) => normalizeBrandFamily(item.nama) === familyKey);
  const productGroups = await Promise.all(
    familyBrands.map((item) => getProductsByBrand(kategoriId, String(item.id))),
  );
  const products = Array.from(
    new Map(productGroups.flat().map((item) => [item.id, item])).values(),
  );
  const lowestPrice = getLowestPrice(products);
  const brandTitle = toTitleCase(brand.nama);

  return buildPageMetadata({
    title: `Top Up ${brandTitle} Online | Isiloka`,
    description: lowestPrice
      ? `Top up ${brandTitle} online di Isiloka dengan nominal tetap dan bebas nominal. Harga mulai Rp ${lowestPrice.toLocaleString("id-ID")}.`
      : `Top up ${brandTitle} di Isiloka dengan pilihan nominal tetap dan bebas nominal yang lebih jelas.`,
    path: `/ewallet/${slug}`,
    keywords: [`top up ${brand.nama.toLowerCase()}`, `${brand.nama.toLowerCase()} pulsakilat`, "ewallet"],
    imageUrl: getRichProductImageUrl({ brandName: brandTitle, categoryName: "E-Wallet", items: products }),
  });
}

function normalizeDest(value: string | undefined) {
  return String(value || "").replace(/\D+/g, "").slice(0, 16);
}

export default async function EwalletBrandPage({ params, searchParams }: PageProps) {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const isLoggedIn = Boolean(session?.backendToken);
  const backendToken = session?.backendToken;
  const buyerRole = String(session?.user?.role || "").trim().toLowerCase();
  const kategoriId = await getEwalletKategoriId();
  const brands = await getBrandsByKategori(kategoriId);
  const { slug } = await params;
  const query = searchParams ? await searchParams : {};
  const initialDest = normalizeDest(query.dest);
  const brand = findBrandByDedicatedSlug(kategoriId, brands, slug);
  if (!brand) notFound();
  const familyKey = normalizeBrandFamily(brand.nama);
  const familyBrands = brands.filter((item) => normalizeBrandFamily(item.nama) === familyKey);
  const productGroups = await Promise.all(
    familyBrands.map((item) => getProductsByBrand(kategoriId, String(item.id))),
  );
  const products = Array.from(
    new Map(productGroups.flat().map((item) => [item.id, item])).values(),
  );
  const brandTitle = toTitleCase(brand.nama);

  const collectionJsonLd = buildCollectionJsonLd({
    title: `Top Up ${brandTitle} Online | Isiloka`,
    description: `Top up ${brandTitle} di Isiloka dengan pilihan nominal tetap dan bebas nominal yang lebih jelas.`,
    path: `/ewallet/${slug}`,
    itemNames: [brandTitle, ...products.slice(0, 5).map((item) => item.nama)],
  });
  const productJsonLd = buildProductItemListJsonLd({
    title: `Top Up ${brandTitle} Online | Isiloka`,
    path: `/ewallet/${slug}`,
    brandName: brandTitle,
    categoryName: "E-Wallet",
    items: products.slice(0, 24),
  });
  const faqJsonLd = buildFaqJsonLd([
    {
      question: `Apakah top up ${brandTitle} di Isiloka bisa pilih nominal tetap dan bebas nominal?`,
      answer: `Ya. Isiloka menampilkan produk ${brandTitle} yang aktif, termasuk nominal tetap dan varian bebas nominal bila tersedia.`,
    },
    {
      question: `Bagaimana cara top up ${brandTitle} di Isiloka?`,
      answer: `Pilih produk ${brandTitle}, masukkan nomor akun atau nomor tujuan sesuai kebutuhan, lalu selesaikan pembayaran.`,
    },
    {
      question: `Apakah halaman ${brandTitle} ini cocok untuk pembeli umum dan member?`,
      answer: `Bisa. Halaman ini dipakai untuk kebutuhan top up pribadi maupun penjualan ulang oleh member Isiloka.`,
    },
  ]);
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "E-Wallet", path: "/ewallet" },
    { name: brandTitle, path: `/ewallet/${slug}` },
  ]);
  const relatedLinks = [
    { label: "Pulsa Telkomsel", href: "/pulsa/telkomsel" },
    { label: "Paket Data Telkomsel", href: "/paket-data/telkomsel" },
    { label: "Top Up Mobile Legends", href: "/game/mobilelegends" },
    { label: "Top Up GoPay", href: "/ewallet/gopay" },
  ];

  return (
    <main className="min-h-screen bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(productJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(faqJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4">
        <section>
          {products.length === 0 ? (
            <div className="grid min-h-40 place-items-center rounded-3xl border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_8px_22px_rgba(15,23,42,0.13)]">
              Belum ada produk aktif untuk brand ini.
            </div>
          ) : (
            <EMoneyBrandFlow
              items={products}
              isLoggedIn={isLoggedIn}
              authToken={backendToken}
              mode={isLoggedIn ? "user" : "guest"}
              buyerRole={buyerRole}
              initialDest={initialDest}
            />
          )}
        </section>
        <section className="rounded-md border border-slate-200 bg-white p-4 shadow-[0_16px_36px_rgba(15,23,42,0.06)]">
          <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-sky-700">Layanan Terkait</p>
          <h2 className="mt-2 text-lg font-black tracking-tight text-slate-950">Kategori lain yang sering dipakai pelanggan digital</h2>
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

      <GuestBottomNav isLoggedIn={isLoggedIn} />
    </main>
  );
}
