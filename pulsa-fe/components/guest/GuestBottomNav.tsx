"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { History, House, Tag, UserRound } from "lucide-react";

function navClass(active: boolean) {
  return active
    ? "flex min-w-0 flex-col items-center gap-1.5 rounded-2xl px-2 py-1.5 text-[#057b73]! visited:text-[#057b73]!"
    : "flex min-w-0 flex-col items-center gap-1.5 rounded-2xl px-2 py-1.5 text-[#7c8aa1]! transition visited:text-[#7c8aa1]! hover:text-[#057b73]!";
}

const iconClass = "h-5.5 w-5.5";
const textClass = "text-[11px] font-extrabold leading-none";

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
    <section className="brand-bottom-nav fixed bottom-0 left-1/2 z-[90] w-[calc(100%-1.5rem)] max-w-[398px] -translate-x-1/2 overflow-hidden rounded-t-[28px] border border-white/80 bg-white/94 shadow-[0_-14px_34px_rgba(25,73,78,0.12)] backdrop-blur-xl md:bottom-4 md:max-w-[398px] md:rounded-[28px]">
      <div className="grid grid-cols-4 px-4 pb-[calc(0.65rem+env(safe-area-inset-bottom))] pt-3">
        <Link href="/" prefetch={false} className={navClass(homeActive)}>
          <span className={homeActive ? "grid h-8 w-8 place-items-center rounded-2xl bg-[#e2fbf4]" : "grid h-8 w-8 place-items-center rounded-2xl"}>
            <House className={iconClass} fill={homeActive ? "currentColor" : "none"} strokeWidth={homeActive ? 2.2 : 1.9} />
          </span>
          <span className={textClass}>Beranda</span>
        </Link>

        <Link href="/transaksi" prefetch={false} className={navClass(historyActive)}>
          <span className={historyActive ? "grid h-8 w-8 place-items-center rounded-2xl bg-[#e2fbf4]" : "grid h-8 w-8 place-items-center rounded-2xl"}>
            <History className={iconClass} strokeWidth={historyActive ? 2.2 : 1.9} />
          </span>
          <span className={textClass}>Riwayat</span>
        </Link>

        <Link href={promoHref} prefetch={false} className={navClass(promoActive)}>
          <span className={promoActive ? "grid h-8 w-8 place-items-center rounded-2xl bg-[#e2fbf4]" : "grid h-8 w-8 place-items-center rounded-2xl"}>
            <Tag className={iconClass} strokeWidth={promoActive ? 2.2 : 1.9} />
          </span>
          <span className={textClass}>Promo</span>
        </Link>

        <Link href={accountHref} prefetch={false} className={navClass(accountActive)}>
          <span className={accountActive ? "grid h-8 w-8 place-items-center rounded-2xl bg-[#e2fbf4]" : "grid h-8 w-8 place-items-center rounded-2xl"}>
            <UserRound className={iconClass} strokeWidth={accountActive ? 2.2 : 1.9} />
          </span>
          <span className={textClass}>Akun</span>
        </Link>
      </div>
    </section>
  );
}
