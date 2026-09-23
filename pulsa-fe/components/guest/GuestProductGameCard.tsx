"use client";

import { Gamepad2 } from "lucide-react";
import type { GuestProductCardCommonProps } from "@/components/guest/product-card-shared";
import {
  formatRupiah,
  getDisplayProductName,
  getProductPricing,
} from "@/components/guest/product-card-shared";

function getMainLabel(name: string) {
  const compact = name.replace(/\s+/g, " ").trim();
  const number = compact.match(/\b(\d+(?:\.\d{3})*)\b/);
  if (number) return number[1];
  if (/weekly|mingguan/i.test(compact)) return "Weekly";
  if (/monthly|bulanan/i.test(compact)) return "Monthly";
  if (/pass/i.test(compact)) return "Pass";
  return compact;
}

export function GuestProductGameCard({
  item,
  isLoggedIn,
  onBuy,
  canBuy,
  buyBlockedLabel,
  hidePrice,
}: GuestProductCardCommonProps) {
  const { fixedPrice, openAmountPrice, feeActive, isFixed } = getProductPricing(item, isLoggedIn);
  const displayName = getDisplayProductName(item);
  const mainLabel = getMainLabel(displayName);
  const priceLabel = isFixed && fixedPrice !== null
    ? formatRupiah(fixedPrice)
    : `+${formatRupiah(openAmountPrice ?? feeActive)}`;

  return (
    <button
      type="button"
      onClick={() => {
        if (!canBuy) return;
        onBuy(item);
      }}
      disabled={!canBuy}
      className="group relative min-h-[116px] w-full overflow-hidden rounded-[22px] border border-emerald-100/70 bg-linear-to-br from-[#13464b] via-[#087e8b] to-[#22c55e] p-3 text-left text-white shadow-[0_16px_34px_rgba(6,78,59,0.18)] transition duration-300 hover:-translate-y-0.5 hover:shadow-[0_20px_42px_rgba(6,78,59,0.24)] disabled:cursor-not-allowed disabled:opacity-60"
    >
      <div className="absolute -right-10 -top-10 h-28 w-28 rounded-full bg-lime-300/35 blur-2xl transition duration-300 group-hover:bg-lime-200/45" />
      <div className="absolute -bottom-14 left-3 h-24 w-32 rounded-full bg-white/12 blur-2xl" />
      <div className="absolute inset-0 opacity-20 bg-[repeating-radial-gradient(circle_at_0_100%,rgba(255,255,255,0.55)_0,rgba(255,255,255,0.55)_1px,transparent_1px,transparent_11px)] bg-size-[150%_120%]" />
      <div className="absolute right-3 top-3 grid h-8 w-8 place-items-center rounded-2xl bg-white/14 text-lime-100 ring-1 ring-white/20">
        <Gamepad2 className="h-4 w-4" />
      </div>

      <div className="relative flex min-h-[92px] flex-col justify-between gap-3">
        <div className="pr-9">
          <h2 className="line-clamp-2 text-base font-black leading-tight tracking-tight text-white">
            {mainLabel}
          </h2>
          <p className="mt-1 line-clamp-1 text-[10px] font-semibold text-white/68">{displayName}</p>
        </div>

        <div className="flex items-end justify-end">
          <div className="text-right">
            <p className="text-[9px] font-semibold uppercase tracking-wide text-white/55">Harga</p>
            <p className="text-sm font-black leading-tight text-white">
              {hidePrice ? "Cek dulu" : canBuy ? priceLabel : (buyBlockedLabel || "Lengkapi dulu")}
            </p>
          </div>
        </div>
      </div>
    </button>
  );
}
