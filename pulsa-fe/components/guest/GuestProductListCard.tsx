"use client";

import { ChevronRight } from "lucide-react";
import type { GuestProductCardCommonProps } from "@/components/guest/product-card-shared";
import { formatRupiah, getDisplayProductName, getProductPricing } from "@/components/guest/product-card-shared";

export function GuestProductListCard({
  item,
  isLoggedIn,
  onBuy,
  canBuy,
  buyBlockedLabel,
  hidePrice,
}: GuestProductCardCommonProps) {
  const { isFixed, feeActive, fixedPrice, openAmountPrice } = getProductPricing(item, isLoggedIn);
  const displayName = getDisplayProductName(item);

  return (
    <button
      type="button"
      onClick={() => {
        if (!canBuy) return;
        onBuy(item);
      }}
      disabled={!canBuy}
      className="group flex w-full items-center gap-3 border border-slate-100 bg-white px-4 py-3 text-left shadow-[0_8px_22px_rgba(15,23,42,0.10)] transition-all duration-300 hover:border-sky-200/60 hover:shadow-[0_12px_28px_rgba(15,23,42,0.16)] disabled:cursor-not-allowed disabled:opacity-60"
    >
      <div className="min-w-0 flex-1">
        <h2 className="line-clamp-2 text-sm font-bold text-slate-900">{displayName}</h2>
        {hidePrice ? (
          <p className="mt-1 text-[11px] text-slate-500">Biaya admin ditambahkan setelah hasil cek tagihan diterima.</p>
        ) : (
          <p className="mt-1 text-sm font-semibold text-slate-700">
            {isFixed && fixedPrice !== null ? formatRupiah(fixedPrice) : `+ ${formatRupiah(openAmountPrice ?? feeActive)}`}
          </p>
        )}
      </div>
      {canBuy ? (
        <ChevronRight className="h-5 w-5 shrink-0 text-sky-600" />
      ) : (
        <span className="text-xs font-semibold text-slate-400">{buyBlockedLabel || "Lengkapi dulu"}</span>
      )}
    </button>
  );
}
