"use client";

import Link from "next/link";
import { useMemo } from "react";
import {
  ChevronRight,
  LayoutGrid,
} from "lucide-react";
import type { UserCategoryItem } from "@/components/user/types";
import { getGuestCategoryPath } from "@/lib/category-routes";
import { CategoryShortcutLink } from "@/components/shared/CategoryShortcutLink";

type GuestCategoryGridProps = {
  items: UserCategoryItem[];
  showAll?: boolean;
};

type CategoryCardProps = {
  item: UserCategoryItem;
};

const PRIORITY: Record<string, number> = {
  pulsa: 1,
  "paket data": 1,
  game: 2,
  "e-money": 3,
  "e-wallet": 3,
  listrik: 4,
  pln: 4,
  tv: 6,
  pdam: 7,
  bpjs: 8,
  "internet pascabayar": 9,
  "hp pascabayar": 10,
  "masa aktif": 11,
  "paket telepon": 12,
  "aktivasi perdana": 13,
  "gas negara": 14,
  lainnya: 15,
};

const HOME_FALLBACK_ITEMS: UserCategoryItem[] = [
  { id: 1, nama: "Pulsa", aktif: true },
];

function normalizeName(name: string) {
  return name.trim().toLowerCase();
}

function getCategoryHref(item: UserCategoryItem) {
  const name = normalizeName(item.nama);
  if (name === "pulsa" || name === "paket data") return "/pulsa-data";
  return getGuestCategoryPath(item);
}

function sortCategories(items: UserCategoryItem[]) {
  return items
  .filter((item) => normalizeName(item.nama) !== "paket data")
  .sort((a, b) => {
    const aKey = normalizeName(a.nama);
    const bKey = normalizeName(b.nama);
    const pa = PRIORITY[aKey] ?? 999;
    const pb = PRIORITY[bKey] ?? 999;
    if (pa !== pb) return pa - pb;
    return a.nama.localeCompare(b.nama, "id-ID");
  });
}

function getCategoryLabel(item: UserCategoryItem) {
  const name = normalizeName(item.nama);
  if (name === "pulsa") {
    return "Pulsa & Data";
  }
  if (name === "e-money" || name === "e-wallet") {
    return "E-Wallet";
  }
  return item.nama;
}

function getCategoryVisualName(item: UserCategoryItem) {
  return normalizeName(item.nama) === "pulsa" ? "pulsa data" : item.nama;
}

function CategoryCard({ item }: CategoryCardProps) {
  const label = getCategoryLabel(item);

  return (
    <CategoryShortcutLink href={getCategoryHref(item)} label={label} visualName={getCategoryVisualName(item)} />
  );
}

export function GuestCategoryGrid({ items, showAll = false }: GuestCategoryGridProps) {
  const sortedItems = useMemo(() => sortCategories(items), [items]);
  const homeItems = useMemo(() => {
    const activeItems = sortedItems.filter((item) => item.aktif);
    return (activeItems.length > 0 ? activeItems : HOME_FALLBACK_ITEMS).slice(0, 9);
  }, [sortedItems]);

  return (
    <section>
      <div className="brand-categories overflow-hidden rounded-[18px] border border-sky-100 bg-white px-4 pb-4 pt-4 shadow-[0_16px_38px_rgba(15,56,104,0.08)]">
        <div className="mb-4 flex items-center justify-between gap-3">
          <h2 className="text-[17px] font-black leading-none text-slate-950">Semua Layanan</h2>
          {!showAll ? (
            <Link
              href="/kategori"
              prefetch={false}
              className="inline-flex h-8 shrink-0 items-center gap-1 rounded-full px-2 text-xs font-black text-[#06489f] transition hover:bg-sky-50"
            >
              Lihat Semua
              <ChevronRight className="h-4 w-4" strokeWidth={2.8} />
            </Link>
          ) : null}
        </div>
        <div className={showAll ? "grid grid-cols-4 gap-x-3 gap-y-5" : "grid grid-cols-5 gap-x-2 gap-y-5"}>
          {!showAll ? homeItems.map((item) => (
            <CategoryCard key={item.id} item={item} />
          )) : null}
          {showAll ? sortedItems.map((item) => (
            <CategoryCard key={item.id} item={item} />
          )) : null}
          {!showAll ? (
            <Link
              href="/kategori"
              prefetch={false}
              aria-label="Lainnya"
              className="group flex min-h-[72px] flex-col items-center justify-start gap-2 text-center transition duration-300 hover:-translate-y-0.5"
            >
              <div className="grid h-12 w-12 place-items-center rounded-[16px] bg-slate-100 text-slate-500 shadow-[inset_0_0_0_1px_rgba(15,23,42,0.04)] transition-transform duration-300 group-hover:scale-105">
                <LayoutGrid className="h-5 w-5" strokeWidth={2.4} />
              </div>
              <span className="line-clamp-2 min-h-[24px] px-0.5 text-[10px] font-black leading-tight text-slate-950">
                Lainnya
              </span>
            </Link>
          ) : null}
        </div>
      </div>
    </section>
  );
}
