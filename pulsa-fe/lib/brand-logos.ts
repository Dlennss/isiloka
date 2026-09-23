export type BrandLogoMeta = {
  src: string;
  alt: string;
  sourcePage: string;
};

const BRAND_LOGOS: Record<string, BrandLogoMeta> = {
  dana: {
    src: "/brand-logos/dana.svg",
    alt: "Logo DANA",
    sourcePage: "https://commons.wikimedia.org/wiki/File:Logo_Dana_Wiki.svg",
  },
  telkomsel: {
    src: "/images/providers/logo_telkomsel.webp",
    alt: "Logo Telkomsel",
    sourcePage: "local:/public/images/providers/logo_telkomsel.webp",
  },
  byu: {
    src: "/images/providers/logo_byu.webp",
    alt: "Logo by.U",
    sourcePage: "local:/public/images/providers/logo_byu.webp",
  },
  axis: {
    src: "/images/providers/logo_axis.webp",
    alt: "Logo AXIS",
    sourcePage: "local:/public/images/providers/logo_axis.webp",
  },
  tri: {
    src: "/images/providers/logo_tri.webp",
    alt: "Logo Tri",
    sourcePage: "local:/public/images/providers/logo_tri.webp",
  },
  gopay: {
    src: "/brand-logos/gopay.svg",
    alt: "Logo GoPay",
    sourcePage: "https://commons.wikimedia.org/wiki/File:Gopay_logo.svg",
  },
  indosat: {
    src: "/images/providers/logo_im3.webp",
    alt: "Logo Indosat",
    sourcePage: "local:/public/images/providers/logo_im3.webp",
  },
  im3: {
    src: "/images/providers/logo_im3.webp",
    alt: "Logo Indosat",
    sourcePage: "local:/public/images/providers/logo_im3.webp",
  },
  mentari: {
    src: "/images/providers/logo_im3.webp",
    alt: "Logo Indosat",
    sourcePage: "local:/public/images/providers/logo_im3.webp",
  },
  xl: {
    src: "/images/providers/logo_xl.png",
    alt: "Logo XL",
    sourcePage: "local:/public/images/providers/logo_xl.png",
  },
  "xl axiata": {
    src: "/images/providers/logo_xl.png",
    alt: "Logo XL Axiata",
    sourcePage: "local:/public/images/providers/logo_xl.png",
  },
  "link aja": {
    src: "/images/ewallet/logo_linkaja.png",
    alt: "Logo LinkAja",
    sourcePage: "local:/public/images/ewallet/logo_linkaja.png",
  },
  linkaja: {
    src: "/images/ewallet/logo_linkaja.png",
    alt: "Logo LinkAja",
    sourcePage: "local:/public/images/ewallet/logo_linkaja.png",
  },
  ovo: {
    src: "/brand-logos/ovo.svg",
    alt: "Logo OVO",
    sourcePage: "https://commons.wikimedia.org/wiki/File:Logo_ovo_purple.svg",
  },
  shopee: {
    src: "/brand-logos/shopee.svg",
    alt: "Logo Shopee",
    sourcePage: "https://commons.wikimedia.org/wiki/File:Shopee.svg",
  },
  shopeepay: {
    src: "/brand-logos/shopee.svg",
    alt: "Logo ShopeePay",
    sourcePage: "https://commons.wikimedia.org/wiki/File:Shopee.svg",
  },
  "shopee pay": {
    src: "/brand-logos/shopee.svg",
    alt: "Logo ShopeePay",
    sourcePage: "https://commons.wikimedia.org/wiki/File:Shopee.svg",
  },
  astrapay: {
    src: "/images/ewallet/logo_astrapay.svg",
    alt: "Logo AstraPay",
    sourcePage: "local:/public/images/ewallet/logo_astrapay.svg",
  },
  "astra pay": {
    src: "/images/ewallet/logo_astrapay.svg",
    alt: "Logo AstraPay",
    sourcePage: "local:/public/images/ewallet/logo_astrapay.svg",
  },
  isaku: {
    src: "/images/ewallet/logo_isaku.svg",
    alt: "Logo i.saku",
    sourcePage: "local:/public/images/ewallet/logo_isaku.svg",
  },
  "i.saku": {
    src: "/images/ewallet/logo_isaku.svg",
    alt: "Logo i.saku",
    sourcePage: "local:/public/images/ewallet/logo_isaku.svg",
  },
  smartfren: {
    src: "/images/providers/logo_smartfren.webp",
    alt: "Logo Smartfren",
    sourcePage: "local:/public/images/providers/logo_smartfren.webp",
  },
  pln: {
    src: "/images/pln/logo_pln.png",
    alt: "Logo PLN",
    sourcePage: "local:/public/images/pln/logo_pln.png",
  },
  pdam: {
    src: "/images/pdam/logo_pdam.png",
    alt: "Logo PDAM",
    sourcePage: "local:/public/images/pdam/logo_pdam.png",
  },
  "bpjs kesehatan": {
    src: "/images/bpjs/icon_bpjs_kesehatan.png",
    alt: "Logo BPJS Kesehatan",
    sourcePage: "local:/public/images/bpjs/icon_bpjs_kesehatan.png",
  },
  "bpjs ketenagakerjaan": {
    src: "/images/bpjs/icon_bpjs_ketenagakerjaan.png",
    alt: "Logo BPJS Ketenagakerjaan",
    sourcePage: "local:/public/images/bpjs/icon_bpjs_ketenagakerjaan.png",
  },
  bca: {
    src: "/images/banks/bca.jpeg",
    alt: "Logo BCA",
    sourcePage: "local:/public/images/banks/bca.jpeg",
  },
  bni: {
    src: "/images/banks/bni.jpg",
    alt: "Logo BNI",
    sourcePage: "local:/public/images/banks/bni.jpg",
  },
  bri: {
    src: "/images/banks/bri.jpg",
    alt: "Logo BRI",
    sourcePage: "local:/public/images/banks/bri.jpg",
  },
  bsi: {
    src: "/images/banks/bsi.jpg",
    alt: "Logo BSI",
    sourcePage: "local:/public/images/banks/bsi.jpg",
  },
  mandiri: {
    src: "/images/banks/mandiri.jpeg",
    alt: "Logo Bank Mandiri",
    sourcePage: "local:/public/images/banks/mandiri.jpeg",
  },
  "bank mandiri": {
    src: "/images/banks/mandiri.jpeg",
    alt: "Logo Bank Mandiri",
    sourcePage: "local:/public/images/banks/mandiri.jpeg",
  },
  danamon: {
    src: "/images/banks/danamon.png",
    alt: "Logo Bank Danamon",
    sourcePage: "local:/public/images/banks/danamon.png",
  },
  permata: {
    src: "/images/banks/permatabank.jpeg",
    alt: "Logo PermataBank",
    sourcePage: "local:/public/images/banks/permatabank.jpeg",
  },
  "bank permata": {
    src: "/images/banks/permatabank.jpeg",
    alt: "Logo PermataBank",
    sourcePage: "local:/public/images/banks/permatabank.jpeg",
  },
  seabank: {
    src: "/images/banks/seabank.jpeg",
    alt: "Logo SeaBank",
    sourcePage: "local:/public/images/banks/seabank.jpeg",
  },
  "sea bank": {
    src: "/images/banks/seabank.jpeg",
    alt: "Logo SeaBank",
    sourcePage: "local:/public/images/banks/seabank.jpeg",
  },
  jago: {
    src: "/images/banks/jago.jpg",
    alt: "Logo Bank Jago",
    sourcePage: "local:/public/images/banks/jago.jpg",
  },
  "bank jago": {
    src: "/images/banks/jago.jpg",
    alt: "Logo Bank Jago",
    sourcePage: "local:/public/images/banks/jago.jpg",
  },
  biznet: {
    src: "/images/internet/logo_biznet.png",
    alt: "Logo Biznet",
    sourcePage: "local:/public/images/internet/logo_biznet.png",
  },
  bizznet: {
    src: "/images/internet/logo_biznet.png",
    alt: "Logo Bizznet",
    sourcePage: "local:/public/images/internet/logo_biznet.png",
  },
  speedy: {
    src: "/images/internet/logo_speedy.png",
    alt: "Logo Speedy",
    sourcePage: "local:/public/images/internet/logo_speedy.png",
  },
  "speedy dan indihome": {
    src: "/images/internet/logo_speedy.png",
    alt: "Logo Speedy dan IndiHome",
    sourcePage: "local:/public/images/internet/logo_speedy.png",
  },
  indihome: {
    src: "/images/internet/logo_speedy.png",
    alt: "Logo IndiHome",
    sourcePage: "local:/public/images/internet/logo_speedy.png",
  },
  iconet: {
    src: "/images/internet/logo_iconet.png",
    alt: "Logo Iconnet",
    sourcePage: "local:/public/images/internet/logo_iconet.png",
  },
  iconnet: {
    src: "/images/internet/logo_iconet.png",
    alt: "Logo Iconnet",
    sourcePage: "local:/public/images/internet/logo_iconet.png",
  },
  indovision: {
    src: "/images/tv/logo_indovision.png",
    alt: "Logo Indovision",
    sourcePage: "local:/public/images/tv/logo_indovision.png",
  },
  "mnc play": {
    src: "/images/tv/log_mnc_play.png",
    alt: "Logo MNC Play",
    sourcePage: "local:/public/images/tv/log_mnc_play.png",
  },
  "my republik": {
    src: "/images/tv/logo_myrepublik.png",
    alt: "Logo My Republik",
    sourcePage: "local:/public/images/tv/logo_myrepublik.png",
  },
  "nex parabola": {
    src: "/images/tv/logo_nex.png",
    alt: "Logo Nex Parabola",
    sourcePage: "local:/public/images/tv/logo_nex.png",
  },
  nexparabola: {
    src: "/images/tv/logo_nex.png",
    alt: "Logo Nex Parabola",
    sourcePage: "local:/public/images/tv/logo_nex.png",
  },
  okevision: {
    src: "/images/tv/logo_okevision.png",
    alt: "Logo Okevision",
    sourcePage: "local:/public/images/tv/logo_okevision.png",
  },
  telkomvision: {
    src: "/images/tv/logo_telekom_vision.png",
    alt: "Logo Telkomvision",
    sourcePage: "local:/public/images/tv/logo_telekom_vision.png",
  },
  "top tv": {
    src: "/images/tv/logo_top_tv.png",
    alt: "Logo Top TV",
    sourcePage: "local:/public/images/tv/logo_top_tv.png",
  },
  transvision: {
    src: "/images/tv/logo_transvision.png",
    alt: "Logo Transvision",
    sourcePage: "local:/public/images/tv/logo_transvision.png",
  },
  "yes tv": {
    src: "/images/tv/logo_yestv.png",
    alt: "Logo YES TV",
    sourcePage: "local:/public/images/tv/logo_yestv.png",
  },
  "k-vision": {
    src: "/images/tv/logo_kvision.png",
    alt: "Logo K-Vision",
    sourcePage: "local:/public/images/tv/logo_kvision.png",
  },
  "k vision": {
    src: "/images/tv/logo_kvision.png",
    alt: "Logo K-Vision",
    sourcePage: "local:/public/images/tv/logo_kvision.png",
  },
  kvision: {
    src: "/images/tv/logo_kvision.png",
    alt: "Logo K-Vision",
    sourcePage: "local:/public/images/tv/logo_kvision.png",
  },
  pgn: {
    src: "/images/gas/Logo_PGN.png",
    alt: "Logo PGN",
    sourcePage: "local:/public/images/gas/Logo_PGN.png",
  },
  "free fire": {
    src: "/images/games/banner_freefire.png",
    alt: "Logo Free Fire",
    sourcePage: "local:/public/images/games/banner_freefire.png",
  },
  "pubg mobile": {
    src: "/images/games/banner_pubg.png",
    alt: "Logo PUBG Mobile",
    sourcePage: "local:/public/images/games/banner_pubg.png",
  },
  "mobile legend": {
    src: "/images/games/banner_mobile_legends.png",
    alt: "Mobile Legends: Bang Bang",
    sourcePage: "local:/public/images/games/banner_mobile_legends.png",
  },
  "mobile legends": {
    src: "/images/games/banner_mobile_legends.png",
    alt: "Mobile Legends: Bang Bang",
    sourcePage: "local:/public/images/games/banner_mobile_legends.png",
  },
  "mobile legends bang bang": {
    src: "/images/games/banner_mobile_legends.png",
    alt: "Mobile Legends: Bang Bang",
    sourcePage: "local:/public/images/games/banner_mobile_legends.png",
  },
  "magic chess go go": {
    src: "/images/games/banner_magic_chess.jpg",
    alt: "Magic Chess: Go Go",
    sourcePage: "local:/public/images/games/banner_magic_chess.jpg",
  },
  "free fire max": {
    src: "/images/games/banner_freefire_max.svg",
    alt: "Free Fire MAX Top-up",
    sourcePage: "local:/public/images/games/banner_freefire_max.svg",
  },
  "call of duty mobile": {
    src: "/images/games/codm_logo.png",
    alt: "Call of Duty Mobile",
    sourcePage: "local:/public/images/games/codm_logo.png",
  },
  hago: {
    src: "/images/games/banner_hago.svg",
    alt: "Hago",
    sourcePage: "local:/public/images/games/banner_hago.svg",
  },
  "genshin impact": {
    src: "/images/games/genshin_impact_logo_square_transparent.png",
    alt: "Genshin Impact",
    sourcePage: "local:/public/images/games/genshin_impact_logo_square_transparent.png",
  },
  zepeto: {
    src: "/images/games/banner_zepeto.svg",
    alt: "ZEPETO",
    sourcePage: "local:/public/images/games/banner_zepeto.svg",
  },
  "blood strike": {
    src: "/images/games/banner_blood_strike.svg",
    alt: "Blood Strike",
    sourcePage: "local:/public/images/games/banner_blood_strike.svg",
  },
  roblox: {
    src: "/images/games/banner_roblox.png",
    alt: "Logo Roblox",
    sourcePage: "local:/public/images/games/banner_roblox.png",
  },
  "point blank": {
    src: "/images/games/banner_point_blank.png",
    alt: "Logo Point Blank",
    sourcePage: "local:/public/images/games/banner_point_blank.png",
  },
};

function normalizeBrandName(name: string) {
  return name
    .trim()
    .toLowerCase()
    .replace(/\./g, "")
    .replace(/&/g, "and")
    .replace(/-/g, " ")
    .replace(/\s+/g, " ");
}

export function getBrandLogo(name: string): BrandLogoMeta | null {
  const key = normalizeBrandName(name);
  const withoutBankPrefix = key.replace(/^bank\s+/, "");
  return BRAND_LOGOS[key] ?? BRAND_LOGOS[withoutBankPrefix] ?? null;
}
