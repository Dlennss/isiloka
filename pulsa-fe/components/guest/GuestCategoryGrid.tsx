"use client";

import Link from "next/link";
import { useMemo } from "react";
import {
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
    return (activeItems.length > 0 ? activeItems : HOME_FALLBACK_ITEMS).slice(0, 4);
  }, [sortedItems]);

  return (
    <section>
      <div className="brand-categories rounded-[24px] border border-emerald-950/5 bg-linear-to-br from-white via-emerald-50/70 to-lime-50/80 p-2.5 shadow-[0_18px_38px_rgba(6,78,59,0.12)]">
        <div className={showAll ? "grid grid-cols-3 gap-2.5" : "grid grid-cols-3 gap-2.5"}>
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
              className="group flex min-h-[76px] flex-col items-center justify-center gap-2 rounded-2xl border border-white/70 bg-white/72 px-1.5 py-2 text-center shadow-[0_8px_22px_rgba(15,23,42,0.06)] ring-1 ring-emerald-950/[0.03] transition duration-300 hover:-translate-y-0.5 hover:bg-white hover:shadow-[0_14px_30px_rgba(6,78,59,0.13)]"
            >
              <div className="grid h-11 w-11 place-items-center rounded-2xl bg-linear-to-br from-[#13464b] via-[#087e8b] to-[#f6b49d] text-white shadow-[0_12px_26px_rgba(6,78,59,0.26)] ring-1 ring-white/40 transition-transform duration-300 group-hover:scale-105">
                <LayoutGrid className="h-4.5 w-4.5" strokeWidth={2.2} />
              </div>
              <span className="line-clamp-2 px-1 text-[10px] font-black leading-tight text-[#13464b]">
                Lainnya
              </span>
            </Link>
          ) : null}
        </div>
      </div>
    </section>
  );
}
