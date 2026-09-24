"use client";

import Link from "next/link";
import Image from "next/image";
import { usePathname } from "next/navigation";
import { Bell, UserRound } from "lucide-react";

type AppTopHeaderProps = {
  isLoggedIn?: boolean;
  userName?: string | null;
  saldo?: number | null;
  role?: string | null;
};

export function AppTopHeader({ isLoggedIn = false, userName, saldo, role }: AppTopHeaderProps) {
  const pathname = usePathname();
  const normalizedRole = String(role || "").trim().toLowerCase();
  const isRetailLoggedIn = isLoggedIn && (normalizedRole === "user" || normalizedRole === "agent" || normalizedRole === "master");
  const homeHref = isRetailLoggedIn ? "/user" : "/";
  void userName;
  void saldo;

  if (pathname === "/") return null;

  return (
    <header className="brand-app-header sticky top-0 z-30 bg-[#f7fffc]/92 px-5 pb-2.5 pt-5 text-[#073b43] backdrop-blur-xl">
      <div className="flex h-14 items-center justify-between gap-3">
        <div className="flex min-w-0 flex-1 items-center">
          <Link
            href={homeHref}
            prefetch={false}
            className="flex min-w-0 items-center gap-3"
            aria-label="Isiloka"
          >
            <Image src="/isiloka-concept/logo_symbol.png" alt="" width={71} height={73} className="h-12 w-auto shrink-0" priority />
            <Image src="/isiloka-concept/logo_wordmark_tagline.png" alt="Isiloka - Isi hari, dari sini" width={196} height={73} className="h-12 w-auto min-w-0 object-contain" priority />
          </Link>
        </div>

        <div className="flex shrink-0 items-center gap-2">
          <Link
            href={isLoggedIn ? "/user/transaksi" : "/login"}
            prefetch={false}
            className="relative grid h-11 w-11 place-items-center rounded-[17px] bg-white text-[#087e8b]! shadow-[0_12px_28px_rgba(12,74,76,0.08)] ring-1 ring-teal-900/5 transition visited:text-[#087e8b]! hover:-translate-y-0.5"
            aria-label="Lihat transaksi"
          >
            <Bell className="h-5 w-5" strokeWidth={2.4} />
            <span className="absolute right-2.5 top-2.5 h-2.5 w-2.5 rounded-full border-2 border-white bg-[#ff315f]" />
          </Link>
          <Link
            href={isLoggedIn ? "/user/account" : "/login"}
            prefetch={false}
            className={
              isLoggedIn
                ? "grid h-11 w-11 place-items-center overflow-hidden rounded-full bg-white text-sm font-black text-[#087e8b]! shadow-[0_12px_28px_rgba(12,74,76,0.08)] ring-1 ring-teal-900/5 transition visited:text-[#087e8b]! hover:-translate-y-0.5"
                : "flex h-11 items-center gap-1.5 rounded-[17px] bg-white px-3 text-[13px] font-black text-[#087e8b]! shadow-[0_12px_28px_rgba(12,74,76,0.08)] ring-1 ring-teal-900/5 transition visited:text-[#087e8b]! hover:-translate-y-0.5"
            }
            aria-label={isLoggedIn ? "Akun" : "Masuk"}
          >
            {isLoggedIn ? (
              <UserRound className="h-5 w-5" strokeWidth={2.4} />
            ) : (
              <>
                <UserRound className="h-4 w-4" strokeWidth={2.4} />
                <span>Masuk</span>
              </>
            )}
          </Link>
        </div>
      </div>
    </header>
  );
}
