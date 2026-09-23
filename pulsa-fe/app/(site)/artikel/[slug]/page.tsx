import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { ChevronLeft } from "lucide-react";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import {
  CANONICAL_SITE_URL,
  DEFAULT_OG_IMAGE_URL,
  getCanonicalArticleUrl,
  getCanonicalUrl,
  getArticleProductLinks,
  getRelatedArticles,
  getSeoArticle,
  seoArticles,
} from "@/lib/seo-articles";

type Props = {
  params: Promise<{ slug: string }>;
};

export async function generateStaticParams() {
  return seoArticles.map((article) => ({ slug: article.slug }));
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;
  const article = getSeoArticle(slug);

  if (!article) {
    return {
      title: "Artikel Tidak Ditemukan | Isiloka",
    };
  }

  return {
    metadataBase: new URL(CANONICAL_SITE_URL),
    title: `${article.title} | Isiloka`,
    description: article.description,
    keywords: article.keywords,
    alternates: {
      canonical: getCanonicalArticleUrl(article.slug),
    },
    openGraph: {
      title: article.title,
      description: article.description,
      url: getCanonicalArticleUrl(article.slug),
      type: "article",
      siteName: "Isiloka",
      publishedTime: article.publishedAt,
      modifiedTime: article.updatedAt,
      authors: ["Isiloka"],
      tags: article.keywords,
      images: [
        {
          url: DEFAULT_OG_IMAGE_URL,
          alt: article.title,
        },
      ],
    },
    twitter: {
      card: "summary_large_image",
      title: article.title,
      description: article.description,
      images: [DEFAULT_OG_IMAGE_URL],
    },
    robots: {
      index: true,
      follow: true,
    },
  };
}

