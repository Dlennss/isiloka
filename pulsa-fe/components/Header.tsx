"use client";

import React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Menu, X, Zap } from "lucide-react";

const navItems = [
  { href: "/", label: "Beranda" },
  { href: "/tentang", label: "Tentang" },
  { href: "/docs", label: "Dokumentasi API" },
];

export default function Header() {
  const [open, setOpen] = React.useState(false);
  const pathname = usePathname();
  const inDashboard = pathname?.startsWith("/dashboard");

  if (inDashboard) return null;

  return (
    <header className="brand-app-header sticky top-0 z-50 overflow-hidden bg-[#13464b] text-white shadow-[0_16px_34px_rgba(5,46,38,0.22)]">
      <div className="pointer-events-none absolute -right-16 -top-20 h-44 w-44 rounded-full bg-lime-300/25 blur-2xl" />
      <div className="pointer-events-none absolute inset-x-0 bottom-0 h-px bg-linear-to-r from-transparent via-lime-300/80 to-transparent" />
      <div className="mx-auto flex h-18 w-full max-w-7xl items-center justify-between gap-3 px-4">
        <Link href="/" className="flex items-center gap-3" aria-label="Isiloka">
          <span className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-white shadow-[0_10px_22px_rgba(163,230,53,0.22)] ring-1 ring-lime-200/80">
            <img src="/brand/icon.svg" alt="" width={40} height={40} className="h-10 w-10" />
          </span>
          <span>
            <span className="block text-2xl font-black italic leading-5 tracking-tight">
              <span className="brand-wordmark">Isiloka</span>
            </span>
            <span className="mt-1 block text-[10px] font-bold uppercase tracking-[0.18em] text-lime-100/85">
              Isi hari, dari sini
            </span>
          </span>
        </Link>

        <div className="ml-auto hidden items-center gap-4 lg:flex">
          <nav className="flex items-center gap-6">
            {navItems.map((item) => (
              <Link key={item.label} href={item.href} className="text-sm font-bold text-white/75 transition hover:text-lime-200">
                {item.label}
              </Link>
            ))}
          </nav>
          <div className="h-8 w-px bg-white/15" />
          <Link href="/login" className="inline-flex items-center gap-2 rounded-2xl bg-linear-to-r from-[#f6b49d] to-[#22c55e] px-6 py-2 text-sm font-black shadow-[0_10px_22px_rgba(163,230,53,0.24)] transition hover:brightness-105">
            <Zap className="h-4 w-4 fill-[#13464b] text-[#13464b]" />
            <span className="text-[#13464b]">Masuk</span>
          </Link>
        </div>

        <button
          type="button"
          className="rounded-2xl border border-white/15 bg-white/10 p-2 text-lime-100 lg:hidden"
          onClick={() => setOpen((value) => !value)}
          aria-label="Toggle menu"
        >
          {open ? <X className="h-6 w-6" /> : <Menu className="h-6 w-6" />}
        </button>
      </div>

      {open ? (
        <nav className="space-y-2 border-t border-white/10 bg-[#13464b] px-4 py-4 lg:hidden">
          {navItems.map((item) => (
            <Link
              key={item.label}
              href={item.href}
              className="block rounded-2xl bg-white/10 px-3 py-2 text-sm font-bold text-white/85"
              onClick={() => setOpen(false)}
            >
              {item.label}
            </Link>
          ))}
          <div className="pt-2">
            <Link
              href="/login"
              className="block rounded-2xl bg-linear-to-r from-[#f6b49d] to-[#22c55e] px-3 py-2 text-center text-sm font-black text-[#13464b]"
              onClick={() => setOpen(false)}
            >
              Masuk
            </Link>
          </div>
        </nav>
      ) : null}
    </header>
  );
}
