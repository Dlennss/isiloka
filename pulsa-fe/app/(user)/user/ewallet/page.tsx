import Image from "next/image";
import Link from "next/link";
import { getBrandsByKategori, getCategories } from "@/lib/api.products";
import type { UserBrandItem, UserCategoryItem } from "@/components/user/types";
import { UserBottomNav } from "@/components/user/UserBottomNav";

export const dynamic = "force-dynamic";

function normalizeName(value: string) {
  return value.trim().toLowerCase();
}

function pickCategory(categories: UserCategoryItem[], keyword: string) {
  return categories.find((item) => item.aktif && normalizeName(item.nama).includes(keyword));
}

function pickEwalletCategory(categories: UserCategoryItem[]) {
  return pickCategory(categories, "e-wallet") ?? pickCategory(categories, "e-money");
}

function pickBrand(brands: UserBrandItem[], patterns: string[]) {
  return brands.find((brand) => {
    const normalized = normalizeName(brand.nama);
    return patterns.some((pattern) => normalized.includes(pattern));
  }) ?? null;
}

function getEwalletImageSrc(key: string) {
  switch (key) {
    case "dana":
      return "/images/ewallet/logo_dana.png";
    case "gopay":
      return "/images/ewallet/logo_gopay.png";
    case "shopeepay":
      return "/images/ewallet/logo_shopee.png";
    case "linkaja":
      return "/images/ewallet/logo_linkaja.png";
    case "ovo":
      return "/images/ewallet/logo_ovo.png";
    case "astrapay":
      return "/images/ewallet/logo_astrapay.svg";
    case "isaku":
      return "/images/ewallet/logo_isaku.svg";
    default:
      return "";
  }
}

export default async function UserEwalletPage() {
  const categories = (await getCategories()) as UserCategoryItem[];
  const ewalletCategory = pickEwalletCategory(categories);
  const brands = ewalletCategory ? await getBrandsByKategori(String(ewalletCategory.id)) : [];

  const cards = [
    { key: "dana", title: "DANA", brand: pickBrand(brands, ["dana"]) },
    { key: "gopay", title: "GoPay", brand: pickBrand(brands, ["gopay", "go pay"]) },
    { key: "shopeepay", title: "ShopeePay", brand: pickBrand(brands, ["shopeepay", "shopee pay"]) },
    { key: "linkaja", title: "LinkAja", brand: pickBrand(brands, ["linkaja", "link aja"]) },
    { key: "ovo", title: "OVO", brand: pickBrand(brands, ["ovo"]) },
    { key: "astrapay", title: "AstraPay", brand: pickBrand(brands, ["astrapay", "astra pay"]) },
    { key: "isaku", title: "i.saku", brand: pickBrand(brands, ["i.saku", "isaku"]) },
  ].filter((item) => item.brand) as Array<{ key: string; title: string; brand: UserBrandItem }>;

  return (
    <main className="bg-sky-50">
      <div className="space-y-4 px-4 pt-4">
        {!ewalletCategory ? (
          <section className="grid min-h-40 place-items-center rounded-md border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_10px_28px_rgba(15,23,42,0.08)]">
            Kategori e-wallet belum tersedia.
          </section>
        ) : cards.length === 0 ? (
          <section className="grid min-h-40 place-items-center rounded-md border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_10px_28px_rgba(15,23,42,0.08)]">
            Provider e-wallet belum tersedia.
          </section>
        ) : (
          <section className="grid grid-cols-2 gap-3">
            {cards.map((card) => {
              const imageSrc = getEwalletImageSrc(card.key);
              return (
                <Link
                  key={card.key}
                  href={`/user/kategori/${ewalletCategory.id}/brand/${card.brand.id}?name=${encodeURIComponent(card.brand.nama)}`}
                  aria-label={card.title}
                  className="group rounded-md bg-white px-2 py-3 text-center shadow-[0_10px_28px_rgba(15,23,42,0.12)] ring-1 ring-slate-100 transition-transform duration-300 hover:-translate-y-1"
                >
                  <div className="flex flex-col items-center">
                    <div className="grid h-[100px] w-[100px] place-items-center overflow-hidden">
                      {imageSrc ? (
                        <Image src={imageSrc} alt={card.title} title={card.title} width={100} height={100} className="h-full w-full object-contain" />
                      ) : (
                        <span className="text-sm font-black uppercase tracking-tight text-sky-700">{card.title.slice(0, 2)}</span>
                      )}
                    </div>
                  </div>
                </Link>
              );
            })}
          </section>
        )}
      </div>

      <UserBottomNav />
    </main>
  );
}
