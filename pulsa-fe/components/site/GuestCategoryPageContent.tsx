import Image from "next/image";
import Link from "next/link";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/nextauth";
import { getBrandsByKategori } from "@/lib/api.products";
import type { UserSession } from "@/components/user/types";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";
import { getBrandLogo } from "@/lib/brand-logos";
import { getGameBrandSlug } from "@/lib/game-brand-routes";
import { GameCategoryPicker } from "@/components/site/GameCategoryPicker";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

type GuestCategoryPageContentProps = {
  kategoriId: string;
  title: string;
};

const GAME_BRAND_ORDER: Record<string, number> = {
  "mobile legend": 1,
  "mobile legends": 1,
  "mobile legends bang bang": 1,
  "free fire": 2,
  "magic chess go go": 3,
  "call of duty mobile": 4,
  "genshin impact": 5,
  "pubg mobile": 6,
  roblox: 7,
  "point blank": 8,
};

const HIDDEN_GAME_BRANDS = new Set([
  "blood strike",
  "free fire max",
  "hago",
  "honkai star rail",
  "zepeto",
]);

function normalizeName(value: string) {
  return value.trim().toLowerCase();
}

function sortGameBrands<T extends { nama: string }>(items: T[]) {
  return [...items].sort((a, b) => {
    const pa = GAME_BRAND_ORDER[normalizeName(a.nama)] ?? 999;
    const pb = GAME_BRAND_ORDER[normalizeName(b.nama)] ?? 999;
    if (pa !== pb) return pa - pb;
    return a.nama.localeCompare(b.nama, "id-ID");
  });
}

function getGameDisplayName(name: string) {
  const normalized = normalizeName(name);
  if (normalized === "mobile legend" || normalized === "mobile legends" || normalized === "mobile legends bang bang") {
    return "Mobile Legends: Bang Bang";
  }
  if (normalized === "free fire") return "Free Fire Top-up";
  if (normalized === "magic chess go go") return "Magic Chess: Go Go";
  return name;
}

export async function GuestCategoryPageContent({
  kategoriId,
  title,
}: GuestCategoryPageContentProps) {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const brands = await getBrandsByKategori(kategoriId);
  const isGameCategory = String(kategoriId) === "5";

  if (isGameCategory) {
    return (
      <main className="min-h-screen bg-[#eef7f3] pb-24">
        <div className="px-4 pt-4">
          <GameCategoryPicker
            title={title}
            items={sortGameBrands(brands)
              .filter((brand) => brand.aktif !== false && !HIDDEN_GAME_BRANDS.has(normalizeName(brand.nama)))
              .map((brand) => {
                const logo = getBrandLogo(brand.nama);
                return {
                  id: brand.id,
                  name: getGameDisplayName(brand.nama),
                  href: `/game/${getGameBrandSlug(brand.nama)}`,
                  logo: logo ? { src: logo.src, alt: logo.alt || brand.nama } : null,
                };
              })}
          />
        </div>
        <GuestBottomNav isLoggedIn={!!session?.backendToken} />
      </main>
    );
  }

  return (
    <main className="bg-sky-50">
      <div className="space-y-4 px-4 pt-4">
        <section>
          {brands.length === 0 ? (
            <div className="grid min-h-40 place-items-center rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 text-center text-sm text-slate-500">
              Belum ada brand aktif untuk kategori ini.
            </div>
          ) : (
            <div className="grid grid-cols-3 gap-3 sm:grid-cols-4">
              {brands.map((brand) => {
                const logo = getBrandLogo(brand.nama);
                return (
                  <Link
                    key={brand.id}
                    href={`/kategori/${kategoriId}/brand/${brand.id}?name=${brand.nama}`}
                    aria-label={brand.nama}
                    className="group relative flex h-27 items-center justify-center overflow-hidden rounded-3xl border border-slate-200 bg-white px-3 py-4 text-center shadow-[0_16px_38px_rgba(15,23,42,0.10)] transition-all duration-300 hover:-translate-y-1 hover:border-sky-300 hover:shadow-[0_16px_38px_rgba(15,23,42,0.14)]"
                  >
                    <div className="absolute inset-x-3 top-3 h-14 rounded-2xl bg-linear-to-br from-sky-200 via-cyan-100 to-blue-100 opacity-80 blur-xl transition-opacity duration-300 group-hover:opacity-100" />
                    <div className="relative grid h-16 w-16 shrink-0 place-items-center overflow-hidden  border border-sky-100 bg-sky-50 p-3 shadow-sm">
                      {logo ? (
                        <Image
                          src={logo.src}
                          alt={logo.alt}
                          title={logo.alt}
                          width={56}
                          height={56}
                          className="h-full w-full object-contain"
                        />
                      ) : (
                        <span className="text-base font-black uppercase tracking-tight text-sky-700">
                          {brand.nama.slice(0, 2)}
                        </span>
                      )}
                    </div>
                  </Link>
                );
              })}
            </div>
          )}
        </section>
      </div>

      <GuestBottomNav isLoggedIn={!!session?.backendToken} />
    </main>
  );
}
