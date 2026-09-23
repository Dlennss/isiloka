import Link from "next/link";
import { getAppServerSession } from "@/lib/server-auth";
import { ArrowLeft, ReceiptText } from "lucide-react";
import { getOrderByInvoice } from "@/lib/api.transactions";
import type { UserAppOrder, UserSession } from "@/components/user/types";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import { UserTransactionStatusBadge } from "@/components/user/UserTransactionStatusBadge";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

type PageProps = {
  params: Promise<{ invoiceId: string }>;
  searchParams?: Promise<{
    guest_email?: string;
    guest_phone?: string;
  }>;
};

function formatRupiah(value: number) {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    maximumFractionDigits: 0,
  }).format(value || 0);
}

function formatDateTime(value?: string | null) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return new Intl.DateTimeFormat("id-ID", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-start justify-between gap-4 py-3">
      <p className="text-sm text-slate-500">{label}</p>
      <p className="max-w-[60%] text-right text-sm font-medium text-slate-900">{value}</p>
    </div>
  );
}

export default async function UserTransactionDetailPage({ params, searchParams }: PageProps) {
  const session = (await getAppServerSession()) as SessionShape | null;
  const user = session?.user ?? null;
  const backendToken = session?.backendToken;

  const { invoiceId } = await params;
  const sp = searchParams ? await searchParams : undefined;
  const guestEmail = String(sp?.guest_email || "").trim().toLowerCase();
  const guestPhone = String(sp?.guest_phone || "").replace(/\D/g, "");

  // Try to get order details (works for both authenticated users and guests)
  let order: UserAppOrder | null = null;
  let error: string | null = null;

  try {
    if (!backendToken && (!guestEmail || !guestPhone)) {
      error = "Untuk transaksi guest, isi email dan no. HP saat checkout dari halaman cek invoice.";
    } else {
      const result = await getOrderByInvoice(invoiceId, backendToken || undefined, guestEmail || undefined, guestPhone || undefined);
      if (result && Array.isArray(result) && result.length > 0) {
        order = result[0] as UserAppOrder;
      } else {
        error = "Transaksi tidak ditemukan";
      }
    }
  } catch (err) {
    console.error("Error fetching order:", err);
    error = "Terjadi kesalahan saat mengambil data transaksi";
  }

  return (
    <main className="min-h-screen bg-sky-50 px-4 pb-24 pt-5">
      <div className="mx-auto w-full max-w-md space-y-4">
        <Link href={backendToken ? "/user/transaksi" : "/user"} className="inline-flex items-center gap-2 text-sm font-medium text-[#087e8b] hover:text-[#0a5dad]">
          <ArrowLeft className="h-4 w-4" />
          {backendToken ? "Kembali ke riwayat" : "Kembali ke beranda"}
        </Link>

        <section className="rounded-3xl bg-white p-5 shadow-[0_8px_22px_rgba(15,23,42,0.13)]">
          {error ? (
            <div className="text-center py-8">
              <div className="mx-auto h-12 w-12 rounded-full bg-red-100 flex items-center justify-center mb-4">
                <svg className="h-6 w-6 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
                </svg>
              </div>
              <h3 className="text-lg font-medium text-gray-900 mb-2">Transaksi Tidak Ditemukan</h3>
              <p className="text-gray-500">{error}</p>
              {!backendToken ? (
                <Link href="/?focus=invoice-checker" className="mt-4 inline-flex items-center justify-center rounded-xl bg-sky-600 px-4 py-2 text-sm font-semibold text-white">
                  Kembali ke cek invoice
                </Link>
              ) : null}
            </div>
          ) : order ? (
            <>
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="text-xs font-semibold tracking-wide text-slate-500 uppercase">Detail Transaksi</p>
                  <h1 className="mt-1 text-xl font-bold text-slate-900">{order.produk_nama_snapshot}</h1>
                  <p className="mt-2 flex items-center gap-1 text-xs text-slate-500">
                    <ReceiptText className="h-3.5 w-3.5" />
                    {order.invoice_id}
                  </p>
                </div>
                <UserTransactionStatusBadge status={order.status} />
              </div>

              <div className="mt-5 rounded-3xl bg-sky-50 p-4">
                <p className="text-xs font-semibold tracking-wide text-sky-700 uppercase">Total Pembayaran</p>
                <p className="mt-1 text-2xl font-bold text-slate-900">{formatRupiah(order.harga_final)}</p>
              </div>

              <div className="mt-5 divide-y divide-slate-100">
                <DetailRow label="Nama" value={order.member_nama || order.guest_nama || "-"} />
                <DetailRow label="Tujuan" value={order.dest} />
                <DetailRow label="Tipe Buyer" value={order.buyer_type === "user" ? "User" : "Guest"} />
                <DetailRow label="Nominal" value={formatRupiah(order.nominal)} />
                <DetailRow label="Harga Dasar" value={formatRupiah(order.harga_dasar)} />
                <DetailRow label="Fee" value={formatRupiah(order.fee)} />
                <DetailRow label="Dibuat" value={formatDateTime(order.dibuat_pada)} />
                <DetailRow label="Diupdate" value={formatDateTime(order.diubah_pada)} />
              </div>
            </>
          ) : (
            <div className="text-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-sky-600 mx-auto mb-4"></div>
              <p className="text-gray-500">Memuat detail transaksi...</p>
            </div>
          )}
        </section>
      </div>

      <UserBottomNav />
    </main>
  );
}
