"use client";

import { useEffect, useState } from "react";
import { Coins } from "lucide-react";

type Summary = {
  total_earned: number;
  total_pending_withdraw: number;
  total_approved_withdraw: number;
  total_rejected_withdraw: number;
  available_saldo: number;
};

type CommissionRow = {
  id: number;
  invoice_id: string;
  level: string;
  amount: number;
  note: string;
  source_member_nama?: string | null;
  source_member_role?: string | null;
  created_at?: string | null;
};

type Props = {
  authToken: string;
};

function fmtIDR(v: number) {
  return new Intl.NumberFormat("id-ID").format(Number(v || 0));
}

export function RetailCommissionClient({ authToken }: Props) {
  const [summary, setSummary] = useState<Summary | null>(null);
  const [items, setItems] = useState<CommissionRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState("");

  useEffect(() => {
    let cancelled = false;
    async function load() {
      setLoading(true);
      setErr("");
      try {
        const [sumRes, listRes] = await Promise.all([
          fetch("/api/me/retail/commissions/summary", { headers: { Authorization: `Bearer ${authToken}` }, cache: "no-store" }),
          fetch("/api/me/retail/commissions?limit=50&offset=0", { headers: { Authorization: `Bearer ${authToken}` }, cache: "no-store" }),
        ]);
        const sumJson = await sumRes.json().catch(() => ({}));
        const listJson = await listRes.json().catch(() => ({}));
        if (!sumRes.ok || !sumJson?.ok) {
          throw new Error(sumJson?.error || "Gagal memuat ringkasan komisi.");
        }
        if (!listRes.ok || !listJson?.ok) {
          throw new Error(listJson?.error || "Gagal memuat riwayat komisi.");
        }
        if (!cancelled) {
          setSummary(sumJson.item || null);
          setItems(Array.isArray(listJson.items) ? listJson.items : []);
        }
      } catch (e) {
        if (!cancelled) setErr(e instanceof Error ? e.message : "Gagal memuat komisi retail.");
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, [authToken]);

  return (
    <div className="space-y-4">
      <section className="relative overflow-hidden rounded-md border border-sky-100 bg-[linear-gradient(135deg,#f8fbff_0%,#eef7ff_45%,#fff7db_100%)] p-5 shadow-[0_10px_28px_rgba(15,23,42,0.1)]">
        <div className="pointer-events-none absolute inset-y-0 right-0 w-28 bg-[radial-gradient(circle_at_top_right,rgba(59,130,246,0.18),transparent_58%)]" />
        <div className="pointer-events-none absolute inset-x-0 bottom-0 h-16 bg-[linear-gradient(90deg,rgba(14,165,233,0.06),rgba(245,158,11,0.12),rgba(59,130,246,0.08))]" />
        <div className="relative flex items-center gap-3">
          <span className="grid h-11 w-11 place-items-center rounded-md bg-[linear-gradient(135deg,#f59e0b,#f97316)] text-white shadow-[0_10px_20px_rgba(249,115,22,0.24)]">
            <Coins className="h-5 w-5" />
          </span>
          <div>
            <div className="text-[11px] font-semibold uppercase tracking-[0.24em] text-sky-700">Komisi</div>
            <h1 className="text-xl font-bold text-slate-900">Fee Retail</h1>
            <p className="text-sm text-slate-600">Komisi retail yang sudah masuk ke saldo akun anda.</p>
          </div>
        </div>
      </section>

      {err ? <div className="rounded-md border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">{err}</div> : null}

      <section className="grid grid-cols-2 gap-3">
        {[
          {
            label: "Total Fee",
            value: summary?.total_earned || 0,
            cardClass:
              "border-sky-200 bg-[linear-gradient(180deg,#ffffff_0%,#eef7ff_100%)] before:bg-sky-500",
            valueClass: "text-sky-700",
          },
          {
            label: "Saldo Tersedia",
            value: summary?.available_saldo || 0,
            cardClass:
              "border-emerald-200 bg-[linear-gradient(180deg,#ffffff_0%,#effcf6_100%)] before:bg-emerald-500",
            valueClass: "text-emerald-700",
          },
          {
            label: "Withdraw Pending",
            value: summary?.total_pending_withdraw || 0,
            cardClass:
              "border-amber-200 bg-[linear-gradient(180deg,#ffffff_0%,#fff8eb_100%)] before:bg-amber-500",
            valueClass: "text-amber-700",
          },
          {
            label: "Withdraw Approved",
            value: summary?.total_approved_withdraw || 0,
            cardClass:
              "border-indigo-200 bg-[linear-gradient(180deg,#ffffff_0%,#eef2ff_100%)] before:bg-indigo-500",
            valueClass: "text-indigo-700",
          },
        ].map((item) => (
          <div
            key={item.label}
            className={`relative overflow-hidden rounded-md border p-4 shadow-[0_10px_20px_rgba(15,23,42,0.06)] before:absolute before:left-0 before:top-0 before:h-full before:w-1 ${item.cardClass}`}
          >
            <div className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-500">{item.label}</div>
            <div className={`mt-2 text-lg font-bold ${item.valueClass}`}>Rp {fmtIDR(item.value)}</div>
          </div>
        ))}
      </section>

      <section className="overflow-hidden rounded-md border border-slate-200 bg-white shadow-[0_10px_24px_rgba(15,23,42,0.08)]">
        <div className="border-b border-slate-200 bg-[linear-gradient(90deg,#f8fafc,#eef6ff)] px-5 py-4">
          <h2 className="font-bold text-slate-900">Riwayat Fee</h2>
          <p className="mt-1 text-xs text-slate-500">Daftar komisi retail terbaru yang masuk ke akun anda.</p>
        </div>
        <div className="p-5">
        <div className="mt-4 space-y-3">
          {loading ? <div className="text-sm text-slate-500">Memuat riwayat fee...</div> : null}
          {!loading && items.length === 0 ? <div className="text-sm text-slate-500">Belum ada komisi retail yang masuk.</div> : null}
          {items.map((item) => (
            <div key={item.id} className="rounded-md border border-slate-200 bg-[linear-gradient(180deg,#ffffff_0%,#f8fafc_100%)] px-4 py-3 shadow-[0_4px_14px_rgba(15,23,42,0.04)]">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <div className="text-sm font-semibold text-slate-900">Rp {fmtIDR(item.amount)}</div>
                  <div className="mt-1 inline-flex rounded-sm bg-sky-50 px-2 py-1 text-[11px] font-semibold uppercase tracking-[0.14em] text-sky-700">
                    {item.level?.toUpperCase()} • {item.invoice_id}
                  </div>
                  <div className="mt-1 text-xs text-slate-500">
                    Sumber: {item.source_member_nama || "-"}{item.source_member_role ? ` (${item.source_member_role})` : ""}
                  </div>
                  {item.note ? <div className="mt-1 text-xs text-slate-500">{item.note}</div> : null}
                </div>
                <div className="text-xs text-slate-400">{item.created_at ? new Date(item.created_at).toLocaleString("id-ID") : "-"}</div>
              </div>
            </div>
          ))}
        </div>
        </div>
      </section>
    </div>
  );
}
