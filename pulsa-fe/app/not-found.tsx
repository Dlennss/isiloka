import Link from "next/link";

export default function NotFound() {
  return (
    <main className="min-h-[100svh] bg-[#f0f8f8] px-4 py-12 text-slate-950">
      <section className="mx-auto flex min-h-[calc(100svh-6rem)] w-full max-w-md items-center">
        <div className="w-full rounded-md border border-emerald-950/10 bg-white p-6 shadow-[0_24px_70px_rgba(6,78,59,0.14)]">
          <p className="inline-flex rounded-full bg-emerald-50 px-3 py-1 text-xs font-semibold uppercase tracking-[0.16em] text-emerald-700">
            Error 404
          </p>

          <div className="mt-5 space-y-4">
            <h1 className="text-3xl font-black leading-tight text-slate-950">
              Halaman tidak tersedia
            </h1>
            <p className="text-sm leading-6 text-slate-600">
              Link ini mungkin sudah berubah. Anda tetap dapat melihat dan membeli produk Isiloka tanpa membuat akun.
            </p>
          </div>

          <div className="mt-7 grid gap-3">
            <Link href="/pulsa-data" className="inline-flex h-12 items-center justify-center rounded-md bg-emerald-700 px-5 text-sm font-bold text-white hover:bg-emerald-800">
              Beli Pulsa & Paket Data
            </Link>
            <Link href="/" className="inline-flex h-12 items-center justify-center rounded-md border border-emerald-200 bg-white px-5 text-sm font-bold text-emerald-800 hover:bg-emerald-50">
              Kembali ke Beranda
            </Link>
          </div>
        </div>
      </section>
    </main>
  );
}
