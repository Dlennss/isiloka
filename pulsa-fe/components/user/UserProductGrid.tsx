"use client";

import * as React from "react";
import Link from "next/link";
import { X } from "lucide-react";
import type { UserProductItem } from "@/components/user/types";
import { UserCheckoutModal } from "@/components/user/UserCheckoutModal";
import { getDisplayProductName, getDisplayedFixedPrice } from "@/components/guest/product-card-shared";
import { getRetailFeeForProduct } from "@/lib/retailRoles";

type UserProductGridProps = {
  items: UserProductItem[];
  isLoggedIn: boolean;
  authToken?: string;
  buyerRole?: string;
  initialDest?: string;
  canBuy?: boolean;
  buyBlockedLabel?: string;
  enableGuestHint?: boolean;
  buyLabel?: string;
  hidePrice?: boolean;
};

function formatRupiah(value: number) {
  return `Rp ${new Intl.NumberFormat("id-ID").format(value || 0)}`;
}

function isPackageStyleItem(item: UserProductItem) {
  const upper = String(item.kategori_nama || item.nama || "").toUpperCase();
  return upper.includes("PAKET DATA") || upper.includes("PAKET TELEPON") || upper.includes("PAKET SMS");
}

function isEMoneyItem(item: UserProductItem) {
  const category = String(item.kategori_nama || "").toUpperCase();
  return category.includes("E-MONEY") || category.includes("E-WALLET");
}

