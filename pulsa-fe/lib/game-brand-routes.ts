import type { UserBrandItem } from "@/components/user/types";

function normalizeSlugPart(value: string) {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

const GAME_BRAND_SLUG_OVERRIDES: Record<string, string> = {
  "free fire": "freefire",
  "free fire max": "freefiremax",
  "mobile legend": "mobilelegends",
  "mobile legends": "mobilelegends",
  "mobile legends bang bang": "mobilelegends",
  "magic chess go go": "magicchessgogo",
  "pubg mobile": "pubgmobile",
  "call of duty mobile": "callofdutymobile",
  "genshin impact": "genshinimpact",
  "blood strike": "bloodstrike",
  "point blank": "pointblank",
  "roblox gift card": "roblox",
};

export function getGameBrandSlug(name: string) {
  const normalized = name.trim().toLowerCase();
  return GAME_BRAND_SLUG_OVERRIDES[normalized] ?? normalizeSlugPart(name);
}

export function findGameBrandBySlug(brands: UserBrandItem[], slug: string) {
  const target = normalizeSlugPart(slug);
  return brands.find((brand) => getGameBrandSlug(brand.nama) === target) ?? null;
}
