"use client";

import * as React from "react";
import type { UserProductItem } from "@/components/user/types";
import { GuestProductGrid } from "@/components/guest/GuestProductGrid";
import { getProductGroupLabel } from "@/lib/product-grouping";

type GuestEwalletProductSectionProps = {
  items: UserProductItem[];
  isLoggedIn: boolean;
};

export function GuestEwalletProductSection({ items, isLoggedIn }: GuestEwalletProductSectionProps) {
  const groupedProducts = React.useMemo(() => {
    const grouped = new Map<string, UserProductItem[]>();
    for (const item of items) {
      const label = getProductGroupLabel(item);
      const bucket = grouped.get(label) || [];
      bucket.push(item);
      grouped.set(label, bucket);
    }

    return Array.from(grouped.entries())
      .map(([label, groupItems]) => ({
        label,
        items: [...groupItems].sort((a, b) => Number(a.nominal || 0) - Number(b.nominal || 0)),
      }))
      .sort((a, b) => a.label.localeCompare(b.label));
  }, [items]);

  const [selectedGroup, setSelectedGroup] = React.useState("");

  React.useEffect(() => {
    if (!groupedProducts.length) {
      setSelectedGroup("");
      return;
    }

    setSelectedGroup((current) => {
      if (current && groupedProducts.some((group) => group.label === current)) return current;
      return groupedProducts[0].label;
    });
  }, [groupedProducts]);

  const activeItems = React.useMemo(() => {
    if (!groupedProducts.length) return [];
    return groupedProducts.find((group) => group.label === selectedGroup)?.items || groupedProducts[0].items;
  }, [groupedProducts, selectedGroup]);

  return (
    <div className="space-y-3">
      {groupedProducts.length > 1 ? (
        <div className="flex snap-x snap-mandatory gap-2 overflow-x-auto pb-1 [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
          {groupedProducts.map((group) => (
            <button
              key={group.label}
              type="button"
              onClick={() => setSelectedGroup(group.label)}
              className={`shrink-0 snap-start whitespace-nowrap rounded-full px-3 py-2 text-xs font-semibold transition ${
                selectedGroup === group.label
                  ? "bg-sky-600 text-white shadow-[0_8px_18px_rgba(15,111,203,0.22)]"
                  : "bg-white text-slate-700 ring-1 ring-slate-200 hover:bg-slate-50"
              }`}
            >
              {group.label}
            </button>
          ))}
        </div>
      ) : null}

      <GuestProductGrid items={activeItems} isLoggedIn={isLoggedIn} layout="list" />
    </div>
  );
}
