import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import { buildBreadcrumbJsonLd, buildCollectionJsonLd, buildPageMetadata } from "@/lib/site-search";

const BPJS_OPTIONS = [
  {
    slug: "kesehatan",
    title: "BPJS Kesehatan",
    logo: "/images/bpjs/icon_bpjs_kesehatan.png",
  },
  {
    slug: "ketenagakerjaan",
    title: "BPJS Ketenagakerjaan",
    logo: "/images/bpjs/icon_bpjs_ketenagakerjaan.png",
  },
];

export const metadata: Metadata = buildPageMetadata({
  title: "BPJS | Isiloka",
  description: "Cek dan bayar BPJS Kesehatan dan BPJS Ketenagakerjaan di Isiloka dengan alur yang ringkas.",
  path: "/bpjs",
  keywords: ["bpjs kesehatan", "bpjs ketenagakerjaan", "bayar bpjs", "pulsakilat"],
});

export default function BPJSPage() {
  const collectionJsonLd = buildCollectionJsonLd({
    title: "BPJS | Isiloka",
    description: "Cek dan bayar BPJS Kesehatan dan BPJS Ketenagakerjaan di Isiloka dengan alur yang ringkas.",
    path: "/bpjs",
    itemNames: BPJS_OPTIONS.map((item) => item.title),
  });
  const breadcrumbJsonLd = buildBreadcrumbJsonLd([
    { name: "Beranda", path: "/" },
    { name: "BPJS", path: "/bpjs" },
  ]);

  return (
    <main className="bg-sky-50">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(collectionJsonLd) }} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
      <div className="space-y-4 px-4 pt-4">
        <section className="grid grid-cols-1 gap-3">
          {BPJS_OPTIONS.map((item) => (
            <Link
              key={item.slug}
              href={`/bpjs/${item.slug}`}
              className="flex items-center gap-3 border border-slate-200 bg-white px-4 py-4 shadow-[0_8px_22px_rgba(15,23,42,0.08)] transition hover:border-sky-300 hover:shadow-[0_12px_26px_rgba(15,23,42,0.12)]"
            >
              <div className="grid h-12 w-12 shrink-0 place-items-center overflow-hidden">
                <Image src={item.logo} alt={item.title} width={48} height={48} className="h-full w-full object-contain" />
              </div>
              <p className="text-base font-semibold text-slate-900">{item.title}</p>
            </Link>
          ))}
        </section>
      </div>
    </main>
  );
}
