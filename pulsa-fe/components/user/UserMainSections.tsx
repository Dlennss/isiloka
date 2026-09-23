import Link from "next/link";
import { ArrowUpRight, BookOpen, CalendarClock, Code2, Gamepad2, PlugZap, ReceiptText, Smartphone, Wifi, Zap } from "lucide-react";
import type { AgentCreditApplication } from "@/lib/api.auth";

const accessibleActionButtonClass =
  "inline-flex h-10 shrink-0 items-center justify-center gap-1.5 rounded-2xl border border-[#13464b] bg-[#13464b] px-3.5 text-[11px] font-black text-white! shadow-[0_10px_20px_rgba(6,78,59,0.18)] transition visited:text-white! hover:-translate-y-0.5 hover:bg-[#13464b] hover:text-white! focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-emerald-200 active:translate-y-0";

export function UserPulsaDataShortcut() {
  return (
    <section>
      <Link
        href="/user/pulsa-data"
        prefetch={false}
        className="group block overflow-hidden rounded-lg border border-[#13464b]/10 bg-linear-to-r from-[#ecfdf5] via-white to-[#f0f8f8] p-5 shadow-[0_10px_24px_rgba(6,78,59,0.08)] transition hover:-translate-y-0.5 hover:shadow-[0_16px_36px_rgba(6,78,59,0.12)]"
      >
        <div className="flex items-start justify-between gap-3">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[#087e8b]">Shortcut Cepat</p>
            <h2 className="mt-1 text-xl font-bold text-slate-900">Pulsa & Data by Nomor</h2>
            <p className="mt-2 text-sm text-slate-600">Masukkan nomor HP, deteksi operator otomatis, lalu pilih tab pulsa atau data tanpa cari brand manual.</p>
          </div>
          <div className="grid h-11 w-11 place-items-center rounded-lg bg-[#dcfce7] text-[#087e8b] transition group-hover:scale-105">
            <Code2 className="h-5 w-5" />
          </div>
        </div>
      </Link>
    </section>
  );
}

export function UserApiCTA() {
  return (
    <section>
      <div className="relative overflow-hidden rounded-[28px] bg-[#13464b] p-4 text-white shadow-[0_20px_44px_rgba(5,46,38,0.24)] ring-1 ring-lime-200/15">
        <div className="absolute -right-16 -top-16 h-44 w-44 rounded-full bg-lime-300/25 blur-3xl" />
        <div className="absolute -bottom-20 left-8 h-40 w-40 rounded-full bg-emerald-400/18 blur-3xl" />
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_18%_8%,rgba(255,255,255,0.18),transparent_30%),linear-gradient(135deg,rgba(16,185,129,0.20),rgba(163,230,53,0.06))]" />
        <div className="absolute inset-0 opacity-[0.12] bg-[repeating-radial-gradient(circle_at_0_100%,rgba(255,255,255,0.8)_0,rgba(255,255,255,0.8)_1px,transparent_1px,transparent_12px)] bg-size-[160%_130%]" />

        <div className="relative">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <div className="inline-flex items-center gap-2 rounded-full bg-white/10 px-3 py-1.5 text-[10px] font-black uppercase tracking-[0.18em] text-lime-200 ring-1 ring-white/15">
                <PlugZap className="h-3.5 w-3.5" />
                Kemitraan
              </div>
              <h2 className="mt-3 text-2xl leading-none font-black tracking-tight">API Reseller</h2>
              <p className="mt-2 max-w-[270px] text-sm leading-6 font-semibold text-white/78">
                Integrasi H2H untuk reseller, agen, dan website yang ingin transaksi otomatis.
              </p>
            </div>
            <div className="grid h-13 w-13 shrink-0 place-items-center rounded-[20px] bg-white/12 text-lime-200 shadow-[0_12px_24px_rgba(0,0,0,0.16)] ring-1 ring-white/20">
              <Code2 className="h-6 w-6" strokeWidth={2.4} />
            </div>
          </div>

          <div className="mt-4 grid grid-cols-2 gap-2">
            <div className="rounded-2xl bg-white/10 px-3 py-2 ring-1 ring-white/12">
              <p className="text-[10px] font-black uppercase tracking-[0.12em] text-white/52">Mode</p>
              <p className="mt-0.5 text-xs font-black text-white">H2H API</p>
            </div>
            <div className="rounded-2xl bg-white/10 px-3 py-2 ring-1 ring-white/12">
              <p className="text-[10px] font-black uppercase tracking-[0.12em] text-white/52">Cocok</p>
              <p className="mt-0.5 text-xs font-black text-white">Agen & Web</p>
            </div>
          </div>

          <div className="mt-4 grid grid-cols-1 gap-2 sm:grid-cols-2">
            <Link
              href="/docs"
              prefetch={false}
              className="inline-flex h-11 items-center justify-center gap-2 rounded-2xl bg-lime-300 px-3 text-sm font-black text-[#13464b] shadow-[0_12px_24px_rgba(163,230,53,0.22)] transition hover:bg-lime-200"
            >
              <BookOpen className="h-4 w-4" strokeWidth={2.5} />
              Dokumentasi
            </Link>
            <Link
              href="/artikel/cara-menjadi-member-h2h-pulsakilat"
              prefetch={false}
              className="inline-flex h-11 items-center justify-center gap-2 rounded-2xl border border-white/28 bg-white/10 px-3 text-sm font-black text-white! shadow-[0_12px_24px_rgba(0,0,0,0.12)] transition hover:bg-white/15 visited:text-white! hover:text-white!"
            >
              Pelajari H2H
              <ArrowUpRight className="h-4 w-4" strokeWidth={2.5} />
            </Link>
          </div>
        </div>
      </div>
    </section>
  );
}

