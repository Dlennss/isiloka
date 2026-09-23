"use client";

export function SiteShell({ children }: { children: React.ReactNode }) {
  return <div className="box-border min-h-[calc(100dvh-65px)] pb-24">{children}</div>;
}
