"use client";

import { useEffect, useState } from "react";
import { Landmark } from "lucide-react";
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
  total_pending_withdraw: number;
  available_saldo: number;
};

type WithdrawRow = {
  id: number;
  ref_id: string;
  amount: number;
  bank_name: string;
  account_name: string;
  account_number: string;
  status: string;
  note?: string;
  reject_reason?: string;
  created_at?: string;
};

type Props = {
  authToken: string;
};

function fmtIDR(v: number) {
  return new Intl.NumberFormat("id-ID").format(Number(v || 0));
}

export function H2HWithdrawClient({ authToken }: Props) {
  const [summary, setSummary] = useState<Summary | null>(null);
  const [items, setItems] = useState<WithdrawRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [err, setErr] = useState("");
  const [ok, setOk] = useState("");
  const [amount, setAmount] = useState("");
  const [bankName, setBankName] = useState("");
  const [accountName, setAccountName] = useState("");
  const [accountNumber, setAccountNumber] = useState("");
  const [note, setNote] = useState("");
  const [dateFrom, setDateFrom] = useState(sevenDaysAgo());
  const [dateTo, setDateTo] = useState(todayDate());

  async function load(nextFrom = dateFrom, nextTo = dateTo) {
    setLoading(true);
    try {
      const listQs = new URLSearchParams({ limit: "50", offset: "0" });
      if (nextFrom) listQs.set("from", nextFrom);
      if (nextTo) listQs.set("to", nextTo);
      const [sumRes, listRes] = await Promise.all([
        fetch("/api/me/h2h/commissions/summary", { headers: { Authorization: `Bearer ${authToken}` }, cache: "no-store" }),
        fetch(`/api/me/h2h/withdraw-requests?${listQs.toString()}`, { headers: { Authorization: `Bearer ${authToken}` }, cache: "no-store" }),
      ]);
      const sumJson = await sumRes.json().catch(() => ({}));
      const listJson = await listRes.json().catch(() => ({}));
      if (!sumRes.ok || !sumJson?.ok) throw new Error(sumJson?.error || "Gagal memuat ringkasan.");
      if (!listRes.ok || !listJson?.ok) throw new Error(listJson?.error || "Gagal memuat withdraw.");
      setSummary(sumJson.item || null);
      setItems(Array.isArray(listJson.items) ? listJson.items : []);
    } catch (e) {
      setErr(e instanceof Error ? e.message : "Gagal memuat withdraw H2H.");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, [authToken]);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setErr("");
    setOk("");
    const amt = Number(amount || 0);
    if (!Number.isFinite(amt) || amt <= 0) {
      setErr("Nominal withdraw tidak valid.");
      return;
    }
    if (!bankName.trim() || !accountName.trim() || !accountNumber.trim()) {
      setErr("Data rekening wajib lengkap.");
      return;
    }
    setSaving(true);
    try {
      const r = await fetch("/api/me/h2h/withdraw-requests", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${authToken}`,
        },
        body: JSON.stringify({
          amount: Math.floor(amt),
          bank_name: bankName.trim(),
          account_name: accountName.trim(),
          account_number: accountNumber.trim(),
          note: note.trim(),
        }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) {
        setErr(j?.error || "Gagal membuat withdraw.");
        return;
      }
      setOk(`Withdraw berhasil diajukan. Ref ID: ${j.item?.ref_id || "-"}`);
      setAmount("");
      setBankName("");
      setAccountName("");
      setAccountNumber("");
      setNote("");
      await load();
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="space-y-4">
      <section className="rounded-2xl border border-white/10 bg-card/80 p-5 shadow-[0_16px_40px_-26px_rgba(0,0,0,0.8)]">
        <div className="flex items-center gap-3">
          <span className="grid h-11 w-11 place-items-center rounded-2xl border border-violet-400/20 bg-violet-400/10 text-violet-200">
            <Landmark className="h-5 w-5" />
          </span>
          <div>
            <h1 className="text-lg font-bold text-white">Withdraw Fee H2H</h1>
            <p className="text-sm text-white/60">Ajukan pencairan fee H2H ke rekening tujuan anda.</p>
          </div>
        </div>
      </section>

      <section className="grid gap-3 md:grid-cols-2">
        <div className="rounded-2xl border border-white/10 bg-card/80 p-4 shadow-[0_16px_40px_-26px_rgba(0,0,0,0.8)]">
          <div className="text-xs font-semibold uppercase tracking-wide text-white/45">Saldo Tersedia</div>
          <div className="mt-2 text-lg font-bold text-white">Rp {fmtIDR(summary?.available_saldo || 0)}</div>
        </div>
        <div className="rounded-2xl border border-white/10 bg-card/80 p-4 shadow-[0_16px_40px_-26px_rgba(0,0,0,0.8)]">
          <div className="text-xs font-semibold uppercase tracking-wide text-white/45">Pending Withdraw</div>
          <div className="mt-2 text-lg font-bold text-white">Rp {fmtIDR(summary?.total_pending_withdraw || 0)}</div>
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

      <section className="rounded-2xl border border-white/10 bg-card/80 p-5 shadow-[0_16px_40px_-26px_rgba(0,0,0,0.8)]">
        {err ? <div className="mb-3 rounded-2xl border border-rose-400/25 bg-rose-500/10 px-3 py-2 text-sm text-rose-300">{err}</div> : null}
        {ok ? <div className="mb-3 rounded-2xl border border-emerald-400/25 bg-emerald-500/10 px-3 py-2 text-sm text-emerald-300">{ok}</div> : null}

        <form className="grid gap-3" onSubmit={submit}>
          <input className="h-11 rounded-xl border border-white/10 bg-slate-950/60 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400" placeholder="Nominal withdraw" inputMode="numeric" value={amount} onChange={(e) => setAmount(e.target.value)} />
          <input className="h-11 rounded-xl border border-white/10 bg-slate-950/60 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400" placeholder="Nama bank" value={bankName} onChange={(e) => setBankName(e.target.value)} />
          <input className="h-11 rounded-xl border border-white/10 bg-slate-950/60 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400" placeholder="Nama pemilik rekening" value={accountName} onChange={(e) => setAccountName(e.target.value)} />
          <input className="h-11 rounded-xl border border-white/10 bg-slate-950/60 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400" placeholder="Nomor rekening / akun" value={accountNumber} onChange={(e) => setAccountNumber(e.target.value)} />
          <textarea className="min-h-[92px] rounded-xl border border-white/10 bg-slate-950/60 px-3 py-3 text-sm text-slate-100 outline-none focus:border-cyan-400" placeholder="Catatan opsional" value={note} onChange={(e) => setNote(e.target.value)} />
          <button className="h-11 rounded-xl bg-cyan-500 text-sm font-semibold text-slate-950 hover:bg-cyan-400 disabled:opacity-70" disabled={saving} type="submit">
            {saving ? "Mengajukan..." : "Ajukan Withdraw"}
          </button>
        </form>
      </section>

      <section className="rounded-2xl border border-white/10 bg-card/80 p-5 shadow-[0_16px_40px_-26px_rgba(0,0,0,0.8)]">
        <h2 className="font-bold text-white">Riwayat Withdraw</h2>
        <div className="mt-4 space-y-3">
          {loading ? <div className="text-sm text-white/60">Memuat withdraw...</div> : null}
          {!loading && items.length === 0 ? <div className="text-sm text-white/60">Belum ada request withdraw.</div> : null}
          {items.map((item) => (
            <div key={item.id} className="rounded-2xl border border-white/10 bg-slate-950/35 px-4 py-3">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <div className="text-sm font-semibold text-white">{item.ref_id}</div>
                  <div className="text-xs text-white/55">{item.bank_name} • {item.account_name} • {item.account_number}</div>
                  <div className="mt-1 text-xs text-white/50">{item.note || item.reject_reason || "-"}</div>
                </div>
                <div className="text-right">
                  <div className="text-sm font-semibold text-cyan-300">Rp {fmtIDR(item.amount)}</div>
                  <div className="mt-1 text-xs uppercase tracking-wide text-white/45">{item.status}</div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}
