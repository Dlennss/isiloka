"use client";

import { useEffect, useMemo, useState } from "react";
import { Coins } from "lucide-react";
import { DateField } from "@/components/ui/date-field";
import { Button } from "@/components/ui/button";

function todayDate() {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

function sevenDaysAgo() {
  const d = new Date();
  d.setDate(d.getDate() - 7);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

type Summary = {
  total_earned: number;
  total_pending_withdraw: number;
  total_approved_withdraw: number;
  total_rejected_withdraw: number;
  available_saldo: number;
};

type CommissionRow = {
  id: number;
  ref_id: string;
  kategori: string;
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

export function H2HCommissionClient({ authToken }: Props) {
  const [summary, setSummary] = useState<Summary | null>(null);
  const [items, setItems] = useState<CommissionRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState("");
  const [dateFrom, setDateFrom] = useState(sevenDaysAgo());
  const [dateTo, setDateTo] = useState(todayDate());

  async function load(nextFrom = dateFrom, nextTo = dateTo) {
    setLoading(true);
    setErr("");
    try {
      const listQs = new URLSearchParams({ limit: "50", offset: "0" });
      if (nextFrom) listQs.set("from", nextFrom);
      if (nextTo) listQs.set("to", nextTo);
      const [sumRes, listRes] = await Promise.all([
        fetch("/api/me/h2h/commissions/summary", { headers: { Authorization: `Bearer ${authToken}` }, cache: "no-store" }),
        fetch(`/api/me/h2h/commissions?${listQs.toString()}`, { headers: { Authorization: `Bearer ${authToken}` }, cache: "no-store" }),
      ]);
      const sumJson = await sumRes.json().catch(() => ({}));
      const listJson = await listRes.json().catch(() => ({}));
      if (!sumRes.ok || !sumJson?.ok) throw new Error(sumJson?.error || "Gagal memuat ringkasan fee H2H.");
      if (!listRes.ok || !listJson?.ok) throw new Error(listJson?.error || "Gagal memuat riwayat fee H2H.");
      setSummary(sumJson.item || null);
      setItems(Array.isArray(listJson.items) ? listJson.items : []);
    } catch (e) {
      setErr(e instanceof Error ? e.message : "Gagal memuat fee H2H.");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, [authToken]);

  const cards = useMemo(
    () => [
      { label: "Total Fee", value: summary?.total_earned || 0 },
      { label: "Saldo Tersedia", value: summary?.available_saldo || 0 },
      { label: "Withdraw Pending", value: summary?.total_pending_withdraw || 0 },
      { label: "Withdraw Approved", value: summary?.total_approved_withdraw || 0 },
    ],
    [summary],
  );

  return (
    <div className="space-y-4">
      <section className="rounded-2xl border border-white/10 bg-card/80 p-5 shadow-[0_16px_40px_-26px_rgba(0,0,0,0.8)]">
        <div className="flex items-center gap-3">
          <span className="grid h-11 w-11 place-items-center rounded-2xl border border-amber-400/20 bg-amber-400/10 text-amber-200">
            <Coins className="h-5 w-5" />
          </span>
          <div>
            <h1 className="text-lg font-bold text-white">Fee H2H</h1>
            <p className="text-sm text-white/60">Ringkasan fee downline H2H yang sudah masuk ke saldo anda.</p>
          </div>
        </div>
      </section>

      <section className="rounded-2xl border border-white/10 bg-card/80 p-4 shadow-[0_16px_40px_-26px_rgba(0,0,0,0.8)]">
        <div className="flex flex-wrap items-end gap-2">
          <DateField label="Dari" className="h-10 w-44" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)} />
          <DateField label="Sampai" className="h-10 w-44" value={dateTo} onChange={(e) => setDateTo(e.target.value)} />
          <Button className="h-10" onClick={() => void load()} disabled={loading}>
            {loading ? "Memuat..." : "Terapkan Filter"}
          </Button>
          <Button
            variant="outline"
            className="h-10"
            onClick={() => {
              const nextFrom = sevenDaysAgo();
              const nextTo = todayDate();
              setDateFrom(nextFrom);
              setDateTo(nextTo);
              void load(nextFrom, nextTo);
            }}
            disabled={loading}
          >
            Reset
          </Button>
        </div>
      </section>

      {err ? <div className="rounded-2xl border border-rose-400/25 bg-rose-500/10 px-4 py-3 text-sm text-rose-300">{err}</div> : null}

      <section className="grid gap-3 md:grid-cols-2">
        {cards.map((item) => (
          <div key={item.label} className="rounded-2xl border border-white/10 bg-card/80 p-4 shadow-[0_16px_40px_-26px_rgba(0,0,0,0.8)]">
            <div className="text-xs font-semibold uppercase tracking-wide text-white/45">{item.label}</div>
            <div className="mt-2 text-lg font-bold text-white">Rp {fmtIDR(item.value)}</div>
          </div>
        ))}
      </section>

      <section className="rounded-2xl border border-white/10 bg-card/80 p-5 shadow-[0_16px_40px_-26px_rgba(0,0,0,0.8)]">
        <h2 className="font-bold text-white">Riwayat Fee</h2>
        <div className="mt-4 space-y-3">
          {loading ? <div className="text-sm text-white/60">Memuat riwayat fee...</div> : null}
          {!loading && items.length === 0 ? <div className="text-sm text-white/60">Belum ada fee H2H yang masuk.</div> : null}
          {items.map((item) => (
            <div key={item.id} className="rounded-2xl border border-white/10 bg-slate-950/35 px-4 py-3">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <div className="text-sm font-semibold text-white">Rp {fmtIDR(item.amount)}</div>
                  <div className="text-xs text-white/55">
                    {String(item.level || "").toUpperCase()} • {item.kategori || "-"} • {item.ref_id}
                  </div>
                  <div className="mt-1 text-xs text-white/55">
                    Sumber: {item.source_member_nama || "-"}{item.source_member_role ? ` (${item.source_member_role})` : ""}
                  </div>
                  {item.note ? <div className="mt-1 text-xs text-white/50">{item.note}</div> : null}
                </div>
                <div className="text-xs text-white/40">{item.created_at ? new Date(item.created_at).toLocaleString("id-ID") : "-"}</div>
              </div>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}
