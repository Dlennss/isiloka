import Link from "next/link";
import { ChevronRight, ReceiptText } from "lucide-react";
import type { UserAppOrder } from "@/components/user/types";
import { UserTransactionStatusBadge } from "@/components/user/UserTransactionStatusBadge";

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

type UserTransactionHistoryCardProps = {
  item: UserAppOrder;
};

export function UserTransactionHistoryCard({ item }: UserTransactionHistoryCardProps) {
  return (
    <Link href={`/user/transaksi/${encodeURIComponent(item.invoice_id)}`} className="block rounded-2xl border border-neutral-100 p-4 shadow-sm transition hover:border-sky-200 hover:shadow-md">
      <article>
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <p className="truncate text-[13px] font-semibold text-neutral-900">{item.produk_nama_snapshot}</p>
            <p className="mt-1 flex items-center gap-1 text-[11px] text-neutral-500">
              <ReceiptText className="h-3.5 w-3.5" />
              <span className="truncate">{item.invoice_id}</span>
            </p>
          </div>
          <ChevronRight className="mt-0.5 h-4 w-4 shrink-0 text-slate-400" />
        </div>

        <div className="mt-3 flex items-center justify-between gap-3">
          <p className="text-[13px] font-semibold text-neutral-900">{formatRupiah(item.harga_final)}</p>
          <UserTransactionStatusBadge status={item.status} />
        </div>

        <div className="mt-3 flex items-center justify-between text-[11px] text-neutral-500">
          <span className="truncate">{item.dest}</span>
          <span>{formatDateTime(item.dibuat_pada)}</span>
        </div>
      </article>
    </Link>
  );
}
