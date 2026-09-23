import Link from "next/link";

type UserTransactionPaginationProps = {
  status: string;
  page: number;
  hasNextPage: boolean;
  mode?: "default" | "manual-first";
  onNext?: () => void | Promise<void>;
};

function buildHref(status: string, page: number) {
  const params = new URLSearchParams();
  if (status) params.set("status", status);
  if (page > 1) params.set("page", String(page));
  const qs = params.toString();
  return `/user/transaksi${qs ? `?${qs}` : ""}`;
}

export function UserTransactionPagination({
  status,
  page,
  hasNextPage,
  mode = "default",
  onNext,
}: UserTransactionPaginationProps) {
  if (!hasNextPage) {
    return null;
  }

  return (
    <div className="flex justify-center">
      {mode === "manual-first" && page === 1 ? (
        <button
          type="button"
          onClick={() => void onNext?.()}
          className="inline-flex w-full justify-center rounded-2xl bg-[#087e8b] px-4 py-2 text-sm font-semibold text-white! shadow-sm transition hover:bg-[#0a5dad] hover:text-white!"
        >
          Lainnya
        </button>
      ) : (
        <Link
          href={buildHref(status, page + 1)}
          className="inline-flex w-full justify-center rounded-2xl bg-[#087e8b] px-4 py-2 text-sm font-semibold text-white! visited:text-white! shadow-sm transition hover:bg-[#0a5dad] hover:text-white!"
        >
          Lainnya
        </Link>
      )}
    </div>
  );
}
