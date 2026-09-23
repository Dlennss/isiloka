"use client";

import { useMemo, useState } from "react";
import { Eye, EyeOff } from "lucide-react";

type Props = {
  name: string;
  saldo: number;
  commission: number;
};

function formatIDR(value: number) {
  return `Rp ${new Intl.NumberFormat("id-ID").format(Number(value || 0))}`;
}

function maskedCurrency(value: number, revealed: boolean) {
  if (revealed) return formatIDR(value);
  return "Rp*********";
}

export function UserAccountSummaryCard({ name, saldo, commission }: Props) {
  const [revealed, setRevealed] = useState<{ saldo: boolean; commission: boolean }>({
    saldo: false,
    commission: false,
  });

  const stats = useMemo(
    () => [
      {
        key: "saldo",
        label: "Saldo",
        value: saldo,
        ring: "ring-sky-500/20",
        cardBg: "from-[#087e8b] via-[#087e8b] to-[#087e8b]",
        glow: "from-white/16 via-white/6 to-transparent",
        tone: "text-white/78",
      },
      {
        key: "commission",
        label: "Komisi",
        value: commission,
        ring: "ring-orange-500/20",
        cardBg: "from-[#ff9f1c] via-[#ff7a18] to-[#ff4d6d]",
        glow: "from-white/16 via-white/6 to-transparent",
        tone: "text-white/80",
      },
    ],
    [commission, saldo],
  );

  return (
    <section className="space-y-3">
      <p className="text-sm font-medium tracking-[0.01em] text-slate-700">Halo, {name}</p>

      <div className="grid grid-cols-2 gap-3">
        {stats.map((item) => {
          const isShown = item.key === "saldo" ? revealed.saldo : revealed.commission;

          return (
            <div
              key={item.key}
              className={`relative aspect-video overflow-hidden rounded-md bg-linear-to-br ${item.cardBg} px-3 py-3 shadow-[0_16px_34px_rgba(15,23,42,0.14)] ring-1 ${item.ring}`}
            >
              <div className={`absolute -right-8 -top-8 z-0 h-24 w-24 rounded-full bg-linear-to-br ${item.glow}`} />
              <div className="absolute right-5 top-5 z-0 h-14 w-14 rounded-full border border-white/12 bg-white/6" />
              <div className="absolute -bottom-6 -left-6 z-0 h-16 w-20 rounded-full border border-white/8 bg-white/6 blur-[2px]" />
              <div className="relative z-10">
                <div className="min-w-0">
                  <p className={`text-[9px] font-semibold uppercase tracking-[0.18em] ${item.tone}`}>{item.label}</p>
                  <div className="relative mt-2 pr-7">
                    <p className="min-w-0 whitespace-nowrap text-sm font-bold tracking-tighter text-white sm:text-[12px]">
                      {maskedCurrency(item.value, isShown)}
                    </p>
                    <button
                      type="button"
                      onClick={() =>
                        setRevealed((current) => ({
                          ...current,
                          [item.key]: !current[item.key as keyof typeof current],
                        }))
                      }
                      className="absolute right-0 top-1/2 inline-flex h-5 w-5 -translate-y-1/2 items-center justify-center rounded-full text-white/70 transition hover:text-white"
                      aria-label={isShown ? `Sembunyikan ${item.label}` : `Tampilkan ${item.label}`}
                    >
                      {isShown ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
                    </button>
                  </div>
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </section>
  );
}
