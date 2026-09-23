import type { UserProductItem } from "@/components/user/types";
import { getBrandLogo } from "@/lib/brand-logos";
import generatedBrandMap from "@/lib/generated-yuscom-display-brand-map.json";
import { CANONICAL_SITE_URL, DEFAULT_OG_IMAGE_URL } from "@/lib/seo-articles";
import { getYuscomDisplayBrandVisual } from "@/lib/yuscom-display-brand-visuals";

type ProductRichImageInput = {
  brandName?: string;
  categoryName?: string;
  items?: UserProductItem[];
};

function toAbsoluteUrl(src: string) {
  if (/^https?:\/\//i.test(src)) return src;
  return `${CANONICAL_SITE_URL}${src.startsWith("/") ? src : `/${src}`}`;
}

function normalizeName(value: string) {
  return value.trim().toLowerCase().replace(/\s+/g, " ");
}

function uniqueNames(values: Array<string | undefined>) {
  const seen = new Set<string>();
  const result: string[] = [];

  for (const value of values) {
    const trimmed = String(value || "").trim();
    if (!trimmed) continue;
    const key = normalizeName(trimmed);
    if (seen.has(key)) continue;
    seen.add(key);
    result.push(trimmed);
  }

  return result;
}

function categoryFallbackImage(categoryName: string) {
  const normalized = normalizeName(categoryName);

  if (normalized.includes("bpjs")) return "/yuscom-display-brand-logos-generated/bpjs.svg";
  if (normalized.includes("pdam")) return "/images/logo-pulsakilat.svg";
  if (normalized.includes("listrik")) return "/yuscom-display-brand-logos/pln.svg";
  if (normalized.includes("game")) return "/images/games/banner_mobile_legends.png";
  if (normalized.includes("tv")) return "/images/tv/logo_indovision.png";
  if (normalized.includes("internet")) return "/images/internet/logo_biznet.png";
  if (normalized.includes("gas") || normalized.includes("pgn")) return "/images/gas/Logo_PGN.png";
  if (normalized.includes("e-wallet") || normalized.includes("ewallet")) return "/images/ewallet/logo_dana.png";

  return null;
}

function generatedBrandImage(brandName: string) {
  const generatedFile = (generatedBrandMap as Record<string, string>)[brandName];
  return generatedFile ? `/yuscom-display-brand-logos-generated/${generatedFile}` : null;
}

export function getRichProductImageUrl({ brandName, categoryName, items = [] }: ProductRichImageInput) {
  const candidateNames = uniqueNames([
    brandName,
    ...items.map((item) => String(item.brand_nama || "").trim()),
    ...items.map((item) => String(item.kategori_nama || "").trim()),
  ]);

  for (const candidate of candidateNames) {
    const brandLogo = getBrandLogo(candidate);
    if (brandLogo?.src) {
      return toAbsoluteUrl(brandLogo.src);
    }
  }

  for (const candidate of candidateNames) {
    const visual = getYuscomDisplayBrandVisual(candidate);
    if (visual.kind === "image" && visual.src) {
      return toAbsoluteUrl(visual.src);
    }

    const generatedImage = generatedBrandImage(candidate);
    if (generatedImage) {
      return toAbsoluteUrl(generatedImage);
    }
  }

  const categoryImage = categoryFallbackImage(
    categoryName || String(items[0]?.kategori_nama || "").trim(),
  );
  if (categoryImage) {
    return toAbsoluteUrl(categoryImage);
  }

  return DEFAULT_OG_IMAGE_URL;
}
