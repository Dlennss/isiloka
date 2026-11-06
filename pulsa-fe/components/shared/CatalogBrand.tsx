import Link from "next/link";
import Image from "next/image";
import { ArrowLeft } from "lucide-react";
import { getProductsByBrand } from "@/lib/api.products";
import { getBrandLogo } from "@/lib/brand-logos";
import { CatalogProductList } from "@/components/shared/CatalogProductList";
import { RetailBillingEntryFlow } from "@/components/shared/RetailBillingEntryFlow";
import { BPJSBrandFlow } from "@/components/shared/BPJSBrandFlow";

export async function CatalogBrand({ id, brandId, mode = "guest", authToken, buyerRole }: {
  id: string; brandId: string; mode?: "guest" | "user"; authToken?: string; buyerRole?: string;
}) {
  const items = await getProductsByBrand(id, brandId);
  const category = items[0]?.kategori_nama || "";
  const brand = items[0]?.brand_nama || "";
  const logo = getBrandLogo(brand);
  const billing = ["Tagihan Air", "Tagihan Gas", "Pascabayar", "Internet & Telco", "TV & Streaming"].includes(category);
  return <div className="space-y-4 px-4 pt-4 pb-24">
    <div className="flex items-center justify-between gap-3">
      <Link href={`${mode === "user" ? "/user" : ""}/kategori/${id}`} title="Kembali ke provider" aria-label="Kembali ke provider" className="grid h-10 w-10 place-items-center rounded-md border border-slate-200 bg-white"><ArrowLeft className="h-5 w-5" /></Link>
      {logo ? <Image src={logo.src} alt={brand} width={80} height={48} className="h-12 w-20 object-contain" /> : null}
    </div>
    {items.length === 0 ? <p className="py-8 text-center text-sm text-slate-600">Produk tidak tersedia.</p>
      : category === "BPJS" ? <BPJSBrandFlow items={items} authToken={authToken} buyerRole={buyerRole} initialMode={brand.toLowerCase().includes("ketenagakerjaan") ? "ketenagakerjaan" : "kesehatan"} showModeSelector={false} />
      : billing ? <RetailBillingEntryFlow title={brand} description={category} placeholder="ID / nomor pelanggan" items={items} mode={mode} authToken={authToken} buyerRole={buyerRole} />
      : <CatalogProductList items={items} authToken={authToken} buyerRole={buyerRole} />}
  </div>;
}
