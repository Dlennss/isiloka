import { ChevronDown, LogOut } from "lucide-react";
import { useState } from "react";
import { usePathname } from "next/navigation";
import { BrandLogo } from "./BrandLogo";
import { NavItem } from "./NavItem";
import { type NavSection } from "./nav";

type Props = {
  sections: NavSection[];
  onLogout: () => void;
  contextLabel?: string;
};

export function SidebarDesktop({ sections, onLogout, contextLabel = "Control Center" }: Props) {
  const pathname = usePathname();
  const [openSections, setOpenSections] = useState<Record<string, boolean>>({});
  return (
    <aside className="hidden w-72 shrink-0 self-start border-r border-emerald-950/20 bg-[radial-gradient(circle_at_20%_0%,rgba(190,242,100,0.18),transparent_30%),linear-gradient(180deg,#052e26_0%,#064e3b_46%,#047857_100%)] md:sticky md:top-0 md:flex md:h-screen md:flex-col">
      <div className="border-b border-white/15 px-5 py-5">
        <div className="flex justify-center">
          <BrandLogo variant="dark" />
        </div>
        <p className="mt-3 text-center text-[11px] font-black uppercase tracking-[0.18em] text-lime-100">
          {contextLabel}
        </p>
      </div>

      <div className="flex-1 overflow-y-auto px-4 py-5">
        {sections.map((section, idx) => {
          const key = section.title || `section-${idx}`;
          const active = section.items.some((item) => pathname === item.href || pathname.startsWith(item.href + "/"));
          const isOpen = section.title ? openSections[key] ?? active : true;

          return (
            <div key={key} className={idx === 0 ? "space-y-1" : "mt-5 border-t border-white/15 pt-4"}>
              {section.title ? (
                <button
                  type="button"
                  className={`mb-3 flex w-full items-center justify-between rounded-2xl border px-3 py-3 text-left text-[13px] font-black uppercase tracking-[0.10em] outline-none transition focus-visible:ring-4 focus-visible:ring-emerald-200 ${
                    active
                      ? "border-white bg-white text-[#13464b] shadow-[0_14px_28px_rgba(0,0,0,0.16)]"
                      : "border-white/15 bg-white/8 text-white hover:border-white/35 hover:bg-white/14 hover:text-white"
                  }`}
                  onClick={() => setOpenSections((prev) => ({ ...prev, [key]: !isOpen }))}
                >
                  <span>{section.title}</span>
                  <ChevronDown className={`h-4 w-4 transition ${isOpen ? "rotate-180 text-[#13464b]" : "text-lime-100"}`} />
                </button>
              ) : null}
              {isOpen ? (
                <div className="space-y-1.5 rounded-[24px] border border-white/10 bg-[#033b2e]/25 p-2 shadow-inner shadow-black/5">
                  {section.items.map((item) => (
                    <NavItem key={item.href} href={item.href} label={item.label} variant="dark" />
                  ))}
                </div>
              ) : null}
            </div>
          );
        })}
      </div>

      <div className="border-t border-white/15 p-4">
        <button
          type="button"
          onClick={onLogout}
          className="inline-flex w-full items-center justify-center gap-2 rounded-2xl border border-white/25 bg-white px-3 py-2.5 text-sm font-black text-rose-700 transition hover:bg-rose-50"
        >
          <LogOut className="h-4 w-4" />
          Logout
        </button>
      </div>
    </aside>
  );
}
