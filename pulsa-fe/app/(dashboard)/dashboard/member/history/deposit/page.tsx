"use client";
import { fmtID } from "@/lib/format";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";
import { EmptyState } from "@/components/ui/empty-state";
import { Landmark } from "lucide-react";

import { useCallback, useEffect, useState } from "react";

type DepositRow = {
  id: number;
  member_id: number;
  bank_id?: number | null;
  bank_nama?: string;
  bank_nomor_rekening?: string;
  bank_atas_nama?: string;
  amount: number;
  requested_amount?: number;
  unique_code?: number;
  approved_amount?: number;
  metode: string;
  bukti_url: string;
  status: string;
  note: string;
  dibuat_pada: string;
};

type ApiResp = {
  ok: boolean;
  rows?: DepositRow[];
  error?: string;
};

async function apiFetch(path: string, init?: RequestInit) {
  const t = localStorage.getItem("auth_token") || "";
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (t) headers.Authorization = `Bearer ${t}`;
  const res = await fetch(path, { ...init, headers: { ...headers, ...(init?.headers || {}) }, cache: "no-store" });
  const data = await res.json().catch(() => ({}));
  return { ok: res.ok, status: res.status, data };
}

function saldoMasuk(row: DepositRow) {
  const status = String(row.status || "").toLowerCase();
  const approved = Number(row.approved_amount || 0);
  if (status === "approved" && approved > 0) {
    return approved;
  }
  return Number(row.amount || 0);
}

export default function HistoryDepositPage() {
  const [rows, setRows] = useState<DepositRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState<string>("");

  const load = useCallback(async () => {
    setErr("");
    setLoading(true);
    try {
      const r = await apiFetch("/api/me/history/deposit?limit=50");
      const j: ApiResp = r.data || ({} as ApiResp);
      if (!r.ok || !j.ok) {
        setErr(j.error || "Gagal mengambil data");
        setRows([]);
        return;
      }
      setRows(j.rows || []);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const columns: DataTableColumn<DepositRow>[] = [
    { id: "id", header: "ID", render: (x) => x.id },
    {
      id: "waktu",
      header: "Waktu",
      render: (x) => new Date(x.dibuat_pada).toLocaleString("id-ID"),
    },
    { id: "metode", header: "Metode", render: (x) => x.metode },
    {
      id: "bank",
      header: "Bank Tujuan",
      render: (x) => (
        <div className="text-white/85">
          <div>{x.bank_nama || "-"}</div>
          <div className="text-xs text-white/55">{x.bank_atas_nama || x.bank_nomor_rekening ? `${x.bank_atas_nama || "-"} • ${x.bank_nomor_rekening || "-"}` : "-"}</div>
        </div>
      ),
    },
    {
      id: "amount",
      header: "Saldo Masuk",
      render: (x) => {
        const credited = saldoMasuk(x);
        const corrected = String(x.status || "").toLowerCase() === "approved" && credited !== Number(x.amount || 0);
        return (
          <div className="font-semibold text-white">
            <div>Rp {fmtID(credited)}</div>
            {corrected ? <div className="text-xs font-medium text-amber-200">Tiket Rp {fmtID(x.amount)}</div> : null}
          </div>
        );
      },
    },
    { id: "status", header: "Status", render: (x) => x.status },
    { id: "note", header: "Note", render: (x) => x.note || "-" },
  ];

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <div className="text-xs text-white/60">History</div>
          <h1 className="text-lg font-semibold">Deposit</h1>
        </div>
        <button
          onClick={load}
          className="rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-sm text-white/80 hover:bg-white/10"
          disabled={loading}
        >
          {loading ? "Memuat..." : "Refresh"}
        </button>
      </div>

      {err ? (
        <div className="rounded-xl border border-red-500/30 bg-red-500/10 p-3 text-sm text-red-200">{err}</div>
      ) : null}

      {rows.length > 0 ? (
        <DataTable<DepositRow>
          columns={columns}
          rows={rows}
          rowKey={(x) => x.id}
          emptyText="Belum ada request deposit"
          minWidthClassName="min-w-150"
          showRowNumber={false}
          wrapperClassName="overflow-auto rounded-xl border border-white/10 bg-slate-950/35"
        />
      ) : !loading ? (
        <EmptyState
          title="Belum ada data"
          description="Riwayat request deposit akan tampil di sini."
          icon={<Landmark className="h-6 w-6" />}
        />
      ) : null}
    </div>
  );
}