export default async function ArtikelDetailPage({ params }: Props) {
  const { slug } = await params;
  const article = getSeoArticle(slug);

  if (!article) {
    notFound();
  }
  const productLinks = getArticleProductLinks(article);
  const relatedArticles = getRelatedArticles(article.slug);
  const ctaHeading =
    article.category === "Bisnis Agen"
      ? "Siapkan langkah berikutnya untuk usaha Anda"
      : article.category === "Panduan Member"
        ? "Pilih jalur yang paling cocok untuk mulai"
        : article.category === "H2H"
          ? "Lanjut ke integrasi yang lebih serius"
          : article.category === "Deposit"
            ? "Jaga transaksi tetap lancar"
            : "Lanjutkan ke layanan yang paling relevan";
  const ctaDescription =
    article.category === "Bisnis Agen"
      ? "Setelah memahami arahnya, lanjutkan ke halaman Isiloka yang paling membantu pertumbuhan usaha dan layanan member Anda."
      : article.category === "Panduan Member"
        ? "Gunakan halaman berikut untuk mulai dari produk yang paling mudah dijalankan atau untuk menyiapkan langkah bisnis berikutnya."
        : article.category === "H2H"
          ? "Kalau bisnis Anda sudah siap masuk ke otomasi, lanjutkan ke halaman integrasi dan pelajari fondasi yang dibutuhkan."
          : article.category === "Deposit"
            ? "Masuk ke akun member untuk menyiapkan saldo dan memastikan operasional transaksi tetap jalan tanpa hambatan."
            : "Jika Anda sudah memahami produknya, lanjutkan ke halaman layanan Isiloka yang paling relevan untuk kebutuhan Anda.";

  const articleJsonLd = {
    "@context": "https://schema.org",
    "@type": "Article",
    headline: article.title,
    description: article.description,
    datePublished: article.publishedAt,
    dateModified: article.updatedAt,
    image: [DEFAULT_OG_IMAGE_URL],
    inLanguage: "id-ID",
    keywords: article.keywords.join(", "),
    articleSection: article.category,
    about: article.products,
    author: {
      "@type": "Organization",
      name: "Isiloka",
      url: CANONICAL_SITE_URL,
    },
    publisher: {
      "@type": "Organization",
      name: "Isiloka",
      url: CANONICAL_SITE_URL,
      logo: {
        "@type": "ImageObject",
        url: DEFAULT_OG_IMAGE_URL,
      },
    },
    mainEntityOfPage: getCanonicalArticleUrl(article.slug),
  };

  const faqJsonLd = {
    "@context": "https://schema.org",
    "@type": "FAQPage",
    mainEntity: article.faqs.map((faq) => ({
      "@type": "Question",
      name: faq.question,
      acceptedAnswer: {
        "@type": "Answer",
        text: faq.answer,
      },
    })),
  };

  const breadcrumbJsonLd = {
    "@context": "https://schema.org",
    "@type": "BreadcrumbList",
    itemListElement: [
      {
        "@type": "ListItem",
        position: 1,
        name: "Beranda",
        item: getCanonicalUrl("/"),
      },
      {
        "@type": "ListItem",
        position: 2,
        name: "Artikel",
        item: getCanonicalUrl("/artikel"),
      },
      {
        "@type": "ListItem",
        position: 3,
        name: article.title,
        item: getCanonicalArticleUrl(article.slug),
      },
    ],
  };

  return (
    <main className="relative overflow-hidden bg-slate-50 text-slate-900">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(articleJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(faqJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="pointer-events-none absolute inset-x-0 top-0 h-72 bg-[radial-gradient(560px_220px_at_10%_0%,rgba(45,143,220,0.16),transparent_60%),radial-gradient(680px_260px_at_100%_0%,rgba(6,182,212,0.10),transparent_55%),linear-gradient(180deg,#eef6ff_0%,rgba(248,250,252,0)_100%)]" />

      <article className="relative space-y-4 px-4 pb-6 pt-4">
        <Link
          href="/artikel"
          className="inline-flex items-center gap-1 text-sm font-medium text-sky-700"
          aria-label="Kembali ke artikel"
        >
          <ChevronLeft className="h-4 w-4" />
          <span>Kembali ke artikel</span>
        </Link>

        <header className="overflow-hidden rounded-md border border-sky-100 bg-linear-to-br from-white via-sky-50/70 to-cyan-50/80 p-4 shadow-[0_18px_40px_rgba(15,111,203,0.10)]">
          <div className="flex flex-wrap items-center gap-2">
            <span className="rounded-full border border-sky-200 bg-white/90 px-4 py-1.5 text-[10px] font-semibold uppercase tracking-[0.18em] text-sky-700">
              {article.category}
            </span>
            <span className="rounded-full border border-slate-200 bg-white/90 px-4 py-1.5 text-[10px] font-semibold uppercase tracking-[0.18em] text-slate-500">
              {article.readTime}
            </span>
          </div>
          <h1 className="mt-4 text-2xl font-black leading-tight tracking-tight text-slate-950">{article.title}</h1>
          <p className="mt-3 text-sm leading-7 text-slate-600">{article.description}</p>
          <p className="mt-4 rounded-2xl border border-white/80 bg-white/85 px-4 py-3 text-sm leading-7 text-slate-700">
            {article.hero}
          </p>
          <div className="mt-4 flex flex-wrap gap-2">
            {article.products.map((product) => (
              <span key={product} className="rounded-full border border-sky-100 bg-sky-50 px-3 py-1 text-[11px] font-semibold text-sky-700">
                {product}
              </span>
            ))}
          </div>
        </header>

        <div className="grid gap-4">
          {article.sections.map((section) => (
            <section
              key={section.heading}
              className="rounded-md border border-slate-200 bg-white p-4 shadow-[0_16px_36px_rgba(15,23,42,0.06)]"
            >
              <h2 className="text-lg font-black tracking-tight text-slate-950">{section.heading}</h2>
              <div className="mt-3 space-y-3">
                {section.body.map((paragraph) => (
                  <p key={paragraph} className="text-sm leading-7 text-slate-600">
                    {paragraph}
                  </p>
                ))}
              </div>
            </section>
          ))}
        </div>

        <section className="rounded-md border border-slate-200 bg-white p-4 shadow-[0_16px_36px_rgba(15,23,42,0.06)]">
          <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-sky-700">FAQ</p>
          <h2 className="mt-2 text-lg font-black tracking-tight text-slate-950">Pertanyaan yang sering muncul</h2>
          <div className="mt-4 space-y-3">
            {article.faqs.map((faq) => (
              <details key={faq.question} className="group rounded-2xl border border-slate-200 bg-slate-50/80 p-4">
                <summary className="cursor-pointer list-none font-semibold text-slate-900">
                  {faq.question}
                  <span className="ml-2 inline-block text-slate-400 transition group-open:rotate-180">⌄</span>
                </summary>
                <p className="mt-2 text-sm leading-7 text-slate-600">{faq.answer}</p>
              </details>
            ))}
          </div>
        </section>

        {productLinks.length > 0 ? (
          <section className="rounded-md border border-slate-200 bg-white p-4 shadow-[0_16px_36px_rgba(15,23,42,0.06)]">
            <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-sky-700">Produk Terkait</p>
            <h2 className="mt-2 text-lg font-black tracking-tight text-slate-950">Buka halaman produk yang paling relevan</h2>
            <div className="mt-4 grid gap-3">
              {productLinks.map((item) => (
                <Link
                  key={item.href}
                  href={item.href}
                  className="rounded-2xl border border-slate-200 bg-slate-50/80 px-4 py-3 text-sm font-semibold text-slate-800 transition hover:border-sky-200 hover:text-sky-700"
                >
                  {item.label}
                </Link>
              ))}
            </div>
          </section>
        ) : null}

        {relatedArticles.length > 0 ? (
          <section className="rounded-md border border-slate-200 bg-white p-4 shadow-[0_16px_36px_rgba(15,23,42,0.06)]">
            <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-sky-700">Artikel Terkait</p>
            <h2 className="mt-2 text-lg font-black tracking-tight text-slate-950">Lanjutkan membaca topik yang saling mendukung</h2>
            <div className="mt-4 grid gap-3">
              {relatedArticles.map((related) => (
                <Link
                  key={related.slug}
                  href={`/artikel/${related.slug}`}
                  className="rounded-2xl border border-slate-200 bg-slate-50/80 px-4 py-3 transition hover:border-sky-200"
                >
                  <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-sky-700">{related.category}</p>
                  <p className="mt-1 text-sm font-semibold leading-6 text-slate-900">{related.title}</p>
                  <p className="mt-1 text-xs leading-5 text-slate-600">{related.excerpt}</p>
                </Link>
              ))}
            </div>
          </section>
        ) : null}

        <section className="rounded-md border border-sky-200 bg-linear-to-r from-[#087e8b] via-[#087e8b] to-[#087e8b] p-4 text-white shadow-[0_24px_60px_rgba(15,111,203,0.22)]">
          <p className="text-xs font-semibold uppercase tracking-[0.16em] text-white/70">Lanjutkan dari sini</p>
          <h2 className="mt-2 text-lg font-black tracking-tight">{ctaHeading}</h2>
          <p className="mt-2 text-sm leading-7 text-white/90">{ctaDescription}</p>
          <Link
            href={article.ctaHref}
            className="mt-4 inline-flex h-11 items-center justify-center rounded-full bg-white px-5 text-sm font-semibold text-[#087e8b]! visited:text-[#087e8b]! hover:bg-sky-50 hover:text-[#087e8b]!"
          >
            {article.ctaLabel}
          </Link>
        </section>
      </article>

      <GuestBottomNav />
    </main>
  );
}
