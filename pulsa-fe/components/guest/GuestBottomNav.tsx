"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { History, House, Tag, UserRound } from "lucide-react";

function navClass(active: boolean) {
  return active
    ? "flex min-w-0 flex-col items-center gap-1.5 py-1 text-[#057b73]! visited:text-[#057b73]!"
    : "flex min-w-0 flex-col items-center gap-1.5 py-1 text-[#79859a]! transition visited:text-[#79859a]! hover:text-[#057b73]!";
}

const iconClass = "h-6 w-6";
const textClass = "text-[12px] font-bold leading-none";

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
    <section className="brand-bottom-nav fixed bottom-0 left-1/2 z-[90] w-full max-w-md -translate-x-1/2 overflow-hidden rounded-t-[28px] border-t border-teal-900/5 bg-white/96 shadow-[0_-18px_40px_rgba(25,73,78,0.10)] backdrop-blur-xl md:bottom-0 md:w-97.5 md:max-w-none">
      <div className="grid grid-cols-4 px-6 pb-[calc(0.75rem+env(safe-area-inset-bottom))] pt-3.5">
        <Link href="/" prefetch={false} className={navClass(homeActive)}>
          <House className={iconClass} fill={homeActive ? "currentColor" : "none"} strokeWidth={1.8} />
          <span className={textClass}>Beranda</span>
        </Link>

        <Link href="/transaksi" prefetch={false} className={navClass(historyActive)}>
          <History className={iconClass} strokeWidth={1.8} />
          <span className={textClass}>Riwayat</span>
        </Link>

        <Link href={promoHref} prefetch={false} className={navClass(promoActive)}>
          <Tag className={iconClass} strokeWidth={1.8} />
          <span className={textClass}>Promo</span>
        </Link>

        <Link href={accountHref} prefetch={false} className={navClass(accountActive)}>
          <UserRound className={iconClass} strokeWidth={1.8} />
          <span className={textClass}>Akun</span>
        </Link>
      </div>
    </section>
  );
}
