import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { notFound } from "next/navigation";
import { getBrandsByKategori, getCategories } from "@/lib/api.products";
import type { UserCategoryItem } from "@/components/user/types";
import { ProviderBrandPicker } from "@/components/guest/ProviderBrandPicker";

export async function CatalogCategory({ id, mode = "guest" }: { id: string; mode?: "guest" | "user" }) {
  const categories = await getCategories() as UserCategoryItem[];
  const category = categories.find((item) => String(item.id) === id && item.aktif);
  if (!category) notFound();
  const brands = await getBrandsByKategori(id);
  const prefix = mode === "user" ? "/user/kategori" : "/kategori";
  return (
    <section className="space-y-5 px-4 pt-4 pb-28">
      <div className="flex items-center gap-3">
        <Link href={prefix} aria-label="Kembali ke layanan" title="Kembali ke layanan" className="grid h-10 w-10 shrink-0 place-items-center rounded-md border border-slate-200 bg-white"><ArrowLeft className="h-5 w-5" /></Link>
        <h1 className="min-w-0 break-words text-xl font-bold text-slate-900">{category.nama}</h1>
      </div>
      <ProviderBrandPicker items={brands.filter((brand) => brand.aktif).map((brand) => ({ brand, href: `${prefix}/${id}/brand/${brand.id}` }))} />
      {brands.length === 0 ? <p role="status" className="py-8 text-center text-sm text-slate-600">Belum ada provider aktif.</p> : null}
    </section>
  );
}
