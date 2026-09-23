import type { Metadata } from "next";
import Link from "next/link";
import { ArrowRight, ChevronLeft, Clock3, FileText, Handshake, PhoneCall, ShieldCheck, Zap } from "lucide-react";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { CANONICAL_SITE_URL, seoArticles } from "@/lib/seo-articles";

const services = [
  {
    title: "Pulsa & Paket Data",
    description: "Isi pulsa semua operator dan paket data harian sampai bulanan dalam alur yang cepat.",
    href: "/pulsa",
  },
  {
    title: "E-Wallet & Game",
    description: "Top up DANA, OVO, GoPay, LinkAja, dan kebutuhan top up game populer dalam satu tempat.",
    href: "/ewallet",
  },
  {
    title: "Token Listrik & PPOB",
    description: "Listrik, BPJS, PDAM, internet pascabayar, TV, dan kebutuhan rumah tangga lain tersusun rapi.",
    href: "/listrik",
  },
  {
    title: "Member, Agen & H2H",
    description: "Isiloka mendukung pelanggan retail sampai kebutuhan partner bisnis dan integrasi H2H.",
    href: "/docs",
  },
];

const commitments = [
  {
    title: "Transaksi Harian yang Praktis",
    description: "Halaman produk dibangun agar pelanggan cepat menemukan nominal, paket, atau layanan yang dicari tanpa alur yang berbelit.",
    icon: Zap,
  },
  {
    title: "Layanan Stabil untuk Jangka Panjang",
    description: "Isiloka diarahkan sebagai partner transaksi digital yang siap dipakai untuk kebutuhan pribadi, warung, agen, dan reseller.",
    icon: ShieldCheck,
  },
  {
    title: "Siap Melayani Kebutuhan Kemitraan",
    description: "Untuk kebutuhan kemitraan, member, dan H2H, kami menyiapkan jalur layanan yang lebih jelas agar bisnis bisa bertumbuh bertahap.",
    icon: Handshake,
  },
];

const metrics = [
  { label: "Operasional", value: "24/7" },
  { label: "Layanan", value: "Retail, Agen, H2H" },
  { label: "Badan Usaha", value: "PT Pulsa Mitra Nasional" },
];

const articleBacklinks = [
  "produk-pulsakilat-yang-paling-cocok-untuk-calon-member",
  "keuntungan-berlangganan-dan-bertumbuh-bersama-pulsakilat",
  "cara-menjadi-agen-pulsa-di-pulsakilat",
  "cara-menjadi-member-h2h-pulsakilat",
  "cara-membangun-warung-pulsa-dan-ppob-yang-lengkap",
  "cara-membuat-website-jualan-pulsa-sendiri",
]
  .map((slug) => seoArticles.find((article) => article.slug === slug))
  .filter((article): article is NonNullable<(typeof seoArticles)[number]> => Boolean(article));

const pageTitle = "Tentang Isiloka | Layanan Produk Digital untuk Retail, Agen, dan H2H";
const pageDescription =
  "Kenali Isiloka sebagai layanan produk digital untuk pulsa, paket data, e-wallet, token listrik, PPOB, agen, dan kebutuhan H2H.";

