import fs from "node:fs";
import path from "node:path";

const root = "/home/syarif/app/pulsakilat-fe";
const inputPath = path.join(root, "scripts", "yuscom-display-brands.txt");
const outputDir = path.join(root, "public", "yuscom-display-brand-logos-generated");
const manifestPath = path.join(root, "lib", "generated-yuscom-display-brand-map.json");

function slugify(value) {
  return value
    .normalize("NFKD")
    .replace(/[\u0300-\u036f]/g, "")
    .replace(/&/g, " and ")
    .replace(/[^a-zA-Z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .toLowerCase();
}

function initialsOf(value) {
  const words = value
    .replace(/[()]/g, " ")
    .split(/[^a-zA-Z0-9]+/)
    .filter(Boolean);
  if (words.length === 0) return "BR";
  if (words.length === 1) return words[0].slice(0, 3).toUpperCase();
  return words.slice(0, 3).map((word) => word[0].toUpperCase()).join("");
}

function paletteFor(value) {
  const upper = value.toUpperCase();
  if (upper.startsWith("PDAM") || upper.startsWith("PAM") || upper.startsWith("PERUMDA") || upper.startsWith("PERUMDAM") || upper.startsWith("PT AIR")) {
    return { bg1: "#E0F2FE", bg2: "#BAE6FD", fg: "#0369A1", chip: "#0EA5E9", icon: "water" };
  }
  if (upper.includes("TV") || upper.includes("VISION") || upper.includes("PARABOLA")) {
    return { bg1: "#F3E8FF", bg2: "#E9D5FF", fg: "#6B21A8", chip: "#A855F7", icon: "tv" };
  }
  if (upper.includes("GAME") || upper.includes("FIRE") || upper.includes("PUBG") || upper.includes("ROBLOX") || upper.includes("LEGEND") || upper.includes("POINT BLANK")) {
    return { bg1: "#E2E8F0", bg2: "#CBD5E1", fg: "#0F172A", chip: "#475569", icon: "game" };
  }
  if (upper.includes("NET") || upper.includes("TELKOM") || upper.includes("INDIHOME") || upper.includes("BIZZNET") || upper.includes("ICONNET")) {
    return { bg1: "#DCFCE7", bg2: "#BBF7D0", fg: "#166534", chip: "#22C55E", icon: "signal" };
  }
  if (upper.includes("PLN") || upper.includes("PGN") || upper.includes("BPJS")) {
    return { bg1: "#FEF3C7", bg2: "#FDE68A", fg: "#92400E", chip: "#F59E0B", icon: "bill" };
  }
  return { bg1: "#F8FAFC", bg2: "#E2E8F0", fg: "#334155", chip: "#64748B", icon: "brand" };
}

function iconSvg(kind, color) {
  if (kind === "water") {
    return `<path d="M30 14c6.5 7.5 11 13.5 11 18.6C41 39.4 36 44 30 44s-11-4.6-11-11.4C19 27.5 23.5 21.5 30 14Z" fill="${color}" opacity="0.18"/><path d="M30 18c4.9 5.7 8 10 8 14.6 0 5-3.6 8.4-8 8.4s-8-3.4-8-8.4c0-4.6 3.1-8.9 8-14.6Z" fill="${color}"/>`;
  }
  if (kind === "tv") {
    return `<rect x="16" y="18" width="28" height="18" rx="4" fill="${color}" opacity="0.18"/><rect x="18" y="20" width="24" height="14" rx="3" fill="${color}"/><rect x="27" y="37" width="6" height="2.5" rx="1.25" fill="${color}"/>`;
  }
  if (kind === "game") {
    return `<path d="M20 28c0-4.4 3.6-8 8-8h4c4.4 0 8 3.6 8 8 0 5.5-4.5 10-10 10h-0c-5.5 0-10-4.5-10-10Z" fill="${color}" opacity="0.18"/><path d="M23 30h4v-2h2v-4h-2v-2h-4v2h-2v4h2v2Zm13-4.5a2 2 0 1 0 0 .01Zm-3 5a2 2 0 1 0 0 .01Z" fill="${color}"/>`;
  }
  if (kind === "signal") {
    return `<path d="M18 38c7-10 17-16 24-19l2 3c-6 2.6-15.1 8.3-21.6 17.6L18 38Z" fill="${color}" opacity="0.2"/><path d="M25 40c4.7-6.6 10.5-10.6 15.4-13.1l2 3c-4.4 2.2-9.4 5.7-13.5 11.5L25 40Zm7-2c2.1-2.8 4.5-4.7 7-6.1l2 3c-2 .9-3.9 2.4-5.7 4.8L32 38Z" fill="${color}"/>`;
  }
  if (kind === "bill") {
    return `<rect x="18" y="16" width="24" height="28" rx="4" fill="${color}" opacity="0.15"/><rect x="22" y="21" width="16" height="3" rx="1.5" fill="${color}"/><rect x="22" y="27" width="12" height="3" rx="1.5" fill="${color}"/><rect x="22" y="33" width="10" height="3" rx="1.5" fill="${color}"/>`;
  }
  return `<circle cx="30" cy="24" r="10" fill="${color}" opacity="0.18"/><path d="M21 37c2.8-4.2 6.3-6 9-6s6.2 1.8 9 6" fill="none" stroke="${color}" stroke-width="3" stroke-linecap="round"/>`;
}

function makeSvg(label, palette) {
  const initials = initialsOf(label);
  const safeLabel = label.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="160" height="160" viewBox="0 0 160 160" role="img" aria-label="${safeLabel}">
  <defs>
    <linearGradient id="bg" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0%" stop-color="${palette.bg1}"/>
      <stop offset="100%" stop-color="${palette.bg2}"/>
    </linearGradient>
  </defs>
  <rect width="160" height="160" rx="32" fill="url(#bg)"/>
  <rect x="18" y="18" width="124" height="124" rx="28" fill="#ffffff" opacity="0.9"/>
  <g transform="translate(50 28)">
    ${iconSvg(palette.icon, palette.fg)}
  </g>
  <rect x="28" y="104" width="104" height="28" rx="14" fill="${palette.chip}" opacity="0.14"/>
  <text x="80" y="122" text-anchor="middle" font-family="Arial, Helvetica, sans-serif" font-size="26" font-weight="700" fill="${palette.fg}">${initials}</text>
</svg>`;
}

const brands = fs
  .readFileSync(inputPath, "utf8")
  .split(/\r?\n/)
  .map((line) => line.trim())
  .filter(Boolean);

fs.mkdirSync(outputDir, { recursive: true });

const usedFileNames = new Map();
const manifest = {};

for (const brand of brands) {
  const base = slugify(brand) || "brand";
  const seq = usedFileNames.get(base) ?? 0;
  usedFileNames.set(base, seq + 1);
  const fileName = seq === 0 ? `${base}.svg` : `${base}-${seq + 1}.svg`;
  const filePath = path.join(outputDir, fileName);
  fs.writeFileSync(filePath, makeSvg(brand, paletteFor(brand)));
  manifest[brand] = fileName;
}

fs.writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);

console.log(`generated=${brands.length} unique_files=${usedFileNames.size}`);
