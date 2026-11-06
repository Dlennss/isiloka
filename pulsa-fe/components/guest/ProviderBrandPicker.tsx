"use client";

import Image from "next/image";
import Link from "next/link";
import { getBrandLogo } from "@/lib/brand-logos";
import type { UserBrandItem } from "@/components/user/types";

type ProviderBrandPickerProps = {
  items: Array<{
    brand: UserBrandItem;
    href: string;
  }>;
  layout?: "grid" | "list";
  columns?: 2 | 3;
};

export function ProviderBrandPicker({ items, layout = "grid", columns = 3 }: ProviderBrandPickerProps) {
  return (
    <section className={layout === "list" ? "space-y-2" : columns === 2 ? "grid grid-cols-2 gap-3" : "grid grid-cols-3 gap-3"}>
      {items.map(({ brand, href }) => {
        const logo = getBrandLogo(brand.nama);
        return (
          <Link
            key={brand.id}
            href={href}
            aria-label={brand.nama}
            className={
              layout === "list"
                ? "group flex items-center gap-3 border border-slate-200 bg-white px-3 py-3 shadow-[0_8px_18px_rgba(15,23,42,0.06)] transition-transform duration-300 hover:-translate-y-1 hover:border-sky-300 hover:shadow-[0_10px_24px_rgba(15,111,203,0.08)]"
                : "group rounded-md border border-slate-200 bg-white px-2 py-3 text-center shadow-[0_8px_18px_rgba(15,23,42,0.06)] transition-transform duration-300 hover:-translate-y-1 hover:border-sky-300 hover:shadow-[0_10px_24px_rgba(15,111,203,0.08)]"
            }
          >
            <div className={layout === "list" ? "grid h-12 w-12 shrink-0 place-items-center overflow-hidden" : "flex flex-col items-center gap-3"}>
              <div className={layout === "list" ? "grid h-12 w-12 place-items-center overflow-hidden" : "grid h-14 w-14 place-items-center overflow-hidden"}>
                {logo ? (
                  <Image src={logo.src} alt={logo.alt} title={logo.alt} width={56} height={56} className="h-full w-full object-contain" />
                ) : (
                  <span aria-hidden="true" className="text-base font-bold uppercase tracking-normal text-sky-700">{brand.nama.slice(0, 2)}</span>
                )}
              </div>
            </div>
            <span className={layout === "list" ? "min-w-0 break-words text-sm font-semibold text-slate-900" : "mt-2 flex min-h-10 items-center justify-center break-words text-xs font-semibold leading-5 text-slate-900"}>{brand.nama}</span>
          </Link>
        );
      })}
    </section>
  );
}
