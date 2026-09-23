"use client";

import { useMemo, useState } from "react";
import { FileDown, Printer, Search, Sheet, SlidersHorizontal } from "lucide-react";
import type { AgentCreditApplication } from "@/lib/api.auth";

type ReportTab = "agents" | "applications" | "survey" | "credit";
type ReportRow = Record<string, string | number>;

type Props = {
  applications: AgentCreditApplication[];
};

const tabs: Array<{ value: ReportTab; label: string }> = [
  { value: "agents", label: "Agent Binaan" },
  { value: "applications", label: "Pengajuan Kredit" },
  { value: "survey", label: "Survei & Dokumen" },
  { value: "credit", label: "Modal & Aktivitas" },
];

function formatIDR(value: number) {
  return `Rp ${new Intl.NumberFormat("id-ID").format(Number(value || 0))}`;
}

function formatDate(value?: string) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return new Intl.DateTimeFormat("id-ID", { day: "2-digit", month: "short", year: "numeric" }).format(date);
}

function applicantValue(item: AgentCreditApplication, key: string, fallback = "-") {
  const value = item.applicant_data?.[key];
  return typeof value === "string" && value.trim() ? value.trim() : fallback;
}

function agentName(item: AgentCreditApplication) {
  return applicantValue(item, "agent_name", item.member_name || item.agent_name || "Agent Isiloka");
}

function storeName(item: AgentCreditApplication) {
  return applicantValue(item, "store_name", item.store_name || "-");
}

function statusLabel(item: AgentCreditApplication) {
  const status = String(item.status || "").toLowerCase();
  const loanStatus = String(item.loan_status || "").toLowerCase();
  if (status === "approved" && loanStatus === "paid") return "Siklus selesai";
  if (status === "approved" && loanStatus === "overdue") return "Menunggak";
  if (status === "approved") return "Kredit Aktif";
  if (status === "analysis_review" || status === "master_review") return "Di Operator";
  if (status === "submitted" || status === "marketing_review") return "Perlu Survei";
  if (status.includes("reject")) return "Ditolak";
  return status || "-";
}

function documentExists(item: AgentCreditApplication, key: string) {
  const value = item.document_data?.[key] || (key === "selfie_marketing" ? item.document_data?.selfie : undefined);
  if (!value || typeof value !== "object") return false;
  return String((value as { data_url?: unknown }).data_url || "").startsWith("data:image/");
}

function storedSignature(item: AgentCreditApplication, role: "agent" | "marketing" | "operator") {
  const value = role === "agent"
    ? item.agent_signature_data
    : role === "marketing"
      ? item.applicant_data?.marketing_signature_data
      : item.applicant_data?.master_signature_data;
  return typeof value === "string" && value.startsWith("data:image/") ? value : "";
}

function signatureDate(item: AgentCreditApplication, role: "agent" | "marketing" | "operator") {
  const value = role === "agent"
    ? item.agent_signature_at
    : role === "marketing"
      ? item.applicant_data?.marketing_signature_at
      : item.applicant_data?.master_signature_at;
  return typeof value === "string" ? formatDate(value) : "-";
}

function imageTypeFromDataUrl(value: string) {
  return value.startsWith("data:image/png") ? "PNG" : "JPEG";
}

function matchesStatus(item: AgentCreditApplication, filter: string) {
  if (filter === "all") return true;
  const label = statusLabel(item);
  if (filter === "survey") return label === "Perlu Survei";
  if (filter === "operator") return label === "Di Operator";
  if (filter === "active") return label === "Kredit Aktif" || label === "Menunggak";
  if (filter === "paid") return label === "Siklus selesai";
  if (filter === "rejected") return label === "Ditolak";
  return true;
}

