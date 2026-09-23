import type { Metadata } from "next";
import { notFound } from "next/navigation";
import Image from "next/image";
import { Gamepad2, ShieldCheck, Zap } from "lucide-react";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/nextauth";
import { getBrandsByKategori, getProductsByBrand } from "@/lib/api.products";
import type { UserProductItem, UserSession } from "@/components/user/types";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { GuestProductGrid } from "@/components/guest/GuestProductGrid";
import { findGameBrandBySlug } from "@/lib/game-brand-routes";
import { getBrandLogo } from "@/lib/brand-logos";
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

function getLowestPrice(products: UserProductItem[]) {
  const values = products
    .map((item) => Number(item.harga_guest_final ?? item.harga_dasar_app ?? 0))
    .filter((value) => Number.isFinite(value) && value > 0);

  return values.length > 0 ? Math.min(...values) : null;
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const brands = await getBrandsByKategori("5");
  const { slug } = await params;
  const brand = findGameBrandBySlug(brands, slug);
  if (!brand) {
    return buildPageMetadata({
      title: "Top Up Game Online | Isiloka",
      description: "Top up game online favorit Anda di Isiloka.",
      path: "/game",
    });
  }

  const products = prepareGameProducts(await getProductsByBrand("5", String(brand.id)));
  const lowestPrice = getLowestPrice(products);
  const brandTitle = toTitleCase(brand.nama);

  return buildPageMetadata({
    title: `Top Up ${brandTitle} Murah | Isiloka`,
    description: lowestPrice
      ? `Top up ${brandTitle} di Isiloka dengan pilihan nominal aktif. Harga mulai Rp ${lowestPrice.toLocaleString("id-ID")} dan transaksi cepat.`
      : `Top up ${brandTitle} di Isiloka dengan pilihan nominal yang jelas dan proses transaksi cepat.`,
    path: `/game/${slug}`,
    keywords: [`top up ${brand.nama.toLowerCase()}`, `${brand.nama.toLowerCase()} murah`, "voucher game", "pulsakilat"],
    imageUrl: getRichProductImageUrl({ brandName: brandTitle, categoryName: "Game", items: products }),
  });
}

function parseLeadingNumber(value: string) {
  const match = value.match(/(\d+(?:\.\d{3})*)/);
  if (!match) return Number.POSITIVE_INFINITY;
  return Number.parseInt(match[1].replace(/\./g, ""), 10);
}

function buildGameSortKey(item: UserProductItem) {
  const name = String(item.nama || "").toUpperCase();
  const sku = String(item.sku || "").toUpperCase();

  if (name.includes("CEK NICK")) return [0, 0, sku] as const;
  if (name.includes("MEMBER MINGGUAN")) return [1, 0, sku] as const;
  if (name.includes("MEMBER BULANAN")) return [1, 1, sku] as const;
  if (name.includes("LEVEL UP PASS")) return [2, parseLeadingNumber(name), sku] as const;
  return [3, parseLeadingNumber(name), sku] as const;
}

function prepareGameProducts(items: UserProductItem[]) {
  const counts = new Map<string, number>();
  for (const item of items) {
    const key = String(item.nama || "").trim();
    counts.set(key, (counts.get(key) || 0) + 1);
  }

  return items
    .map((item) => {
      const rawName = String(item.nama || "").trim();
      const sku = String(item.sku || "").trim().toUpperCase();
      const duplicateCount = counts.get(rawName) || 0;
      const duplicateLabel = sku.startsWith("VG") ? "VIP" : sku.startsWith("G") ? "REG" : sku;
      return duplicateCount > 1
        ? { ...item, nama: `${rawName} ${duplicateLabel}` }
        : item;
    })
    .sort((left, right) => {
      const a = buildGameSortKey(left);
      const b = buildGameSortKey(right);
      return a[0] - b[0] || a[1] - b[1] || a[2].localeCompare(b[2]);
    });
}

