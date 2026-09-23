"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Check, ChevronLeft, ChevronRight, Copy, Filter, RotateCcw, Search, X } from "lucide-react";
import Swal from "sweetalert2";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { alertConfirm, alertError, alertSuccess } from "@/components/ui/alerts";
import { fmtID } from "@/lib/format";

type DepositVAReq = {
  id: number;
  member_id: number;
  ref_id?: string;
  member_nama?: string;
  bank_id?: number | null;
  bank_nama?: string;
  bank_nomor_rekening?: string;
  bank_atas_nama?: string;
  amount: number;
  requested_amount?: number;
  unique_code?: number;
  approved_amount?: number;
  metode: string;
  bukti_url?: string;
  status: "ticket" | "pending" | "approved" | "rejected" | "cancelled" | string;
  note?: string;
  dibuat_pada: string;
  diproses_pada?: string | null;
  diproses_oleh?: number | null;
  diproses_nama?: string | null;
};

type MemberOption = {
  id: number;
  nama?: string;
};

type DepositVARequestsPageProps = {
  readOnly?: boolean;
  subtitle?: string;
};

const PAGE_SIZE = 10;
const POLL_MS = 10_000;

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  if (!t) return {};
  return { Authorization: `Bearer ${t}` };
}

function toLocalDateInputValue(date: Date) {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
}

function depositTime(row: DepositVAReq) {
  const parsed = new Date(row.dibuat_pada).getTime();
  return Number.isFinite(parsed) ? parsed : 0;
}

function sortRows(rows: DepositVAReq[]) {
  return [...rows].sort((a, b) => depositTime(a) - depositTime(b) || a.id - b.id);
}

function sameRow(a: DepositVAReq, b: DepositVAReq) {
  return (
    a.id === b.id &&
    a.member_id === b.member_id &&
    (a.ref_id || "") === (b.ref_id || "") &&
    (a.member_nama || "") === (b.member_nama || "") &&
    (a.bank_nama || "") === (b.bank_nama || "") &&
    (a.bank_nomor_rekening || "") === (b.bank_nomor_rekening || "") &&
    (a.bank_atas_nama || "") === (b.bank_atas_nama || "") &&
    Number(a.amount || 0) === Number(b.amount || 0) &&
    Number(a.requested_amount || 0) === Number(b.requested_amount || 0) &&
    Number(a.unique_code || 0) === Number(b.unique_code || 0) &&
    Number(a.approved_amount || 0) === Number(b.approved_amount || 0) &&
    a.status === b.status &&
    (a.note || "") === (b.note || "") &&
    a.dibuat_pada === b.dibuat_pada &&
    (a.diproses_pada || "") === (b.diproses_pada || "") &&
    (a.diproses_nama || "") === (b.diproses_nama || "")
  );
}

function sameRows(a: DepositVAReq[], b: DepositVAReq[]) {
  return a.length === b.length && a.every((row, index) => sameRow(row, b[index]));
}

function memberLabel(row: DepositVAReq) {
  return row.member_nama ? `${row.member_nama} (#${row.member_id})` : `Member #${row.member_id}`;
}

function formatDate(value?: string | null) {
  if (!value) return "-";
  const date = new Date(value);
  if (!Number.isFinite(date.getTime())) return "-";
  return date.toLocaleString("id-ID");
}

function statusClass(status: string) {
  const s = String(status || "").toLowerCase();
  if (s === "approved") return "border-emerald-400/30 bg-emerald-400/10 text-emerald-200";
  if (s === "rejected" || s === "cancelled") return "border-red-400/30 bg-red-400/10 text-red-200";
  if (s === "pending") return "border-amber-400/30 bg-amber-400/10 text-amber-200";
  return "border-cyan-400/30 bg-cyan-400/10 text-cyan-200";
}

function canProcess(row: DepositVAReq) {
  const status = String(row.status || "").toLowerCase();
  return status === "ticket" || status === "pending";
}

function requestedAmount(row: DepositVAReq) {
  return Number(row.requested_amount || row.amount || 0);
}

