"use client";

import { useMemo, useState } from "react";
import { ChevronLeft, ChevronRight, Search, ShoppingCart } from "lucide-react";
import type { UserProductItem } from "@/components/user/types";
import { UserCheckoutModal } from "@/components/user/UserCheckoutModal";
import { getRetailFeeForProduct } from "@/lib/retailRoles";
import { EWalletProductCard } from "@/components/shared/EWalletProductCard";

const PAGE_SIZE = 60;
const rupiah = (value: number) => `Rp ${new Intl.NumberFormat("id-ID").format(value)}`;

export function CatalogProductList({ items, authToken, buyerRole }: {
  items: UserProductItem[];
  authToken?: string;
  buyerRole?: string;
}) {
  const [query, setQuery] = useState("");
  const [page, setPage] = useState(0);
  const [selected, setSelected] = useState<UserProductItem | null>(null);
  const filtered = useMemo(() => {
    const term = query.trim().toLowerCase();
    return items.filter((item) => item.aktif && `${item.nama} ${item.sku}`.toLowerCase().includes(term));
  }, [items, query]);
  const pages = Math.max(1, Math.ceil(filtered.length / PAGE_SIZE));
  const currentPage = Math.min(page, pages - 1);
  const walletCards = items[0]?.kategori_nama.trim().toUpperCase() === "E-WALLET";
  const brand = selected?.brand_nama.toUpperCase() || "";
  const category = selected?.kategori_nama.toUpperCase() || "";
  const destMode = category === "GAME" && brand.startsWith("MOBILE LEGEND") ? "ml_id_server"
    : ["BANK TRANSFER", "TRANSFER BANK", "GAME", "VOUCHER DIGITAL"].includes(category) ? "alphanumeric" : "single";
  return (
    <section className="space-y-4 pb-28">
      <h1 className="break-words text-xl font-bold text-slate-900">{items[0]?.brand_nama || "Produk"}</h1>
      <label className="flex h-12 items-center gap-3 rounded-md border border-slate-200 bg-white px-3">
        <Search className="h-5 w-5 shrink-0 text-slate-500" aria-hidden="true" />
        <input value={query} onChange={(event) => { setQuery(event.target.value); setPage(0); }} aria-label="Cari nama atau SKU" placeholder="Cari nama atau SKU" className="min-w-0 flex-1 bg-transparent text-sm outline-none" />
      </label>
      <p role="status" className="text-xs text-slate-600">{filtered.length.toLocaleString("id-ID")} produk</p>
      <div className={walletCards ? "grid grid-cols-2 items-stretch gap-3" : "divide-y divide-slate-200"}>
        {filtered.slice(currentPage * PAGE_SIZE, (currentPage + 1) * PAGE_SIZE).map((item) => {
          const price = Number(item.harga_dasar_app) + getRetailFeeForProduct(item, authToken ? buyerRole : "guest");
          const validPrice = Number.isFinite(price) && price >= 0;
          const open = item.tipe_harga === "OPEN_AMOUNT";
          if (walletCards) {
            return <EWalletProductCard key={item.id} item={item} price={price} onBuy={setSelected} />;
          }
          return (
            <article key={item.id} className="flex items-center gap-3 py-4">
              <div className="min-w-0 flex-1 space-y-1">
                <p className="break-words font-mono text-[11px] text-slate-500">{item.sku}</p>
                <h2 className="break-words text-sm font-semibold leading-5 text-slate-900">{item.nama}</h2>
                <p className="text-sm font-bold text-teal-700">{validPrice ? `${open ? "Biaya tambahan " : ""}${rupiah(price)}` : "Harga belum tersedia"}</p>
                {open && item.maksimal_nominal ? <p className="text-xs text-slate-600">Maks. {rupiah(item.maksimal_nominal)}</p> : null}
              </div>
              <button type="button" onClick={() => setSelected(item)} disabled={!validPrice} aria-label={`Beli ${item.sku}`} title={`Beli ${item.sku}`} className="grid h-11 w-11 shrink-0 place-items-center rounded-md bg-teal-700 text-white hover:bg-teal-800 disabled:opacity-40">
                <ShoppingCart className="h-5 w-5" />
              </button>
            </article>
          );
        })}
      </div>
      {filtered.length === 0 ? <p className="py-8 text-center text-sm text-slate-600">Produk tidak ditemukan.</p> : null}
      {pages > 1 ? <nav aria-label="Halaman produk" className="flex items-center justify-between gap-3">
        <button type="button" aria-label="Halaman sebelumnya" title="Halaman sebelumnya" disabled={currentPage === 0} onClick={() => setPage(currentPage - 1)} className="grid h-11 w-11 place-items-center rounded-md border border-slate-200 bg-white disabled:opacity-40"><ChevronLeft className="h-5 w-5" /></button>
        <span className="text-sm text-slate-600">{currentPage + 1} / {pages}</span>
        <button type="button" aria-label="Halaman berikutnya" title="Halaman berikutnya" disabled={currentPage + 1 >= pages} onClick={() => setPage(currentPage + 1)} className="grid h-11 w-11 place-items-center rounded-md border border-slate-200 bg-white disabled:opacity-40"><ChevronRight className="h-5 w-5" /></button>
      </nav> : null}
      <UserCheckoutModal open={!!selected} product={selected} onClose={() => setSelected(null)} authToken={authToken} buyerRole={buyerRole} destMode={destMode} />
    </section>
  );
}
