import { BrandLogo } from "./BrandLogo";
import { Menu } from "lucide-react";

type Props = {
  onOpenMenu: () => void;
};

export function HeaderMobile({ onOpenMenu }: Props) {
  return (
    <header className="sticky top-0 z-20 border-b border-emerald-100 bg-white/90 px-4 py-3 shadow-[0_10px_24px_rgba(6,78,59,0.06)] backdrop-blur md:hidden">
      <div className="relative flex items-center justify-end gap-3">
        <div className="pointer-events-none absolute inset-x-0 flex justify-center">
          <BrandLogo />
        </div>
        <button
          type="button"
          className="inline-flex h-11 w-11 items-center justify-center rounded-2xl border border-emerald-100 bg-emerald-50 text-[#13464b] shadow-sm"
          onClick={onOpenMenu}
          aria-label="Open menu"
        >
          <Menu className="h-5 w-5" />
        </button>
      </div>
    </header>
  );
}

