"use client";

import { useEffect, useMemo, useState } from "react";
import { SlidersHorizontal } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { fmtID } from "@/lib/format";

type AppOrderRow = {
  id: number;
  invoice_id: string;
  member_id?: number | null;
  member_nama?: string | null;
  guest_nama?: string | null;
  guest_email?: string | null;
  guest_phone?: string | null;
  produk_id: number;
  produk_sku_snapshot: string;
  produk_nama_snapshot: string;
  dest: string;
  qty: number;
  nominal: number;
  buyer_type: "user" | "guest";
  harga_dasar: number;
  fee: number;
  harga_final: number;
  status: string;
  dibuat_pada?: string | null;
  diubah_pada?: string | null;
};

type AppOrderPaymentRow = {
  id: number;
  order_id: string;
  transaction_id?: string | null;
  gross_amount: number;
  payment_type?: string | null;
  transaction_status?: string | null;
  fraud_status?: string | null;
  acquirer?: string | null;
  qr_url?: string | null;
  paid_at?: string | null;
  expired_at?: string | null;
  settlement_time?: string | null;
};

type AppOrderProviderTrxRow = {
  id: number;
  provider: string;
  ref_id: string;
  harga_provider: number;
  status: string;
  kode_respon?: string | null;
  pesan?: string | null;
  sn?: string | null;
  dibuat_pada?: string | null;
  diubah_pada?: string | null;
};

type AppOrderDetail = {
  order: AppOrderRow;
  payment?: AppOrderPaymentRow | null;
  provider_trx?: AppOrderProviderTrxRow | null;
};

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  if (!t) return {};
  return { Authorization: `Bearer ${t}` };
}

const PAGE_SIZE = 10;

function money(v: number) {
  return `Rp ${fmtID(v || 0)}`;
}

function statusTone(status: string) {
  const s = (status || "").toLowerCase();
  if (s === "success") return "border border-emerald-400 bg-emerald-100 text-emerald-900";
  if (s === "paid") return "border border-sky-400 bg-sky-100 text-sky-900";
  if (s === "processing_provider") return "border border-violet-400 bg-violet-100 text-violet-900";
  if (["failed", "cancelled", "expired"].includes(s)) {
    return "border border-rose-400 bg-rose-100 text-rose-900";
  }
  if (s === "refunded") return "border border-amber-400 bg-amber-100 text-amber-900";
  return "border border-slate-400 bg-slate-100 text-slate-800";
}

