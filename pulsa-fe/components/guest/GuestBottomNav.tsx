"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { LucideIcon } from "lucide-react";
import { History, House, UserRound } from "lucide-react";

function navClass(active: boolean) {
  return active
    ? "flex h-[56px] min-w-0 flex-col items-center justify-center gap-1 rounded-[18px] bg-[#edfbf7] px-2 text-[#057b73]! ring-1 ring-[#d5f3ee] visited:text-[#057b73]!"
    : "flex h-[56px] min-w-0 flex-col items-center justify-center gap-1 rounded-[18px] px-2 text-[#7b8798]! transition visited:text-[#7b8798]! hover:bg-[#f4fbfa] hover:text-[#057b73]!";
}

const iconClass = "h-5 w-5";
const textClass = "text-[10.5px] font-bold leading-none";

type NavItemProps = {
  active: boolean;
  href: string;
  icon: LucideIcon;
  label: string;
};

function NavItem({ active, href, icon: Icon, label }: NavItemProps) {
  return (
    <Link href={href} prefetch={false} className={navClass(active)}>
      <Icon className={iconClass} strokeWidth={active ? 2.35 : 2} />
      <span className={textClass}>{label}</span>
    </Link>
  );
}

type GuestBottomNavProps = {
  isLoggedIn?: boolean;
};

export function GuestBottomNav({ isLoggedIn = false }: GuestBottomNavProps) {
  const pathname = usePathname() || "";
  const homeHref = isLoggedIn ? "/user" : "/";
  const historyHref = isLoggedIn ? "/user/transaksi" : "/transaksi";
  const homeActive = pathname === "/" || pathname === "/user";
  const historyActive = pathname.startsWith("/transaksi") || pathname.startsWith("/user/transaksi");
  const accountHref = isLoggedIn ? "/user/account" : "/login";
  const accountActive = isLoggedIn
    ? pathname.startsWith("/user/account")
    : pathname.startsWith("/login");

  return (
    <section className="isiloka-bottom-nav fixed bottom-3 left-1/2 z-[90] w-[calc(100%-2rem)] max-w-[398px] -translate-x-1/2 overflow-hidden rounded-[28px] border border-white/90 bg-white/96 shadow-[0_14px_38px_rgba(13,71,70,0.14)] backdrop-blur-2xl">
      <div className="grid grid-cols-3 gap-1 px-3 pb-[calc(0.55rem+env(safe-area-inset-bottom))] pt-2.5">
        <NavItem active={homeActive} href={homeHref} icon={House} label="Beranda" />
        <NavItem active={historyActive} href={historyHref} icon={History} label="Riwayat" />
        <NavItem active={accountActive} href={accountHref} icon={UserRound} label="Akun" />
      </div>
    </section>
  );
}
