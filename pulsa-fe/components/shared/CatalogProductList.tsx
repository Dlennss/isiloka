"use client";

import { useMemo, useState } from "react";
import { ChevronLeft, ChevronRight, Search, ShoppingCart } from "lucide-react";
import type { UserProductItem } from "@/components/user/types";
import { UserCheckoutModal } from "@/components/user/UserCheckoutModal";
import { getRetailFeeForProduct } from "@/lib/retailRoles";

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
            return (
              <article key={item.id} className="relative flex min-w-0 flex-col overflow-hidden rounded-[22px] border border-slate-200 bg-linear-to-bl from-lime-50 via-white to-white p-3 shadow-[0_8px_20px_rgba(22,56,120,0.08)]">
                <div aria-hidden="true" className="absolute inset-x-0 top-0 h-1.5 bg-linear-to-r from-[#163878] via-sky-500 to-lime-300" />
                <p className="mt-1 text-[10px] font-black uppercase leading-4 text-blue-600">Nominal Top Up</p>
                <h2 title={`${item.nama} (${item.sku})`} className="mt-2 line-clamp-2 h-12 break-words text-[20px] font-extrabold leading-6 text-slate-950">{item.nama}</h2>
                <div className="mt-3 flex-1">
                  <p className="truncate text-xs leading-4 text-slate-600" title={item.brand_nama}>Top up {item.brand_nama}</p>
                  <p className="mt-1 min-h-10 break-words text-sm font-bold leading-5 text-[#163878]">
                    {validPrice ? <>{open ? <span className="block text-[11px] font-medium leading-4 text-slate-600">Biaya tambahan </span> : null}{rupiah(price)}</> : "Harga belum tersedia"}
                  </p>
                </div>
                <button type="button" onClick={() => setSelected(item)} disabled={!validPrice} aria-label={`Top Up ${item.sku}`} title={`Top Up ${item.nama} (${item.sku})`} className="mt-2 flex h-10 w-full shrink-0 items-center justify-center rounded-full bg-[#163878] px-2 text-xs font-extrabold text-white transition-colors hover:bg-[#214c98] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:cursor-not-allowed disabled:opacity-40">
                  Top Up
                </button>
              </article>
            );
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
