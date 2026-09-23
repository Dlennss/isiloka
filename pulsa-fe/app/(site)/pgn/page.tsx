import type { Metadata } from "next";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/nextauth";
import { getBrandsByKategori, getProductsByBrand } from "@/lib/api.products";
import type { UserBrandItem, UserSession } from "@/components/user/types";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { RetailBillingEntryFlow } from "@/components/shared/RetailBillingEntryFlow";
import { buildBreadcrumbJsonLd, buildCollectionJsonLd, buildPageMetadata } from "@/lib/site-search";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

function pickPGNBrand(brands: UserBrandItem[]) {
  return brands.find((item) => item.aktif && String(item.nama || "").trim().toLowerCase() === "pgn") ?? brands[0] ?? null;
}

export const metadata: Metadata = buildPageMetadata({
  title: "Tagihan PGN | Isiloka",
  description: "Cek dan bayar tagihan gas PGN di Isiloka dengan langkah pembayaran yang ringkas.",
  path: "/pgn",
  keywords: ["bayar pgn online", "tagihan gas pgn", "cek tagihan pgn", "pulsakilat"],
});

export default async function PGNPage() {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const brands = await getBrandsByKategori("20");
  const pgnBrand = pickPGNBrand(brands);
  const items = pgnBrand ? await getProductsByBrand("20", String(pgnBrand.id)) : [];
  const collectionJsonLd = buildCollectionJsonLd({
    title: "Tagihan PGN | Isiloka",
    description: "Cek dan bayar tagihan gas PGN di Isiloka dengan langkah pembayaran yang ringkas.",
    path: "/pgn",
    itemNames: items.slice(0, 8).map((item) => String(item.nama || "").trim()).filter(Boolean),
  });
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "PGN", path: "/pgn" },
  ]);

  return (
    <main className="bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4 pt-4">
        {!pgnBrand ? (
          <section className="grid min-h-40 place-items-center rounded-md border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_10px_28px_rgba(15,23,42,0.08)]">
            Brand gas negara belum tersedia.
          </section>
        ) : (
          <RetailBillingEntryFlow
            title="Tagihan PGN"
            description=""
            placeholder="Masukkan nomor pelanggan"
            items={items}
            mode="guest"
            logoSrc="/images/gas/Logo_PGN.png"
            logoAlt="PGN"
          />
        )}
      </div>

      <GuestBottomNav isLoggedIn={!!session?.backendToken} />
    </main>
  );
}