export function UserWeeklyPromo() {
  return (
    <section>
      <div className="h-44 rounded-lg bg-linear-to-r from-[#13464b] via-[#087e8b] to-[#f6b49d] p-5 text-white shadow-sm">
        <p className="text-sm font-semibold text-white/90">Promo Mingguan</p>
        <h2 className="mt-2 max-w-70 text-2xl leading-tight font-bold">Yuk, isi kebutuhan digital lebih hemat!</h2>
      </div>
    </section>
  );
}

type UserRecentActivityProps = {
  href?: string;
};

export function UserRecentActivity({ href = "/kategori" }: UserRecentActivityProps) {
  return (
    <section className="space-y-3">
      <h2 className="px-0.5 text-lg font-black tracking-tight text-slate-950">Aktivitas Terakhir</h2>
      <div className="rounded-[22px] border border-emerald-950/10 bg-white p-4 shadow-[0_12px_30px_rgba(6,78,59,0.08)]">
        <div className="flex items-center gap-3">
          <div className="relative grid h-12 w-12 shrink-0 place-items-center rounded-[18px] bg-emerald-50 text-[#087e8b] ring-1 ring-emerald-100">
            <span className="absolute -right-1 -top-1 h-4 w-4 rounded-full bg-lime-300 ring-2 ring-white" />
            <ReceiptText className="h-6 w-6" strokeWidth={2.2} />
          </div>

          <div className="min-w-0 flex-1">
            <p className="text-sm font-black text-slate-950">Belum ada aktivitas</p>
            <p className="mt-1 text-[11px] font-semibold leading-4 text-slate-500">
              Transaksi pertamamu akan tercatat otomatis di sini.
            </p>
          </div>

          <Link href={href} prefetch={false} className={accessibleActionButtonClass}>
            Pilih Layanan
            <ArrowUpRight className="h-3.5 w-3.5" strokeWidth={2.6} />
          </Link>
        </div>
      </div>
    </section>
  );
}

type UserFavoriteTransactionsProps = {
  href?: string;
};

export function UserFavoriteTransactions({ href = "/kategori" }: UserFavoriteTransactionsProps) {
  return (
    <section className="space-y-3">
      <div className="px-0.5">
        <h2 className="text-lg font-black tracking-tight text-slate-950">Transaksi Favorit</h2>
        <p className="mt-0.5 text-[11px] font-semibold text-slate-500">Terbentuk otomatis dari transaksimu</p>
      </div>

      <div className="rounded-[22px] border border-emerald-950/10 bg-white p-4 shadow-[0_12px_30px_rgba(6,78,59,0.08)]">
        <div className="flex items-center gap-3">
          <div className="grid h-12 w-12 shrink-0 grid-cols-2 gap-1 rounded-[18px] bg-emerald-50 p-2 ring-1 ring-emerald-100">
            <span className="grid place-items-center rounded-lg bg-linear-to-br from-[#13464b] to-[#087e8b] text-white">
              <Smartphone className="h-3.5 w-3.5" strokeWidth={2.5} />
            </span>
            <span className="grid place-items-center rounded-lg bg-linear-to-br from-[#087e8b] to-[#f6b49d] text-white">
              <Wifi className="h-3.5 w-3.5" strokeWidth={2.5} />
            </span>
            <span className="grid place-items-center rounded-lg bg-linear-to-br from-[#65a30d] to-[#bef264] text-white">
              <Zap className="h-3.5 w-3.5" strokeWidth={2.5} />
            </span>
            <span className="grid place-items-center rounded-lg bg-linear-to-br from-[#13464b] to-[#06b6d4] text-white">
              <Gamepad2 className="h-3.5 w-3.5" strokeWidth={2.5} />
            </span>
          </div>

          <div className="min-w-0 flex-1">
            <p className="text-sm font-black text-slate-950">Belum ada transaksi favorit</p>
            <p className="mt-1 text-[11px] font-semibold leading-4 text-slate-500">
              Layanan yang sering kamu gunakan akan muncul otomatis.
            </p>
          </div>

          <Link href={href} prefetch={false} className={accessibleActionButtonClass}>
            Mulai Transaksi
            <ArrowUpRight className="h-3.5 w-3.5" strokeWidth={2.6} />
          </Link>
        </div>
      </div>
    </section>
  );
}

