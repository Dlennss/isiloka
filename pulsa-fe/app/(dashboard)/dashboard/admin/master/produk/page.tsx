"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableActions, type DataTableColumn } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { AppModal } from "@/components/ui/app-modal";
import { alertConfirm, alertError, alertSuccess, alertWarning } from "@/components/ui/alerts";
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, Eye, Pencil, Plus, RefreshCw, SlidersHorizontal, Trash2 } from "lucide-react";

type MasterRow = { id: number; nama: string; aktif: boolean };

type ProdukRow = {
  id: number;
  sku: string;
  nama: string;
  group_name: string;
  kategori_id: number;
  kategori_nama: string;
  brand_id: number;
  brand_nama: string;
  tipe_harga: "FIXED" | "OPEN_AMOUNT";
  nominal?: number | null;
  maksimal_nominal?: number | null;
  jam_buka?: string;
  jam_tutup?: string;
  aktif: boolean;
};

type FormState = {
  sku: string;
  nama: string;
  group_name: string;
  kategori_id: string;
  brand_id: string;
  tipe_harga: "FIXED" | "OPEN_AMOUNT";
  nominal: string;
  maksimal_nominal: string;
  jam_buka: string;
  jam_tutup: string;
  aktif: boolean;
};

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

function toNumOrNull(v: string): number | null {
  const n = Number(String(v || "").replace(/[^\d-]/g, ""));
  if (!Number.isFinite(n) || n <= 0) return null;
  return Math.floor(n);
}

