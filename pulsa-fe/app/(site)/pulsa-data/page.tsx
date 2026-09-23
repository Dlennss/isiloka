import type { Metadata } from "next";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/nextauth";
import { getCategories, getBrandsByKategori } from "@/lib/api.products";
import type { UserBrandItem, UserCategoryItem, UserSession } from "@/components/user/types";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { UserPulsaDataExplorer } from "@/components/user/UserPulsaDataExplorer";
import { buildBreadcrumbJsonLd, buildCollectionJsonLd, buildPageMetadata } from "@/lib/site-search";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

type CategoryWithBrands = {
  category: UserCategoryItem;
  brands: UserBrandItem[];
};

function pickCategory(categories: UserCategoryItem[], keyword: string) {
  return categories.find((item) => item.aktif && item.nama.toLowerCase().includes(keyword));
}

export const metadata: Metadata = buildPageMetadata({
  title: "Pulsa dan Paket Data | Isiloka",
  description: "Jelajahi pulsa dan paket data semua operator di Isiloka dalam satu halaman yang ringkas.",
  path: "/pulsa-data",
  keywords: ["pulsa dan paket data", "pulsa online", "paket data operator", "pulsakilat"],
});

export default async function GuestPulsaDataPage({
  searchParams,
}: {
  searchParams?: Promise<{ tab?: string }>;
}) {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const isLoggedIn = Boolean(session?.backendToken);
  const resolvedSearchParams = await searchParams;
  const requestedTab = resolvedSearchParams?.tab === "data" ? "data" : "pulsa";

  const categories = (await getCategories()) as UserCategoryItem[];
  const pulsaCategory = pickCategory(categories, "pulsa");
  const dataCategory = pickCategory(categories, "data");

  const [pulsaBrands, dataBrands] = await Promise.all([
    pulsaCategory ? getBrandsByKategori(String(pulsaCategory.id)) : Promise.resolve([]),
    dataCategory ? getBrandsByKategori(String(dataCategory.id)) : Promise.resolve([]),
  ]);

  const tabs: Partial<Record<"pulsa" | "data", CategoryWithBrands>> = {
    ...(pulsaCategory ? { pulsa: { category: pulsaCategory, brands: pulsaBrands } } : {}),
    ...(dataCategory ? { data: { category: dataCategory, brands: dataBrands } } : {}),
  };
  const collectionJsonLd = buildCollectionJsonLd({
    title: "Pulsa dan Paket Data | Isiloka",
    description: "Jelajahi pulsa dan paket data semua operator di Isiloka dalam satu halaman yang ringkas.",
    path: "/pulsa-data",
    itemNames: [...pulsaBrands, ...dataBrands].map((brand) => brand.nama),
  });
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "Pulsa dan Paket Data", path: "/pulsa-data" },
  ]);

  return (
    <main className="bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <UserPulsaDataExplorer
        tabs={tabs}
        isLoggedIn={isLoggedIn}
        backHref="/"
        loginCallbackUrl="/pulsa-data"
        showBackLink={false}
        initialTab={requestedTab}
      />
      <GuestBottomNav isLoggedIn={isLoggedIn} />
    </main>
  );
}
