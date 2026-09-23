"use client";

import { useEffect, useMemo, useState } from "react";
import { Copy, Pencil, Plus, RefreshCw, RotateCcw, Search, Trash2 } from "lucide-react";
import { AppModal } from "@/components/ui/app-modal";
import { alertConfirm, alertError, alertSuccess, alertWarning } from "@/components/ui/alerts";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableActions, type DataTableColumn } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";

type ProviderAccountRow = {
  id: number;
  provider: string;
  nama: string;
  bank: string;
  nomor_rekening: string;
  nomor_rekening_digits: string;
  catatan: string;
  aktif: boolean;
  dibuat_pada: string;
  diubah_pada: string;
};

type ProviderOption = {
  provider?: string;
  nama?: string;
};

const PAGE_SIZE = 25;

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

function onlyDigits(value: string): string {
  return (value || "").replace(/[^\d]/g, "");
}

function providerLabel(value: string): string {
  return (value || "-").toUpperCase();
}

export default function ProviderAccountsPage() {
  const [items, setItems] = useState<ProviderAccountRow[]>([]);
  const [providers, setProviders] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);

  const [q, setQ] = useState("");
  const [provider, setProvider] = useState("");
  const [aktifOnly, setAktifOnly] = useState(false);
  const [appliedQ, setAppliedQ] = useState("");
  const [appliedProvider, setAppliedProvider] = useState("");
  const [appliedAktifOnly, setAppliedAktifOnly] = useState(false);

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<ProviderAccountRow | null>(null);
  const [formProvider, setFormProvider] = useState("");
  const [formNama, setFormNama] = useState("");
  const [formBank, setFormBank] = useState("");
  const [formNomor, setFormNomor] = useState("");
  const [formCatatan, setFormCatatan] = useState("");
  const [formAktif, setFormAktif] = useState(true);
  const [saving, setSaving] = useState(false);

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const currentPage = Math.floor(offset / PAGE_SIZE) + 1;
  const activeCount = useMemo(() => items.filter((item) => item.aktif).length, [items]);

  async function loadProviders() {
    try {
      const r = await fetch("/api/admin/provider/wallets", { headers: authHeader(), cache: "no-store" });
      const j = await r.json().catch(() => ({}));
      const rows: ProviderOption[] = Array.isArray(j?.data) ? j.data : [];
      const next = rows
        .map((row) => String(row.provider || row.nama || "").trim().toLowerCase())
        .filter(Boolean)
        .sort((a, b) => a.localeCompare(b, "id"));
      setProviders(Array.from(new Set(next)));
    } catch {
      setProviders([]);
    }
  }

  async function load() {
    setLoading(true);
    try {
      const qs = new URLSearchParams();
      qs.set("limit", String(PAGE_SIZE));
      qs.set("offset", String(offset));
      if (appliedQ.trim()) qs.set("q", appliedQ.trim());
      if (appliedProvider) qs.set("provider", appliedProvider);
      if (appliedAktifOnly) qs.set("aktif", "1");

      const r = await fetch(`/api/admin/provider-accounts?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal memuat rekening provider");
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
    void loadProviders();
  }, []);

  useEffect(() => {
    void load();
  }, [offset, appliedQ, appliedProvider, appliedAktifOnly]);

  function openCreate() {
    setEditing(null);
    setFormProvider(appliedProvider || providers[0] || "");
    setFormNama("");
    setFormBank("");
    setFormNomor("");
    setFormCatatan("");
    setFormAktif(true);
    setFormOpen(true);
  }

  function openEdit(item: ProviderAccountRow) {
    setEditing(item);
    setFormProvider(item.provider);
    setFormNama(item.nama);
    setFormBank(item.bank || "");
    setFormNomor(item.nomor_rekening);
    setFormCatatan(item.catatan || "");
    setFormAktif(item.aktif);
    setFormOpen(true);
  }

  async function save() {
    if (saving) return;
    if (!formProvider.trim()) {
      await alertWarning("Provider wajib dipilih.");
      return;
    }
    if (!formNama.trim()) {
      await alertWarning("Nama rekening wajib diisi.");
      return;
    }
    if (!onlyDigits(formNomor)) {
      await alertWarning("Nomor rekening wajib diisi.");
      return;
    }

    setSaving(true);
    try {
      const payload = {
        provider: formProvider.trim(),
        nama: formNama.trim(),
        bank: formBank.trim(),
        nomor_rekening: formNomor.trim(),
        catatan: formCatatan.trim(),
        aktif: formAktif,
      };
      const url = editing ? `/api/admin/provider-accounts/${editing.id}` : "/api/admin/provider-accounts";
      const r = await fetch(url, {
        method: editing ? "PUT" : "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify(payload),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal menyimpan rekening provider");
        return;
      }
      await alertSuccess(editing ? "Rekening provider diperbarui." : "Rekening provider ditambahkan.");
      setFormOpen(false);
      await load();
    } finally {
      setSaving(false);
    }
  }

  async function remove(item: ProviderAccountRow) {
    const ok = await alertConfirm({
      title: "Hapus rekening provider?",
      text: `${providerLabel(item.provider)} ${item.bank || ""} ${item.nomor_rekening}`,
      confirmButtonText: "Hapus",
    });
    if (!ok) return;

    const r = await fetch(`/api/admin/provider-accounts/${item.id}`, {
      method: "DELETE",
      headers: authHeader(),
      cache: "no-store",
    });
    const j = await r.json().catch(() => ({}));
    if (!r.ok || !j.ok) {
      await alertError(j.error || "Gagal menghapus rekening provider");
      return;
    }
    await alertSuccess("Rekening provider dihapus.");
    await load();
  }

  async function copy(value: string) {
    if (!value) return;
    try {
      await navigator.clipboard.writeText(value);
      await alertSuccess("Nomor rekening disalin.");
    } catch {
      await alertError("Gagal menyalin nomor rekening.");
    }
  }

  const columns: DataTableColumn<ProviderAccountRow>[] = [
    {
      id: "provider",
      header: "Provider",
      thClassName: "whitespace-nowrap",
      tdClassName: "whitespace-nowrap",
      render: (item) => (
        <span className="rounded-md border border-white/10 bg-white/5 px-2 py-1 text-xs font-semibold uppercase tracking-wide text-slate-100">
          {item.provider}
        </span>
      ),
    },
    { id: "nama", header: "Nama", thClassName: "whitespace-nowrap", tdClassName: "min-w-60 text-slate-100", render: (item) => item.nama },
    { id: "bank", header: "Bank", thClassName: "whitespace-nowrap", tdClassName: "whitespace-nowrap text-slate-200", render: (item) => item.bank || "-" },
    {
      id: "nomor",
      header: "Nomor Rekening",
      thClassName: "whitespace-nowrap",
      tdClassName: "whitespace-nowrap font-mono text-cyan-200",
      render: (item) => item.nomor_rekening,
    },
    { id: "catatan", header: "Catatan", thClassName: "whitespace-nowrap", tdClassName: "min-w-56 text-slate-300", render: (item) => item.catatan || "-" },
    {
      id: "aktif",
      header: "Status",
      thClassName: "whitespace-nowrap",
      tdClassName: "whitespace-nowrap",
      render: (item) => (
        <span className={item.aktif ? "rounded-full bg-emerald-500/15 px-2 py-1 text-xs font-medium text-emerald-300" : "rounded-full bg-slate-500/15 px-2 py-1 text-xs font-medium text-slate-300"}>
          {item.aktif ? "Aktif" : "Nonaktif"}
        </span>
      ),
    },
    { id: "updated", header: "Update", thClassName: "whitespace-nowrap", tdClassName: "whitespace-nowrap text-slate-400", render: (item) => new Date(item.diubah_pada).toLocaleString("id-ID") },
  ];

  const actions: DataTableActions<ProviderAccountRow> = {
    header: "Aksi",
    align: "right",
    tdClassName: "whitespace-nowrap",
    render: (item) => (
      <div className="flex flex-nowrap items-center justify-end gap-2">
        <Button type="button" variant="outline" className="h-9 w-9 rounded-xl border-white/10 bg-slate-900/70 p-0 text-slate-200" onClick={() => copy(item.nomor_rekening_digits || item.nomor_rekening)} title="Salin nomor rekening">
          <Copy className="h-4 w-4" />
        </Button>
        <Button type="button" variant="warning" className="h-9 w-9 rounded-xl p-0" onClick={() => openEdit(item)} title="Edit rekening provider">
          <Pencil className="h-4 w-4" />
        </Button>
        <Button type="button" variant="danger" className="h-9 w-9 rounded-xl p-0" onClick={() => remove(item)} title="Hapus rekening provider">
          <Trash2 className="h-4 w-4" />
        </Button>
      </div>
    ),
  };

  return (
    <div className="space-y-4 p-2">
      <div className="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
        <div>
          <div className="text-lg font-semibold tracking-tight text-white">Rekening Provider</div>
          <div className="text-sm text-muted-foreground">Master rekening tujuan provider untuk operasional wallet dan proyek kantor24.</div>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button type="button" variant="outline" className="h-10 border-white/10 bg-slate-900/70 text-slate-200" onClick={() => load()} disabled={loading}>
            <RefreshCw className="mr-2 h-4 w-4" />
            Refresh
          </Button>
          <Button type="button" variant="primary" className="h-10" onClick={openCreate}>
            <Plus className="mr-2 h-4 w-4" />
            Tambah Rekening
          </Button>
        </div>
      </div>

      <div className="grid gap-3 rounded-md border border-white/10 bg-slate-950/70 p-4 shadow-xl md:grid-cols-12">
        <div className="space-y-2 md:col-span-5">
          <label className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">Pencarian</label>
          <Input value={q} onChange={(e) => setQ(e.target.value)} placeholder="Cari nama, nomor rekening, bank, provider" className="h-11 rounded-xl border-white/10 bg-slate-900/80 text-slate-100" />
        </div>
        <div className="space-y-2 md:col-span-3">
          <label className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">Provider</label>
          <select value={provider} onChange={(e) => setProvider(e.target.value)} className="h-11 w-full rounded-xl border border-white/10 bg-slate-900/80 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50">
            <option value="">Semua provider</option>
            {providers.map((p) => (
              <option key={p} value={p}>{providerLabel(p)}</option>
            ))}
          </select>
        </div>
        <div className="flex items-end gap-2 md:col-span-4">
          <label className="mb-3 flex items-center gap-2 text-sm text-slate-200">
            <input type="checkbox" checked={aktifOnly} onChange={(e) => setAktifOnly(e.target.checked)} className="h-4 w-4 rounded border-white/20 bg-slate-900" />
            Aktif saja
          </label>
          <div className="ml-auto flex gap-2">
            <Button type="button" variant="primary" className="h-11" onClick={() => { setOffset(0); setAppliedQ(q); setAppliedProvider(provider); setAppliedAktifOnly(aktifOnly); }} disabled={loading}>
              <Search className="mr-2 h-4 w-4" />
              Cari
            </Button>
            <Button type="button" variant="outline" className="h-11 border-white/10 bg-slate-900/70 text-slate-200" onClick={() => { setQ(""); setProvider(""); setAktifOnly(false); setOffset(0); setAppliedQ(""); setAppliedProvider(""); setAppliedAktifOnly(false); }} disabled={loading}>
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
        emptyText={loading ? "Memuat rekening provider..." : "Belum ada rekening provider."}
        loading={loading}
        minWidthClassName="min-w-250"
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
        title={editing ? "Edit Rekening Provider" : "Tambah Rekening Provider"}
        subtitle="Data ini dipakai untuk identifikasi rekening tujuan provider."
        footer={
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" className="border-white/10 bg-slate-900/70 text-slate-200" onClick={() => setFormOpen(false)} disabled={saving}>Batal</Button>
            <Button type="button" variant="primary" onClick={save} disabled={saving}>{saving ? "Menyimpan..." : "Simpan"}</Button>
          </div>
        }
      >
        <div className="grid gap-3 md:grid-cols-2">
          <div className="space-y-2">
            <label className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">Provider</label>
            <select value={formProvider} onChange={(e) => setFormProvider(e.target.value)} className="h-11 w-full rounded-md border border-white/10 bg-slate-950 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50">
              <option value="">Pilih provider</option>
              {providers.map((p) => (
                <option key={p} value={p}>{providerLabel(p)}</option>
              ))}
            </select>
          </div>
          <div className="space-y-2">
            <label className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">Bank</label>
            <Input value={formBank} onChange={(e) => setFormBank(e.target.value)} placeholder="BCA / BRI / BNI / Mandiri" className="h-11 border-white/10 bg-slate-950 text-slate-100" />
          </div>
          <div className="space-y-2 md:col-span-2">
            <label className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">Nama Rekening</label>
            <Input value={formNama} onChange={(e) => setFormNama(e.target.value)} placeholder="A/N rekening provider" className="h-11 border-white/10 bg-slate-950 text-slate-100" />
          </div>
          <div className="space-y-2">
            <label className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">Nomor Rekening</label>
            <Input value={formNomor} onChange={(e) => setFormNomor(e.target.value)} placeholder="Nomor rekening" className="h-11 border-white/10 bg-slate-950 font-mono text-slate-100" />
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
            <Input value={formCatatan} onChange={(e) => setFormCatatan(e.target.value)} placeholder="Otomatis, manual, konfirmasi CS, atau catatan lain" className="h-11 border-white/10 bg-slate-950 text-slate-100" />
          </div>
        </div>
      </AppModal>
    </div>
  );
}
