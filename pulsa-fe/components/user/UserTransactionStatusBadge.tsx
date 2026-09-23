import { CircleCheck, CircleX, Clock3, Wallet } from "lucide-react";

function formatStatusLabel(status: string) {
  if (status === "pending_payment") return "Pending";
  if (status === "paid") return "Dibayar";
  if (status === "processing_provider") return "Diproses";
  if (status === "success") return "Berhasil";
  if (status === "failed") return "Gagal";
  if (status === "refunded") return "Dana Kembali";
  if (status === "expired") return "Kedaluwarsa";
  if (status === "cancelled") return "Dibatalkan";
  return status
    .replace(/_/g, " ")
    .split(" ")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

type UserTransactionStatusBadgeProps = {
  status: string;
};

export function UserTransactionStatusBadge({ status }: UserTransactionStatusBadgeProps) {
  if (status === "success") {
    return (
      <span className="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-semibold text-emerald-700">
        <CircleCheck className="h-3.5 w-3.5" /> {formatStatusLabel(status)}
      </span>
    );
  }
  if (status === "paid") {
    return (
      <span className="inline-flex items-center gap-1 rounded-full bg-sky-50 px-2.5 py-1 text-xs font-semibold text-sky-700">
        <CircleCheck className="h-3.5 w-3.5" /> {formatStatusLabel(status)}
      </span>
    );
  }
  if (status === "pending_payment" || status === "processing_provider") {
    return (
      <span className="inline-flex items-center gap-1 rounded-full bg-amber-50 px-2.5 py-1 text-xs font-semibold text-amber-700">
        <Clock3 className="h-3.5 w-3.5" /> {formatStatusLabel(status)}
      </span>
    );
  }
  if (status === "refunded") {
    return (
      <span className="inline-flex items-center gap-1 rounded-full bg-orange-50 px-2.5 py-1 text-xs font-semibold text-orange-700">
        <Wallet className="h-3.5 w-3.5" /> {formatStatusLabel(status)}
      </span>
    );
  }
  return (
    <span className="inline-flex items-center gap-1 rounded-full bg-rose-50 px-2.5 py-1 text-xs font-semibold text-rose-700">
      <CircleX className="h-3.5 w-3.5" /> {formatStatusLabel(status)}
    </span>
  );
}
