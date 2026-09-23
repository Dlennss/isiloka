"use client";

import Image from "next/image";
import Link from "next/link";
import { Headset } from "lucide-react";

type GuestHeroProps = {
  isLoggedIn?: boolean;
};

export function GuestHero({ isLoggedIn = false }: GuestHeroProps) {
  return (
    <section className="sticky top-0 z-30 border-b border-white/10 bg-linear-to-r from-[#13464b] via-[#087e8b] to-[#10b981] px-5 pb-2 pt-2 text-white shadow-[0_10px_24px_rgba(6,78,59,0.22)] backdrop-blur-sm">
      <div className="flex items-center justify-between gap-3">
        <div className="min-w-0 flex-1 pr-2">
          <Link href="/" className="inline-flex h-12 max-w-[58vw] items-center rounded-lg bg-white px-2.5 shadow-[0_8px_18px_rgba(6,78,59,0.22)]">
            <Image
              src="/images/logo-pulsakilat-header.svg"
              alt="Isiloka"
              width={210}
              height={48}
              className="h-10 w-auto max-w-full object-contain"
              priority
            />
          </Link>
        </div>

        <div className="flex items-center gap-2">
          <a
            href="https://wa.me/6282219107558"
            target="_blank"
            rel="noreferrer"
            className="grid h-7 w-7 place-items-center rounded-full border border-white/60 bg-white/10 text-white transition hover:bg-white/20"
            aria-label="Hubungi bantuan via WhatsApp"
          >
            <Headset className="h-4 w-4" />
          </a>
          <Link
            href={isLoggedIn ? "/user/account" : "/login"}
            className="inline-flex h-7 items-center rounded-full border border-white/60 bg-white/10 px-3 text-xs font-semibold text-white hover:bg-white/20"
          >
            {isLoggedIn ? "Akun" : "Masuk"}
          </Link>
        </div>
      </div>
    </section>
  );
}
