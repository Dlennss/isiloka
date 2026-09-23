import type { MetadataRoute } from "next";
import { CANONICAL_SITE_URL, seoArticles } from "@/lib/seo-articles";
import { getBrandsByKategori } from "@/lib/api.products";
import { getDedicatedGuestBrandPath } from "@/lib/dedicated-category-brand-routes";

export const dynamic = "force-dynamic";

const staticRoutes = [
  "",
  "/tentang",
  "/docs",
  "/pulsa",
  "/pulsa-data",
  "/paket-data",
  "/paket-telepon",
  "/ewallet",
  "/game",
  "/listrik",
  "/listrik/token",
  "/listrik/tagihan",
  "/bpjs",
  "/pdam",
  "/pgn",
  "/internet-pascabayar",
  "/hp-pascabayar",
  "/tv",
  "/masa-aktif",
  "/aktivasi-perdana",
  "/artikel",
];

const detailRouteKategoriIds = ["1", "2", "3", "4", "5", "7", "8", "9", "10", "17", "18", "20"] as const;

const bpjsDetailRoutes = ["/bpjs/kesehatan", "/bpjs/ketenagakerjaan"];

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const now = new Date();

  const pages = staticRoutes.map((path) => ({
    url: `${CANONICAL_SITE_URL}${path}`,
    lastModified: now,
    changeFrequency: path === "" ? "daily" : "weekly",
    priority: path === "" ? 1 : path === "/artikel" ? 0.9 : 0.8,
  })) satisfies MetadataRoute.Sitemap;

  const articlePages = seoArticles.map((article) => ({
    url: `${CANONICAL_SITE_URL}/artikel/${article.slug}`,
    lastModified: new Date(article.updatedAt),
    changeFrequency: "monthly" as const,
    priority: 0.7,
  }));

  const detailBrandGroups = await Promise.all(
    detailRouteKategoriIds.map(async (kategoriId) => {
      const brands = await getBrandsByKategori(kategoriId);
      return brands
        .filter((brand) => brand.aktif)
        .map((brand) => getDedicatedGuestBrandPath(kategoriId, brand))
        .filter((path): path is string => Boolean(path));
    }),
  );

  const detailBrandPages = Array.from(new Set([...bpjsDetailRoutes, ...detailBrandGroups.flat()])).map((path) => ({
    url: `${CANONICAL_SITE_URL}${path}`,
    lastModified: now,
    changeFrequency: "weekly" as const,
    priority: 0.75,
  }));

  return [...pages, ...detailBrandPages, ...articlePages];
}
