import type { Metadata } from "next";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/nextauth";
import { getCategories } from "@/lib/api.products";
import type { UserCategoryItem, UserSession } from "@/components/user/types";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { ServiceDirectory } from "@/components/shared/ServiceDirectory";
import { buildBreadcrumbJsonLd, buildCollectionJsonLd, buildPageMetadata } from "@/lib/site-search";
import { UserUniversalServicePageContent } from "@/components/user/UserUniversalServicePageContent";
import { UserTransferBankPageContent } from "@/components/user/UserTransferBankPageContent";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

type PageProps = {
  searchParams?: Promise<{ layanan?: string }>;
};

export const metadata: Metadata = buildPageMetadata({
  title: "Semua Kategori Produk | Isiloka",
  description: "Lihat semua kategori produk Isiloka, mulai dari pulsa, paket data, e-wallet, game, sampai tagihan.",
  path: "/kategori",
  keywords: ["kategori produk pulsakilat", "semua produk", "pulsakilat"],
});

export default async function GuestKategoriPage({ searchParams }: PageProps) {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const resolvedSearchParams = searchParams ? await searchParams : undefined;
  const serviceSlug = String(resolvedSearchParams?.layanan || "").trim().toLowerCase();

  if (serviceSlug === "transfer-bank") {
    return (
      <main className="min-h-screen bg-[#f0f8f8] pb-24">
        <UserTransferBankPageContent backHref="/kategori" />
        <GuestBottomNav isLoggedIn={!!session?.backendToken} />
      </main>
    );
  }

  if (serviceSlug) {
    return (
      <>
        <UserUniversalServicePageContent serviceSlug={serviceSlug} />
        <GuestBottomNav isLoggedIn={!!session?.backendToken} />
      </>
    );
  }

  const categories = (await getCategories()) as UserCategoryItem[];
  const collectionJsonLd = buildCollectionJsonLd({
    title: "Semua Kategori Produk | Isiloka",
    description: "Lihat semua kategori produk Isiloka, mulai dari pulsa, paket data, e-wallet, game, sampai tagihan.",
    path: "/kategori",
    itemNames: categories.map((category) => category.nama),
  });
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "Kategori", path: "/kategori" },
  ]);

  return (
    <main className="bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4 pt-4">
        <ServiceDirectory mode="guest" />
      </div>

      <GuestBottomNav isLoggedIn={!!session?.backendToken} />
    </main>
  );
}
