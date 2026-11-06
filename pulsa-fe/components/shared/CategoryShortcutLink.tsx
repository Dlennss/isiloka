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
    case "voucher data":
      return { iconSrc: "/service-icons/paket-data.png" };
    case "listrik":
    case "pln":
      return { iconSrc: "/service-icons/token-pln.png" };
    case "game":
      return { iconSrc: "/service-icons/voucher-game.png" };
    case "tv":
    case "tv & streaming":
      return { iconSrc: "/service-icons/tv-kabel.png" };
    case "pdam":
    case "tagihan air":
      return { iconSrc: "/service-icons/pdam.png" };
    case "bpjs":
      return { iconSrc: "/service-icons/bpjs.png" };
    case "internet pascabayar":
    case "internet & telco":
      return { iconSrc: "/service-icons/internet-wifi.png" };
    case "hp pascabayar":
    case "pascabayar":
      return { iconSrc: "/service-icons/hp-pascabayar.png" };
    case "masa aktif":
      return { iconSrc: "/service-icons/pulsa.png" };
    case "paket telepon":
    case "telepon & sms":
      return { iconSrc: "/service-icons/hp-pascabayar.png" };
    case "aktivasi perdana":
      return { iconSrc: "/service-icons/esim-roaming.png" };
    case "gas negara":
    case "tagihan gas":
      return { iconSrc: "/service-icons/gas-pgn.png" };
    case "transfer bank":
    case "bank transfer":
      return { iconSrc: "/service-icons/transfer-bank.png" };
    case "asuransi":
      return { iconSrc: "/service-icons/asuransi.png" };
    case "multifinance":
      return { iconSrc: "/service-icons/cicilan-multifinance.png" };
    case "pajak daerah":
    case "samsat":
      return { iconSrc: "/service-icons/pbb.png" };
    case "voucher digital":
      return { iconSrc: "/service-icons/voucher-digital.png" };
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
      className="group flex min-h-[72px] flex-col items-center justify-start gap-2 text-center transition duration-300 hover:-translate-y-0.5"
    >
      <div className="relative grid h-12 w-12 place-items-center overflow-hidden rounded-[16px] bg-slate-50 shadow-[inset_0_0_0_1px_rgba(15,23,42,0.04)] transition-transform duration-300 group-hover:scale-105">
        <Image
          src={visual.iconSrc}
          alt=""
          fill
          sizes="48px"
          className="object-contain p-0.5"
        />
      </div>
      <span className="min-h-10 break-words px-0.5 text-xs font-semibold leading-5 text-slate-950">
        {label}
      </span>
    </Link>
  );
}
