"use client";

import * as React from "react";
import { Eye, EyeOff, Headset } from "lucide-react";
import Link from "next/link";
import type { UserProfile, UserSession } from "@/components/user/types";
import { UserLogoutButton } from "@/components/user/UserLogoutButton";

type UserHeroProps = {
  user: UserSession | null;
  profile?: UserProfile | null;
};

function formatRupiah(value: number) {
  return `Rp ${new Intl.NumberFormat("id-ID").format(value || 0)}`;
}

export function UserHero({ user, profile }: UserHeroProps) {
  const [showSaldo, setShowSaldo] = React.useState(false);
  const isLoggedIn = Boolean(user);

  return (
    <section className="sticky top-0 z-30 border-b border-white/10 bg-linear-to-r from-[#087e8b] via-[#087e8b] to-[#087e8b] px-5 pb-2 pt-2 text-white shadow-[0_10px_24px_rgba(15,111,203,0.20)] backdrop-blur-sm">
      <div className="flex items-center justify-between gap-3">
        <div className="min-w-0 flex-1 pr-2">
          {isLoggedIn ? (
            <>
              <p className="leading-none font-bold tracking-tight">
                Hai, {profile?.nama?.trim() || user?.name || "User"}!
              </p>
              <div className="mt-2 flex items-center gap-2 text-sm text-white/90">
                <span className="font-medium">{showSaldo ? formatRupiah(Number(profile?.saldo || 0)) : "Rp *****"}</span>
                <button
                  type="button"
                  onClick={() => setShowSaldo((v) => !v)}
                  className="grid h-4 w-4 place-items-center rounded-full bg-white/10 text-white/85 transition hover:bg-white/20"
                  aria-label={showSaldo ? "Sembunyikan saldo" : "Tampilkan saldo"}
                >
                  {showSaldo ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </div>
            </>
          ) : (
            <div className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/10 px-2.5 py-1 backdrop-blur-sm">
              <span className="grid h-6 w-6 place-items-center rounded-full bg-white text-[10px] font-black tracking-[0.18em] text-sky-700">
                P
              </span>
              <span className="text-sm font-black tracking-[0.2em] text-white/95">ISILOKA</span>
            </div>
          )}
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
          {isLoggedIn ? (
            <UserLogoutButton
              iconOnly
              className="h-7 w-7 rounded-full border border-white/60 bg-white/10 text-white hover:bg-white/20"
            />
          ) : (
            <Link
              href="/login"
              className="inline-flex h-7 items-center rounded-full border border-white/60 bg-white/10 px-3 text-xs font-semibold text-white hover:bg-white/20"
            >
              Masuk
            </Link>
          )}
        </div>
      </div>
    </section>
  );
}
