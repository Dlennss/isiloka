import Link from "next/link";
import { Bell, ChevronDown, UserRound } from "lucide-react";

type AppTopHeaderProps = {
  isLoggedIn?: boolean;
  userName?: string | null;
  saldo?: number | null;
  role?: string | null;
};

export function AppTopHeader({ isLoggedIn = false, userName, saldo, role }: AppTopHeaderProps) {
  const normalizedRole = String(role || "").trim().toLowerCase();
  const isRetailLoggedIn = isLoggedIn && (normalizedRole === "user" || normalizedRole === "agent" || normalizedRole === "master");
  const homeHref = isRetailLoggedIn ? "/user" : "/";
  void userName;
  void saldo;

  return (
    <header className="brand-app-header sticky top-0 z-30 overflow-hidden bg-[#0863d8] px-4 pb-3 pt-3 text-white shadow-[0_12px_28px_rgba(8,99,216,0.24)]">
      <div className="pointer-events-none absolute -right-16 -top-20 h-44 w-44 rounded-full bg-sky-300/22 blur-2xl" />
      <div className="pointer-events-none absolute left-36 top-0 h-24 w-28 rotate-12 rounded-[28px] bg-white/8" />

      <div className="relative flex h-13 items-center justify-between gap-3">
        <div className="flex min-w-0 flex-1 items-center">
          <Link
            href={homeHref}
            prefetch={false}
            className="flex min-w-0 items-center gap-2.5"
            aria-label="Isiloka"
          >
            <span className="grid h-12 w-12 shrink-0 place-items-center rounded-[15px] bg-white shadow-[0_10px_22px_rgba(2,36,91,0.18)] ring-1 ring-white/70">
              <img src="/brand/icon.svg" alt="" width={38} height={38} className="h-9.5 w-9.5" />
            </span>
            <span className="min-w-0">
              <span className="block text-[24px] font-black leading-6 tracking-tight text-white">
                <span className="brand-wordmark">Isiloka</span>
              </span>
              <span className="mt-0.5 block text-[9px] font-black uppercase text-sky-50/90">
                Isi hari, dari sini
              </span>
            </span>
          </Link>
        </div>

        <div className="flex shrink-0 items-center gap-2">
          <Link
            href={isLoggedIn ? "/user/transaksi" : "/transaksi"}
            prefetch={false}
            className="grid h-10 w-10 place-items-center rounded-[14px] border border-white/22 bg-white/12 text-white shadow-sm transition hover:bg-white/18"
            aria-label="Lihat transaksi"
          >
            <Bell className="h-5 w-5" strokeWidth={2.4} />
          </Link>
          <Link
            href={isLoggedIn ? "/user/account" : "/login"}
            prefetch={false}
            className="inline-flex h-10 items-center gap-1.5 rounded-[14px] border border-white/18 bg-white/14 px-3 text-sm font-black text-white shadow-sm transition hover:bg-white/20"
          >
            <UserRound className="h-4.5 w-4.5" />
            <span>{isLoggedIn ? "Akun" : "Masuk"}</span>
            <ChevronDown className="h-4 w-4 opacity-80" />
          </Link>
        </div>
      </div>
    </header>
  );
}
