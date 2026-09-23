"use client";

import type { ReactNode } from "react";
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from "lucide-react";
import { Button } from "@/components/ui/button";

type Align = "left" | "center" | "right";

export type DataTableColumn<TRow> = {
  id: string;
  header: ReactNode;
  render: (row: TRow, index: number) => ReactNode;
  thClassName?: string;
  tdClassName?: string;
};

export type DataTableActions<TRow> = {
  header?: ReactNode;
  align?: Align;
  thClassName?: string;
  tdClassName?: string;
  render: (row: TRow, index: number) => ReactNode;
};

export type DataTablePagination = {
  page: number;
  totalPages: number;
  onPrev: () => void;
  onNext: () => void;
  onPageChange?: (page: number) => void;
  disablePrev?: boolean;
  disableNext?: boolean;
};

type DataTableProps<TRow> = {
  columns: DataTableColumn<TRow>[];
  rows?: TRow[];
  data?: TRow[];
  rowKey: (row: TRow, index: number) => string | number;
  emptyText?: string;
  emptyMessage?: string;
  minWidthClassName?: string;
  wrapperClassName?: string;
  showRowNumber?: boolean;
  rowNumberStart?: number;
  pagination?: DataTablePagination;
  actions?: DataTableActions<TRow>;
  loading?: boolean;
  variant?: "dark" | "light";
};

function alignClass(align: Align | undefined): string {
  if (align === "center") return "text-center";
  if (align === "right") return "text-right";
  return "text-left";
}

function visiblePages(current: number, total: number): number[] {
  const safeCurrent = Math.max(1, current);
  const safeTotal = Math.max(1, total);
  const start = Math.max(1, safeCurrent - 1);
  const end = Math.min(safeTotal, safeCurrent + 1);
  const pages: number[] = [];
  for (let page = start; page <= end; page += 1) pages.push(page);
  return pages;
}

