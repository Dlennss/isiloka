"use client";

import * as React from "react";
import Link from "next/link";
import { ChevronLeft } from "lucide-react";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";

function Block({
  id,
  title,
  desc,
  children,
}: {
  id: string;
  title: string;
  desc?: string;
  children: React.ReactNode;
}) {
  return (
    <section
      id={id}
      className="scroll-mt-24 rounded-md border border-slate-200 bg-white p-4 shadow-[0_16px_36px_rgba(15,23,42,0.06)]"
    >
      <div className="inline-flex rounded-full bg-sky-50 px-3 py-1 text-[10px] font-semibold uppercase tracking-[0.16em] text-sky-700">
        Section
      </div>
      <h2 className="mt-3 text-lg font-black tracking-tight text-slate-950">{title}</h2>
      {desc ? <p className="mt-2 text-sm leading-6 text-slate-600">{desc}</p> : null}
      <div className="mt-4">{children}</div>
    </section>
  );
}

export default function DocsPage() {
  const [copied, setCopied] = React.useState<string>("");

  const jsonExample = `{
  "commands": "PAY",
  "product": "DNID",
  "dest": "0812676767",
  "qty": 11000,
  "refid": "1700000000123",
  "pin": "00000"
}`;

  const curlExample = `curl -sS -X POST "https://api.pulsakilat.net/v1/trx" \\
  -H "Content-Type: application/json" \\
  -H "X-Api-Key: YOUR_API_KEY" \\
  --data '${jsonExample.replace(/\n/g, "\\n").replace(/'/g, "\\'")}'`;

  const copy = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(text === jsonExample ? "JSON copied" : "cURL copied");
      window.setTimeout(() => setCopied(""), 1200);
    } catch {
      setCopied("Copy gagal");
      window.setTimeout(() => setCopied(""), 1200);
    }
  };

  return (
    <main className="relative overflow-hidden bg-slate-50 text-slate-900">
      <div className="pointer-events-none absolute inset-x-0 top-0 h-56 bg-[radial-gradient(520px_220px_at_15%_0%,rgba(79,124,255,0.18),transparent_60%),radial-gradient(560px_240px_at_100%_0%,rgba(6,182,212,0.12),transparent_55%),linear-gradient(180deg,#eef4ff_0%,rgba(248,250,252,0)_100%)]" />

      <div className="relative space-y-4 px-4 py-4">
        <Link
          href="/"
          className="inline-flex items-center gap-1 text-sm font-medium text-sky-700"
          aria-label="Kembali ke beranda"
        >
          <ChevronLeft className="h-4 w-4" />
          <span>Kembali</span>
        </Link>

        <section className="overflow-hidden rounded-md border border-sky-100 bg-linear-to-br from-white via-sky-50/70 to-cyan-50/80 p-4 shadow-[0_18px_40px_rgba(15,111,203,0.10)]">
          <p className="inline-flex rounded-full border border-sky-200 bg-white/90 px-4 py-1.5 text-[10px] font-semibold uppercase tracking-[0.18em] text-sky-700">
            Developer & Operasional Docs
          </p>
          <h1 className="mt-4 text-balance text-2xl font-black leading-tight tracking-tight text-slate-950">
            Dokumentasi Isiloka
          </h1>
          <p className="mt-3 text-sm leading-6 text-slate-600">
            Referensi penggunaan platform untuk agen dan developer: cara kerja, panduan operasional, API, dan FAQ.
          </p>
        </section>

        <div className="flex flex-wrap gap-2 rounded-md border border-slate-200 bg-white p-3 shadow-[0_12px_28px_rgba(15,23,42,0.05)]">
          <a className="rounded-full border border-sky-100 bg-sky-50 px-4 py-2 text-sm text-sky-700 hover:bg-sky-100" href="#cara-kerja">Cara Kerja</a>
          <a className="rounded-full border border-sky-100 bg-sky-50 px-4 py-2 text-sm text-sky-700 hover:bg-sky-100" href="#panduan-agen">Panduan Agen</a>
          <a className="rounded-full border border-sky-100 bg-sky-50 px-4 py-2 text-sm text-sky-700 hover:bg-sky-100" href="#api">Panduan API</a>
          <a className="rounded-full border border-sky-100 bg-sky-50 px-4 py-2 text-sm text-sky-700 hover:bg-sky-100" href="#faq">FAQ</a>
        </div>

        <Block id="cara-kerja" title="Cara Kerja" desc="Flow singkat dari daftar sampai transaksi berjalan.">
          <div className="grid gap-3">
            {["Daftar akun", "Isi saldo", "Pilih produk", "Kirim transaksi", "Pantau hasil"].map((step, idx) => (
              <div key={step} className="flex items-center gap-3 rounded-2xl border border-slate-200 bg-slate-50/80 p-4">
                <div className="grid h-8 w-8 shrink-0 place-items-center rounded-xl bg-linear-to-br from-sky-500 to-cyan-500 text-xs font-bold text-white shadow-[0_8px_18px_rgba(15,111,203,0.22)]">
                  {idx + 1}
                </div>
                <div className="font-semibold text-slate-800">{step}</div>
              </div>
            ))}
          </div>
        </Block>

        <Block id="panduan-agen" title="Panduan Agen" desc="Saran operasional agar transaksi rapi dan minim kesalahan.">
          <div className="grid gap-3">
            <div className="rounded-2xl border border-slate-200 bg-linear-to-br from-white to-slate-50 p-4">
              <div className="font-semibold text-slate-900">Checklist harian</div>
              <ul className="mt-2 list-disc space-y-1 pl-5 text-sm text-slate-600">
                <li>Pastikan saldo tersedia sebelum jam sibuk.</li>
                <li>Gunakan format order yang konsisten.</li>
                <li>Cek status transaksi sebelum retry.</li>
                <li>Gunakan histori untuk rekonsiliasi.</li>
              </ul>
            </div>
            <div className="rounded-2xl border border-slate-200 bg-linear-to-br from-sky-50/80 to-cyan-50/70 p-4">
              <div className="font-semibold text-slate-900">Tips margin</div>
              <p className="mt-2 text-sm leading-7 text-slate-600">
                Fokus ke volume dengan markup terukur. Jaga repeat order dari pelanggan dengan response cepat dan catatan transaksi rapi.
              </p>
            </div>
          </div>
        </Block>

        <Block id="api" title="Panduan API" desc="Untuk developer yang ingin integrasi transaksi otomatis.">
          <div className="grid gap-3">
            <div className="rounded-2xl border border-slate-200 bg-linear-to-br from-white to-slate-50 p-4">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div className="font-semibold text-slate-900">Contoh payload</div>
                <button
                  className="rounded-lg border border-slate-200 bg-white px-3 py-1 text-sm text-slate-700 hover:bg-slate-100"
                  onClick={() => void copy(jsonExample)}
                  type="button"
                >
                  Copy JSON
                </button>
              </div>
              <pre className="mt-3 max-w-full overflow-x-auto rounded-xl border border-slate-200 bg-white p-4 text-xs leading-6 text-slate-700">{jsonExample}</pre>
            </div>

            <div className="min-w-0 overflow-hidden rounded-2xl border border-slate-200 bg-linear-to-br from-sky-50/70 to-white p-4">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div className="font-semibold text-slate-900">Contoh cURL</div>
                <button
                  className="rounded-lg border border-slate-200 bg-white px-3 py-1 text-sm text-slate-700 hover:bg-slate-100"
                  onClick={() => void copy(curlExample)}
                  type="button"
                >
                  Copy cURL
                </button>
              </div>
              <pre className="mt-3 max-w-full overflow-x-auto rounded-xl border border-slate-200 bg-white p-4 text-xs leading-6 text-slate-700">{curlExample}</pre>
              <p className="mt-2 text-xs text-slate-500">
                Ganti <b>YOUR_API_KEY</b> dengan API key milik Anda dan jalankan dari backend server.
              </p>
            </div>
          </div>
        </Block>

        <Block id="faq" title="FAQ" desc="Pertanyaan yang paling sering muncul.">
          <div className="space-y-3">
            {[
              {
                q: "Kenapa transaksi bisa gagal?",
                a: "Biasanya karena saldo tidak cukup, data tujuan tidak valid, atau provider sedang gangguan sementara.",
              },
              {
                q: "Bagaimana kalau status pending?",
                a: "Simpan refid lalu cek kembali riwayat atau status transaksi. Hindari retry berulang tanpa validasi status terakhir.",
              },
              {
                q: "Bagaimana mulai jadi agen?",
                a: "Daftar akun, isi saldo, lalu mulai transaksi. Untuk volume tinggi, gunakan API agar proses lebih otomatis.",
              },
            ].map((item) => (
              <details key={item.q} className="group rounded-2xl border border-slate-200 bg-linear-to-br from-white to-slate-50 p-4">
                <summary className="cursor-pointer list-none font-semibold text-slate-900">
                  {item.q}
                  <span className="ml-2 inline-block text-slate-400 transition group-open:rotate-180">⌄</span>
                </summary>
                <p className="mt-2 text-sm leading-7 text-slate-600">{item.a}</p>
              </details>
            ))}
          </div>
        </Block>

        {copied ? (
          <p className="rounded-xl border border-emerald-200 bg-emerald-50 px-3 py-2 text-xs text-emerald-700">
            {copied}
          </p>
        ) : null}
      </div>

      <GuestBottomNav />
    </main>
  );
}
