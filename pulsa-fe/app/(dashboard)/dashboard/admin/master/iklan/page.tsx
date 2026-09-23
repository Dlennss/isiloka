"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableActions, type DataTableColumn } from "@/components/ui/data-table";
import { AppModal } from "@/components/ui/app-modal";
import { Input } from "@/components/ui/input";
import { alertConfirm, alertError, alertSuccess, alertWarning } from "@/components/ui/alerts";
import { Eye, Pencil, Plus, RefreshCw, Trash2 } from "lucide-react";

type AdRow = {
  id: number;
  judul: string;
  keterangan: string;
  image_url: string;
  link_url: string;
  urutan: number;
  aktif: boolean;
  created_at?: string;
  updated_at?: string;
};

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

export default function AdminMasterIklanPage() {
  const [items, setItems] = useState<AdRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [open, setOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [editing, setEditing] = useState<AdRow | null>(null);
  const [judul, setJudul] = useState("");
  const [keterangan, setKeterangan] = useState("");
  const [imageUrl, setImageUrl] = useState("");
  const [imageFile, setImageFile] = useState<File | null>(null);
  const [linkUrl, setLinkUrl] = useState("");
  const [urutan, setUrutan] = useState("0");
  const [aktif, setAktif] = useState(true);
  const [openActionKey, setOpenActionKey] = useState<string | null>(null);

  async function loadItems() {
    setLoading(true);
    try {
      const r = await fetch("/api/admin/master/iklan", {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal memuat iklan");
        setItems([]);
        return;
      }
      setItems(Array.isArray(j.items) ? j.items : []);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void loadItems();
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

  function resetForm() {
    setJudul("");
    setKeterangan("");
    setImageUrl("");
    setImageFile(null);
    setLinkUrl("");
    setUrutan("0");
    setAktif(true);
  }

  function openCreate() {
    setEditing(null);
    resetForm();
    setOpen(true);
  }

  function openEdit(x: AdRow) {
    setEditing(x);
    setJudul(x.judul || "");
    setKeterangan(x.keterangan || "");
    setImageUrl(x.image_url || "");
    setImageFile(null);
    setLinkUrl(x.link_url || "");
    setUrutan(String(x.urutan ?? 0));
    setAktif(Boolean(x.aktif));
    setOpen(true);
  }

  async function save() {
    if (!imageUrl.trim() && !imageFile) return alertWarning("Gambar wajib diisi.");
    if (!judul.trim() && !keterangan.trim()) return alertWarning("Judul atau keterangan wajib diisi.");
    const order = Number(urutan || "0");
    if (!Number.isFinite(order) || order < 0) return alertWarning("Urutan tidak valid.");

    setSaving(true);
    try {
      const isEdit = Boolean(editing?.id);
      const url = isEdit ? `/api/admin/master/iklan/${editing?.id}` : "/api/admin/master/iklan";
      const method = isEdit ? "PUT" : "POST";
      const formData = new FormData();
      formData.set("judul", judul.trim());
      formData.set("keterangan", keterangan.trim());
      formData.set("image_url", imageUrl.trim());
      formData.set("link_url", linkUrl.trim());
      formData.set("urutan", String(Math.floor(order)));
      formData.set("aktif", aktif ? "true" : "false");
      if (imageFile) {
        formData.set("image", imageFile);
        formData.set("delete_old_image", "1");
      }
      const r = await fetch(url, {
        method,
        headers: { ...authHeader() },
        body: formData,
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) return alertError(j.error || "Gagal menyimpan iklan");

      setOpen(false);
      await loadItems();
      await alertSuccess(isEdit ? "Iklan berhasil diupdate." : "Iklan berhasil ditambahkan.");
    } finally {
      setSaving(false);
    }
  }

  async function removeItem(x: AdRow) {
    const ok = await alertConfirm({
      title: "Hapus iklan",
      text: `Yakin hapus iklan "${x.judul || "Tanpa Judul"}"?`,
      confirmButtonText: "Ya, hapus",
    });
    if (!ok) return;

    const r = await fetch(`/api/admin/master/iklan/${x.id}`, {
      method: "DELETE",
      headers: authHeader(),
    });
    const j = await r.json().catch(() => ({}));
    if (!r.ok || !j.ok) return alertError(j.error || "Gagal menghapus iklan");

    await loadItems();
    await alertSuccess("Iklan berhasil dihapus.");
  }

  const columns: DataTableColumn<AdRow>[] = [
    {
      id: "image",
      header: "Gambar",
      tdClassName: "whitespace-nowrap",
      render: (x) => (
        <div className="h-14 w-24 overflow-hidden rounded-xl bg-slate-100 ring-1 ring-white/10">
          <img src={x.image_url} alt={x.judul || "Iklan"} className="h-full w-full object-cover" />
        </div>
      ),
    },
    {
      id: "judul",
      header: "Judul",
      tdClassName: "min-w-[180px] text-slate-100",
      render: (x) => x.judul || "-",
    },
    {
      id: "keterangan",
      header: "Keterangan",
      tdClassName: "min-w-[220px] text-slate-300",
      render: (x) => (
        <div className="max-w-[260px] whitespace-normal text-sm leading-5 text-slate-300">
          {x.keterangan || "-"}
        </div>
      ),
    },
    {
      id: "urutan",
      header: "Urutan",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (x) => String(x.urutan ?? 0),
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

  const actions: DataTableActions<AdRow> = {
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
          <div className="text-lg font-semibold tracking-tight">Iklan Guest</div>
          <div className="text-sm text-muted-foreground">Kelola banner iklan yang tampil otomatis di halaman guest.</div>
        </div>

        <div className="flex gap-2">
          <Button variant="outline" className="h-10" onClick={() => loadItems()} disabled={loading}>
            <RefreshCw className={`mr-2 h-4 w-4 ${loading ? "animate-spin" : ""}`} />
            Refresh
          </Button>
          <Button className="h-10" onClick={openCreate}>
            <Plus className="mr-2 h-4 w-4" />
            Tambah Iklan
          </Button>
        </div>
      </div>

      <DataTable
        columns={columns}
        rows={items}
        rowKey={(row) => row.id}
        actions={actions}
        emptyMessage="Belum ada iklan."
        loading={loading}
      />

      <AppModal
        open={open}
        onClose={() => {
          if (saving) return;
          setOpen(false);
        }}
        title={editing ? "Edit Iklan" : "Tambah Iklan"}
        maxWidthClassName="max-w-2xl"
        footer={
          <>
            <Button variant="outline" onClick={() => setOpen(false)} disabled={saving}>
              Batal
            </Button>
            <Button onClick={() => save()} disabled={saving}>
              {saving ? "Menyimpan..." : editing ? "Simpan Perubahan" : "Tambah Iklan"}
            </Button>
          </>
        }
      >
        <div className="grid gap-4">
          <div className="grid gap-2">
            <label className="text-sm font-medium text-slate-200">Judul</label>
            <Input value={judul} onChange={(e) => setJudul(e.target.value)} placeholder="Promo Pulsa Harian" />
          </div>

          <div className="grid gap-2">
            <label className="text-sm font-medium text-slate-200">Keterangan</label>
            <textarea
              value={keterangan}
              onChange={(e) => setKeterangan(e.target.value)}
              placeholder="Isi keterangan singkat yang tampil di atas gambar."
              className="min-h-24 rounded-xl border border-emerald-200 bg-white px-3 py-3 text-sm text-emerald-950 outline-none transition placeholder:text-emerald-700/50 focus:border-emerald-500"
            />
          </div>

          <div className="grid gap-2">
            <label className="text-sm font-medium text-slate-200">Upload Gambar</label>
            <Input
              type="file"
              accept="image/png,image/jpeg,image/webp"
              onChange={(e) => {
                const nextFile = e.target.files?.[0] || null;
                setImageFile(nextFile);
                if (nextFile) {
                  setImageUrl(URL.createObjectURL(nextFile));
                }
              }}
            />
            <div className="text-xs text-slate-400">Format: JPG, PNG, WEBP. Maksimal 10MB.</div>
          </div>

          <div className="grid gap-2">
            <label className="text-sm font-medium text-slate-200">URL Gambar Manual</label>
            <Input value={imageUrl} onChange={(e) => { setImageUrl(e.target.value); if (e.target.value.trim()) setImageFile(null); }} placeholder="https://... atau kosongkan jika upload file" />
          </div>

          <div className="grid gap-2">
            <label className="text-sm font-medium text-slate-200">URL Tujuan</label>
            <Input value={linkUrl} onChange={(e) => setLinkUrl(e.target.value)} placeholder="https://... atau kosongkan" />
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div className="grid gap-2">
              <label className="text-sm font-medium text-slate-200">Urutan</label>
              <Input value={urutan} onChange={(e) => setUrutan(e.target.value)} inputMode="numeric" />
            </div>

            <label className="mt-7 inline-flex items-center gap-2 text-sm text-slate-200">
              <input
                type="checkbox"
                checked={aktif}
                onChange={(e) => setAktif(e.target.checked)}
                className="h-4 w-4 rounded border-emerald-300 bg-white text-emerald-600"
              />
              Iklan aktif
            </label>
          </div>

          {imageUrl.trim() ? (
            <div className="overflow-hidden rounded-2xl ring-1 ring-white/10">
              <img src={imageUrl.trim()} alt={judul || "Preview iklan"} className="h-48 w-full object-cover" />
            </div>
          ) : null}
        </div>
      </AppModal>
    </div>
  );
}