function amountDigits(value: string | number) {
  return String(value || "").replace(/\D/g, "");
}

function formatAmountInput(value: string | number) {
  const digits = amountDigits(value);
  return digits ? Number(digits).toLocaleString("id-ID") : "";
}

export default function DepositVARequestsPage({ readOnly = false, subtitle = "Tiket VA LoketBayar untuk operator wallet." }: DepositVARequestsPageProps) {
  const now = useMemo(() => new Date(), []);
  const defaultFrom = useMemo(() => toLocalDateInputValue(new Date(now.getFullYear(), now.getMonth(), 1)), [now]);
  const defaultTo = useMemo(() => toLocalDateInputValue(now), [now]);

  const [items, setItems] = useState<DepositVAReq[]>([]);
  const [members, setMembers] = useState<MemberOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [showFiltersMobile, setShowFiltersMobile] = useState(false);
  const [offset, setOffset] = useState(0);
  const [hasNext, setHasNext] = useState(false);
  const [status, setStatus] = useState("ticket");
  const [ticket, setTicket] = useState("");
  const [memberID, setMemberID] = useState("");
  const [from, setFrom] = useState(defaultFrom);
  const [to, setTo] = useState(defaultTo);
  const [appliedStatus, setAppliedStatus] = useState("ticket");
  const [appliedTicket, setAppliedTicket] = useState("");
  const [appliedMemberID, setAppliedMemberID] = useState("");
  const [appliedFrom, setAppliedFrom] = useState(defaultFrom);
  const [appliedTo, setAppliedTo] = useState(defaultTo);
  const [note, setNote] = useState("");
  const [copiedID, setCopiedID] = useState<number | null>(null);

  const pollInFlightRef = useRef(false);
  const hasLoadedRef = useRef(false);
  const intervalRef = useRef<number | null>(null);
  const currentPage = Math.floor(offset / PAGE_SIZE) + 1;
  const totalPages = hasNext ? currentPage + 1 : currentPage;

  const copyText = useCallback(async (value: string) => {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(value);
      return;
    }

    const input = document.createElement("textarea");
    input.value = value;
    input.setAttribute("readonly", "true");
    input.style.position = "fixed";
    input.style.left = "-9999px";
    document.body.appendChild(input);
    input.select();

    try {
      const ok = document.execCommand("copy");
      if (!ok) throw new Error("copy failed");
    } finally {
      document.body.removeChild(input);
    }
  }, []);

  const copyAmount = useCallback(
    async (row: DepositVAReq) => {
      try {
        await copyText(String(Number(row.amount || 0)));
        setCopiedID(row.id);
        window.setTimeout(() => {
          setCopiedID((current) => (current === row.id ? null : current));
        }, 1400);
      } catch {
        await alertError("Gagal copy nominal.");
      }
    },
    [copyText],
  );

  const load = useCallback(
    async (
      nextOffset = offset,
      nextStatus = appliedStatus,
      nextTicket = appliedTicket,
      nextMemberID = appliedMemberID,
      nextFrom = appliedFrom,
      nextTo = appliedTo,
      options: { showLoading?: boolean } = {},
    ) => {
      if (pollInFlightRef.current) return;
      pollInFlightRef.current = true;
      const showLoading = options.showLoading ?? !hasLoadedRef.current;
      if (showLoading) setLoading(true);
      else setRefreshing(true);

      try {
        const qs = new URLSearchParams();
        if (nextStatus !== "all") qs.set("status", nextStatus);
        if (nextTicket.trim()) qs.set("ref_id", nextTicket.trim().replace(/^#/, ""));
        if (nextMemberID.trim()) qs.set("member_id", nextMemberID.trim());
        if (nextFrom) qs.set("from", nextFrom);
        if (nextTo) qs.set("to", nextTo);
        qs.set("limit", String(PAGE_SIZE + 1));
        qs.set("offset", String(nextOffset));
        qs.set("order", "asc");

        const r = await fetch(`/api/admin/deposit/va/requests?${qs.toString()}`, {
          headers: authHeader(),
          cache: "no-store",
        });
        const j = await r.json().catch(() => ({}));
        if (!r.ok || !j.ok) throw new Error(j.error || "Gagal memuat data VA");

        const all: DepositVAReq[] = Array.isArray(j.items) ? j.items : [];
        const nextRows = sortRows(all.slice(0, PAGE_SIZE));
        setHasNext(all.length > PAGE_SIZE);
        setItems((prev) => (sameRows(prev, nextRows) ? prev : nextRows));
      } catch (err) {
        console.error(err);
        setItems([]);
        setHasNext(false);
      } finally {
        hasLoadedRef.current = true;
        if (showLoading) setLoading(false);
        else setRefreshing(false);
        pollInFlightRef.current = false;
      }
    },
    [offset, appliedStatus, appliedTicket, appliedMemberID, appliedFrom, appliedTo],
  );

  useEffect(() => {
    void load(offset, appliedStatus, appliedTicket, appliedMemberID, appliedFrom, appliedTo);
  }, [load, offset, appliedStatus, appliedTicket, appliedMemberID, appliedFrom, appliedTo]);

  useEffect(() => {
    if (intervalRef.current) {
      window.clearInterval(intervalRef.current);
    }

    intervalRef.current = window.setInterval(() => {
      if (document.visibilityState !== "visible") return;
      void load(offset, appliedStatus, appliedTicket, appliedMemberID, appliedFrom, appliedTo, { showLoading: false });
    }, POLL_MS);

    return () => {
      if (intervalRef.current) {
        window.clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [load, offset, appliedStatus, appliedTicket, appliedMemberID, appliedFrom, appliedTo]);

  useEffect(() => {
    let cancelled = false;

    async function loadMembers() {
      try {
        const qs = new URLSearchParams();
        qs.set("limit", "500");
        qs.set("role", "member");
        const r = await fetch(`/api/admin/members?${qs.toString()}`, {
          headers: authHeader(),
          cache: "no-store",
        });
        const j = await r.json().catch(() => ({}));
        const rows = Array.isArray(j?.items) ? j.items : [];
        const next = rows
          .filter((x: MemberOption) => Number(x?.id || 0) > 0)
          .map((x: MemberOption) => ({
            id: Number(x.id),
            nama: String(x.nama || `Member #${x.id}`),
          }));
        if (!cancelled) setMembers(next);
      } catch {
        if (!cancelled) setMembers([]);
      }
    }

    void loadMembers();
    return () => {
      cancelled = true;
    };
  }, []);

  async function approve(row: DepositVAReq) {
    if (!canProcess(row)) return;

    const result = await Swal.fire<{ approvedAmount: number }>({
      background: "#0b1220",
      color: "#e5e7eb",
      confirmButtonColor: "#16a34a",
      cancelButtonColor: "#6b7280",
      customClass: {
        popup: "rounded-2xl border border-white/10 shadow-2xl",
        title: "text-base font-semibold",
        htmlContainer: "text-sm text-slate-300",
        confirmButton: "rounded-lg px-4 py-2 font-semibold",
        cancelButton: "rounded-lg px-4 py-2 font-semibold",
      },
      icon: "question",
      title: "Approve Deposit VA",
      html: `
        <div class="space-y-3 text-left">
          <div>Tiket: <b>${row.ref_id || row.id}</b></div>
          <div>Nominal tiket: <b>Rp ${fmtID(row.amount)}</b></div>
          <div>Nominal diminta: <b>Rp ${fmtID(requestedAmount(row))}</b></div>
          <label for="va-approved-amount" class="block text-xs font-semibold uppercase tracking-wide text-slate-400">Nominal transfer aktual</label>
          <input id="va-approved-amount" class="swal2-input" inputmode="numeric" autocomplete="off" placeholder="Masukkan nominal transfer" style="width:100%;margin:0;" />
          <div class="text-xs text-slate-400">Saldo member dan saldo provider LoketBayar akan bertambah sebesar nominal ini.</div>
        </div>
      `,
      showCancelButton: true,
      reverseButtons: true,
      focusCancel: true,
      confirmButtonText: "Approve",
      cancelButtonText: "Batal",
      didOpen: () => {
        const amountInput = document.getElementById("va-approved-amount") as HTMLInputElement | null;
        if (!amountInput) return;
        const format = () => {
          amountInput.value = formatAmountInput(amountInput.value);
        };
        amountInput.addEventListener("input", format);
        amountInput.focus();
      },
      preConfirm: () => {
        const amountInput = document.getElementById("va-approved-amount") as HTMLInputElement | null;
        const approvedAmount = Number(amountDigits(String(amountInput?.value || "")));
        if (!Number.isFinite(approvedAmount) || approvedAmount <= 0) {
          Swal.showValidationMessage("Nominal transfer aktual wajib diisi.");
          return false;
        }
        return { approvedAmount };
      },
    });
    if (!result.isConfirmed) return;

    const approvedAmount = Number(result.value?.approvedAmount || 0);

    const r = await fetch(`/api/admin/deposit/va/approve?id=${row.id}`, {
      method: "POST",
      headers: { "Content-Type": "application/json", ...authHeader() },
      body: JSON.stringify({ note: note.trim(), approved_amount: approvedAmount }),
    });
    const j = await r.json().catch(() => ({}));
    if (!r.ok || !j.ok) {
      await alertError(j.error || "Gagal approve deposit VA");
      return;
    }

    setNote("");
    if (appliedStatus === "ticket" || appliedStatus === "pending") {
      setItems((prev) => prev.filter((item) => item.id !== row.id));
    }
    await alertSuccess(`Deposit VA berhasil di-approve Rp ${fmtID(j.approved_amount || row.amount)}.`);
    void load(offset, appliedStatus, appliedTicket, appliedMemberID, appliedFrom, appliedTo, { showLoading: false });
  }

  async function reject(row: DepositVAReq) {
    if (!canProcess(row)) return;

    const confirmed = await alertConfirm({
      title: "Reject Deposit VA",
      text: `Reject tiket ${row.ref_id || row.id}?`,
      confirmButtonText: "Reject",
    });
    if (!confirmed) return;

    const r = await fetch(`/api/admin/deposit/va/reject?id=${row.id}`, {
      method: "POST",
      headers: { "Content-Type": "application/json", ...authHeader() },
      body: JSON.stringify({ note: note.trim() }),
    });
    const j = await r.json().catch(() => ({}));
    if (!r.ok || !j.ok) {
      await alertError(j.error || "Gagal reject deposit VA");
      return;
    }

    setNote("");
    if (appliedStatus === "ticket" || appliedStatus === "pending") {
      setItems((prev) => prev.filter((item) => item.id !== row.id));
    }
    await alertSuccess("Deposit VA berhasil ditolak.");
    void load(offset, appliedStatus, appliedTicket, appliedMemberID, appliedFrom, appliedTo, { showLoading: false });
  }

  const columns: DataTableColumn<DepositVAReq>[] = [
    {
      id: "id",
      header: "ID",
      thClassName: "whitespace-nowrap",
      tdClassName: "whitespace-nowrap font-medium text-slate-100",
      render: (d) => d.id,
    },
    {
      id: "member",
      header: "Member",
      thClassName: "whitespace-nowrap",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (d) => memberLabel(d),
    },
    {
      id: "ticket",
      header: "Tiket",
      thClassName: "whitespace-nowrap",
      tdClassName: "whitespace-nowrap font-mono text-cyan-200",
      render: (d) => d.ref_id || "-",
    },
    {
      id: "amount",
      header: "Nominal VA",
      thClassName: "whitespace-nowrap",
      tdClassName: "whitespace-nowrap text-cyan-100",
      render: (d) => (
        <div>
          <div className="flex items-center gap-1.5">
            <span className="font-semibold">Rp {fmtID(d.amount)}</span>
            <Button
              type="button"
              className="h-7 w-7 rounded-lg px-0 text-cyan-100"
              variant="outline"
              onClick={(event) => {
                event.stopPropagation();
                void copyAmount(d);
              }}
              aria-label="Copy nominal VA"
              title="Copy nominal VA"
            >
              {copiedID === d.id ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
            </Button>
          </div>
          {copiedID === d.id ? <div className="text-[11px] font-medium text-emerald-200">Tersalin</div> : null}
        </div>
      ),
    },
    {
      id: "requested",
      header: "Diminta",
      thClassName: "whitespace-nowrap",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (d) => (
        <div>
          <div>Rp {fmtID(requestedAmount(d))}</div>
          <div className="text-xs text-slate-400">Kode {fmtID(d.unique_code || 0)}</div>
        </div>
      ),
    },
    {
      id: "bank",
      header: "Bank VA",
      thClassName: "whitespace-nowrap",
      tdClassName: "text-slate-200",
      render: (d) => (
        <div className="min-w-56">
          <div className="font-medium text-slate-100">{d.bank_nama || "-"}</div>
          <div className="text-xs text-slate-300">{d.bank_atas_nama || "-"}</div>
          <div className="font-mono text-xs text-cyan-200">{d.bank_nomor_rekening || "-"}</div>
        </div>
      ),
    },
    {
      id: "status",
      header: "Status",
      thClassName: "whitespace-nowrap",
      tdClassName: "whitespace-nowrap",
      render: (d) => (
        <span className={`inline-flex rounded-md border px-2 py-1 text-xs font-semibold uppercase ${statusClass(d.status)}`}>{d.status}</span>
      ),
    },
    {
      id: "waktu",
      header: "Waktu",
      thClassName: "whitespace-nowrap",
      tdClassName: "whitespace-nowrap text-slate-300",
      render: (d) => (
        <div>
          <div>{formatDate(d.dibuat_pada)}</div>
          {d.diproses_pada ? <div className="text-xs text-slate-500">{formatDate(d.diproses_pada)}</div> : null}
        </div>
      ),
    },
  ];

  return (
    <div className="space-y-4 p-2">
      <div className="rounded-3xl border border-white/10 bg-[linear-gradient(135deg,rgba(15,23,42,.95),rgba(30,41,59,.82))] px-5 py-5 shadow-[0_18px_60px_rgba(2,6,23,.28)]">
        <h1 className="text-2xl font-semibold tracking-tight text-white">Deposit VA</h1>
        <p className="mt-2 max-w-3xl text-sm leading-6 text-slate-300">{subtitle}</p>
      </div>

      <div className="rounded-3xl border border-white/10 bg-slate-950/70 p-4 shadow-[0_18px_60px_rgba(2,6,23,.2)] backdrop-blur">
        <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div className="text-sm text-slate-300">{refreshing ? "Memperbarui data VA..." : "Refresh otomatis tiap 10 detik."}</div>
          <Button
            className="h-10"
            variant="primary"
            onClick={() => void load(offset, appliedStatus, appliedTicket, appliedMemberID, appliedFrom, appliedTo, { showLoading: items.length === 0 })}
            disabled={loading || refreshing}
          >
            {loading || refreshing ? "Loading..." : "Refresh"}
          </Button>
        </div>
      </div>

      <div className="rounded-3xl border border-white/10 bg-slate-950/70 p-4 shadow-[0_18px_60px_rgba(2,6,23,.2)] backdrop-blur">
        <div className="flex items-center justify-between gap-3 md:hidden">
          <div>
            <div className="text-sm font-semibold text-white">Filter Deposit VA</div>
            <div className="text-xs text-slate-400">Status, tiket, member, dan tanggal.</div>
          </div>
          <Button type="button" variant="outline" className="h-10" onClick={() => setShowFiltersMobile((v) => !v)}>
            <Filter className="mr-2 h-4 w-4" />
            {showFiltersMobile ? "Tutup Filter" : "Buka Filter"}
          </Button>
        </div>

        <div className={`mt-4 grid gap-3 md:mt-0 md:grid-cols-12 ${showFiltersMobile ? "" : "hidden md:grid"}`}>
          <div className="md:col-span-2">
            <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-slate-400">Status</div>
            <select
              className="h-11 w-full rounded-xl border border-white/10 bg-slate-900 px-3 text-sm text-white outline-none"
              value={status}
              onChange={(e) => setStatus(e.target.value)}
            >
              <option value="ticket">Ticket</option>
              <option value="pending">Pending</option>
              <option value="approved">Approved</option>
              <option value="rejected">Rejected</option>
              <option value="cancelled">Cancelled</option>
              <option value="all">Semua</option>
            </select>
          </div>

          <div className="md:col-span-3">
            <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-slate-400">Tiket</div>
            <Input className="h-11 font-mono" value={ticket} onChange={(e) => setTicket(e.target.value)} placeholder="15003122" />
          </div>

          <div className="md:col-span-3">
            <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-slate-400">Member</div>
            <select
              className="h-11 w-full rounded-xl border border-white/10 bg-slate-900 px-3 text-sm text-white outline-none"
              value={memberID}
              onChange={(e) => setMemberID(e.target.value)}
            >
              <option value="">Semua Member</option>
              {members.map((m) => (
                <option key={m.id} value={String(m.id)}>
                  {m.nama || `Member #${m.id}`}
                </option>
              ))}
            </select>
          </div>

          <div className="md:col-span-2">
            <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-slate-400">Dari</div>
            <Input className="h-11" type="date" value={from} onChange={(e) => setFrom(e.target.value)} />
          </div>

          <div className="md:col-span-2">
            <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-slate-400">Sampai</div>
            <Input className="h-11" type="date" value={to} onChange={(e) => setTo(e.target.value)} />
          </div>

          <div className="md:col-span-3">
            <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-transparent">Aksi</div>
            <div className="flex flex-wrap gap-2">
              <Button
                className="h-11"
                variant="primary"
                onClick={() => {
                  hasLoadedRef.current = false;
                  setItems([]);
                  setOffset(0);
                  setAppliedStatus(status);
                  setAppliedTicket(ticket);
                  setAppliedMemberID(memberID);
                  setAppliedFrom(from);
                  setAppliedTo(to);
                }}
              >
                <Search className="h-4 w-4" />
                Terapkan
              </Button>

              <Button
                className="h-11"
                variant="outline"
                onClick={() => {
                  hasLoadedRef.current = false;
                  setItems([]);
                  setOffset(0);
                  setStatus("ticket");
                  setTicket("");
                  setMemberID("");
                  setFrom(defaultFrom);
                  setTo(defaultTo);
                  setAppliedStatus("ticket");
                  setAppliedTicket("");
                  setAppliedMemberID("");
                  setAppliedFrom(defaultFrom);
                  setAppliedTo(defaultTo);
                }}
              >
                <RotateCcw className="h-4 w-4" />
                Reset
              </Button>
            </div>
          </div>
        </div>
      </div>

      {!readOnly ? (
        <div className="rounded-3xl border border-white/10 bg-slate-950/70 p-4 shadow-[0_18px_60px_rgba(2,6,23,.2)] backdrop-blur">
          <div className="mb-2 text-sm text-slate-300">Note untuk approve/reject (opsional)</div>
          <Input className="h-11" value={note} onChange={(e) => setNote(e.target.value)} placeholder="contoh: callback valid manual" />
        </div>
      ) : null}

      <div className="space-y-2 md:hidden">
        {loading && items.length === 0 ? (
          <div className="rounded-xl border border-white/10 bg-slate-950/70 px-4 py-5 text-center text-sm text-slate-300">Memuat data...</div>
        ) : items.length === 0 ? (
          <div className="rounded-xl border border-white/10 bg-slate-950/70 px-4 py-5 text-center text-sm text-slate-300">Belum ada deposit VA.</div>
        ) : (
          items.map((d) => (
            <div key={d.id} className="rounded-xl border border-white/10 bg-slate-950/80 p-3 shadow-[0_10px_24px_rgba(2,6,23,.2)]">
              <div className="flex items-start justify-between gap-2">
                <div className="min-w-0 flex-1">
                  <div className="truncate text-sm font-semibold leading-5 text-white">{memberLabel(d)}</div>
                  <div className="truncate text-xs leading-5 text-cyan-200">Tiket {d.ref_id || "-"}</div>
                  <div className="mt-0.5 truncate text-xs leading-5 text-slate-300">{d.bank_nama || "-"}</div>
                  <div className="truncate text-xs leading-5 text-slate-400">{d.bank_nomor_rekening || "-"}</div>
                </div>
                <div className="shrink-0 text-right">
                  <div className="text-sm font-semibold leading-5 text-cyan-100">Rp {fmtID(d.amount)}</div>
                  <div className="text-xs text-slate-400">Diminta Rp {fmtID(requestedAmount(d))}</div>
                  <span className={`mt-1 inline-flex rounded-md border px-2 py-1 text-[11px] font-semibold uppercase ${statusClass(d.status)}`}>{d.status}</span>
                </div>
              </div>

              {!readOnly && canProcess(d) ? (
                <div className="mt-3 grid grid-cols-2 gap-2">
                  <Button className="h-7 rounded-lg px-2 text-xs" variant="success" onClick={() => approve(d)}>
                    <Check className="h-3.5 w-3.5" />
                    Approve
                  </Button>
                  <Button className="h-7 rounded-lg px-2 text-xs" variant="danger" onClick={() => reject(d)}>
                    <X className="h-3.5 w-3.5" />
                    Reject
                  </Button>
                </div>
              ) : null}
            </div>
          ))
        )}

        {items.length > 0 || currentPage > 1 || hasNext ? (
          <div className="flex items-center justify-between gap-3 rounded-xl border border-white/10 bg-slate-950/70 p-2.5">
            <Button
              className="h-10 w-10 px-0"
              variant="outline"
              onClick={() => {
                if (currentPage > 1) setOffset((v) => Math.max(0, v - PAGE_SIZE));
              }}
              disabled={currentPage <= 1}
              aria-label="Halaman sebelumnya"
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <div className="text-sm font-medium text-slate-200">Halaman {currentPage} / {totalPages}</div>
            <Button
              className="h-10 w-10 px-0"
              variant="outline"
              onClick={() => {
                if (hasNext) setOffset((v) => v + PAGE_SIZE);
              }}
              disabled={!hasNext}
              aria-label="Halaman berikutnya"
            >
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        ) : null}
      </div>

      <div className="hidden rounded-3xl border border-white/10 bg-slate-950/70 p-4 shadow-[0_18px_60px_rgba(2,6,23,.2)] backdrop-blur md:block">
        <DataTable
          columns={columns}
          rows={items}
          rowKey={(d) => d.id}
          rowNumberStart={offset + 1}
          minWidthClassName="min-w-[1120px]"
          actions={
            readOnly
              ? undefined
              : {
                  header: "Aksi",
                  align: "left",
                  render: (d) =>
                    canProcess(d) ? (
                      <div className="flex gap-2">
                        <Button className="h-9 w-9 px-0" variant="success" onClick={() => approve(d)} aria-label="Approve VA" title="Approve VA">
                          <Check className="h-4 w-4" />
                        </Button>
                        <Button className="h-9 w-9 px-0" variant="danger" onClick={() => reject(d)} aria-label="Reject VA" title="Reject VA">
                          <X className="h-4 w-4" />
                        </Button>
                      </div>
                    ) : (
                      <span className="text-xs text-slate-500">Selesai</span>
                    ),
                }
          }
          loading={loading && items.length === 0}
          pagination={{
            page: currentPage,
            totalPages,
            onPrev: () => {
              if (currentPage > 1) setOffset((v) => Math.max(0, v - PAGE_SIZE));
            },
            onNext: () => {
              if (hasNext) setOffset((v) => v + PAGE_SIZE);
            },
            onPageChange: (page) => setOffset((page - 1) * PAGE_SIZE),
            disablePrev: currentPage <= 1,
            disableNext: !hasNext,
          }}
          emptyText="Belum ada deposit VA."
        />
      </div>
    </div>
  );
}
