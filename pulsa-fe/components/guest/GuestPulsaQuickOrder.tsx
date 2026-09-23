"use client";

import * as React from "react";
import { LoaderCircle } from "lucide-react";
import type { UserBrandItem, UserProductItem } from "@/components/user/types";
import { UserCheckoutModal } from "@/components/user/UserCheckoutModal";
import { QuickProductOptionGrid } from "@/components/guest/QuickProductOptionGrid";
import { ProviderBrandPicker } from "@/components/guest/ProviderBrandPicker";
import { PulsaEntryCard } from "@/components/guest/PulsaEntryCard";
import { getDedicatedGuestBrandPath } from "@/lib/dedicated-category-brand-routes";
import { findDetectedOperatorBrand, normalizeOperatorDigits } from "@/lib/operator-brand-detection";
import { getDisplayedFixedPrice } from "@/components/guest/product-card-shared";
import { getProductGroupLabel } from "@/lib/product-grouping";

type GuestPulsaQuickOrderProps = {
  kategoriId: string;
  brands: UserBrandItem[];
  authToken?: string;
  buyerRole?: string;
  forcedBrand?: UserBrandItem;
  relatedKategoriIds?: string[];
};

const BRAND_FAMILY_ALIASES: Record<string, string[]> = {
  xl: ["xl", "axis"],
  axis: ["axis", "xl"],
};

function formatNominal(value: number) {
  return new Intl.NumberFormat("id-ID").format(value || 0);
}

function extractPulsaLabel(item: UserProductItem) {
  const nominal = extractPulsaNominalValue(item);
  if (nominal > 0) {
    return formatNominal(nominal);
  }

  const upper = item.nama.toUpperCase();
  const compact = upper.match(/(\d+)\s*K\b/);
  if (compact) return `${Number.parseInt(compact[1], 10)}K`;

  return item.nama;
}

function extractPulsaNominalValue(item: UserProductItem) {
  const upper = item.nama.toUpperCase();
  const dotted = upper.match(/(\d{1,3}(?:\.\d{3})+)/);
  if (dotted) {
    const parsed = Number.parseInt(dotted[1].replace(/\./g, ""), 10);
    if (!Number.isNaN(parsed)) return parsed;
  }

  const compact = upper.match(/(\d+)\s*K\b/);
  if (compact) {
    const parsed = Number.parseInt(compact[1], 10);
    if (!Number.isNaN(parsed)) return parsed * 1000;
  }

  return Number(item.nominal || 0);
}

function normalizeVariantSpaces(value: string) {
  return value
    .replace(/\s+/g, " ")
    .replace(/REGULERVIP/g, "REGULER VIP")
    .trim();
}

