import type { LucideIcon } from "lucide-react";
import Image from "next/image";

type PurchaseBrandBannerProps = {
  title: string;
  logoSrc?: string | null;
  logoAlt?: string;
  fallbackText?: string;
  icon?: LucideIcon;
};

export function PurchaseBrandBanner({ title, logoSrc, logoAlt, fallbackText, icon: Icon }: PurchaseBrandBannerProps) {
  const initials = String(fallbackText || title || "")
    .trim()
    .replace(/\s+/g, " ")
    .slice(0, 2)
    .toUpperCase();

  return (
    <div className="bg-linear-to-r from-sky-50 via-white to-cyan-50 px-4 py-4">
      <div className="flex items-center gap-3">
        <div className="grid h-11 w-11 place-items-center overflow-hidden rounded-2xl bg-white text-sky-600 shadow-[0_8px_20px_rgba(14,165,233,0.16)] ring-1 ring-sky-100">
          {logoSrc ? (
            <Image src={logoSrc} alt={logoAlt || title} width={44} height={44} className="h-7 w-7 object-contain" />
          ) : Icon ? (
            <Icon className="h-5 w-5" />
          ) : (
            <span className="text-[10px] font-black uppercase text-sky-700">{initials || "PJ"}</span>
          )}
        </div>
        <div className="min-w-0">
          <h2 className="truncate text-[15px] font-bold tracking-tight text-slate-900">{title}</h2>
        </div>
      </div>
    </div>
  );
}
