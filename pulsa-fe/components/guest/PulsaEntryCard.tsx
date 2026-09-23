"use client";

import Image from "next/image";
import { Phone, Sparkles } from "lucide-react";
import { getBrandLogo } from "@/lib/brand-logos";
import type { UserBrandItem, UserProductItem } from "@/components/user/types";

type PulsaEntryCardProps = {
  phone: string;
  nominalInput: string;
  onPhoneChange: (value: string) => void;
  onNominalChange: (value: string) => void;
  onQuickBuy: () => void;
  buyDisabled: boolean;
  detectedBrand: UserBrandItem | null;
  matchedProduct: UserProductItem | null;
  normalizedPhone: string;
  nominalValue: number;
  formatNominal: (value: number) => string;
  getDisplayedFixedPrice: (item: UserProductItem) => number;
};

export function PulsaEntryCard({
  phone,
  nominalInput,
  onPhoneChange,
  onNominalChange,
  onQuickBuy,
  buyDisabled,
  detectedBrand,
  matchedProduct,
  normalizedPhone,
  nominalValue,
  formatNominal,
  getDisplayedFixedPrice,
}: PulsaEntryCardProps) {
  const detectedLogo = detectedBrand ? getBrandLogo(detectedBrand.nama) : null;

  return (
    <section className="overflow-hidden rounded-md border border-sky-100 bg-white shadow-[0_18px_40px_rgba(15,23,42,0.10)]">
      <div className="bg-linear-to-r from-sky-50 via-white to-cyan-50 px-4 py-4">
        <div className="flex items-center gap-3">
          <div className="grid h-11 w-11 place-items-center rounded-2xl bg-white text-sky-600 shadow-[0_8px_20px_rgba(14,165,233,0.16)] ring-1 ring-sky-100">
            <Phone className="h-5 w-5" />
          </div>
          <div className="min-w-0">
            <h2 className="text-[15px] font-bold tracking-tight text-slate-900">Isi Pulsa</h2>
          </div>
        </div>
      </div>

      <div className="space-y-4 px-4 py-4">
        <div className="relative">
          <input
            type="tel"
            value={phone}
            onChange={(e) => onPhoneChange(e.target.value.replace(/\D/g, "").slice(0, 13))}
            minLength={8}
            maxLength={13}
            placeholder="Contoh: 081234567890"
            className={`w-full rounded-md border border-slate-200 bg-slate-50 px-4 py-3 pl-11 text-sm text-slate-900 placeholder:text-slate-400 focus:border-sky-500 focus:bg-white focus:outline-none focus:ring-2 focus:ring-sky-100 ${detectedBrand ? "pr-14" : ""}`}
          />
          <Phone className="absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
          {detectedBrand ? (
            <div className="absolute right-3 top-1/2 grid h-8 w-8 -translate-y-1/2 place-items-center overflow-hidden">
              {detectedLogo ? (
                <Image src={detectedLogo.src} alt={detectedLogo.alt} width={32} height={32} className="h-full w-full object-contain" />
              ) : (
                <span className="text-[10px] font-black uppercase text-sky-700">{detectedBrand.nama.slice(0, 2)}</span>
              )}
            </div>
          ) : null}
        </div>

        <div className="flex items-stretch gap-2">
          <div className="relative flex-1">
            <input
              type="tel"
              value={nominalInput}
              onChange={(e) => onNominalChange(e.target.value)}
              placeholder="Masukkan nominal pulsa"
              className="w-full rounded-md border border-slate-200 bg-slate-50 px-4 py-3 pl-11 text-sm text-slate-900 placeholder:text-slate-400 focus:border-sky-500 focus:bg-white focus:outline-none focus:ring-2 focus:ring-sky-100"
            />
            <Sparkles className="absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
          </div>
          <button
            type="button"
            onClick={onQuickBuy}
            disabled={buyDisabled}
            className="inline-flex h-auto min-w-20 items-center justify-center rounded-md bg-linear-to-r from-[#087e8b] to-[#087e8b] px-4 text-sm font-semibold text-white shadow-[0_8px_20px_rgba(15,111,203,0.24)] transition hover:opacity-95 disabled:cursor-not-allowed disabled:opacity-40"
          >
            Beli
          </button>
        </div>

      </div>
    </section>
  );
}
