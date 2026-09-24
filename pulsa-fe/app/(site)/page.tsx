import Script from "next/script";
import type { Metadata } from "next";
import { getCategories } from "@/lib/api.products";
import { getAppServerSession } from "@/lib/server-auth";
import type { UserAppOrder, UserCategoryItem, UserProfile } from "@/components/user/types";
import { GuestConceptHome } from "@/components/guest/GuestConceptHome";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { CANONICAL_SITE_URL } from "@/lib/seo-articles";

const homeTitle = "Isiloka | Pulsa, Paket Data, E-Wallet, Token Listrik, Game & PPOB";
const homeDescription =
  "Isiloka melayani isi pulsa, paket data, top up e-wallet, token listrik, top up game, dan pembayaran PPOB dengan alur cepat untuk pelanggan, member, dan agen.";

export const metadata: Metadata = {
  title: homeTitle,
  description: homeDescription,
  keywords: [
    "Isiloka",
    "isi pulsa online",
    "paket data murah",
    "top up e-wallet",
    "token listrik online",
    "top up game",
    "PPOB online",
  ],
  alternates: {
    canonical: CANONICAL_SITE_URL,
  },
  openGraph: {
    title: homeTitle,
    description: homeDescription,
    url: CANONICAL_SITE_URL,
    siteName: "Isiloka",
    type: "website",
    images: [
      {
        url: "/opengraph-image",
        width: 1200,
        height: 630,
        alt: "Isiloka",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: homeTitle,
    description: homeDescription,
    images: ["/twitter-image"],
  },
};

type ProfileResponse = {
  ok?: boolean;
  profile?: UserProfile;
};

type OrdersResponse = {
  ok?: boolean;
  items?: UserAppOrder[];
};

const apiBase = () => (process.env.NEXT_PUBLIC_API_BASE || process.env.API_BASE || "http://127.0.0.1:8083").replace(/\/+$/, "");

async function getHomeProfile(token?: string): Promise<UserProfile | null> {
  if (!token) return null;
  try {
    const res = await fetch(`${apiBase()}/v1/me/profile`, {
      cache: "no-store",
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!res.ok) return null;
    const json = (await res.json().catch(() => ({}))) as ProfileResponse;
    return json.ok && json.profile ? json.profile : null;
  } catch (error) {
    console.error("[home] gagal mengambil profile", error);
    return null;
  }
}

async function getHomeOrders(token?: string): Promise<UserAppOrder[]> {
  if (!token) return [];
  try {
    const res = await fetch(`${apiBase()}/v1/app/me/orders?limit=3`, {
      cache: "no-store",
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!res.ok) return [];
    const json = (await res.json().catch(() => ({}))) as OrdersResponse;
    return Array.isArray(json.items) ? json.items : [];
  } catch (error) {
    console.error("[home] gagal mengambil aktivitas", error);
    return [];
  }
}

export default async function GuestHomePage() {
  const session = await getAppServerSession();
  const token = session?.backendToken;
  const [categories, profile, recentOrders] = await Promise.all([
    getCategories() as Promise<UserCategoryItem[]>,
    getHomeProfile(token),
    getHomeOrders(token),
  ]);
  const activeCategories = categories.filter((item) => item.aktif);

  const websiteJsonLd = {
    "@context": "https://schema.org",
    "@type": "WebSite",
    name: "Isiloka",
    url: CANONICAL_SITE_URL,
    description: homeDescription,
    inLanguage: "id-ID",
  };

  const organizationJsonLd = {
    "@context": "https://schema.org",
    "@type": "Organization",
    name: "Isiloka",
    url: CANONICAL_SITE_URL,
    logo: `${CANONICAL_SITE_URL}/images/logo-pulsakilat.svg`,
    image: `${CANONICAL_SITE_URL}/opengraph-image`,
    description: homeDescription,
  };

  const catalogJsonLd = {
    "@context": "https://schema.org",
    "@type": "CollectionPage",
    name: "Isiloka",
    url: CANONICAL_SITE_URL,
    description: homeDescription,
    about: activeCategories.map((item) => item.nama),
    mainEntity: {
      "@type": "OfferCatalog",
      name: "Kategori Produk Isiloka",
      itemListElement: activeCategories.map((item, index) => ({
        "@type": "ListItem",
        position: index + 1,
        item: {
          "@type": "Thing",
          name: item.nama,
        },
      })),
    },
  };

  const faqJsonLd = {
    "@context": "https://schema.org",
    "@type": "FAQPage",
    mainEntity: [
      {
        "@type": "Question",
        name: "Produk apa saja yang tersedia di Isiloka?",
        acceptedAnswer: {
          "@type": "Answer",
          text: "Isiloka menyediakan isi pulsa, paket data, top up e-wallet, token listrik, top up game, BPJS, PDAM, internet pascabayar, TV, dan layanan PPOB lain untuk pelanggan, member, dan agen.",
        },
      },
      {
        "@type": "Question",
        name: "Apakah Isiloka cocok untuk calon member dan agen?",
        acceptedAnswer: {
          "@type": "Answer",
          text: "Ya. Isiloka bisa dipakai untuk kebutuhan transaksi harian sekaligus untuk member, agen, reseller, dan kebutuhan H2H dengan katalog produk digital yang lengkap.",
        },
      },
      {
        "@type": "Question",
        name: "Apa keunggulan Isiloka untuk transaksi produk digital?",
        acceptedAnswer: {
          "@type": "Answer",
          text: "Isiloka menata kategori produk secara jelas, menyediakan banyak layanan dalam satu tempat, dan memudahkan pembeli maupun penjual untuk melayani kebutuhan digital harian dengan lebih cepat.",
        },
      },
    ],
  };

  return (
    <main className="brand-retail-main bg-[#f0f8f8]">
      <Script id="homepage-website-jsonld" type="application/ld+json">
        {JSON.stringify(websiteJsonLd)}
      </Script>
      <Script id="homepage-organization-jsonld" type="application/ld+json">
        {JSON.stringify(organizationJsonLd)}
      </Script>
      <Script id="homepage-catalog-jsonld" type="application/ld+json">
        {JSON.stringify(catalogJsonLd)}
      </Script>
      <Script id="homepage-faq-jsonld" type="application/ld+json">
        {JSON.stringify(faqJsonLd)}
      </Script>
      <GuestConceptHome
        user={profile ? {
          isLoggedIn: true,
          name: profile.nama || profile.email,
          email: profile.email,
          image: profile.profile_photo_url || session?.user?.image || null,
          balance: profile.saldo,
        } : { isLoggedIn: false }}
        recentOrders={recentOrders}
      />
      <GuestBottomNav isLoggedIn={Boolean(profile)} />
    </main>
  );
}
