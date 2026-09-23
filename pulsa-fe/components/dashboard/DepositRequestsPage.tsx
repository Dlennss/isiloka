"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { usePathname } from "next/navigation";
import { CalendarRange, Check, ChevronLeft, ChevronRight, Copy, Filter, RotateCcw, Search, X } from "lucide-react";
import Swal from "sweetalert2";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { alertConfirm, alertError, alertSuccess } from "@/components/ui/alerts";
import { fmtID } from "@/lib/format";

type DepositReq = {
  id: number;
  member_id: number;
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
  bukti_url: string;
  status: "pending" | "approved" | "rejected";
  note?: string;
  dibuat_pada: string;
};

type MemberOption = {
  id: number;
  nama?: string;
  role?: string;
};

type Props = {
  title?: string;
  description?: string;
};

const PAGE_SIZE = 10;
const POLL_MS = 10_000;
const DEPOSIT_ALARM_SRC = "/audio/cash.mp3";

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

function depositTime(row: DepositReq) {
  const parsed = new Date(row.dibuat_pada).getTime();
  return Number.isFinite(parsed) ? parsed : 0;
}

function sortDepositRows(rows: DepositReq[]) {
  return [...rows].sort((a, b) => depositTime(a) - depositTime(b) || a.id - b.id);
}

function sameDepositRow(a: DepositReq, b: DepositReq) {
  return (
    a.id === b.id &&
    a.member_id === b.member_id &&
    (a.member_nama || "") === (b.member_nama || "") &&
    (a.bank_id || null) === (b.bank_id || null) &&
    (a.bank_nama || "") === (b.bank_nama || "") &&
    (a.bank_nomor_rekening || "") === (b.bank_nomor_rekening || "") &&
    (a.bank_atas_nama || "") === (b.bank_atas_nama || "") &&
    Number(a.amount || 0) === Number(b.amount || 0) &&
    Number(a.requested_amount || 0) === Number(b.requested_amount || 0) &&
    Number(a.unique_code || 0) === Number(b.unique_code || 0) &&
    Number(a.approved_amount || 0) === Number(b.approved_amount || 0) &&
    a.metode === b.metode &&
    a.bukti_url === b.bukti_url &&
    a.status === b.status &&
    (a.note || "") === (b.note || "") &&
    a.dibuat_pada === b.dibuat_pada
  );
}

function sameDepositRows(a: DepositReq[], b: DepositReq[]) {
  return a.length === b.length && a.every((row, index) => sameDepositRow(row, b[index]));
}

function memberLabel(row: DepositReq) {
  return row.member_nama ? `${row.member_nama} (#${row.member_id})` : `Member #${row.member_id}`;
}

function amountDigits(value: string | number) {
  return String(value || "").replace(/\D/g, "");
}

function formatAmountInput(value: string | number) {
  const digits = amountDigits(value);
  return digits ? Number(digits).toLocaleString("id-ID") : "";
}

function parseBankRefIDs(value: string) {
  return String(value || "")
    .split(/[\s,;]+/g)
    .map((x) => x.trim())
    .filter(Boolean);
}

function depositCopySeparator(bankName?: string) {
  const bank = String(bankName || "").toUpperCase();
  if (bank.includes("BRI") || bank.includes("RAKYAT")) return ".";
  return ",";
}

function formatDepositCopyAmount(value: string | number, bankName?: string) {
  const digits = amountDigits(value);
  if (!digits) return "";
  return digits.replace(/\B(?=(\d{3})+(?!\d))/g, depositCopySeparator(bankName));
}

