"use client";

import { useState } from "react";
import { Eye, EyeOff } from "lucide-react";

type Props = {
  name?: string;
  saldo: number;
  label?: string;
};

function formatIDR(value: number) {
  return `Rp ${new Intl.NumberFormat("id-ID").format(Number(value || 0))}`;
}

export function UserBalanceHeader({ name, saldo, label }: Props) {
  const [revealed, setRevealed] = useState(false);
  const displayLabel = label || (name ? `Halo, ${name}` : "Saldo");

  return (
    <section className="flex items-center justify-between gap-3">
      <div className="min-w-0">
        <p className="truncate text-[13px] font-medium tracking-[0.01em] text-sky-900/75">{displayLabel}</p>
      </div>

      <div className="flex shrink-0 items-center gap-2">
        <p className="text-[13px] font-semibold tracking-[-0.01em] text-sky-950/85 sm:text-sm">
          {revealed ? formatIDR(saldo) : "Rp *******"}
        </p>
        <button
          type="button"
          onClick={() => setRevealed((value) => !value)}
          className="inline-flex h-6 w-6 items-center justify-center "
          aria-label={revealed ? "Sembunyikan saldo" : "Tampilkan saldo"}
        >
          {revealed ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
        </button>
      </div>
    </section>
  );
}
