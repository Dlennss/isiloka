"use client";

import Image from "next/image";
import Link from "next/link";

type CategoryVisual = {
  iconSrc: string;
};

type CategoryShortcutLinkProps = {
  href: string;
  label: string;
  visualName: string;
};

function normalizeName(name: string) {
  return name.trim().toLowerCase();
}

function getCategoryVisual(name: string): CategoryVisual {
  const value = normalizeName(name);

  switch (value) {
    case "pulsa":
    case "pulsa data":
    case "pulsa & data":
      return { iconSrc: "/service-icons/pulsa.png" };
    case "e-money":
    case "e-wallet":
      return { iconSrc: "/service-icons/ewallet.png" };
    case "paket data":
      return { iconSrc: "/service-icons/paket-data.png" };
    case "listrik":
    case "pln":
      return { iconSrc: "/service-icons/token-pln.png" };
    case "game":
      return { iconSrc: "/service-icons/voucher-game.png" };
    case "tv":
      return { iconSrc: "/service-icons/tv-kabel.png" };
    case "pdam":
      return { iconSrc: "/service-icons/pdam.png" };
    case "bpjs":
      return { iconSrc: "/service-icons/bpjs.png" };
    case "internet pascabayar":
      return { iconSrc: "/service-icons/internet-wifi.png" };
    case "hp pascabayar":
      return { iconSrc: "/service-icons/hp-pascabayar.png" };
    case "masa aktif":
      return { iconSrc: "/service-icons/pulsa.png" };
    case "paket telepon":
      return { iconSrc: "/service-icons/hp-pascabayar.png" };
    case "aktivasi perdana":
      return { iconSrc: "/service-icons/esim-roaming.png" };
    case "gas negara":
      return { iconSrc: "/service-icons/gas-pgn.png" };
    case "lainnya":
      return { iconSrc: "/service-icons/lainnya.png" };
    default:
      return { iconSrc: "/service-icons/lainnya.png" };
  }
}

export function CategoryShortcutLink({ href, label, visualName }: CategoryShortcutLinkProps) {
  const visual = getCategoryVisual(visualName);

  return (
    <Link
      href={href}
      prefetch={false}
      aria-label={label}
      className="group flex min-h-[76px] flex-col items-center justify-center gap-2 rounded-2xl border border-white/70 bg-white/72 px-1.5 py-2 text-center shadow-[0_8px_22px_rgba(15,23,42,0.06)] ring-1 ring-emerald-950/[0.03] transition duration-300 hover:-translate-y-0.5 hover:bg-white hover:shadow-[0_14px_30px_rgba(6,78,59,0.13)]"
    >
      <div className="relative grid h-12 w-12 place-items-center overflow-hidden rounded-2xl bg-white shadow-[0_10px_22px_rgba(6,78,59,0.12)] ring-1 ring-slate-200/80 transition-transform duration-300 group-hover:scale-105">
        <Image
          src={visual.iconSrc}
          alt=""
          fill
          sizes="48px"
          className="object-contain"
        />
      </div>
      <span className="line-clamp-2 px-1 text-[10px] font-black leading-tight text-[#13464b]">
        {label}
      </span>
    </Link>
  );
}
