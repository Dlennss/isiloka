import type { Metadata } from "next";
import { GuestCategoryPageContent } from "@/components/site/GuestCategoryPageContent";
import { buildBreadcrumbJsonLd, buildCollectionJsonLd, buildPageMetadata } from "@/lib/site-search";

export const metadata: Metadata = buildPageMetadata({
  title: "Top Up Game Online | Isiloka",
  description: "Top up Mobile Legends, Free Fire, PUBG Mobile, Roblox, dan game online lain di Isiloka dengan pilihan nominal yang jelas.",
  path: "/game",
  keywords: ["top up game", "voucher game online", "diamond mobile legends", "top up free fire", "pulsakilat"],
});

export default async function GamePage() {
  const collectionJsonLd = buildCollectionJsonLd({
    title: "Top Up Game Online | Isiloka",
    description: "Top up Mobile Legends, Free Fire, PUBG Mobile, Roblox, dan game online lain di Isiloka dengan pilihan nominal yang jelas.",
    path: "/game",
  });
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "Game", path: "/game" },
  ]);

  return (
    <main>
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <GuestCategoryPageContent kategoriId="5" title="Game" />
    </main>
  );
}
