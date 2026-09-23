"use client";

import * as React from "react";

type QuickProductOptionItem = {
  id: number;
  title: React.ReactNode;
  subtitle?: React.ReactNode;
};

type QuickProductOptionGridProps = {
  items: QuickProductOptionItem[];
  selectedId?: number | null;
  onSelect: (id: number) => void;
  columns?: 1 | 2;
  variant?: "default" | "pulsa";
};

export function QuickProductOptionGrid({
  items,
  selectedId,
  onSelect,
  columns = 2,
  variant = "default",
}: QuickProductOptionGridProps) {
  if (columns === 1) {
    return (
      <div className="space-y-2">
        {items.map((item) => {
          const selected = selectedId === item.id;
          return (
            <button
              key={item.id}
              type="button"
              onClick={() => onSelect(item.id)}
              className={`relative w-full overflow-hidden rounded-md px-5 py-3 text-left ${
                selected
                  ? "bg-[#0084D1] shadow-[0_18px_36px_rgba(0,132,209,0.28)]"
                  : "bg-[#1491db] shadow-[0_14px_30px_rgba(0,132,209,0.22)]"
              }`}
            >
              <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(255,255,255,0.22),transparent_30%),linear-gradient(180deg,rgba(255,255,255,0.06),rgba(255,255,255,0.02))]" />
              <div className="absolute inset-0 opacity-30 bg-[repeating-radial-gradient(circle_at_0_100%,rgba(255,255,255,0.35)_0,rgba(255,255,255,0.35)_2px,transparent_2px,transparent_12px)] bg-size-[160%_120%]" />

              <div className="relative flex min-h-18 flex-col justify-between gap-2">
                <div className="min-w-0">
                  <div className="text-white **:text-inherit text-lg font-semibold">{item.title}</div>
                </div>
                <div className="min-w-0">
                  {item.subtitle ? <div className="text-sm font-semibold leading-none tracking-tight text-white/72">{item.subtitle}</div> : null}
                </div>
              </div>
            </button>
          );
        })}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-2 gap-2">
      {items.map((item) => {
        const selected = selectedId === item.id;
        const isPulsaCard = variant === "pulsa";
        return (
          <button
            key={item.id}
            type="button"
            onClick={() => onSelect(item.id)}
            className={`relative overflow-hidden rounded-md px-4 py-4 ${
              selected
                ? "bg-[#0084D1] shadow-[0_18px_36px_rgba(0,132,209,0.28)]"
                : "bg-[#1491db] shadow-[0_14px_30px_rgba(0,132,209,0.22)]"
            }`}
          >
            <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(255,255,255,0.22),transparent_30%),linear-gradient(180deg,rgba(255,255,255,0.06),rgba(255,255,255,0.02))]" />
            <div className="absolute inset-0 opacity-30 bg-[repeating-radial-gradient(circle_at_0_100%,rgba(255,255,255,0.35)_0,rgba(255,255,255,0.35)_2px,transparent_2px,transparent_12px)] bg-size-[170%_130%]" />

            <div
              className={`relative ${
                isPulsaCard
                  ? "flex min-h-18 flex-col justify-between"
                  : "flex min-h-20 flex-col justify-between gap-4"
              }`}
            >
              <div
                className={`min-w-0 text-white **:text-inherit ${
                  isPulsaCard ? "flex flex-1 items-center justify-center text-center text-sm font-semibold" : "text-sm font-semibold"
                }`}
              >
                {item.title}
              </div>
              {item.subtitle ? (
                <div
                  className={`font-bold leading-none tracking-tight text-white ${
                    isPulsaCard ? "text-right text-[11px] opacity-65" : "text-sm opacity-80"
                  }`}
                >
                  {item.subtitle}
                </div>
              ) : null}
            </div>
          </button>
        );
      })}
    </div>
  );
}
