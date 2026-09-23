"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableActions, type DataTableColumn } from "@/components/ui/data-table";
import { AppModal } from "@/components/ui/app-modal";
import { Input } from "@/components/ui/input";
import { alertConfirm, alertError, alertSuccess, alertWarning } from "@/components/ui/alerts";
import { Eye, Pencil, Plus, RefreshCw, Trash2 } from "lucide-react";

type KategoriFeeRow = {
  id: number;
  kategori_id: number;
  kategori_nama: string;
  fee_master: number;
  fee_agent: number;
  fee_user: number;
  fee_non_user: number;
  aktif: boolean;
  created_at?: string;
  updated_at?: string;
};

type KategoriRow = {
  id: number;
  nama: string;
  aktif: boolean;
};

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

function fmtIDR(n: number): string {
  return new Intl.NumberFormat("id-ID").format(Number.isFinite(n) ? n : 0);
}

function normalizeFeeInput(value: string): string {
  const cleaned = value.replace(/[^\d-]/g, "");
  const negative = cleaned.startsWith("-");
  const digits = cleaned.replace(/-/g, "");
  return `${negative ? "-" : ""}${digits}`;
}

export default function FeeKategoriAplikasiPage() {
  const [items, setItems] = useState<KategoriFeeRow[]>([]);
  const [kategori, setKategori] = useState<KategoriRow[]>([]);
  const [loading, setLoading] = useState(false);

  const [open, setOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [editing, setEditing] = useState<KategoriFeeRow | null>(null);
  const [kategoriID, setKategoriID] = useState("");
  const [feeMaster, setFeeMaster] = useState("0");
  const [feeAgent, setFeeAgent] = useState("0");
  const [feeUser, setFeeUser] = useState("0");
  const [feeNonUser, setFeeNonUser] = useState("0");
  const [aktif, setAktif] = useState(true);
  const [openActionKey, setOpenActionKey] = useState<string | null>(null);

  async function loadItems() {
    setLoading(true);
    try {
      const r = await fetch("/api/admin/master/kategori-fee-app", {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal memuat fee kategori aplikasi");
        setItems([]);
        return;
      }
      setItems(Array.isArray(j.items) ? j.items : []);
    } finally {
      setLoading(false);
    }
  }

  async function loadKategori() {
    const r = await fetch("/api/admin/master/kategori", {
      headers: authHeader(),
      cache: "no-store",
    });
    const j = await r.json().catch(() => ({}));
    const rows = Array.isArray(j.items) ? j.items : [];
    setKategori(
      rows.filter((x: KategoriRow) => {
        const nama = String(x?.nama || "").toLowerCase();
        return x?.aktif !== false || nama === "bebas nominal";
      })
    );
  }

  useEffect(() => {
    void Promise.all([loadItems(), loadKategori()]);
  }, []);

  useEffect(() => {
    function handleOutsideClick(e: MouseEvent) {
      const target = e.target as HTMLElement | null;
      if (!target) return;
      if (target.closest("[data-action-dropdown]")) return;
      setOpenActionKey(null);
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
    setKategoriID("");
    setFeeMaster("0");
    setFeeAgent("0");
    setFeeUser("0");
    setFeeNonUser("0");
    setAktif(true);
    setOpen(true);
  }

  function openEdit(x: KategoriFeeRow) {
    setEditing(x);
    setKategoriID(String(x.kategori_id));
    setFeeMaster(String(x.fee_master || 0));
    setFeeAgent(String(x.fee_agent || 0));
    setFeeUser(String(x.fee_user));
    setFeeNonUser(String(x.fee_non_user));
    setAktif(Boolean(x.aktif));
    setOpen(true);
  }

  async function save() {
    const kID = Number(kategoriID || "0");
    const fMaster = Number(feeMaster || "0");
    const fAgent = Number(feeAgent || "0");
    const fUser = Number(feeUser || "0");
    const fNonUser = Number(feeNonUser || "0");

    if (!kID || kID <= 0) return alertWarning("Kategori wajib dipilih.");
    if (!Number.isFinite(fMaster)) return alertWarning("Fee master tidak valid.");
    if (!Number.isFinite(fAgent)) return alertWarning("Fee agent tidak valid.");
    if (!Number.isFinite(fUser)) return alertWarning("Fee user tidak valid.");
    if (!Number.isFinite(fNonUser)) return alertWarning("Fee non-user tidak valid.");

    setSaving(true);
    try {
      const isEdit = Boolean(editing?.id);
      const url = isEdit ? `/api/admin/master/kategori-fee-app/${editing?.id}` : "/api/admin/master/kategori-fee-app";
      const method = isEdit ? "PUT" : "POST";
      const r = await fetch(url, {
        method,
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({
          kategori_id: kID,
          fee_master: Math.floor(fMaster),
          fee_agent: Math.floor(fAgent),
          fee_user: Math.floor(fUser),
          fee_non_user: Math.floor(fNonUser),
          aktif,
        }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) return alertError(j.error || "Gagal menyimpan data");

      setOpen(false);
      await loadItems();
      await alertSuccess(isEdit ? "Fee kategori berhasil diupdate." : "Fee kategori berhasil ditambahkan.");
    } finally {
      setSaving(false);
    }
  }

  async function removeItem(x: KategoriFeeRow) {
    const ok = await alertConfirm({
      title: "Hapus fee kategori",
      text: `Yakin hapus fee untuk kategori "${x.kategori_nama}"?`,
      confirmButtonText: "Ya, hapus",
    });
    if (!ok) return;

    const r = await fetch(`/api/admin/master/kategori-fee-app/${x.id}`, {
      method: "DELETE",
      headers: authHeader(),
    });
    const j = await r.json().catch(() => ({}));
    if (!r.ok || !j.ok) return alertError(j.error || "Gagal menghapus data");

    await loadItems();
    await alertSuccess("Data berhasil dihapus.");
  }

  const columns: DataTableColumn<KategoriFeeRow>[] = [
    {
      id: "kategori",
      header: "Kategori",
      tdClassName: "whitespace-nowrap text-slate-100",
      render: (x) => x.kategori_nama,
    },
    {
      id: "feeMaster",
      header: "Fee Master",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (x) => `Rp ${fmtIDR(x.fee_master || 0)}`,
    },
    {
      id: "feeAgent",
      header: "Fee Agent",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (x) => `Rp ${fmtIDR(x.fee_agent || 0)}`,
    },
    {
      id: "feeUser",
      header: "Fee User",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (x) => `Rp ${fmtIDR(x.fee_user || 0)}`,
    },
    {
      id: "feeNonUser",
      header: "Fee Non User",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (x) => `Rp ${fmtIDR(x.fee_non_user || 0)}`,
    },
    {
      id: "status",
      header: "Status",
      tdClassName: "whitespace-nowrap",
      render: (x) =>
        x.aktif ? (
          <span className="rounded-full border border-emerald-400 bg-emerald-100 px-2 py-0.5 text-xs font-bold text-emerald-900">Aktif</span>
        ) : (
          <span className="rounded-full border border-rose-400 bg-rose-100 px-2 py-0.5 text-xs font-bold text-rose-900">Nonaktif</span>
        ),
    },
    {
      id: "updated",
      header: "Diubah",
      tdClassName: "whitespace-nowrap text-slate-300",
      render: (x) => (x.updated_at ? new Date(x.updated_at).toLocaleString("id-ID") : "-"),
    },
  ];

  const actions: DataTableActions<KategoriFeeRow> = {
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
                void removeItem(x);
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
          <div className="text-lg font-semibold tracking-tight">Fee Kategori Aplikasi</div>
          <div className="text-sm text-muted-foreground">Kelola fee master, agent, user, dan guest per kategori untuk app commerce.</div>
        </div>

        <div className="flex gap-2">
          <Button variant="outline" className="h-10" onClick={() => void loadItems()} disabled={loading}>
            <RefreshCw className={`mr-2 h-4 w-4 ${loading ? "animate-spin" : ""}`} />
            Refresh
          </Button>
          <Button variant="primary" className="h-10" onClick={openCreate}>
            <Plus className="mr-2 h-4 w-4" />
            Tambah
          </Button>
        </div>
      </div>

      <DataTable
        columns={columns}
        rows={items}
        rowKey={(x) => x.id}
        emptyText={loading ? "Loading..." : "Belum ada fee kategori aplikasi."}
        actions={actions}
      />

      <AppModal
        open={open}
        onClose={() => setOpen(false)}
        title={editing ? "Edit Fee Kategori Aplikasi" : "Tambah Fee Kategori Aplikasi"}
        subtitle="Perubahan ini akan dipakai untuk perhitungan harga jual aplikasi."
        maxWidthClassName="max-w-lg"
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
        <div className="space-y-3">
          <div>
            <div className="mb-1 text-xs text-muted-foreground">Kategori</div>
            <select
              value={kategoriID}
              onChange={(e) => setKategoriID(e.target.value)}
              className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm text-foreground outline-none ring-offset-background transition focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            >
              <option value="">Pilih kategori</option>
              {kategori.map((k) => (
                <option key={k.id} value={String(k.id)}>
                  {k.nama}
                </option>
              ))}
            </select>
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <div className="mb-1 text-xs text-muted-foreground">Fee Master</div>
              <Input
                value={feeMaster}
                onChange={(e) => setFeeMaster(normalizeFeeInput(e.target.value))}
                inputMode="text"
                placeholder="0"
              />
            </div>
            <div>
              <div className="mb-1 text-xs text-muted-foreground">Fee Agent</div>
              <Input
                value={feeAgent}
                onChange={(e) => setFeeAgent(normalizeFeeInput(e.target.value))}
                inputMode="text"
                placeholder="0"
              />
            </div>
            <div>
              <div className="mb-1 text-xs text-muted-foreground">Fee User</div>
              <Input
                value={feeUser}
                onChange={(e) => setFeeUser(normalizeFeeInput(e.target.value))}
                inputMode="text"
                placeholder="0"
              />
            </div>
            <div>
              <div className="mb-1 text-xs text-muted-foreground">Fee Guest</div>
              <Input
                value={feeNonUser}
                onChange={(e) => setFeeNonUser(normalizeFeeInput(e.target.value))}
                inputMode="text"
                placeholder="0"
              />
            </div>
          </div>

          <label className="flex items-center gap-2 text-sm text-slate-200">
            <input type="checkbox" checked={aktif} onChange={(e) => setAktif(e.target.checked)} />
            Aktif
          </label>
        </div>
      </AppModal>
    </div>
  );
}
