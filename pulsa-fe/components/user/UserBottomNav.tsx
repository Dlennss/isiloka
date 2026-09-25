"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { LucideIcon } from "lucide-react";
import { History, House, UserRound, WalletCards } from "lucide-react";

function navClass(active: boolean) {
  return active
    ? "flex h-[56px] min-w-0 flex-col items-center justify-center gap-1 rounded-[18px] bg-[#edfbf7] px-2 text-[#057b73]! ring-1 ring-[#d5f3ee] visited:text-[#057b73]!"
    : "flex h-[56px] min-w-0 flex-col items-center justify-center gap-1 rounded-[18px] px-2 text-[#7b8798]! transition visited:text-[#7b8798]! hover:bg-[#f4fbfa] hover:text-[#057b73]!";
}

function isActivePath(pathname: string, basePath: string) {
  return pathname === basePath || pathname.startsWith(`${basePath}/`);
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

export function UserBottomNav() {
  const pathname = usePathname() || "";
  const trxActive = isActivePath(pathname, "/user/transaksi");
  const saldoActive = isActivePath(pathname, "/user/saldo") || isActivePath(pathname, "/user/account/topup") || isActivePath(pathname, "/user/account/mutasi");
  const accountActive = isActivePath(pathname, "/user/account") && !saldoActive;
  const homeActive = isActivePath(pathname, "/user") && !trxActive && !accountActive && !saldoActive;

  return (
    <section className="isiloka-bottom-nav fixed bottom-3 left-1/2 z-[90] w-[calc(100%-2rem)] max-w-[398px] -translate-x-1/2 overflow-hidden rounded-[28px] border border-white/90 bg-white/96 shadow-[0_14px_38px_rgba(13,71,70,0.14)] backdrop-blur-2xl">
      <div className="grid grid-cols-4 gap-1 px-3 pb-[calc(0.55rem+env(safe-area-inset-bottom))] pt-2.5">
        <NavItem active={homeActive} href="/user" icon={House} label="Beranda" />
        <NavItem active={trxActive} href="/user/transaksi" icon={History} label="Riwayat" />
        <NavItem active={saldoActive} href="/user/saldo" icon={WalletCards} label="Saldo" />
        <NavItem active={accountActive} href="/user/account" icon={UserRound} label="Akun" />
      </div>
    </section>
  );
}
