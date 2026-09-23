'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Search, FileText } from 'lucide-react';

type GuestInvoiceCheckerProps = {
  isLoggedIn?: boolean;
};

export function GuestInvoiceChecker({ isLoggedIn = false }: GuestInvoiceCheckerProps) {
  const [invoiceId, setInvoiceId] = useState('');
  const [guestEmail, setGuestEmail] = useState('');
  const [guestPhone, setGuestPhone] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const router = useRouter();

  if (isLoggedIn) {
    return null;
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!invoiceId.trim() || !guestEmail.trim() || !guestPhone.trim()) return;
    if (!guestEmail.trim().includes('@')) return;
    if (guestPhone.replace(/\D/g, '').length < 8) return;

    setIsLoading(true);
    try {
      const qs = new URLSearchParams({
        guest_email: guestEmail.trim().toLowerCase(),
        guest_phone: guestPhone.replace(/\D/g, ''),
      });
      router.push(`/transaksi/${invoiceId.trim()}?${qs.toString()}`);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <section className="rounded-md bg-white p-4 shadow-[0_10px_22px_rgba(15,23,42,0.1)] ring-1 ring-slate-100">
      <div className="mb-3 flex items-center gap-3">
        <div className="grid h-8 w-8 place-items-center rounded-lg bg-sky-100 text-sky-600">
          <FileText className="h-4 w-4" />
        </div>
        <div>
          <h3 className="font-semibold text-slate-900">Cek Status Transaksi</h3>
          <p className="text-sm text-slate-500">Masukkan invoice, email, dan no. HP saat checkout guest</p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-3">
        <div className="relative">
          <input
            type="text"
            value={invoiceId}
            onChange={(e) => setInvoiceId(e.target.value)}
            placeholder="Contoh: INV-2024-001234"
            className="w-full rounded-lg border border-slate-200 px-4 py-3 pl-11 text-sm placeholder:text-slate-400 focus:border-sky-500 focus:outline-none focus:ring-1 focus:ring-sky-500"
            disabled={isLoading}
          />
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
        </div>

        <input
          type="email"
          value={guestEmail}
          onChange={(e) => setGuestEmail(e.target.value)}
          placeholder="Email saat checkout"
          className="w-full rounded-lg border border-slate-200 px-4 py-3 text-sm placeholder:text-slate-400 focus:border-sky-500 focus:outline-none focus:ring-1 focus:ring-sky-500"
          disabled={isLoading}
        />

        <input
          type="tel"
          value={guestPhone}
          onChange={(e) => setGuestPhone(e.target.value)}
          placeholder="No. HP saat checkout"
          className="w-full rounded-lg border border-slate-200 px-4 py-3 text-sm placeholder:text-slate-400 focus:border-sky-500 focus:outline-none focus:ring-1 focus:ring-sky-500"
          disabled={isLoading}
        />

        <button
          type="submit"
          disabled={!invoiceId.trim() || !guestEmail.trim() || !guestPhone.trim() || !guestEmail.includes('@') || guestPhone.replace(/\D/g, '').length < 8 || isLoading}
          className="w-full rounded-lg bg-sky-600 px-4 py-3 text-sm font-medium text-white transition-colors hover:bg-sky-700 focus:outline-none focus:ring-2 focus:ring-sky-500 focus:ring-offset-2 disabled:cursor-not-allowed disabled:bg-slate-300 disabled:opacity-50"
        >
          {isLoading ? 'Mencari...' : 'Cek Transaksi'}
        </button>
      </form>
    </section>
  );
}
