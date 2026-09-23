"use client";

import type { ReactNode } from "react";

type OverviewStatCardProps = {
  title: string;
  value: string;
  subtitle?: string;
  icon?: ReactNode;
  tone?: "sky" | "emerald" | "amber" | "violet";
};

function toneClass(tone: OverviewStatCardProps["tone"]): string {
  if (tone === "emerald") {
    return "border-emerald-200 border-l-[#087e8b] bg-[linear-gradient(135deg,#ffffff_0%,#f1fff8_100%)]";
  }
  if (tone === "amber") {
    return "border-lime-200 border-l-[#65a30d] bg-[linear-gradient(135deg,#ffffff_0%,#f5ffe7_100%)]";
  }
  if (tone === "violet") {
    return "border-teal-200 border-l-[#0f766e] bg-[linear-gradient(135deg,#ffffff_0%,#ecfffb_100%)]";
  }
  return "border-cyan-200 border-l-[#0891b2] bg-[linear-gradient(135deg,#ffffff_0%,#effcff_100%)]";
}

function iconClass(tone: OverviewStatCardProps["tone"]): string {
  if (tone === "emerald") return "bg-[#e8fff4] text-[#13464b] ring-emerald-300";
  if (tone === "amber") return "bg-[#f5ffe7] text-[#3f6212] ring-lime-300";
  if (tone === "violet") return "bg-[#ecfffb] text-[#115e59] ring-teal-300";
  return "bg-[#effcff] text-[#155e75] ring-cyan-300";
}

export default function OverviewStatCard(props: OverviewStatCardProps) {
  const { title, value, subtitle, icon, tone = "sky" } = props;

  return (
    <div
      className={`rounded-[22px] border border-l-4 p-4 shadow-[0_14px_30px_rgba(6,78,59,0.08)] ${toneClass(tone)}`}
    >
      <div className="flex items-center justify-between gap-3">
        <div className="text-xs font-black uppercase tracking-[0.12em] text-[#13464b]">{title}</div>
        {icon ? <div className={`grid h-10 w-10 place-items-center rounded-2xl ring-2 ${iconClass(tone)}`}>{icon}</div> : null}
      </div>
      <div className="mt-2 text-2xl font-black tracking-tight text-[#071b14]">{value}</div>
      {subtitle ? <div className="mt-1 text-xs font-semibold text-[#315847]">{subtitle}</div> : null}
    </div>
  );
}