function escapeHtml(value: string | number) {
  return String(value).replace(/[&<>'"]/g, (character) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "'": "&#39;", '"': "&quot;" })[character] || character);
}

export function MarketingReportCenter({ applications }: Props) {
  const [tab, setTab] = useState<ReportTab>("agents");
  const [query, setQuery] = useState("");
  const [status, setStatus] = useState("all");
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");
  const [exporting, setExporting] = useState(false);

  const filteredApplications = useMemo(() => {
    const needle = query.trim().toLowerCase();
    return applications.filter((item) => {
      const itemDate = String(item.created_at || item.updated_at || "").slice(0, 10);
      const matchesDate = (!dateFrom || itemDate >= dateFrom) && (!dateTo || itemDate <= dateTo);
      const haystack = [agentName(item), storeName(item), item.member_phone, item.member_email, item.member_id].join(" ").toLowerCase();
      return matchesDate && matchesStatus(item, status) && (!needle || haystack.includes(needle));
    });
  }, [applications, dateFrom, dateTo, query, status]);

  const rows = useMemo<ReportRow[]>(() => {
    if (tab === "agents") {
      const latest = new Map<number, AgentCreditApplication>();
      filteredApplications.forEach((item) => {
        const current = latest.get(item.member_id);
        if (!current || new Date(item.updated_at || item.created_at).getTime() > new Date(current.updated_at || current.created_at).getTime()) latest.set(item.member_id, item);
      });
      return Array.from(latest.values()).map((item) => ({
        "ID Agent": item.member_id,
        Agent: agentName(item),
        Toko: storeName(item),
        WhatsApp: applicantValue(item, "whatsapp", item.member_phone || "-"),
        Status: statusLabel(item),
        Tier: item.credit_level_name || "Kilat Start",
        Limit: formatIDR(Number(item.credit_limit_amount || item.approved_amount || 0)),
        "Modal Berjalan": formatIDR(Number(item.outstanding_amount || 0)),
      }));
    }
    if (tab === "survey") {
      return filteredApplications.map((item) => ({
        Tanggal: formatDate(item.created_at),
        Agent: agentName(item),
        Toko: storeName(item),
        KTP: documentExists(item, "ktp") ? "Ada" : "Kosong",
        "Foto Toko": documentExists(item, "store") ? "Ada" : "Kosong",
        "Selfie KTP": documentExists(item, "selfie_ktp") ? "Ada" : "Kosong",
        "Bersama Marketing": documentExists(item, "selfie_marketing") ? "Ada" : "Kosong",
        Status: statusLabel(item),
      }));
    }
    if (tab === "credit") {
      return filteredApplications.filter((item) => item.status === "approved").map((item) => ({
        Tanggal: formatDate(item.loan_approved_at || item.updated_at),
        Agent: agentName(item),
        Tier: item.credit_level_name || "Kilat Start",
        Limit: formatIDR(Number(item.credit_limit_amount || item.approved_amount || 0)),
        "Saldo Kredit": formatIDR(Number(item.credit_available_amount || 0)),
        "Modal Berjalan": formatIDR(Number(item.outstanding_amount || 0)),
        Terbayar: formatIDR(Number(item.paid_amount || 0)),
        Status: statusLabel(item),
      }));
    }
    return filteredApplications.map((item) => ({
      Tanggal: formatDate(item.created_at),
      "ID Pengajuan": item.id,
      Agent: agentName(item),
      Toko: storeName(item),
      Diajukan: formatIDR(Number(item.requested_amount || 0)),
      Disetujui: formatIDR(Number(item.approved_amount || 0)),
      Status: statusLabel(item),
    }));
  }, [filteredApplications, tab]);

  const reportTitle = tabs.find((item) => item.value === tab)?.label || "Laporan";
  const periodLabel = dateFrom || dateTo ? `${dateFrom || "Awal"} s.d. ${dateTo || "Sekarang"}` : "Semua periode";
  const totalRequested = filteredApplications.reduce((sum, item) => sum + Number(item.requested_amount || 0), 0);
  const totalOutstanding = filteredApplications.reduce((sum, item) => sum + Number(item.outstanding_amount || 0), 0);
  const signatureApplications = useMemo(() => {
    const scoped = tab === "credit" ? filteredApplications.filter((item) => item.status === "approved") : filteredApplications;
    if (tab !== "agents") return scoped;
    const latest = new Map<number, AgentCreditApplication>();
    scoped.forEach((item) => {
      const current = latest.get(item.member_id);
      if (!current || new Date(item.updated_at || item.created_at).getTime() > new Date(current.updated_at || current.created_at).getTime()) latest.set(item.member_id, item);
    });
    return Array.from(latest.values());
  }, [filteredApplications, tab]);

  async function downloadExcel() {
    if (!rows.length) return;
    setExporting(true);
    try {
      const XLSX = await import("xlsx");
      const sheet = XLSX.utils.json_to_sheet(rows);
      const book = XLSX.utils.book_new();
      XLSX.utils.book_append_sheet(book, sheet, reportTitle.slice(0, 31));
      const output = XLSX.write(book, { bookType: "xlsx", type: "array" });
      const url = URL.createObjectURL(new Blob([output], { type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" }));
      const anchor = document.createElement("a");
      anchor.href = url;
      anchor.download = `laporan-marketing-${tab}.xlsx`;
      anchor.click();
      URL.revokeObjectURL(url);
    } finally {
      setExporting(false);
    }
  }

  async function downloadPdf() {
    if (!rows.length) return;
    setExporting(true);
    try {
      const [{ jsPDF }, autoTableModule] = await Promise.all([import("jspdf"), import("jspdf-autotable")]);
      const doc = new jsPDF({ orientation: "landscape", unit: "pt", format: "a4" });
      const headers = Object.keys(rows[0]);
      doc.setFontSize(16);
      doc.text(`Isiloka - ${reportTitle}`, 32, 34);
      doc.setFontSize(9);
      doc.text(`Periode: ${periodLabel} | Dicetak: ${new Date().toLocaleString("id-ID")}`, 32, 50);
      autoTableModule.default(doc, {
        head: [headers],
        body: rows.map((row) => headers.map((header) => String(row[header] ?? "-"))),
        startY: 62,
        margin: { left: 32, right: 32, bottom: 175 },
        styles: { fontSize: 7, cellPadding: 4 },
        headStyles: { fillColor: [4, 120, 87], textColor: 255 },
      });

      const approvalItem = signatureApplications[0];
      if (approvalItem) {
        const pageWidth = doc.internal.pageSize.getWidth();
        const tableEndY = (doc as typeof doc & { lastAutoTable?: { finalY?: number } }).lastAutoTable?.finalY || 210;
        const approvalY = tableEndY + 18;
        doc.setTextColor(4, 120, 87);
        doc.setFont("helvetica", "bold");
        doc.setFontSize(12);
        doc.text("Persetujuan Laporan", 36, approvalY);
        doc.setFont("helvetica", "normal");
        doc.setTextColor(100, 116, 139);
        doc.setFontSize(8);
        doc.text("Ditandatangani secara elektronik sesuai tahapan verifikasi kredit.", 36, approvalY + 13);

        const signers = [
          { role: "agent" as const, label: "Agent", name: agentName(approvalItem) },
          { role: "marketing" as const, label: "Marketing", name: "Petugas Marketing" },
          { role: "operator" as const, label: "Operator Kredit", name: "Operator Kredit" },
        ];
        const cardWidth = 235;
        const gap = 22;
        const startX = (pageWidth - (cardWidth * 3 + gap * 2)) / 2;
        signers.forEach((signer, index) => {
          const x = startX + index * (cardWidth + gap);
          const signature = storedSignature(approvalItem, signer.role);
          doc.setDrawColor(203, 213, 225);
          doc.setFillColor(255, 255, 255);
          doc.roundedRect(x, approvalY + 25, cardWidth, 118, 5, 5, "FD");
          doc.setTextColor(4, 120, 87);
          doc.setFont("helvetica", "bold");
          doc.setFontSize(9);
          doc.text(signer.label, x + cardWidth / 2, approvalY + 42, { align: "center" });
          if (signature) {
            try {
              doc.addImage(signature, imageTypeFromDataUrl(signature), x + 62, approvalY + 48, cardWidth - 124, 42, undefined, "FAST");
            } catch {
              doc.setTextColor(148, 163, 184);
              doc.setFontSize(7);
              doc.text("Tanda tangan tidak dapat dibaca", x + cardWidth / 2, approvalY + 72, { align: "center" });
            }
          } else {
            doc.setTextColor(148, 163, 184);
            doc.setFont("helvetica", "normal");
            doc.setFontSize(7);
            doc.text("Menunggu tanda tangan", x + cardWidth / 2, approvalY + 72, { align: "center" });
          }
          doc.setDrawColor(100, 116, 139);
          doc.line(x + 32, approvalY + 96, x + cardWidth - 32, approvalY + 96);
          doc.setTextColor(15, 23, 42);
          doc.setFont("helvetica", "bold");
          doc.setFontSize(8);
          doc.text(signer.name, x + cardWidth / 2, approvalY + 108, { align: "center", maxWidth: cardWidth - 30 });
          doc.setTextColor(100, 116, 139);
          doc.setFont("helvetica", "normal");
          doc.setFontSize(7);
          doc.text(signatureDate(approvalItem, signer.role), x + cardWidth / 2, approvalY + 122, { align: "center" });
        });
      }
      doc.save(`laporan-marketing-${tab}.pdf`);
    } finally {
      setExporting(false);
    }
  }

  function printReport() {
    if (!rows.length) return;
    const headers = Object.keys(rows[0]);
    const printWindow = window.open("", "_blank", "width=1200,height=800");
    if (!printWindow) return;
    const approvalItem = signatureApplications[0];
    const approvalSection = approvalItem ? (() => {
      const signers = [
        { role: "agent" as const, label: "Agent", name: agentName(approvalItem) },
        { role: "marketing" as const, label: "Marketing", name: "Petugas Marketing" },
        { role: "operator" as const, label: "Operator Kredit", name: "Operator Kredit" },
      ];
      return `<section class="approval"><h2>Persetujuan Laporan</h2><p>Ditandatangani secara elektronik sesuai tahapan verifikasi kredit.</p><div class="signatures">${signers.map((signer) => { const signature = storedSignature(approvalItem, signer.role); return `<div class="signature-card"><strong>${escapeHtml(signer.label)}</strong><div class="signature-image">${signature ? `<img src="${signature}" alt="Tanda tangan ${escapeHtml(signer.label)}">` : `<span>Menunggu tanda tangan</span>`}</div><div class="line">${escapeHtml(signer.name)}</div><small>${escapeHtml(signatureDate(approvalItem, signer.role))}</small></div>`; }).join("")}</div></section>`;
    })() : "";
    printWindow.document.write(`<!doctype html><html><head><title>${escapeHtml(reportTitle)}</title><style>body{font-family:Arial,sans-serif;color:#111827;margin:28px}header{border-bottom:3px solid #087e8b;padding-bottom:12px;margin-bottom:18px}h1{font-size:22px;margin:0;color:#065f46}h2{font-size:15px;color:#065f46;margin:0 0 4px}p{font-size:11px;margin:5px 0;color:#475569}table{width:100%;border-collapse:collapse;font-size:9px}th{background:#087e8b;color:#fff;text-align:left}th,td{border:1px solid #cbd5e1;padding:6px;vertical-align:top}.approval{margin-top:16px;break-inside:avoid;page-break-inside:avoid}.signatures{display:grid;grid-template-columns:repeat(3,1fr);gap:18px;margin-top:14px;text-align:center}.signature-card{border:1px solid #cbd5e1;border-radius:6px;padding:10px}.signature-card strong{color:#087e8b;font-size:11px}.signature-image{height:58px;display:flex;align-items:center;justify-content:center;color:#94a3b8;font-size:9px}.signature-image img{max-width:120px;max-height:52px;object-fit:contain}.line{border-top:1px solid #334155;padding-top:5px;font-size:10px;font-weight:700}.signature-card small{display:block;margin-top:3px;color:#64748b;font-size:8px}@page{size:landscape;margin:14mm}</style></head><body><header><h1>Isiloka - ${escapeHtml(reportTitle)}</h1><p>Periode: ${escapeHtml(periodLabel)} | Dibuat: ${escapeHtml(new Date().toLocaleString("id-ID"))}</p><p>Total data: ${rows.length} | Total pengajuan: ${escapeHtml(formatIDR(totalRequested))} | Total tagihan: ${escapeHtml(formatIDR(totalOutstanding))}</p></header><table><thead><tr>${headers.map((header) => `<th>${escapeHtml(header)}</th>`).join("")}</tr></thead><tbody>${rows.map((row) => `<tr>${headers.map((header) => `<td>${escapeHtml(row[header] ?? "-")}</td>`).join("")}</tr>`).join("")}</tbody></table>${approvalSection}<script>window.onload=()=>window.print();</script></body></html>`);
    printWindow.document.close();
  }

  return (
    <section className="mx-auto w-full max-w-7xl overflow-hidden rounded-lg border border-emerald-200 bg-white shadow-[0_18px_42px_rgba(6,78,59,0.08)]">
      <header className="grid gap-4 bg-[linear-gradient(135deg,#13464b,#087e8b)] px-4 py-5 text-white sm:px-6 sm:py-6">
        <div className="min-w-0"><p className="text-[9px] font-black uppercase tracking-[0.18em] text-lime-200">Administrasi Marketing</p><h1 className="mt-1 text-2xl font-black leading-tight sm:text-3xl">Laporan Agent Binaan</h1><p className="mt-2 max-w-2xl text-xs font-semibold leading-5 text-emerald-100/75">Rekap onboarding, survei, dan aktivitas agent binaan.</p></div>
        <div className="flex w-full flex-wrap gap-2 sm:w-auto">
          <button type="button" onClick={printReport} disabled={!rows.length || exporting} className="inline-flex h-10 min-w-[92px] flex-1 items-center justify-center gap-2 rounded-lg border border-white/20 bg-white/10 px-3 text-xs font-black disabled:opacity-40 sm:flex-none"><Printer className="h-4 w-4" />Cetak</button>
          <button type="button" onClick={() => void downloadPdf()} disabled={!rows.length || exporting} className="inline-flex h-10 min-w-[92px] flex-1 items-center justify-center gap-2 rounded-lg bg-white px-3 text-xs font-black text-emerald-800 disabled:opacity-40 sm:flex-none"><FileDown className="h-4 w-4" />PDF</button>
          <button type="button" onClick={() => void downloadExcel()} disabled={!rows.length || exporting} className="inline-flex h-10 min-w-[92px] flex-1 items-center justify-center gap-2 rounded-lg bg-lime-300 px-3 text-xs font-black text-emerald-950 disabled:opacity-40 sm:flex-none"><Sheet className="h-4 w-4" />Excel</button>
        </div>
      </header>

      <div className="border-b border-emerald-100 bg-[#f8fffb] p-4">
        <div className="grid min-w-0 grid-cols-[repeat(auto-fit,minmax(min(100%,220px),1fr))] gap-3">
          <label className="flex h-11 items-center gap-2 rounded-lg border border-slate-200 bg-white px-3 text-sm text-slate-500"><Search className="h-4 w-4 text-emerald-600" /><input value={query} onChange={(event) => setQuery(event.target.value)} className="min-w-0 flex-1 bg-transparent font-semibold outline-none" placeholder="Cari agent, toko, WA, atau ID" /></label>
          <input type="date" value={dateFrom} onChange={(event) => setDateFrom(event.target.value)} aria-label="Tanggal awal" className="h-11 rounded-lg border border-slate-200 bg-white px-3 text-xs font-bold text-slate-600 outline-none focus:border-emerald-400" />
          <input type="date" value={dateTo} onChange={(event) => setDateTo(event.target.value)} aria-label="Tanggal akhir" className="h-11 rounded-lg border border-slate-200 bg-white px-3 text-xs font-bold text-slate-600 outline-none focus:border-emerald-400" />
          <label className="flex h-11 items-center gap-2 rounded-lg border border-slate-200 bg-white px-3"><SlidersHorizontal className="h-4 w-4 text-emerald-600" /><select value={status} onChange={(event) => setStatus(event.target.value)} className="min-w-0 flex-1 bg-transparent text-xs font-black text-slate-600 outline-none"><option value="all">Semua status</option><option value="survey">Perlu survei</option><option value="operator">Di operator</option><option value="active">Modal aktif</option><option value="paid">Siklus selesai</option><option value="rejected">Ditolak</option></select></label>
        </div>
      </div>

      <div className="flex min-w-0 gap-2 overflow-x-auto border-b border-slate-200 px-4 pt-3 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
        {tabs.map((item) => <button key={item.value} type="button" onClick={() => setTab(item.value)} className={tab === item.value ? "shrink-0 border-b-2 border-emerald-700 px-3 pb-3 text-xs font-black text-emerald-700" : "shrink-0 border-b-2 border-transparent px-3 pb-3 text-xs font-black text-slate-500 hover:text-emerald-700"}>{item.label}</button>)}
      </div>

      <div className="grid grid-cols-3 border-b border-slate-200 bg-white text-center"><div className="p-3"><p className="text-[9px] font-black uppercase text-slate-400">Data Tampil</p><p className="mt-1 text-lg font-black text-slate-950">{rows.length}</p></div><div className="border-x border-slate-100 p-3"><p className="text-[9px] font-black uppercase text-slate-400">Total Pengajuan</p><p className="mt-1 text-sm font-black text-emerald-700 sm:text-lg">{formatIDR(totalRequested)}</p></div><div className="p-3"><p className="text-[9px] font-black uppercase text-slate-400">Modal Berjalan</p><p className="mt-1 text-sm font-black text-amber-700 sm:text-lg">{formatIDR(totalOutstanding)}</p></div></div>

      <div className="overflow-x-auto p-4">
        {rows.length ? <table className="w-full min-w-[760px] border-separate border-spacing-0 overflow-hidden text-left text-xs"><thead><tr>{Object.keys(rows[0]).map((header) => <th key={header} className="border-y border-slate-200 bg-slate-50 px-3 py-3 text-[10px] font-black uppercase tracking-[0.08em] text-slate-500 first:border-l first:rounded-l-lg last:border-r last:rounded-r-lg">{header}</th>)}</tr></thead><tbody>{rows.map((row, index) => <tr key={`${tab}-${index}`} className="hover:bg-emerald-50/35">{Object.keys(rows[0]).map((header) => <td key={header} className="border-b border-slate-100 px-3 py-3 font-semibold text-slate-700">{row[header]}</td>)}</tr>)}</tbody></table> : <div className="grid min-h-36 place-items-center rounded-lg border border-dashed border-emerald-300 bg-emerald-50/35 px-4 text-center text-sm font-semibold text-slate-500">Tidak ada data laporan sesuai filter.</div>}
      </div>
    </section>
  );
}
