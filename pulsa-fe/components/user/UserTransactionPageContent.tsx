"use client";

import { useState } from "react";
import { CalendarDays, ReceiptText, Search } from "lucide-react";
import type { UserAppOrder } from "@/components/user/types";
import { UserTransactionHistoryList } from "@/components/user/UserTransactionHistoryList";

type UserTransactionPageContentProps = {
  initialItems: UserAppOrder[];
  initialHasNextPage: boolean;
  status: string;
  authToken: string;
};

const RANGE_FILTERS = ["Semua", "Hari ini", "Kemarin", "7 Hari"] as const;

function formatDateDisplay(value: string) {
  if (!value) return "dd/mm/yyyy";
  const [year, month, day] = value.split("-");
  if (!year || !month || !day) return "dd/mm/yyyy";
  return `${day}/${month}/${year}`;
}

export function UserTransactionPageContent({
  initialItems,
  initialHasNextPage,
  status,
  authToken,
}: UserTransactionPageContentProps) {
  const [query, setQuery] = useState("");
  const [selectedRange, setSelectedRange] = useState("Semua");
  const [selectedDate, setSelectedDate] = useState("");

  return (
    <section className="space-y-4">
      <div>
        <h1 className="text-xl font-black tracking-tight text-slate-950">Riwayat Transaksi</h1>
        <p className="mt-1 text-[11px] font-semibold text-slate-500">Pantau semua transaksi Isiloka</p>
      </div>

      <label className="flex h-13 items-center gap-3 rounded-[18px] border border-slate-200 bg-white px-4 shadow-[0_10px_24px_rgba(15,23,42,0.04)]">
        <Search className="h-5 w-5 shrink-0 text-slate-400" strokeWidth={2.1} />
        <input
          type="search"
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="Cari nomor atau ID transaksi"
          className="h-full min-w-0 flex-1 bg-transparent text-sm font-semibold text-slate-700 outline-none placeholder:text-slate-400"
        />
      </label>

      <div className="flex gap-2 overflow-x-auto pb-1 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
        {RANGE_FILTERS.map((label) => (
          <button
            key={label}
            type="button"
            onClick={() => {
              setSelectedRange(label);
              if (label !== "Semua") setSelectedDate("");
            }}
            className={
              selectedRange === label
                ? "h-9 shrink-0 rounded-full bg-[#087e8b] px-4 text-xs font-black text-white shadow-[0_10px_20px_rgba(4,120,87,0.20)]"
                : "h-9 shrink-0 rounded-full border border-slate-200 bg-white px-4 text-xs font-black text-slate-500 transition hover:border-emerald-200 hover:text-[#087e8b]"
            }
          >
            {label}
          </button>
        ))}
      </div>

      <label className="relative flex h-13 cursor-pointer items-center gap-3 rounded-[16px] border border-slate-200 bg-white px-4 text-slate-500 shadow-[0_10px_24px_rgba(15,23,42,0.04)]">
        <CalendarDays className="h-4.5 w-4.5 text-[#087e8b]" strokeWidth={2.1} />
        <span className="text-xs font-bold text-slate-500">Pilih tanggal</span>
        <span className={selectedDate ? "ml-auto text-xs font-black text-slate-700" : "ml-auto text-xs font-bold text-slate-400"}>
          {formatDateDisplay(selectedDate)}
        </span>
        <input
          type="date"
          value={selectedDate}
          onChange={(event) => {
            setSelectedDate(event.target.value);
            if (event.target.value) setSelectedRange("Semua");
          }}
          aria-label="Pilih tanggal transaksi"
          className="absolute inset-0 cursor-pointer opacity-0"
        />
      </label>

      <UserTransactionHistoryList
        initialItems={initialItems}
        initialHasNextPage={initialHasNextPage}
        status={status}
        authToken={authToken}
        searchQuery={query}
        selectedRange={selectedRange}
        selectedDate={selectedDate}
        emptyIcon={<ReceiptText className="h-12 w-12 text-slate-400" strokeWidth={1.8} />}
      />
    </section>
  );
}
