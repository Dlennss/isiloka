import Link from "next/link";
import { getAppServerSession } from "@/lib/server-auth";
import { redirect } from "next/navigation";
import {
  ArrowLeft,
  Building2,
  ChevronRight,
  Landmark,
  ScanLine,
  Send,
  ShieldCheck,
  Sparkles,
  UsersRound,
  WalletCards,
  Zap,
} from "lucide-react";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import type { UserSession } from "@/components/user/types";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

const sendOptions = [
  {
    title: "Rekening Bank",
    desc: "Transfer ke bank Indonesia",
    href: "/user/transfer-bank",
    icon: Landmark,
    tone: "from-emerald-500 to-lime-400",
    badge: "Bank",
  },
  {
    title: "E-Wallet",
    desc: "DANA, OVO, GoPay, dan lainnya",
    href: "/user/ewallet",
    icon: WalletCards,
    tone: "from-cyan-500 to-emerald-400",
    badge: "Cepat",
  },
  {
    title: "Sesama Isiloka",
    desc: "Kirim saldo ke pengguna lain",
    href: "/user/saldo/kirim",
    icon: UsersRound,
    tone: "from-[#087e8b] to-[#84cc16]",
    badge: "Gratis",
  },
  {
    title: "Nomor Virtual Account",
    desc: "Bayar atau kirim via VA",
    href: "/user/account/topup",
    icon: ScanLine,
    tone: "from-slate-700 to-emerald-500",
    badge: "VA",
  },
];

export default async function UserSaldoKirimPage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  if (!session?.backendToken) redirect("/login");

  return (
    <main className="min-h-screen bg-[#f0f8f8] pb-24">
      <section className="relative overflow-hidden bg-[linear-gradient(135deg,#052e26_0%,#047857_62%,#84cc16_135%)] px-4 pb-7 pt-5 text-white shadow-[0_18px_42px_rgba(4,120,87,0.22)]">
        <div className="pointer-events-none absolute -right-10 -top-14 h-36 w-36 rounded-full bg-white/10" />
        <div className="pointer-events-none absolute right-10 top-10 h-20 w-20 rounded-full border border-white/10" />
        <div className="mx-auto flex w-full max-w-md items-center justify-between gap-3">
          <Link href="/user/saldo" className="grid h-10 w-10 shrink-0 place-items-center rounded-2xl bg-white/12 text-white ring-1 ring-white/15">
            <ArrowLeft className="h-5 w-5" strokeWidth={2.4} />
          </Link>
          <div className="min-w-0 flex-1">
            <h1 className="text-lg font-black tracking-tight">Kirim Saldo</h1>
            <p className="mt-0.5 text-[11px] font-semibold text-white/75">Pilih tujuan pengiriman</p>
          </div>
          <div className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-white text-[#087e8b] shadow-[0_12px_26px_rgba(6,78,59,0.18)]">
            <Zap className="h-5 w-5 fill-lime-300 text-[#087e8b]" strokeWidth={2.4} />
          </div>
        </div>
      </section>

      <div className="mx-auto -mt-3 w-full max-w-md space-y-4 px-4">
        <section className="rounded-[26px] border border-emerald-950/5 bg-white p-4 shadow-[0_18px_42px_rgba(6,78,59,0.10)]">
          <div className="flex items-center justify-between gap-3">
            <div>
              <p className="text-[10px] font-black uppercase tracking-[0.22em] text-[#087e8b]">Pilih Jenis</p>
              <h2 className="mt-1 text-xl font-black tracking-tight text-slate-950">Mau kirim ke mana?</h2>
            </div>
            <span className="grid h-10 w-10 place-items-center rounded-2xl bg-lime-100 text-[#087e8b]">
              <Send className="h-5 w-5" strokeWidth={2.4} />
            </span>
          </div>

          <div className="mt-4 grid grid-cols-2 gap-3">
            {sendOptions.map((item) => {
              const Icon = item.icon;
              return (
                <Link
                  key={item.title}
                  href={item.href}
                  className="group relative min-h-[136px] overflow-hidden rounded-[22px] border border-slate-200 bg-[#fbfffd] p-3 shadow-[0_10px_24px_rgba(15,23,42,0.04)] transition hover:-translate-y-0.5 hover:border-emerald-200 hover:shadow-[0_18px_34px_rgba(6,78,59,0.12)]"
                >
                  <div className={`absolute inset-x-0 top-0 h-1.5 bg-linear-to-r ${item.tone}`} />
                  <div className="flex items-start justify-between gap-2">
                    <span className={`grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-linear-to-br ${item.tone} text-white shadow-[0_10px_20px_rgba(6,78,59,0.14)]`}>
                      <Icon className="h-5 w-5" strokeWidth={2.4} />
                    </span>
                    <span className="rounded-full bg-emerald-50 px-2 py-1 text-[9px] font-black text-[#087e8b]">{item.badge}</span>
                  </div>
                  <h3 className="mt-4 text-sm font-black leading-4 text-slate-950">{item.title}</h3>
                  <p className="mt-1 text-[10px] font-semibold leading-4 text-slate-500">{item.desc}</p>
                  <span className="absolute bottom-3 right-3 grid h-7 w-7 place-items-center rounded-full bg-slate-100 text-slate-500 transition group-hover:bg-lime-300 group-hover:text-[#13464b]">
                    <ChevronRight className="h-4 w-4" strokeWidth={2.5} />
                  </span>
                </Link>
              );
            })}
          </div>
        </section>

        <section className="rounded-[24px] border border-emerald-200 bg-emerald-50 px-4 py-3">
          <div className="flex items-center gap-3">
            <span className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-white text-[#087e8b] ring-1 ring-emerald-100">
              <ShieldCheck className="h-5 w-5" strokeWidth={2.4} />
            </span>
            <div className="min-w-0">
              <p className="text-xs font-black text-slate-950">Pengiriman aman</p>
              <p className="mt-0.5 text-[10px] font-semibold leading-4 text-slate-500">Pastikan tujuan benar sebelum konfirmasi.</p>
            </div>
          </div>
        </section>

        <section className="rounded-[24px] border border-emerald-950/5 bg-white p-4 shadow-[0_16px_36px_rgba(6,78,59,0.08)]">
          <div className="flex items-center gap-3">
            <span className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-slate-100 text-slate-500">
              <Building2 className="h-5 w-5" strokeWidth={2.4} />
            </span>
            <div className="min-w-0 flex-1">
              <p className="text-xs font-black text-slate-950">Tujuan tersimpan</p>
              <p className="mt-0.5 text-[10px] font-semibold text-slate-400">Belum ada penerima favorit.</p>
            </div>
            <Sparkles className="h-4 w-4 shrink-0 text-lime-500" strokeWidth={2.5} />
          </div>
        </section>
      </div>

      <UserBottomNav />
    </main>
  );
}
