import type { GuestProductCardCommonProps } from "@/components/guest/product-card-shared";
import { formatRupiah, getDisplayProductName, getProductPricing } from "@/components/guest/product-card-shared";

export function GuestProductVoucherCard({
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
      className="relative w-full overflow-hidden rounded-md bg-[#0084D1] px-5 py-2 text-left text-white shadow-[0_14px_30px_rgba(0,132,209,0.22)] disabled:cursor-not-allowed disabled:opacity-60"
    >
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(255,255,255,0.22),transparent_30%),linear-gradient(180deg,rgba(255,255,255,0.06),rgba(255,255,255,0.02))]" />
      <div className="absolute inset-0 opacity-30 bg-[repeating-radial-gradient(circle_at_0_100%,rgba(255,255,255,0.35)_0,rgba(255,255,255,0.35)_2px,transparent_2px,transparent_12px)] bg-size-[160%_120%]" />

      <div className="relative flex min-h-18 flex-col gap-2 ">
        <div className="min-w-0">
          <h2 className="line-clamp-3 text-xl font-bold leading-tight tracking-tight text-white">
            {displayName}
          </h2>
          {hidePrice ? (
            <p className="mt-2 text-[12px] text-white/75">Biaya admin ditambahkan setelah hasil cek tagihan diterima.</p>
          ) : null}
        </div>

        <div className="min-w-0">
          <p className="text-sm font-bold leading-none tracking-tight text-white">
            {isFixed && fixedPrice !== null ? formatRupiah(fixedPrice).replace("Rp ", "Rp") : `+${formatRupiah(openAmountPrice ?? feeActive).replace("Rp ", "Rp")}`}
          </p>
        </div>
      </div>
    </button>
  );
}