export default async function GameBrandPage({ params, searchParams }: PageProps) {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const isLoggedIn = Boolean(session?.backendToken);
  const backendToken = session?.backendToken;
  const buyerRole = String(session?.user?.role || "").trim().toLowerCase();
  const { slug } = await params;
  const initialDest = String((await searchParams)?.dest || "").replace(/[^a-zA-Z0-9]/g, "").slice(0, 32);

  const brands = await getBrandsByKategori("5");
  const brand = findGameBrandBySlug(brands, slug);
  if (!brand) {
    notFound();
  }

  const products = prepareGameProducts(await getProductsByBrand("5", String(brand.id)));
  const brandTitle = toTitleCase(brand.nama);
  const brandLogo = getBrandLogo(brand.nama);
  const collectionJsonLd = buildCollectionJsonLd({
    title: `Top Up ${brandTitle} Murah | Isiloka`,
    description: `Top up ${brandTitle} di Isiloka dengan pilihan nominal yang jelas dan proses transaksi cepat.`,
    path: `/game/${slug}`,
    itemNames: products.slice(0, 12).map((item) => String(item.nama || "").trim()).filter(Boolean),
  });
  const productJsonLd = buildProductItemListJsonLd({
    title: `Top Up ${brandTitle} Murah | Isiloka`,
    path: `/game/${slug}`,
    brandName: brandTitle,
    categoryName: "Game",
    items: products.slice(0, 24),
  });
  const faqJsonLd = buildFaqJsonLd([
    {
      question: `Apakah top up ${brandTitle} di Isiloka tersedia dalam banyak pilihan nominal?`,
      answer: `Ya. Isiloka menampilkan produk ${brandTitle} yang aktif agar pembeli bisa memilih nominal yang sesuai kebutuhan.`,
    },
    {
      question: `Bagaimana cara top up ${brandTitle} di Isiloka?`,
      answer: `Pilih nominal ${brandTitle}, isi data tujuan sesuai kebutuhan game, lalu lanjutkan pembayaran dengan metode yang tersedia.`,
    },
    {
      question: `Apakah halaman ${brandTitle} ini bisa dipakai pembeli umum dan member?`,
      answer: `Bisa. Halaman ini bisa dipakai pembeli umum maupun member Isiloka untuk transaksi pribadi dan penjualan ulang.`,
    },
  ]);
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "Game", path: "/game" },
    { name: brandTitle, path: `/game/${slug}` },
  ]);
  return (
    <main className="min-h-screen bg-[#eef7f3] pb-24">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(productJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(faqJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4 pt-4">
        <section className="overflow-hidden rounded-[26px] border border-emerald-950/5 bg-linear-to-br from-[#13464b] via-[#087e8b] to-[#22c55e] p-4 text-white shadow-[0_18px_42px_rgba(6,78,59,0.22)]">
          <div className="relative">
            <div className="absolute -right-12 -top-16 h-32 w-32 rounded-full bg-lime-300/30 blur-2xl" />
            <div className="relative flex items-center gap-3">
              <div className="grid h-16 w-24 shrink-0 place-items-center overflow-hidden rounded-2xl bg-white shadow-[0_14px_30px_rgba(5,46,38,0.20)]">
                {brandLogo?.src ? (
                  <Image
                    src={brandLogo.src}
                    alt={brandLogo.alt || brand.nama}
                    width={160}
                    height={96}
                    className="h-full w-full object-contain p-2"
                  />
                ) : (
                  <Gamepad2 className="h-8 w-8 text-[#087e8b]" />
                )}
              </div>
              <div className="min-w-0 flex-1">
                <p className="inline-flex items-center gap-1 rounded-full bg-white/12 px-2 py-1 text-[10px] font-black uppercase tracking-wide text-lime-100 ring-1 ring-white/15">
                  <Zap className="h-3 w-3" fill="currentColor" />
                  Top Up Game
                </p>
                <h1 className="mt-2 truncate text-xl font-black tracking-tight text-white">{brandTitle}</h1>
                <p className="mt-1 text-xs font-semibold leading-5 text-white/72">Pilih nominal, isi ID, lalu lanjut bayar.</p>
              </div>
            </div>
            <div className="relative mt-4 flex items-center gap-2 rounded-2xl bg-white/10 px-3 py-2 text-xs font-bold text-white/86 ring-1 ring-white/15">
              <ShieldCheck className="h-4 w-4 text-lime-200" />
              Produk aktif dan siap diproses otomatis.
            </div>
          </div>
        </section>
        <section>
          {products.length === 0 ? (
            <div className="grid min-h-40 place-items-center rounded-3xl border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_8px_22px_rgba(15,23,42,0.13)]">
              Belum ada produk aktif untuk brand ini.
            </div>
          ) : (
            <GuestProductGrid
              items={products}
              isLoggedIn={isLoggedIn}
              authToken={backendToken}
              buyerRole={buyerRole}
              initialDest={initialDest}
              destLabel={slug === "roblox" ? "Nomor HP" : "ID"}
              destPlaceholder={slug === "roblox" ? "Masukkan nomor HP" : "Masukkan ID"}
              destMode={slug === "mobilelegends" ? "ml_id_server" : slug === "pointblank" ? "alphanumeric" : "single"}
              showBrandBanner={false}
              visualStyle="game"
              enableGuestHint={false}
            />
          )}
        </section>
      </div>

      <GuestBottomNav isLoggedIn={isLoggedIn} />
    </main>
  );
}
