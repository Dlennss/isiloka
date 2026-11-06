"use client";

import { useState } from "react";
import { Search } from "lucide-react";
import type { UserCategoryItem } from "@/components/user/types";
import { CategoryShortcutLink } from "@/components/shared/CategoryShortcutLink";

export function ServiceDirectory({ categories, mode = "guest" }: {
  categories: UserCategoryItem[];
  mode?: "guest" | "user";
}) {
  const [query, setQuery] = useState("");
  const items = categories.filter((item) => item.aktif && item.nama.toLocaleLowerCase("id").includes(query.trim().toLocaleLowerCase("id")));
  const prefix = mode === "user" ? "/user/kategori" : "/kategori";
  return (
    <section className="space-y-5 pb-28">
      <h1 className="text-xl font-bold text-slate-900">Semua Layanan</h1>
      <label className="flex h-12 items-center gap-3 rounded-md border border-slate-200 bg-white px-3">
        <Search aria-hidden="true" className="h-5 w-5 shrink-0 text-slate-500" />
        <input aria-label="Cari layanan" placeholder="Cari layanan" value={query} onChange={(event) => setQuery(event.target.value)} className="min-w-0 flex-1 bg-transparent text-sm outline-none" />
      </label>
      <div className="grid grid-cols-3 gap-x-3 gap-y-5">
        {items.map((item) => <CategoryShortcutLink key={item.id} href={`${prefix}/${item.id}`} label={item.nama} visualName={item.nama} />)}
      </div>
      {items.length === 0 ? <p role="status" className="py-8 text-center text-sm text-slate-600">Layanan tidak tersedia.</p> : null}
    </section>
  );
}
