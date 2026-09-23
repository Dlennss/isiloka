"use client";

import * as React from "react";
import { useRouter } from "next/navigation";
import { ArrowRight, Flame, Smartphone } from "lucide-react";
import { USER_MENU_ITEMS, type UserMenuKey } from "@/components/user/menu-items";

type HomeCategory = UserMenuKey;

const fieldLabel: Record<HomeCategory, string> = {
  pulsa: "Nomor HP",
  paket_data: "Nomor HP",
  listrik_prabayar: "Nomor Meter",
  e_wallet: "Nomor E-Wallet",
  gaming: "User ID Game",
  semua: "Nomor Tujuan",
};

const nominalByCategory: Record<HomeCategory, string[]> = {
  pulsa: ["10000", "20000", "50000", "100000"],
  paket_data: ["10000", "25000", "50000", "100000"],
  listrik_prabayar: ["20000", "50000", "100000", "200000"],
  e_wallet: ["20000", "50000", "100000", "200000"],
  gaming: ["10000", "20000", "50000", "100000"],
  semua: ["10000", "25000", "50000", "100000"],
};

function toCheckoutCategory(cat: HomeCategory) {
  if (cat === "paket_data") return "paket_data";
  if (cat === "listrik_prabayar") return "token_listrik";
  if (cat === "e_wallet") return "ovo";
  if (cat === "gaming") return "games";
  return "pulsa";
}

export function HomeQuickOrder() {
  const router = useRouter();
  const [category, setCategory] = React.useState<HomeCategory>("listrik_prabayar");
  const [dest, setDest] = React.useState("");
  const [nominal, setNominal] = React.useState("20000");

  React.useEffect(() => {
    const list = nominalByCategory[category];
    if (!list.includes(nominal)) setNominal(list[0]);
  }, [category, nominal]);

  function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const clean = dest.trim();
    if (!clean) return;
    const qs = new URLSearchParams({
      cat: toCheckoutCategory(category),
      dest: clean,
      nominal,
    });
    router.push(`/beli?${qs.toString()}`);
  }

  return (
    <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-[0_10px_30px_rgba(15,23,42,0.14)]">
      <div className="overflow-x-auto border-b border-slate-200">
        <div className="flex min-w-max">
        {USER_MENU_ITEMS.map((item) => {
          const Icon = item.icon;
          const active = category === item.key;
          return (
              <button
                key={item.key}
                type="button"
                onClick={() => setCategory(item.key)}
                className={`shrink-0 border-r border-slate-200 px-4 py-3 text-sm font-medium last:border-r-0 md:px-5 md:py-3 ${
                  active ? "bg-sky-50 text-slate-800" : "text-slate-500"
                }`}
              >
                <span className={`mr-2 inline-grid h-8 w-8 place-items-center rounded-full ${active ? "bg-[#c9ecff] text-[#2d8fdc]" : "bg-slate-100 text-slate-400"}`}>
                  <Icon className="h-4 w-4" />
                </span>
                {item.title}
              </button>
          );
        })}
        </div>
      </div>

      <form onSubmit={onSubmit} className="grid gap-4 p-4 md:grid-cols-[1.15fr_0.85fr] md:p-5 lg:p-6">
        <label className="flex min-h-12 items-center justify-between rounded-lg border border-slate-300 px-4 py-3">
          <input
            value={dest}
            onChange={(e) => setDest(e.target.value)}
            placeholder={fieldLabel[category]}
            className="w-full border-0 bg-transparent text-sm text-slate-700 outline-none placeholder:text-slate-400 md:text-base lg:text-lg"
            required
          />
          <Smartphone className="h-5 w-5 text-slate-500" />
        </label>

        <label className="flex min-h-12 items-center justify-between rounded-lg border border-slate-300 px-4 py-3">
          <select
            value={nominal}
            onChange={(e) => setNominal(e.target.value)}
            className="w-full appearance-none border-0 bg-transparent text-sm text-slate-700 outline-none md:text-base lg:text-lg"
          >
            {nominalByCategory[category].map((x) => (
              <option key={x} value={x}>
                Rp {Number(x).toLocaleString("id-ID")}
              </option>
            ))}
          </select>
          <ArrowRight className="h-5 w-5 text-slate-400 md:h-7 md:w-7" />
        </label>

        <div className="md:col-span-2">
          <button type="submit" className="inline-flex w-full items-center justify-center gap-2 rounded-full bg-[#2d8fdc] px-6 py-3 text-sm font-semibold text-white md:w-auto">
            <Flame className="h-4 w-4" />
            Beli
          </button>
        </div>
      </form>
    </div>
  );
}
