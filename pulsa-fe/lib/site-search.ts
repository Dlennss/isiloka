import type { Metadata } from "next";
import { CANONICAL_SITE_URL, DEFAULT_OG_IMAGE_URL } from "@/lib/seo-articles";
import type { UserProductItem } from "@/components/user/types";
import { getRichProductImageUrl } from "@/lib/product-rich-images";

type MetaInput = {
  title: string;
  description: string;
  path: string;
  keywords?: string[];
  imageUrl?: string;
};

type CollectionJsonLdInput = {
  title: string;
  description: string;
  path: string;
  itemNames?: string[];
};

type BreadcrumbItem = {
  name: string;
  path: string;
};

type FaqItem = {
  question: string;
  answer: string;
};

type ProductItemListInput = {
  title: string;
  path: string;
  brandName?: string;
  categoryName?: string;
  items: UserProductItem[];
};

export function canonicalUrl(path: string) {
  return `${CANONICAL_SITE_URL}${path}`;
}

export function buildPageMetadata({ title, description, path, keywords = [], imageUrl = DEFAULT_OG_IMAGE_URL }: MetaInput): Metadata {
  return {
    metadataBase: new URL(CANONICAL_SITE_URL),
    title,
    description,
    keywords,
    alternates: {
      canonical: canonicalUrl(path),
    },
    openGraph: {
      title,
      description,
      url: canonicalUrl(path),
      siteName: "Isiloka",
      type: "website",
      images: [
        {
          url: imageUrl,
          alt: title,
        },
      ],
    },
    twitter: {
      card: "summary_large_image",
      title,
      description,
      images: [imageUrl],
    },
    robots: {
      index: true,
      follow: true,
    },
  };
}

export function buildCollectionJsonLd({ title, description, path, itemNames = [] }: CollectionJsonLdInput) {
  return {
    "@context": "https://schema.org",
    "@type": "CollectionPage",
    name: title,
    description,
    url: canonicalUrl(path),
    publisher: {
      "@type": "Organization",
      name: "Isiloka",
      url: CANONICAL_SITE_URL,
      logo: {
        "@type": "ImageObject",
        url: DEFAULT_OG_IMAGE_URL,
      },
    },
    mainEntity: {
      "@type": "ItemList",
      itemListElement: itemNames.map((item, index) => ({
        "@type": "ListItem",
        position: index + 1,
        name: item,
      })),
    },
  };
}

export function buildBreadcrumbJsonLd(items: BreadcrumbItem[]) {
  return {
    "@context": "https://schema.org",
    "@type": "BreadcrumbList",
    itemListElement: items.map((item, index) => ({
      "@type": "ListItem",
      position: index + 1,
      name: item.name,
      item: canonicalUrl(item.path),
    })),
  };
}

export function buildFaqJsonLd(items: FaqItem[]) {
  return {
    "@context": "https://schema.org",
    "@type": "FAQPage",
    mainEntity: items.map((item) => ({
      "@type": "Question",
      name: item.question,
      acceptedAnswer: {
        "@type": "Answer",
        text: item.answer,
      },
    })),
  };
}

function pickProductPrice(item: UserProductItem) {
  const candidates = [
    item.harga_guest_final,
    item.harga_user_final,
    item.harga_dasar_app,
  ];

  for (const candidate of candidates) {
    const value = Number(candidate ?? 0);
    if (Number.isFinite(value) && value > 0) {
      return value;
    }
  }

  return null;
}

function normalizeWhitespace(value: string) {
  return value.replace(/\s+/g, " ").trim();
}

function cleanProductName(name: string, brandName?: string, categoryName?: string) {
  let cleaned = normalizeWhitespace(name);

  const prefixes = [brandName, categoryName]
    .filter(Boolean)
    .map((value) => String(value).trim())
    .filter(Boolean)
    .sort((a, b) => b.length - a.length);

  for (const prefix of prefixes) {
    const escaped = prefix.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
    cleaned = cleaned.replace(new RegExp(`^${escaped}\\s+`, "i"), "");
  }

  return normalizeWhitespace(cleaned);
}

export function buildProductItemListJsonLd({ title, path, brandName, categoryName, items }: ProductItemListInput) {
  return {
    "@context": "https://schema.org",
    "@type": "ItemList",
    name: title,
    url: canonicalUrl(path),
    numberOfItems: items.length,
    itemListElement: items.map((item, index) => {
      const price = pickProductPrice(item);
      const imageUrl = getRichProductImageUrl({
        brandName: brandName || String(item.brand_nama || "").trim(),
        categoryName: categoryName || String(item.kategori_nama || "").trim(),
        items: [item],
      });

      return {
        "@type": "ListItem",
        position: index + 1,
        item: {
          "@type": "Product",
          name: cleanProductName(String(item.nama || "").trim(), brandName, categoryName),
          image: [imageUrl],
          sku: String(item.sku || "").trim(),
          category: categoryName || String(item.kategori_nama || "").trim() || undefined,
          brand: {
            "@type": "Brand",
            name: brandName || String(item.brand_nama || "").trim() || "Isiloka",
          },
          offers: price
            ? {
                "@type": "Offer",
                priceCurrency: "IDR",
                price,
                availability: "https://schema.org/InStock",
                url: canonicalUrl(path),
                seller: {
                  "@type": "Organization",
                  name: "Isiloka",
                  url: CANONICAL_SITE_URL,
                },
              }
            : undefined,
        },
      };
    }),
  };
}
