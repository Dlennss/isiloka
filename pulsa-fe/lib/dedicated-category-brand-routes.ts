import type { UserBrandItem } from "@/components/user/types";
import { getGameBrandSlug } from "@/lib/game-brand-routes";

function normalizeSlugPart(value: string) {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

function getGenericBrandSlug(name: string) {
  return normalizeSlugPart(name);
}

const GUEST_CATEGORY_BRAND_BASE: Record<string, string> = {
  "1": "/pulsa",
  "2": "/paket-data",
  "3": "/ewallet",
  "4": "/listrik",
  "5": "/game",
  "7": "/tv",
  "8": "/aktivasi-perdana",
  "9": "/masa-aktif",
  "10": "/paket-telepon",
  "17": "/pdam",
  "18": "/hp-pascabayar",
  "20": "/pgn",
};

export function getDedicatedGuestBrandPath(kategoriId: string, brand: Pick<UserBrandItem, "nama">) {
  const base = GUEST_CATEGORY_BRAND_BASE[String(kategoriId)];
  if (!base) return null;
  const slug = String(kategoriId) === "5" ? getGameBrandSlug(brand.nama) : getGenericBrandSlug(brand.nama);
  return `${base}/${slug}`;
}

export function findBrandByDedicatedSlug(kategoriId: string, brands: UserBrandItem[], slug: string) {
  const normalized = normalizeSlugPart(slug);
  return brands.find((brand) => {
    const current = String(kategoriId) === "5" ? getGameBrandSlug(brand.nama) : getGenericBrandSlug(brand.nama);
    return current === normalized;
  }) ?? null;
}
