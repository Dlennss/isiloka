"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { History, House, UserRound, WalletCards } from "lucide-react";

function navClass(active: boolean) {
  return active
    ? "flex min-w-0 flex-col items-center gap-1.5 py-1 text-[#087e8b]! visited:text-[#087e8b]!"
    : "flex min-w-0 flex-col items-center gap-1.5 py-1 text-slate-400! transition visited:text-slate-400! hover:text-[#13464b]!";
}

const iconClass = "h-5 w-5";
const textClass = "text-[11px] font-bold leading-none";

type GuestBottomNavProps = {
  isLoggedIn?: boolean;
};

export function GuestBottomNav({ isLoggedIn = false }: GuestBottomNavProps) {
  const pathname = usePathname() || "";
  const homeActive = pathname === "/";
  const historyActive = pathname.startsWith("/transaksi");
  const accountHref = isLoggedIn ? "/user/account" : "/login";
  const saldoHref = isLoggedIn ? "/user/saldo" : "/login";
  const saldoActive = isLoggedIn ? pathname.startsWith("/user/saldo") || pathname.startsWith("/user/account/topup") || pathname.startsWith("/user/account/mutasi") : false;
  const accountActive = isLoggedIn
    ? pathname.startsWith("/user/account") && !saldoActive
    : pathname.startsWith("/login");

  return (
    <section className="brand-bottom-nav fixed bottom-0 left-1/2 z-[90] w-full max-w-md -translate-x-1/2 overflow-hidden rounded-t-[24px] border-t border-[#13464b]/10 bg-white/96 shadow-[0_-14px_34px_rgba(6,78,59,0.10)] backdrop-blur-xl md:bottom-0 md:w-97.5 md:max-w-none">
      <div className="grid grid-cols-4 px-4 pb-[calc(0.55rem+env(safe-area-inset-bottom))] pt-2.5">
        <Link href="/" prefetch={false} className={navClass(homeActive)}>
          <House className={iconClass} strokeWidth={1.65} />
          <span className={textClass}>Beranda</span>
        </Link>

        <Link href="/transaksi" prefetch={false} className={navClass(historyActive)}>
          <History className={iconClass} strokeWidth={1.65} />
          <span className={textClass}>Riwayat</span>
        </Link>

        <Link href={saldoHref} prefetch={false} className={navClass(saldoActive)}>
          <WalletCards className={iconClass} strokeWidth={1.65} />
          <span className={textClass}>Saldo</span>
        </Link>

        <Link href={accountHref} prefetch={false} className={navClass(accountActive)}>
          <UserRound className={iconClass} strokeWidth={1.65} />
          <span className={textClass}>Akun</span>
        </Link>
      </div>
    </section>
  );
}
