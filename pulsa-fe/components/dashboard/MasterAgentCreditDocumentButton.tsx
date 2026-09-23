"use client";

import { Download, Eye, FileText, X } from "lucide-react";
import Image from "next/image";
import { useEffect, useState } from "react";
import { createPortal } from "react-dom";

type CreditDocumentImage = {
  label: string;
  src: string;
};

export function MasterAgentCreditDocumentButton({
  agentName,
  documents,
}: {
  agentName: string;
  documents: CreditDocumentImage[];
}) {
  const readyDocuments = documents.filter((doc) => doc.src);
  const canPreview = readyDocuments.length > 0;
  const [open, setOpen] = useState(false);
  const [mounted, setMounted] = useState(false);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (!open) return;
    const originalBodyOverflow = document.body.style.overflow;
    const originalHtmlOverflow = document.documentElement.style.overflow;
    document.body.style.overflow = "hidden";
    document.documentElement.style.overflow = "hidden";

    function onKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") setOpen(false);
    }

    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.body.style.overflow = originalBodyOverflow;
      document.documentElement.style.overflow = originalHtmlOverflow;
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [open]);

  function imageTypeFromDataUrl(src: string) {
    return src.startsWith("data:image/png") ? "PNG" : "JPEG";
  }

  async function buildPDF() {
    const { jsPDF } = await import("jspdf");
    const pdf = new jsPDF({ unit: "pt", format: "a4" });
    const pageWidth = pdf.internal.pageSize.getWidth();
    const pageHeight = pdf.internal.pageSize.getHeight();
    const margin = 40;

    readyDocuments.forEach((doc, index) => {
      if (index > 0) pdf.addPage();
      pdf.setFillColor(5, 46, 38);
      pdf.rect(0, 0, pageWidth, 64, "F");
      pdf.setTextColor(255, 255, 255);
      pdf.setFontSize(16);
      pdf.text("Isiloka - Dokumen Kredit Agent", margin, 38);

      pdf.setTextColor(15, 23, 42);
      pdf.setFontSize(12);
      pdf.text(`${doc.label} - ${agentName}`, margin, 92);

      const maxWidth = pageWidth - margin * 2;
      const maxHeight = pageHeight - 142;
      pdf.addImage(doc.src, imageTypeFromDataUrl(doc.src), margin, 112, maxWidth, maxHeight, undefined, "FAST");
    });

    return pdf;
  }

  async function openPreview() {
    if (!canPreview || loading) return;
    setOpen(true);
  }

  async function downloadPDF() {
    if (!canPreview || loading) return;
    setLoading(true);
    try {
      const pdf = await buildPDF();
      const fileName = `dokumen-kredit-${agentName.toLowerCase().replace(/[^a-z0-9]+/g, "-") || "agent"}.pdf`;
      pdf.save(fileName);
    } finally {
      setLoading(false);
    }
  }

  const previewModal = open ? (
    <div className="fixed inset-0 z-[100] flex items-end justify-center bg-slate-950/65 p-0 backdrop-blur-sm sm:items-center sm:p-3">
      <section className="flex max-h-[100svh] w-full max-w-4xl flex-col overflow-hidden rounded-t-[24px] bg-white shadow-[0_28px_80px_rgba(15,23,42,0.35)] sm:max-h-[92vh] sm:rounded-[28px]">
        <div className="flex min-w-0 items-center gap-2 border-b border-slate-100 px-3 py-3 sm:gap-3 sm:px-4">
          <div className="grid h-10 w-10 shrink-0 place-items-center rounded-2xl bg-emerald-50 text-emerald-700 sm:h-11 sm:w-11">
            <FileText className="h-5 w-5" strokeWidth={2.4} />
          </div>
          <div className="min-w-0 flex-1">
            <h3 className="truncate text-sm font-black text-slate-950">Preview Dokumen</h3>
            <p className="truncate text-[11px] font-bold text-slate-500">{agentName}</p>
          </div>
          <button
            type="button"
            onClick={downloadPDF}
            className="inline-flex h-10 shrink-0 items-center gap-1.5 rounded-2xl bg-emerald-600 px-2.5 text-[11px] font-black text-white transition hover:bg-emerald-700 sm:gap-2 sm:px-3 sm:text-xs"
          >
            <Download className="h-4 w-4" strokeWidth={2.5} />
            <span className="hidden min-[390px]:inline">Download</span>
          </button>
          <button
            type="button"
            onClick={() => setOpen(false)}
            className="grid h-10 w-10 shrink-0 place-items-center rounded-2xl bg-slate-100 text-slate-600 transition hover:bg-slate-200"
            aria-label="Tutup preview"
          >
            <X className="h-4 w-4" strokeWidth={2.6} />
          </button>
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto bg-slate-100 p-2 pb-[calc(0.75rem+env(safe-area-inset-bottom))] sm:p-5">
          <div className="mx-auto flex w-full max-w-3xl flex-col gap-4">
            {readyDocuments.map((doc, index) => (
              <figure
                key={`${doc.label}-${index}`}
                className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-[0_16px_36px_rgba(15,23,42,0.12)]"
              >
                <figcaption className="border-b border-slate-100 px-4 py-3 text-xs font-black text-slate-700">
                  {doc.label}
                </figcaption>
                <div className="relative h-[min(62svh,520px)] min-h-[260px] bg-slate-50 sm:min-h-[460px]">
                  <Image
                    src={doc.src}
                    alt={`${doc.label} ${agentName}`}
                    fill
                    unoptimized
                    sizes="(max-width: 768px) 92vw, 768px"
                    className="object-contain p-2"
                  />
                </div>
              </figure>
            ))}
          </div>
        </div>
      </section>
    </div>
  ) : null;

  return (
    <>
      <button
        type="button"
        disabled={!canPreview || loading}
        onClick={openPreview}
        className="flex w-full items-center gap-3 rounded-3xl border border-emerald-100 bg-white p-3 text-left shadow-[0_12px_26px_rgba(5,122,69,0.08)] transition hover:-translate-y-0.5 hover:border-emerald-300 disabled:cursor-not-allowed disabled:opacity-55"
        title={agentName}
      >
        <span className="grid h-12 w-12 shrink-0 place-items-center rounded-2xl bg-emerald-50 text-emerald-700">
          <FileText className="h-6 w-6" strokeWidth={2.4} />
        </span>
        <span className="min-w-0 flex-1">
          <span className="block text-xs font-black text-slate-950">PDF Dokumen</span>
          <span className="mt-0.5 block text-[10px] font-bold text-slate-500">
            {canPreview ? `${readyDocuments.length} foto tersimpan` : "Belum ada foto"}
          </span>
        </span>
        <span className="grid h-10 w-10 shrink-0 place-items-center rounded-2xl bg-emerald-600 text-white">
          <Eye className="h-4 w-4" strokeWidth={2.6} />
        </span>
      </button>

      {mounted && previewModal ? createPortal(previewModal, document.body) : null}
    </>
  );
}
