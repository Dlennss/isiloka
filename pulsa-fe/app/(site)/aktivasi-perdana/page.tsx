import type { Metadata } from "next";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/nextauth";
import { getBrandsByKategori } from "@/lib/api.products";
import type { UserSession } from "@/components/user/types";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { GuestPaketDataQuickOrder } from "@/components/guest/GuestPaketDataQuickOrder";
import { buildBreadcrumbJsonLd, buildCollectionJsonLd, buildPageMetadata } from "@/lib/site-search";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

export const metadata: Metadata = buildPageMetadata({
  title: "Aktivasi Perdana | Isiloka",
  description: "Pilih produk aktivasi perdana semua operator di Isiloka dengan pilihan yang ringkas dan jelas.",
  path: "/aktivasi-perdana",
  keywords: ["aktivasi perdana", "kartu perdana", "aktivasi kartu operator", "pulsakilat"],
});

export default async function AktivasiPerdanaPage() {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const brands = await getBrandsByKategori("8");
  const collectionJsonLd = buildCollectionJsonLd({
    title: "Aktivasi Perdana | Isiloka",
    description: "Pilih produk aktivasi perdana semua operator di Isiloka dengan pilihan yang ringkas dan jelas.",
    path: "/aktivasi-perdana",
    itemNames: brands.map((brand) => brand.nama),
  });
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "Aktivasi Perdana", path: "/aktivasi-perdana" },
  ]);

  return (
    <main className="bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4 pt-4">
        {brands.length === 0 ? (
          <section className="grid min-h-40 place-items-center rounded-md border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_10px_28px_rgba(15,23,42,0.08)]">
            Brand aktivasi perdana belum tersedia.
          </section>
        ) : (
          <GuestPaketDataQuickOrder
            kategoriId="8"
            brands={brands}
            authToken={session?.backendToken}
            buyerRole={session?.user?.role}
            title="Aktivasi Perdana"
            productLabel="aktivasi perdana"
          />
        )}
      </div>

      <GuestBottomNav isLoggedIn={!!session?.backendToken} />
    </main>
  );
}
