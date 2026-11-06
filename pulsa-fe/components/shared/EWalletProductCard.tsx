"use client";

import type { UserProductItem } from "@/components/user/types";

export function EWalletProductCard({ item, price, onBuy, canBuy = true, buyLabel = "Top Up", blockedLabel = "Lengkapi dulu" }: {
  item: UserProductItem;
  price: number;
  onBuy: (item: UserProductItem) => void;
  canBuy?: boolean;
  buyLabel?: string;
  blockedLabel?: string;
}) {
  const validPrice = Number.isFinite(price) && price >= 0;
  return (
    <article className="relative flex min-w-0 flex-col overflow-hidden rounded-[22px] border border-slate-200 bg-linear-to-bl from-lime-50 via-white to-white p-3 shadow-[0_8px_20px_rgba(22,56,120,0.08)]">
      <div aria-hidden="true" className="absolute inset-x-0 top-0 h-1.5 bg-linear-to-r from-[#163878] via-sky-500 to-lime-300" />
      <p className="mt-1 text-[10px] font-black uppercase leading-4 text-blue-600">Nominal Top Up</p>
      <h2 title={`${item.nama} (${item.sku})`} className="mt-2 line-clamp-2 h-12 break-words text-[20px] font-extrabold leading-6 text-slate-950">{item.nama}</h2>
      <div className="mt-3 flex-1">
        <p className="truncate text-xs leading-4 text-slate-600" title={item.brand_nama}>Top up {item.brand_nama}</p>
        <p className="mt-1 min-h-10 break-words text-sm font-bold leading-5 text-[#163878]">
          {validPrice ? <>{item.tipe_harga === "OPEN_AMOUNT" ? <span className="block text-[11px] font-medium leading-4 text-slate-600">Biaya tambahan </span> : null}Rp {new Intl.NumberFormat("id-ID").format(price)}</> : "Harga belum tersedia"}
        </p>
      </div>
      <button type="button" onClick={() => onBuy(item)} disabled={!validPrice || !canBuy} aria-label={`${buyLabel} ${item.sku}`} title={`${buyLabel} ${item.nama} (${item.sku})`} className="mt-2 flex h-10 w-full shrink-0 items-center justify-center rounded-full bg-[#163878] px-2 text-xs font-extrabold text-white transition-colors hover:bg-[#214c98] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:cursor-not-allowed disabled:opacity-40">
        {canBuy ? buyLabel : blockedLabel}
      </button>
    </article>
  );
}
