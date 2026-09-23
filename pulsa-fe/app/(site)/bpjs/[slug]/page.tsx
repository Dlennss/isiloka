import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/nextauth";
import type { UserSession } from "@/components/user/types";
import { getProductsByBrand } from "@/lib/api.products";
import { BPJSBrandFlow } from "@/components/shared/BPJSBrandFlow";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { buildBreadcrumbJsonLd, buildCollectionJsonLd, buildPageMetadata } from "@/lib/site-search";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

type PageProps = {
  params: Promise<{ slug: string }>;
};

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { slug } = await params;
  if (slug !== "kesehatan" && slug !== "ketenagakerjaan") {
    return buildPageMetadata({
      title: "BPJS | Isiloka",
      description: "Cek dan bayar BPJS di Isiloka.",
      path: "/bpjs",
    });
  }

  const title = slug === "kesehatan" ? "BPJS Kesehatan" : "BPJS Ketenagakerjaan";
  return buildPageMetadata({
    title: `${title} | Isiloka`,
    description: `Cek dan bayar ${title} di Isiloka dengan alur yang ringkas.`,
    path: `/bpjs/${slug}`,
    keywords: [title.toLowerCase(), "bayar bpjs", "pulsakilat"],
  });
}

export default async function BPJSSlugPage({ params }: PageProps) {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const buyerRole = String(session?.user?.role || "").trim().toLowerCase();
  const { slug } = await params;
  if (slug !== "kesehatan" && slug !== "ketenagakerjaan") {
    notFound();
  }

  const items = await getProductsByBrand("19", "171");
  const title = slug === "kesehatan" ? "BPJS Kesehatan" : "BPJS Ketenagakerjaan";
  const collectionJsonLd = buildCollectionJsonLd({
    title: `${title} | Isiloka`,
    description: `Cek dan bayar ${title} di Isiloka dengan alur yang ringkas.`,
    path: `/bpjs/${slug}`,
    itemNames: items.slice(0, 8).map((item) => item.nama),
  });
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "BPJS", path: "/bpjs" },
    { name: title, path: `/bpjs/${slug}` },
  ]);

  return (
    <main className="bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4 pt-4">
        <BPJSBrandFlow items={items} authToken={session?.backendToken} buyerRole={buyerRole} initialMode={slug} showModeSelector={false} />
      </div>

      <GuestBottomNav isLoggedIn={!!session?.backendToken} />
    </main>
  );
}
