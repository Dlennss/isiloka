import type { MetadataRoute } from "next";
import { CANONICAL_SITE_URL } from "@/lib/seo-articles";

export default function robots(): MetadataRoute.Robots {
  const host = CANONICAL_SITE_URL.replace(/^https?:\/\//, "");

  return {
    rules: {
      userAgent: "*",
      allow: "/",
      disallow: ["/api/", "/dashboard/", "/user/"],
    },
    sitemap: `${CANONICAL_SITE_URL}/sitemap.xml`,
    host,
  };
}
