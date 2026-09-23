"use client";

import { useEffect, useMemo, useState } from "react";
import { Search, SlidersHorizontal } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";

type LogRow = {
  id: number;
  transaksi_member_id: number;
  ref_id: string;
  status_sebelum?: string;
  status_sesudah?: string;
  keterangan_sebelum?: string;
  keterangan_sesudah?: string;
  biaya_aktual_sebelum: number;
  biaya_aktual_sesudah: number;
  aksi: string;
  diubah_oleh?: number;
  diubah_oleh_nama?: string;
  dibuat_pada: string;
};

type ProfileResponse = {
  id?: number;
  nama?: string;
  email?: string;
};

type ActorOption = {
  id: number;
  nama: string;
  role: string;
};

type RawActorItem = {
  id?: number;
  nama?: string;
  role?: string;
};

type ApiResponse = {
  ok?: boolean;
  items?: LogRow[];
  total?: number;
  limit?: number;
  offset?: number;
  page?: number;
  total_pages?: number;
  error?: string;
};

type Props = {
  mode: "admin" | "operator";
  manualOnly?: boolean;
};

const DEFAULT_LIMIT = 10;

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

function fmtDate(s: string) {
  try {
    return new Date(s).toLocaleString("id-ID");
  } catch {
    return s;
  }
}

function startOfMonthISO() {
  const d = new Date();
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  return `${y}-${m}-01`;
}

