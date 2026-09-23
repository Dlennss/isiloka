"use client";

import * as React from "react";
import Image from "next/image";
import { Sparkles, Wallet } from "lucide-react";
import { UserProductGrid } from "@/components/user/UserProductGrid";
import type { UserProductItem } from "@/components/user/types";
import { getBrandLogo } from "@/lib/brand-logos";

type EMoneyBrandFlowProps = {
  items: UserProductItem[];
  isLoggedIn: boolean;
  authToken?: string;
  mode: "guest" | "user";
  buyerRole?: string;
  initialDest?: string;
};

type EMoneyBucket = {
  key: string;
  label: string;
  items: UserProductItem[];
};

const SUBCATEGORY_PRIORITY = [
  "NON ADMIN DIRECT",
  "NON ADMIN",
  "DIRECT",
  "REGULER",
  "DRIVER",
  "PROMO",
  "ADMIN",
  "LAINNYA",
  "BEBAS NOMINAL",
] as const;

function formatBucketLabel(key: string) {
  return key.replace(/\s+/g, " ").trim();
}

function isCheckAccountProduct(item: UserProductItem) {
  const upperName = String(item.nama || "").toUpperCase();
  const upperSku = String(item.sku || "").toUpperCase();
  return upperSku.startsWith("CEK") || upperName.includes("CEK NAMA AKUN") || upperName.includes("CEK AKUN");
}

function extractNominalSortValue(item: UserProductItem) {
  const upper = String(item.nama || "").toUpperCase();
  const dotted = upper.match(/(\d{1,3}(?:\.\d{3})+)/);
  if (dotted) {
    return Number.parseInt(dotted[1].replaceAll(".", ""), 10);
  }

  const compact = upper.match(/(\d+)\s*K\b/);
  if (compact) {
    return Number.parseInt(compact[1], 10) * 1000;
  }

  return Number(item.nominal || item.harga_dasar_app || 0);
}

function classifyEMoneySubcategory(item: UserProductItem) {
  const backendGroup = String(item.group_name || "").trim();
  if (backendGroup) {
    return backendGroup;
  }

  const upper = String(item.nama || "").toUpperCase();

  if (item.tipe_harga === "OPEN_AMOUNT" || upper.includes("BEBAS NOMINAL")) {
    return "BEBAS NOMINAL";
  }
  if (upper.includes("NON ADMIN") && upper.includes("DIRECT")) {
    return "NON ADMIN DIRECT";
  }
  if (upper.includes("NON ADMIN")) {
    return "NON ADMIN";
  }
  if (upper.includes("DIRECT")) {
    return "DIRECT";
  }
  if (upper.includes("DRIVER")) {
    return "DRIVER";
  }
  if (upper.includes("PROMO")) {
    return "PROMO";
  }
  if (upper.includes("ADMIN")) {
    return "ADMIN";
  }
  return "REGULER";
}

function sortBuckets(a: EMoneyBucket, b: EMoneyBucket) {
  const aIsOpenAmount = a.key.toUpperCase().includes("BEBAS NOMINAL");
  const bIsOpenAmount = b.key.toUpperCase().includes("BEBAS NOMINAL");
  if (aIsOpenAmount !== bIsOpenAmount) {
    return aIsOpenAmount ? 1 : -1;
  }

  const aIndex = SUBCATEGORY_PRIORITY.indexOf(a.key as (typeof SUBCATEGORY_PRIORITY)[number]);
  const bIndex = SUBCATEGORY_PRIORITY.indexOf(b.key as (typeof SUBCATEGORY_PRIORITY)[number]);
  const aRank = aIndex === -1 ? SUBCATEGORY_PRIORITY.length : aIndex;
  const bRank = bIndex === -1 ? SUBCATEGORY_PRIORITY.length : bIndex;
  if (aRank !== bRank) return aRank - bRank;
  return a.label.localeCompare(b.label);
}

function buildBuckets(items: UserProductItem[]) {
  const visibleItems = items.filter((item) => !isCheckAccountProduct(item));
  const buckets = new Map<string, UserProductItem[]>();

  for (const item of visibleItems) {
    const key = classifyEMoneySubcategory(item);
    const current = buckets.get(key) ?? [];
    current.push(item);
    buckets.set(key, current);
  }

  return Array.from(buckets.entries())
    .map(([key, bucketItems]) => ({
      key,
      label: formatBucketLabel(key),
      items: [...bucketItems].sort((a, b) => {
        if (a.tipe_harga !== b.tipe_harga) return a.tipe_harga === "FIXED" ? -1 : 1;
        const nominalDiff = extractNominalSortValue(a) - extractNominalSortValue(b);
        if (nominalDiff !== 0) return nominalDiff;
        return a.nama.localeCompare(b.nama);
      }),
    }))
    .sort(sortBuckets);
}

