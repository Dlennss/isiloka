"use client";

import { useEffect, useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { fmtID } from "@/lib/format";

type GuestRefundPendingRow = {
  id: number;
  app_order_id: number;
  invoice_id: string;
  guest_nama?: string | null;
  guest_email?: string | null;
  guest_phone?: string | null;
  amount_refund: number;
  status: string;
  reason?: string | null;
  produk_nama_snapshot?: string | null;
  dest?: string | null;
  order_status?: string | null;
  dibuat_pada?: string | null;
};

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  if (!t) return {};
  return { Authorization: `Bearer ${t}` };
}

function money(v: number) {
  return `Rp ${fmtID(v || 0)}`;
}

const PAGE_SIZE = 10;

export default function GuestRefundPendingClient({ title }: { title: string }) {
  const [items, setItems] = useState<GuestRefundPendingRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [offset, setOffset] = useState(0);
  const [hasNext, setHasNext] = useState(false);
  const [invoiceFilter, setInvoiceFilter] = useState("");
  const [draftInvoiceFilter, setDraftInvoiceFilter] = useState("");
  const [claimEmail, setClaimEmail] = useState<Record<number, string>>({});
  const [claimingId, setClaimingId] = useState<number | null>(null);
  const [message, setMessage] = useState<string>("");
  const [error, setError] = useState<string>("");

  async function load(nextOffset = offset) {
    setLoading(true);
    setError("");
    try {
      const qs = new URLSearchParams();
      qs.set("limit", String(PAGE_SIZE + 1));
      qs.set("offset", String(nextOffset));
      if (invoiceFilter.trim()) qs.set("invoice_id", invoiceFilter.trim());
      const res = await fetch(`/api/admin/app/guest-refunds?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const json = await res.json().catch(() => ({}));
      if (!res.ok || !json?.ok) {
        throw new Error(json?.error || "Gagal memuat refund guest pending.");
      }
      const rows: GuestRefundPendingRow[] = Array.isArray(json.items) ? json.items : [];
      setHasNext(rows.length > PAGE_SIZE);
      setItems(rows.slice(0, PAGE_SIZE));
    } catch (err) {
      setItems([]);
      setHasNext(false);
      setError(err instanceof Error ? err.message : "Gagal memuat refund guest pending.");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load(offset);
     
  }, [offset]);

  useEffect(() => {
    void load(0);
     
  }, [invoiceFilter]);

  async function handleClaim(row: GuestRefundPendingRow) {
    const targetEmail = (claimEmail[row.id] || "").trim().toLowerCase();
    if (!targetEmail) {
      setError(`Email akun tujuan untuk ${row.invoice_id} wajib diisi.`);
      return;
    }

    setClaimingId(row.id);
    setError("");
    setMessage("");
    try {
      const res = await fetch("/api/admin/app/guest-refunds/claim", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...authHeader(),
        },
        body: JSON.stringify({
          invoice_id: row.invoice_id,
          target_email: targetEmail,
        }),
      });
      const json = await res.json().catch(() => ({}));
      if (!res.ok || !json?.ok) {
        throw new Error(json?.error || "Gagal claim refund guest.");
      }
      setMessage(`Invoice ${row.invoice_id} berhasil di-claim ke ${targetEmail}.`);
      await load(offset);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gagal claim refund guest.");
    } finally {
      setClaimingId(null);
    }
  }

  const columns = useMemo<DataTableColumn<GuestRefundPendingRow>[]>(
    () => [
      {
        id: "dibuat_pada",
        header: "Waktu",
        tdClassName: "whitespace-nowrap text-slate-100",
        render: (x) => (x.dibuat_pada ? new Date(x.dibuat_pada).toLocaleString("id-ID") : "-"),
      },
      {
        id: "invoice_id",
        header: "Invoice",
        tdClassName: "whitespace-nowrap font-mono text-xs text-slate-300",
        render: (x) => x.invoice_id,
      },
      {
        id: "guest",
        header: "Guest",
        tdClassName: "text-slate-200",
        render: (x) => (
          <div>
            <div>{x.guest_nama || "Guest"}</div>
            <div className="text-xs text-slate-400">{x.guest_email || "-"}</div>
            <div className="text-xs text-slate-400">{x.guest_phone || "-"}</div>
          </div>
        ),
      },
      {
        id: "produk",
        header: "Produk",
        tdClassName: "text-slate-200",
        render: (x) => (
          <div>
            <div>{x.produk_nama_snapshot || "-"}</div>
            <div className="text-xs text-slate-400">{x.dest || "-"}</div>
          </div>
        ),
      },
      {
        id: "amount_refund",
        header: "Refund",
        tdClassName: "whitespace-nowrap text-slate-200",
        render: (x) => money(x.amount_refund),
      },
      {
        id: "aksi",
        header: "Claim ke User",
        tdClassName: "min-w-[240px]",
        render: (x) => (
          <div className="space-y-2">
            <Input
              value={claimEmail[x.id] || ""}
              onChange={(e) => setClaimEmail((prev) => ({ ...prev, [x.id]: e.target.value }))}
              placeholder="email user tujuan"
              className="h-9 border-white/10 bg-slate-950/50 text-slate-100 placeholder:text-slate-500"
            />
            <Button
              type="button"
              size="sm"
              disabled={claimingId === x.id}
              onClick={() => void handleClaim(x)}
              className="h-8 bg-emerald-600 hover:bg-emerald-500"
            >
              {claimingId === x.id ? "Memproses..." : "Claim"}
            </Button>
          </div>
        ),
      },
    ],
    [claimEmail, claimingId],
  );

  return (
    <section className="space-y-4">
      <div className="space-y-1">
        <h1 className="text-2xl font-semibold tracking-tight text-slate-50">{title}</h1>
        <div className="text-sm text-slate-400">Daftar refund guest yang masih pending claim dan bisa dipindahkan ke akun user tertentu.</div>
      </div>

      <div className="rounded-2xl border border-white/10 bg-slate-900/70 p-4 shadow-[0_18px_38px_-20px_rgba(15,23,42,0.85)]">
        <div className="flex flex-col gap-3 md:flex-row md:items-end">
          <div className="flex-1">
            <label className="mb-1 block text-sm text-slate-300">Cari Invoice</label>
            <Input
              value={draftInvoiceFilter}
              onChange={(e) => setDraftInvoiceFilter(e.target.value)}
              placeholder="INV-..."
              className="border-white/10 bg-slate-950/60 text-slate-100 placeholder:text-slate-500"
            />
          </div>
          <div className="flex gap-2">
            <Button
              type="button"
              className="bg-sky-600 hover:bg-sky-500"
              onClick={() => {
                setOffset(0);
                setInvoiceFilter(draftInvoiceFilter.trim());
              }}
            >
              Terapkan
            </Button>
            <Button
              type="button"
              variant="outline"
              className="border-white/10 bg-transparent text-slate-200 hover:bg-white/5"
              onClick={() => {
                setDraftInvoiceFilter("");
                setOffset(0);
                setInvoiceFilter("");
              }}
            >
              Reset
            </Button>
          </div>
        </div>
      </div>

      {message ? <div className="rounded-xl border border-emerald-500/30 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-200">{message}</div> : null}
      {error ? <div className="rounded-xl border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-sm text-rose-200">{error}</div> : null}

      <DataTable<GuestRefundPendingRow>
        rows={items}
        columns={columns}
        loading={loading}
        emptyText={loading ? "Memuat refund guest pending..." : "Tidak ada refund guest pending."}
        rowKey={(row) => row.id}
        pagination={{
          page: Math.floor(offset / PAGE_SIZE) + 1,
          totalPages: hasNext ? Math.floor(offset / PAGE_SIZE) + 2 : Math.max(1, Math.floor(offset / PAGE_SIZE) + 1),
          onPrev: () => setOffset((prev) => Math.max(0, prev - PAGE_SIZE)),
          onNext: () => setOffset((prev) => prev + PAGE_SIZE),
          disablePrev: offset === 0 || loading,
          disableNext: !hasNext || loading,
        }}
      />
    </section>
  );
}
