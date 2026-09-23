"use client";

import { useEffect, useMemo, useState } from "react";
import { PackageSearch, Search } from "lucide-react";

type H2HProductRow = {
  id: number;
  sku: string;
  nama: string;
  group_name?: string;
  kategori_nama?: string;
  brand_nama?: string;
  tipe_harga?: string;
  harga?: number | null;
  fee_tambahan?: number | null;
  maksimal_nominal?: number | null;
};

type ProductResponse = {
  ok?: boolean;
  items?: H2HProductRow[];
  error?: string;
};

type Props = {
  authToken: string;
};

function fmtIDR(v: number) {
  return new Intl.NumberFormat("id-ID").format(Number(v || 0));
}

export function H2HProductClient({ authToken }: Props) {
  const [query, setQuery] = useState("");
  const [selectedKategori, setSelectedKategori] = useState("");
  const [selectedBrand, setSelectedBrand] = useState("");
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState("");
  const [items, setItems] = useState<H2HProductRow[]>([]);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setErr("");
      try {
        const params = new URLSearchParams();
        if (query.trim()) params.set("q", query.trim());
        if (selectedKategori) params.set("kategori", selectedKategori);
        if (selectedBrand) params.set("brand", selectedBrand);
        const r = await fetch(`/api/me/h2h/produk?${params.toString()}`, {
          headers: { Authorization: `Bearer ${authToken}` },
          cache: "no-store",
        });
        const j = (await r.json().catch(() => ({}))) as ProductResponse;
        if (!r.ok || !j?.ok) throw new Error(j?.error || "Gagal memuat produk H2H.");
        if (!cancelled) setItems(Array.isArray(j.items) ? j.items : []);
      } catch (e) {
        if (!cancelled) setErr(e instanceof Error ? e.message : "Gagal memuat produk H2H.");
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    void load();
    return () => {
      cancelled = true;
    };
  }, [authToken, query, selectedKategori, selectedBrand]);

  const stats = useMemo(() => {
    const priced = items.filter((item) => typeof item.harga === "number");
    return {
      total: items.length,
      priced: priced.length,
    };
  }, [items]);

  const groupedItems = useMemo(() => {
    const kategoriMap = new Map<
      string,
      Map<string, H2HProductRow[]>
    >();

    const sorted = [...items].sort((a, b) => {
      const kategoriCmp = String(a.kategori_nama || "").localeCompare(String(b.kategori_nama || ""), "id");
      if (kategoriCmp !== 0) return kategoriCmp;
      const brandCmp = String(a.brand_nama || "").localeCompare(String(b.brand_nama || ""), "id");
      if (brandCmp !== 0) return brandCmp;
      const skuCmp = String(a.sku || "").localeCompare(String(b.sku || ""), "id");
      if (skuCmp !== 0) return skuCmp;
      const hargaA = typeof a.harga === "number" ? a.harga : Number.MAX_SAFE_INTEGER;
      const hargaB = typeof b.harga === "number" ? b.harga : Number.MAX_SAFE_INTEGER;
      if (hargaA !== hargaB) return hargaA - hargaB;
      return String(a.nama || "").localeCompare(String(b.nama || ""), "id");
    });

    for (const item of sorted) {
      const kategori = String(item.kategori_nama || "Lainnya");
      const brand = String(item.brand_nama || "Tanpa Brand");
      if (!kategoriMap.has(kategori)) kategoriMap.set(kategori, new Map<string, H2HProductRow[]>());
      const brandMap = kategoriMap.get(kategori)!;
      if (!brandMap.has(brand)) brandMap.set(brand, []);
      brandMap.get(brand)!.push(item);
    }

    return Array.from(kategoriMap.entries()).map(([kategori, brandMap]) => ({
      kategori,
      brands: Array.from(brandMap.entries()).map(([brand, rows]) => ({
        brand,
        rows,
      })),
    }));
  }, [items]);

  const kategoriOptions = useMemo(
    () => Array.from(new Set(items.map((item) => String(item.kategori_nama || "").trim()).filter(Boolean))).sort((a, b) => a.localeCompare(b, "id")),
    [items],
  );

  const brandOptions = useMemo(
    () => Array.from(new Set(items.map((item) => String(item.brand_nama || "").trim()).filter(Boolean))).sort((a, b) => a.localeCompare(b, "id")),
    [items],
  );

  return (
    <div className="space-y-4">
      <section className="rounded-2xl border border-white/10 bg-card/80 p-5 shadow-[0_16px_40px_-26px_rgba(0,0,0,0.8)]">
        <div className="flex items-center gap-3">
          <span className="grid h-11 w-11 place-items-center rounded-2xl border border-sky-400/20 bg-sky-400/10 text-sky-200">
            <PackageSearch className="h-5 w-5" />
          </span>
          <div>
            <h1 className="text-lg font-bold text-white">Produk H2H</h1>
            <p className="text-sm text-white/60">Daftar SKU internal dan harga jual member H2H. Harga mengikuti harga operator Yuscom ditambah fee kategori member.</p>
          </div>
        </div>
      </section>

      {err ? <div className="rounded-2xl border border-rose-400/25 bg-rose-500/10 px-4 py-3 text-sm text-rose-300">{err}</div> : null}

      <section className="grid gap-3 md:grid-cols-2">
        <div className="rounded-2xl border border-white/10 bg-card/80 p-4 shadow-[0_16px_40px_-26px_rgba(0,0,0,0.8)]">
          <div className="text-xs font-semibold uppercase tracking-wide text-white/45">Total Produk Aktif</div>
          <div className="mt-2 text-lg font-bold text-white">{stats.total}</div>
        </div>
        <div className="rounded-2xl border border-white/10 bg-card/80 p-4 shadow-[0_16px_40px_-26px_rgba(0,0,0,0.8)]">
          <div className="text-xs font-semibold uppercase tracking-wide text-white/45">Produk Dengan Harga Tetap</div>
          <div className="mt-2 text-lg font-bold text-white">{stats.priced}</div>
        </div>
      </section>

      <section className="rounded-2xl border border-white/10 bg-card/80 p-5 shadow-[0_16px_40px_-26px_rgba(0,0,0,0.8)]">
        <div className="flex flex-col gap-3">
          <div>
            <h2 className="font-bold text-white">Daftar Produk</h2>
            <p className="text-sm text-white/60">Gunakan SKU internal ini untuk transaksi H2H.</p>
          </div>
          <div className="grid gap-3 md:grid-cols-3">
            <label className="flex items-center gap-2 rounded-2xl border border-white/10 bg-slate-950/35 px-3 py-2 text-sm text-white/70">
              <Search className="h-4 w-4 text-white/45" />
              <input
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Cari SKU / nama produk"
                className="w-full bg-transparent outline-hidden placeholder:text-white/35"
              />
            </label>
            <select
              value={selectedKategori}
              onChange={(e) => setSelectedKategori(e.target.value)}
              className="rounded-2xl border border-white/10 bg-slate-950/35 px-3 py-2 text-sm text-white outline-hidden"
            >
              <option value="">Semua kategori</option>
              {kategoriOptions.map((item) => (
                <option key={item} value={item}>
                  {item}
                </option>
              ))}
            </select>
            <select
              value={selectedBrand}
              onChange={(e) => setSelectedBrand(e.target.value)}
              className="rounded-2xl border border-white/10 bg-slate-950/35 px-3 py-2 text-sm text-white outline-hidden"
            >
              <option value="">Semua brand</option>
              {brandOptions.map((item) => (
                <option key={item} value={item}>
                  {item}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="mt-4 space-y-3">
          {loading ? <div className="text-sm text-white/60">Memuat produk H2H...</div> : null}
          {!loading && items.length === 0 ? <div className="text-sm text-white/60">Belum ada produk aktif yang cocok.</div> : null}
          {groupedItems.map((kategoriGroup) => (
            <div key={kategoriGroup.kategori} className="rounded-2xl border border-white/10 bg-slate-950/25 p-4">
              <div className="mb-3 text-sm font-bold uppercase tracking-wide text-cyan-200/90">{kategoriGroup.kategori}</div>
              <div className="space-y-4">
                {kategoriGroup.brands.map((brandGroup) => (
                  <div key={`${kategoriGroup.kategori}-${brandGroup.brand}`} className="rounded-2xl border border-white/8 bg-slate-950/35 p-3">
                    <div className="mb-3 text-sm font-semibold text-white">{brandGroup.brand}</div>
                    <div className="overflow-hidden rounded-2xl border border-white/10">
                      <div className="grid grid-cols-[140px_minmax(0,1fr)_140px] gap-3 bg-white/5 px-4 py-2 text-[11px] font-semibold uppercase tracking-wide text-white/45">
                        <div>SKU</div>
                        <div>Keterangan</div>
                        <div className="text-right">Harga</div>
                      </div>
                      {brandGroup.rows.map((item) => (
                        <div key={item.id} className="grid grid-cols-[140px_minmax(0,1fr)_140px] gap-3 border-t border-white/10 bg-slate-950/35 px-4 py-3">
                          <div className="text-sm font-semibold text-cyan-200">{item.sku}</div>
                          <div className="min-w-0">
                            <div className="truncate text-sm font-semibold text-white">{item.nama}</div>
                            <div className="mt-1 text-xs text-white/55">{item.group_name || "-"}</div>
                            <div className="mt-1 text-xs text-white/45">
                              {item.tipe_harga || "-"}
                              {String(item.tipe_harga || "").toUpperCase() === "OPEN_AMOUNT" && typeof item.maksimal_nominal === "number"
                                ? ` • Maks ${fmtIDR(item.maksimal_nominal)}`
                                : ""}
                            </div>
                          </div>
                          <div className="text-right text-base font-bold text-white">
                            {typeof item.harga === "number"
                              ? `Rp ${fmtIDR(item.harga)}`
                              : String(item.tipe_harga || "").toUpperCase() === "OPEN_AMOUNT"
                                ? `+ Rp ${fmtIDR(typeof item.fee_tambahan === "number" ? item.fee_tambahan : 0)}`
                                : "-"}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}
