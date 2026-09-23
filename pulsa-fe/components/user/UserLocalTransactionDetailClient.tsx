"use client";

import { useEffect, useState } from "react";
import type { UserAppOrder } from "@/components/user/types";
import { UserTransactionStatusBadge } from "@/components/user/UserTransactionStatusBadge";
import { ReceiptText } from "lucide-react";

const LOCAL_SERVICE_ORDER_KEY = "pulsakilat_local_service_orders";

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

export function UserLocalTransactionDetailClient({ invoiceId }: { invoiceId: string }) {
  const [order, setOrder] = useState<UserAppOrder | null>(null);

  useEffect(() => {
    queueMicrotask(() => {
      try {
        const raw = window.localStorage.getItem(LOCAL_SERVICE_ORDER_KEY);
        const items = JSON.parse(raw || "[]") as UserAppOrder[];
        setOrder(items.find((item) => item.invoice_id === invoiceId) || null);
      } catch {
        setOrder(null);
      }
    });
  }, [invoiceId]);

  if (!order) return null;

  return (
    <>
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Detail Transaksi</p>
          <h1 className="mt-1 text-xl font-bold text-slate-900">{order.produk_nama_snapshot}</h1>
          <p className="mt-2 flex items-center gap-1 text-xs text-slate-500">
            <ReceiptText className="h-3.5 w-3.5" />
            {order.invoice_id}
          </p>
        </div>
        <UserTransactionStatusBadge status={order.status} />
      </div>

      <div className="mt-5 rounded-3xl bg-emerald-50 p-4">
        <p className="text-xs font-semibold uppercase tracking-wide text-[#087e8b]">Total Pembayaran</p>
        <p className="mt-1 text-2xl font-bold text-slate-900">{formatRupiah(order.harga_final)}</p>
      </div>

      <div className="mt-5 divide-y divide-slate-100">
        <DetailRow label="Nama" value={order.member_nama || "-"} />
        <DetailRow label="Tujuan" value={order.dest} />
        <DetailRow label="Tipe Buyer" value="User" />
        <DetailRow label="Nominal" value={formatRupiah(order.nominal)} />
        <DetailRow label="Harga Dasar" value={formatRupiah(order.harga_dasar)} />
        <DetailRow label="Fee" value={formatRupiah(order.fee)} />
        <DetailRow label="Dibuat" value={formatDateTime(order.dibuat_pada)} />
        <DetailRow label="Diupdate" value={formatDateTime(order.diubah_pada)} />
      </div>
    </>
  );
}
