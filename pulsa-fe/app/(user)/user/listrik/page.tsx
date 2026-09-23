import Link from "next/link";
import { Bolt, ReceiptText } from "lucide-react";
import { getCategories } from "@/lib/api.products";
import type { UserCategoryItem } from "@/components/user/types";
import { UserBottomNav } from "@/components/user/UserBottomNav";

export const dynamic = "force-dynamic";

function normalizeName(value: string) {
  return value.trim().toLowerCase();
}

function pickCategory(categories: UserCategoryItem[], keyword: string) {
  return categories.find((item) => item.aktif && normalizeName(item.nama).includes(keyword));
}

export default async function UserListrikPage() {
  const categories = (await getCategories()) as UserCategoryItem[];
  const listrikCategory = pickCategory(categories, "listrik");

  const cards = [
    { key: "token", title: "Token Listrik", href: "/user/listrik/token", icon: Bolt, accent: "from-amber-400 via-yellow-400 to-orange-500" },
    { key: "tagihan", title: "Tagihan Listrik", href: "/user/listrik/tagihan", icon: ReceiptText, accent: "from-sky-500 via-cyan-500 to-blue-600" },
  ];

  return (
    <main className="bg-sky-50">
      <div className="space-y-4 px-4 pt-4">
        {!listrikCategory ? (
          <section className="grid min-h-40 place-items-center rounded-md border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_10px_28px_rgba(15,23,42,0.08)]">
            Kategori listrik belum tersedia.
          </section>
        ) : (
          <section className="grid grid-cols-2 gap-3">
            {cards.map((card) => {
              const Icon = card.icon;
              return (
                <Link
                  key={card.key}
                  href={card.href}
                  className="group rounded-md bg-white px-3 py-4 text-center shadow-[0_10px_28px_rgba(15,23,42,0.12)] ring-1 ring-slate-100 transition-transform duration-300 hover:-translate-y-1"
                >
                  <div className="flex flex-col items-center gap-3">
                    <div className={`grid h-12 w-12 shrink-0 place-items-center rounded-full bg-linear-to-br ${card.accent} text-white shadow-[0_12px_26px_rgba(14,165,233,0.18)]`}>
                      <Icon className="h-5 w-5" />
                    </div>
                    <div className="min-w-0">
                      <h2 className="text-base font-bold text-slate-900">{card.title}</h2>
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
