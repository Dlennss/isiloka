"use client";

import Link from "next/link";
import { useState } from "react";
import {
  ArrowDownToLine,
  ArrowUpRight,
  ChevronRight,
  Eye,
  EyeOff,
  Landmark,
  Plus,
  ReceiptText,
  ShieldCheck,
  WalletCards,
} from "lucide-react";

type UserSaldoPageContentProps = {
  saldo: number;
  userCode: string;
  showCredit?: boolean;
};

function formatIDR(value: number) {
  return `Rp ${new Intl.NumberFormat("id-ID").format(Number(value || 0))}`;
}

export function UserSaldoPageContent({ saldo, userCode, showCredit = false }: UserSaldoPageContentProps) {
  const [revealed, setRevealed] = useState(true);
  const displaySaldo = revealed ? formatIDR(saldo) : "Rp ******";

  const actions = [
    { label: "Isi Saldo", sub: "Tambah dana", href: "/user/account/topup", icon: Plus, tone: "bg-lime-100 text-[#087e8b]" },
    { label: "Kirim", sub: "Ke pengguna", href: "/user/saldo/kirim", icon: ArrowUpRight, tone: "bg-emerald-100 text-[#087e8b]" },
    { label: "Tarik", sub: "Ke rekening", href: "/user/account/withdraw", icon: ArrowDownToLine, tone: "bg-sky-100 text-sky-700" },
    ...(showCredit
      ? [{ label: "Tagihan", sub: "Pinjaman agent", href: "/user/saldo/tagihan", icon: ReceiptText, tone: "bg-amber-100 text-amber-700" }]
      : []),
  ];

  const topupMethods = [
    { title: "Transfer Bank", desc: "BCA, BRI, BNI, dan Mandiri", href: "/user/account/topup", icon: Landmark, tone: "bg-emerald-50 text-[#087e8b]" },
    { title: "Virtual Account", desc: "Verifikasi otomatis lebih cepat", href: "/user/account/topup", icon: WalletCards, tone: "bg-lime-50 text-lime-700" },
  ];

  return (
    <div className="space-y-3.5">
      <section className="relative overflow-hidden rounded-[28px] bg-[linear-gradient(135deg,#052e26_0%,#047857_52%,#84cc16_135%)] p-5 text-white shadow-[0_22px_50px_rgba(4,120,87,0.25)]">
        <div className="pointer-events-none absolute -right-12 -top-14 h-36 w-36 rounded-full border border-white/10 bg-white/10" />
        <div className="pointer-events-none absolute right-5 top-6 h-24 w-24 rounded-full border border-white/8 bg-white/6" />
        <div className="relative">
          <div className="flex items-start justify-between gap-3">
            <div>
              <div className="flex items-center gap-2">
                <p className="text-xs font-bold text-white/85">Saldo tersedia</p>
                <button
                  type="button"
                  onClick={() => setRevealed((value) => !value)}
                  aria-label={revealed ? "Sembunyikan saldo" : "Tampilkan saldo"}
                  className="grid h-7 w-7 place-items-center rounded-full bg-white/12 text-white/85"
                >
                  {revealed ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
                </button>
              </div>
              <h1 className="mt-4 text-3xl font-black tracking-tight">{displaySaldo}</h1>
              <p className="mt-1 text-[11px] font-semibold text-white/78">Siap dipakai untuk semua transaksi Isiloka</p>
            </div>
            <div className="grid h-12 w-12 place-items-center rounded-2xl bg-white/12 text-lime-200 ring-1 ring-white/15">
              <WalletCards className="h-5 w-5" strokeWidth={2.4} />
            </div>
          </div>

          <div className="mt-5 flex items-center justify-between border-t border-white/15 pt-3">
            <p className="text-[10px] font-black uppercase tracking-[0.22em] text-lime-100">Isiloka Wallet</p>
            <p className="max-w-[160px] truncate text-right text-[10px] font-black uppercase tracking-[0.18em] text-white/80">{userCode}</p>
          </div>
        </div>
      </section>

      <section className="rounded-[24px] border border-emerald-950/5 bg-white p-3 shadow-[0_16px_36px_rgba(6,78,59,0.08)]">
        <div className={`grid gap-1.5 ${showCredit ? "grid-cols-4" : "grid-cols-3"}`}>
          {actions.map((item) => {
            const Icon = item.icon;
            return (
              <Link key={item.label} href={item.href} className="group min-w-0 rounded-2xl px-1 py-2 text-center transition hover:bg-emerald-50">
                <span className={`mx-auto grid h-10 w-10 place-items-center rounded-2xl ${item.tone} transition group-hover:scale-105`}>
                  <Icon className="h-5 w-5" strokeWidth={2.4} />
                </span>
                <span className="mt-2 block truncate text-[10px] font-black text-slate-950">{item.label}</span>
                <span className="mt-0.5 block truncate text-[9px] font-semibold text-slate-400">{item.sub}</span>
              </Link>
            );
          })}
        </div>
      </section>

      <Link href="/user/account/mutasi" className="flex items-center gap-3 rounded-[22px] border border-emerald-200 bg-emerald-50/90 px-4 py-3 text-[#087e8b] shadow-[0_10px_24px_rgba(4,120,87,0.07)]">
        <span className="grid h-10 w-10 shrink-0 place-items-center rounded-2xl bg-white text-[#087e8b] ring-1 ring-emerald-100">
          <ShieldCheck className="h-5 w-5" strokeWidth={2.4} />
        </span>
        <span className="min-w-0 flex-1">
          <span className="block text-[11px] font-black">Saldo aman dan terlindungi</span>
          <span className="mt-0.5 block text-[10px] font-semibold text-[#087e8b]/70">Setiap transaksi dijaga dengan PIN dan verifikasi.</span>
        </span>
        <ChevronRight className="h-4 w-4 shrink-0" />
      </Link>

      <section className="rounded-[24px] border border-emerald-950/5 bg-white px-4 py-4 shadow-[0_16px_36px_rgba(6,78,59,0.08)]">
        <div className="grid grid-cols-[1fr_auto_1fr] items-center gap-4">
          <div className="flex min-w-0 items-center gap-3">
            <span className="grid h-12 w-12 shrink-0 place-items-center rounded-full bg-emerald-50 text-[#087e8b]">
              <ArrowDownToLine className="h-6 w-6" strokeWidth={2.5} />
            </span>
            <div className="min-w-0">
              <p className="truncate text-sm font-semibold text-slate-500">Uang Masuk</p>
              <p className="mt-1 text-xl font-black tracking-tight text-slate-950">Rp0</p>
            </div>
          </div>

          <div className="h-14 w-px bg-slate-200" />

          <div className="flex min-w-0 items-center gap-3">
            <span className="grid h-12 w-12 shrink-0 place-items-center rounded-full bg-amber-50 text-amber-600">
              <ArrowUpRight className="h-6 w-6" strokeWidth={2.5} />
            </span>
            <div className="min-w-0">
              <p className="truncate text-sm font-semibold text-slate-500">Uang Keluar</p>
              <p className="mt-1 text-xl font-black tracking-tight text-slate-950">Rp0</p>
            </div>
          </div>
        </div>
      </section>

      <section className="overflow-hidden rounded-[24px] border border-emerald-950/5 bg-white shadow-[0_16px_36px_rgba(6,78,59,0.08)]">
        <div className="flex items-center justify-between border-b border-slate-100 px-4 py-3.5">
          <h2 className="text-sm font-black text-slate-950">Cara Isi Saldo</h2>
          <Link href="/user/account/topup" className="text-[10px] font-black text-[#087e8b]">Lihat semua</Link>
        </div>
        <div className="divide-y divide-slate-100">
          {topupMethods.map((item) => {
            const Icon = item.icon;
            return (
              <Link key={item.title} href={item.href} className="flex items-center gap-3 px-4 py-3.5 transition hover:bg-emerald-50/60">
                <span className={`grid h-11 w-11 shrink-0 place-items-center rounded-2xl ${item.tone}`}>
                  <Icon className="h-5 w-5" strokeWidth={2.4} />
                </span>
                <span className="min-w-0 flex-1">
                  <span className="block text-xs font-black text-slate-950">{item.title}</span>
                  <span className="mt-0.5 block truncate text-[10px] font-semibold text-slate-400">{item.desc}</span>
                </span>
                <ChevronRight className="h-4 w-4 shrink-0 text-slate-400" />
              </Link>
            );
          })}
        </div>
      </section>

      <section className="overflow-hidden rounded-[24px] border border-emerald-950/5 bg-white shadow-[0_16px_36px_rgba(6,78,59,0.08)]">
        <div className="border-b border-slate-100 px-4 py-3.5">
          <h2 className="text-sm font-black text-slate-950">Aktivitas Dompet</h2>
        </div>
        <div className="grid min-h-[220px] place-items-center px-6 py-8 text-center">
          <div>
            <div className="mx-auto grid h-14 w-14 place-items-center rounded-2xl bg-emerald-50 text-[#087e8b] ring-1 ring-emerald-100">
              <WalletCards className="h-6 w-6" strokeWidth={2.4} />
            </div>
            <p className="mt-4 text-sm font-black text-slate-950">Belum ada aktivitas</p>
            <p className="mx-auto mt-1 max-w-[240px] text-[11px] font-semibold leading-4 text-slate-400">Isi saldo atau lakukan transaksi pertamamu. Riwayat dompet akan muncul di sini.</p>
            <Link href="/user/account/topup" className="mt-5 inline-flex h-10 items-center justify-center rounded-2xl bg-lime-100 px-4 text-xs font-black text-[#087e8b]">
              Isi Saldo Sekarang
            </Link>
          </div>
        </div>
      </section>
    </div>
  );
}
