"use client";

import * as React from "react";
import Link from "next/link";
import { LoaderCircle } from "lucide-react";

type UserRefundClaimFormProps = {
  authToken: string;
  initialInvoiceID?: string;
  initialGuestEmail?: string;
  initialGuestPhone?: string;
};

type ApiErrorResponse = {
  ok?: boolean;
  error?: string;
};

type GuestRefundItem = {
  invoice_id: string;
  amount_refund: number;
  status: string;
};

type ApiItemResponse<T> = {
  ok?: boolean;
  item?: T;
  error?: string;
};

function formatRupiah(value: number) {
  return `Rp ${new Intl.NumberFormat("id-ID").format(value || 0)}`;
}

export function UserRefundClaimForm({ authToken, initialInvoiceID, initialGuestEmail, initialGuestPhone }: UserRefundClaimFormProps) {
  const [invoiceID, setInvoiceID] = React.useState(initialInvoiceID || "");
  const [guestEmail, setGuestEmail] = React.useState(initialGuestEmail || "");
  const [guestPhone, setGuestPhone] = React.useState(initialGuestPhone || "");
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);
  const [success, setSuccess] = React.useState<string | null>(null);

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);
    setSuccess(null);

    if (!invoiceID.trim() || !guestEmail.trim() || !guestPhone.trim()) {
      setError("Invoice, email, dan no. HP wajib diisi.");
      return;
    }

    setLoading(true);
    try {
      const res = await fetch("/api/app/me/refunds/claim", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${authToken}`,
        },
        body: JSON.stringify({
          invoice_id: invoiceID.trim(),
          guest_email: guestEmail.trim(),
          guest_phone: guestPhone.trim(),
        }),
      });

      const json = (await res.json().catch(() => ({}))) as ApiItemResponse<GuestRefundItem> & ApiErrorResponse;
      if (!res.ok || !json.ok || !json.item) {
        throw new Error(json.error || "Klaim refund gagal.");
      }

      setSuccess(`Klaim berhasil. Refund ${formatRupiah(json.item.amount_refund)} sudah masuk ke saldo akun anda.`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Klaim refund gagal.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <section className="rounded-3xl bg-white p-5 shadow-[0_8px_22px_rgba(15,23,42,0.13)]">
      <p className="text-xs font-semibold tracking-wide text-neutral-500 uppercase">Klaim Refund</p>
      <h1 className="mt-2 text-lg font-bold text-neutral-900">Klaim refund transaksi guest</h1>
      <p className="mt-1 text-sm text-neutral-500">Masukkan data checkout guest yang gagal untuk memasukkan refund ke saldo akun ini.</p>

      <form onSubmit={handleSubmit} className="mt-4 space-y-3">
        <label className="block space-y-1">
          <span className="text-xs font-medium text-slate-600">No. Invoice</span>
          <input
            value={invoiceID}
            onChange={(e) => setInvoiceID(e.target.value)}
            placeholder="INV-2026..."
            className="h-11 w-full rounded-2xl border border-slate-200 bg-white px-3 text-sm text-slate-900 outline-none focus:border-sky-300"
          />
        </label>

        <label className="block space-y-1">
          <span className="text-xs font-medium text-slate-600">Email saat checkout guest</span>
          <input
            value={guestEmail}
            onChange={(e) => setGuestEmail(e.target.value)}
            placeholder="guest@email.com"
            className="h-11 w-full rounded-2xl border border-slate-200 bg-white px-3 text-sm text-slate-900 outline-none focus:border-sky-300"
          />
        </label>

        <label className="block space-y-1">
          <span className="text-xs font-medium text-slate-600">No. HP saat checkout guest</span>
          <input
            value={guestPhone}
            onChange={(e) => setGuestPhone(e.target.value)}
            placeholder="08xxxxxxxxxx"
            className="h-11 w-full rounded-2xl border border-slate-200 bg-white px-3 text-sm text-slate-900 outline-none focus:border-sky-300"
          />
        </label>

        {error ? <div className="rounded-2xl border border-rose-100 bg-rose-50 px-3 py-2 text-sm text-rose-700">{error}</div> : null}
        {success ? <div className="rounded-2xl border border-emerald-100 bg-emerald-50 px-3 py-2 text-sm text-emerald-700">{success}</div> : null}

        <div className="grid grid-cols-2 gap-2">
          <button
            type="submit"
            disabled={loading}
            className="inline-flex h-11 items-center justify-center gap-2 rounded-2xl bg-linear-to-r from-[#087e8b] to-[#087e8b] px-4 text-sm font-semibold text-white shadow-[0_8px_20px_rgba(15,111,203,0.24)] disabled:cursor-not-allowed disabled:opacity-60"
          >
            {loading ? <LoaderCircle className="h-4 w-4 animate-spin" /> : null}
            {loading ? "Memproses..." : "Klaim Refund"}
          </button>
          <Link
            href="/user/account"
            className="inline-flex h-11 items-center justify-center rounded-2xl border border-slate-200 bg-slate-50 px-4 text-sm font-semibold text-slate-700"
          >
            Kembali
          </Link>
        </div>
      </form>
    </section>
  );
}
