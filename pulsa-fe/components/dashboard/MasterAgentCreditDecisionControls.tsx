"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { CheckCircle2, FileWarning, Loader2, LockKeyhole, RotateCcw, UnlockKeyhole, XCircle } from "lucide-react";

type Props = {
  applicationId: number;
  requestedAmount: number;
  approvedAmount?: number;
  marketingNote?: string;
  analystNote?: string;
  analystRecommendation?: string;
  analystRecommendedAmount?: number;
  status: string;
  loanStatus?: string;
  mode?: "marketing" | "master" | "analyst" | "admin";
  canApprove?: boolean;
  approveBlockReason?: string;
};

type ApiBody = {
  ok?: boolean;
  error?: string;
};

type DecisionAction = "approved" | "rejected" | "revision_required" | "forward_to_analysis";

const revisionOptions = [
  { key: "ktp", label: "Foto KTP" },
  { key: "store", label: "Foto Toko" },
  { key: "selfie_ktp", label: "Dokumen Formulir" },
  { key: "selfie_marketing", label: "Foto Bersama Marketing" },
] as const;

function ReviewSignaturePad({
  label,
  signerName,
  value,
  onChange,
}: {
  label: string;
  signerName: string;
  value: string;
  onChange: (value: string) => void;
}) {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const drawingRef = useRef(false);
  const inkRef = useRef(Boolean(value));
  const [hasInk, setHasInk] = useState(Boolean(value));

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const resize = () => {
      const rect = canvas.getBoundingClientRect();
      const ratio = window.devicePixelRatio || 1;
      canvas.width = Math.max(1, Math.floor(rect.width * ratio));
      canvas.height = Math.max(1, Math.floor(rect.height * ratio));
      const ctx = canvas.getContext("2d");
      if (!ctx) return;
      ctx.setTransform(ratio, 0, 0, ratio, 0, 0);
      ctx.lineCap = "round";
      ctx.lineJoin = "round";
      ctx.lineWidth = 1.35;
      ctx.strokeStyle = "#087e8b";
    };
    resize();
    window.addEventListener("resize", resize);
    return () => window.removeEventListener("resize", resize);
  }, []);

  function point(event: React.PointerEvent<HTMLCanvasElement>) {
    const canvas = canvasRef.current;
    if (!canvas) return { x: 0, y: 0 };
    const rect = canvas.getBoundingClientRect();
    return { x: event.clientX - rect.left, y: event.clientY - rect.top };
  }

  function start(event: React.PointerEvent<HTMLCanvasElement>) {
    const canvas = canvasRef.current;
    const ctx = canvas?.getContext("2d");
    if (!canvas || !ctx) return;
    drawingRef.current = true;
    canvas.setPointerCapture(event.pointerId);
    const pos = point(event);
    ctx.beginPath();
    ctx.moveTo(pos.x, pos.y);
  }

  function move(event: React.PointerEvent<HTMLCanvasElement>) {
    if (!drawingRef.current) return;
    const canvas = canvasRef.current;
    const ctx = canvas?.getContext("2d");
    if (!canvas || !ctx) return;
    const pos = point(event);
    ctx.lineTo(pos.x, pos.y);
    ctx.stroke();
    inkRef.current = true;
    setHasInk(true);
  }

  function end(event: React.PointerEvent<HTMLCanvasElement>) {
    drawingRef.current = false;
    const canvas = canvasRef.current;
    if (!canvas) return;
    if (canvas.hasPointerCapture(event.pointerId)) canvas.releasePointerCapture(event.pointerId);
    if (inkRef.current || value) onChange(canvas.toDataURL("image/png"));
  }

  function clear() {
    const canvas = canvasRef.current;
    const ctx = canvas?.getContext("2d");
    if (!canvas || !ctx) return;
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    inkRef.current = false;
    setHasInk(false);
    onChange("");
  }

  return (
    <div className="rounded-[24px] border border-emerald-300 bg-[linear-gradient(180deg,#ffffff_0%,#f3fff9_100%)] p-4 text-center shadow-[0_12px_26px_rgba(4,120,87,0.10)]">
      <div className="flex items-center justify-between gap-2">
        <p className="text-[10px] font-black uppercase tracking-[0.18em] text-slate-950">{label}</p>
        <button type="button" onClick={clear} className="grid h-7 w-7 place-items-center rounded-full bg-rose-50 text-rose-500 transition hover:bg-rose-100" aria-label="Hapus tanda tangan">
          <RotateCcw className="h-3.5 w-3.5" />
        </button>
      </div>
      <div className="relative mt-3 overflow-hidden rounded-[22px] border border-slate-200 bg-[#fbfffd]">
        <canvas
          ref={canvasRef}
          className="h-40 w-full touch-none"
          onPointerDown={start}
          onPointerMove={move}
          onPointerUp={end}
          onPointerCancel={end}
          aria-label={`Area tanda tangan ${label.toLowerCase()}`}
        />
        {!hasInk ? <span className="pointer-events-none absolute inset-0 grid place-items-center text-[10px] font-semibold text-slate-400">Tanda tangan</span> : null}
      </div>
      <p className="mt-3 truncate text-[11px] font-black text-slate-500">{signerName}</p>
      <p className={hasInk ? "mt-1 text-[9px] font-black text-[#087e8b]" : "mt-1 text-[9px] font-black text-slate-400"}>
        {hasInk ? "Siap dikirim" : "Belum tanda tangan"}
      </p>
    </div>
  );
}

