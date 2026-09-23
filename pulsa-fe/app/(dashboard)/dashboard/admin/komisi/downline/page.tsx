"use client";

import { useEffect, useMemo, useState } from "react";
import { Coins, TrendingUp } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";
import { DateField } from "@/components/ui/date-field";
import { fmtID } from "@/lib/format";

type Scope = "all" | "retail" | "h2h";

type CommissionRow = {
  scope: string;
  upline_member_id: number;
  upline_email: string;
  upline_nama: string;
  upline_role: string;
  level: string;
  source_member_id: number;
  source_email: string;
  source_nama: string;
  source_role: string;
  transaction_count: number;
  total_commission: number;
  last_created_at?: string | null;
};

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

function fmtIDR(v: number) {
  return `Rp ${fmtID(Number(v || 0))}`;
}

function scopeLabel(scope: string) {
  switch ((scope || "").toLowerCase()) {
    case "retail":
      return "Retail";
    case "h2h":
      return "H2H";
    default:
      return "Gabungan";
  }
}

const PAGE_SIZE = 10;

export default function AdminDownlineCommissionPage() {
  const [scope, setScope] = useState<Scope>("all");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [commissions, setCommissions] = useState<CommissionRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState("");
  const [page, setPage] = useState(0);

  async function load(currentScope: Scope, currentFrom = from, currentTo = to) {
    setLoading(true);
    setErr("");
    try {
      const qs = new URLSearchParams({ scope: currentScope, limit: "1000", offset: "0" });
      if (currentFrom) qs.set("from", currentFrom);
      if (currentTo) qs.set("to", currentTo);
      const res = await fetch(`/api/admin/reports/commissions?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const json = await res.json().catch(() => ({}));
      if (!res.ok || !json?.ok) {
        throw new Error(json?.error || "Gagal memuat komisi downline.");
      }
      setCommissions(Array.isArray(json.items) ? json.items : []);
      setPage(0);
    } catch (e) {
      setErr(e instanceof Error ? e.message : "Gagal memuat komisi downline.");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load(scope);
  }, [scope]);

  const summary = useMemo(() => {
    return {
      uplineCount: new Set(commissions.map((item) => `${item.scope}:${item.upline_member_id}`)).size,
      sourceCount: new Set(commissions.map((item) => `${item.scope}:${item.source_member_id}`)).size,
      totalTransactions: commissions.reduce((sum, item) => sum + Number(item.transaction_count || 0), 0),
      totalCommission: commissions.reduce((sum, item) => sum + Number(item.total_commission || 0), 0),
    };
  }, [commissions]);

  const totalPages = Math.max(1, Math.ceil(commissions.length / PAGE_SIZE));
  const safePage = Math.min(page, totalPages - 1);
  const pageRows = commissions.slice(safePage * PAGE_SIZE, safePage * PAGE_SIZE + PAGE_SIZE);

  const columns: DataTableColumn<CommissionRow>[] = [
    {
      id: "scope",
      header: "Jenis",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (item) => scopeLabel(item.scope),
    },
    {
      id: "upline",
      header: "Penerima Komisi",
      tdClassName: "min-w-[220px] text-slate-100",
      render: (item) => (
        <div>
          <div className="font-semibold">{item.upline_nama || item.upline_email}</div>
          <div className="text-xs text-slate-400">
            #{fmtID(item.upline_member_id)} • {item.upline_role} • {item.upline_email}
          </div>
        </div>
      ),
    },
    {
      id: "source",
      header: "Dari Downline",
      tdClassName: "min-w-[220px] text-slate-100",
      render: (item) => (
        <div>
          <div className="font-semibold">{item.source_nama || item.source_email}</div>
          <div className="text-xs text-slate-400">
            #{fmtID(item.source_member_id)} • {item.source_role} • {item.source_email}
          </div>
        </div>
      ),
    },
    {
      id: "level",
      header: "Level",
      tdClassName: "whitespace-nowrap text-cyan-200 font-semibold",
      render: (item) => (item.level || "-").toUpperCase(),
    },
    {
      id: "count",
      header: "Jml Trx",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (item) => fmtID(Number(item.transaction_count || 0)),
    },
    {
      id: "commission",
      header: "Total Komisi",
      tdClassName: "whitespace-nowrap text-emerald-300 font-semibold",
      render: (item) => fmtIDR(Number(item.total_commission || 0)),
    },
    {
      id: "last",
      header: "Terakhir",
      tdClassName: "whitespace-nowrap text-slate-400",
      render: (item) => (item.last_created_at ? new Date(item.last_created_at).toLocaleString("id-ID") : "-"),
    },
  ];

  function exportCommissionCsv() {
    const rows = [
      [
        "scope",
        "upline_member_id",
        "upline_email",
        "upline_nama",
        "upline_role",
        "level",
        "source_member_id",
        "source_email",
        "source_nama",
        "source_role",
        "transaction_count",
        "total_commission",
        "last_created_at",
      ],
      ...commissions.map((item) => [
        item.scope,
        String(item.upline_member_id),
        item.upline_email,
        item.upline_nama,
        item.upline_role,
        item.level,
        String(item.source_member_id),
        item.source_email,
        item.source_nama,
        item.source_role,
        String(item.transaction_count),
        String(item.total_commission),
        item.last_created_at || "",
      ]),
    ];
    downloadCsv(`komisi-downline-${scope}.csv`, rows);
  }

  return (
    <div className="space-y-4 p-2">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <div className="text-lg font-semibold tracking-tight">Komisi Downline</div>
          <div className="text-sm text-muted-foreground">
            Lihat komisi agent/master per downline tanpa mencampurkannya dengan laporan bisnis harian.
          </div>
        </div>

        <div className="flex flex-wrap gap-2">
          {(["all", "retail", "h2h"] as Scope[]).map((item) => (
            <Button
              key={item}
              variant={scope === item ? "primary" : "outline"}
              className="h-10"
              onClick={() => setScope(item)}
              disabled={loading}
            >
              {scopeLabel(item)}
            </Button>
          ))}
        </div>
      </div>

      <div className="flex flex-wrap items-end gap-2">
        <DateField label="Dari" className="h-10 w-44" value={from} onChange={(e) => setFrom(e.target.value)} />
        <DateField label="Sampai" className="h-10 w-44" value={to} onChange={(e) => setTo(e.target.value)} />
        <Button className="h-10" onClick={() => void load(scope, from, to)} disabled={loading}>
          {loading ? "Memuat..." : "Terapkan Filter"}
        </Button>
        <Button
          variant="outline"
          className="h-10"
          onClick={() => {
            setFrom("");
            setTo("");
            void load(scope, "", "");
          }}
          disabled={loading}
        >
          Reset
        </Button>
        <Button variant="outline" className="h-10" onClick={exportCommissionCsv} disabled={loading || commissions.length === 0}>
          Export Komisi CSV
        </Button>
      </div>

      {err ? (
        <div className="rounded-md border border-rose-400/30 bg-rose-500/10 px-4 py-3 text-sm text-rose-200">
          {err}
        </div>
      ) : null}

      <section className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        {[
          { label: "Penerima Komisi", value: fmtID(summary.uplineCount), icon: Coins },
          { label: "Downline Penghasil", value: fmtID(summary.sourceCount), icon: TrendingUp },
          { label: "Jumlah Trx", value: fmtID(summary.totalTransactions), icon: TrendingUp },
          { label: "Total Komisi", value: fmtIDR(summary.totalCommission), icon: Coins },
        ].map((item) => {
          const Icon = item.icon;
          return (
            <div
              key={item.label}
              className="rounded-md border border-white/15 bg-slate-950/50 p-4 shadow-[0_18px_42px_-26px_rgba(56,189,248,0.45)]"
            >
              <div className="flex items-center justify-between gap-3">
                <div>
                  <div className="text-xs font-semibold uppercase tracking-wide text-slate-400">{item.label}</div>
                  <div className="mt-2 text-lg font-semibold text-slate-100">{item.value}</div>
                </div>
                <span className="grid h-10 w-10 place-items-center rounded-2xl bg-cyan-500/10 text-cyan-200">
                  <Icon className="h-5 w-5" />
                </span>
              </div>
            </div>
          );
        })}
      </section>

      <section className="rounded-md border border-white/15 bg-slate-950/50 p-4 shadow-[0_18px_42px_-26px_rgba(56,189,248,0.45)]">
        <div className="mb-1 text-base font-semibold text-slate-100">Komisi Per Downline</div>
        <div className="text-sm text-slate-400">
          Menunjukkan berapa komisi yang didapat agent/master dari masing-masing user/member di bawahnya.
        </div>
        <DataTable
          columns={columns}
          rows={pageRows}
          rowKey={(row) => `${row.scope}-${row.upline_member_id}-${row.source_member_id}-${row.level}`}
          rowNumberStart={safePage * PAGE_SIZE + 1}
          minWidthClassName="min-w-[1280px]"
          emptyText="Belum ada data komisi."
          loading={loading}
          pagination={{
            page: safePage + 1,
            totalPages,
            onPrev: () => setPage((prev) => Math.max(0, prev - 1)),
            onNext: () => setPage((prev) => Math.min(totalPages - 1, prev + 1)),
            disablePrev: loading || safePage === 0,
            disableNext: loading || safePage >= totalPages - 1,
          }}
        />
      </section>
    </div>
  );
}

function escapeCsvCell(value: string) {
  if (/[",\n]/.test(value)) {
    return `"${value.replace(/"/g, "\"\"")}"`;
  }
  return value;
}

function downloadCsv(filename: string, rows: string[][]) {
  const csv = rows.map((row) => row.map((cell) => escapeCsvCell(cell)).join(",")).join("\n");
  const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
  const href = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = href;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(href);
}
