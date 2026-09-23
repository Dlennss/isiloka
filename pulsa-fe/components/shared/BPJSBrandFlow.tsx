"use client";

import * as React from "react";
import Image from "next/image";
import { FileSearch, LoaderCircle, X } from "lucide-react";
import TurnstileWidget from "@/components/TurnstileWidget";
import { UserCheckoutModal } from "@/components/user/UserCheckoutModal";
import type { UserAppBillingCheck, UserProductItem } from "@/components/user/types";
import { getRetailFeeForProduct } from "@/lib/retailRoles";

type BPJSBrandFlowProps = {
  items: UserProductItem[];
  authToken?: string;
  buyerRole?: string;
  initialMode?: BPJSMode;
  showModeSelector?: boolean;
};

type BPJSMode = "kesehatan" | "ketenagakerjaan";

const TURNSTILE_ENABLED = /^(1|true|yes|on)$/i.test(process.env.NEXT_PUBLIC_TURNSTILE_ENABLED || "");
const TURNSTILE_SITE_KEY = process.env.NEXT_PUBLIC_TURNSTILE_SITE_KEY || "";

function normalizeDigits(value: string) {
  return value.replace(/\D/g, "");
}

function normalizeGuestEmail(value: string) {
  return value.trim().toLowerCase();
}

function normalizeGuestPhone(value: string) {
  return value.replace(/\D/g, "");
}

function buildGuestIdentity(dest: string) {
  const phone = normalizeGuestPhone(dest);
  return {
    guestNama: "Guest Isiloka",
    guestEmail: phone ? `guest-${phone}@pulsakilat.local` : "guest@pulsakilat.local",
    guestPhone: phone,
  };
}

function getProductMap(items: UserProductItem[]) {
  const out: Record<string, UserProductItem | null> = {
    cekKesehatan: null,
    bayarKesehatan: null,
    cekKetenagakerjaan: null,
    bayarKetenagakerjaan: null,
  };

  for (const item of items) {
    const name = item.nama.trim().toLowerCase();
    if (name.includes("cek bpjs kesehatan")) out.cekKesehatan = item;
    else if (name.includes("bpjs kesehatan")) out.bayarKesehatan = item;
    else if (name.includes("cek bpjs tenaga kerja")) out.cekKetenagakerjaan = item;
    else if (name.includes("bpjs tenaga kerja")) out.bayarKetenagakerjaan = item;
  }

  return out;
}

function normalizeBillingFailureMessage(value: string) {
  const raw = String(value || "").trim();
  if (!raw) return "Cek gagal";

  const compact = raw.replace(/\s+/g, " ").trim();
  const directMatch = compact.match(/Data purchase biller.*?(?:\.|$)/i);
  if (directMatch?.[0]) {
    return directMatch[0].trim();
  }

  const gagalMatch = compact.match(/GAGAL\.\s*(.*?)(?:\s+Saldo\s|$)/i);
  if (gagalMatch?.[1]) {
    return gagalMatch[1].trim().replace(/^[:\-\s]+/, "") || "Cek gagal";
  }

  return compact.replace(/\s+Saldo\s[\d.@:.-]+.*$/i, "").trim() || "Cek gagal";
}