export const metadata: Metadata = {
  title: pageTitle,
  description: pageDescription,
  alternates: {
    canonical: `${CANONICAL_SITE_URL}/tentang`,
  },
  openGraph: {
    title: pageTitle,
    description: pageDescription,
    url: `${CANONICAL_SITE_URL}/tentang`,
    siteName: "Isiloka",
    type: "article",
    images: [
      {
        url: "/opengraph-image",
        width: 1200,
        height: 630,
        alt: "Tentang Isiloka",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: pageTitle,
    description: pageDescription,
    images: ["/twitter-image"],
  },
};

export default function TentangPage() {
  return (
    <main className="relative overflow-hidden bg-sky-50 text-slate-900">
      <div className="pointer-events-none absolute inset-x-0 top-0 h-72 bg-[radial-gradient(620px_260px_at_10%_0%,rgba(15,111,203,0.18),transparent_62%),radial-gradient(760px_320px_at_100%_0%,rgba(45,212,191,0.14),transparent_56%),linear-gradient(180deg,#e9f5ff_0%,rgba(233,245,255,0.36)_38%,rgba(233,245,255,0)_100%)]" />

      <section className="relative space-y-5 px-4 pb-5 pt-4">
        <Link
          href="/"
          className="inline-flex items-center gap-1 rounded-full border border-sky-200 bg-white/85 px-3 py-1.5 text-sm font-semibold text-sky-700 shadow-[0_10px_24px_rgba(15,111,203,0.08)]"
          aria-label="Kembali ke beranda"
        >
          <ChevronLeft className="h-4 w-4" />
          <span>Kembali ke beranda</span>
        </Link>

        <div className="overflow-hidden rounded-[28px] border border-white/70 bg-linear-to-br from-[#087e8b] via-[#087e8b] to-[#087e8b] px-5 py-6 text-white shadow-[0_24px_60px_rgba(15,111,203,0.22)]">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div className="max-w-[19rem]">
              <p className="text-[11px] font-semibold uppercase tracking-[0.22em] text-white/72">Tentang Isiloka</p>
              <h1 className="mt-3 text-balance text-[30px] font-black leading-[1.02] tracking-tight">
                Satu tempat untuk transaksi digital harian dan pertumbuhan usaha produk digital.
              </h1>
              <p className="mt-4 text-sm leading-6 text-white/90 text-justify">
                Isiloka melayani kebutuhan pulsa, paket data, e-wallet, token listrik, top up game, dan PPOB untuk
                pelanggan harian, member, agen, sampai kebutuhan kemitraan H2H.
              </p>
            </div>

            <div className="grid w-full gap-3 sm:grid-cols-3">
              {metrics.map((item) => (
                <div
                  key={item.label}
                  className="rounded-2xl border border-white/18 bg-white/10 px-4 py-3 backdrop-blur-sm"
                >
                  <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-white/64">{item.label}</p>
                  <p className="mt-2 text-sm font-bold text-white">{item.value}</p>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      <section className="relative px-4 pb-4">
        <div className="grid gap-4">
          <div className="rounded-[28px] border border-sky-100 bg-white p-5 shadow-[0_18px_40px_rgba(15,23,42,0.08)]">
            <p className="text-[10px] font-semibold uppercase tracking-[0.18em] text-sky-700">Fokus Layanan</p>
            <h2 className="mt-2 text-xl font-black tracking-tight text-slate-950">
              Produk utama Isiloka disusun untuk kebutuhan harian sampai kebutuhan usaha
            </h2>
            <p className="mt-3 text-sm leading-6 text-slate-600 text-justify">
              Kami membangun Isiloka agar pelanggan cepat menemukan layanan yang dicari, dan pada saat yang sama
              member atau agen bisa memperluas usaha tanpa harus pindah platform.
            </p>

            <div className="mt-5 grid gap-3">
              {services.map((service) => (
                <Link
                  key={service.title}
                  href={service.href}
                  className="group rounded-2xl border border-slate-200 bg-slate-50/80 px-4 py-4 transition hover:-translate-y-0.5 hover:border-sky-200 hover:bg-white"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <h3 className="text-sm font-bold text-slate-950">{service.title}</h3>
                      <p className="mt-1 text-sm leading-6 text-slate-600">{service.description}</p>
                    </div>
                    <ArrowRight className="mt-0.5 h-4 w-4 shrink-0 text-sky-600 transition group-hover:translate-x-0.5" />
                  </div>
                </Link>
              ))}
            </div>
          </div>

          <div className="grid gap-4 md:grid-cols-3">
            {commitments.map((item) => {
              const Icon = item.icon;

              return (
                <div
                  key={item.title}
                  className="rounded-[26px] border border-white/70 bg-linear-to-br from-white via-sky-50 to-cyan-50 p-5 shadow-[0_18px_40px_rgba(15,111,203,0.08)]"
                >
                  <div className="grid h-11 w-11 place-items-center rounded-2xl bg-linear-to-br from-sky-500 to-cyan-500 text-white shadow-[0_12px_28px_rgba(15,111,203,0.20)]">
                    <Icon className="h-5 w-5" />
                  </div>
                  <h3 className="mt-4 text-base font-black tracking-tight text-slate-950">{item.title}</h3>
                  <p className="mt-2 text-sm leading-6 text-slate-600 text-justify">{item.description}</p>
                </div>
              );
            })}
          </div>
        </div>
      </section>

      <section className="relative px-4 pb-4">
        <div className="rounded-[28px] border border-slate-200 bg-white p-5 shadow-[0_18px_40px_rgba(15,23,42,0.08)]">
          <div className="flex items-center gap-3">
            <div className="grid h-11 w-11 place-items-center rounded-2xl bg-sky-100 text-sky-700">
              <FileText className="h-5 w-5" />
            </div>
            <div>
              <p className="text-[10px] font-semibold uppercase tracking-[0.18em] text-sky-700">Baca Panduan</p>
              <h2 className="mt-1 text-xl font-black tracking-tight text-slate-950">
                Pelajari Isiloka lebih dalam lewat artikel dan panduan member
              </h2>
            </div>
          </div>

          <div className="mt-5 grid gap-3">
            {articleBacklinks.map((article) => (
              <Link
                key={article.slug}
                href={`/artikel/${article.slug}`}
                className="group rounded-2xl border border-slate-200 bg-slate-50/80 px-4 py-4 transition hover:-translate-y-0.5 hover:border-sky-200 hover:bg-white"
              >
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-sky-700">{article.category}</p>
                    <h3 className="mt-1 text-sm font-bold leading-6 text-slate-950">{article.title}</h3>
                    <p className="mt-1 text-sm leading-6 text-slate-600">{article.excerpt}</p>
                    <div className="mt-2 inline-flex items-center gap-1 text-xs font-semibold text-slate-500">
                      <Clock3 className="h-3.5 w-3.5" />
                      {article.readTime}
                    </div>
                  </div>
                  <ArrowRight className="mt-1 h-4 w-4 shrink-0 text-sky-600 transition group-hover:translate-x-0.5" />
                </div>
              </Link>
            ))}
          </div>
        </div>
      </section>

      <section className="relative px-4 pb-10">
        <aside className="rounded-[28px] border border-sky-200 bg-linear-to-r from-[#087e8b] via-[#087e8b] to-[#087e8b] p-5 text-white shadow-[0_24px_60px_rgba(15,111,203,0.22)]">
          <div className="flex items-start gap-3">
            <div className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-white/12">
              <PhoneCall className="h-5 w-5" />
            </div>
            <div>
              <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-white/70">Kontak Resmi</p>
              <h2 className="mt-2 text-xl font-black tracking-tight">
                Hubungi kami untuk kebutuhan transaksi, member, agen, dan kemitraan
              </h2>
            </div>
          </div>

          <div className="mt-6 grid gap-4 text-sm leading-7 text-white/92">
            <div className="rounded-2xl border border-white/16 bg-white/10 px-4 py-3">
              <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-white/64">Alamat</p>
              <p className="mt-2">
                Perumahan Griya Mulya Indah
                <br />
                Blok JG No. 12 RT 003 / RW 018
                <br />
                Kel. Jayamulya, Kec. Serang Baru
              </p>
            </div>
            <div className="rounded-2xl border border-white/16 bg-white/10 px-4 py-3">
              <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-white/64">WhatsApp & Telepon</p>
              <Link href="tel:+6282219107558" className="mt-2 inline-block text-base font-bold text-cyan-100 hover:text-white">
                0822-1910-7558
              </Link>
            </div>
          </div>

          <div className="mt-6 flex gap-3">
            <Link
              href="https://wa.me/6282219107558"
              className="inline-flex h-12 flex-1 items-center justify-center rounded-full bg-white px-4 text-center text-sm font-semibold text-[#087e8b]! visited:text-[#087e8b]! hover:bg-sky-50 hover:text-[#087e8b]!"
            >
              WhatsApp
            </Link>
            <Link
              href="/artikel"
              className="inline-flex h-12 flex-1 items-center justify-center rounded-full border border-white/30 bg-white/10 px-4 text-center text-sm font-semibold text-white transition hover:bg-white/15"
            >
              Baca Artikel
            </Link>
          </div>
        </aside>
      </section>

      <GuestBottomNav />
    </main>
  );
}
