"use client";

import * as React from "react";
import Image from "next/image";
import Link from "next/link";
import { Search } from "lucide-react";
import type { UserBrandItem } from "@/components/user/types";
import { getDedicatedGuestBrandPath } from "@/lib/dedicated-category-brand-routes";

type PdamBrandSearchProps = {
  brands: UserBrandItem[];
};

export function PdamBrandSearch({ brands }: PdamBrandSearchProps) {
  const [query, setQuery] = React.useState("");

  const filtered = React.useMemo(() => {
    const keyword = query.trim().toLowerCase();
    if (!keyword) return brands;
    return brands.filter((brand) => brand.nama.toLowerCase().includes(keyword));
  }, [brands, query]);

  return (
    <div className="space-y-4">
      <label className="block space-y-2">
        <span className="text-sm font-semibold text-slate-700">Cari PDAM</span>
        <div className="flex items-center gap-3 border border-slate-200 bg-white px-4 py-3 shadow-[0_8px_22px_rgba(15,23,42,0.08)]">
          <Search className="h-4 w-4 text-slate-400" />
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Cari nama PDAM"
            className="w-full bg-transparent text-sm text-slate-900 outline-none"
          />
        </div>
      </label>

      <div className="space-y-2">
        {filtered.length === 0 ? (
          <div className="grid min-h-28 place-items-center border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500">
            Tidak ada PDAM yang cocok.
          </div>
        ) : (
          filtered.map((brand) => {
            const href = getDedicatedGuestBrandPath("17", brand) || `/kategori/17/brand/${brand.id}?name=${encodeURIComponent(brand.nama)}`;
            return (
              <Link
                key={brand.id}
                href={href}
                className="flex items-center gap-3 border border-slate-200 bg-white px-4 py-3 text-left shadow-[0_8px_22px_rgba(15,23,42,0.08)] transition hover:border-sky-300 hover:shadow-[0_12px_26px_rgba(15,23,42,0.12)]"
              >
                <div className="grid h-12 w-12 shrink-0 place-items-center overflow-hidden">
                  <Image
                    src="/images/pdam/logo_pdam.png"
                    alt="PDAM"
                    title="PDAM"
                    width={48}
                    height={48}
                    className="h-full w-full object-contain"
                  />
                </div>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-semibold text-slate-900">{brand.nama}</p>
                </div>
              </Link>
            );
          })
        )}
      </div>
    </div>
  );
}
