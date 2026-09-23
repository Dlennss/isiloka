"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableActions, type DataTableColumn } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { AppModal } from "@/components/ui/app-modal";
import { alertConfirm, alertError, alertSuccess, alertWarning } from "@/components/ui/alerts";
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, Eye, Pencil, Plus, RefreshCw, Search, Trash2 } from "lucide-react";

type ItemRow = {
  id: number;
  nama: string;
  aktif: boolean;
  keterangan?: string;
  dibuat_pada: string;
  diubah_pada: string;
};

type Props = {
  title: string;
  endpoint: string; // ex: /api/admin/master/kategori
  emptyLabel: string;
  showKeterangan?: boolean;
};

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  if (!t) return {};
  return { Authorization: `Bearer ${t}` };
}

export default function MasterSimpleCrud({ title, endpoint, emptyLabel, showKeterangan = false }: Props) {
  const [items, setItems] = useState<ItemRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [qInput, setQInput] = useState("");
  const [q, setQ] = useState("");
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(10);

  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<ItemRow | null>(null);
  const [nama, setNama] = useState("");
  const [aktif, setAktif] = useState(true);
  const [saving, setSaving] = useState(false);
  const [openActionKey, setOpenActionKey] = useState<string | null>(null);

  async function load() {
    setLoading(true);
    try {
      const r = await fetch(endpoint, { headers: authHeader(), cache: "no-store" });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || `Gagal memuat ${title.toLowerCase()}`);
        setItems([]);
        return;
      }
      setItems(Array.isArray(j.items) ? j.items : []);
    } finally {
      setLoading(false);
    }
  }

  function applySearch() {
    setQ(qInput.trim());
    setPage(1);
  }

  useEffect(() => {
    load();
     
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
    setNama("");
    setAktif(true);
    setOpen(true);
  }

  function openEdit(x: ItemRow) {
    setEditing(x);
    setNama(x.nama || "");
    setAktif(Boolean(x.aktif));
    setOpen(true);
  }

  async function save() {
    const nm = nama.trim();
    if (!nm) return alertWarning("Nama wajib diisi.");

    setSaving(true);
    try {
      const isEdit = Boolean(editing?.id);
      const url = isEdit ? `${endpoint}/${editing?.id}` : endpoint;
      const method = isEdit ? "PUT" : "POST";

      const r = await fetch(url, {
        method,
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({ nama: nm, aktif }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) return alertError(j.error || "Gagal menyimpan data");

      setOpen(false);
      await load();
      await alertSuccess(isEdit ? "Data berhasil diperbarui." : "Data berhasil ditambahkan.");
    } finally {
      setSaving(false);
    }
  }

  async function remove(x: ItemRow) {
    const ok = await alertConfirm({
      title: "Hapus data",
      text: `Yakin hapus ${title.toLowerCase()} \"${x.nama}\"?`,
      confirmButtonText: "Ya, hapus",
    });
    if (!ok) return;

    const r = await fetch(`${endpoint}/${x.id}`, {
      method: "DELETE",
      headers: authHeader(),
    });
    const j = await r.json().catch(() => ({}));
    if (!r.ok || !j.ok) return alertError(j.error || "Gagal menghapus data");

    await load();
    await alertSuccess("Data berhasil dihapus.");
  }

  const filteredItems = items.filter((x) => {
    const keyword = q.trim().toLowerCase();
    if (!keyword) return true;
    return (
      x.nama.toLowerCase().includes(keyword) ||
      String(x.id).includes(keyword) ||
      (showKeterangan && String(x.keterangan || "").toLowerCase().includes(keyword))
    );
  });
  const total = filteredItems.length;
  const totalPages = Math.max(1, Math.ceil(total / limit));
  const currentPage = Math.min(page, totalPages);
  const start = (currentPage - 1) * limit;
  const end = start + limit;
  const pageRows = filteredItems.slice(start, end);

  const columns: DataTableColumn<ItemRow>[] = [
    {
      id: "id",
      header: "ID",
      tdClassName: "whitespace-nowrap font-mono text-xs text-slate-400",
      render: (x) => String(x.id),
    },
    {
      id: "nama",
      header: "Nama",
      tdClassName: "whitespace-nowrap text-slate-100",
      render: (x) => x.nama,
    },
    {
      id: "aktif",
      header: "Status",
      tdClassName: "whitespace-nowrap",
      render: (x) =>
        x.aktif ? (
          <span className="rounded-full border border-emerald-400/40 bg-emerald-500/15 px-2 py-0.5 text-xs font-semibold text-emerald-300">Aktif</span>
        ) : (
          <span className="rounded-full border border-rose-400/40 bg-rose-500/15 px-2 py-0.5 text-xs font-semibold text-rose-300">Nonaktif</span>
        ),
    },
    ...(showKeterangan
      ? [
          {
            id: "keterangan",
            header: "Keterangan",
            tdClassName: "min-w-[220px] max-w-[360px] text-slate-300",
            render: (x) => {
              const text = String(x.keterangan || "").trim();
              return text ? <span className="block whitespace-normal break-words text-sm">{text}</span> : <span className="text-slate-500">-</span>;
            },
          } satisfies DataTableColumn<ItemRow>,
        ]
      : []),
    {
      id: "dibuat",
      header: "Dibuat",
      tdClassName: "whitespace-nowrap text-slate-300",
      render: (x) => new Date(x.dibuat_pada).toLocaleString("id-ID"),
    },
    {
      id: "diubah",
      header: "Diubah",
      tdClassName: "whitespace-nowrap text-slate-300",
      render: (x) => new Date(x.diubah_pada).toLocaleString("id-ID"),
    },
  ];

  const actions: DataTableActions<ItemRow> = {
    header: "Aksi",
    align: "right",
    tdClassName: "whitespace-nowrap",
    render: (x, index) => (
      <div className="relative inline-flex" data-action-dropdown>
        <Button
          size="sm"
          variant="info"
          className="h-8 w-8 px-0"
          onClick={() => setOpenActionKey((prev) => (prev === String(x.id) ? null : String(x.id)))}
          aria-label="Aksi"
        >
          <Eye className="h-4 w-4" />
        </Button>

        {openActionKey === String(x.id) ? (
          <div
            className={`absolute right-0 z-20 w-36 rounded-xl border border-white/12 bg-slate-900/95 p-1.5 shadow-2xl backdrop-blur ${
              index >= pageRows.length - 2 ? "bottom-10" : "top-10"
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
              className="mt-1 flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm text-rose-300 transition hover:bg-rose-500/15"
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
          <div className="text-lg font-semibold tracking-tight">{title}</div>
          <div className="text-sm text-muted-foreground">Kelola data master {title.toLowerCase()}.</div>
        </div>

        <div className="flex gap-2">
          <Button variant="outline" className="h-10" onClick={() => load()} disabled={loading}>
            <RefreshCw className={`mr-2 h-4 w-4 ${loading ? "animate-spin" : ""}`} />
            Refresh
          </Button>
          <Button variant="primary" className="h-10" onClick={openCreate}>
            <Plus className="mr-2 h-4 w-4" />
            Tambah
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-2 rounded-md border border-white/15 bg-slate-950/50 p-3 sm:grid-cols-2 lg:grid-cols-12">
        <div className="sm:col-span-2 lg:col-span-8">
          <div className="mb-1 text-xs text-muted-foreground">Cari</div>
          <div className="relative">
            <Search className="pointer-events-none absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-slate-400" />
            <Input
              value={qInput}
              onChange={(e) => setQInput(e.target.value)}
              placeholder={`Cari ${title.toLowerCase()}...`}
              className="pl-9"
              onKeyDown={(e) => {
                if (e.key === "Enter") applySearch();
              }}
            />
          </div>
        </div>
        <div className="sm:col-span-1 lg:col-span-2">
          <div className="mb-1 hidden text-xs text-muted-foreground sm:block">Limit</div>
          <select
            className="h-10 w-full rounded-md border border-input bg-transparent px-3 text-sm"
            value={String(limit)}
            onChange={(e) => {
              setLimit(Number(e.target.value));
              setPage(1);
            }}
          >
            <option value="5">5</option>
            <option value="10">10</option>
            <option value="20">20</option>
            <option value="50">50</option>
            <option value="100">100</option>
          </select>
        </div>
        <div className="sm:col-span-1 lg:col-span-2">
          <div className="mb-1 hidden text-xs text-muted-foreground sm:block opacity-0 select-none">Aksi</div>
          <div className="flex gap-2 lg:justify-end">
            <Button variant="outline" className="h-10" onClick={applySearch}>
              Cari
            </Button>
            <Button
              variant="outline"
              className="h-10"
              onClick={() => {
                setQInput("");
                setQ("");
                setPage(1);
              }}
            >
              Reset
            </Button>
          </div>
        </div>
      </div>

      <DataTable
        columns={columns}
        rows={pageRows}
        rowKey={(x) => x.id}
        emptyText={loading ? "Loading..." : emptyLabel}
        rowNumberStart={start + 1}
        actions={actions}
      />

      <div className="flex flex-wrap items-center justify-between gap-3 rounded-md border border-white/15 bg-linear-to-r from-slate-900/80 via-slate-900/65 to-cyan-950/25 p-3 text-sm">
        <div className="text-slate-300">
          Total {total.toLocaleString("id-ID")} data, halaman {currentPage} / {totalPages}
        </div>
        <div className="flex flex-wrap items-center gap-1.5">
          <Button variant="outline" size="sm" onClick={() => setPage(1)} disabled={currentPage <= 1} aria-label="Halaman awal">
            <ChevronsLeft className="h-4 w-4" />
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={currentPage <= 1}
            aria-label="Halaman sebelumnya"
          >
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            disabled={currentPage >= totalPages}
            aria-label="Halaman berikutnya"
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage(totalPages)}
            disabled={currentPage >= totalPages}
            aria-label="Halaman akhir"
          >
            <ChevronsRight className="h-4 w-4" />
          </Button>
        </div>
      </div>

      <AppModal
        open={open}
        onClose={() => setOpen(false)}
        title={editing ? `Edit ${title}` : `Tambah ${title}`}
        subtitle="Perubahan akan langsung diterapkan ke data master."
        maxWidthClassName="max-w-md"
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
            <div className="mb-1 text-xs text-muted-foreground">Nama</div>
            <Input value={nama} onChange={(e) => setNama(e.target.value)} placeholder={`Nama ${title.toLowerCase()}`} />
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
