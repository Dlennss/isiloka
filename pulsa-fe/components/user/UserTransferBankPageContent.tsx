"use client";

import Image from "next/image";
import Link from "next/link";
import { useState } from "react";
import { ArrowLeft, CheckCircle2, ChevronLeft, ChevronRight, Landmark, RotateCcw, ShieldCheck } from "lucide-react";

const banks = [
  { name: "BCA", image: "/images/banks/bca.jpeg" },
  { name: "BRI", image: "/images/banks/bri.jpg" },
  { name: "BNI", image: "/images/banks/bni.jpg" },
  { name: "Mandiri", image: "/images/banks/mandiri.jpeg" },
  { name: "Bank Syariah Indonesia", shortName: "BSI", image: "/images/banks/bsi.jpg" },
  { name: "Danamon", image: "/images/banks/danamon.png" },
  { name: "PermataBank", image: "/images/banks/permatabank.jpeg" },
  { name: "SeaBank", image: "/images/banks/seabank.jpeg" },
  { name: "Bank Jago", image: "/images/banks/jago.jpg" },
];

type UserTransferBankPageContentProps = {
  backHref?: string;
};

export function UserTransferBankPageContent({ backHref = "/user/kategori" }: UserTransferBankPageContentProps) {
  const [accountNumber, setAccountNumber] = useState("");
  const [selectedBankName, setSelectedBankName] = useState<string | null>(null);
  const selectedBank = banks.find((bank) => bank.name === selectedBankName) || null;
  const canContinue = Boolean(selectedBank && accountNumber.trim().length >= 6);

  function chooseBank(bankName: string) {
    setSelectedBankName(bankName);
    setAccountNumber("");
  }

  function changeBank() {
    setSelectedBankName(null);
    setAccountNumber("");
  }

  return (
    <>
      <section className="bg-[linear-gradient(135deg,#052e26_0%,#047857_72%,#65a30d_145%)] px-4 pb-6 pt-5 text-white shadow-[0_16px_36px_rgba(4,120,87,0.20)]">
        <div className="mx-auto flex w-full max-w-md items-center gap-3">
          {selectedBank ? (
            <button
              type="button"
              onClick={changeBank}
              className="grid h-10 w-10 shrink-0 place-items-center rounded-md bg-white/12 text-white ring-1 ring-white/15"
              aria-label="Kembali ke pilihan bank"
            >
              <ChevronLeft className="h-5 w-5" strokeWidth={2.4} />
            </button>
          ) : (
            <Link
              href={backHref}
              className="grid h-10 w-10 shrink-0 place-items-center rounded-md bg-white/12 text-white ring-1 ring-white/15"
              aria-label="Kembali"
            >
              <ArrowLeft className="h-5 w-5" strokeWidth={2.4} />
            </Link>
          )}

          <div className="min-w-0 flex-1">
            <p className="text-[10px] font-bold uppercase tracking-[0.16em] text-lime-200">
              {selectedBank ? "Langkah 2 dari 2" : "Langkah 1 dari 2"}
            </p>
            <h1 className="mt-0.5 truncate text-lg font-black tracking-tight">Transfer Bank</h1>
            <p className="mt-0.5 truncate text-[11px] font-semibold text-white/75">
              {selectedBank ? "Masukkan rekening tujuan" : "Pilih bank tujuan transfer"}
            </p>
          </div>

          <span className="grid h-11 w-11 shrink-0 place-items-center rounded-md bg-white text-[#087e8b]">
            <Landmark className="h-5 w-5" strokeWidth={2.4} />
          </span>
        </div>
      </section>

      <div className="mx-auto w-full max-w-md px-4 pb-6 pt-4">
        {!selectedBank ? (
          <section>
            <div className="flex items-end justify-between gap-3 px-1">
              <div>
                <h2 className="text-lg font-black text-slate-950">Pilih Bank</h2>
                <p className="mt-1 text-xs font-medium text-slate-500">Ketuk salah satu bank untuk melanjutkan.</p>
              </div>
              <span className="shrink-0 rounded-full bg-emerald-100 px-3 py-1 text-[10px] font-black text-emerald-700">
                {banks.length} bank
              </span>
            </div>

            <div className="mt-4 grid grid-cols-3 gap-2.5">
              {banks.map((bank) => (
                <button
                  key={bank.name}
                  type="button"
                  onClick={() => chooseBank(bank.name)}
                  className="group relative grid aspect-[1/1.04] min-w-0 grid-rows-[1fr_auto] overflow-hidden rounded-md border border-slate-200 bg-white p-2.5 text-center shadow-[0_8px_20px_rgba(15,23,42,0.05)] transition hover:-translate-y-0.5 hover:border-emerald-400 hover:shadow-[0_14px_28px_rgba(6,78,59,0.12)] focus-visible:border-emerald-600 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-emerald-100"
                >
                  <span className="relative block min-h-0 w-full">
                    <Image
                      src={bank.image}
                      alt={`Logo ${bank.name}`}
                      fill
                      sizes="(max-width: 430px) 29vw, 112px"
                      className="object-contain p-2"
                    />
                  </span>
                  <span className="line-clamp-2 min-h-7 text-[10px] font-black leading-3.5 text-slate-800">
                    {bank.shortName || bank.name}
                  </span>
                </button>
              ))}
            </div>
          </section>
        ) : (
          <section className="space-y-4">
            <div className="rounded-md border border-emerald-200 bg-white p-4 shadow-[0_16px_34px_rgba(6,78,59,0.09)]">
              <div className="flex items-center gap-3">
                <span className="relative block h-16 w-24 shrink-0 overflow-hidden rounded-md border border-slate-100 bg-white">
                  <Image
                    src={selectedBank.image}
                    alt={`Logo ${selectedBank.name}`}
                    fill
                    sizes="96px"
                    className="object-contain p-2"
                  />
                </span>
                <div className="min-w-0 flex-1">
                  <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-emerald-700">Bank tujuan</p>
                  <h2 className="mt-1 text-base font-black leading-tight text-slate-950">{selectedBank.name}</h2>
                </div>
                <CheckCircle2 className="h-5 w-5 shrink-0 text-emerald-600" strokeWidth={2.4} />
              </div>

              <button
                type="button"
                onClick={changeBank}
                className="mt-3 inline-flex h-9 items-center gap-2 rounded-md border border-emerald-200 bg-emerald-50 px-3 text-[11px] font-black text-emerald-700"
              >
                <ChevronLeft className="h-4 w-4" />
                Ganti Bank
              </button>
            </div>

            <div className="rounded-md border border-slate-200 bg-white p-4 shadow-[0_16px_34px_rgba(15,23,42,0.06)]">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <label htmlFor="transfer-account-number" className="text-sm font-black text-slate-950">
                    Nomor Rekening Tujuan
                  </label>
                  <p className="mt-1 text-[11px] font-medium text-slate-500">Pastikan rekening sesuai dengan bank yang dipilih.</p>
                </div>
                {accountNumber ? (
                  <button
                    type="button"
                    onClick={() => setAccountNumber("")}
                    className="inline-flex h-8 shrink-0 items-center gap-1.5 rounded-md bg-slate-100 px-2.5 text-[10px] font-black text-slate-600"
                  >
                    <RotateCcw className="h-3.5 w-3.5" />
                    Reset
                  </button>
                ) : null}
              </div>

              <input
                id="transfer-account-number"
                value={accountNumber}
                onChange={(event) => setAccountNumber(event.target.value.replace(/\D+/g, "").slice(0, 24))}
                className="mt-4 h-13 w-full rounded-md border border-slate-200 bg-slate-50 px-4 text-base font-bold text-slate-950 outline-none placeholder:text-sm placeholder:font-medium placeholder:text-slate-400 focus:border-emerald-600 focus:bg-white focus:ring-4 focus:ring-emerald-100"
                placeholder="Masukkan nomor rekening"
                inputMode="numeric"
                autoFocus
              />

              <button
                type="button"
                disabled={!canContinue}
                className="mt-4 flex h-13 w-full items-center justify-center gap-2 rounded-md bg-[linear-gradient(135deg,#13464b,#087e8b,#65a30d)] px-4 text-sm font-black text-white shadow-[0_14px_28px_rgba(4,120,87,0.20)] transition hover:brightness-105 disabled:bg-none disabled:bg-slate-200 disabled:text-slate-400 disabled:shadow-none"
              >
                Lanjutkan
                <ChevronRight className="h-4 w-4" strokeWidth={2.6} />
              </button>
            </div>

            <div className="flex items-center gap-3 rounded-md border border-emerald-200 bg-emerald-50 px-4 py-3">
              <span className="grid h-9 w-9 shrink-0 place-items-center rounded-md bg-white text-emerald-700">
                <ShieldCheck className="h-4.5 w-4.5" strokeWidth={2.4} />
              </span>
              <p className="text-[11px] font-semibold leading-4 text-emerald-800">
                Nomor rekening akan diperiksa sebelum transaksi dilanjutkan.
              </p>
            </div>
          </section>
        )}
      </div>
    </>
  );
}
