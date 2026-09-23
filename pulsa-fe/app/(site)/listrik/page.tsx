import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import { getServerSession } from "next-auth";
import { ChevronRight, FileText, Zap } from "lucide-react";
import { authOptions } from "@/lib/nextauth";
import { getCategories } from "@/lib/api.products";
import type { UserCategoryItem, UserSession } from "@/components/user/types";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { buildBreadcrumbJsonLd, buildCollectionJsonLd, buildPageMetadata } from "@/lib/site-search";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

function normalizeName(value: string) {
  return value.trim().toLowerCase();
}

function pickCategory(categories: UserCategoryItem[], keyword: string) {
  return categories.find((item) => item.aktif && normalizeName(item.nama).includes(keyword));
}

function pickPLNCategory(categories: UserCategoryItem[]) {
  return pickCategory(categories, "pln") ?? pickCategory(categories, "listrik");
}

export const metadata: Metadata = buildPageMetadata({
  title: "Produk Listrik PLN | Isiloka",
  description: "Pilih token listrik atau pembayaran tagihan listrik PLN di Isiloka.",
  path: "/listrik",
  keywords: ["token listrik", "tagihan listrik", "pln", "pulsakilat"],
});

export default async function GuestListrikPage() {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const categories = (await getCategories()) as UserCategoryItem[];
  const listrikCategory = pickPLNCategory(categories);

  const cards = [
    {
      key: "token",
      title: "Token Listrik",
      description: "Beli token prabayar dengan nominal populer.",
      href: "/listrik/token",
      icon: Zap,
      accent: "from-lime-300 via-emerald-500 to-[#13464b]",
    },
    {
      key: "tagihan",
      title: "Tagihan Listrik",
      description: "Cek dan bayar tagihan PLN pascabayar.",
      href: "/listrik/tagihan",
      icon: FileText,
      accent: "from-yellow-300 via-green-500 to-[#13464b]",
    },
  ];
  const collectionJsonLd = buildCollectionJsonLd({
    title: "Produk Listrik PLN | Isiloka",
    description: "Pilih token listrik atau pembayaran tagihan listrik PLN di Isiloka.",
    path: "/listrik",
    itemNames: cards.map((item) => item.title),
  });
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "Listrik", path: "/listrik" },
  ]);

  return (
    <main className="min-h-screen bg-[#f0f8f8]">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4 pt-4 pb-28">
        {!listrikCategory ? (
          <section className="grid min-h-40 place-items-center rounded-md border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_10px_28px_rgba(15,23,42,0.08)]">
            Kategori listrik belum tersedia.
          </section>
        ) : (
          <>
            <section className="relative overflow-hidden rounded-[30px] bg-[#13464b] p-4 text-white shadow-[0_22px_48px_rgba(5,46,38,0.26)]">
              <div className="absolute -right-12 -top-12 h-36 w-36 rounded-full bg-lime-300/30 blur-2xl" />
              <div className="absolute bottom-0 left-0 h-24 w-36 bg-emerald-400/15 blur-2xl" />
              <div className="relative flex items-center gap-3">
                <span className="grid h-16 w-16 shrink-0 place-items-center rounded-[24px] bg-white p-3 shadow-[0_14px_28px_rgba(0,0,0,0.18)]">
                  <Image src="/images/pln/logo_pln.png" alt="Logo PLN" width={56} height={56} className="h-full w-full object-contain" />
                </span>
                <div className="min-w-0">
                  <p className="text-[10px] font-black uppercase tracking-[0.18em] text-lime-200">Isiloka PLN</p>
                  <h1 className="mt-1 text-xl font-black leading-tight tracking-tight">Listrik Lebih Mudah</h1>
                  <p className="mt-1 text-xs font-semibold leading-relaxed text-white/70">Token dan tagihan dalam satu tempat.</p>
                </div>
              </div>
            </section>

            <section className="grid gap-3">
              {cards.map((card) => {
                const Icon = card.icon;
              return (
                <Link
                  key={card.key}
                  href={card.href}
                  className="group relative overflow-hidden rounded-[28px] border border-white bg-white p-3 shadow-[0_16px_36px_rgba(6,78,59,0.10)] ring-1 ring-emerald-950/5 transition duration-300 hover:-translate-y-1 hover:shadow-[0_20px_42px_rgba(6,78,59,0.15)]"
                >
                  <div className={`absolute inset-y-0 right-0 w-32 bg-linear-to-br ${card.accent} opacity-[0.18] blur-xl transition group-hover:opacity-[0.28]`} />
                  <div className="relative flex items-center gap-3">
                    <span className={`grid h-13 w-13 shrink-0 place-items-center rounded-[20px] bg-linear-to-br ${card.accent} text-white shadow-[0_12px_24px_rgba(6,78,59,0.16)]`}>
                      <Icon className="h-6 w-6" />
                    </span>
                    <div className="min-w-0 flex-1">
                      <h2 className="text-base font-black tracking-tight text-slate-950">{card.title}</h2>
                      <p className="mt-0.5 text-xs font-semibold leading-relaxed text-slate-500">{card.description}</p>
                    </div>
                    <span className="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-lime-100 text-[#13464b] transition group-hover:translate-x-0.5 group-hover:bg-lime-300">
                      <ChevronRight className="h-4.5 w-4.5" strokeWidth={2.6} />
                    </span>
                  </div>
                </Link>
              );
            })}
            </section>
          </>
        )}
      </div>

      <GuestBottomNav isLoggedIn={!!session?.backendToken} />
    </main>
  );
}
