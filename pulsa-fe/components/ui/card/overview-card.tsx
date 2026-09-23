import type { ReactNode } from "react";

type OverviewCardTone = "emerald" | "sky" | "amber" | "rose" | "slate";

const toneMap: Record<OverviewCardTone, string> = {
  emerald: "from-emerald-500/20 via-emerald-400/10 to-transparent",
  sky: "from-sky-500/20 via-sky-400/10 to-transparent",
  amber: "from-amber-500/20 via-amber-400/10 to-transparent",
  rose: "from-rose-500/20 via-rose-400/10 to-transparent",
  slate: "from-slate-400/15 via-slate-300/10 to-transparent",
};

export function OverviewCard({
  title,
  value,
  sub,
  icon,
  tone = "slate",
}: {
  title: string;
  value: string;
  sub?: string;
  icon?: ReactNode;
  tone?: OverviewCardTone;
}) {
  return (
    <div className="relative overflow-hidden rounded-2xl border border-white/10 bg-card/80 p-4 shadow-[0_14px_34px_-20px_rgba(0,0,0,0.75)]">
      <div className={`pointer-events-none absolute inset-0 bg-linear-to-br ${toneMap[tone]}`} />
      <div className="relative flex items-start justify-between">
        <div>
          <div className="text-xs font-medium tracking-wide text-white/60">{title}</div>
          <div className="mt-1 text-2xl font-semibold leading-tight">{value}</div>
          {sub ? <div className="mt-1 text-xs text-white/50">{sub}</div> : null}
        </div>
        {icon ? <div className="rounded-xl border border-white/15 bg-white/10 p-2 text-white/85">{icon}</div> : null}
      </div>
    </div>
  );
}