export function MasterAgentCreditDecisionControls({
  applicationId,
  requestedAmount,
  approvedAmount = 0,
  marketingNote = "",
  analystNote = "",
  analystRecommendation = "",
  analystRecommendedAmount = 0,
  status,
  loanStatus = "",
  mode = "master",
  canApprove = true,
  approveBlockReason = "",
}: Props) {
  const router = useRouter();
  const [busy, setBusy] = useState<DecisionAction | "">("");
  const [error, setError] = useState("");
  const [editing, setEditing] = useState(false);
  const [amount, setAmount] = useState(String(approvedAmount || analystRecommendedAmount || requestedAmount));
  const [note, setNote] = useState(mode === "analyst" || mode === "admin" ? analystNote : marketingNote || analystNote);
  const [signatureData, setSignatureData] = useState("");
  const [riskLevel, setRiskLevel] = useState("perhatian");
  const [riskScore, setRiskScore] = useState("50");
  const [statusEditorOpen, setStatusEditorOpen] = useState(false);
  const [statusReason, setStatusReason] = useState("");
  const [statusBusy, setStatusBusy] = useState(false);
  const [revisionDocuments, setRevisionDocuments] = useState<string[]>([]);
  const isFinal = status === "approved" || status === "rejected" || status === "analysis_rejected" || status === "master_rejected";
  const isRevisionWaiting = mode === "analyst" && status === "marketing_review" && analystRecommendation === "revision_required";
  const isMarketingReview = mode === "master" && (status === "submitted" || status === "marketing_review");
  const isAdminReview = mode === "admin" && ["submitted", "marketing_review", "analysis_review", "master_review", "ready_to_disburse"].includes(status);
  const canAct = isMarketingReview || isAdminReview || (mode === "analyst" && ["submitted", "marketing_review", "analysis_review", "master_review", "ready_to_disburse"].includes(status)) || isFinal;
  const needsReviewerSignature = isMarketingReview;
  const approveLabel = isMarketingReview ? "Kirim ke Operator" : mode === "analyst" || mode === "admin" ? "Setujui Pengajuan" : "Kirim ke Operator";
  const rejectLabel = "Tolak";
  const creditSuspended = loanStatus === "suspended";

  async function changeCreditOperationalStatus() {
    if (statusBusy) return;
    if (!statusReason.trim()) {
      setError("Alasan perubahan status kredit wajib diisi");
      return;
    }
    setStatusBusy(true);
    setError("");
    try {
      const token = window.localStorage.getItem("auth_token") || "";
      const response = await fetch("/api/admin/agent-credit/loans/status", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify({
          application_id: applicationId,
          suspended: !creditSuspended,
          reason: statusReason.trim(),
        }),
      });
      const body = (await response.json().catch(() => ({}))) as ApiBody;
      if (!response.ok || !body.ok) throw new Error(body.error || "Status kredit gagal diubah");
      setStatusEditorOpen(false);
      setStatusReason("");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Status kredit gagal diubah");
    } finally {
      setStatusBusy(false);
    }
  }

  async function decide(decision: DecisionAction) {
    if (busy) return;
    const isPositiveDecision = decision === "approved" || decision === "forward_to_analysis";
    if (isPositiveDecision && !canApprove) {
      setError(approveBlockReason || "Data agent wajib lengkap sebelum dikirim ke operator");
      return;
    }
    if (needsReviewerSignature && !signatureData.startsWith("data:image/")) {
      setError("Tanda tangan pemeriksa wajib diisi");
      return;
    }
    if (decision === "revision_required") {
      if (!revisionDocuments.length) {
        setError("Pilih minimal satu dokumen yang perlu diperbaiki");
        return;
      }
    }
    const parsedAmount = Number(amount.replace(/[^\d]/g, ""));
    const defaultNote = isPositiveDecision
        ? mode === "analyst" || mode === "admin"
          ? "Pemeriksaan risiko disetujui. Pinjaman saldo agent aktif."
          : mode === "marketing"
            ? "Data agent sudah diperiksa dan dikirim ke operator."
            : isMarketingReview
              ? "Data agent sudah diperiksa dan dikirim ke operator."
            : "Data agent sesuai dan dikirim ke operator."
        : "Data agent belum sesuai.";
    const apiDecision = isMarketingReview && decision === "forward_to_analysis" ? "kirim_analis" : decision;
    setBusy(decision);
    setError("");
    try {
      const token = window.localStorage.getItem("auth_token") || "";
      const headers: Record<string, string> = { "Content-Type": "application/json" };
      if (token) headers.Authorization = `Bearer ${token}`;
      const response = await fetch("/api/agent-credit/applications/decision", {
        method: "POST",
        headers,
        body: JSON.stringify({
          id: applicationId,
          decision: apiDecision,
          approved_amount: decision === "approved" && !isMarketingReview ? parsedAmount : 0,
          note: note.trim() || (decision === "revision_required" ? "Operator meminta perbaikan pada dokumen yang dipilih." : defaultNote),
          reviewer_mode: mode,
          signature_data: needsReviewerSignature ? signatureData : "",
          risk_level: mode === "analyst" || mode === "admin" ? riskLevel : undefined,
          risk_score: mode === "analyst" || mode === "admin" ? Number(riskScore || 0) : undefined,
          revision_documents: decision === "revision_required" ? revisionDocuments : undefined,
        }),
      });
      const body = (await response.json().catch(() => ({}))) as ApiBody;
      if (!response.ok || !body.ok) {
        throw new Error(body.error || "Keputusan gagal disimpan");
      }
      setEditing(false);
      if (decision === "revision_required") {
        window.location.reload();
        return;
      }
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Keputusan gagal disimpan");
    } finally {
      setBusy("");
    }
  }

  if (isFinal && !editing) {
    return (
      <div className="space-y-2">
        <div className={creditSuspended ? "rounded-2xl border border-amber-300 bg-amber-50 p-2.5 text-amber-800" : status === "approved" ? "rounded-2xl border border-emerald-200 bg-emerald-50 p-2.5 text-emerald-700" : "rounded-2xl border border-rose-200 bg-rose-50 p-2.5 text-rose-600"}>
          <div className="flex items-center gap-2">
            <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-white shadow-sm">
              {creditSuspended ? <LockKeyhole className="h-5 w-5" /> : status === "approved" ? <CheckCircle2 className="h-5 w-5" /> : <XCircle className="h-5 w-5" />}
            </span>
            <span className="min-w-0 flex-1">
              <span className="block text-sm font-black">{creditSuspended ? "Kredit Dibekukan Super Admin" : status === "approved" ? "Disetujui Operator" : "Ditolak Operator"}</span>
              <span className="mt-0.5 block truncate text-[10px] font-bold opacity-70">
                {creditSuspended ? "Kredit agent sedang dibekukan dan pengajuan baru tidak dapat dibuat." : status === "approved" ? "Dana modal sudah masuk ke saldo utama agent dan dapat digunakan selama kemitraan aktif." : "Agent akan melihat pemberitahuan dan bisa memperbaiki data."}
              </span>
            </span>
          </div>
          {marketingNote || analystNote ? <p className="mt-2 rounded-2xl bg-white/70 px-3 py-2 text-[10px] font-bold leading-4 opacity-80">{marketingNote || analystNote}</p> : null}
        </div>
        {mode === "admin" && status === "approved" && ["active", "due", "overdue", "suspended"].includes(loanStatus) ? (
          <div className="rounded-2xl border border-slate-200 bg-white p-3">
            {!statusEditorOpen ? (
              <button
                type="button"
                onClick={() => setStatusEditorOpen(true)}
                className={creditSuspended ? "inline-flex h-10 w-full items-center justify-center gap-2 rounded-xl bg-emerald-700 px-3 text-[11px] font-black text-white" : "inline-flex h-10 w-full items-center justify-center gap-2 rounded-xl border border-amber-300 bg-amber-50 px-3 text-[11px] font-black text-amber-800"}
              >
                {creditSuspended ? <UnlockKeyhole className="h-4 w-4" /> : <LockKeyhole className="h-4 w-4" />}
                {creditSuspended ? "Aktifkan Kembali Kredit" : "Bekukan Kredit Agent"}
              </button>
            ) : (
              <div className="space-y-2">
                <label className="block">
                  <span className="text-[9px] font-black uppercase tracking-[0.12em] text-slate-500">Alasan Super Admin</span>
                  <textarea
                    value={statusReason}
                    onChange={(event) => setStatusReason(event.target.value)}
                    rows={2}
                    placeholder={creditSuspended ? "Alasan mengaktifkan kembali kredit" : "Contoh: pemeriksaan risiko atau tunggakan"}
                    className="mt-1 w-full resize-none rounded-xl border border-slate-200 px-3 py-2 text-xs font-semibold outline-none focus:border-emerald-400"
                  />
                </label>
                <div className="grid grid-cols-2 gap-2">
                  <button type="button" onClick={() => { setStatusEditorOpen(false); setStatusReason(""); setError(""); }} disabled={statusBusy} className="h-10 rounded-xl bg-slate-100 text-[10px] font-black text-slate-600">Batal</button>
                  <button type="button" onClick={() => void changeCreditOperationalStatus()} disabled={statusBusy} className="inline-flex h-10 items-center justify-center gap-2 rounded-xl bg-[#13464b] text-[10px] font-black text-white disabled:opacity-60">
                    {statusBusy ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
                    Simpan Status
                  </button>
                </div>
              </div>
            )}
            {error ? <p className="mt-2 rounded-xl bg-rose-50 px-3 py-2 text-center text-[10px] font-black text-rose-600">{error}</p> : null}
          </div>
        ) : null}
      </div>
    );
  }

  if (isRevisionWaiting) {
    return (
      <div className="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-amber-800">
        <div className="flex items-start gap-2">
          <FileWarning className="mt-0.5 h-4 w-4 shrink-0" />
          <div>
            <p className="text-xs font-black">Menunggu perbaikan agent</p>
            <p className="mt-1 text-[10px] font-semibold leading-4">Permintaan perbaikan sudah dikirim. Agent hanya perlu mengganti dokumen yang dipilih.</p>
            {analystNote ? <p className="mt-2 rounded-xl bg-white/70 px-3 py-2 text-[10px] font-bold">{analystNote}</p> : null}
          </div>
        </div>
      </div>
    );
  }

  if (!canAct && !editing) {
    return (
      <div className="rounded-2xl border border-slate-200 bg-slate-50 px-3 py-2 text-slate-600">
        <div>
          <p className="text-xs font-black">Menunggu review</p>
          <p className="mt-0.5 text-[10px] font-bold text-slate-400">Belum bisa diputuskan di tahap ini.</p>
        </div>
        {analystRecommendation ? (
          <p className="mt-3 rounded-2xl bg-white px-3 py-2 text-[10px] font-bold text-slate-500">
            Catatan lama: {analystRecommendation === "approved" ? "Layak" : "Tidak layak"} {analystRecommendedAmount ? `- Rp ${new Intl.NumberFormat("id-ID").format(analystRecommendedAmount)}` : ""}
          </p>
        ) : null}
      </div>
    );
  }

  return (
    <div className="rounded-2xl border border-emerald-100 bg-[linear-gradient(135deg,#ffffff_0%,#f8fffb_58%,#ecfdf5_100%)] p-3 shadow-[0_10px_24px_rgba(6,78,59,0.05)]">
      <div className="grid gap-2 xl:grid-cols-[160px_minmax(260px,1fr)_210px] xl:items-stretch">
        <label className="block min-w-0 rounded-xl border border-slate-200 bg-white px-3 py-2 shadow-[0_8px_18px_rgba(15,23,42,0.035)] focus-within:border-emerald-300 focus-within:ring-2 focus-within:ring-emerald-100">
          <span className="block truncate text-[9px] font-black uppercase tracking-[0.08em] text-emerald-600">{isMarketingReview ? "Nominal Diajukan" : "Nominal ACC"}</span>
          <input
            value={amount}
            onChange={(event) => setAmount(event.target.value)}
            inputMode="numeric"
            disabled={isMarketingReview}
            className="mt-1 h-8 w-full bg-transparent text-base font-black text-slate-950 outline-none disabled:text-slate-700"
          />
        </label>
        <label className="block min-w-0 rounded-xl border border-slate-200 bg-white px-3 py-2 shadow-[0_8px_18px_rgba(15,23,42,0.035)] focus-within:border-emerald-300 focus-within:ring-2 focus-within:ring-emerald-100">
          <span className="block truncate text-[9px] font-black uppercase tracking-[0.08em] text-emerald-600">Catatan</span>
          <textarea
            value={note}
            onChange={(event) => setNote(event.target.value)}
            placeholder={isMarketingReview ? "Catatan hasil pendampingan lapangan" : "Catatan untuk agent"}
            rows={2}
            className="mt-1 min-h-10 w-full resize-none bg-transparent text-xs font-bold leading-5 text-slate-700 outline-none placeholder:text-slate-400"
          />
        </label>
        {mode === "analyst" || mode === "admin" ? (
          <div className="grid gap-2 rounded-xl border border-slate-200 bg-white p-2 shadow-[0_8px_18px_rgba(15,23,42,0.035)] min-[420px]:grid-cols-2 xl:col-span-2">
            <label className="block">
              <span className="block text-[9px] font-black uppercase tracking-[0.08em] text-emerald-600">Level Risiko</span>
              <select value={riskLevel} onChange={(event) => setRiskLevel(event.target.value)} className="mt-1 h-10 w-full rounded-xl border border-slate-200 bg-slate-50 px-3 text-xs font-black text-slate-950 outline-none">
                <option value="aman">Aman</option>
                <option value="perhatian">Perlu Perhatian</option>
                <option value="tinggi">Risiko Tinggi</option>
              </select>
            </label>
            <label className="block">
              <span className="block text-[9px] font-black uppercase tracking-[0.08em] text-emerald-600">Skor Risiko</span>
              <input value={riskScore} onChange={(event) => setRiskScore(event.target.value)} inputMode="numeric" className="mt-1 h-10 w-full rounded-xl border border-slate-200 bg-slate-50 px-3 text-xs font-black text-slate-950 outline-none" />
            </label>
          </div>
        ) : null}
        {mode === "analyst" && !isFinal ? (
          <div className="xl:col-span-2 rounded-xl border border-amber-200 bg-amber-50/70 p-3">
            <div className="flex items-start gap-2">
              <FileWarning className="mt-0.5 h-4 w-4 shrink-0 text-amber-700" />
              <div>
                <p className="text-[10px] font-black uppercase tracking-[0.08em] text-amber-800">Perlu perbaikan dokumen</p>
                <p className="mt-1 text-[10px] font-semibold leading-4 text-amber-700">Pilih hanya foto yang blur atau tidak jelas. Agent akan mengunggah ulang bagian ini tanpa mengulang data pinjaman.</p>
              </div>
            </div>
            <div className="mt-2 grid grid-cols-1 gap-2 min-[420px]:grid-cols-2">
              {revisionOptions.map((option) => {
                const checked = revisionDocuments.includes(option.key);
                return (
                  <label key={option.key} className={`flex cursor-pointer items-center gap-2 rounded-xl border px-3 py-2 text-[10px] font-black transition ${checked ? "border-amber-400 bg-white text-amber-800" : "border-amber-100 bg-white/60 text-slate-600"}`}>
                    <input
                      type="checkbox"
                      checked={checked}
                      onChange={(event) => setRevisionDocuments((current) => event.target.checked ? [...current, option.key] : current.filter((key) => key !== option.key))}
                      className="h-4 w-4 accent-amber-600"
                    />
                    {option.label}
                  </label>
                );
              })}
            </div>
          </div>
        ) : null}
        {needsReviewerSignature ? (
          <div className="xl:col-span-2">
            <ReviewSignaturePad
              label="Marketing"
              signerName="Marketing Isiloka"
              value={signatureData}
              onChange={setSignatureData}
            />
          </div>
        ) : null}
        <div className={`grid min-w-0 grid-cols-1 gap-2 ${isMarketingReview ? "" : mode === "analyst" ? "min-[390px]:grid-cols-3 xl:grid-cols-1" : "min-[390px]:grid-cols-2 xl:grid-cols-1"}`}>
          <button
            type="button"
            onClick={() => decide(isMarketingReview ? "forward_to_analysis" : "approved")}
            disabled={Boolean(busy) || !canApprove}
            className="inline-flex h-11 min-w-0 items-center justify-center gap-2 rounded-xl bg-[linear-gradient(135deg,#087e8b,#16a34a)] px-3 text-[11px] font-black leading-3 text-white shadow-[0_10px_18px_rgba(5,150,105,0.18)] transition hover:-translate-y-0.5 disabled:translate-y-0 disabled:bg-none disabled:bg-slate-200 disabled:text-slate-500 disabled:shadow-none"
            title={!canApprove ? approveBlockReason || "Agent belum melengkapi persetujuan" : undefined}
          >
            {busy === "approved" || busy === "forward_to_analysis" ? <Loader2 className="h-4 w-4 shrink-0 animate-spin" /> : <CheckCircle2 className="h-4 w-4 shrink-0" />}
            <span className="truncate">{busy === "approved" || busy === "forward_to_analysis" ? "Proses" : approveLabel}</span>
          </button>
          {mode === "analyst" || mode === "admin" ? (
            <button
              type="button"
              onClick={() => decide("rejected")}
              disabled={Boolean(busy)}
              className="inline-flex h-11 min-w-0 items-center justify-center gap-2 rounded-xl border border-rose-200 bg-white px-3 text-[11px] font-black leading-3 text-rose-600 shadow-[0_8px_16px_rgba(225,29,72,0.06)] transition hover:-translate-y-0.5 hover:bg-rose-50 disabled:translate-y-0 disabled:opacity-60"
            >
              {busy === "rejected" ? <Loader2 className="h-4 w-4 shrink-0 animate-spin" /> : <XCircle className="h-4 w-4 shrink-0" />}
              <span className="truncate">{busy === "rejected" ? "Proses" : rejectLabel}</span>
            </button>
          ) : null}
          {mode === "analyst" ? (
            <button
              type="button"
              onClick={() => decide("revision_required")}
              disabled={Boolean(busy)}
              className="inline-flex h-11 min-w-0 items-center justify-center gap-2 rounded-xl border border-amber-300 bg-amber-50 px-3 text-[11px] font-black leading-3 text-amber-800 shadow-[0_8px_16px_rgba(245,158,11,0.08)] transition hover:-translate-y-0.5 hover:bg-amber-100 disabled:translate-y-0 disabled:opacity-60"
            >
              {busy === "revision_required" ? <Loader2 className="h-4 w-4 shrink-0 animate-spin" /> : <FileWarning className="h-4 w-4 shrink-0" />}
              <span className="truncate">{busy === "revision_required" ? "Proses" : "Minta Perbaikan"}</span>
            </button>
          ) : null}
        </div>
      </div>
      {editing ? (
        <button
          type="button"
          onClick={() => {
            setEditing(false);
            setError("");
          }}
          className="mt-3 h-10 w-full rounded-2xl bg-slate-100 text-[10px] font-black text-slate-500 transition hover:bg-slate-200"
        >
          Batal edit
        </button>
      ) : null}
      {!canApprove && !(mode === "analyst" && !isFinal) ? <p className="mt-2 rounded-2xl bg-amber-50 px-3 py-2 text-center text-[10px] font-black text-amber-700">{approveBlockReason || "Data agent wajib lengkap sebelum dikirim ke operator."}</p> : null}
      {error ? <p className="mt-2 rounded-2xl bg-rose-50 px-3 py-2 text-center text-[10px] font-black text-rose-600">{error}</p> : null}
    </div>
  );
}
