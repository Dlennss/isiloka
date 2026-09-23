import type { YuscomFamilyVisual } from "@/lib/yuscom-family-visuals";
import { getYuscomFamilyVisual } from "@/lib/yuscom-family-visuals";
import generatedBrandMap from "@/lib/generated-yuscom-display-brand-map.json";

function image(src: string, alt: string, sourcePage: string): YuscomFamilyVisual {
  return { kind: "image", src, alt, sourcePage };
}

const DISPLAY_BRAND_VISUALS: Record<string, YuscomFamilyVisual> = {
  INDOVISION: image("/yuscom-display-brand-logos/indovision.jpeg", "Logo IndiVision", "https://www.indovision.tv/"),
  OKEVISION: { kind: "badge", label: "OkeVision", tone: "sky" },
  "TOP TV": { kind: "badge", label: "Top TV", tone: "amber" },
  TRANSVISION: image("/yuscom-display-brand-logos/transvision.png", "Logo Transvision", "https://commons.wikimedia.org/wiki/File:TransVision_logo.png"),
  TELKOMVISION: image("/yuscom-family-logos/telkom.svg", "Logo Telkom", "https://commons.wikimedia.org/wiki/File:Telkom-Logo.svg"),
  "MNC PLAY": image("/yuscom-display-brand-logos/mncplay.png", "Logo MNC Play", "https://mncplay.id/"),
  "MY REPUBLIK": image("/yuscom-display-brand-logos/myrepublic.ico", "Logo MyRepublic", "https://www.myrepublic.co.id/"),
  "YES TV": { kind: "badge", label: "Yes TV", tone: "emerald" },
  BIZZNET: image("/yuscom-display-brand-logos/biznet-home.png", "Logo Biznet Home", "https://biznethome.net/"),
  ICONNET: { kind: "badge", label: "Iconnet", tone: "amber" },
  PGN: image("/yuscom-display-brand-logos/pgn.png", "Logo PGN", "https://www.pgn.co.id/"),
  "BY.U": image("/yuscom-display-brand-logos/byu.png", "Logo by.U", "https://www.byu.id/"),
  SMARTFREN: image("/yuscom-display-brand-logos/smartfren.png", "Logo Smartfren", "local:/home/syarif/app/logosmartfern.png"),
  KASPRO: image("/yuscom-display-brand-logos/kaspro.png", "Logo KasPro", "https://kaspro.id/"),
  MAXIM: image("/yuscom-display-brand-logos/maxim.ico", "Logo Maxim", "https://www.maxim.id/"),
  "FREE FIRE": image("/images/games/banner_freefire.png", "Logo Free Fire", "local:/public/images/games/banner_freefire.png"),
  "PUBG MOBILE": image("/images/games/banner_pubg.png", "Logo PUBG Mobile", "local:/public/images/games/banner_pubg.png"),
  "MOBILE LEGEND": image("/images/games/banner_mobile_legends.png", "Logo Mobile Legends", "local:/public/images/games/banner_mobile_legends.png"),
  ROBLOX: image("/images/games/banner_roblox.png", "Logo Roblox", "local:/public/images/games/banner_roblox.png"),
  "POINT BLANK": image("/images/games/banner_point_blank.png", "Logo Point Blank", "local:/public/images/games/banner_point_blank.png"),
  "SPEEDY DAN INDIHOME": image("/yuscom-family-logos/telkom.svg", "Logo IndiHome", "https://commons.wikimedia.org/wiki/File:Telkom-Logo.svg"),
  PLN: image("/yuscom-display-brand-logos/pln.svg", "Logo PLN", "https://commons.wikimedia.org/wiki/File:Logo_PLN_(cropped).svg"),
};

function normalizeKey(v: string) {
  return v.trim().toUpperCase();
}

export function getYuscomDisplayBrandVisual(displayBrand: string): YuscomFamilyVisual {
  const key = normalizeKey(displayBrand);
  if (DISPLAY_BRAND_VISUALS[key]) {
    return DISPLAY_BRAND_VISUALS[key];
  }
  if (displayBrand.trim()) {
    const generatedFile = (generatedBrandMap as Record<string, string>)[displayBrand];
    if (generatedFile) {
      return image(
        `/yuscom-display-brand-logos-generated/${generatedFile}`,
        `Logo ${displayBrand}`,
        "generated-local-brand-mark"
      );
    }
  }
  return getYuscomFamilyVisual(displayBrand);
}
