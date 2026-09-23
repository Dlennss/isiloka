"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { ReactNode } from "react";
import type { UserAppOrder } from "@/components/user/types";
import { UserTransactionHistoryCard } from "@/components/user/UserTransactionHistoryCard";
import { UserTransactionPagination } from "@/components/user/UserTransactionPagination";

const PAGE_SIZE = 10;
const LOCAL_SERVICE_ORDER_KEY = "pulsakilat_local_service_orders";

type UserTransactionHistoryListProps = {
  initialItems: UserAppOrder[];
  initialHasNextPage: boolean;
  status: string;
  authToken: string;
  emptyIcon?: ReactNode;
  searchQuery?: string;
  selectedRange?: string;
  selectedDate?: string;
};

type OrdersResponse = {
  ok?: boolean;
  items?: UserAppOrder[];
};

export function UserTransactionHistoryList({
  initialItems,
  initialHasNextPage,
  status,
  authToken,
  emptyIcon,
  searchQuery = "",
  selectedRange = "Semua",
  selectedDate = "",
}: UserTransactionHistoryListProps) {
  const [items, setItems] = useState<UserAppOrder[]>(initialItems);
  const [page, setPage] = useState(1);
  const [hasNextPage, setHasNextPage] = useState(initialHasNextPage);
  const [autoMode, setAutoMode] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const containerRef = useRef<HTMLDivElement | null>(null);
  const sentinelRef = useRef<HTMLDivElement | null>(null);

  const readLocalOrders = useCallback(() => {
    try {
      const raw = window.localStorage.getItem(LOCAL_SERVICE_ORDER_KEY);
      const parsed = JSON.parse(raw || "[]") as UserAppOrder[];
      return Array.isArray(parsed) ? parsed : [];
    } catch {
      return [];
    }
  }, []);

  const mergeLocalOrders = useCallback(
    (serverItems: UserAppOrder[]) => {
      const localOrders = readLocalOrders();
      const seen = new Set<string>();
      return [...localOrders, ...serverItems]
        .filter((item) => {
          if (!item?.invoice_id || seen.has(item.invoice_id)) return false;
          seen.add(item.invoice_id);
          return true;
        })
        .sort((a, b) => String(b.dibuat_pada || "").localeCompare(String(a.dibuat_pada || "")));
    },
    [readLocalOrders],
  );

  useEffect(() => {
    setItems(mergeLocalOrders(initialItems));
    setPage(1);
    setHasNextPage(initialHasNextPage);
    setAutoMode(false);
    setIsLoading(false);
    if (containerRef.current) {
      containerRef.current.scrollTop = 0;
    }
  }, [initialHasNextPage, initialItems, mergeLocalOrders, status]);

  const loadPage = useCallback(
    async (nextPage: number) => {
      if (isLoading || !hasNextPage) {
        return;
      }

      setIsLoading(true);
      try {
        const qs = new URLSearchParams();
        if (status) qs.set("status", status);
        qs.set("limit", String(PAGE_SIZE + 1));
        qs.set("offset", String((nextPage - 1) * PAGE_SIZE));

        const res = await fetch(`/api/app/me/orders?${qs.toString()}`, {
          headers: {
            Authorization: `Bearer ${authToken}`,
          },
          cache: "no-store",
        });
        const json = (await res.json().catch(() => ({}))) as OrdersResponse;
        if (!res.ok || !json.ok || !Array.isArray(json.items)) {
          return;
        }

        const nextHasNextPage = json.items.length > PAGE_SIZE;
        const nextItems = nextHasNextPage ? json.items.slice(0, PAGE_SIZE) : json.items;

        setItems((prev) => mergeLocalOrders([...prev, ...nextItems]));
        setPage(nextPage);
        setHasNextPage(nextHasNextPage);
      } finally {
        setIsLoading(false);
      }
    },
    [authToken, hasNextPage, isLoading, status],
  );

  const handleEnableAutoLoad = useCallback(async () => {
    setAutoMode(true);
    await loadPage(2);
  }, [loadPage]);

  useEffect(() => {
    if (!autoMode || !hasNextPage || isLoading) {
      return;
    }
    const root = containerRef.current;
    const target = sentinelRef.current;
    if (!root || !target) {
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        const entry = entries[0];
        if (!entry?.isIntersecting) {
          return;
        }
        void loadPage(page + 1);
      },
      {
        root,
        rootMargin: "120px 0px 120px 0px",
        threshold: 0.1,
      },
    );

    observer.observe(target);
    return () => observer.disconnect();
  }, [autoMode, hasNextPage, isLoading, loadPage, page]);

  const filteredItems = useMemo(() => {
    const q = searchQuery.trim().toLowerCase();
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const yesterday = new Date(today);
    yesterday.setDate(today.getDate() - 1);
    const sevenDaysAgo = new Date(today);
    sevenDaysAgo.setDate(today.getDate() - 6);

    function sameDay(a: Date, b: Date) {
      return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate();
    }

    return items.filter((item) => {
      if (q) {
        const haystack = [
          item.invoice_id,
          item.dest,
          item.produk_nama_snapshot,
          item.produk_sku_snapshot,
          item.status,
        ]
          .join(" ")
          .toLowerCase();
        if (!haystack.includes(q)) return false;
      }

      const date = item.dibuat_pada ? new Date(item.dibuat_pada) : null;
      if (!date || Number.isNaN(date.getTime())) {
        return selectedRange === "Semua" && !selectedDate;
      }

      if (selectedDate) {
        return item.dibuat_pada?.slice(0, 10) === selectedDate;
      }

      if (selectedRange === "Hari ini") return sameDay(date, today);
      if (selectedRange === "Kemarin") return sameDay(date, yesterday);
      if (selectedRange === "7 Hari") return date >= sevenDaysAgo;
      return true;
    });
  }, [items, searchQuery, selectedDate, selectedRange]);

  const content = useMemo(() => {
    if (filteredItems.length === 0) {
      return (
        <div className="grid min-h-[360px] place-items-center px-4 text-center">
          <div>
            <div className="mx-auto grid h-18 w-18 place-items-center text-slate-400">
              {emptyIcon}
            </div>
            <p className="mt-3 text-sm font-black text-slate-500">Belum ada transaksi</p>
            <p className="mt-1 max-w-[260px] text-[11px] font-semibold leading-4 text-slate-400">
              Transaksi pada tanggal yang dipilih tidak ditemukan.
            </p>
          </div>
        </div>
      );
    }

    return (
      <>
        {filteredItems.map((trx) => (
          <UserTransactionHistoryCard key={`${trx.invoice_id}-${trx.id}`} item={trx} />
        ))}
        <div ref={sentinelRef} className="h-1" />
      </>
    );
  }, [emptyIcon, filteredItems]);

  return (
    <>
      <div
        ref={containerRef}
        className="mt-4 space-y-3 pb-[calc(5rem+env(safe-area-inset-bottom))]"
      >
        {content}
        {isLoading ? (
          <div className="rounded-2xl border border-slate-100 bg-slate-50 px-4 py-3 text-center text-sm text-slate-500">
            Memuat riwayat berikutnya...
          </div>
        ) : null}

        {page === 1 ? (
          <UserTransactionPagination
            status={status}
            page={page}
            hasNextPage={hasNextPage}
            mode="manual-first"
            onNext={handleEnableAutoLoad}
          />
        ) : null}
      </div>
    </>
  );
}