export function BPJSBrandFlow({ items, authToken, buyerRole, initialMode = "kesehatan", showModeSelector = true }: BPJSBrandFlowProps) {
  const [mode, setMode] = React.useState<BPJSMode>(initialMode);
  const [dest, setDest] = React.useState("");
  const [error, setError] = React.useState<string | null>(null);
  const [checking, setChecking] = React.useState(false);
  const [checkOrder, setCheckOrder] = React.useState<UserAppBillingCheck | null>(null);
  const [showCheckModal, setShowCheckModal] = React.useState(false);
  const [showPayModal, setShowPayModal] = React.useState(false);
  const [turnstileToken, setTurnstileToken] = React.useState("");
  const [turnstileError, setTurnstileError] = React.useState<string | null>(null);
  const [waitingTurnstile, setWaitingTurnstile] = React.useState(false);
  const [turnstileAppearance, setTurnstileAppearance] = React.useState<"always" | "interaction-only">("interaction-only");
  const [inputInvalid, setInputInvalid] = React.useState(false);

  const productMap = React.useMemo(() => getProductMap(items), [items]);
  const guestNeedsTurnstile = !authToken && TURNSTILE_ENABLED && Boolean(TURNSTILE_SITE_KEY);
  const checkProduct = mode === "kesehatan" ? productMap.cekKesehatan : productMap.cekKetenagakerjaan;
  const payProduct = mode === "kesehatan" ? productMap.bayarKesehatan : productMap.bayarKetenagakerjaan;
  const guestIdentity = React.useMemo(() => buildGuestIdentity(dest), [dest]);
  const billingInquiry = checkOrder?.billing_inquiry || null;
  const failureMessage = billingInquiry && !billingInquiry.can_pay
    ? (mode === "kesehatan"
        ? normalizeBillingFailureMessage(billingInquiry.provider_message || billingInquiry.display_message || "")
        : (billingInquiry.provider_message || billingInquiry.display_message || "Cek gagal"))
    : "";
  const feeActive = payProduct ? getRetailFeeForProduct(payProduct, authToken ? (buyerRole || "user") : "guest") : 0;
  const productAdmin = Number(payProduct?.harga_dasar_app || 0);
  const providerBill = Number(billingInquiry?.bill_amount || 0);
  const providerAdmin = Number(billingInquiry?.admin_amount || 0);
  const providerTotal = Number(billingInquiry?.total_amount || 0);
  const subtotalBayar = providerTotal + productAdmin + feeActive;
  const guestQrisAdminFee = !authToken && subtotalBayar > 0 ? Math.ceil(subtotalBayar * 0.007) : 0;
  const adminFeeDisplay = providerAdmin + productAdmin + feeActive + guestQrisAdminFee;
  const estimatedGuestTotal = subtotalBayar + guestQrisAdminFee;
  const isCheckPending = Boolean(checkOrder && !["success", "failed", "refunded", "expired", "cancelled"].includes(checkOrder.status));
  const isCheckFinal = Boolean(checkOrder && ["success", "failed", "refunded", "expired", "cancelled"].includes(checkOrder.status));

  React.useEffect(() => {
    setError(null);
    setCheckOrder(null);
    setShowCheckModal(false);
    setShowPayModal(false);
    setTurnstileToken("");
    setTurnstileError(null);
    setWaitingTurnstile(false);
    setTurnstileAppearance("interaction-only");
  }, [mode]);

  React.useEffect(() => {
    setMode(initialMode);
  }, [initialMode]);

  React.useEffect(() => {
    if (!checkOrder) {
      setShowCheckModal(false);
    }
  }, [checkOrder]);

  React.useEffect(() => {
    if (!showCheckModal || !isCheckFinal || !billingInquiry?.can_pay || !payProduct) return;
    const timer = window.setTimeout(() => {
      setShowCheckModal(false);
      setShowPayModal(true);
    }, 450);
    return () => window.clearTimeout(timer);
  }, [billingInquiry?.can_pay, isCheckFinal, payProduct, showCheckModal]);

  React.useEffect(() => {
    if (!waitingTurnstile || !turnstileToken || !checkProduct) return;
    void performCheck();
  }, [waitingTurnstile, turnstileToken, checkProduct]);

  React.useEffect(() => {
    if (!guestNeedsTurnstile || !waitingTurnstile || turnstileToken) return;
    const timer = window.setTimeout(() => {
      setTurnstileAppearance("always");
    }, 1200);
    return () => window.clearTimeout(timer);
  }, [guestNeedsTurnstile, waitingTurnstile, turnstileToken]);

  React.useEffect(() => {
    if (!checkOrder?.ref_id) return;
    if (["success", "failed", "refunded", "expired", "cancelled"].includes(checkOrder.status)) return;

    let cancelled = false;
    const guestHeaders: Record<string, string> = !authToken
      ? {
          "X-Guest-Email": normalizeGuestEmail(guestIdentity.guestEmail),
          "X-Guest-Phone": normalizeGuestPhone(guestIdentity.guestPhone),
        }
      : {};

    const poll = async () => {
      try {
        const res = await fetch(`/api/app/billing-checks/${encodeURIComponent(checkOrder.ref_id)}`, {
          method: "GET",
          headers: {
            ...(authToken ? { Authorization: `Bearer ${authToken}` } : {}),
            ...guestHeaders,
          },
        });
        const json = (await res.json().catch(() => ({}))) as { ok?: boolean; item?: UserAppBillingCheck };
        if (!cancelled && res.ok && json.ok && json.item) {
          setCheckOrder(json.item);
          setShowCheckModal(true);
        }
      } catch {}
    };

    const timer = window.setInterval(poll, 3000);
    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, [authToken, checkOrder?.ref_id, checkOrder?.status, guestIdentity.guestEmail, guestIdentity.guestPhone]);

  React.useEffect(() => {
    if (!inputInvalid) return;
    const timer = window.setTimeout(() => setInputInvalid(false), 520);
    return () => window.clearTimeout(timer);
  }, [inputInvalid]);

  async function performCheck() {
    const cleanDest = normalizeDigits(dest);
    setChecking(true);
    setError(null);
    try {
      const orderPayload = {
        produk_id: checkProduct!.id,
        dest: cleanDest,
        ...(authToken
          ? {}
          : {
              guest_nama: guestIdentity.guestNama,
              guest_email: guestIdentity.guestEmail,
              guest_phone: guestIdentity.guestPhone,
            }),
      };

      const orderRes = await fetch("/api/app/billing-checks", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(authToken ? { Authorization: `Bearer ${authToken}` } : {}),
          ...(!authToken && guestNeedsTurnstile ? { "X-Turnstile-Required": "1" } : {}),
          ...(!authToken && turnstileToken ? { "X-Turnstile-Token": turnstileToken } : {}),
        },
        body: JSON.stringify(orderPayload),
      });
      const orderJson = (await orderRes.json().catch(() => ({}))) as {
        ok?: boolean;
        error?: string;
        detail?: string;
        item?: UserAppBillingCheck;
      };
      if (!orderRes.ok || !orderJson.ok || !orderJson.item) {
        if ((orderJson.error || "").toLowerCase().includes("turnstile")) {
          throw new Error(orderJson.detail || "Verifikasi keamanan sudah tidak valid. Silakan verifikasi ulang lalu coba cek lagi.");
        }
        throw new Error(orderJson.error || "Gagal membuat order cek.");
      }

      const guestHeaders: Record<string, string> = !authToken
        ? {
            "X-Guest-Email": normalizeGuestEmail(guestIdentity.guestEmail),
            "X-Guest-Phone": normalizeGuestPhone(guestIdentity.guestPhone),
          }
        : {};

      const latestRes = await fetch(`/api/app/billing-checks/${encodeURIComponent(orderJson.item.ref_id)}`, {
        method: "GET",
        headers: {
          ...(authToken ? { Authorization: `Bearer ${authToken}` } : {}),
          ...guestHeaders,
        },
      });
      const latestJson = (await latestRes.json().catch(() => ({}))) as { ok?: boolean; item?: UserAppBillingCheck };

      setDest(cleanDest);
      setTurnstileToken("");
      setTurnstileError(null);
      setWaitingTurnstile(false);
      setTurnstileAppearance("interaction-only");
      setShowCheckModal(true);
      setShowPayModal(false);
      setCheckOrder(latestRes.ok && latestJson.ok && latestJson.item ? latestJson.item : orderJson.item);
    } catch (err) {
      setShowCheckModal(false);
      setCheckOrder(null);
      setError(err instanceof Error ? err.message : "Request cek gagal.");
      if (!authToken) {
        setTurnstileToken("");
        setWaitingTurnstile(false);
        setTurnstileAppearance("always");
        setTurnstileError("Verifikasi keamanan perlu diulang sebelum mengirim cek lagi.");
      }
    } finally {
      if (!authToken) {
        setTurnstileToken("");
      }
      setChecking(false);
    }
  }

  async function handleCheck() {
    const cleanDest = normalizeDigits(dest);
    if (cleanDest.length < 10 || cleanDest.length > 14) {
      setInputInvalid(true);
      return;
    }
    if (!checkProduct || !payProduct) {
      setError("Produk BPJS belum tersedia.");
      return;
    }
    if (!authToken && guestNeedsTurnstile && !turnstileToken) {
      setError(null);
      setTurnstileError(null);
      setTurnstileAppearance("interaction-only");
      setWaitingTurnstile(true);
      return;
    }
    await performCheck();
  }

  function statusLabel(status?: string) {
    switch ((status || "").toLowerCase()) {
      case "success":
        return "Cek berhasil diproses.";
      case "failed":
        return "Cek gagal diproses.";
      case "processing_provider":
        return "Cek sudah dikirim ke provider.";
      case "paid":
        return "Cek sudah dibayar dan menunggu provider.";
      default:
        return "Request cek sedang diproses.";
    }
  }

    return (
    <section className={`space-y-4 rounded-md bg-white p-4 shadow-[0_14px_34px_rgba(15,23,42,0.12)] ring-1 ring-slate-100 ${inputInvalid ? "auth-shake" : ""}`}>

      {showModeSelector ? (
        <div className="grid grid-cols-2 gap-3">
          <button
            type="button"
            onClick={() => setMode("kesehatan")}
            className={`rounded-3xl border px-4 py-4 text-left transition ${
              mode === "kesehatan" ? "border-sky-500 bg-sky-50 shadow-[0_12px_28px_rgba(15,111,203,0.14)]" : "border-slate-200 bg-white hover:border-sky-300"
            }`}
          >
            <p className="text-sm font-bold text-slate-900">BPJS Kesehatan</p>
            <p className="mt-1 text-xs text-slate-500">Iuran kesehatan</p>
          </button>

          <button
            type="button"
            onClick={() => setMode("ketenagakerjaan")}
            className={`rounded-3xl border px-4 py-4 text-left transition ${
              mode === "ketenagakerjaan" ? "border-sky-500 bg-sky-50 shadow-[0_12px_28px_rgba(15,111,203,0.14)]" : "border-slate-200 bg-white hover:border-sky-300"
            }`}
          >
            <p className="text-sm font-bold text-slate-900">BPJS Ketenagakerjaan</p>
            <p className="mt-1 text-xs text-slate-500">Iuran tenaga kerja</p>
          </button>
        </div>
      ) : null}

      <div className="space-y-3">
        <div className="flex items-center gap-3">
          <div className="shrink-0">
            <Image
              src={mode === "kesehatan" ? "/images/bpjs/icon_bpjs_kesehatan.png" : "/images/bpjs/icon_bpjs_ketenagakerjaan.png"}
              alt={mode === "kesehatan" ? "BPJS Kesehatan" : "BPJS Ketenagakerjaan"}
              width={44}
              height={44}
              className="h-11 w-11 object-contain"
            />
          </div>
          <div className="min-w-0">
            <h2 className="text-lg font-bold text-slate-900">{mode === "kesehatan" ? "BPJS Kesehatan" : "BPJS Ketenagakerjaan"}</h2>
          </div>
        </div>

        <div className="relative">
          <input
            type="tel"
            inputMode="numeric"
            maxLength={14}
            value={dest}
            onChange={(e) => {
              setDest(e.target.value.replace(/\D/g, "").slice(0, 14));
              setInputInvalid(false);
              setCheckOrder(null);
              setShowCheckModal(false);
              setShowPayModal(false);
              setError(null);
            }}
            placeholder={mode === "kesehatan" ? "Masukkan nomor VA BPJS Kesehatan" : "Masukkan nomor BPJS Ketenagakerjaan"}
            className={`w-full rounded-2xl border px-4 py-3 pl-11 text-sm text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-1 ${inputInvalid ? "border-rose-300 bg-rose-50/60 focus:border-rose-400 focus:ring-rose-200" : "border-slate-200 focus:border-sky-500 focus:ring-sky-500"}`}
          />
          <FileSearch className="absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
        </div>

        <button
          type="button"
          onClick={handleCheck}
          disabled={checking}
          className="inline-flex h-11 w-full items-center justify-center rounded-2xl bg-linear-to-r from-sky-500 to-cyan-500 px-4 text-sm font-semibold text-white shadow-[0_8px_20px_rgba(15,111,203,0.24)] transition hover:opacity-95"
        >
          {checking ? (
            <span className="inline-flex items-center gap-2">
              <LoaderCircle className="h-4 w-4 animate-spin" />
              Memproses cek...
            </span>
          ) : (
            "Cek"
          )}
        </button>

        {guestNeedsTurnstile ? (
          <div className={waitingTurnstile || turnstileError ? "mt-3" : "mt-0 h-0 overflow-hidden"}>
              <TurnstileWidget
                siteKey={TURNSTILE_SITE_KEY}
                appearance={turnstileAppearance}
                className={waitingTurnstile || turnstileError ? "" : "pointer-events-none opacity-0"}
                onToken={(token) => {
                  setTurnstileToken(token);
                  setTurnstileError(null);
                  setTurnstileAppearance("interaction-only");
                }}
                onExpire={() => {
                  setTurnstileToken("");
                  setWaitingTurnstile(false);
                }}
                onError={() => {
                  setTurnstileToken("");
                  setWaitingTurnstile(false);
                  setTurnstileAppearance("always");
                  setTurnstileError("Verifikasi keamanan belum berhasil. Klik verifikasi lalu coba cek lagi.");
                }}
              />
          </div>
        ) : null}
      </div>

      {error && error !== "Nomor pelanggan tidak valid." ? <div className="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">{error}</div> : null}
      {turnstileError ? <div className="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-700">{turnstileError}</div> : null}

      {showCheckModal && checkOrder ? (
        <div className="fixed inset-0 z-70 bg-slate-950/55">
          <div className="absolute inset-x-0 bottom-0 left-1/2 w-full max-w-md -translate-x-1/2 overflow-hidden rounded-t-[28px] bg-white shadow-[0_26px_70px_rgba(15,23,42,0.38)] md:w-97.5 md:max-w-none">
            <div className="flex items-start justify-between gap-4 border-b border-slate-100 px-5 py-4">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.24em] text-sky-600">Status Cek</p>
                <h3 className="mt-1 text-lg font-bold text-slate-900">{mode === "kesehatan" ? "BPJS Kesehatan" : "BPJS Ketenagakerjaan"}</h3>
                <p className="mt-1 text-sm text-slate-500">Ref cek: {checkOrder.ref_id}</p>
              </div>
              <button
                type="button"
                onClick={() => setShowCheckModal(false)}
                className="inline-flex h-10 w-10 items-center justify-center rounded-full border border-slate-200 text-slate-500"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <div className="space-y-4 px-5 py-5">
              {isCheckPending ? (
                <div className="rounded-2xl border border-sky-100 bg-sky-50 px-4 py-4 text-sm text-sky-700">
                  <div className="flex items-start gap-3">
                    <LoaderCircle className="mt-0.5 h-5 w-5 animate-spin" />
                    <div>
                      <p className="font-semibold">Tagihan anda sedang di cek.</p>
                      <p className="mt-1 text-sky-700/80">Mohon tunggu, sistem sedang mengambil data tagihan dari provider.</p>
                    </div>
                  </div>
                </div>
              ) : null}

              {isCheckFinal && billingInquiry ? (
                <div className={`rounded-2xl px-4 py-4 text-sm ${billingInquiry.can_pay ? "border border-emerald-100 bg-emerald-50 text-emerald-900" : "border border-rose-100 bg-rose-50 text-rose-700"}`}>
                  <p className="font-semibold">{billingInquiry.can_pay ? "Hasil cek tagihan sudah siap" : "Tagihan tidak dapat diproses"}</p>
                  {billingInquiry.can_pay ? (
                    <p className="mt-1 wrap-break-words text-sm opacity-90">{billingInquiry.provider_message}</p>
                  ) : failureMessage ? (
                    <p className="mt-1 wrap-break-words text-sm opacity-90">{failureMessage}</p>
                  ) : null}

                  {billingInquiry.can_pay ? (
                    <div className="mt-4 space-y-2 rounded-2xl bg-white/80 px-4 py-3 text-slate-700 ring-1 ring-emerald-100">
                      {billingInquiry.customer_name ? (
                        <div className="flex items-center justify-between gap-3">
                          <span>Nama</span>
                          <span className="text-right font-semibold text-slate-900">{billingInquiry.customer_name}</span>
                        </div>
                      ) : null}
                      {billingInquiry.usage_label ? (
                        <div className="flex items-center justify-between gap-3">
                          <span>Periode/Pemakaian</span>
                          <span className="font-semibold text-slate-900">{billingInquiry.usage_label}</span>
                        </div>
                      ) : null}
                      <div className="flex items-center justify-between gap-3">
                        <span>Total tagihan</span>
                        <span className="font-semibold text-slate-900">{new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(providerBill > 0 ? providerBill : providerTotal)}</span>
                      </div>
                      <div className="flex items-center justify-between gap-3">
                        <span>Admin fee</span>
                        <span className="font-semibold text-slate-900">{new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(adminFeeDisplay)}</span>
                      </div>
                      <div className="flex items-center justify-between gap-3 border-t border-slate-200 pt-2">
                        <span>{authToken ? "Subtotal bayar" : "Estimasi total bayar"}</span>
                        <span className="font-bold text-sky-700">{new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(authToken ? subtotalBayar : estimatedGuestTotal)}</span>
                      </div>
                    </div>
                  ) : null}
                </div>
              ) : null}

              <div className="flex flex-col gap-3 sm:flex-row">
                <button
                  type="button"
                  onClick={() => setShowCheckModal(false)}
                  className="inline-flex h-11 flex-1 items-center justify-center rounded-2xl border border-slate-200 px-4 text-sm font-semibold text-slate-700"
                >
                  {isCheckPending ? "Tutup" : "Selesai"}
                </button>
              </div>
            </div>
          </div>
        </div>
      ) : null}

      {payProduct ? (
        <UserCheckoutModal
          open={showPayModal}
          product={payProduct}
          authToken={authToken}
          buyerRole={buyerRole}
          initialDest={normalizeDigits(dest)}
          billingCheckRefId={checkOrder?.ref_id}
          billingInquiry={billingInquiry}
          onClose={() => setShowPayModal(false)}
        />
      ) : null}
    </section>
  );
}
