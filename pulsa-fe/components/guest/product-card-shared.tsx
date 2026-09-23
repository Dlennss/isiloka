import type { UserProductItem } from "@/components/user/types";

export type GuestProductCardCommonProps = {
  item: UserProductItem;
  isLoggedIn: boolean;
  onBuy: (item: UserProductItem) => void;
  canBuy: boolean;
  buyBlockedLabel?: string;
  buyLabel?: string;
  hidePrice?: boolean;
};

export function formatRupiah(value: number) {
  return `Rp ${new Intl.NumberFormat("id-ID").format(value || 0)}`;
}

function normalizeDisplayRole(role?: string | null) {
  const value = String(role || "").trim().toLowerCase();
  if (value === "master") return "master";
  if (value === "agent") return "agent";
  if (value === "user") return "user";
  return "guest";
}

export function getDisplayedFixedPrice(item: UserProductItem, role?: string | null) {
  const normalizedRole = normalizeDisplayRole(role);
  const rolePrice = normalizedRole === "master"
    ? Number(item.harga_master_final || 0)
    : normalizedRole === "agent"
      ? Number(item.harga_agent_final || 0)
      : normalizedRole === "user"
        ? Number(item.harga_user_final || 0)
        : Number(item.harga_guest_final || 0);
  if (rolePrice > 0) return rolePrice;

  const basePrice = Number(item.harga_dasar_app || item.nominal || 0);
  switch (normalizedRole) {
    case "master":
      return basePrice + Number(item.fee_master || 0);
    case "agent":
      return basePrice + Number(item.fee_agent || 0);
    case "user":
      return basePrice + Number(item.fee_user || 0);
    default:
      return basePrice + Number(item.fee_guest || 0);
  }
}

export function isPackageStyleItem(item: UserProductItem) {
  const upper = String(item.kategori_nama || item.nama || "").toUpperCase();
  return upper.includes("PAKET DATA") || upper.includes("PAKET TELEPON") || upper.includes("PAKET SMS");
}

export function isEMoneyFixedItem(item: UserProductItem) {
  return String(item.kategori_nama || "").toUpperCase().includes("E-MONEY") && item.tipe_harga === "FIXED";
}

export function isGameItem(item: UserProductItem) {
  return String(item.kategori_nama || "").toUpperCase().includes("GAME");
}

function escapeRegExp(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

export function getDisplayProductName(item: UserProductItem) {
  const raw = String(item.nama || "").trim();
  if (!raw) return raw;

  const category = String(item.kategori_nama || "").trim().toUpperCase();
  const brand = String(item.brand_nama || "").trim();

  if (category.includes("LISTRIK") || brand.toUpperCase() === "PLN") {
    const cleaned = raw.replace(/^token\s+listrik\s*/i, "").trim();
    return cleaned || raw;
  }

  if (!brand) return raw;

  const escapedBrand = escapeRegExp(brand).replace(/\s+/g, "\\s+");

  if (category.includes("TV")) {
    const cleaned = raw.replace(new RegExp(`^${escapedBrand}\\s*[:-]?\\s*`, "i"), "").trim();
    return cleaned || raw;
  }

  if (isPackageStyleItem(item)) {
    const words = raw.replace(/\s+/g, " ").split(" ");
    const firstNumericIndex = words.findIndex((word) => /\d/.test(word));
    if (firstNumericIndex > 0) {
      const cleaned = words.slice(firstNumericIndex).join(" ").trim();
      if (cleaned) return cleaned;
    }
  }

  if (!isGameItem(item)) return raw;

  const cleaned = raw
    .replace(new RegExp(`^(?:voucher\\s+)?${escapedBrand}\\s*[:-]?\\s*`, "i"), "")
    .replace(new RegExp(`^${escapedBrand}\\s+voucher\\s*[:-]?\\s*`, "i"), "")
    .trim();

  return cleaned || raw;
}

export function extractLargeNominalLabel(item: UserProductItem) {
  const upper = String(item.nama || "").toUpperCase();
  const dotted = upper.match(/(\d+(?:\.\d{3})+)/);
  if (dotted) return dotted[1];

  const compact = upper.match(/(\d+)\s*K\b/);
  if (compact) return `${Number.parseInt(compact[1], 10)}.000`;

  return new Intl.NumberFormat("id-ID").format(Number(item.nominal || 0));
}

export function getProductPricing(item: UserProductItem, isLoggedIn: boolean) {
  const isFixed = item.tipe_harga === "FIXED";
  const feeActive = isLoggedIn ? item.fee_user : item.fee_guest;
  const baseCharge = Number(item.harga_dasar_app || 0);
  const fixedPrice = isFixed ? getDisplayedFixedPrice(item, isLoggedIn ? "user" : "guest") : null;
  const openAmountPrice = !isFixed ? baseCharge + feeActive : null;
  const finalPrice = Number(fixedPrice || openAmountPrice || 0);
  return { isFixed, feeActive, fixedPrice, openAmountPrice, finalPrice };
}