type UserMonthlyBillsProps = {
  href?: string;
  variant?: "user" | "agent";
  agentBills?: AgentCreditApplication[];
};

function formatIDR(value: number) {
  return `Rp ${new Intl.NumberFormat("id-ID").format(Number(value || 0))}`;
}

export function UserMonthlyBills({ href = "/kategori", variant = "user", agentBills = [] }: UserMonthlyBillsProps) {
  // Agent credit is an operating capital cycle, not a recurring bill.
  // Settlement is controlled by the operator when the partnership ends.
  if (variant === "agent") return null;

  const isAgentBill = false;
  const billHref = href;
  const outstanding = 0;
  const approved = 0;

  return (
    <section className="space-y-3">
      <div className="px-0.5">
        <h2 className="text-lg font-black tracking-tight text-slate-950">Tagihan Bulan Ini</h2>
        <p className="mt-0.5 text-[11px] font-semibold text-slate-500">
          {isAgentBill ? "Tagihan kredit agent yang sudah terpakai" : "Tagihan aktif milik akunmu"}
        </p>
      </div>

      <div className="rounded-[22px] border border-emerald-950/10 bg-white p-4 shadow-[0_12px_30px_rgba(6,78,59,0.08)]">
        <div className="flex items-center gap-3">
          <div className="relative grid h-12 w-12 shrink-0 place-items-center rounded-[18px] bg-emerald-50 text-[#087e8b] ring-1 ring-emerald-100">
            <span className="absolute -right-1 -top-1 grid h-5 w-5 place-items-center rounded-full bg-lime-300 text-[#13464b] ring-2 ring-white">
              <CalendarClock className="h-3 w-3" strokeWidth={2.5} />
            </span>
            <ReceiptText className="h-6 w-6" strokeWidth={2.2} />
          </div>

          <div className="min-w-0 flex-1">
            <p className="text-sm font-black text-slate-950">
              {isAgentBill ? "Tagihan Kredit Agent" : "Belum ada tagihan"}
            </p>
            <p className="mt-1 text-[11px] font-semibold leading-4 text-slate-500">
              {isAgentBill
                ? `Sisa ${formatIDR(outstanding)} dari limit ${formatIDR(approved)}.`
                : "Tagihan yang kamu cek atau simpan nanti akan tampil di sini."}
            </p>
          </div>

          <Link href={billHref} prefetch={false} className={accessibleActionButtonClass}>
            {isAgentBill ? "Bayar" : "Cek Tagihan"}
            <ArrowUpRight className="h-3.5 w-3.5" strokeWidth={2.6} />
          </Link>
        </div>
      </div>
    </section>
  );
}

export function UserAboutSection() {
  return (
    <section>
      <div className="rounded-lg border border-[#13464b]/10 bg-white px-4 py-6 shadow-[0_10px_24px_rgba(6,78,59,0.08)]">
        <h3 className="text-lg leading-tight font-bold text-slate-900">Isiloka - Solusi Digital untuk Kebutuhan Sehari-hari</h3>
        <p className="mt-3 text-sm leading-relaxed text-slate-600 text-justify">
          Isiloka hadir sebagai platform digital untuk memenuhi kebutuhan transaksi online Anda. Tersedia pembayaran praktis dan harga kompetitif. 
          <span><Link href="/tentang" prefetch={false} className="text-[#087e8b]! hover:underline!">
            Selengkapnya
          </Link></span>
        </p>
      </div>
    </section>
  );
}
