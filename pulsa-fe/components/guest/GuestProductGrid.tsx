"use client";

import * as React from "react";
import Link from "next/link";
import { X } from "lucide-react";
import type { UserProductItem } from "@/components/user/types";
import { GuestProductDefaultCard } from "@/components/guest/GuestProductDefaultCard";
import { GuestProductGameCard } from "@/components/guest/GuestProductGameCard";
import { GuestProductListCard } from "@/components/guest/GuestProductListCard";
import { GuestProductVoucherCard } from "@/components/guest/GuestProductVoucherCard";
import { UserCheckoutModal } from "@/components/user/UserCheckoutModal";
import { PurchaseBrandBanner } from "@/components/shared/PurchaseBrandBanner";
import { getBrandLogo } from "@/lib/brand-logos";

type GuestProductGridProps = {
  items: UserProductItem[];
  isLoggedIn: boolean;
  authToken?: string;
  buyerRole?: string;
  initialDest?: string;
  destLabel?: string;
  destPlaceholder?: string;
  destMode?: "single" | "ml_id_server" | "alphanumeric";
  canBuy?: boolean;
  buyBlockedLabel?: string;
  buyLabel?: string;
  hidePrice?: boolean;
  layout?: "grid" | "list";
  visualStyle?: "default" | "voucher" | "game";
  showBrandBanner?: boolean;
  enableGuestHint?: boolean;
};

function GuestProductCard({
  item,
  isLoggedIn,
  onBuy,
  canBuy,
  buyBlockedLabel,
  buyLabel,
  hidePrice,
  layout,
  visualStyle,
}: {
  item: UserProductItem;
  isLoggedIn: boolean;
  onBuy: (item: UserProductItem) => void;
  canBuy: boolean;
  buyBlockedLabel?: string;
  buyLabel?: string;
  hidePrice?: boolean;
  layout?: "grid" | "list";
  visualStyle?: "default" | "voucher" | "game";
}) {
  if (visualStyle === "game") {
    return <GuestProductGameCard item={item} isLoggedIn={isLoggedIn} onBuy={onBuy} canBuy={canBuy} buyBlockedLabel={buyBlockedLabel} buyLabel={buyLabel} hidePrice={hidePrice} />;
  }

  if (layout === "list" && visualStyle === "voucher") {
    return <GuestProductVoucherCard item={item} isLoggedIn={isLoggedIn} onBuy={onBuy} canBuy={canBuy} buyBlockedLabel={buyBlockedLabel} buyLabel={buyLabel} hidePrice={hidePrice} />;
  }

  if (layout === "list") {
    return <GuestProductListCard item={item} isLoggedIn={isLoggedIn} onBuy={onBuy} canBuy={canBuy} buyBlockedLabel={buyBlockedLabel} buyLabel={buyLabel} hidePrice={hidePrice} />;
  }

  return <GuestProductDefaultCard item={item} isLoggedIn={isLoggedIn} onBuy={onBuy} canBuy={canBuy} buyBlockedLabel={buyBlockedLabel} buyLabel={buyLabel} hidePrice={hidePrice} />;
}

export function GuestProductGrid({
  items,
  isLoggedIn,
  authToken,
  buyerRole,
  initialDest = "",
  destLabel,
  destPlaceholder,
  destMode = "single",
  canBuy = true,
  buyBlockedLabel,
  buyLabel,
  hidePrice = false,
  layout = "grid",
  visualStyle = "default",
  showBrandBanner = true,
  enableGuestHint = true,
}: GuestProductGridProps) {
  const [selectedProduct, setSelectedProduct] = React.useState<UserProductItem | null>(null);
  const [showGuestHint, setShowGuestHint] = React.useState(false);
  const brandName = String(items[0]?.brand_nama || "").trim();
  const brandLogo = brandName ? getBrandLogo(brandName) : null;

  React.useEffect(() => {
    if (isLoggedIn || !enableGuestHint) {
      setShowGuestHint(false);
      return;
    }
    const dismissed = window.sessionStorage.getItem("guest-product-grid-hint-dismissed");
    setShowGuestHint(dismissed !== "1");
  }, [enableGuestHint, isLoggedIn]);

  return (
    <>
      {!isLoggedIn && enableGuestHint && showGuestHint ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-slate-950/40 p-4 md:px-4">
          <div className="max-h-[92vh] w-full max-w-md overflow-y-auto rounded-[28px] bg-white p-5 shadow-[0_18px_46px_rgba(15,23,42,0.24)] md:w-97.5 md:max-w-none">
            <div className="flex items-start justify-between gap-3">
              <div>
                <h3 className="mt-1 text-lg font-bold text-slate-900">Potongan Harga</h3>
              </div>
              <button
                type="button"
                onClick={() => {
                  window.sessionStorage.setItem("guest-product-grid-hint-dismissed", "1");
                  setShowGuestHint(false);
                }}
                className="grid h-10 w-10 place-items-center rounded-2xl bg-slate-100 text-slate-500"
                aria-label="Tutup informasi login"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
            <p className="mt-3 text-sm text-slate-600">Dapatkan potongan harga spesial dengan login atau mendaftar akun.</p>
            <div className="mt-4 grid grid-cols-2 gap-2">
              <Link
                href="/login?callbackUrl=/"
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

      {showBrandBanner && brandName ? (
        <section className="overflow-hidden rounded-md border border-sky-100 bg-white shadow-[0_18px_40px_rgba(15,23,42,0.10)]">
          <PurchaseBrandBanner
            title={brandName}
            logoSrc={brandLogo?.src}
            logoAlt={brandLogo?.alt || brandName}
            fallbackText={brandName}
          />
        </section>
      ) : null}

      <div className={layout === "list" ? " space-y-3 mt-3" : "grid grid-cols-2 gap-3"}>
        {items.map((item) => (
          <GuestProductCard
            key={item.id}
            item={item}
            isLoggedIn={isLoggedIn}
            onBuy={setSelectedProduct}
            canBuy={canBuy}
            buyBlockedLabel={buyBlockedLabel}
            buyLabel={buyLabel}
            hidePrice={hidePrice}
            layout={layout}
            visualStyle={visualStyle}
          />
        ))}
      </div>

      <UserCheckoutModal
        open={Boolean(selectedProduct)}
        product={selectedProduct}
        authToken={authToken}
        buyerRole={buyerRole}
        initialDest={initialDest}
        destLabel={destLabel}
        destPlaceholder={destPlaceholder}
        destMode={destMode}
        onClose={() => setSelectedProduct(null)}
      />
    </>
  );
}
