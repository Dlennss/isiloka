import Link from "next/link";

const STATUS_FILTERS = [
  { label: "Semua", value: "" },
  { label: "Pending", value: "pending_payment" },
  { label: "Paid", value: "paid" },
  { label: "Diproses", value: "processing_provider" },
  { label: "Sukses", value: "success" },
  { label: "Gagal", value: "failed" },
  { label: "Expired", value: "expired" },
  { label: "Batal", value: "cancelled" },
  { label: "Refund", value: "refunded" },
];

type UserTransactionStatusFiltersProps = {
  status: string;
};

export function UserTransactionStatusFilters({ status }: UserTransactionStatusFiltersProps) {
  return (
    <div className="mt-4 flex gap-2 overflow-x-auto pb-1">
      {STATUS_FILTERS.map((filter) => {
        const active = status === filter.value;
        const href = filter.value ? `/user/transaksi?status=${encodeURIComponent(filter.value)}` : "/user/transaksi";

        return (
          <Link
            key={filter.value || "all"}
            href={href}
            className={`whitespace-nowrap rounded-full px-3 py-2 text-xs font-semibold transition ${
              active
                ? "bg-[#087e8b] text-white! visited:text-white! hover:bg-[#0a5dad] hover:text-white! focus-visible:text-white!"
                : "bg-slate-100 text-slate-600 hover:bg-slate-200"
            }`}
          >
            {filter.label}
          </Link>
        );
      })}
    </div>
  );
}