function escapeRegExp(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function toTitleCase(value: string) {
  return value
    .toLowerCase()
    .split(" ")
    .filter(Boolean)
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

function getEMoneyCardTitle(item: UserProductItem) {
  if (item.tipe_harga === "OPEN_AMOUNT") {
    const raw = String(item.nama || "").replace(/\s+/g, " ").trim();
    const brand = String(item.brand_nama || "").trim();
    let cleaned = raw;

    if (brand) {
      const compactBrand = brand.replace(/\s+/g, "");
      const brandPatterns = [brand, compactBrand].filter(Boolean);
      for (const pattern of brandPatterns) {
        cleaned = cleaned.replace(new RegExp(`^${escapeRegExp(pattern)}\\s*`, "i"), "").trim();
      }
    }

    return toTitleCase(cleaned || "Bebas Nominal");
  }
  return extractLargeNominalLabel(item);
}

function extractLargeNominalLabel(item: UserProductItem) {
  const upper = String(item.nama || "").toUpperCase();
  const dotted = upper.match(/(\d+(?:\.\d{3})+)/);
  if (dotted) return dotted[1];

  const compact = upper.match(/(\d+)\s*K\b/);
  if (compact) return `${Number.parseInt(compact[1], 10)}.000`;

  return new Intl.NumberFormat("id-ID").format(Number(item.nominal || 0));
}

function UserProductCard({
  item,
  isLoggedIn,
  buyerRole,
  onBuy,
  canBuy,
  buyBlockedLabel,
  buyLabel,
  hidePrice,
}: {
  item: UserProductItem;
  isLoggedIn: boolean;
  buyerRole?: string;
  onBuy: (item: UserProductItem) => void;
  canBuy: boolean;
  buyBlockedLabel?: string;
  buyLabel?: string;
  hidePrice?: boolean;
}) {
  const isFixed = item.tipe_harga === "FIXED";
  const effectiveRole = isLoggedIn ? buyerRole : "guest";
  const feeActive = getRetailFeeForProduct(item, effectiveRole);
  const fixedPrice = isFixed ? getDisplayedFixedPrice(item, effectiveRole) : null;
  const finalPrice = Number(fixedPrice || feeActive || 0);
  const packageStyle = isPackageStyleItem(item);
  const emoneyStyle = isEMoneyItem(item);
  const displayName = getDisplayProductName(item);

  if (emoneyStyle) {
    const nominalLabel = getEMoneyCardTitle(item);
    const brandLabel = String(item.brand_nama || "E-Wallet").trim();
    return (
      <button
        type="button"
        onClick={() => {
          if (!canBuy) return;
          onBuy(item);
        }}
        disabled={!canBuy}
        className="group relative min-h-36 overflow-hidden rounded-[22px] border border-emerald-950/10 bg-white p-3 text-left shadow-[0_14px_30px_rgba(6,78,59,0.10)] transition-all duration-300 hover:-translate-y-0.5 hover:border-lime-300 hover:shadow-[0_18px_36px_rgba(6,78,59,0.16)] disabled:cursor-not-allowed disabled:opacity-60"
      >
        <div className="absolute inset-x-0 top-0 h-1.5 bg-linear-to-r from-[#13464b] via-[#087e8b] to-[#b8f138]" />
        <div className="absolute -right-8 -top-8 h-24 w-24 rounded-full bg-lime-200/55 blur-2xl" />
        <div className="relative flex h-full min-h-30 flex-col justify-between gap-3">
          <div className="flex items-start justify-between gap-2">
            <div className="min-w-0">
              <p className="text-[10px] font-black uppercase tracking-[0.16em] text-[#087e8b]">Nominal Top Up</p>
              <p className={item.tipe_harga === "OPEN_AMOUNT" ? "mt-2 line-clamp-2 text-[21px] font-black leading-tight text-slate-950" : "mt-2 text-[26px] font-black leading-none tracking-tight text-slate-950"}>
                {nominalLabel}
              </p>
            </div>
          </div>

          <div className="space-y-2">
            <div>
              <p className="line-clamp-1 text-[11px] font-bold text-slate-500">Top up {brandLabel}</p>
              <p className="text-sm font-black text-[#13464b]">
                {canBuy ? formatRupiah(finalPrice) : (buyBlockedLabel || "Lengkapi dulu")}
              </p>
            </div>
            <span className="inline-flex h-8 w-full items-center justify-center rounded-2xl bg-[#13464b] px-3 text-xs font-black text-white shadow-[0_10px_20px_rgba(5,46,38,0.18)] transition group-hover:bg-[#087e8b]">
              {canBuy ? (buyLabel || "Top Up") : (buyBlockedLabel || "Lengkapi dulu")}
            </span>
          </div>
        </div>
      </button>
    );
  }

  return (
    <article className="group relative rounded-xl bg-white p-3 shadow-[0_8px_22px_rgba(15,23,42,0.13)] transition-all duration-300 hover:shadow-[0_16px_40px_rgba(15,23,42,0.2)] hover:-translate-y-1 border border-slate-100/50 hover:border-sky-200/50">
      <div className="absolute inset-0 rounded-xl bg-linear-to-br from-sky-50/30 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
      <div className="relative">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0 flex-1">
            <h2 className={`text-slate-900 transition-colors group-hover:text-sky-900 ${packageStyle ? "line-clamp-3 text-[13px] font-bold leading-tight" : "line-clamp-2 text-sm font-bold"}`}>
              {displayName}
            </h2>
          </div>
        </div>

        {hidePrice ? (
          <p className="mt-2 text-[11px] font-medium text-slate-500">
            Biaya admin ditambahkan setelah hasil cek tagihan diterima.
          </p>
        ) : packageStyle ? (
          <p className="mt-2 text-[11px] font-medium text-slate-500">
            Harga: <span className="font-semibold text-slate-900">{formatRupiah(finalPrice)}</span>
          </p>
        ) : (
          <div className="mt-2 rounded-xl bg-linear-to-r from-slate-50 to-slate-100 px-4 py-3 border border-slate-200/50">
            <p className="text-[10px] font-semibold tracking-wide text-slate-500 uppercase">Harga</p>
            <p className="mt-1 text-sm font-bold text-slate-900">
              {isFixed && fixedPrice !== null ? formatRupiah(fixedPrice) : `+ ${formatRupiah(feeActive)}`}
            </p>
          </div>
        )}

        <div className="mt-2">
          <button
            type="button"
            onClick={() => {
              if (!canBuy) return;
              onBuy(item);
            }}
            disabled={!canBuy}
            className="inline-flex h-7 w-full items-center justify-center rounded-md bg-linear-to-r from-[#087e8b] to-[#087e8b] px-4 text-sm font-semibold text-white shadow-[0_8px_20px_rgba(15,111,203,0.24)] transition-all duration-300 hover:shadow-[0_12px_30px_rgba(15,111,203,0.4)] hover:scale-[1.02] active:scale-[0.98] disabled:cursor-not-allowed disabled:from-slate-300 disabled:to-slate-400 disabled:text-white disabled:shadow-none disabled:hover:scale-100"
          >
            {canBuy ? (buyLabel || "Beli") : (buyBlockedLabel || "Lengkapi dulu")}
          </button>
        </div>
      </div>
    </article>
  );
}

export function UserProductGrid({
  items,
  isLoggedIn,
  authToken,
  buyerRole,
  initialDest = "",
  canBuy = true,
  buyBlockedLabel,
  enableGuestHint = true,
  buyLabel,
  hidePrice = false,
}: UserProductGridProps) {
  const [selectedProduct, setSelectedProduct] = React.useState<UserProductItem | null>(null);
  const [showGuestHint, setShowGuestHint] = React.useState(false);

  React.useEffect(() => {
    if (isLoggedIn || !enableGuestHint) {
      setShowGuestHint(false);
      return;
    }

    const dismissed = window.sessionStorage.getItem("user-product-grid-guest-hint-dismissed");
    setShowGuestHint(dismissed !== "1");
  }, [enableGuestHint, isLoggedIn]);

  return (
    <>
      {!isLoggedIn && showGuestHint ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-slate-950/40 p-4 md:px-4">
          <div className="max-h-[92vh] w-full max-w-md overflow-y-auto rounded-[28px] bg-white p-5 shadow-[0_18px_46px_rgba(15,23,42,0.24)] md:w-97.5 md:max-w-none">
            <div className="flex items-start justify-between gap-3">
              <div>
                <h3 className="mt-1 text-lg font-bold text-slate-900">Potongan Harga</h3>
              </div>
              <button
                type="button"
                onClick={() => { window.sessionStorage.setItem("user-product-grid-guest-hint-dismissed", "1"); setShowGuestHint(false); }}
                className="grid h-10 w-10 place-items-center rounded-2xl bg-slate-100 text-slate-500"
                aria-label="Tutup informasi login"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
            <p className="mt-3 text-sm text-slate-600">Dapatkan potongan harga spesial dengan login atau mendaftar akun.</p>
            <div className="mt-4 grid grid-cols-2 gap-2">
              <Link
                href="/login?callbackUrl=/user"
                className="inline-flex h-10 items-center justify-center rounded-2xl bg-linear-to-r from-[#087e8b] to-[#087e8b] px-4 text-sm font-semibold text-white! visited:text-white! hover:text-white!"
              >
                Masuk
              </Link>
              <Link
                href="/register"
                className="inline-flex h-10 items-center justify-center rounded-2xl border border-sky-200 bg-white px-4 text-sm font-semibold text-sky-700"
              >
                Daftar
              </Link>
            </div>
          </div>
        </div>
      ) : null}

      <div className="grid grid-cols-2 gap-3">
        {items.map((item) => (
          <UserProductCard
            key={item.id}
            item={item}
            isLoggedIn={isLoggedIn}
            buyerRole={buyerRole}
            onBuy={setSelectedProduct}
            canBuy={canBuy}
            buyBlockedLabel={buyBlockedLabel}
            buyLabel={buyLabel}
            hidePrice={hidePrice}
          />
        ))}
      </div>

      <UserCheckoutModal
        open={Boolean(selectedProduct)}
        product={selectedProduct}
        authToken={authToken}
        buyerRole={buyerRole}
        initialDest={initialDest}
        onClose={() => setSelectedProduct(null)}
      />
    </>
  );
}