export default function DepositRequestsPage({
  title = "Deposit Request",
  description = "Daftar request deposit untuk dipantau tanpa aksi approve atau reject.",
}: Props) {
  const pathname = usePathname();
  const isOperatorMode = pathname?.startsWith("/dashboard/operator") ?? false;
  const isDepositAlarmPage =
    (pathname?.startsWith("/dashboard/admin/deposits") ?? false) ||
    (pathname?.startsWith("/dashboard/wallet/deposits") ?? false);
  const canProcessDeposit = !isOperatorMode;
  const now = useMemo(() => new Date(), []);
  const defaultFrom = useMemo(
    () => toLocalDateInputValue(new Date(now.getFullYear(), now.getMonth(), 1)),
    [now],
  );
  const defaultTo = useMemo(() => toLocalDateInputValue(now), [now]);

  const [items, setItems] = useState<DepositReq[]>([]);
  const [members, setMembers] = useState<MemberOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [showFiltersMobile, setShowFiltersMobile] = useState(false);
  const [offset, setOffset] = useState(0);
  const [hasNext, setHasNext] = useState(false);
  const [status, setStatus] = useState("pending");
  const [memberID, setMemberID] = useState("");
  const [from, setFrom] = useState(defaultFrom);
  const [to, setTo] = useState(defaultTo);
  const [appliedStatus, setAppliedStatus] = useState("pending");
  const [appliedMemberID, setAppliedMemberID] = useState("");
  const [appliedFrom, setAppliedFrom] = useState(defaultFrom);
  const [appliedTo, setAppliedTo] = useState(defaultTo);
  const [note, setNote] = useState("");
  const [copiedAmountID, setCopiedAmountID] = useState<number | null>(null);

  const pollInFlightRef = useRef(false);
  const hasLoadedRef = useRef(false);
  const intervalRef = useRef<number | null>(null);
  const alarmAudioRef = useRef<HTMLAudioElement | null>(null);
  const alarmPrimedRef = useRef(false);
  const alarmShouldRingRef = useRef(false);
  const currentPage = Math.floor(offset / PAGE_SIZE) + 1;
  const totalPages = hasNext ? currentPage + 1 : currentPage;
  const getAlarmAudio = useCallback(() => {
    if (!alarmAudioRef.current) {
      const audio = new Audio(DEPOSIT_ALARM_SRC);
      audio.preload = "auto";
      audio.loop = true;
      alarmAudioRef.current = audio;
    }
    return alarmAudioRef.current;
  }, []);
  const stopDepositAlarm = useCallback(() => {
    alarmShouldRingRef.current = false;
    const audio = alarmAudioRef.current;
    if (!audio) return;
    audio.pause();
    audio.currentTime = 0;
  }, []);
  const startDepositAlarm = useCallback(async () => {
    alarmShouldRingRef.current = true;
    const audio = getAlarmAudio();
    audio.loop = true;
    audio.volume = 1;
    if (!audio.paused) return;

    try {
      audio.currentTime = 0;
      await audio.play();
      alarmPrimedRef.current = true;
    } catch {
      alarmPrimedRef.current = false;
    }
  }, [getAlarmAudio]);
  const primeDepositAlarm = useCallback(async () => {
    if (!isDepositAlarmPage || alarmPrimedRef.current) return;

    const audio = getAlarmAudio();
    const previousVolume = audio.volume;

    try {
      audio.volume = 0;
      audio.loop = false;
      audio.currentTime = 0;
      await audio.play();
      audio.pause();
      audio.currentTime = 0;
      audio.loop = true;
      audio.volume = previousVolume || 1;
      alarmPrimedRef.current = true;

      if (alarmShouldRingRef.current) {
        void startDepositAlarm();
      }
    } catch {
      audio.loop = true;
      audio.volume = previousVolume || 1;
      alarmPrimedRef.current = false;
    }
  }, [getAlarmAudio, isDepositAlarmPage, startDepositAlarm]);
  const syncDepositAlarm = useCallback(
    (rows: DepositReq[], rowStatus: string) => {
      if (isDepositAlarmPage && rowStatus === "pending" && rows.length > 0) {
        void startDepositAlarm();
      } else {
        stopDepositAlarm();
      }
    },
    [isDepositAlarmPage, startDepositAlarm, stopDepositAlarm],
  );
  const load = useCallback(
    async (
      nextOffset = offset,
      nextStatus = appliedStatus,
      nextMemberID = appliedMemberID,
      nextFrom = appliedFrom,
      nextTo = appliedTo,
      options: { showLoading?: boolean } = {},
    ) => {
      if (pollInFlightRef.current) return;
      pollInFlightRef.current = true;
      const showLoading = options.showLoading ?? !hasLoadedRef.current;
      if (showLoading) {
        setLoading(true);
      } else {
        setRefreshing(true);
      }

      try {
        const qs = new URLSearchParams();
        if (nextStatus !== "all") qs.set("status", nextStatus);
        if (nextMemberID.trim()) qs.set("member_id", nextMemberID.trim());
        if (nextFrom) qs.set("from", nextFrom);
        if (nextTo) qs.set("to", nextTo);
        qs.set("limit", String(PAGE_SIZE + 1));
        qs.set("offset", String(nextOffset));
        qs.set("order", "asc");

        const r = await fetch(`/api/admin/deposit/requests?${qs.toString()}`, {
          headers: authHeader(),
          cache: "no-store",
        });
        const j = await r.json().catch(() => ({}));
        const all: DepositReq[] = Array.isArray(j.items) ? j.items : [];
        const nextRows = sortDepositRows(all.slice(0, PAGE_SIZE));
        setHasNext(all.length > PAGE_SIZE);
        setItems((prev) => (sameDepositRows(prev, nextRows) ? prev : nextRows));
        syncDepositAlarm(nextRows, nextStatus);
      } finally {
        hasLoadedRef.current = true;
        if (showLoading) {
          setLoading(false);
        } else {
          setRefreshing(false);
        }
        pollInFlightRef.current = false;
      }
    },
    [offset, appliedStatus, appliedMemberID, appliedFrom, appliedTo, syncDepositAlarm],
  );

  useEffect(() => {
    void load(offset, appliedStatus, appliedMemberID, appliedFrom, appliedTo);
  }, [load, offset, appliedStatus, appliedMemberID, appliedFrom, appliedTo]);

  useEffect(() => {
    if (intervalRef.current) {
      window.clearInterval(intervalRef.current);
    }

    intervalRef.current = window.setInterval(() => {
      if (document.visibilityState !== "visible") return;
      void load(offset, appliedStatus, appliedMemberID, appliedFrom, appliedTo, { showLoading: false });
    }, POLL_MS);

    return () => {
      if (intervalRef.current) {
        window.clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [load, offset, appliedStatus, appliedMemberID, appliedFrom, appliedTo]);

  useEffect(() => {
    syncDepositAlarm(items, appliedStatus);
  }, [appliedStatus, items, syncDepositAlarm]);

  useEffect(() => {
    if (!isDepositAlarmPage) return;

    const handleUserGesture = () => {
      if (alarmShouldRingRef.current) {
        void startDepositAlarm();
      } else {
        void primeDepositAlarm();
      }
    };

    window.addEventListener("pointerdown", handleUserGesture, { passive: true });
    window.addEventListener("touchstart", handleUserGesture, { passive: true });
    window.addEventListener("keydown", handleUserGesture);

    return () => {
      window.removeEventListener("pointerdown", handleUserGesture);
      window.removeEventListener("touchstart", handleUserGesture);
      window.removeEventListener("keydown", handleUserGesture);
      stopDepositAlarm();
    };
  }, [isDepositAlarmPage, primeDepositAlarm, startDepositAlarm, stopDepositAlarm]);

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

  const copyDepositAmount = useCallback(
    async (row: DepositReq) => {
      const copyAmount = formatDepositCopyAmount(row.amount, row.bank_nama);
      if (!copyAmount) {
        await alertError("Nominal tidak valid untuk dicopy.");
        return;
      }

      try {
        await copyText(copyAmount);
        setCopiedAmountID(row.id);
        window.setTimeout(() => {
          setCopiedAmountID((current) => (current === row.id ? null : current));
        }, 1400);
      } catch {
        await alertError("Gagal copy nominal.");
      }
    },
    [copyText],
  );

  function renderAmount(row: DepositReq, align: "left" | "right" = "left") {
    const approvedAmount = Number(row.approved_amount || 0);
    const requestedAmount = Number(row.requested_amount || 0);
    const corrected = row.status === "approved" && approvedAmount > 0 && approvedAmount !== Number(row.amount || 0);
    const copied = copiedAmountID === row.id;

    return (
      <div className={align === "right" ? "text-right" : undefined}>
        <div className={`flex items-center gap-1.5 ${align === "right" ? "justify-end" : ""}`}>
          <span>Rp {fmtID(row.amount)}</span>
          <Button
            type="button"
            className="h-7 w-7 rounded-lg px-0 text-emerald-700"
            variant="outline"
            onClick={(event) => {
              event.stopPropagation();
              void copyDepositAmount(row);
            }}
            aria-label="Copy nominal deposit"
            title="Copy nominal sesuai format bank"
          >
            {copied ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
          </Button>
        </div>
        {copied ? <div className="text-[11px] font-bold text-emerald-800">Tersalin</div> : null}
        {requestedAmount > 0 && requestedAmount !== Number(row.amount || 0) ? <div className="text-xs font-bold text-emerald-800">Saldo masuk Rp {fmtID(requestedAmount)}</div> : null}
        {corrected ? <div className="text-xs font-bold text-amber-800">Saldo masuk Rp {fmtID(approvedAmount)}</div> : null}
      </div>
    );
  }

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
            role: String(x.role || ""),
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

  async function approve(row: DepositReq) {
    const ticketAmount = Number(row.amount || 0);
    const defaultAmount = Number(row.approved_amount || row.requested_amount || row.amount || 0);
    const result = await Swal.fire<{ approvedAmount: number; bankRefIDs: string[] }>({
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
      title: "Approve Deposit",
      html: `
        <div class="space-y-3 text-left">
          <div>Nominal tiket: <b>Rp ${fmtID(ticketAmount)}</b></div>
          <label for="deposit-approve-amount" class="block text-xs font-semibold uppercase tracking-wide text-slate-400">Nominal saldo masuk</label>
          <input id="deposit-approve-amount" class="swal2-input" inputmode="numeric" autocomplete="off" value="${formatAmountInput(defaultAmount)}" style="width:100%;margin:0;" />
          <label for="deposit-bank-refids" class="block text-xs font-semibold uppercase tracking-wide text-slate-400">Ref ID mutasi bank</label>
          <textarea id="deposit-bank-refids" class="swal2-textarea" rows="3" placeholder="BMIN-...&#10;BMIN-..." style="width:100%;margin:0;"></textarea>
          <div class="text-xs text-slate-400">Opsional, maksimal 4 refid. Jika dikosongkan, sistem mencatat mutasi kredit bank otomatis saat approval.</div>
        </div>
      `,
      showCancelButton: true,
      reverseButtons: true,
      focusCancel: true,
      confirmButtonText: "Approve",
      cancelButtonText: "Batal",
      didOpen: () => {
        const amountInput = document.getElementById("deposit-approve-amount") as HTMLInputElement | null;
        if (!amountInput) return;
        const format = () => {
          amountInput.value = formatAmountInput(amountInput.value);
        };
        amountInput.addEventListener("input", format);
        amountInput.focus();
        amountInput.select();
      },
      preConfirm: () => {
        const amountInput = document.getElementById("deposit-approve-amount") as HTMLInputElement | null;
        const refInput = document.getElementById("deposit-bank-refids") as HTMLTextAreaElement | null;
        const approvedAmount = Number(amountDigits(String(amountInput?.value || "")));
        const bankRefIDs = parseBankRefIDs(refInput?.value || "");
        if (bankRefIDs.length > 4) {
          Swal.showValidationMessage("Maksimal 4 refid mutasi bank.");
          return false;
        }
        if (bankRefIDs.length === 0 && (!Number.isFinite(approvedAmount) || approvedAmount <= 0)) {
          Swal.showValidationMessage("Nominal saldo masuk wajib diisi.");
          return false;
        }
        return { approvedAmount, bankRefIDs };
      },
    });
    if (!result.isConfirmed) return;

    const approvedAmount = Number(result.value?.approvedAmount || 0);
    const bankRefIDs = result.value?.bankRefIDs || [];

    const r = await fetch(`/api/admin/deposit/approve?id=${row.id}`, {
      method: "POST",
      headers: { "Content-Type": "application/json", ...authHeader() },
      body: JSON.stringify({ note: note.trim(), approved_amount: approvedAmount, bank_ref_ids: bankRefIDs }),
    });
    const j = await r.json().catch(() => ({}));
    if (!r.ok || !j.ok) {
      await alertError(j.error || "Gagal approve");
      return;
    }
    const creditedAmount = Number(j.approved_amount || approvedAmount);
    setNote("");
    if (appliedStatus === "pending") {
      setItems((prev) => prev.filter((item) => item.id !== row.id));
    }
    await alertSuccess(`Deposit berhasil di-approve Rp ${fmtID(creditedAmount)}. Ref ID: ${j.ref_id || "-"}`);
    void load(offset, appliedStatus, appliedMemberID, appliedFrom, appliedTo, { showLoading: false });
  }

  async function reject(id: number) {
    const confirmed = await alertConfirm({
      title: "Reject Deposit",
      text: "Request deposit ini akan ditolak. Lanjutkan?",
      confirmButtonText: "Reject",
    });
    if (!confirmed) return;

    const r = await fetch(`/api/admin/deposit/reject?id=${id}`, {
      method: "POST",
      headers: { "Content-Type": "application/json", ...authHeader() },
      body: JSON.stringify({ note: note.trim() }),
    });
    const j = await r.json().catch(() => ({}));
    if (!r.ok || !j.ok) {
      await alertError(j.error || "Gagal reject");
      return;
    }
    setNote("");
    if (appliedStatus === "pending") {
      setItems((prev) => prev.filter((row) => row.id !== id));
    }
    await alertSuccess("Request deposit berhasil ditolak.");
    void load(offset, appliedStatus, appliedMemberID, appliedFrom, appliedTo, { showLoading: false });
  }

  const columns: DataTableColumn<DepositReq>[] = [
    { id: "id", header: "ID", thClassName: "whitespace-nowrap", tdClassName: "whitespace-nowrap text-slate-100 font-medium", render: (d) => d.id },
    {
      id: "member",
      header: "Member",
      thClassName: "whitespace-nowrap",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (d) => (d.member_nama ? `${d.member_nama} (#${d.member_id})` : `#${d.member_id}`),
    },
    {
      id: "amount",
      header: "Nominal",
      thClassName: "whitespace-nowrap",
      tdClassName: "whitespace-nowrap font-semibold text-emerald-700",
      render: (d) => renderAmount(d),
    },
    {
      id: "metode",
      header: "Metode",
      thClassName: "whitespace-nowrap",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (d) => d.metode,
    },
    {
      id: "bank",
      header: "Bank Tujuan",
      thClassName: "whitespace-nowrap",
      tdClassName: "text-slate-200",
      render: (d) => (
        <div className="min-w-55">
          <div className="font-medium text-slate-100">{d.bank_nama || "-"}</div>
          <div className="text-xs text-slate-300">{d.bank_atas_nama || "-"}</div>
          <div className="font-mono text-xs text-emerald-700">{d.bank_nomor_rekening || "-"}</div>
        </div>
      ),
    },
    {
      id: "status",
      header: "Status",
      thClassName: "whitespace-nowrap",
      tdClassName: "whitespace-nowrap text-slate-200 uppercase",
      render: (d) => d.status,
    },
    {
      id: "waktu",
      header: "Waktu",
      thClassName: "whitespace-nowrap",
      tdClassName: "whitespace-nowrap text-slate-300",
      render: (d) => new Date(d.dibuat_pada).toLocaleString("id-ID"),
    },
  ];

  return (
    <div className="space-y-4 p-2">
      <div className="rounded-3xl border border-emerald-900 bg-[linear-gradient(135deg,#13464b,#087e8b)] px-5 py-5 shadow-[0_18px_46px_rgba(6,78,59,.22)]">
        <div className="inline-flex items-center gap-2 rounded-full border border-emerald-200/70 bg-white px-3 py-1 text-[11px] font-bold uppercase tracking-[0.24em] text-emerald-800">
          <CalendarRange className="h-3.5 w-3.5" />
          Deposit Monitor
        </div>
        <h1 className="mt-3 text-2xl font-semibold tracking-tight text-white">{title}</h1>
        <p className="mt-2 max-w-3xl text-sm font-medium leading-6 text-emerald-100">{description}</p>
      </div>

      <div className="rounded-3xl border border-white/10 bg-slate-950/70 p-4 shadow-[0_18px_60px_rgba(2,6,23,.2)] backdrop-blur">
        <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div className="text-sm text-slate-300">{refreshing ? "Memperbarui data deposit..." : "Monitor request deposit pending dengan refresh otomatis tiap 10 detik."}</div>

          <div className="flex flex-wrap items-center gap-2">
            <Button
              className="h-10"
              variant="primary"
              onClick={() => void load(offset, appliedStatus, appliedMemberID, appliedFrom, appliedTo, { showLoading: items.length === 0 })}
              disabled={loading || refreshing}
            >
              {loading || refreshing ? "Loading..." : "Refresh"}
            </Button>
          </div>
        </div>
      </div>

      <div className="rounded-3xl border border-white/10 bg-slate-950/70 p-4 shadow-[0_18px_60px_rgba(2,6,23,.2)] backdrop-blur">
        <div className="flex items-center justify-between gap-3 md:hidden">
          <div>
            <div className="text-sm font-semibold text-white">Filter Deposit</div>
            <div className="text-xs text-slate-400">Atur status, member, dan tanggal.</div>
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
              <option value="pending">Pending</option>
              <option value="approved">Approved</option>
              <option value="rejected">Rejected</option>
              <option value="all">Semua</option>
            </select>
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
                  setStatus("pending");
                  setMemberID("");
                  setFrom(defaultFrom);
                  setTo(defaultTo);
                  setAppliedStatus("pending");
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

      {canProcessDeposit ? (
        <div className="rounded-3xl border border-white/10 bg-slate-950/70 p-4 shadow-[0_18px_60px_rgba(2,6,23,.2)] backdrop-blur">
          <div className="mb-2 text-sm text-slate-300">Note untuk approve/reject (opsional)</div>
          <Input className="h-11" value={note} onChange={(e) => setNote(e.target.value)} placeholder="contoh: bukti valid" />
        </div>
      ) : null}

      <div className="space-y-2 md:hidden">
        {loading && items.length === 0 ? (
          <div className="rounded-xl border border-white/10 bg-slate-950/70 px-4 py-5 text-center text-sm text-slate-300">Memuat data...</div>
        ) : items.length === 0 ? (
          <div className="rounded-xl border border-white/10 bg-slate-950/70 px-4 py-5 text-center text-sm text-slate-300">Belum ada request deposit.</div>
        ) : (
          items.map((d) => (
            <div key={d.id} className="rounded-xl border border-white/10 bg-slate-950/80 p-3 shadow-[0_10px_24px_rgba(2,6,23,.2)]">
              <div className="flex items-start justify-between gap-2">
                <div className="min-w-0 flex-1">
                  <div className="truncate text-sm font-semibold leading-5 text-white">{memberLabel(d)}</div>
                  <div className="mt-0.5 truncate text-xs leading-5 text-slate-300">{d.bank_nama || "-"}</div>
                  <div className="truncate text-xs leading-5 text-slate-400">a.n. {d.bank_atas_nama || "-"}</div>
                </div>
                <div className="shrink-0 text-sm font-semibold leading-5 text-emerald-700">
                  {renderAmount(d, "right")}
                </div>
              </div>

              {canProcessDeposit ? (
                <div className="mt-3 grid grid-cols-2 gap-2">
                  <Button className="h-7 rounded-lg px-2 text-xs" variant="success" onClick={() => approve(d)}>
                    <Check className="h-3.5 w-3.5" />
                    Approve
                  </Button>
                  <Button className="h-7 rounded-lg px-2 text-xs" variant="danger" onClick={() => reject(d.id)}>
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
          minWidthClassName="min-w-[980px]"
          actions={
            !canProcessDeposit
              ? undefined
              : {
                  header: "Aksi",
                  align: "left",
                  render: (d) => (
                    <div className="flex gap-2">
                      <Button className="h-9 w-9 px-0" variant="success" onClick={() => approve(d)} aria-label="Approve deposit" title="Approve">
                        <Check className="h-4 w-4" />
                      </Button>
                      <Button className="h-9 w-9 px-0" variant="danger" onClick={() => reject(d.id)} aria-label="Reject deposit" title="Reject">
                        <X className="h-4 w-4" />
                      </Button>
                    </div>
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
          emptyText="Belum ada request deposit."
        />
      </div>
    </div>
  );
}