export function EMoneyBrandFlow({ items, isLoggedIn, authToken, buyerRole, initialDest = "" }: EMoneyBrandFlowProps) {
  const buckets = React.useMemo(() => buildBuckets(items), [items]);
  const defaultBucketKey = React.useMemo(
    () => buckets.find((bucket) => bucket.key !== "BEBAS NOMINAL")?.key ?? buckets[0]?.key ?? "",
    [buckets],
  );
  const [selectedKey, setSelectedKey] = React.useState<string>(defaultBucketKey);

  React.useEffect(() => {
    const hasSelected = buckets.some((bucket) => bucket.key === selectedKey);
    if (!hasSelected) {
      setSelectedKey(defaultBucketKey);
    }
  }, [buckets, defaultBucketKey, selectedKey]);

  const selectedBucket = buckets.find((bucket) => bucket.key === selectedKey) ?? buckets[0] ?? null;
  const brandName = String(items[0]?.brand_nama || "").trim();
  const brandLogo = brandName ? getBrandLogo(brandName) : null;

  if (!selectedBucket) {
    return (
      <div className="grid min-h-40 place-items-center rounded-3xl border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_8px_22px_rgba(15,23,42,0.13)]">
        Belum ada produk aktif untuk brand ini.
      </div>
    );
  }

  return (
    <div className="mt-2 space-y-4">
      {brandName ? (
        <section className="overflow-hidden rounded-[26px] border border-emerald-950/10 bg-linear-to-br from-[#13464b] via-[#087e8b] to-[#b8f138] p-4 text-white shadow-[0_18px_42px_rgba(5,46,38,0.22)]">
          <div className="flex items-center gap-4">
            <div className="grid h-18 w-18 shrink-0 place-items-center rounded-[24px] bg-white p-3 shadow-[0_14px_30px_rgba(5,46,38,0.18)]">
              {brandLogo?.src ? (
                <Image
                  src={brandLogo.src}
                  alt={brandLogo.alt || brandName}
                  width={96}
                  height={96}
                  className="h-full w-full object-contain"
                />
              ) : (
                <Wallet className="h-8 w-8 text-[#087e8b]" />
              )}
            </div>
            <div className="min-w-0 flex-1">
              <p className="text-[10px] font-black uppercase tracking-[0.2em] text-lime-200">Isiloka E-Wallet</p>
              <h1 className="mt-1 text-2xl font-black tracking-tight">{brandName}</h1>
              <p className="mt-1 text-xs font-semibold text-white/78">Pilih nominal top up, lalu masukkan nomor e-wallet saat checkout.</p>
            </div>
            <div className="hidden h-12 w-12 shrink-0 place-items-center rounded-2xl bg-white/18 text-lime-100 ring-1 ring-white/20 min-[380px]:grid">
              <Sparkles className="h-5 w-5" />
            </div>
          </div>
        </section>
      ) : null}

      {buckets.length > 1 ? (
        <div className="flex snap-x snap-mandatory gap-2 overflow-x-auto pb-1 [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
          {buckets.map((bucket) => {
            const active = bucket.key === selectedBucket.key;
            return (
              <button
                key={bucket.key}
                type="button"
                onClick={() => setSelectedKey(bucket.key)}
                className={`h-9 shrink-0 snap-start whitespace-nowrap rounded-full px-3 text-[10px] font-black leading-tight transition ${
                  active
                    ? "bg-[#13464b] text-white shadow-[0_10px_20px_rgba(5,46,38,0.22)]"
                    : "bg-white text-slate-700 ring-1 ring-emerald-950/10 hover:bg-emerald-50"
                }`}
              >
                {bucket.label}
              </button>
            );
          })}
        </div>
      ) : null}

      <UserProductGrid
        items={selectedBucket.items}
        isLoggedIn={isLoggedIn}
        authToken={authToken}
        buyerRole={buyerRole}
        initialDest={initialDest}
        buyLabel="Top Up"
        enableGuestHint={false}
      />
    </div>
  );
}
