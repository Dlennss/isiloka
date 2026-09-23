import Link from "next/link";
import { ArrowDownLeft, ArrowUpRight, ChevronRight } from "lucide-react";

type MutasiRow = {
  id: number;
  member_id: number;
  ref_id: string;
  arah: string;
  jumlah: number;
  alasan: string;
  catatan?: string | null;
  saldo_sebelum?: number | null;
  saldo_sesudah?: number | null;
  dibuat_pada: string;
};

function formatRupiah(value?: number | null) {
  return `Rp ${new Intl.NumberFormat("id-ID").format(Number(value || 0))}`;
}

function formatDateTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function formatReason(row: MutasiRow) {
  if (row.alasan === "APP_ORDER_REFUND") return "Refund transaksi aplikasi";
  if (row.alasan === "GUEST_REFUND_CLAIM") return "Klaim refund transaksi guest";
  return row.alasan || "Mutasi saldo";
}

type Props = {
  item: MutasiRow;
};

export function UserSaldoMutationHistoryCard({ item }: Props) {
  const arah = String(item.arah || "").toUpperCase();
  const isCredit = arah === "CREDIT";

  return (
    <Link
      href={`/user/account/mutasi/${item.id}`}
      className="flex items-center justify-between gap-3 rounded-md border border-slate-200 bg-slate-50 px-4 py-3 transition hover:border-slate-300 hover:bg-slate-100"
    >
      <div className="flex min-w-0 items-center gap-3">
        <span className={`grid h-9 w-9 shrink-0 place-items-center rounded-full ${isCredit ? "bg-emerald-100 text-emerald-700" : "bg-rose-100 text-rose-700"}`}>
          {isCredit ? <ArrowDownLeft className="h-4 w-4" /> : <ArrowUpRight className="h-4 w-4" />}
        </span>
        <div className="min-w-0">
          <p className="truncate text-sm font-semibold text-slate-900">{formatReason(item)}</p>
          <p className="mt-0.5 text-[11px] text-slate-500">{formatDateTime(item.dibuat_pada)}</p>
        </div>
      </div>

      <div className="flex shrink-0 items-center gap-2">
        <p className={`text-sm font-bold ${isCredit ? "text-emerald-700" : "text-rose-700"}`}>
          {isCredit ? "+" : "-"}{formatRupiah(item.jumlah)}
        </p>
        <ChevronRight className="h-4 w-4 text-slate-400" />
      </div>
    </Link>
  );
}
