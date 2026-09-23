"use client";

import type { ReactNode } from "react";
import { Inbox } from "lucide-react";

type EmptyStateProps = {
  title: string;
  description?: string;
  icon?: ReactNode;
  className?: string;
};

export function EmptyState({
  title,
  description,
  icon,
  className = "",
}: EmptyStateProps) {
  return (
    <div
      className={`grid min-h-55 place-items-center rounded-2xl border border-dashed border-white/12 bg-linear-to-br from-slate-950/60 via-slate-900/40 to-cyan-950/10 p-6 text-center shadow-[inset_0_1px_0_rgba(255,255,255,0.03)] ${className}`}
    >
      <div className="max-w-sm">
        <div className="mx-auto grid h-14 w-14 place-items-center rounded-2xl border border-cyan-400/15 bg-cyan-400/8 text-cyan-200 shadow-[0_14px_30px_-22px_rgba(34,211,238,0.55)]">
          {icon ?? <Inbox className="h-6 w-6" />}
        </div>
        <div className="mt-4 text-base font-semibold text-white">{title}</div>
        {description ? <div className="mt-1 text-sm text-white/55">{description}</div> : null}
      </div>
    </div>
  );
}
