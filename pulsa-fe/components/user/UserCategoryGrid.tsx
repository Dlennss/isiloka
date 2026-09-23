"use client";

import { useMemo } from "react";
import type { UserCategoryItem } from "@/components/user/types";
import { CategoryShortcutLink } from "@/components/shared/CategoryShortcutLink";
import { getGuestCategoryPath } from "@/lib/category-routes";

type UserCategoryGridProps = {
  items: UserCategoryItem[];
  showAll?: boolean;
};

type CategoryCardProps = {
  item: UserCategoryItem;
};

const DEFAULT_SHORTCUTS = [
  { href: "/user/pulsa-data", label: "Pulsa & Data", visualName: "pulsa data" },
  { href: "/game", label: "Game", visualName: "game" },
  { href: "/user/ewallet", label: "E-Wallet", visualName: "e-money" },
  { href: "/user/listrik", label: "PLN", visualName: "pln" },
  { href: "/user/kategori", label: "Lainnya", visualName: "lainnya" },
];

const PRIORITY: Record<string, number> = {
  pulsa: 1,
  "paket data": 1,
  game: 2,
  "e-money": 3,
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

function normalizeName(name: string) {
  return name.trim().toLowerCase();
}

function getCategoryHref(item: UserCategoryItem) {
  const name = normalizeName(item.nama);
  if (name === "pulsa" || name === "paket data") return "/user/pulsa-data";
  return getGuestCategoryPath(item).replace(/^\/kategori\//, "/user/kategori/");
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
  if (name === "e-money") {
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

export function UserCategoryGrid({ items, showAll = false }: UserCategoryGridProps) {
  const sortedItems = useMemo(() => sortCategories(items), [items]);

  return (
    <section>
      <div className="rounded-[24px] border border-emerald-950/5 bg-linear-to-br from-white via-emerald-50/70 to-lime-50/80 p-2.5 shadow-[0_18px_38px_rgba(6,78,59,0.12)]">
        <div className={showAll ? "grid grid-cols-3 gap-2.5" : "grid grid-cols-3 gap-2.5"}>
          {!showAll
            ? DEFAULT_SHORTCUTS.map((item) => (
                <CategoryShortcutLink
                  key={item.href}
                  href={item.href}
                  label={item.label}
                  visualName={item.visualName}
                />
              ))
            : null}
          {showAll
            ? sortedItems.map((item) => (
                <CategoryCard key={item.id} item={item} />
              ))
            : null}
        </div>
      </div>
    </section>
  );
}
