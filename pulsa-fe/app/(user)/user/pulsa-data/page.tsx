import { getAppServerSession } from "@/lib/server-auth";
import { getCategories, getBrandsByKategori } from "@/lib/api.products";
import type { UserBrandItem, UserCategoryItem, UserSession } from "@/components/user/types";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import { UserAuthClientSync } from "@/components/user/UserAuthClientSync";
import { UserPulsaDataExplorer } from "@/components/user/UserPulsaDataExplorer";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

type PageProps = {
  searchParams?: Promise<{ tab?: string }>;
};

type CategoryWithBrands = {
  category: UserCategoryItem;
  brands: UserBrandItem[];
};

function pickCategory(categories: UserCategoryItem[], keyword: string) {
  return categories.find((item) => item.aktif && item.nama.toLowerCase().includes(keyword));
}

export default async function UserPulsaDataPage({ searchParams }: PageProps) {
  const session = (await getAppServerSession()) as SessionShape | null;
  const backendToken = session?.backendToken;
  const isLoggedIn = Boolean(backendToken);
  const resolvedSearchParams = searchParams ? await searchParams : undefined;
  const rawTab = String(resolvedSearchParams?.tab || "").trim().toLowerCase();
  const initialTab = rawTab === "data" ? "data" : "pulsa";

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

  return (
    <main className="min-h-screen bg-sky-50">
      {backendToken ? <UserAuthClientSync backendToken={backendToken} /> : null}
      <UserPulsaDataExplorer tabs={tabs} isLoggedIn={isLoggedIn} authToken={backendToken} initialTab={initialTab} />
      <UserBottomNav />
    </main>
  );
}
