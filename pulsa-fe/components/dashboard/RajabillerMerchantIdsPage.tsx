"use client";

import { useEffect, useMemo, useState } from "react";
import { Copy, Pencil, Plus, Power, RefreshCw, RotateCcw, Search } from "lucide-react";
import { AppModal } from "@/components/ui/app-modal";
import { alertError, alertSuccess, alertWarning } from "@/components/ui/alerts";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableActions, type DataTableColumn } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";

type MerchantIDRow = {
  id: number;
  provider: string;
  merchant_id: string;
  label: string;
  catatan: string;
  aktif: boolean;
  dibuat_pada: string;
  diubah_pada: string;
};

const PAGE_SIZE = 25;
const PROVIDER = "rajabiller";

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

function statusClass(active: boolean) {
  return active
    ? "rounded-full bg-emerald-500/15 px-2 py-1 text-xs font-medium text-emerald-300"
    : "rounded-full bg-slate-500/15 px-2 py-1 text-xs font-medium text-slate-300";
}

export default function RajabillerMerchantIdsPage() {
  const [items, setItems] = useState<MerchantIDRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);

  const [q, setQ] = useState("");
  const [aktifOnly, setAktifOnly] = useState(false);
  const [appliedQ, setAppliedQ] = useState("");
  const [appliedAktifOnly, setAppliedAktifOnly] = useState(false);

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<MerchantIDRow | null>(null);
  const [formMerchantID, setFormMerchantID] = useState("");
  const [formLabel, setFormLabel] = useState("");
  const [formCatatan, setFormCatatan] = useState("");
  const [formAktif, setFormAktif] = useState(true);
  const [saving, setSaving] = useState(false);

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const currentPage = Math.floor(offset / PAGE_SIZE) + 1;
  const activeCount = useMemo(() => items.filter((item) => item.aktif).length, [items]);

  async function load() {
    setLoading(true);
    try {
      const qs = new URLSearchParams();
      qs.set("provider", PROVIDER);
      qs.set("limit", String(PAGE_SIZE));
      qs.set("offset", String(offset));
      if (appliedQ.trim()) qs.set("q", appliedQ.trim());
      if (appliedAktifOnly) qs.set("aktif", "1");

      const r = await fetch(`/api/admin/provider-merchant-ids?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal memuat merchant ID Rajabiller");
        setItems([]);
        setTotal(0);
        return;
      }
      setItems(Array.isArray(j.items) ? j.items : []);
      setTotal(Number(j.total || 0));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, [offset, appliedQ, appliedAktifOnly]);

  function openCreate() {
    setEditing(null);
    setFormMerchantID("");
    setFormLabel("");
    setFormCatatan("");
    setFormAktif(true);
    setFormOpen(true);
  }

  function openEdit(item: MerchantIDRow) {
    setEditing(item);
    setFormMerchantID(item.merchant_id);
    setFormLabel(item.label || "");
    setFormCatatan(item.catatan || "");
    setFormAktif(item.aktif);
    setFormOpen(true);
  }

  async function save() {
    if (saving) return;
    if (!formMerchantID.trim()) {
      await alertWarning("Merchant ID wajib diisi.");
      return;
    }

    setSaving(true);
    try {
      const payload = {
        provider: PROVIDER,
        merchant_id: formMerchantID.trim(),
        label: formLabel.trim(),
        catatan: formCatatan.trim(),
        aktif: formAktif,
      };
      const url = editing ? `/api/admin/provider-merchant-ids/${editing.id}` : "/api/admin/provider-merchant-ids";
      const r = await fetch(url, {
        method: editing ? "PUT" : "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify(payload),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal menyimpan merchant ID");
        return;
      }
      await alertSuccess(editing ? "Merchant ID diperbarui." : "Merchant ID ditambahkan.");
      setFormOpen(false);
      await load();
    } finally {
      setSaving(false);
    }
  }

  async function toggleActive(item: MerchantIDRow) {
    const nextActive = !item.aktif;
    const r = await fetch(`/api/admin/provider-merchant-ids/${item.id}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json", ...authHeader() },
      body: JSON.stringify({
        provider: PROVIDER,
        merchant_id: item.merchant_id,
        label: item.label || "",
        catatan: item.catatan || "",
        aktif: nextActive,
      }),
      cache: "no-store",
    });
    const j = await r.json().catch(() => ({}));
    if (!r.ok || !j.ok) {
      await alertError(j.error || "Gagal mengubah status merchant ID");
      return;
    }
    await alertSuccess(nextActive ? "Merchant ID diaktifkan." : "Merchant ID dinonaktifkan.");
    await load();
  }

  async function copy(value: string) {
    if (!value) return;
    try {
      await navigator.clipboard.writeText(value);
      await alertSuccess("Merchant ID disalin.");
    } catch {
      await alertError("Gagal menyalin merchant ID.");
    }
  }

  const columns: DataTableColumn<MerchantIDRow>[] = [
    {
      id: "merchant",
      header: "Merchant ID",
      thClassName: "whitespace-nowrap",
      tdClassName: "whitespace-nowrap font-mono text-cyan-200",
      render: (item) => item.merchant_id,
    },
    { id: "label", header: "Label", thClassName: "whitespace-nowrap", tdClassName: "min-w-52 text-slate-100", render: (item) => item.label || "-" },
    { id: "catatan", header: "Catatan", thClassName: "whitespace-nowrap", tdClassName: "min-w-64 text-slate-300", render: (item) => item.catatan || "-" },
    {
      id: "aktif",
      header: "Status",
      thClassName: "whitespace-nowrap",
      tdClassName: "whitespace-nowrap",
      render: (item) => <span className={statusClass(item.aktif)}>{item.aktif ? "Aktif" : "Nonaktif"}</span>,
    },
    { id: "updated", header: "Update", thClassName: "whitespace-nowrap", tdClassName: "whitespace-nowrap text-slate-400", render: (item) => new Date(item.diubah_pada).toLocaleString("id-ID") },
  ];

  const actions: DataTableActions<MerchantIDRow> = {
    header: "Aksi",
    align: "right",
    tdClassName: "whitespace-nowrap",
    render: (item) => (
      <div className="flex flex-nowrap items-center justify-end gap-2">
        <Button type="button" variant="outline" className="h-9 w-9 rounded-xl border-white/10 bg-slate-900/70 p-0 text-slate-200" onClick={() => copy(item.merchant_id)} title="Salin merchant ID">
          <Copy className="h-4 w-4" />
        </Button>
        <Button type="button" variant="warning" className="h-9 w-9 rounded-xl p-0" onClick={() => openEdit(item)} title="Edit merchant ID">
          <Pencil className="h-4 w-4" />
        </Button>
        <Button type="button" variant={item.aktif ? "danger" : "primary"} className="h-9 w-9 rounded-xl p-0" onClick={() => toggleActive(item)} title={item.aktif ? "Nonaktifkan" : "Aktifkan"}>
          <Power className="h-4 w-4" />
        </Button>
      </div>
    ),
  };

  return (
    <div className="space-y-4 p-2">
      <div className="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
        <div>
          <div className="text-lg font-semibold tracking-tight text-white">Merchant ID Rajabiller</div>
          <div className="text-sm text-muted-foreground">Bank transfer Rajabiller memilih salah satu merchant ID aktif secara random.</div>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button type="button" variant="outline" className="h-10 border-white/10 bg-slate-900/70 text-slate-200" onClick={() => load()} disabled={loading}>
            <RefreshCw className="mr-2 h-4 w-4" />
            Refresh
          </Button>
          <Button type="button" variant="primary" className="h-10" onClick={openCreate}>
            <Plus className="mr-2 h-4 w-4" />
            Tambah Merchant ID
          </Button>
        </div>
      </div>

      <div className="grid gap-3 rounded-md border border-white/10 bg-slate-950/70 p-4 shadow-xl md:grid-cols-12">
        <div className="space-y-2 md:col-span-6">
          <label className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">Pencarian</label>
          <Input value={q} onChange={(e) => setQ(e.target.value)} placeholder="Cari merchant ID, label, atau catatan" className="h-11 rounded-xl border-white/10 bg-slate-900/80 text-slate-100" />
        </div>
        <div className="flex items-end gap-2 md:col-span-6">
          <label className="mb-3 flex items-center gap-2 text-sm text-slate-200">
            <input type="checkbox" checked={aktifOnly} onChange={(e) => setAktifOnly(e.target.checked)} className="h-4 w-4 rounded border-white/20 bg-slate-900" />
            Aktif saja
          </label>
          <div className="ml-auto flex gap-2">
            <Button type="button" variant="primary" className="h-11" onClick={() => { setOffset(0); setAppliedQ(q); setAppliedAktifOnly(aktifOnly); }} disabled={loading}>
              <Search className="mr-2 h-4 w-4" />
              Cari
            </Button>
            <Button type="button" variant="outline" className="h-11 border-white/10 bg-slate-900/70 text-slate-200" onClick={() => { setQ(""); setAktifOnly(false); setOffset(0); setAppliedQ(""); setAppliedAktifOnly(false); }} disabled={loading}>
              <RotateCcw className="mr-2 h-4 w-4" />
              Reset
            </Button>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3">
        <div className="rounded-md border border-cyan-400/20 bg-cyan-400/10 px-4 py-3">
          <div className="text-xs uppercase tracking-[0.2em] text-cyan-200">Total Filter</div>
          <div className="mt-1 text-xl font-semibold text-white">{total}</div>
        </div>
        <div className="rounded-md border border-emerald-400/20 bg-emerald-400/10 px-4 py-3">
          <div className="text-xs uppercase tracking-[0.2em] text-emerald-200">Aktif di Halaman</div>
          <div className="mt-1 text-xl font-semibold text-white">{activeCount}</div>
        </div>
      </div>

      <DataTable
        columns={columns}
        rows={items}
        rowKey={(item) => item.id}
        rowNumberStart={offset + 1}
        emptyText={loading ? "Memuat merchant ID..." : "Belum ada merchant ID Rajabiller."}
        loading={loading}
        minWidthClassName="min-w-180"
        actions={actions}
      />

      <div className="flex flex-wrap items-center justify-end gap-2 rounded-md border border-white/10 bg-slate-950/70 p-2">
        <div className="mr-auto text-sm text-slate-400">Halaman {currentPage} / {totalPages}</div>
        <Button type="button" variant="outline" className="h-9 border-white/10 bg-slate-900/70 text-slate-200" onClick={() => setOffset(0)} disabled={loading || currentPage <= 1}>Awal</Button>
        <Button type="button" variant="outline" className="h-9 border-white/10 bg-slate-900/70 text-slate-200" onClick={() => setOffset((v) => Math.max(0, v - PAGE_SIZE))} disabled={loading || currentPage <= 1}>Prev</Button>
        <Button type="button" variant="outline" className="h-9 border-white/10 bg-slate-900/70 text-slate-200" onClick={() => setOffset((v) => v + PAGE_SIZE)} disabled={loading || currentPage >= totalPages}>Next</Button>
        <Button type="button" variant="outline" className="h-9 border-white/10 bg-slate-900/70 text-slate-200" onClick={() => setOffset((totalPages - 1) * PAGE_SIZE)} disabled={loading || currentPage >= totalPages}>Akhir</Button>
      </div>

      <AppModal
        open={formOpen}
        onClose={() => {
          if (!saving) setFormOpen(false);
        }}
        title={editing ? "Edit Merchant ID" : "Tambah Merchant ID"}
        subtitle="Merchant ID aktif dipilih random untuk transaksi bank Rajabiller."
        footer={
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" className="border-white/10 bg-slate-900/70 text-slate-200" onClick={() => setFormOpen(false)} disabled={saving}>Batal</Button>
            <Button type="button" variant="primary" onClick={save} disabled={saving}>{saving ? "Menyimpan..." : "Simpan"}</Button>
          </div>
        }
      >
        <div className="grid gap-3 md:grid-cols-2">
          <div className="space-y-2 md:col-span-2">
            <label className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">Merchant ID</label>
            <Input value={formMerchantID} onChange={(e) => setFormMerchantID(e.target.value)} placeholder="Contoh: SentosaLintasLink" className="h-11 border-white/10 bg-slate-950 font-mono text-slate-100" />
          </div>
          <div className="space-y-2">
            <label className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">Label</label>
            <Input value={formLabel} onChange={(e) => setFormLabel(e.target.value)} placeholder="Nama outlet / akun" className="h-11 border-white/10 bg-slate-950 text-slate-100" />
          </div>
          <div className="space-y-2">
            <label className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">Status</label>
            <select value={formAktif ? "1" : "0"} onChange={(e) => setFormAktif(e.target.value === "1")} className="h-11 w-full rounded-md border border-white/10 bg-slate-950 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50">
              <option value="1">Aktif</option>
              <option value="0">Nonaktif</option>
            </select>
          </div>
          <div className="space-y-2 md:col-span-2">
            <label className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">Catatan</label>
            <Input value={formCatatan} onChange={(e) => setFormCatatan(e.target.value)} placeholder="Catatan internal" className="h-11 border-white/10 bg-slate-950 text-slate-100" />
          </div>
        </div>
      </AppModal>
    </div>
  );
}
