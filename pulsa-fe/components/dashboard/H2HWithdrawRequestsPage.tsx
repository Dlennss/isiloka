"use client";

import { useEffect, useState } from "react";
import { alertError, alertSuccess } from "@/components/ui/alerts";
import WithdrawApproveModal from "@/components/dashboard/WithdrawApproveModal";
import { DateField } from "@/components/ui/date-field";

function todayDate() {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

function sevenDaysAgo() {
  const d = new Date();
  d.setDate(d.getDate() - 7);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

type WithdrawRow = {
  id: number;
  ref_id: string;
  member_nama?: string | null;
  member_email?: string | null;
  amount: number;
  bank_name: string;
  account_name: string;
  account_number: string;
  status: string;
  note?: string;
  reject_reason?: string;
  created_at?: string;
};

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

function fmtIDR(v: number) {
  return new Intl.NumberFormat("id-ID").format(Number(v || 0));
}

export default function H2HWithdrawRequestsPage() {
  const [items, setItems] = useState<WithdrawRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [status, setStatus] = useState("pending");
  const [q, setQ] = useState("");
  const [approveTarget, setApproveTarget] = useState<WithdrawRow | null>(null);
  const [approving, setApproving] = useState(false);
  const [dateFrom, setDateFrom] = useState(sevenDaysAgo());
  const [dateTo, setDateTo] = useState(todayDate());

  async function load() {
    setLoading(true);
    const qs = new URLSearchParams({ status, q, limit: "50", offset: "0" });
    if (dateFrom) qs.set("from", dateFrom);
    if (dateTo) qs.set("to", dateTo);
    const r = await fetch(`/api/admin/h2h/withdraw-requests?${qs.toString()}`, {
      headers: authHeader(),
      cache: "no-store",
    });
    const j = await r.json().catch(() => ({}));
    setItems(Array.isArray(j?.items) ? j.items : []);
    setLoading(false);
  }

  useEffect(() => {
    void load();
  }, [status]);

  async function approve(payload: { bank_id: number; fee: number; note: string }) {
    if (!approveTarget) return;
    setApproving(true);
    try {
      const r = await fetch(`/api/admin/h2h/withdraw-requests/approve?id=${approveTarget.id}`, {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify(payload),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) {
        await alertError(j?.error || "Gagal approve withdraw.");
        return;
      }
      await alertSuccess("Withdraw H2H berhasil di-approve.");
      setApproveTarget(null);
      await load();
    } catch {
      await alertError("Gagal approve withdraw.");
    } finally {
      setApproving(false);
    }
  }

  async function reject(id: number) {
    const reason = window.prompt("Masukkan alasan reject withdraw H2H:");
    if (!reason || !reason.trim()) return;
    const r = await fetch(`/api/admin/h2h/withdraw-requests/reject?id=${id}`, {
      method: "POST",
      headers: { "Content-Type": "application/json", ...authHeader() },
      body: JSON.stringify({ reason: reason.trim() }),
    });
    const j = await r.json().catch(() => ({}));
    if (!r.ok || !j?.ok) {
      await alertError(j?.error || "Gagal reject withdraw.");
      return;
    }
    await alertSuccess("Withdraw H2H berhasil ditolak dan saldo dikembalikan.");
    await load();
  }

  return (
    <div className="space-y-4 p-2">
      <div className="rounded-3xl border border-white/10 bg-[linear-gradient(135deg,rgba(15,23,42,.95),rgba(30,41,59,.82))] px-5 py-5 shadow-[0_18px_60px_rgba(2,6,23,.28)]">
        <h1 className="text-2xl font-semibold tracking-tight text-white">H2H Withdraw</h1>
        <p className="mt-2 text-sm leading-6 text-slate-300">Pantau dan proses request pencairan fee H2H dari agent member atau master member.</p>
      </div>

      <div className="rounded-3xl border border-white/10 bg-slate-950/70 p-4">
        <div className="flex flex-col gap-3 md:flex-row">
          <select className="h-11 rounded-xl border border-white/10 bg-slate-900 px-3 text-sm text-white outline-none" value={status} onChange={(e) => setStatus(e.target.value)}>
            <option value="pending">Pending</option>
            <option value="approved">Approved</option>
            <option value="rejected">Rejected</option>
            <option value="all">Semua</option>
          </select>
          <input className="h-11 flex-1 rounded-xl border border-white/10 bg-slate-900 px-3 text-sm text-white outline-none" placeholder="Cari ref, nama, email, bank" value={q} onChange={(e) => setQ(e.target.value)} />
          <DateField label="Dari" className="h-11 w-44" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)} />
          <DateField label="Sampai" className="h-11 w-44" value={dateTo} onChange={(e) => setDateTo(e.target.value)} />
          <button className="h-11 rounded-xl bg-cyan-500 px-4 text-sm font-semibold text-slate-950" onClick={() => void load()}>Refresh</button>
        </div>
      </div>

      <div className="space-y-3">
        {loading ? <div className="rounded-2xl bg-white p-4 text-sm text-slate-500">Memuat withdraw H2H...</div> : null}
        {!loading && items.length === 0 ? <div className="rounded-2xl bg-white p-4 text-sm text-slate-500">Tidak ada request withdraw untuk filter ini.</div> : null}
        {items.map((item) => (
          <div key={item.id} className="rounded-3xl bg-white p-5 shadow-[0_8px_22px_rgba(15,23,42,0.13)]">
            <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
              <div className="min-w-0">
                <div className="text-sm font-semibold text-slate-900">{item.ref_id}</div>
                <div className="text-xs text-slate-500">{item.member_nama || "-"} • {item.member_email || "-"}</div>
                <div className="mt-1 text-xs text-slate-500">{item.bank_name} • {item.account_name} • {item.account_number}</div>
                <div className="mt-1 text-xs text-slate-500">{item.note || item.reject_reason || "-"}</div>
              </div>
              <div className="md:text-right">
                <div className="text-base font-bold text-sky-700">Rp {fmtIDR(item.amount)}</div>
                <div className="mt-1 text-xs uppercase tracking-wide text-slate-400">{item.status}</div>
                <div className="mt-1 text-xs text-slate-400">{item.created_at ? new Date(item.created_at).toLocaleString("id-ID") : "-"}</div>
              </div>
            </div>
            {item.status === "pending" ? (
              <div className="mt-4 flex flex-wrap gap-2">
                <button className="h-10 rounded-xl bg-emerald-600 px-4 text-sm font-semibold text-white" onClick={() => setApproveTarget(item)}>Approve</button>
                <button className="h-10 rounded-xl bg-rose-600 px-4 text-sm font-semibold text-white" onClick={() => void reject(item.id)}>Reject</button>
              </div>
            ) : null}
          </div>
        ))}
      </div>

      <WithdrawApproveModal
        open={Boolean(approveTarget)}
        title="Approve Withdraw H2H"
        amount={approveTarget?.amount || 0}
        sourceAccountNumber="8761518267"
        sourceAccountLabel="BCA H2H LISA OKTARIA 8761518267"
        submitting={approving}
        onClose={() => {
          if (!approving) setApproveTarget(null);
        }}
        onSubmit={approve}
      />
    </div>
  );
}
