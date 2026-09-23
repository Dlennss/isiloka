"use client";

import * as React from "react";
import Image from "next/image";
import Link from "next/link";
import { CheckCircle2, ChevronRight, RotateCcw, WalletCards } from "lucide-react";

type EwalletProviderItem = {
  key: string;
  title: string;
  href: string;
  imageSrc?: string;
  accent: string;
};

type EwalletProviderPickerProps = {
  items: EwalletProviderItem[];
};

function normalizeDigits(value: string) {
  return value.replace(/\D+/g, "").slice(0, 16);
}

function withDest(href: string, dest: string) {
  const cleanDest = normalizeDigits(dest);
  if (!cleanDest) return href;
  return `${href}?dest=${encodeURIComponent(cleanDest)}`;
}

function getInitials(value: string) {
  return value
    .replace(/[^a-z0-9 ]+/gi, " ")
    .trim()
    .split(/\s+/)
    .map((word) => word.charAt(0))
    .join("")
    .slice(0, 3)
    .toUpperCase();
}

export function EwalletProviderPicker({ items }: EwalletProviderPickerProps) {
  const [dest, setDest] = React.useState("");
  const [selectedKey, setSelectedKey] = React.useState(items[0]?.key ?? "");
  const selectedItem = items.find((item) => item.key === selectedKey) ?? items[0] ?? null;
  const cleanDest = normalizeDigits(dest);

  return (
    <div className="space-y-3 pb-28">
      <section className="overflow-hidden rounded-[28px] border border-emerald-950/5 bg-linear-to-br from-white via-emerald-50/80 to-lime-50/70 p-3 shadow-[0_18px_42px_rgba(6,78,59,0.12)]">
        <div className="flex items-center gap-3">
          <span className="grid h-13 w-13 shrink-0 place-items-center rounded-[20px] bg-linear-to-br from-[#13464b] to-[#16a34a] text-lime-200 shadow-[0_14px_28px_rgba(5,46,38,0.18)]">
            <WalletCards className="h-6 w-6" />
          </span>
          <div className="min-w-0 flex-1">
            <h1 className="text-base font-black tracking-tight text-slate-950">Nomor E-Wallet</h1>
            <p className="mt-0.5 text-xs font-semibold text-slate-500">Masukkan nomor akun tujuan.</p>
          </div>
          <button
            type="button"
            onClick={() => setDest("")}
            className="inline-flex h-9 items-center gap-1.5 rounded-full bg-emerald-50 px-3 text-[11px] font-black text-emerald-700 ring-1 ring-emerald-200"
          >
            <RotateCcw className="h-3.5 w-3.5" />
            Reset
          </button>
        </div>

        <label className="mt-4 flex h-14 items-center overflow-hidden rounded-2xl border border-emerald-500/25 bg-white shadow-inner focus-within:border-emerald-500 focus-within:ring-4 focus-within:ring-emerald-100">
          <span className="grid h-full w-14 shrink-0 place-items-center border-r border-emerald-100 text-sm font-black text-emerald-700">
            +62
          </span>
          <input
            value={dest}
            onChange={(event) => setDest(normalizeDigits(event.target.value))}
            inputMode="numeric"
            placeholder="Masukkan nomor e-wallet"
            className="h-full min-w-0 flex-1 bg-transparent px-3 text-sm font-bold text-slate-900 outline-none placeholder:text-slate-400"
          />
        </label>
      </section>

      <section className="overflow-hidden rounded-[30px] border border-emerald-950/5 bg-linear-to-br from-white via-[#f8fffb] to-emerald-50/85 p-3 shadow-[0_18px_42px_rgba(6,78,59,0.12)]">
        <div className="flex items-center justify-between gap-3 px-1 pb-3 pt-1">
          <div>
            <h2 className="text-base font-black tracking-tight text-slate-950">Pilih E-Wallet</h2>
            <p className="mt-0.5 text-xs font-semibold text-slate-500">{items.length} penyedia tersedia.</p>
          </div>
          {selectedItem ? (
            <span className="max-w-24 truncate rounded-full bg-[#13464b] px-3 py-1.5 text-[10px] font-black text-lime-200 shadow-[0_10px_18px_rgba(5,46,38,0.16)]">
              {selectedItem.title}
            </span>
          ) : null}
        </div>

        {items.length === 0 ? (
          <div className="grid min-h-40 place-items-center rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 text-center text-sm text-slate-500">
            Provider e-wallet belum tersedia.
          </div>
        ) : (
          <div className="grid grid-cols-3 gap-2.5">
            {items.map((item) => {
              const selected = item.key === selectedKey;
              return (
                <button
                  key={item.key}
                  type="button"
                  onClick={() => setSelectedKey(item.key)}
                  className={`group relative min-h-[104px] overflow-hidden rounded-[23px] border px-2.5 pb-2 pt-2.5 text-left transition duration-300 hover:-translate-y-0.5 ${
                    selected
                      ? "border-lime-300 bg-[#13464b] shadow-[0_18px_34px_rgba(5,46,38,0.24)] ring-2 ring-lime-200/70"
                      : "border-white bg-white shadow-[0_14px_30px_rgba(15,23,42,0.08)] ring-1 ring-slate-200/70 hover:border-emerald-200 hover:shadow-[0_18px_34px_rgba(6,78,59,0.13)]"
                  }`}
                >
                  <div className={`absolute inset-x-2 top-2 h-11 rounded-[19px] bg-linear-to-br ${item.accent} ${selected ? "opacity-45" : "opacity-16"} blur-[1px]`} />
                  <div className={`absolute -right-8 -top-8 h-22 w-22 rounded-full bg-linear-to-br ${item.accent} ${selected ? "opacity-35" : "opacity-18"} blur-xl`} />
                  <div className="absolute -bottom-10 left-3 h-18 w-18 rounded-full bg-white/25 blur-2xl" />
                  {selected ? (
                    <span className="absolute right-2 top-2 z-10 grid h-5.5 w-5.5 place-items-center rounded-full bg-lime-300 text-[#13464b] shadow-[0_8px_16px_rgba(0,0,0,0.18)]">
                      <CheckCircle2 className="h-3.5 w-3.5" strokeWidth={2.6} />
                    </span>
                  ) : null}
                  <div className={`relative mx-auto grid h-12.5 w-12.5 place-items-center overflow-hidden rounded-[18px] bg-white p-2.5 shadow-[0_12px_24px_rgba(15,23,42,0.13)] ring-1 ${selected ? "ring-lime-200" : "ring-slate-100"}`}>
                    {item.imageSrc ? (
                      <Image src={item.imageSrc} alt={item.title} width={56} height={56} className="h-full w-full object-contain" />
                    ) : (
                      <span className={`grid h-full w-full place-items-center rounded-xl bg-linear-to-br ${item.accent} text-sm font-black text-white`}>
                        {getInitials(item.title)}
                      </span>
                    )}
                  </div>
                  <p className={`relative mt-2 text-center text-[9px] font-black uppercase leading-tight ${selected ? "text-white" : "text-slate-950"}`}>
                    {item.title}
                  </p>
                </button>
              );
            })}
          </div>
        )}

        <Link
          href={selectedItem ? withDest(selectedItem.href, cleanDest) : "#"}
          aria-disabled={!selectedItem}
          className={`group relative mt-4 flex h-13 items-center justify-center gap-2 overflow-hidden rounded-[22px] text-sm font-black transition ${
            selectedItem
              ? "bg-linear-to-r from-[#13464b] via-[#13464b] to-[#13464b] text-white shadow-[0_18px_34px_rgba(5,46,38,0.28)] ring-1 ring-white/20"
              : "pointer-events-none bg-slate-200 text-slate-400"
          }`}
        >
          <span className="absolute inset-y-0 left-0 w-1/2 bg-linear-to-r from-emerald-400/28 to-transparent" />
          <span className="absolute -right-8 -top-10 h-24 w-24 rounded-full bg-lime-300/35 blur-xl transition group-hover:bg-lime-200/45" />
          <span className="relative tracking-wide drop-shadow-sm">Lihat Produk & Nominal</span>
          <span className="relative grid h-6 w-6 place-items-center rounded-full bg-lime-300 text-[#13464b] shadow-[0_8px_16px_rgba(0,0,0,0.18)] transition group-hover:translate-x-0.5">
            <ChevronRight className="h-4 w-4" strokeWidth={2.6} />
          </span>
        </Link>
      </section>
    </div>
  );
}