export function DataTable<TRow>(props: DataTableProps<TRow>) {
  const {
    columns,
    rows,
    data,
    rowKey,
    emptyText,
    emptyMessage,
    minWidthClassName,
    wrapperClassName,
    showRowNumber = true,
    rowNumberStart = 1,
    pagination,
    actions,
    loading = false,
    variant = "dark",
  } = props;

  const safeRows = Array.isArray(rows) ? rows : Array.isArray(data) ? data : [];
  const actionAlign = alignClass(actions?.align);
  const currentPage = Math.max(1, pagination?.page ?? 1);
  const totalPages = Math.max(1, pagination?.totalPages ?? 1);
  const pageButtons = visiblePages(currentPage, totalPages);
  const isLight = variant === "light";
  const stateClassName = isLight
    ? "rounded-lg border border-dashed border-emerald-200 bg-white px-3 py-10 text-center text-sm text-slate-500 shadow-sm sm:px-4 sm:py-12"
    : "mt-5 rounded-md border border-white/15 bg-slate-950/50 px-3 py-6 text-center text-sm text-slate-400 shadow-[0_18px_42px_-26px_rgba(56,189,248,0.45)] sm:px-4 sm:py-8";

  if (loading) {
    return (
      <div
        className={
          wrapperClassName ||
          stateClassName
        }
      >
        Memuat data...
      </div>
    );
  }

  if (!safeRows.length) {
    return (
      <div
        className={
          wrapperClassName ||
          stateClassName
        }
      >
        {emptyMessage || emptyText || "Tidak ada data."}
      </div>
    );
  }

  return (
    <>
      <div
        className={
          wrapperClassName ||
          (isLight
            ? "overflow-x-auto overflow-y-visible rounded-lg border border-slate-200 bg-white shadow-sm"
            : "mt-5 overflow-x-auto overflow-y-visible rounded-md border border-white/15 bg-slate-950/50 shadow-[0_18px_42px_-26px_rgba(56,189,248,0.45)]")
        }
      >
        <table className={`w-full text-sm ${minWidthClassName || "min-w-245"}`}>
          <thead className={isLight ? "sticky top-0 z-10 bg-emerald-50" : "sticky top-0 z-10 bg-linear-to-r from-cyan-500/20 via-sky-500/10 to-indigo-500/20 backdrop-blur"}>
            <tr className="text-left">
              {showRowNumber ? (
                <th className={`px-3 py-1.5 text-[11px] font-semibold uppercase tracking-wide sm:px-4 sm:py-3 sm:text-xs ${isLight ? "text-emerald-900" : "text-slate-200"}`}>No</th>
              ) : null}
              {columns.map((col) => (
                <th
                  key={col.id}
                  className={`px-3 py-1.5 text-[11px] font-semibold uppercase tracking-wide sm:px-4 sm:py-3 sm:text-xs ${isLight ? "text-emerald-900" : "text-slate-200"} ${col.thClassName || ""}`}
                >
                  {col.header}
                </th>
              ))}
              {actions ? (
                <th
                  className={`px-3 py-1.5 text-[11px] font-semibold uppercase tracking-wide sm:px-4 sm:py-3 sm:text-xs ${isLight ? "text-emerald-900" : "text-slate-200"} ${actionAlign} ${
                    actions.thClassName || ""
                  }`}
                >
                  {actions.header || "Aksi"}
                </th>
              ) : null}
            </tr>
          </thead>
          <tbody>
            {safeRows.map((row, index) => (
              <tr key={rowKey(row, index)} className={isLight ? "border-t border-slate-100 bg-white transition hover:bg-emerald-50/60" : "border-t border-white/10 bg-white/1.5 transition hover:bg-cyan-400/[0.07]"}>
                {showRowNumber ? <td className={`whitespace-nowrap px-3 py-2.5 sm:px-4 sm:py-3 ${isLight ? "text-slate-500" : "text-slate-300"}`}>{rowNumberStart + index}</td> : null}
                {columns.map((col) => (
                  <td key={col.id} className={`px-3 py-2.5 sm:px-4 sm:py-3 ${col.tdClassName || ""}`}>
                    {col.render(row, index)}
                  </td>
                ))}
                {actions ? (
                  <td className={`px-3 py-2.5 sm:px-4 sm:py-3 ${actionAlign} ${actions.tdClassName || ""}`}>{actions.render(row, index)}</td>
                ) : null}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {pagination ? (
        <div className={`mt-3 flex flex-col gap-2 rounded-md p-2.5 text-sm sm:flex-row sm:items-center sm:justify-between sm:p-3 ${isLight ? "border border-slate-200 bg-white shadow-sm" : "border border-white/15 bg-linear-to-r from-slate-900/80 via-slate-900/65 to-cyan-950/25 shadow-[0_16px_36px_-28px_rgba(6,182,212,0.8)]"}`}>
          <div className={`text-xs sm:text-sm ${isLight ? "text-slate-600" : "text-slate-300"}`}>
            Halaman {currentPage} / {totalPages}
          </div>
          <div className="flex flex-wrap items-center gap-2 self-end sm:self-auto">
            <Button
              variant="outline"
              className="h-8 w-8 p-0 text-xs sm:h-9 sm:w-9"
              onClick={() => pagination.onPageChange?.(1)}
              disabled={pagination.disablePrev || !pagination.onPageChange}
              aria-label="Halaman pertama"
            >
              <ChevronsLeft className="h-4 w-4" />
            </Button>
            <Button
              variant="outline"
              className="h-8 w-8 p-0 text-xs sm:h-9 sm:w-9"
              onClick={pagination.onPrev}
              disabled={pagination.disablePrev}
              aria-label="Halaman sebelumnya"
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
            {pageButtons.map((page) => {
              const isActive = page === currentPage;
              return (
                <Button
                  key={page}
                  variant={isActive ? "primary" : "outline"}
                  className="h-8 min-w-8 px-2 text-xs sm:h-9 sm:min-w-9 sm:px-3 sm:text-sm"
                  onClick={() => pagination.onPageChange?.(page)}
                  disabled={isActive || !pagination.onPageChange}
                >
                  {page}
                </Button>
              );
            })}
            <Button
              variant="outline"
              className="h-8 w-8 p-0 text-xs sm:h-9 sm:w-9"
              onClick={pagination.onNext}
              disabled={pagination.disableNext}
              aria-label="Halaman berikutnya"
            >
              <ChevronRight className="h-4 w-4" />
            </Button>
            <Button
              variant="outline"
              className="h-8 w-8 p-0 text-xs sm:h-9 sm:w-9"
              onClick={() => pagination.onPageChange?.(totalPages)}
              disabled={pagination.disableNext || !pagination.onPageChange}
              aria-label="Halaman terakhir"
            >
              <ChevronsRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      ) : null}
    </>
  );
}