export default function AdminMasterProdukPage() {
  const [items, setItems] = useState<ProdukRow[]>([]);
  const [kategori, setKategori] = useState<MasterRow[]>([]);
  const [brand, setBrand] = useState<MasterRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [qInput, setQInput] = useState("");
  const [q, setQ] = useState("");
  const [skuInput, setSkuInput] = useState("");
  const [sku, setSku] = useState("");
  const [groupOptions, setGroupOptions] = useState<string[]>([]);
  const [groupName, setGroupName] = useState("");
  const [kategoriFilter, setKategoriFilter] = useState("");
  const [brandFilter, setBrandFilter] = useState("");
  const [groupDropdownOpen, setGroupDropdownOpen] = useState(false);
  const [groupSearch, setGroupSearch] = useState("");
  const [mobileFilterOpen, setMobileFilterOpen] = useState(false);
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(10);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(1);

  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<ProdukRow | null>(null);
  const [saving, setSaving] = useState(false);
  const [openActionKey, setOpenActionKey] = useState<string | null>(null);

  const [form, setForm] = useState<FormState>({
    sku: "",
    nama: "",
    group_name: "",
    kategori_id: "",
    brand_id: "",
    tipe_harga: "FIXED",
    nominal: "",
    maksimal_nominal: "",
    jam_buka: "00:31",
    jam_tutup: "23:29",
    aktif: true,
  });

  const activeKategori = useMemo(() => kategori.filter((x) => x.aktif), [kategori]);
  const activeBrand = useMemo(() => brand.filter((x) => x.aktif), [brand]);

  async function loadMasters() {
    const [rk, rb] = await Promise.all([
      fetch("/api/admin/master/kategori", { headers: authHeader(), cache: "no-store" }),
      fetch("/api/admin/master/brand", { headers: authHeader(), cache: "no-store" }),
    ]);

    const jk = await rk.json().catch(() => ({}));
    const jb = await rb.json().catch(() => ({}));

    if (rk.ok && jk.ok) setKategori(Array.isArray(jk.items) ? jk.items : []);
    if (rb.ok && jb.ok) setBrand(Array.isArray(jb.items) ? jb.items : []);
  }

  async function loadGroupOptions() {
    const r = await fetch("/api/admin/master/produk?groups=1", { headers: authHeader(), cache: "no-store" });
    const j = await r.json().catch(() => ({}));
    if (r.ok && j.ok && Array.isArray(j.items)) {
      setGroupOptions(j.items.filter((x: unknown): x is string => typeof x === "string" && x.trim() !== ""));
    }
  }

  const loadProduk = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (q) params.set("q", q);
      if (sku) params.set("sku", sku);
      if (groupName) params.set("group_name", groupName);
      if (kategoriFilter) params.set("kategori_id", kategoriFilter);
      if (brandFilter) params.set("brand_id", brandFilter);
      params.set("page", String(page));
      params.set("limit", String(limit));
      const r = await fetch(`/api/admin/master/produk?${params.toString()}`, { headers: authHeader(), cache: "no-store" });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal memuat produk");
        setItems([]);
        setTotal(0);
        setTotalPages(1);
        return;
      }
      const nextTotal = Number(j.total || 0);
      const nextTotalPages = Math.max(1, Number(j.total_pages || 1));
      if (page > nextTotalPages) {
        setPage(nextTotalPages);
        return;
      }
      setItems(Array.isArray(j.items) ? j.items : []);
      setTotal(nextTotal);
      setTotalPages(nextTotalPages);
    } finally {
      setLoading(false);
    }
  }, [brandFilter, groupName, kategoriFilter, limit, page, q, sku]);

  async function refreshAll() {
    await Promise.all([loadMasters(), loadGroupOptions(), loadProduk()]);
  }

  useEffect(() => {
    loadMasters();
    loadGroupOptions();
  }, []);

  useEffect(() => {
    loadProduk();
  }, [loadProduk]);

  useEffect(() => {
    function handleOutsideClick(e: MouseEvent) {
      const target = e.target as HTMLElement | null;
      if (!target) return;
      if (target.closest("[data-action-dropdown]")) return;
      setOpenActionKey(null);
      if (target.closest("[data-group-dropdown]")) return;
      setGroupDropdownOpen(false);
    }

    function handleEsc(e: KeyboardEvent) {
      if (e.key === "Escape") setOpenActionKey(null);
    }

    document.addEventListener("mousedown", handleOutsideClick);
    document.addEventListener("keydown", handleEsc);
    return () => {
      document.removeEventListener("mousedown", handleOutsideClick);
      document.removeEventListener("keydown", handleEsc);
    };
  }, []);

  function openCreate() {
    setEditing(null);
    setForm({
      sku: "",
      nama: "",
      group_name: "",
      kategori_id: activeKategori[0]?.id ? String(activeKategori[0].id) : "",
      brand_id: activeBrand[0]?.id ? String(activeBrand[0].id) : "",
      tipe_harga: "FIXED",
      nominal: "",
      maksimal_nominal: "",
      jam_buka: "00:31",
      jam_tutup: "23:29",
      aktif: true,
    });
    setOpen(true);
  }

  function openEdit(x: ProdukRow) {
    setEditing(x);
    setForm({
      sku: x.sku || "",
      nama: x.nama || "",
      group_name: x.group_name || "",
      kategori_id: String(x.kategori_id || ""),
      brand_id: String(x.brand_id || ""),
      tipe_harga: x.tipe_harga,
      nominal: x.nominal != null ? String(x.nominal) : "",
      maksimal_nominal: x.maksimal_nominal != null ? String(x.maksimal_nominal) : "",
      jam_buka: (x.jam_buka || "00:31:00").slice(0, 5),
      jam_tutup: (x.jam_tutup || "23:29:00").slice(0, 5),
      aktif: Boolean(x.aktif),
    });
    setOpen(true);
  }

  async function save() {
    const sku = form.sku.trim().toUpperCase();
    const nama = form.nama.trim();
    const groupName = form.group_name.trim();
    const kategoriID = Number(form.kategori_id);
    const brandID = Number(form.brand_id);

    if (!sku || !nama) return alertWarning("SKU dan Nama wajib diisi.");
    if (!kategoriID || !brandID) return alertWarning("Kategori dan Brand wajib dipilih.");

    const nominal = toNumOrNull(form.nominal);
    const maksimal = toNumOrNull(form.maksimal_nominal);

    if (form.tipe_harga === "FIXED" && !nominal) {
      return alertWarning("Untuk FIXED, nominal wajib diisi.");
    }
    if (form.tipe_harga === "OPEN_AMOUNT" && !maksimal) {
      return alertWarning("Untuk OPEN_AMOUNT, maksimal nominal wajib diisi.");
    }

    setSaving(true);
    try {
      const payload = {
        sku,
        nama,
        group_name: groupName,
        kategori_id: kategoriID,
        brand_id: brandID,
        tipe_harga: form.tipe_harga,
        nominal: form.tipe_harga === "FIXED" ? nominal : null,
        maksimal_nominal: form.tipe_harga === "OPEN_AMOUNT" ? maksimal : null,
        jam_buka: form.jam_buka || "",
        jam_tutup: form.jam_tutup || "",
        aktif: form.aktif,
      };

      const isEdit = Boolean(editing?.id);
      const url = isEdit ? `/api/admin/master/produk/${editing?.id}` : "/api/admin/master/produk";
      const method = isEdit ? "PUT" : "POST";

      const r = await fetch(url, {
        method,
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify(payload),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) return alertError(j.error || "Gagal menyimpan produk");

      setOpen(false);
      await loadProduk();
      await alertSuccess(isEdit ? "Produk berhasil diperbarui." : "Produk berhasil ditambahkan.");
    } finally {
      setSaving(false);
    }
  }

  async function remove(x: ProdukRow) {
    const ok = await alertConfirm({
      title: "Hapus Produk",
      text: `Yakin hapus produk ${x.sku}?`,
      confirmButtonText: "Ya, hapus",
    });
    if (!ok) return;

    const r = await fetch(`/api/admin/master/produk/${x.id}`, {
      method: "DELETE",
      headers: authHeader(),
    });
    const j = await r.json().catch(() => ({}));
    if (!r.ok || !j.ok) return alertError(j.error || "Gagal hapus produk");

    await loadProduk();
    await alertSuccess("Produk berhasil dihapus.");
  }

  function applySearch() {
    setPage(1);
    setQ(qInput.trim());
    setSku(skuInput.trim().toUpperCase());
    setMobileFilterOpen(false);
  }

  function visiblePages(current: number, maxPage: number): number[] {
    const start = Math.max(1, current - 2);
    const end = Math.min(maxPage, current + 2);
    const out: number[] = [];
    for (let i = start; i <= end; i += 1) out.push(i);
    return out;
  }

  const pageButtons = visiblePages(page, totalPages);
  const offset = (page - 1) * limit;
  const filteredGroupOptions = groupOptions.filter((g) => g.toLowerCase().includes(groupSearch.trim().toLowerCase()));
  const activeFilterCount = [q, sku, groupName, kategoriFilter, brandFilter].filter((v) => v.trim()).length;

  const columns: DataTableColumn<ProdukRow>[] = [
    { id: "id", header: "ID", tdClassName: "whitespace-nowrap font-mono text-xs text-slate-400", render: (x) => String(x.id) },
    { id: "sku", header: "SKU", tdClassName: "whitespace-nowrap font-mono text-xs text-slate-200", render: (x) => x.sku },
    { id: "nama", header: "Nama", tdClassName: "whitespace-nowrap text-slate-100", render: (x) => x.nama },
    { id: "group_name", header: "Grup", tdClassName: "whitespace-nowrap text-slate-300", render: (x) => x.group_name || "-" },
    { id: "kategori", header: "Kategori", tdClassName: "whitespace-nowrap text-slate-300", render: (x) => x.kategori_nama || "-" },
    { id: "brand", header: "Brand", tdClassName: "whitespace-nowrap text-slate-300", render: (x) => x.brand_nama || "-" },
    { id: "tipe", header: "Tipe", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => x.tipe_harga },
    { id: "nominal", header: "Nominal", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => (x.nominal != null ? x.nominal.toLocaleString("id-ID") : "-") },
    { id: "maks", header: "Maksimal", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => (x.maksimal_nominal != null ? x.maksimal_nominal.toLocaleString("id-ID") : "-") },
    {
      id: "jam_online",
      header: "Jam Online",
      tdClassName: "whitespace-nowrap text-xs text-slate-300",
      render: (x) => {
        const buka = (x.jam_buka || "00:31").slice(0, 5);
        const tutup = (x.jam_tutup || "23:29").slice(0, 5);
        const is24h = buka === "00:00" && tutup === "23:59";
        return is24h
          ? <span className="rounded-full border border-emerald-400 bg-emerald-100 px-2 py-0.5 text-xs font-bold text-emerald-900">24 Jam</span>
          : <span>{buka} - {tutup}</span>;
      },
    },
    {
      id: "aktif",
      header: "Status",
      tdClassName: "whitespace-nowrap",
      render: (x) =>
        x.aktif ? (
          <span className="rounded-full border border-emerald-400 bg-emerald-100 px-2 py-0.5 text-xs font-bold text-emerald-900">Aktif</span>
        ) : (
          <span className="rounded-full border border-rose-400 bg-rose-100 px-2 py-0.5 text-xs font-bold text-rose-900">Nonaktif</span>
        ),
    },
  ];

  const actions: DataTableActions<ProdukRow> = {
    header: "Aksi",
    align: "right",
    tdClassName: "whitespace-nowrap",
    render: (x, index) => (
      <div className="relative inline-flex" data-action-dropdown>
        <Button
          size="sm"
          className="h-8 w-8 bg-emerald-700 px-0 text-white hover:bg-emerald-600"
          onClick={() => setOpenActionKey((prev) => (prev === String(x.id) ? null : String(x.id)))}
          aria-label="Aksi"
        >
          <Eye className="h-4 w-4" />
        </Button>

        {openActionKey === String(x.id) ? (
          <div
            className={`absolute right-0 z-20 w-36 rounded-xl border border-white/12 bg-slate-900/95 p-1.5 shadow-2xl backdrop-blur ${
              index >= items.length - 2 ? "bottom-10" : "top-10"
            }`}
          >
            <button
              type="button"
              className="flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm text-slate-100 transition hover:bg-white/10"
              onClick={() => {
                setOpenActionKey(null);
                openEdit(x);
              }}
            >
              <Pencil className="h-3.5 w-3.5" />
              Edit
            </button>
            <button
              type="button"
              className="mt-1 flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm font-semibold text-rose-800 transition hover:bg-rose-50"
              onClick={() => {
                setOpenActionKey(null);
                void remove(x);
              }}
            >
              <Trash2 className="h-3.5 w-3.5" />
              Hapus
            </button>
          </div>
        ) : null}
      </div>
    ),
  };

  return (
    <div className="space-y-4 p-2">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <div className="text-lg font-semibold tracking-tight">Master Produk</div>
          <div className="text-sm text-muted-foreground">Kelola produk internal, brand produk, dan status aktif produk. Harga retail mengikuti harga Yuscom ditambah fee sesuai tier.</div>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" className="h-10" onClick={refreshAll} disabled={loading}>
            <RefreshCw className={`mr-2 h-4 w-4 ${loading ? "animate-spin" : ""}`} />
            Refresh
          </Button>
          <Button variant="primary" className="h-10" onClick={openCreate}>
            <Plus className="mr-2 h-4 w-4" />
            Tambah Produk
          </Button>
        </div>
      </div>

      <div className="md:hidden">
        <Button
          type="button"
          variant="outline"
          className="h-10 w-full justify-between border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50"
          onClick={() => setMobileFilterOpen((v) => !v)}
        >
          <span className="inline-flex items-center gap-2">
            <SlidersHorizontal className="h-4 w-4" />
            {mobileFilterOpen ? "Tutup Filter" : "Buka Filter"}
          </span>
          {activeFilterCount > 0 ? (
            <span className="rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-semibold text-emerald-700">{activeFilterCount}</span>
          ) : null}
        </Button>
      </div>

      <div
        className={`${mobileFilterOpen ? "grid" : "hidden"} grid-cols-2 gap-2 rounded-md border border-white/15 bg-slate-950/50 p-3 sm:grid-cols-2 md:grid lg:grid-cols-12`}
      >
        <div className="col-span-2 sm:col-span-2 lg:col-span-4">
          <div className="mb-1 text-xs text-muted-foreground">Cari Produk</div>
          <Input
            value={qInput}
            onChange={(e) => setQInput(e.target.value)}
            placeholder="Cari SKU / nama produk"
            onKeyDown={(e) => {
              if (e.key === "Enter") applySearch();
            }}
          />
        </div>
        <div className="col-span-2 sm:col-span-1 lg:col-span-2">
          <div className="mb-1 text-xs text-muted-foreground">SKU / Kode Produk</div>
          <Input
            value={skuInput}
            onChange={(e) => setSkuInput(e.target.value.toUpperCase())}
            placeholder="Contoh: AXM10"
            onKeyDown={(e) => {
              if (e.key === "Enter") applySearch();
            }}
          />
        </div>
        <div className="col-span-2 sm:col-span-2 lg:col-span-4">
          <div className="mb-1 text-xs text-muted-foreground">Filter Grup</div>
          <div className="relative" data-group-dropdown>
            <button
              type="button"
              className="flex h-10 w-full items-center justify-between rounded-md border border-input bg-transparent px-3 text-left text-sm"
              onClick={() => {
                setGroupDropdownOpen((v) => !v);
                setGroupSearch("");
              }}
            >
              <span className={groupName ? "text-slate-100" : "text-slate-400"}>{groupName || "Semua grup"}</span>
              <ChevronRight className={`h-4 w-4 text-slate-400 transition ${groupDropdownOpen ? "rotate-90" : ""}`} />
            </button>

            {groupDropdownOpen ? (
              <div className="absolute z-30 mt-1 w-full rounded-md border border-white/15 bg-slate-900 p-2 shadow-2xl">
                <Input value={groupSearch} onChange={(e) => setGroupSearch(e.target.value)} placeholder="Cari grup..." />
                <div className="mt-2 max-h-52 overflow-auto rounded-md border border-white/10">
                  <button
                    type="button"
                    className={`block w-full px-3 py-2 text-left text-sm transition hover:bg-emerald-50 ${groupName === "" ? "bg-emerald-100 text-emerald-800" : "text-slate-200"}`}
                    onClick={() => {
                      setGroupName("");
                      setPage(1);
                      setGroupDropdownOpen(false);
                    }}
                  >
                    Semua grup
                  </button>
                  {filteredGroupOptions.map((g) => (
                    <button
                      key={g}
                      type="button"
                      className={`block w-full px-3 py-2 text-left text-sm transition hover:bg-emerald-50 ${groupName === g ? "bg-emerald-100 text-emerald-800" : "text-slate-200"}`}
                      onClick={() => {
                        setGroupName(g);
                        setPage(1);
                        setGroupDropdownOpen(false);
                      }}
                    >
                      {g}
                    </button>
                  ))}
                  {!filteredGroupOptions.length ? <div className="px-3 py-2 text-sm text-slate-400">Grup tidak ditemukan.</div> : null}
                </div>
              </div>
            ) : null}
          </div>
        </div>
        <div className="col-span-2 sm:col-span-1 lg:col-span-2">
          <div className="mb-1 text-xs text-muted-foreground">Filter Kategori</div>
          <select
            className="h-10 w-full rounded-md border border-input bg-transparent px-3 text-sm"
            value={kategoriFilter}
            onChange={(e) => {
              setKategoriFilter(e.target.value);
              setPage(1);
            }}
          >
            <option value="">Semua kategori</option>
            {activeKategori.map((k) => (
              <option key={k.id} value={String(k.id)}>
                {k.nama}
              </option>
            ))}
          </select>
        </div>
        <div className="col-span-2 sm:col-span-1 lg:col-span-2">
          <div className="mb-1 text-xs text-muted-foreground">Filter Brand</div>
          <select
            className="h-10 w-full rounded-md border border-input bg-transparent px-3 text-sm"
            value={brandFilter}
            onChange={(e) => {
              setBrandFilter(e.target.value);
              setPage(1);
            }}
          >
            <option value="">Semua brand</option>
            {activeBrand.map((b) => (
              <option key={b.id} value={String(b.id)}>
                {b.nama}
              </option>
            ))}
          </select>
        </div>
        <div className="col-span-1 sm:col-span-1 lg:col-span-2">
          <div className="mb-1 hidden text-xs text-muted-foreground sm:block whitespace-nowrap">Limit</div>
          <select
            className="h-10 w-full rounded-md border border-input bg-transparent px-3 text-sm"
            value={String(limit)}
            onChange={(e) => {
              setLimit(Number(e.target.value));
              setPage(1);
            }}
          >
            <option value="10">10</option>
            <option value="20">20</option>
            <option value="50">50</option>
            <option value="100">100</option>
          </select>
        </div>
        <div className="col-span-1 sm:col-span-1 lg:col-span-2">
          <div className="mb-1 hidden text-xs text-muted-foreground sm:block whitespace-nowrap">Aksi</div>
          <div className="flex gap-2 lg:justify-start">
            <Button variant="outline" className="h-10" onClick={applySearch}>
              Cari
            </Button>
            <Button
              variant="outline"
              className="h-10"
              onClick={() => {
                setQInput("");
                setQ("");
                setSkuInput("");
                setSku("");
                setGroupName("");
                setKategoriFilter("");
                setBrandFilter("");
                setPage(1);
                setMobileFilterOpen(false);
              }}
            >
              Reset
            </Button>
          </div>
        </div>
      </div>

      <DataTable
        columns={columns}
        rows={items}
        rowKey={(x) => x.id}
        emptyText={loading ? "Loading..." : "Belum ada produk."}
        rowNumberStart={offset + 1}
        actions={actions}
      />

      <div className="flex flex-wrap items-center justify-between gap-3 rounded-md border border-white/15 bg-linear-to-r from-slate-900/80 via-slate-900/65 to-cyan-950/25 p-3 text-sm">
        <div className="text-slate-300">
          Total {total.toLocaleString("id-ID")} data, halaman {page} / {totalPages}
        </div>
        <div className="flex flex-wrap items-center gap-1.5">
          <Button variant="outline" size="sm" onClick={() => setPage(1)} disabled={page <= 1} aria-label="Halaman awal">
            <ChevronsLeft className="h-4 w-4" />
          </Button>
          <Button variant="outline" size="sm" onClick={() => setPage((p) => Math.max(1, p - 1))} disabled={page <= 1} aria-label="Halaman sebelumnya">
            <ChevronLeft className="h-4 w-4" />
          </Button>

          {pageButtons[0] > 1 ? <span className="px-1 text-slate-400">...</span> : null}
          {pageButtons.map((p) => (
            <Button
              key={p}
              variant={p === page ? "outline" : "ghost"}
              size="sm"
              className={p === page ? "border-emerald-300 bg-emerald-100 text-emerald-800" : "border-transparent"}
              onClick={() => setPage(p)}
            >
              {p}
            </Button>
          ))}
          {pageButtons[pageButtons.length - 1] < totalPages ? <span className="px-1 text-slate-400">...</span> : null}

          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            disabled={page >= totalPages}
            aria-label="Halaman berikutnya"
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
          <Button variant="outline" size="sm" onClick={() => setPage(totalPages)} disabled={page >= totalPages} aria-label="Halaman akhir">
            <ChevronsRight className="h-4 w-4" />
          </Button>
        </div>
      </div>

      <AppModal
        open={open}
        onClose={() => setOpen(false)}
        title={editing ? "Edit Produk" : "Tambah Produk"}
        subtitle="Pastikan kategori, brand, dan tipe harga sudah sesuai."
        maxWidthClassName="max-w-2xl"
        footer={
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={() => setOpen(false)} disabled={saving}>
              Batal
            </Button>
            <Button variant="success" onClick={save} disabled={saving}>
              {saving ? "Menyimpan..." : "Simpan"}
            </Button>
          </div>
        }
      >
        <div className="grid gap-3 sm:grid-cols-2">
              <div>
                <div className="mb-1 text-xs text-muted-foreground">SKU</div>
                <Input value={form.sku} onChange={(e) => setForm((p) => ({ ...p, sku: e.target.value }))} placeholder="DNID" />
              </div>
              <div>
                <div className="mb-1 text-xs text-muted-foreground">Nama</div>
                <Input value={form.nama} onChange={(e) => setForm((p) => ({ ...p, nama: e.target.value }))} placeholder="DANA 10K" />
              </div>
              <div className="sm:col-span-2">
                <div className="mb-1 text-xs text-muted-foreground">Group Name</div>
                <Input
                  value={form.group_name}
                  onChange={(e) => setForm((p) => ({ ...p, group_name: e.target.value }))}
                  placeholder="AXIS XL REGULER PROMO"
                />
              </div>

              <div>
                <div className="mb-1 text-xs text-muted-foreground">Kategori</div>
                <select
                  className="h-10 w-full rounded-md border border-input bg-transparent px-3 text-sm"
                  value={form.kategori_id}
                  onChange={(e) => setForm((p) => ({ ...p, kategori_id: e.target.value }))}
                >
                  <option value="">Pilih kategori</option>
                  {activeKategori.map((x) => (
                    <option key={x.id} value={x.id}>
                      {x.nama}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <div className="mb-1 text-xs text-muted-foreground">Brand</div>
                <select
                  className="h-10 w-full rounded-md border border-input bg-transparent px-3 text-sm"
                  value={form.brand_id}
                  onChange={(e) => setForm((p) => ({ ...p, brand_id: e.target.value }))}
                >
                  <option value="">Pilih brand</option>
                  {activeBrand.map((x) => (
                    <option key={x.id} value={x.id}>
                      {x.nama}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <div className="mb-1 text-xs text-muted-foreground">Tipe Harga</div>
                <select
                  className="h-10 w-full rounded-md border border-input bg-transparent px-3 text-sm"
                  value={form.tipe_harga}
                  onChange={(e) => setForm((p) => ({ ...p, tipe_harga: e.target.value as "FIXED" | "OPEN_AMOUNT" }))}
                >
                  <option value="FIXED">FIXED</option>
                  <option value="OPEN_AMOUNT">OPEN_AMOUNT</option>
                </select>
              </div>

              <div>
                <div className="mb-1 text-xs text-muted-foreground">Status</div>
                <select
                  className="h-10 w-full rounded-md border border-input bg-transparent px-3 text-sm"
                  value={form.aktif ? "1" : "0"}
                  onChange={(e) => setForm((p) => ({ ...p, aktif: e.target.value === "1" }))}
                >
                  <option value="1">Aktif</option>
                  <option value="0">Nonaktif</option>
                </select>
              </div>

              <div>
                <div className="mb-1 text-xs text-muted-foreground">Jam Buka</div>
                <Input type="time" value={form.jam_buka} onChange={(e) => setForm((p) => ({ ...p, jam_buka: e.target.value }))} />
              </div>
              <div>
                <div className="mb-1 text-xs text-muted-foreground">Jam Tutup</div>
                <Input type="time" value={form.jam_tutup} onChange={(e) => setForm((p) => ({ ...p, jam_tutup: e.target.value }))} />
              </div>

              {form.tipe_harga === "FIXED" ? (
                <div>
                  <div className="mb-1 text-xs text-muted-foreground">Nominal</div>
                  <Input
                    value={form.nominal}
                    onChange={(e) => setForm((p) => ({ ...p, nominal: e.target.value }))}
                    placeholder="10000"
                    inputMode="numeric"
                  />
                </div>
              ) : (
                <div>
                  <div className="mb-1 text-xs text-muted-foreground">Maksimal Nominal</div>
                  <Input
                    value={form.maksimal_nominal}
                    onChange={(e) => setForm((p) => ({ ...p, maksimal_nominal: e.target.value }))}
                    placeholder="2000000"
                    inputMode="numeric"
                  />
                </div>
              )}
        </div>
      </AppModal>
    </div>
  );
}