export default function TrxMemberStatusLogsPage({ mode, manualOnly = false }: Props) {
  const isOperator = mode === "operator";
  const endpoint = manualOnly ? "/api/admin/history/transaksi/logs/manual" : "/api/admin/history/transaksi/logs";
  const [rows, setRows] = useState<LogRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [limit] = useState(DEFAULT_LIMIT);
  const [totalPages, setTotalPages] = useState(1);

  const [trxID, setTrxID] = useState("");
  const [refID, setRefID] = useState("");
  const [diubahOleh, setDiubahOleh] = useState("");
  const [from, setFrom] = useState(isOperator ? startOfMonthISO() : "");
  const [to, setTo] = useState("");
  const [mobileFilterOpen, setMobileFilterOpen] = useState(false);
  const [actorOptions, setActorOptions] = useState<ActorOption[]>([]);
  const activeFilterCount = [trxID, refID, diubahOleh, from, to].filter((v) => v.trim()).length;

  const page = Math.floor(offset / limit) + 1;

  async function hydrateOperatorActor(): Promise<string> {
    if (!isOperator) return "";
    try {
      const r = await fetch("/api/me/profile", {
        headers: authHeader(),
        cache: "no-store",
      });
      const j: ProfileResponse = await r.json().catch(() => ({}));
      if (r.ok && Number(j.id || 0) > 0) {
        const actor = String(j.id);
        setDiubahOleh(actor);
        return actor;
      }
    } catch {
      // noop
    }
    return "";
  }

  async function loadActorOptions() {
    if (isOperator) return;
    try {
      const roles = ["admin", "staff", "operator_trx", "operator_wallet"];
      const collected: ActorOption[] = [];

      for (const role of roles) {
        let nextOffset = 0;
        const limitPerPage = 100;

        for (let i = 0; i < 50; i += 1) {
          const qs = new URLSearchParams({
            role,
            limit: String(limitPerPage),
            offset: String(nextOffset),
          });
          const r = await fetch(`/api/admin/members?${qs.toString()}`, {
            headers: authHeader(),
            cache: "no-store",
          });
          const j = await r.json().catch(() => ({}));
          if (!r.ok) break;

          const items: RawActorItem[] = Array.isArray(j?.items) ? j.items : [];
          const mapped = items
            .filter((x): x is { id: number; nama?: string; role?: string } => Number(x?.id || 0) > 0)
            .map((x) => ({
              id: Number(x.id),
              nama: String(x.nama || `user-${x.id}`),
              role: String(x.role || role),
            }));
          collected.push(...mapped);

          if (items.length < limitPerPage) break;
          nextOffset += limitPerPage;
        }
      }

      const uniq = Array.from(new Map(collected.map((x) => [x.id, x])).values()).sort((a, b) => a.nama.localeCompare(b.nama, "id"));
      setActorOptions(uniq);
    } catch {
      setActorOptions([]);
    }
  }

  async function load(nextOffset = offset, overrideActor = diubahOleh) {
    setLoading(true);
    try {
      const qs = new URLSearchParams();
      qs.set("limit", String(limit));
      qs.set("offset", String(nextOffset));
      if (trxID.trim()) qs.set("trx_id", trxID.trim());
      if (refID.trim()) qs.set("ref_id", refID.trim());
      if (overrideActor.trim()) qs.set("diubah_oleh", overrideActor.trim());
      if (from.trim()) qs.set("from", from.trim());
      if (to.trim()) qs.set("to", to.trim());

      const r = await fetch(`${endpoint}?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j: ApiResponse = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        setRows([]);
        setTotal(0);
        setTotalPages(1);
        return;
      }
      setRows(Array.isArray(j.items) ? j.items : []);
      setTotal(Number(j.total || 0));
      setTotalPages(Math.max(1, Number(j.total_pages || 1)));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    let mounted = true;
    (async () => {
      let actor = diubahOleh;
      if (!isOperator) {
        await loadActorOptions();
      }
      if (isOperator && !actor.trim()) {
        try {
          const r = await fetch("/api/me/profile", {
            headers: authHeader(),
            cache: "no-store",
          });
          const j: ProfileResponse = await r.json().catch(() => ({}));
          if (mounted && r.ok && Number(j.id || 0) > 0) {
            actor = String(j.id);
            setDiubahOleh(actor);
          }
        } catch {
          // noop
        }
      }
      if (mounted) {
        await load(0, actor);
      }
    })();
    return () => {
      mounted = false;
    };
     
  }, []);

  const title = manualOnly
    ? (isOperator ? "Audit Transaksi Diubah Saya" : "Audit Transaksi Diubah")
    : (isOperator ? "Audit Log Status Saya" : "Audit Log Status Member");
  const description = manualOnly
    ? (isOperator
        ? "Riwayat perubahan status transaksi yang diubah manual oleh akun operator login. Default bulan berjalan."
        : "Riwayat perubahan status transaksi member yang diubah manual oleh admin/operator.")
    : (isOperator
        ? "Riwayat semua log perubahan status transaksi milik operator login. Default bulan berjalan."
        : "Riwayat semua perubahan status transaksi member untuk audit dan pelaporan.");

  const columns: DataTableColumn<LogRow>[] = useMemo(
    () => [
      {
        id: "dibuat",
        header: "Waktu",
        tdClassName: "whitespace-nowrap text-slate-200",
        render: (x) => fmtDate(x.dibuat_pada),
      },
      {
        id: "ref",
        header: "Ref ID",
        tdClassName: "whitespace-nowrap font-mono text-xs text-slate-300",
        render: (x) => x.ref_id,
      },
      {
        id: "trx",
        header: "Trx ID",
        tdClassName: "whitespace-nowrap text-slate-300",
        render: (x) => String(x.transaksi_member_id),
      },
      {
        id: "status",
        header: "Status",
        tdClassName: "whitespace-nowrap text-slate-200",
        render: (x) => `${x.status_sebelum || "-"} -> ${x.status_sesudah || "-"}`,
      },
      {
        id: "aksi",
        header: "Aksi",
        tdClassName: "whitespace-nowrap text-cyan-200",
        render: (x) => x.aksi,
      },
      {
        id: "actor",
        header: "Diubah Oleh",
        tdClassName: "whitespace-nowrap text-slate-200",
        render: (x) =>
          x.diubah_oleh_nama && x.diubah_oleh
            ? `${x.diubah_oleh_nama} (${x.diubah_oleh})`
            : x.diubah_oleh
              ? String(x.diubah_oleh)
              : "-",
      },
      {
        id: "ket",
        header: "Keterangan",
        tdClassName: "text-slate-300",
        render: (x) => x.keterangan_sesudah || x.keterangan_sebelum || "-",
      },
    ],
    []
  );

  return (
    <div className="space-y-4 p-2">
      <div>
        <div className="text-lg font-semibold tracking-tight">{title}</div>
        <div className="text-sm text-muted-foreground">{description}</div>
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
            <span className="rounded-full bg-cyan-500/20 px-2 py-0.5 text-xs font-semibold text-cyan-200">{activeFilterCount}</span>
          ) : null}
        </Button>
      </div>

      <div
        className={`${mobileFilterOpen ? "block" : "hidden"} rounded-2xl border border-white/12 bg-linear-to-br from-slate-900/85 via-slate-900/65 to-cyan-950/25 p-3 shadow-[0_22px_48px_-34px_rgba(56,189,248,0.75)] md:block`}
      >
        <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-6">
          <div className="relative">
            <Search className="pointer-events-none absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-slate-400" />
            <input
              value={refID}
              onChange={(e) => setRefID(e.target.value)}
              placeholder="Filter ref_id"
              className="h-10 w-full rounded-xl border border-white/15 bg-slate-950/55 pr-3 pl-9 text-sm text-slate-100 outline-none focus:border-cyan-400/50"
            />
          </div>

          <input
            value={trxID}
            onChange={(e) => setTrxID(e.target.value.replace(/\D/g, ""))}
            placeholder="trx_id"
            className="h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50"
          />

          {!isOperator ? (
            <select
              value={diubahOleh}
              onChange={(e) => setDiubahOleh(e.target.value)}
              className="h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50"
            >
              <option value="">semua admin/operator</option>
              {actorOptions.map((actor) => (
                <option key={actor.id} value={String(actor.id)}>
                  {actor.nama} ({actor.role})
                </option>
              ))}
            </select>
          ) : null}

          <input
            type="date"
            value={from}
            onChange={(e) => setFrom(e.target.value)}
            className="h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50 scheme-dark [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:opacity-90 [&::-webkit-calendar-picker-indicator]:invert"
          />

          <input
            type="date"
            value={to}
            onChange={(e) => setTo(e.target.value)}
            className="h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50 scheme-dark [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:opacity-90 [&::-webkit-calendar-picker-indicator]:invert"
          />

          <div className="grid grid-cols-2 gap-2">
            <Button
              type="button"
              variant="primary"
              className="h-10"
              disabled={loading}
              onClick={() => {
                const nextOffset = 0;
                setOffset(nextOffset);
                void load(nextOffset);
                setMobileFilterOpen(false);
              }}
            >
              {loading ? "Memuat..." : "Terapkan"}
            </Button>
            <Button
              type="button"
              variant="outline"
              disabled={loading}
              onClick={async () => {
                setTrxID("");
                setRefID("");
                setFrom(isOperator ? startOfMonthISO() : "");
                setTo("");
                const nextOffset = 0;
                setOffset(nextOffset);
                let nextActor = "";
                if (isOperator) {
                  nextActor = await hydrateOperatorActor();
                } else {
                  setDiubahOleh("");
                }
                await load(nextOffset, nextActor);
                setMobileFilterOpen(false);
              }}
            >
              Reset
            </Button>
          </div>
        </div>
      </div>

      <div className="text-sm text-slate-300">Total log: {total}</div>

      <DataTable
        columns={columns}
        rows={rows}
        rowKey={(x) => x.id}
        emptyText={loading ? "Memuat..." : "Tidak ada data."}
        minWidthClassName="min-w-260"
        showRowNumber
        rowNumberStart={offset + 1}
        pagination={{
          page,
          totalPages,
          onPrev: () => {
            const nextOffset = Math.max(0, offset - limit);
            setOffset(nextOffset);
            void load(nextOffset);
          },
          onNext: () => {
            const nextOffset = offset + limit;
            setOffset(nextOffset);
            void load(nextOffset);
          },
          disablePrev: loading || page <= 1,
          disableNext: loading || page >= totalPages,
        }}
      />
    </div>
  );
}
