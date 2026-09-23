"use client";

import type { GuestProductCardCommonProps } from "@/components/guest/product-card-shared";
import {
  extractLargeNominalLabel,
  formatRupiah,
  getDisplayProductName,
  getProductPricing,
  isEMoneyFixedItem,
  isPackageStyleItem,
} from "@/components/guest/product-card-shared";

export function GuestProductDefaultCard({
  item,
  isLoggedIn,
  onBuy,
  canBuy,
  buyBlockedLabel,
  hidePrice,
}: GuestProductCardCommonProps) {
  const { fixedPrice, openAmountPrice, feeActive, isFixed } = getProductPricing(item, isLoggedIn);
  const packageStyle = isPackageStyleItem(item);
  const emoneyStyle = isEMoneyFixedItem(item);
  const plnStyle = String(item.kategori_nama || "").toUpperCase().includes("PLN")
    || String(item.brand_nama || "").toUpperCase() === "PLN"
    || String(item.sku || "").toUpperCase().includes("PLN");
  const displayName = getDisplayProductName(item);
  const priceLabel = isFixed && fixedPrice !== null
    ? formatRupiah(fixedPrice).replace("Rp ", "Rp")
    : `+${formatRupiah((openAmountPrice ?? feeActive)).replace("Rp ", "Rp")}`;

  if (plnStyle) {
    return (
      <button
        type="button"
        onClick={() => {
          if (!canBuy) return;
          onBuy(item);
        }}
        disabled={!canBuy}
        className="group relative w-full overflow-hidden rounded-[24px] bg-[#13464b] px-4 py-4 text-left text-white shadow-[0_16px_32px_rgba(5,46,38,0.22)] ring-1 ring-lime-200/20 transition duration-300 hover:-translate-y-0.5 hover:shadow-[0_20px_40px_rgba(5,46,38,0.28)] disabled:cursor-not-allowed disabled:opacity-60"
      >
        <div className="absolute -right-8 -top-10 h-26 w-26 rounded-full bg-lime-300/35 blur-2xl transition group-hover:bg-lime-200/45" />
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_15%_10%,rgba(255,255,255,0.16),transparent_28%),linear-gradient(135deg,rgba(16,185,129,0.22),rgba(163,230,53,0.06))]" />
        <div className="absolute inset-0 opacity-[0.16] bg-[repeating-radial-gradient(circle_at_0_100%,rgba(255,255,255,0.75)_0,rgba(255,255,255,0.75)_1px,transparent_1px,transparent_11px)] bg-size-[150%_120%]" />

        <div className="relative flex min-h-20 flex-col justify-between gap-4">
          <div className="min-w-0">
            <p className="text-2xl font-black leading-none tracking-tight text-lime-200">
              {extractLargeNominalLabel(item)}
            </p>
            <p className="mt-1 text-[10px] font-black uppercase tracking-[0.14em] text-white/55">Token PLN</p>
          </div>

          {hidePrice ? (
            <p className="text-right text-[10px] font-medium text-white/75">Biaya admin ditambahkan setelah hasil cek tagihan diterima.</p>
          ) : (
            <div className="text-right text-[12px] font-black leading-none tracking-tight text-white">
              {canBuy ? priceLabel : (buyBlockedLabel || "Lengkapi dulu")}
            </div>
          )}
        </div>
      </button>
    );
  }

  return (
    <button
      type="button"
      onClick={() => {
        if (!canBuy) return;
        onBuy(item);
      }}
      disabled={!canBuy}
      className="relative w-full overflow-hidden rounded-md bg-[#1491db] px-4 py-4 text-left text-white shadow-[0_14px_30px_rgba(0,132,209,0.22)] transition-transform duration-200 hover:-translate-y-0.5 disabled:cursor-not-allowed disabled:opacity-60"
    >
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(255,255,255,0.22),transparent_30%),linear-gradient(180deg,rgba(255,255,255,0.06),rgba(255,255,255,0.02))]" />
      <div className="absolute inset-0 opacity-30 bg-[repeating-radial-gradient(circle_at_0_100%,rgba(255,255,255,0.35)_0,rgba(255,255,255,0.35)_2px,transparent_2px,transparent_12px)] bg-size-[170%_130%]" />

      <div
        className={`relative flex ${
          emoneyStyle ? "min-h-18 flex-col justify-between" : "min-h-18 flex-col justify-between gap-4"
        }`}
      >
        <div className={`min-w-0 text-white ${emoneyStyle ? "flex flex-1 items-center justify-center text-center" : ""}`}>
          {emoneyStyle ? (
            <p className="text-xl font-bold tracking-tight text-white">{extractLargeNominalLabel(item)}</p>
          ) : (
            <h2 className={packageStyle ? "line-clamp-2 text-[13px] font-bold leading-tight text-white" : "line-clamp-3 text-[13px] font-bold leading-tight text-white"}>
              {displayName}
            </h2>
          )}
        </div>

        {hidePrice ? (
          <p className="text-right text-[10px] font-medium text-white/75">Biaya admin ditambahkan setelah hasil cek tagihan diterima.</p>
        ) : (
          <div className="text-right text-[11px] font-semibold leading-none tracking-tight text-white/65">
            {canBuy ? priceLabel : (buyBlockedLabel || "Lengkapi dulu")}
          </div>
        )}
      </div>
    </button>
  );
}
