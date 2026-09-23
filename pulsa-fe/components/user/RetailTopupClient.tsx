"use client";

import { useCallback, useEffect, useState } from "react";
import { CheckCircle2, ChevronRight, Copy, LoaderCircle, QrCode, RefreshCcw, ShieldCheck } from "lucide-react";

type DepositRow = {
  id: number;
  ref_id: string;
  amount: number;
  requested_amount?: number;
  metode: string;
  status: string;
  dibuat_pada: string;
};

type QrisDeposit = {
  ref_id: string;
  amount: number;
  fee_admin: number;
  gross_amount: number;
  status: string;
  qr_url: string;
  expired_at?: string;
};

type Props = { authToken: string; initialAmount?: number };

const MIN_TOPUP = 100_000;
const QUICK_AMOUNTS = [100_000, 200_000, 500_000, 1_000_000] as const;

function fmtIDR(value: number) {
  return new Intl.NumberFormat("id-ID").format(Number(value || 0));
}

function statusLabel(value: string) {
  switch (String(value || "").toLowerCase()) {
    case "pending": return "Menunggu pembayaran";
    case "approved": return "Saldo sudah masuk";
    case "rejected": return "Pembayaran kedaluwarsa";
    default: return value || "-";
  }
}

export function RetailTopupClient({ authToken, initialAmount = 0 }: Props) {
  const suggestedAmount = initialAmount > 0 ? Math.max(Math.ceil(initialAmount), MIN_TOPUP) : 0;
  const [amount, setAmount] = useState(suggestedAmount ? String(suggestedAmount) : "");
  const [payment, setPayment] = useState<QrisDeposit | null>(null);
  const [history, setHistory] = useState<DepositRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [checking, setChecking] = useState(false);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const nominal = Number(amount || 0);

  const request = useCallback(async (url: string, init?: RequestInit) => {
    const response = await fetch(url, {
      ...init,
      headers: {
        ...(init?.body ? { "Content-Type": "application/json" } : {}),
        Authorization: `Bearer ${authToken}`,
        ...(init?.headers || {}),
      },
      cache: "no-store",
    });
    const json = await response.json().catch(() => ({}));
    if (!response.ok || !json?.ok) throw new Error(json?.error || "Permintaan gagal diproses.");
    return json;
  }, [authToken]);

  const checkStatus = useCallback(async (refID: string, showLoading = false) => {
    if (showLoading) setChecking(true);
    try {
      const data = await request(`/api/me/deposit/request/qris/status?ref_id=${encodeURIComponent(refID)}&refresh=1`);
      const item = data.item as QrisDeposit;
      setPayment(item);
      if (item.status === "approved") {
        setMessage("Pembayaran berhasil. Saldo Isiloka sudah bertambah.");
      } else if (item.status === "rejected") {
        setError("QR pembayaran sudah kedaluwarsa atau ditolak. Buat QR baru untuk melanjutkan.");
      }
      return item;
    } finally {
      if (showLoading) setChecking(false);
    }
  }, [request]);

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const data = await request("/api/me/history/deposit?limit=25");
      const rows = Array.isArray(data.rows) ? data.rows as DepositRow[] : [];
      setHistory(rows);
      const active = rows.find((row) => String(row.metode).toLowerCase() === "qris" && String(row.status).toLowerCase() === "pending");
      if (active?.ref_id) await checkStatus(active.ref_id);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gagal memuat data topup.");
    } finally {
      setLoading(false);
    }
  }, [checkStatus, request]);

  useEffect(() => { void load(); }, [load]);

  useEffect(() => {
    if (!payment?.ref_id || payment.status !== "pending") return;
    const timer = window.setInterval(() => {
      void checkStatus(payment.ref_id).catch(() => undefined);
    }, 5000);
    return () => window.clearInterval(timer);
  }, [checkStatus, payment?.ref_id, payment?.status]);

  async function createPayment(event: React.FormEvent) {
    event.preventDefault();
    setError("");
    setMessage("");
    if (!Number.isFinite(nominal) || nominal < MIN_TOPUP) return setError("Minimum topup adalah Rp 100.000.");
    setSaving(true);
    try {
      const data = await request("/api/me/deposit/request/qris", {
        method: "POST",
        body: JSON.stringify({ amount: Math.floor(nominal) }),
      });
      setPayment(data.item as QrisDeposit);
      setMessage("QR pembayaran berhasil dibuat. Selesaikan pembayaran sebelum masa berlaku berakhir.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gagal membuat QR pembayaran.");
    } finally {
      setSaving(false);
    }
  }

  async function copy(value: string) {
    await navigator.clipboard.writeText(value);
    setMessage("Kode referensi berhasil disalin.");
  }

  const isPending = payment?.status === "pending";

  return (
    <div className="space-y-4">
      <section className="px-1">
        <p className="text-[10px] font-black uppercase tracking-[0.22em] text-[#087e8b]">Isi Saldo</p>
        <h1 className="mt-1 text-2xl font-black text-slate-950">Topup saldo</h1>
        <p className="mt-1 text-xs font-semibold text-slate-500">Pembayaran diproses melalui QRIS dan tercatat otomatis di Pulsa24Jam.</p>
      </section>

      {error ? <div className="rounded-2xl border border-rose-300 bg-rose-50 px-4 py-3 text-xs font-bold text-rose-700">{error}</div> : null}
      {message ? <div className="rounded-2xl border border-emerald-300 bg-emerald-50 px-4 py-3 text-xs font-bold text-emerald-800">{message}</div> : null}

      {loading ? (
        <div className="grid min-h-48 place-items-center rounded-[28px] bg-white"><LoaderCircle className="h-6 w-6 animate-spin text-[#087e8b]" /></div>
      ) : payment ? (
        <section className="rounded-[28px] border border-emerald-950/5 bg-white p-5 shadow-[0_18px_42px_rgba(6,78,59,0.11)]">
          <div className="flex items-start justify-between gap-3">
            <div>
              <p className="text-[10px] font-black uppercase tracking-[0.2em] text-[#087e8b]">Pembayaran QRIS</p>
              <h2 className="mt-1 text-lg font-black text-slate-950">{statusLabel(payment.status)}</h2>
            </div>
            <button type="button" onClick={() => void checkStatus(payment.ref_id, true)} disabled={checking} className="grid h-10 w-10 place-items-center rounded-2xl bg-emerald-50 text-[#087e8b] disabled:opacity-60" aria-label="Periksa status pembayaran">
              <RefreshCcw className={`h-4 w-4 ${checking ? "animate-spin" : ""}`} />
            </button>
          </div>

          {isPending && payment.qr_url ? (
            <div className="mt-5 rounded-2xl border border-emerald-100 bg-[#f8fffb] p-4 text-center">
              <img src={payment.qr_url} alt="QRIS topup saldo" className="mx-auto aspect-square w-full max-w-56 rounded-xl bg-white object-contain p-2" />
              <p className="mt-3 text-[11px] font-bold text-slate-600">Pindai dengan aplikasi bank atau dompet digital</p>
            </div>
          ) : payment.status === "approved" ? (
            <div className="mt-5 grid min-h-44 place-items-center rounded-2xl bg-emerald-50 text-center">
              <div><CheckCircle2 className="mx-auto h-12 w-12 text-[#087e8b]" /><p className="mt-3 text-sm font-black text-emerald-900">Topup berhasil</p></div>
            </div>
          ) : null}

          <div className="mt-4 space-y-3 rounded-2xl bg-slate-50 p-4">
            <div className="flex justify-between gap-4 text-xs"><span className="font-semibold text-slate-500">Saldo diterima</span><strong className="text-slate-950">Rp {fmtIDR(payment.amount)}</strong></div>
            {payment.fee_admin > 0 ? <div className="flex justify-between gap-4 text-xs"><span className="font-semibold text-slate-500">Biaya layanan</span><strong className="text-slate-950">Rp {fmtIDR(payment.fee_admin)}</strong></div> : null}
            <div className="flex justify-between gap-4 border-t border-slate-200 pt-3 text-sm"><span className="font-black text-slate-700">Total pembayaran</span><strong className="text-[#087e8b]">Rp {fmtIDR(payment.gross_amount || payment.amount)}</strong></div>
            <button type="button" onClick={() => void copy(payment.ref_id)} className="flex w-full items-center justify-between rounded-xl bg-white px-3 py-3 text-left"><span><span className="block text-[9px] font-bold uppercase text-slate-400">Referensi</span><span className="text-xs font-black text-slate-950">{payment.ref_id}</span></span><Copy className="h-4 w-4 text-[#087e8b]" /></button>
          </div>

          {!isPending ? <button type="button" onClick={() => { setPayment(null); setError(""); setMessage(""); void load(); }} className="mt-4 h-12 w-full rounded-2xl bg-[#087e8b] text-xs font-black text-white">Buat Topup Baru</button> : null}
        </section>
      ) : (
        <form className="space-y-4" onSubmit={createPayment}>
          <section className="rounded-[28px] border border-emerald-950/5 bg-white p-4 shadow-[0_18px_42px_rgba(6,78,59,0.11)]">
            <h2 className="text-base font-black text-slate-950">Pilih nominal</h2>
            <div className="mt-4 grid grid-cols-2 gap-2.5">{QUICK_AMOUNTS.map((value) => <button key={value} type="button" onClick={() => setAmount(String(value))} className={`h-12 rounded-2xl border text-xs font-black ${nominal === value ? "border-[#087e8b] bg-emerald-50 text-[#087e8b]" : "border-slate-200 bg-slate-50 text-slate-700"}`}>Rp {fmtIDR(value)}</button>)}</div>
            <label className="mt-4 flex h-12 items-center gap-3 rounded-2xl border border-slate-200 bg-[#f8fffb] px-4"><span className="text-xs font-black text-[#087e8b]">Rp</span><input className="min-w-0 flex-1 bg-transparent text-sm font-black text-slate-950 outline-none" inputMode="numeric" placeholder="Minimum 100.000" value={amount ? fmtIDR(nominal) : ""} onChange={(event) => setAmount(event.target.value.replace(/\D/g, ""))} /></label>
          </section>

          <section className="rounded-[28px] border border-[#087e8b] bg-emerald-50 p-4 shadow-[0_18px_42px_rgba(6,78,59,0.08)]">
            <div className="flex items-center gap-3">
              <span className="grid h-11 w-11 place-items-center rounded-2xl bg-white text-[#087e8b]"><QrCode className="h-5 w-5" /></span>
              <span className="min-w-0 flex-1"><span className="block text-sm font-black text-slate-950">QRIS</span><span className="text-[10px] font-semibold text-slate-600">Pembayaran otomatis melalui Pulsa24Jam</span></span>
              <ShieldCheck className="h-5 w-5 text-[#087e8b]" />
            </div>
          </section>

          <button type="submit" disabled={saving} className="inline-flex h-12 w-full items-center justify-center gap-2 rounded-2xl bg-[#087e8b] text-xs font-black text-white disabled:opacity-60">{saving ? <LoaderCircle className="h-4 w-4 animate-spin" /> : null}{saving ? "Membuat QR..." : "Lanjutkan Pembayaran"}<ChevronRight className="h-4 w-4" /></button>
        </form>
      )}

      {history.length ? <section className="rounded-[28px] bg-white p-4"><h2 className="text-sm font-black text-slate-950">Riwayat terbaru</h2><div className="mt-3 space-y-2">{history.slice(0, 5).map((row) => <div key={row.id} className="flex items-center justify-between rounded-2xl bg-slate-50 px-3 py-3"><div><p className="text-xs font-black text-slate-900">Rp {fmtIDR(row.requested_amount || row.amount)}</p><p className="text-[10px] font-semibold text-slate-500">QRIS</p></div><span className="text-[10px] font-black text-[#087e8b]">{statusLabel(row.status)}</span></div>)}</div></section> : null}
    </div>
  );
}
