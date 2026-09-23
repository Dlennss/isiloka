import type { Metadata } from "next";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/nextauth";
import { getBrandsByKategori, getCategories, getProductsByBrand } from "@/lib/api.products";
import type { UserBrandItem, UserCategoryItem, UserProductItem, UserSession } from "@/components/user/types";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { GuestElectricityEntryFlow } from "@/components/guest/GuestElectricityEntryFlow";
import { getRichProductImageUrl } from "@/lib/product-rich-images";
import { buildBreadcrumbJsonLd, buildCollectionJsonLd, buildPageMetadata, buildProductItemListJsonLd } from "@/lib/site-search";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

function normalizeName(value: string) {
  return value.trim().toLowerCase();
}

function pickCategory(categories: UserCategoryItem[], keyword: string) {
  return categories.find((item) => item.aktif && normalizeName(item.nama).includes(keyword));
}

function pickPLNCategory(categories: UserCategoryItem[]) {
  return pickCategory(categories, "pln") ?? pickCategory(categories, "listrik");
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

  if (normalizedSku.startsWith("cekidpln")) {
    return false;
  }

  if (normalizedSku === "plnc" || normalizedSku === "plnb") {
    return false;
  }

  if (nominal <= 0) {
    return false;
  }

  if (normalizedName.includes("pascabayar")) {
    return false;
  }

  return normalizedName.includes("token listrik");
}

export const metadata: Metadata = buildPageMetadata({
  title: "Token Listrik PLN | Isiloka",
  description: "Beli token listrik PLN di Isiloka dengan pilihan nominal dan grup produk yang lebih jelas.",
  path: "/listrik/token",
  keywords: ["token listrik", "token pln", "beli token listrik", "pulsakilat listrik"],
  imageUrl: getRichProductImageUrl({ brandName: "PLN", categoryName: "Token Listrik" }),
});

export default async function GuestListrikTokenPage() {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const categories = (await getCategories()) as UserCategoryItem[];
  const listrikCategory = pickPLNCategory(categories);
  const brands = listrikCategory ? await getBrandsByKategori(String(listrikCategory.id)) : [];
  const tokenBrand = pickBrand(brands, ["pln", "token", "prabayar"]);
  const allProducts = listrikCategory && tokenBrand
    ? ((await getProductsByBrand(String(listrikCategory.id), String(tokenBrand.id))) as UserProductItem[])
    : [];
  const checkProduct = allProducts.find((item) => normalizeName(item.sku) === "cekidpln") ?? null;
  const products = allProducts.filter(isTokenProduct);
  const collectionJsonLd = buildCollectionJsonLd({
    title: "Token Listrik PLN | Isiloka",
    description: "Beli token listrik PLN di Isiloka dengan pilihan nominal dan grup produk yang lebih jelas.",
    path: "/listrik/token",
    itemNames: products.slice(0, 8).map((item) => item.nama),
  });
  const productJsonLd = buildProductItemListJsonLd({
    title: "Token Listrik PLN | Isiloka",
    path: "/listrik/token",
    brandName: tokenBrand?.nama || "PLN",
    categoryName: "Token Listrik",
    items: products.slice(0, 24),
  });
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "Listrik", path: "/listrik" },
    { name: "Token Listrik", path: "/listrik/token" },
  ]);

  return (
    <main className="bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(productJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4 pt-4">
        <GuestElectricityEntryFlow
          title="Token Listrik"
          description=""
          placeholder="Masukkan nomor meter PLN"
          items={products}
          checkProduct={checkProduct}
          authToken={session?.backendToken}
        />
      </div>

      <GuestBottomNav isLoggedIn={!!session?.backendToken} />
    </main>
  );
}
