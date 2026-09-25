"use client";

import { usePathname } from "next/navigation";

export function SiteFrame({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const isHome = pathname === "/" || pathname === "/user";

  if (isHome) {
    return (
      <div className="min-h-dvh bg-[#eaf8f5] text-neutral-900 md:grid md:place-items-start md:py-4 [&_.site-frame-shell]:max-w-[430px] [&_.site-frame-shell]:overflow-visible [&_.site-frame-shell]:bg-transparent [&_.site-frame-shell]:shadow-none md:[&_.site-frame-shell]:w-[430px] md:[&_.site-frame-shell]:rounded-[32px] md:[&_.site-frame-shell]:border-0">
        {children}
      </div>
    );
  }

  return (
    <div className="min-h-dvh bg-[radial-gradient(circle_at_50%_10%,#effffc_0%,#e6f8f5_38%,#f7fffc_100%)] text-neutral-900 md:grid md:place-items-start md:py-4">
      {children}
    </div>
  );
}
