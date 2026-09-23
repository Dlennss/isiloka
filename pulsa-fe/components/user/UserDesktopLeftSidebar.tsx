import Link from "next/link";
import { ChevronDown, Headset, UserRound } from "lucide-react";
import type { UserSession } from "@/components/user/types";

type UserDesktopLeftSidebarProps = {
  user: UserSession | null;
};

export function UserDesktopLeftSidebar({ user }: UserDesktopLeftSidebarProps) {
  return (
    <aside className="hidden lg:mt-9 lg:block lg:space-y-4">
      <div className="rounded-3xl bg-white p-5 shadow-[0_8px_22px_rgba(15,23,42,0.13)]">
        <p className="text-xs font-semibold tracking-wide text-neutral-500 uppercase">Akun</p>
        <p className="mt-2 text-lg font-bold text-neutral-800">{user?.name || "User"}</p>
        <p className="truncate text-sm text-neutral-500">{user?.email || "-"}</p>
        <div className="mt-4 rounded-xl bg-neutral-100 p-3">
          <p className="text-xs text-neutral-500">Metode Login</p>
          <p className="text-sm font-semibold text-neutral-700">Google</p>
        </div>

        <Link href="/user/account" className="mt-4 inline-flex w-full items-center justify-center gap-2 rounded-xl bg-sky-600 px-3 py-2 text-sm font-semibold text-white hover:bg-sky-700">
          <UserRound className="h-4 w-4" />
          Buka Halaman Akun
        </Link>
      </div>

      <div className="rounded-3xl bg-white p-4 shadow-[0_8px_22px_rgba(15,23,42,0.13)]">
        <p className="text-sm font-semibold text-neutral-700">Bantuan Cepat</p>
        <button type="button" className="mt-3 flex w-full items-center justify-between rounded-xl bg-neutral-100 px-3 py-2 text-sm font-medium text-neutral-700">
          <span className="flex items-center gap-2">
            <Headset className="h-4 w-4 text-[#2d8fdc]" />
            Hubungi CS
          </span>
          <ChevronDown className="h-4 w-4 -rotate-90 text-neutral-400" />
        </button>
      </div>
    </aside>
  );
}