function escapeRegex(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function extractOperatorVariant(item: UserProductItem, brandName: string) {
  const fromGroup = normalizeVariantSpaces(getProductGroupLabel(item, ""));
  if (fromGroup) return fromGroup;

  const kategoriUpper = String(item.kategori_nama || "").toUpperCase();
  if (kategoriUpper.includes("PAKET TELEPON")) return "PAKET TELEPON";
  if (kategoriUpper.includes("PAKET SMS")) return "PAKET SMS";

  const brandUpper = brandName.toUpperCase();
  const upper = item.nama.toUpperCase();

  let variant = upper;
  variant = variant.replace(new RegExp(`^${escapeRegex(brandUpper)}\\s+`), "");
  variant = variant.replace(/\b\d{1,3}(?:\.\d{3})+\b/g, "");
  variant = variant.replace(/\b\d+\s*K\b/g, "");
  variant = normalizeVariantSpaces(variant);

  if (brandUpper === "BY.U") {
    variant = variant.replace(/\bBY\.U\b/g, "");
    variant = normalizeVariantSpaces(variant);
  }

  if (!variant) return "REGULER";
  return variant;
}

function getVariantPriority(label: string, brandName?: string) {
  const upper = label.toUpperCase();
  const brandUpper = String(brandName || "").toUpperCase();
  if (brandUpper === "BY.U" && upper.includes("DETIK")) return -1;
  if (upper.includes("VIP")) return 0;
  if (upper.includes("REGULER")) return 1;
  if (upper.includes("PROMO")) return 2;
  if (upper.includes("TRANSFER")) return 3;
  if (upper.includes("PAKET TELEPON")) return 4;
  if (upper.includes("PAKET SMS")) return 5;
  return 6;
}

function isDescriptiveVariant(label: string) {
  const upper = label.toUpperCase();
  return upper.includes("PAKET TELEPON") || upper.includes("PAKET SMS");
}

async function getClientProductsByBrand(kategoriIds: string[], brandId: string): Promise<UserProductItem[]> {
  const rows = await Promise.all(
    kategoriIds.map(async (kategoriId) => {
      const res = await fetch(
        `/api/app/produk?kategori_id=${encodeURIComponent(kategoriId)}&brand_id=${encodeURIComponent(brandId)}`,
        {
          method: "GET",
          cache: "no-store",
          headers: {
            "Content-Type": "application/json",
          },
        }
      );

      if (!res.ok) {
        throw new Error(`Gagal mengambil produk pulsa (${res.status})`);
      }

      const json = (await res.json().catch(() => ({}))) as { items?: UserProductItem[] };
      return Array.isArray(json.items) ? json.items : [];
    })
  );

  const merged = new Map<number, UserProductItem>();
  for (const items of rows) {
    for (const item of items) {
      merged.set(item.id, item);
    }
  }

  return Array.from(merged.values());
}

function resolveBrandFamilyIds(selectedBrand: UserBrandItem, brands: UserBrandItem[]) {
  const normalized = selectedBrand.nama.trim().toLowerCase();
  const aliases = BRAND_FAMILY_ALIASES[normalized];
  if (!aliases) return [String(selectedBrand.id)];

  const ids = brands
    .filter((brand) => aliases.includes(brand.nama.trim().toLowerCase()))
    .map((brand) => String(brand.id));

  return ids.length ? Array.from(new Set(ids)) : [String(selectedBrand.id)];
}

export function GuestPulsaQuickOrder({ kategoriId, brands, authToken, buyerRole, forcedBrand, relatedKategoriIds }: GuestPulsaQuickOrderProps) {
  const [phone, setPhone] = React.useState("");
  const [nominalInput, setNominalInput] = React.useState("");
  const [products, setProducts] = React.useState<UserProductItem[]>([]);
  const [loading, setLoading] = React.useState(false);
  const [selectedProduct, setSelectedProduct] = React.useState<UserProductItem | null>(null);
  const [selectedVariant, setSelectedVariant] = React.useState("");
  const cacheRef = React.useRef<Record<string, UserProductItem[]>>({});
  const sourceKategoriIds = React.useMemo(() => {
    const values = [kategoriId, ...(relatedKategoriIds || [])].map((value) => String(value).trim()).filter(Boolean);
    return Array.from(new Set(values));
  }, [kategoriId, relatedKategoriIds]);
  const allowedKategoriIds = React.useMemo(() => new Set(sourceKategoriIds.map((value) => Number(value))), [sourceKategoriIds]);

  const detectedBrand = React.useMemo(() => {
    if (forcedBrand) return forcedBrand;
    return findDetectedOperatorBrand(phone, brands);
  }, [forcedBrand, phone, brands]);
  const normalizedPhone = React.useMemo(() => normalizeOperatorDigits(phone), [phone]);
  const nominalValue = Number.parseInt(nominalInput.replace(/\D/g, ""), 10) || 0;
  const brandFamilyIds = React.useMemo(() => {
    if (!detectedBrand) return [];
    return resolveBrandFamilyIds(detectedBrand, brands);
  }, [detectedBrand, brands]);
  const cacheKey = React.useMemo(() => brandFamilyIds.slice().sort().join(","), [brandFamilyIds]);

  React.useEffect(() => {
    if (!detectedBrand?.id || !brandFamilyIds.length) {
      setProducts([]);
      return;
    }

    const cached = cacheRef.current[cacheKey];
    if (cached) {
      setProducts(cached);
      return;
    }

    let active = true;
    setLoading(true);
    void Promise.all(brandFamilyIds.map((brandId) => getClientProductsByBrand(sourceKategoriIds, brandId)))
      .then((brandRows) => {
        if (!active) return;
        const merged = new Map<number, UserProductItem>();
        for (const rows of brandRows) {
          for (const item of rows) {
            merged.set(item.id, item);
          }
        }
        const sorted = Array.from(merged.values())
          .filter((item) => item.aktif !== false && item.tipe_harga === "FIXED" && allowedKategoriIds.has(Number(item.kategori_id)))
          .sort((a, b) => Number(a.nominal || 0) - Number(b.nominal || 0));
        cacheRef.current[cacheKey] = sorted;
        setProducts(sorted);
      })
      .catch(() => {
        if (!active) return;
        setProducts([]);
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, [allowedKategoriIds, brandFamilyIds, cacheKey, detectedBrand?.id, sourceKategoriIds]);

  const variantGroups = React.useMemo(() => {
    if (!detectedBrand) return [];

    const grouped = new Map<string, UserProductItem[]>();
    for (const item of products) {
      const key = extractOperatorVariant(item, detectedBrand.nama);
      const bucket = grouped.get(key) || [];
      bucket.push(item);
      grouped.set(key, bucket);
    }

    return Array.from(grouped.entries())
      .map(([label, items]) => ({
        label,
        items: [...items].sort((a, b) => extractPulsaNominalValue(a) - extractPulsaNominalValue(b)),
      }))
      .sort((a, b) => {
        const priority = getVariantPriority(a.label, detectedBrand.nama) - getVariantPriority(b.label, detectedBrand.nama);
        if (priority !== 0) return priority;
        return a.label.localeCompare(b.label);
      });
  }, [detectedBrand, products]);

  React.useEffect(() => {
    if (!variantGroups.length) {
      setSelectedVariant("");
      return;
    }

    setSelectedVariant((current) => {
      if (current && variantGroups.some((group) => group.label === current)) return current;
      return variantGroups[0].label;
    });
  }, [variantGroups]);

  const activeProducts = React.useMemo(() => {
    if (!variantGroups.length) return products;
    return variantGroups.find((group) => group.label === selectedVariant)?.items || variantGroups[0]?.items || products;
  }, [products, selectedVariant, variantGroups]);

  const matchedProduct = React.useMemo(() => {
    if (!nominalValue) return null;
    return products.find((item) => extractPulsaNominalValue(item) === nominalValue) || null;
  }, [nominalValue, products]);

  const visibleProducts = React.useMemo(() => {
    if (matchedProduct) return [matchedProduct];
    return activeProducts;
  }, [activeProducts, matchedProduct]);

  return (
    <>
      <div className="space-y-4">
        <PulsaEntryCard
          phone={phone}
          nominalInput={nominalInput}
          onPhoneChange={setPhone}
          onNominalChange={setNominalInput}
          onQuickBuy={() => {
            if (matchedProduct && normalizedPhone.length >= 8 && normalizedPhone.length <= 13) {
              setSelectedProduct(matchedProduct);
            }
          }}
          buyDisabled={!matchedProduct || normalizedPhone.length < 8 || normalizedPhone.length > 13}
          detectedBrand={detectedBrand}
          matchedProduct={matchedProduct}
          normalizedPhone={normalizedPhone}
          nominalValue={nominalValue}
          formatNominal={formatNominal}
          getDisplayedFixedPrice={(item) => getDisplayedFixedPrice(item, buyerRole || (authToken ? "user" : "guest"))}
        />

        <section className="">
          <div className="space-y-4">
            {!detectedBrand && !forcedBrand ? (
            <div className="space-y-3">
              <ProviderBrandPicker
                items={brands.map((brand) => ({
                  brand,
                  href: getDedicatedGuestBrandPath(kategoriId, brand) || `/kategori/${kategoriId}/brand/${brand.id}?name=${encodeURIComponent(brand.nama)}`,
                }))}
              />
            </div>
          ) : null}

          {detectedBrand ? (
            <div>
              <div className="mb-2 flex items-center justify-between gap-3">
                {loading ? <LoaderCircle className="h-4 w-4 animate-spin text-sky-600" /> : null}
              </div>

              {variantGroups.length > 1 ? (
                <div className="mb-3 flex snap-x snap-mandatory gap-1.5 overflow-x-auto pb-1 [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
                  {variantGroups.map((group) => (
                    <button
                      type="button"
                      key={group.label}
                      onClick={() => setSelectedVariant(group.label)}
                      className={`shrink-0 snap-start whitespace-nowrap rounded-full px-2.5 py-2 text-[10px] font-semibold leading-tight transition ${
                        selectedVariant === group.label
                          ? "bg-sky-600 text-white shadow-[0_8px_18px_rgba(15,111,203,0.22)]"
                          : "bg-white text-slate-700 ring-1 ring-slate-200 hover:bg-slate-50"
                      }`}
                    >
                      {detectedBrand.nama.toUpperCase()} {group.label}
                    </button>
                  ))}
                </div>
              ) : null}

              {loading ? (
                <div className="grid min-h-28 place-items-center rounded-md border border-dashed border-slate-200 bg-slate-50 text-sm text-slate-500">
                  Memuat produk pulsa...
                </div>
              ) : visibleProducts.length > 0 ? (
                <QuickProductOptionGrid
                  items={visibleProducts.map((item) => {
                    const showDescription = isDescriptiveVariant(selectedVariant);
                    const nominalLabel = extractPulsaLabel(item);
                    return {
                      id: item.id,
                      title: (
                        <p
                          className={`text-white ${
                            showDescription
                              ? "text-[14px] font-bold leading-tight"
                              : "text-[28px] font-extrabold leading-none tracking-tight"
                          }`}
                        >
                          {showDescription ? item.nama : nominalLabel}
                        </p>
                      ),
                      subtitle: <>Rp. {formatNominal(getDisplayedFixedPrice(item, buyerRole || (authToken ? "user" : "guest")))}</>,
                    };
                  })}
                  selectedId={matchedProduct?.id}
                  onSelect={(id) => {
                    const found = visibleProducts.find((item) => item.id === id) || null;
                    setSelectedProduct(found);
                  }}
                  variant="pulsa"
                />
              ) : (
                <div className="grid min-h-28 place-items-center rounded-md border border-dashed border-slate-200 bg-slate-50 text-sm text-slate-500">
                  Belum ada produk aktif untuk operator ini.
                </div>
              )}
            </div>
          ) : null}
          </div>
        </section>
      </div>

      <UserCheckoutModal
        open={Boolean(selectedProduct)}
        product={selectedProduct}
        authToken={authToken}
        buyerRole={buyerRole}
        initialDest={normalizedPhone}
        onClose={() => setSelectedProduct(null)}
      />
    </>
  );
}
