import Link from "next/link";
import { Headset, Zap } from "lucide-react";

type AppTopHeaderProps = {
  isLoggedIn?: boolean;
  userName?: string | null;
  saldo?: number | null;
  role?: string | null;
};

export function AppTopHeader({ isLoggedIn = false, userName, saldo, role }: AppTopHeaderProps) {
  const normalizedRole = String(role || "").trim().toLowerCase();
  const isRetailLoggedIn = isLoggedIn && (normalizedRole === "user" || normalizedRole === "agent" || normalizedRole === "master");
  const homeHref = isRetailLoggedIn ? "/user" : "/";
  void userName;
  void saldo;

  return (
    <header className="brand-app-header sticky top-0 z-30 overflow-hidden bg-[#13464b] px-3 pb-3 pt-2 text-white shadow-[0_16px_34px_rgba(5,46,38,0.22)]">
      <div className="pointer-events-none absolute -right-10 -top-14 h-32 w-32 rounded-full bg-lime-300/25 blur-2xl" />
      <div className="pointer-events-none absolute -bottom-10 left-20 h-20 w-40 rotate-[-10deg] bg-emerald-400/15 blur-2xl" />
      <div className="pointer-events-none absolute inset-x-0 bottom-0 h-px bg-linear-to-r from-transparent via-lime-300/80 to-transparent" />

      <div className="relative flex h-14 items-center justify-between gap-3">
        <div className="flex min-w-0 flex-1 items-center">
          <Link
            href={homeHref}
            prefetch={false}
            className="flex min-w-0 items-center gap-2.5"
            aria-label="Isiloka"
          >
            <span className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-white shadow-[0_10px_22px_rgba(163,230,53,0.22)] ring-1 ring-lime-200/80">
              <img src="/brand/icon.svg" alt="" width={40} height={40} className="h-10 w-10" />
            </span>
            <span className="min-w-0">
              <span className="block text-[22px] font-black italic leading-5 tracking-tight">
                <span className="brand-wordmark">Isiloka</span>
              </span>
              <span className="mt-1 block text-[10px] font-bold uppercase tracking-[0.18em] text-lime-100/85">
                Isi hari, dari sini
              </span>
            </span>
          </Link>
        </div>

        <div className="flex items-center gap-2">
          <a
            href="#"
            target="_blank"
            rel="noreferrer"
            className="grid h-10 w-10 shrink-0 place-items-center rounded-2xl border border-white/15 bg-white/10 text-lime-100 shadow-sm transition hover:bg-white/18"
            aria-label="Hubungi bantuan via WhatsApp"
          >
            <Headset className="h-4 w-4" />
          </a>
        </div>
      </div>
    </header>
  );
}
