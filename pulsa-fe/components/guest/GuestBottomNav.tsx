"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { History, House, Tag, UserRound } from "lucide-react";

function navClass(active: boolean) {
  return active
    ? "flex min-w-0 flex-col items-center gap-1.5 rounded-2xl px-2 py-1.5 text-[#057b73]! visited:text-[#057b73]!"
    : "flex min-w-0 flex-col items-center gap-1.5 rounded-2xl px-2 py-1.5 text-[#7b8798]! transition visited:text-[#7b8798]! hover:text-[#057b73]!";
}

const iconClass = "h-5 w-5";
const textClass = "text-[11px] font-extrabold leading-none";

function iconShellClass(active: boolean) {
  return active
    ? "grid h-9 w-9 place-items-center rounded-2xl bg-[#057b73] text-white shadow-[0_9px_18px_rgba(5,123,115,0.24)]"
    : "grid h-9 w-9 place-items-center rounded-2xl text-[#7b8798]";
}

type GuestBottomNavProps = {
  isLoggedIn?: boolean;
};

export function GuestBottomNav({ isLoggedIn = false }: GuestBottomNavProps) {
  const pathname = usePathname() || "";
  const homeActive = pathname === "/";
  const historyActive = pathname.startsWith("/transaksi");
  const accountHref = isLoggedIn ? "/user/account" : "/login";
  const promoHref = "/artikel";
  const promoActive = pathname.startsWith("/artikel");
  const accountActive = isLoggedIn
    ? pathname.startsWith("/user/account")
    : pathname.startsWith("/login");

  return (
    <section className="isiloka-bottom-nav fixed bottom-3 left-1/2 z-[90] w-[calc(100%-2rem)] max-w-[398px] -translate-x-1/2 overflow-hidden rounded-[30px] border border-white/85 bg-white/95 shadow-[0_18px_46px_rgba(13,71,70,0.18)] backdrop-blur-2xl">
      <div className="grid grid-cols-4 px-3 pb-[calc(0.6rem+env(safe-area-inset-bottom))] pt-2.5">
        <Link href="/" prefetch={false} className={navClass(homeActive)}>
          <span className={iconShellClass(homeActive)}>
            <House className={iconClass} fill={homeActive ? "currentColor" : "none"} strokeWidth={homeActive ? 2.1 : 1.9} />
          </span>
          <span className={textClass}>Beranda</span>
        </Link>

        <Link href="/transaksi" prefetch={false} className={navClass(historyActive)}>
          <span className={iconShellClass(historyActive)}>
            <History className={iconClass} strokeWidth={historyActive ? 2.1 : 1.9} />
          </span>
          <span className={textClass}>Riwayat</span>
        </Link>

        <Link href={promoHref} prefetch={false} className={navClass(promoActive)}>
          <span className={iconShellClass(promoActive)}>
            <Tag className={iconClass} strokeWidth={promoActive ? 2.1 : 1.9} />
          </span>
          <span className={textClass}>Promo</span>
        </Link>

        <Link href={accountHref} prefetch={false} className={navClass(accountActive)}>
          <span className={iconShellClass(accountActive)}>
            <UserRound className={iconClass} strokeWidth={accountActive ? 2.1 : 1.9} />
          </span>
          <span className={textClass}>Akun</span>
        </Link>
      </div>
    </section>
  );
}
