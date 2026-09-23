/**
 * Product & Category API Calls
 * Semua fetch untuk kategori, brand, dan produk
 */

import { fetchAPI } from './api';
import type { UserBrandItem, UserProductItem } from '@/components/user/types';


function extractSortNumber(value: string, fallback = Number.POSITIVE_INFINITY) {
  const dotted = value.match(/(\d{1,3}(?:\.\d{3})+)/);
  if (dotted) return Number.parseInt(dotted[1].replaceAll(".", ""), 10);

  const compact = value.match(/(\d+)\s*K\b/i);
  if (compact) return Number.parseInt(compact[1], 10) * 1000;

  const plain = value.match(/(\d+)/);
  if (plain) return Number.parseInt(plain[1], 10);

  return fallback;
}

function buildProductSortKey(item: UserProductItem) {
  const category = String(item.kategori_nama || "").toUpperCase();
  const name = String(item.nama || "").trim();
  const upperName = name.toUpperCase();
  const sku = String(item.sku || "").toUpperCase();
  const nominalBase = Number(item.nominal || item.harga_dasar_app || 0);

  if (category.includes("GAME")) {
    if (upperName.includes("CEK NICK")) return [0, 0, sku] as const;
    if (upperName.includes("MEMBER MINGGUAN")) return [1, 0, sku] as const;
    if (upperName.includes("MEMBER BULANAN")) return [1, 1, sku] as const;
    if (upperName.includes("LEVEL UP PASS")) return [2, extractSortNumber(upperName, nominalBase), sku] as const;
    return [3, extractSortNumber(upperName, nominalBase), sku] as const;
  }

  if (category.includes("TV")) {
    const upperSku = String(item.sku || "").toUpperCase();
    if (upperSku.startsWith("CEK")) {
      return [0, 0, sku] as const;
    }
  }

  const typeRank = item.tipe_harga === "FIXED" ? 0 : 1;
  return [typeRank, extractSortNumber(upperName, nominalBase), sku] as const;
}

function normalizeProductsForDisplay(items: UserProductItem[]) {
  if (items.length === 0) return items;

  const category = String(items[0]?.kategori_nama || "").toUpperCase();
  const byName = new Map<string, number>();
  for (const item of items) {
    const key = String(item.nama || "").trim();
    byName.set(key, (byName.get(key) || 0) + 1);
  }

  const normalized = items.map((item) => {
    if (!category.includes("GAME")) return item;

    const rawName = String(item.nama || "").trim();
    const duplicateCount = byName.get(rawName) || 0;
    if (duplicateCount <= 1) return item;

    const sku = String(item.sku || "").trim().toUpperCase();
    const suffix = sku.startsWith("VG") ? "VIP" : sku.startsWith("G") ? "REG" : sku;
    return { ...item, nama: `${rawName} ${suffix}` };
  });

  return normalized.sort((left, right) => {
    const a = buildProductSortKey(left);
    const b = buildProductSortKey(right);
    return a[0] - b[0] || a[1] - b[1] || a[2].localeCompare(b[2]);
  });
}

// ============================================
// KATEGORI
// ============================================

/**
 * Get all categories
 * Cache: 1 jam (kategori jarang berubah)
 */
export async function getCategories() {
  return fetchAPI('/v1/app/kategori', {
    cache: 'no-store',
  });
}

// ============================================
// BRAND
// ============================================

/**
 * Get brands by kategori
 * Cache: 1 jam (brand jarang berubah)
 */
export async function getBrandsByKategori(kategoriId: string): Promise<UserBrandItem[]> {
  return fetchAPI<UserBrandItem>(
    `/v1/app/brand?kategori_id=${encodeURIComponent(kategoriId)}`,
    { cache: 'no-store' }
  );
}

export async function getBrandsByKategoriIds(kategoriIds: string[]): Promise<UserBrandItem[]> {
  const uniqueIds = Array.from(new Set(kategoriIds.map((value) => value.trim()).filter(Boolean)));
  if (uniqueIds.length === 0) return [];
  if (uniqueIds.length === 1) return getBrandsByKategori(uniqueIds[0]);

  const results = await Promise.all(uniqueIds.map((id) => getBrandsByKategori(id)));
  const byId = new Map<number, UserBrandItem>();

  for (const items of results) {
    for (const item of items) {
      if (!item.aktif) continue;
      if (!byId.has(item.id)) {
        byId.set(item.id, item);
      }
    }
  }

  return Array.from(byId.values()).sort((a, b) => a.nama.localeCompare(b.nama));
}

/**
 * Get all brands
 * Cache: 1 jam
 */
export async function getAllBrands(): Promise<UserBrandItem[]> {
  return fetchAPI<UserBrandItem>('/v1/app/brand', {
    revalidate: 30, // 1 menit
  });
}

// ============================================
// PRODUCTS
// ============================================

/**
 * Get products by kategori & brand
 * Cache: 5 menit (harga bisa berubah)
 */
export async function getProductsByBrand(
  kategoriId: string,
  brandId: string
): Promise<UserProductItem[]> {
  const items = await fetchAPI<UserProductItem>(
    `/v1/app/produk?kategori_id=${encodeURIComponent(kategoriId)}&brand_id=${encodeURIComponent(brandId)}`,
    { cache: 'no-store' }
  );

  return normalizeProductsForDisplay(items);
}

/**
 * Get all products
 * Cache: 5 menit (harga bisa berubah)
 */
export async function getAllProducts() {
  return fetchAPI('/v1/app/produk', {
    revalidate: 30, // 30 detik
  });
}

/**
 * Search products
 * Cache: 5 menit
 */
export async function searchProducts(query: string) {
  return fetchAPI(`/v1/app/produk/search?q=${encodeURIComponent(query)}`, {
    revalidate: 30,
  });
}

// ============================================
// FEATURED / PROMO
// ============================================

/**
 * Get featured products / on sale
 * Cache: 30 menit (promo bisa berubah)
 */
export async function getFeaturedProducts() {
  return fetchAPI('/v1/app/produk/featured', {
    revalidate: 30, // 30 detik
  });
}
