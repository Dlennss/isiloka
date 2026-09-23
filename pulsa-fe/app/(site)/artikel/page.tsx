import type { Metadata } from "next";
import Link from "next/link";
import { ChevronLeft } from "lucide-react";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { CANONICAL_SITE_URL, DEFAULT_OG_IMAGE_URL, getArticleProductLinks, getCanonicalUrl, seoArticles, suggestedTopics } from "@/lib/seo-articles";

export const metadata: Metadata = {
  metadataBase: new URL(CANONICAL_SITE_URL),
  title: "Artikel Isiloka",
  description:
    "Panduan lengkap tentang produk Isiloka, keuntungan menjadi member, layanan H2H, website jualan sendiri, dan kebutuhan transaksi digital harian.",
  alternates: {
    canonical: getCanonicalUrl("/artikel"),
  },
  openGraph: {
    title: "Artikel Isiloka",
    description:
      "Panduan lengkap tentang produk Isiloka, keuntungan menjadi member, layanan H2H, website jualan sendiri, dan kebutuhan transaksi digital harian.",
    url: getCanonicalUrl("/artikel"),
    siteName: "Isiloka",
    type: "website",
    images: [
      {
        url: DEFAULT_OG_IMAGE_URL,
        alt: "Artikel Isiloka",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "Artikel Isiloka",
    description:
      "Panduan lengkap tentang produk Isiloka, keuntungan menjadi member, layanan H2H, website jualan sendiri, dan kebutuhan transaksi digital harian.",
    images: [DEFAULT_OG_IMAGE_URL],
  },
  robots: {
    index: true,
    follow: true,
  },
};

export default function ArtikelIndexPage() {
  const itemListJsonLd = {
    "@context": "https://schema.org",
    "@type": "CollectionPage",
    name: "Artikel Isiloka",
    description:
      "Panduan lengkap tentang produk Isiloka, keuntungan menjadi member, layanan H2H, website jualan sendiri, dan kebutuhan transaksi digital harian.",
    url: getCanonicalUrl("/artikel"),
    about: [
      "Pulsa",
      "Paket Data",
      "Paket Telepon",
      "Pulsa Data",
      "E-Wallet",
      "Game",
      "Listrik",
      "BPJS",
      "PDAM",
      "PGN",
      "Internet Pascabayar",
      "HP Pascabayar",
      "TV",
      "H2H",
      "Agen Pulsa",
    ],
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
      itemListElement: seoArticles.map((article, index) => ({
        "@type": "ListItem",
        position: index + 1,
        url: getCanonicalUrl(`/artikel/${article.slug}`),
        name: article.title,
      })),
    },
  };

  return (
    <main className="relative overflow-hidden bg-slate-50 text-slate-900">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(itemListJsonLd) }} />
      <div className="pointer-events-none absolute inset-x-0 top-0 h-64 bg-[radial-gradient(560px_220px_at_10%_0%,rgba(45,143,220,0.16),transparent_60%),radial-gradient(680px_260px_at_100%_0%,rgba(6,182,212,0.10),transparent_55%),linear-gradient(180deg,#eef6ff_0%,rgba(248,250,252,0)_100%)]" />

      <section className="relative space-y-4 px-4 pb-6 pt-4">
        <Link
          href="/"
          className="inline-flex items-center gap-1 text-sm font-medium text-sky-700"
          aria-label="Kembali ke beranda"
        >
          <ChevronLeft className="h-4 w-4" />
          <span>Kembali</span>
        </Link>

        <div className="overflow-hidden rounded-md border border-sky-100 bg-linear-to-br from-white via-sky-50/70 to-cyan-50/80 p-4 shadow-[0_18px_40px_rgba(15,111,203,0.10)]">
          <p className="inline-flex rounded-full border border-sky-200 bg-white/90 px-4 py-1.5 text-[10px] font-semibold uppercase tracking-[0.18em] text-sky-700">
            Artikel & Panduan
          </p>
          <h1 className="mt-4 text-2xl font-black leading-tight tracking-tight text-slate-950">
            Panduan lengkap tentang produk dan peluang usaha di Isiloka
          </h1>
          <p className="mt-3 text-sm leading-7 text-slate-600">
            Halaman ini berisi panduan untuk calon pelanggan, calon member, agen, dan pemilik bisnis yang ingin memahami produk Isiloka
            dengan lebih jelas.
          </p>
        </div>

        <div className="grid gap-3">
          {seoArticles.map((article) => (
            <Link
              key={article.slug}
              href={`/artikel/${article.slug}`}
              className="rounded-md border border-slate-200 bg-white p-4 shadow-[0_16px_36px_rgba(15,23,42,0.06)] transition hover:-translate-y-0.5 hover:border-sky-200"
            >
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-sky-700">{article.category}</p>
                  <h2 className="mt-2 text-lg font-black leading-tight tracking-tight text-slate-950">{article.title}</h2>
                </div>
                <span className="shrink-0 rounded-full bg-slate-100 px-3 py-1 text-[11px] font-semibold text-slate-600">{article.readTime}</span>
              </div>
              <p className="mt-3 text-sm leading-7 text-slate-600">{article.excerpt}</p>
              <div className="mt-4 flex flex-wrap gap-2">
                {article.products.slice(0, 4).map((product) => (
                  <span
                    key={product}
                    className="rounded-full border border-sky-100 bg-sky-50 px-3 py-1 text-[11px] font-semibold text-sky-700"
                  >
                    {product}
                  </span>
                ))}
              </div>
              <div className="mt-4 flex flex-wrap gap-2">
                {getArticleProductLinks(article).slice(0, 2).map((item) => (
                  <span
                    key={item.href}
                    className="rounded-full border border-slate-200 bg-white px-3 py-1 text-[11px] font-semibold text-slate-700"
                  >
                    {item.label}
                  </span>
                ))}
              </div>
            </Link>
          ))}
        </div>

        <section className="rounded-md border border-slate-200 bg-white p-4 shadow-[0_16px_36px_rgba(15,23,42,0.06)]">
          <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-sky-700">Topik Berikutnya</p>
          <h2 className="mt-2 text-lg font-black tracking-tight text-slate-950">Artikel lain yang layak ditambahkan</h2>
          <div className="mt-4 grid gap-2">
            {suggestedTopics.map((topic) => (
              <div key={topic} className="rounded-2xl border border-slate-200 bg-slate-50/80 px-4 py-3 text-sm text-slate-700">
                {topic}
              </div>
            ))}
          </div>
        </section>
      </section>

      <GuestBottomNav />
    </main>
  );
}
