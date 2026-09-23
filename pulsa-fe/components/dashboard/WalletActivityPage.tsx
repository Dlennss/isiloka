"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { fmtID } from "@/lib/format";

type ActorOption = {
  id: number;
  nama: string;
  role: string;
};

type WalletActivitySummary = {
  member_adjust_count: number;
  member_credit_count: number;
  member_debit_count: number;
  provider_adjust_count: number;
  provider_credit_count: number;
  provider_debit_count: number;
};

type WalletActivityRow = {
  scope: "member" | "provider";
  target: string;
  ref_id: string;
  arah: string;
  jumlah: number;
  alasan: string;
  catatan: string;
  dibuat_pada: string;
};

type ApiItem = {
  from: string;
  to: string;
  summary: WalletActivitySummary;
  recent: WalletActivityRow[];
};

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

function formatDateTime(v: string): string {
  try {
    return new Date(v).toLocaleString("id-ID");
  } catch {
    return v;
  }
}

function startOfMonthISO() {
  const d = new Date();
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  return `${y}-${m}-01`;
}

function localDateISO(d: Date) {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

function todayISO() {
  return localDateISO(new Date());
}

type WalletActivityPageProps = {
  title?: string;
  description?: string;
  showHeader?: boolean;
};

const filterControlClassName =
  "h-11 w-full rounded-xl border border-cyan-200/20 bg-slate-950/75 px-3 text-sm font-semibold text-white shadow-inner shadow-black/20 outline-none transition focus:border-cyan-300/70 focus:ring-2 focus:ring-cyan-400/25 [color-scheme:dark]";

const optionClassName = "bg-slate-950 text-slate-50";

export default function WalletActivityPage({
  title = "Audit Aktivitas Wallet",
  description = "Laporan koreksi saldo member dan provider oleh admin/operator wallet.",
  showHeader = true,
}: WalletActivityPageProps) {
  const [loading, setLoading] = useState(false);
  const [actors, setActors] = useState<ActorOption[]>([]);
  const [actorID, setActorID] = useState("");
  const [from, setFrom] = useState(startOfMonthISO());
  const [to, setTo] = useState(todayISO());
  const [item, setItem] = useState<ApiItem | null>(null);

  async function loadActors() {
    const roles = ["admin", "staff", "operator_wallet"];
    const out: ActorOption[] = [];

    for (const role of roles) {
      let offset = 0;
      const limit = 100;
      for (let i = 0; i < 50; i += 1) {
        const qs = new URLSearchParams({ role, limit: String(limit), offset: String(offset) });
        const r = await fetch(`/api/admin/members?${qs.toString()}`, {
          headers: authHeader(),
          cache: "no-store",
        });
        const j = await r.json().catch(() => ({}));
        if (!r.ok) break;
        const rows = Array.isArray(j?.items) ? j.items : [];
        for (const row of rows) {
          const id = Number(row?.id || 0);
          if (id <= 0) continue;
          out.push({
            id,
            nama: String(row?.nama || `user-${id}`),
            role: String(row?.role || role),
          });
        }
        if (rows.length < limit) break;
        offset += limit;
      }
    }

    const uniq = Array.from(new Map(out.map((x) => [x.id, x])).values()).sort((a, b) => a.nama.localeCompare(b.nama, "id"));
    setActors(uniq);
    if (!actorID && uniq.length > 0) {
      setActorID(String(uniq[0].id));
    }
  }

  async function load(nextActorID = actorID) {
    if (!nextActorID) {
      setItem(null);
      return;
    }
    setLoading(true);
    try {
      const qs = new URLSearchParams();
      qs.set("actor_id", nextActorID);
      qs.set("from", from);
      qs.set("to", to);
      qs.set("limit", "20");
      const r = await fetch(`/api/admin/wallet/activity?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok || !j.item) {
        setItem(null);
        return;
      }
      setItem(j.item as ApiItem);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void loadActors();
     
  }, []);

  useEffect(() => {
    if (!actorID) return;
    void load(actorID);
     
  }, [actorID]);

  const actorName = actors.find((x) => String(x.id) === actorID);
  const summary = item?.summary;

  return (
    <div className="space-y-4 p-2">
      {showHeader ? (
        <div>
          <div className="text-lg font-semibold tracking-tight">{title}</div>
          <div className="mt-1 text-sm text-muted-foreground">{description}</div>
        </div>
      ) : null}

      <div className="rounded-2xl border border-white/10 bg-linear-to-br from-slate-900/85 via-slate-900/70 to-cyan-950/25 p-4 shadow-[0_16px_40px_-24px_rgba(34,211,238,0.28)]">
        <div className="grid gap-3 md:grid-cols-[minmax(0,1.4fr)_minmax(0,1fr)_minmax(0,1fr)_auto]">
          <div>
            <label className="mb-1.5 block text-xs font-semibold text-slate-100">Aktor</label>
            <select
              value={actorID}
              onChange={(e) => setActorID(e.target.value)}
              className={filterControlClassName}
            >
              <option value="" className={optionClassName}>
                Pilih aktor
              </option>
              {actors.map((actor) => (
                <option key={actor.id} value={actor.id} className={optionClassName}>
                  {actor.nama} ({actor.role})
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="mb-1.5 block text-xs font-semibold text-slate-100">Dari</label>
            <input
              type="date"
              value={from}
              onChange={(e) => setFrom(e.target.value)}
              className={filterControlClassName}
            />
          </div>
          <div>
            <label className="mb-1.5 block text-xs font-semibold text-slate-100">Sampai</label>
            <input
              type="date"
              value={to}
              onChange={(e) => setTo(e.target.value)}
              className={filterControlClassName}
            />
          </div>
          <div className="flex items-end">
            <Button
              type="button"
              variant="primary"
              className="h-11 w-full"
              onClick={() => void load()}
              disabled={loading || !actorID}
            >
              {loading ? "Memuat..." : "Terapkan"}
            </Button>
          </div>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <div className="rounded-2xl border border-emerald-400/25 bg-linear-to-r from-emerald-500/15 to-teal-500/10 px-4 py-4 text-sm text-emerald-100 shadow-[0_14px_30px_-20px_rgba(16,185,129,0.8)]">
          <div className="text-emerald-200/90">Koreksi Saldo Member</div>
          <div className="mt-1 text-2xl font-semibold text-emerald-50">{summary?.member_adjust_count ?? 0}</div>
          <div className="mt-2 text-xs text-emerald-100/80">
            Credit {summary?.member_credit_count ?? 0} • Debit {summary?.member_debit_count ?? 0}
          </div>
        </div>
        <div className="rounded-2xl border border-sky-400/25 bg-linear-to-r from-sky-500/15 to-cyan-500/10 px-4 py-4 text-sm text-sky-100 shadow-[0_14px_30px_-20px_rgba(14,165,233,0.8)]">
          <div className="text-sky-200/90">Koreksi Saldo Provider</div>
          <div className="mt-1 text-2xl font-semibold text-sky-50">{summary?.provider_adjust_count ?? 0}</div>
          <div className="mt-2 text-xs text-sky-100/80">
            Credit {summary?.provider_credit_count ?? 0} • Debit {summary?.provider_debit_count ?? 0}
          </div>
        </div>
        <div className="rounded-2xl border border-indigo-400/25 bg-linear-to-r from-indigo-500/15 to-violet-500/10 px-4 py-4 text-sm text-indigo-100 shadow-[0_14px_30px_-20px_rgba(99,102,241,0.75)]">
          <div className="text-indigo-200/90">Aktor Terpilih</div>
          <div className="mt-1 text-lg font-semibold text-indigo-50">{actorName ? `${actorName.nama} (${actorName.role})` : "-"}</div>
          <div className="mt-2 text-xs text-indigo-100/80">{item ? `${item.from} s/d ${item.to}` : "Pilih actor untuk melihat laporan."}</div>
        </div>
        <div className="rounded-2xl border border-white/10 bg-linear-to-br from-slate-900/85 via-slate-900/70 to-slate-800/40 px-4 py-4 text-sm text-slate-100 shadow-[0_16px_34px_-24px_rgba(15,23,42,0.95)]">
          <div className="text-slate-300">Total Aktivitas</div>
          <div className="mt-1 text-2xl font-semibold text-white">
            {(summary?.member_adjust_count ?? 0) + (summary?.provider_adjust_count ?? 0)}
          </div>
          <div className="mt-2 text-xs text-slate-400">Gabungan koreksi saldo member dan provider.</div>
        </div>
      </div>

      <div className="rounded-2xl border border-white/10 bg-linear-to-br from-slate-900/85 via-slate-900/70 to-indigo-950/25 p-4 shadow-[0_16px_40px_-24px_rgba(99,102,241,0.28)]">
        <div className="flex items-center justify-between gap-3">
          <div>
            <div className="text-lg font-semibold text-white">Aktivitas Wallet</div>
            <div className="mt-1 text-sm text-white/60">Detail aktivitas saldo member dan saldo provider dari actor terpilih.</div>
          </div>
          <div className="rounded-full border border-white/10 bg-white/[0.03] px-3 py-1 text-xs text-slate-300">
            {item?.recent?.length ?? 0} data
          </div>
        </div>

        {!item?.recent?.length ? (
          <div className="mt-5 rounded-md border border-white/15 bg-slate-950/50 px-4 py-8 text-center text-sm text-slate-400">
            Tidak ada aktivitas wallet untuk filter ini.
          </div>
        ) : (
          <div className="mt-5 overflow-auto rounded-md border border-white/15 bg-slate-950/50 shadow-[0_18px_42px_-26px_rgba(56,189,248,0.45)]">
            <table className="min-w-[1080px] w-full text-sm">
              <thead className="sticky top-0 z-10 bg-linear-to-r from-cyan-500/20 via-sky-500/10 to-indigo-500/20 backdrop-blur">
                <tr className="text-left">
                  <th className="px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-200">Waktu</th>
                  <th className="px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-200">Scope</th>
                  <th className="px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-200">Target</th>
                  <th className="px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-200">Arah</th>
                  <th className="px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-200">Jumlah</th>
                  <th className="px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-200">Alasan</th>
                  <th className="px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-200">Catatan</th>
                  <th className="px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-200">Ref</th>
                </tr>
              </thead>
              <tbody>
                {item.recent.map((row, idx) => (
                  <tr key={`${row.scope}-${row.ref_id}-${row.dibuat_pada}-${idx}`} className="border-t border-white/10 bg-white/1.5 transition hover:bg-cyan-400/[0.07]">
                    <td className="whitespace-nowrap px-4 py-3 text-slate-100">{formatDateTime(row.dibuat_pada)}</td>
                    <td className="whitespace-nowrap px-4 py-3">
                      <span className={`inline-flex rounded-full px-2.5 py-1 text-xs font-semibold ${row.scope === "member" ? "bg-emerald-500/15 text-emerald-300" : "bg-sky-500/15 text-sky-300"}`}>
                        {row.scope}
                      </span>
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-slate-200">{row.target}</td>
                    <td className="whitespace-nowrap px-4 py-3 text-slate-200">{row.arah}</td>
                    <td className="whitespace-nowrap px-4 py-3 font-semibold text-cyan-200">Rp {fmtID(row.jumlah)}</td>
                    <td className="whitespace-nowrap px-4 py-3 text-slate-200">{row.alasan}</td>
                    <td className="px-4 py-3 text-slate-300">{row.catatan || "-"}</td>
                    <td className="whitespace-nowrap px-4 py-3 font-mono text-xs text-slate-300">{row.ref_id || "-"}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
