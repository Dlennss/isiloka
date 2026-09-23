"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { UserSaldoMutationHistoryCard } from "@/components/user/UserSaldoMutationHistoryCard";

const PAGE_SIZE = 10;

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

type ResponseShape = {
  ok?: boolean;
  rows?: MutasiRow[];
};

type Props = {
  initialItems: MutasiRow[];
  initialHasNextPage: boolean;
  arah: string;
  from?: string;
  to?: string;
  authToken: string;
};

export function UserSaldoMutationHistoryList({ initialItems, initialHasNextPage, arah, from = "", to = "", authToken }: Props) {
  const [items, setItems] = useState<MutasiRow[]>(initialItems);
  const [page, setPage] = useState(1);
  const [hasNextPage, setHasNextPage] = useState(initialHasNextPage);
  const [autoMode, setAutoMode] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const containerRef = useRef<HTMLDivElement | null>(null);
  const sentinelRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    setItems(initialItems);
    setPage(1);
    setHasNextPage(initialHasNextPage);
    setAutoMode(false);
    setIsLoading(false);
    if (containerRef.current) containerRef.current.scrollTop = 0;
  }, [arah, from, to, initialHasNextPage, initialItems]);

  const loadPage = useCallback(async (nextPage: number) => {
    if (isLoading || !hasNextPage) return;
    setIsLoading(true);
    try {
      const qs = new URLSearchParams();
      qs.set("limit", String(PAGE_SIZE + 1));
      qs.set("offset", String((nextPage - 1) * PAGE_SIZE));
      if (arah) qs.set("arah", arah);
      if (from) qs.set("from", from);
      if (to) qs.set("to", to);

      const res = await fetch(`/api/me/history/mutasi?${qs.toString()}`, {
        headers: { Authorization: `Bearer ${authToken}` },
        cache: "no-store",
      });
      const json = (await res.json().catch(() => ({}))) as ResponseShape;
      if (!res.ok || !json.ok || !Array.isArray(json.rows)) return;

      const nextHasNextPage = json.rows.length > PAGE_SIZE;
      const nextItems = nextHasNextPage ? json.rows.slice(0, PAGE_SIZE) : json.rows;
      setItems((prev) => [...prev, ...nextItems]);
      setPage(nextPage);
      setHasNextPage(nextHasNextPage);
    } finally {
      setIsLoading(false);
    }
  }, [arah, authToken, from, hasNextPage, isLoading, to]);

  const handleEnableAutoLoad = useCallback(async () => {
    setAutoMode(true);
    await loadPage(2);
  }, [loadPage]);

  useEffect(() => {
    if (!autoMode || !hasNextPage || isLoading) return;
    const root = containerRef.current;
    const target = sentinelRef.current;
    if (!root || !target) return;

    const observer = new IntersectionObserver(
      (entries) => {
        const entry = entries[0];
        if (!entry?.isIntersecting) return;
        void loadPage(page + 1);
      },
      { root, rootMargin: "120px 0px 120px 0px", threshold: 0.1 },
    );

    observer.observe(target);
    return () => observer.disconnect();
  }, [autoMode, hasNextPage, isLoading, loadPage, page]);

  const content = useMemo(() => {
    if (items.length === 0) {
      return (
        <div className="grid min-h-30 place-items-center rounded-md border border-dashed border-slate-200 bg-slate-50 px-4 text-center text-sm text-slate-500">
          Belum ada riwayat mutasi saldo.
        </div>
      );
    }

    return (
      <>
        {items.map((item) => (
          <UserSaldoMutationHistoryCard key={item.id} item={item} />
        ))}
        <div ref={sentinelRef} className="h-1" />
      </>
    );
  }, [items]);

  return (
    <div ref={containerRef} className="mt-4 min-h-0 flex-1 space-y-3 overflow-y-auto pb-[calc(5rem+env(safe-area-inset-bottom))]">
      {content}
      {isLoading ? (
        <div className="rounded-md border border-slate-100 bg-slate-50 px-4 py-3 text-center text-sm text-slate-500">
          Memuat riwayat berikutnya...
        </div>
      ) : null}

      {hasNextPage && page === 1 ? (
        <div className="flex justify-center">
          <button
            type="button"
            onClick={() => void handleEnableAutoLoad()}
            className="inline-flex w-full justify-center rounded-md bg-[#087e8b] px-4 py-2 text-sm font-semibold text-white! shadow-sm transition hover:bg-[#0a5dad] hover:text-white!"
          >
            Lainnya
          </button>
        </div>
      ) : null}
    </div>
  );
}
