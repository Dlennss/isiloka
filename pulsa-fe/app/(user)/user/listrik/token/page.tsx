import { getBrandsByKategori, getCategories, getProductsByBrand } from "@/lib/api.products";
import type { UserBrandItem, UserCategoryItem, UserProductItem } from "@/components/user/types";
import { GuestElectricityEntryFlow } from "@/components/guest/GuestElectricityEntryFlow";
import { UserBottomNav } from "@/components/user/UserBottomNav";

export const dynamic = "force-dynamic";

function normalizeName(value: string) {
  return value.trim().toLowerCase();
}

function pickCategory(categories: UserCategoryItem[], keyword: string) {
  return categories.find((item) => item.aktif && normalizeName(item.nama).includes(keyword));
}

function pickBrand(brands: UserBrandItem[], patterns: string[]) {
  return brands.find((brand) => {
    const normalized = normalizeName(brand.nama);
    return patterns.some((pattern) => normalized.includes(pattern));
  }) ?? null;
}

function isTokenProduct(product: UserProductItem) {
  const normalizedName = normalizeName(product.nama);
  const normalizedSku = normalizeName(product.sku);
  const nominal = Number(product.nominal ?? 0);

  if (normalizedSku.startsWith("cekidpln") || normalizedSku === "plnc" || normalizedSku === "plnb") return false;
  if (nominal <= 0) return false;
  if (normalizedName.includes("pascabayar")) return false;
  return normalizedName.includes("token listrik");
}

export default async function UserListrikTokenPage() {
  const categories = (await getCategories()) as UserCategoryItem[];
  const listrikCategory = pickCategory(categories, "listrik");
  const brands = listrikCategory ? await getBrandsByKategori(String(listrikCategory.id)) : [];
  const tokenBrand = pickBrand(brands, ["pln", "token", "prabayar"]);
  const products =
    listrikCategory && tokenBrand
      ? ((await getProductsByBrand(String(listrikCategory.id), String(tokenBrand.id))) as UserProductItem[]).filter(isTokenProduct)
      : [];

  return (
    <main className="bg-sky-50">
      <div className="space-y-4 px-4 pt-4">
        <GuestElectricityEntryFlow title="Token Listrik" description="" placeholder="Masukkan nomor meter PLN" items={products} />
      </div>

      <UserBottomNav />
    </main>
  );
}
