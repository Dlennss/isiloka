"use client";

type ProductItem = {
  id: number;
  sku: string;
  nama: string;
  group_name?: string;
  kategori_id?: number;
  brand_id?: number;
  tipe_harga?: "FIXED" | "OPEN_AMOUNT";
  nominal?: number | null;
  maksimal_nominal?: number | null;
  aktif: boolean;
};

type ProductListResponse = {
  ok?: boolean;
  items?: ProductItem[];
  total_pages?: number;
  page?: number;
};

export async function fetchAllAdminProducts(
  authHeader: Record<string, string>,
  limit = 200,
): Promise<ProductItem[]> {
  const out: ProductItem[] = [];
  let page = 1;
  let totalPages = 1;

  while (page <= totalPages) {
    const r = await fetch(`/api/admin/master/produk?limit=${limit}&page=${page}`, {
      headers: authHeader,
      cache: "no-store",
    });
    const j: ProductListResponse = await r.json().catch(() => ({}));
    if (!r.ok || !j.ok) {
      throw new Error("Gagal mengambil daftar produk");
    }

    const items = Array.isArray(j.items) ? j.items : [];
    out.push(...items);

    totalPages = Math.max(1, Number(j.total_pages || 1));
    page += 1;
  }

  return out;
}
