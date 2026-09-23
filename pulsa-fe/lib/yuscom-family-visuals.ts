export type YuscomFamilyVisual =
  | {
      kind: "image";
      src: string;
      alt: string;
      sourcePage: string;
    }
  | {
      kind: "badge";
      label: string;
      tone: "sky" | "emerald" | "violet" | "rose" | "amber" | "slate";
    };

function image(src: string, alt: string, sourcePage: string): YuscomFamilyVisual {
  return { kind: "image", src, alt, sourcePage };
}

function badge(
  label: string,
  tone: "sky" | "emerald" | "violet" | "rose" | "amber" | "slate"
): YuscomFamilyVisual {
  return { kind: "badge", label, tone };
}

const FAMILY_VISUALS: Record<string, YuscomFamilyVisual> = {
  DANA: image("/brand-logos/dana.svg", "Logo DANA", "https://commons.wikimedia.org/wiki/File:Logo_Dana_Wiki.svg"),
  GOPAY: image("/brand-logos/gopay.svg", "Logo GoPay", "https://commons.wikimedia.org/wiki/File:Gopay_logo.svg"),
  SHOPEEPAY: image("/brand-logos/shopee.svg", "Logo ShopeePay", "https://commons.wikimedia.org/wiki/File:Shopee.svg"),
  LINKAJA: image("/brand-logos/linkaja.svg", "Logo LinkAja", "https://commons.wikimedia.org/wiki/File:LinkAja.svg"),
  OVO: image("/brand-logos/ovo.svg", "Logo OVO", "https://commons.wikimedia.org/wiki/File:Logo_ovo_purple.svg"),
  TELKOMSEL: image("/brand-logos/telkomsel.svg", "Logo Telkomsel", "https://commons.wikimedia.org/wiki/File:Telkomsel_2021_icon.svg"),
  AXIS: image("/brand-logos/axis.svg", "Logo AXIS", "https://commons.wikimedia.org/wiki/File:Axis_logo_2014.svg"),
  TRI: image("/brand-logos/tri.svg", "Logo Tri", "https://commons.wikimedia.org/wiki/File:Three_logo.svg"),
  INDOSAT: image("/yuscom-family-logos/indosat.svg", "Logo Indosat", "https://commons.wikimedia.org/wiki/File:Indosat_Ooredoo_(2).svg"),
  XL: image("/yuscom-family-logos/xl.svg", "Logo XL Axiata", "https://commons.wikimedia.org/wiki/File:XL_Axiata_2014.svg"),
  "BY.U": image("/yuscom-display-brand-logos/byu.png", "Logo by.U", "https://www.byu.id/"),
  SMARTFREN: image("/yuscom-family-logos/smartfren.png", "Logo Smartfren", "local:/home/syarif/app/logosmartfern.png"),
  MAXIM: image("/yuscom-display-brand-logos/maxim.ico", "Logo Maxim", "https://www.maxim.id/"),
  KASPRO: image("/yuscom-display-brand-logos/kaspro.png", "Logo KasPro", "https://kaspro.id/"),
  GRAB: image("/yuscom-family-logos/grab.svg", "Logo Grab", "https://commons.wikimedia.org/wiki/File:Grab_Logo.svg"),
  ISAKU: badge("i.saku", "rose"),
  "FREE FIRE": image("/images/games/banner_freefire.png", "Logo Free Fire", "local:/public/images/games/banner_freefire.png"),
  "PUBG MOBILE": image("/images/games/banner_pubg.png", "Logo PUBG Mobile", "local:/public/images/games/banner_pubg.png"),
  "MOBILE LEGEND": image("/images/games/banner_mobile_legends.png", "Logo Mobile Legends", "local:/public/images/games/banner_mobile_legends.png"),
  ROBLOX: image("/images/games/banner_roblox.png", "Logo Roblox", "local:/public/images/games/banner_roblox.png"),
  "ROBLOX GIFT CARD": image("/images/games/banner_roblox.png", "Logo Roblox", "local:/public/images/games/banner_roblox.png"),
  "VOUCHER POINT BLANK ZEPETTO": image("/images/games/banner_point_blank.png", "Logo Point Blank", "local:/public/images/games/banner_point_blank.png"),
  PDAM: badge("PDAM", "sky"),
  LISTRIK: badge("PLN", "amber"),
  "TV PASCABAYAR": badge("TV", "violet"),
  "K-VISION": image("/yuscom-family-logos/kvision.png", "Logo K-Vision", "https://www.k-vision.tv/"),
  "TV NEX PARABOLA 30 HARI": image("/yuscom-family-logos/nexparabola.svg", "Logo Nex Parabola", "https://www.mynex.co.id/"),
  BPJS: image("/yuscom-family-logos/bpjs-kesehatan.svg", "Logo BPJS Kesehatan", "https://commons.wikimedia.org/wiki/File:BPJS_Kesehatan_logo.svg"),
  "HP PASCABAYAR": badge("HP", "slate"),
  "INTERNET PASCABAYAR": image("/yuscom-family-logos/telkom.svg", "Logo Telkom", "https://commons.wikimedia.org/wiki/File:Telkom-Logo.svg"),
  "PASCABAYAR GAS NEGARA": badge("Gas", "amber"),
  TELKOM: image("/yuscom-family-logos/telkom.svg", "Logo Telkom", "https://commons.wikimedia.org/wiki/File:Telkom-Logo.svg"),
};

function normalizeKey(v: string) {
  return v.trim().toUpperCase();
}

export function getYuscomFamilyVisual(family: string): YuscomFamilyVisual {
  const key = normalizeKey(family);
  return FAMILY_VISUALS[key] ?? badge(family || "Produk", "slate");
}
