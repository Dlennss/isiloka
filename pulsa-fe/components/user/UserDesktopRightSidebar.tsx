import { Clock3 } from "lucide-react";
import type { UserSession } from "@/components/user/types";

type UserDesktopRightSidebarProps = {
  user: UserSession | null;
  transactions: string[];
};

export function UserDesktopRightSidebar({ user, transactions }: UserDesktopRightSidebarProps) {
  return (
    <aside className="hidden lg:mt-9 lg:block lg:space-y-4">
      <div className="rounded-3xl bg-white p-4 shadow-[0_8px_22px_rgba(15,23,42,0.13)]">
        <p className="text-sm font-semibold text-neutral-700">Riwayat Transaksi</p>
        <div className="mt-3 space-y-2">
          {transactions.map((item) => (
            <div key={item} className="rounded-xl bg-neutral-100 px-3 py-2">
              <p className="text-sm font-medium text-neutral-700">{item}</p>
              <p className="mt-1 flex items-center gap-1 text-xs text-neutral-500">
                <Clock3 className="h-3.5 w-3.5" />
                Sedang diproses
              </p>
            </div>
          ))}
        </div>
      </div>

      <div className="rounded-3xl bg-white p-4 shadow-[0_8px_22px_rgba(15,23,42,0.13)]">
        <p className="text-sm font-semibold text-neutral-700">Status Akun</p>
        <p className="mt-2 text-sm text-neutral-500">Role: {user?.role || "user"}</p>
        <p className="text-sm text-[#2d8fdc]">Aktif</p>
      </div>
    </aside>
  );
}