export default function AdminAppOrdersPage() {
  const [items, setItems] = useState<AppOrderRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [offset, setOffset] = useState(0);
  const [hasNext, setHasNext] = useState(false);
  const [detail, setDetail] = useState<AppOrderDetail | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [mobileFilterOpen, setMobileFilterOpen] = useState(false);

  const [draftQ, setDraftQ] = useState("");
  const [draftStatus, setDraftStatus] = useState("");
  const [draftBuyerType, setDraftBuyerType] = useState("");
  const [draftDateFrom, setDraftDateFrom] = useState("");
  const [draftDateTo, setDraftDateTo] = useState("");

  const [q, setQ] = useState("");
  const [status, setStatus] = useState("");
  const [buyerType, setBuyerType] = useState("");
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");
  const activeFilterCount = [draftQ, draftStatus, draftBuyerType, draftDateFrom, draftDateTo].filter((v) => v.trim()).length;
  const modalOpen = detailLoading || !!detail;

  async function load(nextOffset = offset) {
    setLoading(true);
    try {
      const qs = new URLSearchParams();
      if (q) qs.set("q", q);
      if (status) qs.set("status", status);
      if (buyerType) qs.set("buyer_type", buyerType);
      if (dateFrom) qs.set("date_from", dateFrom);
      if (dateTo) qs.set("date_to", dateTo);
      qs.set("limit", String(PAGE_SIZE + 1));
      qs.set("offset", String(nextOffset));

      const r = await fetch(`/api/admin/app/orders?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      const all: AppOrderRow[] = Array.isArray(j.items) ? j.items : [];
      setHasNext(all.length > PAGE_SIZE);
      setItems(all.slice(0, PAGE_SIZE));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load(offset);
     
  }, [offset]);

  useEffect(() => {
    void load(0);
     
  }, [q, status, buyerType, dateFrom, dateTo]);

  useEffect(() => {
    if (!modalOpen) return;

    const prevBodyOverflow = document.body.style.overflow;
    const prevHtmlOverflow = document.documentElement.style.overflow;
    const prevBodyTouchAction = document.body.style.touchAction;
    const prevHtmlTouchAction = document.documentElement.style.touchAction;

    document.body.style.overflow = "hidden";
    document.documentElement.style.overflow = "hidden";
    document.body.style.touchAction = "none";
    document.documentElement.style.touchAction = "none";

    return () => {
      document.body.style.overflow = prevBodyOverflow;
      document.documentElement.style.overflow = prevHtmlOverflow;
      document.body.style.touchAction = prevBodyTouchAction;
      document.documentElement.style.touchAction = prevHtmlTouchAction;
    };
  }, [modalOpen]);

  function applyFilters() {
    setOffset(0);
    setQ(draftQ.trim());
    setStatus(draftStatus);
    setBuyerType(draftBuyerType);
    setDateFrom(draftDateFrom);
    setDateTo(draftDateTo);
  }

  function resetFilters() {
    setDraftQ("");
    setDraftStatus("");
    setDraftBuyerType("");
    setDraftDateFrom("");
    setDraftDateTo("");
    setOffset(0);
    setQ("");
    setStatus("");
    setBuyerType("");
    setDateFrom("");
    setDateTo("");
  }

  async function openDetail(invoiceId: string) {
    setDetailLoading(true);
    setDetail(null);
    try {
      const r = await fetch(`/api/admin/app/orders/${encodeURIComponent(invoiceId)}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (r.ok && j?.ok && j?.item) {
        setDetail(j.item as AppOrderDetail);
      }
    } finally {
      setDetailLoading(false);
    }
  }

  const columns = useMemo<DataTableColumn<AppOrderRow>[]>(
    () => [
      {
        id: "dibuat_pada",
        header: "Waktu",
        tdClassName: "whitespace-nowrap text-slate-100",
        render: (x) => (x.dibuat_pada ? new Date(x.dibuat_pada).toLocaleString("id-ID") : "-"),
      },
      {
        id: "invoice_id",
        header: "Invoice",
        tdClassName: "whitespace-nowrap font-mono text-xs text-slate-300",
        render: (x) => x.invoice_id,
      },
      {
        id: "buyer_type",
        header: "Buyer",
        tdClassName: "whitespace-nowrap text-slate-200",
        render: (x) => (x.buyer_type === "user" ? x.member_nama || "User" : x.guest_nama || "Guest"),
      },
      {
        id: "produk",
        header: "Produk",
        tdClassName: "text-slate-200",
        render: (x) => x.produk_nama_snapshot,
      },
      {
        id: "dest",
        header: "Tujuan",
        tdClassName: "whitespace-nowrap text-slate-200",
        render: (x) => x.dest,
      },
      {
        id: "harga_final",
        header: "Total",
        tdClassName: "whitespace-nowrap text-slate-200",
        render: (x) => money(x.harga_final),
      },
      {
        id: "status",
        header: "Status",
        tdClassName: "whitespace-nowrap",
        render: (x) => (
          <span className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[11px] font-semibold capitalize ${statusTone(x.status)}`}>
            <span className="h-1.5 w-1.5 rounded-full bg-current opacity-80" />
            {x.status.replace(/_/g, " ")}
          </span>
        ),
      },
    ],
    [],
  );

  const currentPage = Math.floor(offset / PAGE_SIZE) + 1;
  const shouldShowPagination = offset > 0 || hasNext;

  return (
    <div className="space-y-4 p-2">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <div className="text-lg font-semibold tracking-tight">Order Aplikasi</div>
          <div className="text-sm text-muted-foreground">Daftar seluruh order app commerce dari user maupun guest.</div>
        </div>

        <div className="w-full md:hidden">
          <Button
            type="button"
            variant="outline"
            className="h-10 w-full justify-between border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50"
            onClick={() => setMobileFilterOpen((v) => !v)}
          >
            <span className="inline-flex items-center gap-2">
              <SlidersHorizontal className="h-4 w-4" />
              {mobileFilterOpen ? "Tutup Filter" : "Buka Filter"}
            </span>
            {activeFilterCount > 0 ? (
              <span className="rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-semibold text-emerald-700">{activeFilterCount}</span>
            ) : null}
          </Button>
        </div>
      </div>

      <div
        className={`${mobileFilterOpen ? "block" : "hidden"} rounded-2xl border border-white/10 bg-linear-to-br from-slate-900/85 via-slate-900/65 to-slate-800/45 p-3 shadow-[0_22px_48px_-34px_rgba(56,189,248,0.75)] md:block`}
      >
        <div className="mb-2 flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.12em] text-slate-300 md:hidden">
          <SlidersHorizontal className="h-3.5 w-3.5 text-emerald-600" />
          Filter Order
        </div>
        <div className="grid grid-cols-2 gap-2 xl:grid-cols-6">
          <Input
            value={draftQ}
            onChange={(e) => setDraftQ(e.target.value)}
            placeholder="Cari invoice, tujuan, produk"
            className="col-span-2 h-10 border-white/15 bg-slate-950/55 text-slate-100 placeholder:text-slate-500 xl:col-span-1"
          />
          <select
            value={draftStatus}
            onChange={(e) => setDraftStatus(e.target.value)}
            className="col-span-1 h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50 xl:col-span-1"
          >
            <option value="">semua status</option>
            <option value="pending_payment">pending_payment</option>
            <option value="paid">paid</option>
            <option value="processing_provider">processing_provider</option>
            <option value="success">success</option>
            <option value="failed">failed</option>
            <option value="expired">expired</option>
            <option value="cancelled">cancelled</option>
            <option value="refunded">refunded</option>
          </select>
          <select
            value={draftBuyerType}
            onChange={(e) => setDraftBuyerType(e.target.value)}
            className="h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50"
          >
            <option value="">semua buyer</option>
            <option value="user">user</option>
            <option value="guest">guest</option>
          </select>
          <div className="col-span-2 grid grid-cols-2 gap-2 xl:col-span-2">
            <div className="space-y-1">
              <div className="text-[11px] font-medium text-slate-400 md:hidden">Dari</div>
              <Input
                type="date"
                value={draftDateFrom}
                onChange={(e) => setDraftDateFrom(e.target.value)}
                className="h-10 border-white/15 bg-slate-950/55 text-slate-100"
              />
            </div>
            <div className="space-y-1">
              <div className="text-[11px] font-medium text-slate-400 md:hidden">Sampai</div>
              <Input
                type="date"
                value={draftDateTo}
                onChange={(e) => setDraftDateTo(e.target.value)}
                className="h-10 border-white/15 bg-slate-950/55 text-slate-100"
              />
            </div>
          </div>
          <div className="col-span-2 grid grid-cols-2 gap-2 xl:col-span-1">
            <Button
              onClick={() => {
                applyFilters();
                setMobileFilterOpen(false);
              }}
              className="h-10 bg-emerald-700 text-white hover:bg-emerald-600"
            >
              Terapkan
            </Button>
            <Button
              onClick={() => {
                resetFilters();
                setMobileFilterOpen(false);
              }}
              variant="outline"
              className="h-10 border-white/15 bg-slate-950/55 text-slate-100 hover:bg-slate-800/50"
            >
              Reset
            </Button>
          </div>
        </div>
      </div>

      {loading ? (
        <div className="rounded-md border border-white/15 bg-slate-950/50 px-4 py-10 text-center text-sm text-slate-400">
          Memuat order aplikasi...
        </div>
      ) : (
        <DataTable<AppOrderRow>
          columns={columns}
          rows={items}
          rowKey={(row) => row.id}
          emptyText="Tidak ada data order aplikasi."
          actions={{
            header: "Detail",
            align: "center",
            render: (row) => (
              <Button
                size="sm"
                variant="outline"
                className="border-white/15 bg-slate-950/55 text-slate-100 hover:bg-slate-800/50"
                onClick={() => void openDetail(row.invoice_id)}
              >
                Lihat
              </Button>
            ),
          }}
          pagination={
            shouldShowPagination
              ? {
                  page: currentPage,
                  totalPages: hasNext ? currentPage + 1 : currentPage,
                  onPrev: () => setOffset((v) => Math.max(0, v - PAGE_SIZE)),
                  onNext: () => setOffset((v) => v + PAGE_SIZE),
                  onPageChange: (nextPage) => setOffset((nextPage - 1) * PAGE_SIZE),
                  disablePrev: loading || offset === 0,
                  disableNext: loading || !hasNext,
                }
              : undefined
          }
        />
      )}

      {modalOpen ? (
        <div className="fixed inset-0 z-50 bg-slate-950/70 backdrop-blur-sm">
          <div className="flex min-h-dvh items-end justify-center p-0 sm:grid sm:min-h-full sm:place-items-center sm:p-4">
            <div className="flex max-h-[92dvh] w-full flex-col rounded-t-2xl border border-white/10 bg-slate-950/95 p-3 shadow-[0_26px_80px_-32px_rgba(34,211,238,0.55)] sm:max-h-[min(88dvh,820px)] sm:max-w-3xl sm:rounded-2xl sm:p-4">
              <div className="mb-3 flex items-start justify-between gap-3 sm:mb-4">
                <div className="min-w-0">
                  <div className="text-base font-semibold tracking-tight text-slate-100 sm:text-lg">Detail Order Aplikasi</div>
                  <div className="text-xs text-slate-400 sm:text-sm">Order, payment, dan transaksi provider dalam satu tampilan.</div>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  className="shrink-0 border-white/15 bg-slate-950/55 text-slate-100 hover:bg-slate-800/50"
                  onClick={() => setDetail(null)}
                >
                  Tutup
                </Button>
              </div>

              {detailLoading ? (
                <div className="rounded-xl border border-white/10 bg-slate-900/50 px-4 py-10 text-center text-sm text-slate-400">Memuat detail order...</div>
              ) : detail ? (
                <div className="min-h-0 space-y-3 overflow-y-auto pr-1 sm:space-y-4">
                  <section className="rounded-xl border border-white/10 bg-slate-900/55 p-3 sm:p-4">
                    <div className="mb-3 flex flex-wrap items-start justify-between gap-3">
                      <div className="min-w-0">
                        <div className="truncate font-semibold text-slate-100">{detail.order.produk_nama_snapshot}</div>
                        <div className="mt-1 break-all font-mono text-[11px] text-slate-400 sm:text-xs">{detail.order.invoice_id}</div>
                      </div>
                      <span className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[11px] font-semibold capitalize ${statusTone(detail.order.status)}`}>
                        <span className="h-1.5 w-1.5 rounded-full bg-current opacity-80" />
                        {detail.order.status.replace(/_/g, " ")}
                      </span>
                    </div>
                    <div className="grid gap-3 text-sm sm:grid-cols-2">
                      <div>
                        <div className="text-slate-400">Buyer</div>
                        <div className="mt-1 text-slate-100">{detail.order.buyer_type === "user" ? detail.order.member_nama || "User" : detail.order.guest_nama || "Guest"}</div>
                      </div>
                      <div>
                        <div className="text-slate-400">Tujuan</div>
                        <div className="mt-1 break-all text-slate-100">{detail.order.dest}</div>
                      </div>
                      <div>
                        <div className="text-slate-400">Harga Final</div>
                        <div className="mt-1 text-slate-100">{money(detail.order.harga_final)}</div>
                      </div>
                      <div>
                        <div className="text-slate-400">Dibuat</div>
                        <div className="mt-1 text-slate-100">{detail.order.dibuat_pada ? new Date(detail.order.dibuat_pada).toLocaleString("id-ID") : "-"}</div>
                      </div>
                    </div>
                  </section>

                  <section className="rounded-xl border border-white/10 bg-slate-900/55 p-3 sm:p-4">
                    <div className="mb-3 text-sm font-semibold text-slate-100">Pembayaran</div>
                    {detail.payment ? (
                      <div className="grid gap-3 text-sm sm:grid-cols-2">
                        <div>
                          <div className="text-slate-400">Status</div>
                          <div className="mt-1 text-slate-100">{detail.payment.transaction_status || "-"}</div>
                        </div>
                        <div>
                          <div className="text-slate-400">Tipe</div>
                          <div className="mt-1 text-slate-100">{detail.payment.payment_type || "-"}</div>
                        </div>
                        <div>
                          <div className="text-slate-400">Transaction ID</div>
                          <div className="mt-1 break-all font-mono text-[11px] text-slate-100 sm:text-xs">{detail.payment.transaction_id || "-"}</div>
                        </div>
                        <div>
                          <div className="text-slate-400">Acquirer</div>
                          <div className="mt-1 text-slate-100">{detail.payment.acquirer || "-"}</div>
                        </div>
                        <div>
                          <div className="text-slate-400">Paid At</div>
                          <div className="mt-1 text-slate-100">{detail.payment.paid_at ? new Date(detail.payment.paid_at).toLocaleString("id-ID") : "-"}</div>
                        </div>
                        <div>
                          <div className="text-slate-400">Expired At</div>
                          <div className="mt-1 text-slate-100">{detail.payment.expired_at ? new Date(detail.payment.expired_at).toLocaleString("id-ID") : "-"}</div>
                        </div>
                      </div>
                    ) : (
                      <div className="text-sm text-slate-400">Belum ada data pembayaran.</div>
                    )}
                  </section>

                  <section className="rounded-xl border border-white/10 bg-slate-900/55 p-3 sm:p-4">
                    <div className="mb-3 text-sm font-semibold text-slate-100">Transaksi Provider</div>
                    {detail.provider_trx ? (
                      <div className="grid gap-3 text-sm sm:grid-cols-2">
                        <div>
                          <div className="text-slate-400">Provider</div>
                          <div className="mt-1 text-slate-100">{detail.provider_trx.provider}</div>
                        </div>
                        <div>
                          <div className="text-slate-400">Status</div>
                          <div className="mt-1 text-slate-100">{detail.provider_trx.status}</div>
                        </div>
                        <div>
                          <div className="text-slate-400">Ref ID</div>
                          <div className="mt-1 break-all font-mono text-[11px] text-slate-100 sm:text-xs">{detail.provider_trx.ref_id}</div>
                        </div>
                        <div>
                          <div className="text-slate-400">Harga Provider</div>
                          <div className="mt-1 text-slate-100">{money(detail.provider_trx.harga_provider)}</div>
                        </div>
                        <div>
                          <div className="text-slate-400">Kode Respon</div>
                          <div className="mt-1 text-slate-100">{detail.provider_trx.kode_respon || "-"}</div>
                        </div>
                        <div>
                          <div className="text-slate-400">SN</div>
                          <div className="mt-1 break-all text-slate-100">{detail.provider_trx.sn || "-"}</div>
                        </div>
                        <div className="sm:col-span-2">
                          <div className="text-slate-400">Pesan</div>
                          <div className="mt-1 break-all text-slate-100">{detail.provider_trx.pesan || "-"}</div>
                        </div>
                      </div>
                    ) : (
                      <div className="text-sm text-slate-400">Transaksi provider belum tersedia.</div>
                    )}
                  </section>
                </div>
              ) : null}
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
