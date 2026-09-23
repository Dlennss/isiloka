'use client';

import { useCallback, useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import {
  CalendarDays,
  CircleCheck,
  CircleX,
  Clock3,
  Wallet,
  ReceiptText,
  Trash2,
  Loader2,
  ScrollText,
  ChevronRight,
  RefreshCw,
  Search,
} from 'lucide-react';
import {
  loadGuestTransactions,
  saveGuestTransaction,
  removeGuestTransaction,
  type GuestTransactionEntry,
} from '@/lib/guest-transaction-storage';
import { getOrderByInvoice } from '@/lib/api.transactions';

type SyncedGuestOrder = {
  invoice_id?: string;
  dest?: string;
  produk_nama_snapshot?: string;
  dibuat_pada?: string;
  diubah_pada?: string;
  status?: string;
  harga_final?: number;
  sn?: string;
};

function statusLabel(status: string) {
  const s = status.trim().toLowerCase();
  if (s === 'pending_payment') return 'Menunggu';
  if (s === 'paid') return 'Dibayar';
  if (s === 'processing_provider') return 'Diproses';
  if (s === 'success') return 'Berhasil';
  if (s === 'failed') return 'Gagal';
  if (s === 'refunded') return 'Refund';
  if (s === 'expired') return 'Expired';
  if (s === 'cancelled') return 'Batal';
  return status || '-';
}

function statusStyle(status: string) {
  const s = status.trim().toLowerCase();
  if (s === 'success')
    return { bg: 'bg-emerald-50', text: 'text-emerald-600', border: 'border-emerald-100', Icon: CircleCheck };
  if (s === 'paid')
    return { bg: 'bg-sky-50', text: 'text-sky-600', border: 'border-sky-100', Icon: CircleCheck };
  if (s === 'pending_payment' || s === 'processing_provider')
    return { bg: 'bg-amber-50', text: 'text-amber-600', border: 'border-amber-100', Icon: Clock3 };
  if (s === 'refunded')
    return { bg: 'bg-orange-50', text: 'text-orange-600', border: 'border-orange-100', Icon: Wallet };
  return { bg: 'bg-rose-50', text: 'text-rose-600', border: 'border-rose-100', Icon: CircleX };
}

function formatCurrency(n: number) {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(n);
}

function formatDate(iso: string) {
  if (!iso) return '-';
  try {
    return new Intl.DateTimeFormat('id-ID', { day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' }).format(new Date(iso));
  } catch {
    return iso;
  }
}

export function GuestTransactionHistory() {
  const router = useRouter();
  const [entries, setEntries] = useState<GuestTransactionEntry[]>([]);
  const [selectedRange, setSelectedRange] = useState('Semua');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedDate, setSelectedDate] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [isSyncing, setIsSyncing] = useState(false);

  const load = useCallback(() => {
    setEntries(loadGuestTransactions());
    setIsLoading(false);
  }, []);

  useEffect(() => { load(); }, [load]);

  const syncWithServer = useCallback(async () => {
    const current = loadGuestTransactions();
    if (current.length === 0) return;
    setIsSyncing(true);
    try {
      const toSync = current.filter((e) => e.guest_email && e.guest_phone).slice(0, 10);
      for (const entry of toSync) {
        try {
          const res = await getOrderByInvoice(entry.invoice_id, undefined, entry.guest_email, entry.guest_phone);
          const data = (Array.isArray(res) ? res[0] : res) as SyncedGuestOrder | undefined;
          if (data && data.invoice_id) {
            saveGuestTransaction({
              invoice_id: data.invoice_id,
              guest_email: entry.guest_email,
              guest_phone: entry.guest_phone,
              dest: data.dest || entry.dest,
              title: data.produk_nama_snapshot || entry.title,
              created_at: data.dibuat_pada || entry.created_at,
              updated_at: data.diubah_pada || data.dibuat_pada || entry.updated_at,
              status: data.status || entry.status,
              amount: Number(data.harga_final || 0) > 0 ? Number(data.harga_final || 0) : entry.amount,
              serial_number: data.sn || entry.serial_number,
            });
          }
        } catch { /* skip */ }
      }
    } finally {
      setIsSyncing(false);
      load();
    }
  }, [load]);

  useEffect(() => { syncWithServer(); }, [syncWithServer]);

  const handleOpen = (entry: GuestTransactionEntry) => {
    const qs = new URLSearchParams({ guest_email: entry.guest_email, guest_phone: entry.guest_phone });
    router.push(`/transaksi/${entry.invoice_id}?${qs.toString()}`);
  };

  const handleDelete = (invoiceId: string) => {
    removeGuestTransaction(invoiceId);
    load();
  };

  const filteredEntries = entries.filter((entry) => {
    const q = searchQuery.trim().toLowerCase();
    if (q) {
      const haystack = [entry.invoice_id, entry.dest, entry.title, entry.status].join(' ').toLowerCase();
      if (!haystack.includes(q)) return false;
    }

    const dateValue = entry.updated_at || entry.created_at;
    const date = dateValue ? new Date(dateValue) : null;
    if (!date || Number.isNaN(date.getTime())) {
      return selectedRange === 'Semua' && !selectedDate;
    }

    if (selectedDate) {
      return date.toISOString().slice(0, 10) === selectedDate;
    }

    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const yesterday = new Date(today);
    yesterday.setDate(today.getDate() - 1);
    const sevenDaysAgo = new Date(today);
    sevenDaysAgo.setDate(today.getDate() - 6);
    const sameDay = (a: Date, b: Date) => a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate();

    if (selectedRange === 'Hari ini') return sameDay(date, today);
    if (selectedRange === 'Kemarin') return sameDay(date, yesterday);
    if (selectedRange === '7 Hari') return date >= sevenDaysAgo;
    return true;
  });

  function formatDateDisplay(value: string) {
    if (!value) return 'dd/mm/yyyy';
    const [year, month, day] = value.split('-');
    if (!year || !month || !day) return 'dd/mm/yyyy';
    return `${day}/${month}/${year}`;
  }

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-xl font-black tracking-tight text-slate-950">Riwayat Transaksi</h1>
        <p className="mt-1 text-[11px] font-semibold text-slate-500">Pantau semua transaksi Isiloka</p>
      </div>

      <label className="flex h-13 items-center gap-3 rounded-[18px] border border-slate-200 bg-white px-4 shadow-[0_10px_24px_rgba(15,23,42,0.04)]">
        <Search className="h-5 w-5 shrink-0 text-slate-400" strokeWidth={2.1} />
        <input
          type="search"
          value={searchQuery}
          onChange={(event) => setSearchQuery(event.target.value)}
          placeholder="Cari nomor atau ID transaksi"
          className="h-full min-w-0 flex-1 bg-transparent text-sm font-semibold text-slate-700 outline-none placeholder:text-slate-400"
        />
      </label>

      <div className="flex gap-2 overflow-x-auto pb-1 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
        {['Semua', 'Hari ini', 'Kemarin', '7 Hari'].map((label) => (
          <button
            key={label}
            type="button"
            onClick={() => {
              setSelectedRange(label);
              if (label !== 'Semua') setSelectedDate('');
            }}
            className={
              selectedRange === label
                ? 'h-9 shrink-0 rounded-full bg-[#087e8b] px-4 text-xs font-black text-white shadow-[0_10px_20px_rgba(4,120,87,0.20)]'
                : 'h-9 shrink-0 rounded-full border border-slate-200 bg-white px-4 text-xs font-black text-slate-500'
            }
          >
            {label}
          </button>
        ))}
      </div>

      <label className="relative flex h-13 cursor-pointer items-center gap-3 rounded-[16px] border border-slate-200 bg-white px-4 text-slate-500 shadow-[0_10px_24px_rgba(15,23,42,0.04)]">
        <CalendarDays className="h-4.5 w-4.5 text-[#087e8b]" strokeWidth={2.1} />
        <span className="text-xs font-bold text-slate-500">Pilih tanggal</span>
        <span className={selectedDate ? 'ml-auto text-xs font-black text-slate-700' : 'ml-auto text-xs font-bold text-slate-400'}>
          {formatDateDisplay(selectedDate)}
        </span>
        <input
          type="date"
          value={selectedDate}
          onChange={(event) => {
            setSelectedDate(event.target.value);
            if (event.target.value) setSelectedRange('Semua');
          }}
          aria-label="Pilih tanggal transaksi"
          className="absolute inset-0 cursor-pointer opacity-0"
        />
      </label>

      {isSyncing && (
        <div className="flex items-center gap-2.5 rounded-xl bg-sky-50 px-4 py-3 ring-1 ring-sky-100">
          <RefreshCw className="h-3.5 w-3.5 animate-spin text-sky-500" />
          <span className="text-[12px] font-semibold text-sky-600">Memperbarui status...</span>
        </div>
      )}

      {isLoading ? (
        <div className="flex items-center justify-center gap-3 rounded-2xl bg-white px-5 py-10 shadow-sm ring-1 ring-slate-100">
          <Loader2 className="h-5 w-5 animate-spin text-slate-300" />
          <span className="text-sm font-medium text-slate-400">Memuat transaksi...</span>
        </div>
      ) : filteredEntries.length === 0 ? (
        <div className="grid min-h-[360px] place-items-center px-4 text-center">
          <div>
            <div className="mx-auto grid h-18 w-18 place-items-center text-slate-400">
              <ScrollText className="h-12 w-12" strokeWidth={1.8} />
            </div>
            <p className="mt-3 text-sm font-black text-slate-500">Belum ada transaksi</p>
            <p className="mt-1 max-w-[260px] text-[11px] font-semibold leading-4 text-slate-400">
              Transaksi pada tanggal yang dipilih tidak ditemukan.
            </p>
          </div>
        </div>
      ) : (
        <div className="space-y-2.5">
          {filteredEntries.map((entry) => {
            const st = statusStyle(entry.status);
            return (
              <button
                key={entry.invoice_id}
                onClick={() => handleOpen(entry)}
                className="group w-full rounded-2xl bg-white p-4 text-left shadow-sm ring-1 ring-slate-100 transition-all hover:shadow-md hover:ring-sky-200 active:scale-[0.995]"
              >
                {/* Top: product + chevron */}
                <div className="flex items-start gap-3">
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-[13px] font-semibold text-slate-900">
                      {entry.title || 'Transaksi Guest'}
                    </p>
                    <p className="mt-1 flex items-center gap-1 text-[11px] text-slate-400">
                      <ReceiptText className="h-3 w-3 shrink-0" />
                      <span className="truncate">{entry.invoice_id}</span>
                    </p>
                  </div>
                  <div className="flex items-center gap-1">
                    <button
                      onClick={(e) => { e.stopPropagation(); handleDelete(entry.invoice_id); }}
                      className="rounded-lg p-1 text-slate-300 opacity-0 transition-all hover:bg-slate-50 hover:text-rose-400 group-hover:opacity-100"
                      title="Hapus"
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </button>
                    <ChevronRight className="h-4 w-4 text-slate-300 transition-colors group-hover:text-sky-400" />
                  </div>
                </div>

                {/* Middle: price + badge */}
                <div className="mt-3 flex items-center justify-between gap-3">
                  <p className="text-[15px] font-bold tracking-tight text-slate-900">
                    {entry.amount > 0 ? formatCurrency(entry.amount) : '-'}
                  </p>
                  <span className={`inline-flex items-center gap-1 rounded-full border px-2.5 py-[3px] text-[10px] font-semibold ${st.bg} ${st.text} ${st.border}`}>
                    <st.Icon className="h-3 w-3" />
                    {statusLabel(entry.status)}
                  </span>
                </div>

                {/* Bottom: dest + date */}
                <div className="mt-2.5 flex items-center justify-between text-[11px] text-slate-400">
                  <span className="truncate">{entry.dest || '-'}</span>
                  <span className="shrink-0 ml-3 tabular-nums">
                    {formatDate(entry.updated_at || entry.created_at)}
                  </span>
                </div>
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
