import Image from "next/image";
import Link from "next/link";
import { ShoppingBag } from "lucide-react";
import { getAppServerSession } from "@/lib/server-auth";
import { getBrandsByKategori } from "@/lib/api.products";
import type { UserSession } from "@/components/user/types";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import { UserAuthClientSync } from "@/components/user/UserAuthClientSync";
import { getBrandLogo } from "@/lib/brand-logos";
import { GuestPulsaQuickOrder } from "@/components/guest/GuestPulsaQuickOrder";
import { UserUniversalServicePageContent } from "@/components/user/UserUniversalServicePageContent";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

type PageProps = {
  params: Promise<{ id: string }>;
};

const UNIVERSAL_SERVICE_BY_CATEGORY_ID: Record<string, string> = {
  "18": "hp-pascabayar",
  "hp-pascabayar": "hp-pascabayar",
  "esim-roaming": "esim-roaming",
};

export default async function UserKategoriPage({ params }: PageProps) {
  const session = (await getAppServerSession()) as SessionShape | null;
  const backendToken = session?.backendToken;

  const { id } = await params;
  const universalService = UNIVERSAL_SERVICE_BY_CATEGORY_ID[String(id)];
  if (universalService) {
    return (
      <>
        {backendToken ? <UserAuthClientSync backendToken={backendToken} /> : null}
        <UserUniversalServicePageContent serviceSlug={universalService} />
        <UserBottomNav />
      </>
    );
  }

  const brands = await getBrandsByKategori(id);

  return (
    <main className="min-h-screen bg-sky-50">
      {backendToken ? <UserAuthClientSync backendToken={backendToken} /> : null}
      <div className=" space-y-4 px-4">
        {String(id) === "1" ? (
          <GuestPulsaQuickOrder
            kategoriId={String(id)}
            brands={brands}
            authToken={backendToken}
            buyerRole={session?.user?.role}
          />
        ) : (
          <>
            <section className="rounded-xl bg-white p-5 shadow-[0_8px_22px_rgba(15,23,42,0.13)]">
              <div className="flex items-start justify-between gap-3">
                <div />
                <div className="grid h-11 w-11 place-items-center rounded-2xl bg-sky-50 text-sky-600">
                  <ShoppingBag className="h-5 w-5" />
                </div>
              </div>
            </section>
            <section>
              {brands.length === 0 ? (
                <div className="grid min-h-40 place-items-center rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 text-center text-sm text-slate-500">
                  Belum ada brand aktif untuk kategori ini.
                </div>
              ) : (
                <div className="grid grid-cols-3 gap-x-3 gap-y-4 sm:grid-cols-4">
                  {brands.map((brand) => (
                    (() => {
                      const logo = getBrandLogo(brand.nama);
                      return (
                        <Link
                          key={brand.id}
                          href={`/user/kategori/${id}/brand/${brand.id}?name=${brand.nama}`}
                          aria-label={brand.nama}
                          className="group flex flex-col items-center gap-2 text-center transition-transform duration-200 hover:-translate-y-1"
                        >
                          <div className="grid h-16 w-16 shrink-0 place-items-center overflow-hidden rounded-[1.4rem] bg-transparent">
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
                          <span className="line-clamp-2 text-xs font-semibold leading-tight text-slate-700">
                            {brand.nama}
                          </span>
                        </Link>
                      );
                    })()
                  ))}
                </div>
              )}
            </section>
          </>
        )}
      </div>

      <UserBottomNav />
    </main>
  );
}
