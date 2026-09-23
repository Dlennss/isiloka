"use client";

import * as React from "react";
import Image from "next/image";
import { FileSearch, LoaderCircle, X } from "lucide-react";
import TurnstileWidget from "@/components/TurnstileWidget";
import { UserCheckoutModal } from "@/components/user/UserCheckoutModal";
import type { UserAppBillingCheck, UserProductItem } from "@/components/user/types";
import { getRetailFeeForProduct } from "@/lib/retailRoles";
import { getBrandLogo } from "@/lib/brand-logos";

type RetailBillingEntryFlowProps = {
  title: string;
  description: string;
  placeholder: string;
  items: UserProductItem[];
  mode: "guest" | "user";
  authToken?: string;
  buyerRole?: string;
  logoSrc?: string;
  logoAlt?: string;
  minDestLength?: number;
  maxDestLength?: number;
};

const TURNSTILE_ENABLED = /^(1|true|yes|on)$/i.test(process.env.NEXT_PUBLIC_TURNSTILE_ENABLED || "");
const TURNSTILE_SITE_KEY = process.env.NEXT_PUBLIC_TURNSTILE_SITE_KEY || "";

function normalizeDest(value: string) {
  return value.replace(/\D/g, "").trim();
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

function isZeroPriceCheckProduct(item: UserProductItem) {
  return Number(item.harga_dasar_app || 0) === 0;
}

function isCheckProduct(item: UserProductItem) {
  const upperSku = String(item.sku || "").toUpperCase();
  const upperName = String(item.nama || "").toUpperCase();
  return isZeroPriceCheckProduct(item) || upperSku.startsWith("CEK") || upperName.includes("CEK ");
}

function isBillingPayProduct(item: UserProductItem) {
  const category = String(item.kategori_nama || "").toUpperCase();
  const upperSku = String(item.sku || "").toUpperCase();
  const upperName = String(item.nama || "").toUpperCase();

  if (isCheckProduct(item)) return false;
  if (category.includes("LISTRIK")) {
    if (upperName.includes("TOKEN")) return false;
    if (upperSku.startsWith("PLN") && upperSku !== "PLNB") return false;
    return upperName.includes("PASCABAYAR") || upperSku === "PLNB";
  }
  return true;
}

function toCheckVariantSKU(paySKU: string) {
  const upper = String(paySKU || "").toUpperCase().trim();
  if (!upper) return "";
  if (upper.endsWith("B")) return `${upper.slice(0, -1)}C`;
  return "";
}

function resolveCheckItem(items: UserProductItem[]) {
  const zeroPriceItems = items.filter((item) => isZeroPriceCheckProduct(item));
  if (zeroPriceItems.length === 0) return null;
  const category = String(items[0]?.kategori_nama || "").toUpperCase().trim();

  if (category.includes("LISTRIK")) {
    const plnCheck = zeroPriceItems.find((item) => String(item.sku || "").toUpperCase().trim() === "PLNC");
    if (plnCheck) return plnCheck;
  }

  const payCandidates = items.filter((item) => isBillingPayProduct(item));
  const paySkus = new Set(
    payCandidates
      .map((item) => toCheckVariantSKU(item.sku))
      .filter(Boolean),
  );

  const exactPairItems: UserProductItem[] = [];
  for (const item of zeroPriceItems) {
    const sku = String(item.sku || "").toUpperCase().trim();
    if (paySkus.has(sku)) {
      exactPairItems.push(item);
    }
  }
  if (exactPairItems.length === 1) return exactPairItems[0];

  const explicitNamedCheck = zeroPriceItems.find((item) => {
    const upperSku = String(item.sku || "").toUpperCase();
    const upperName = String(item.nama || "").toUpperCase();
    return upperSku.startsWith("CEK") || upperName.includes("CEK ") || upperSku.endsWith("C");
  });
  if (explicitNamedCheck && zeroPriceItems.length === 1) return explicitNamedCheck;
  if (exactPairItems.length > 0) return exactPairItems[0];
  return zeroPriceItems.length === 1 ? zeroPriceItems[0] : null;
}

function resolveBillingLogo(title: string, items: UserProductItem[], logoSrc?: string, logoAlt?: string) {
  if (logoSrc) {
    return { src: logoSrc, alt: logoAlt || title || "Logo layanan" };
  }

  const titleValue = String(title || "").trim().toLowerCase();
  const categoryValue = String(items[0]?.kategori_nama || "").trim().toLowerCase();
  const firstSKU = String(items[0]?.sku || "").trim().toUpperCase();

  if (titleValue.includes("listrik") || categoryValue.includes("listrik") || firstSKU.startsWith("PLN")) {
    return { src: "/images/pln/logo_pln.png", alt: "PLN" };
  }

  if (titleValue.includes("pgn") || categoryValue.includes("gas")) {
    return { src: "/images/gas/Logo_PGN.png", alt: "PGN" };
  }
  if (titleValue.includes("pdam") || categoryValue.includes("pdam")) {
    return { src: "/images/pdam/logo_pdam.png", alt: "PDAM" };
  }
  if (titleValue.includes("bpjs") || categoryValue.includes("bpjs")) {
    return { src: "/images/bpjs/icon_bpjs_kesehatan.png", alt: "BPJS" };
  }
  if (titleValue.includes("iconnet") || titleValue.includes("iconet") || categoryValue.includes("internet")) {
    return { src: "/images/internet/logo_iconet.png", alt: "Internet Pascabayar" };
  }
  const matchedBrandLogo = getBrandLogo(title);
  if (matchedBrandLogo) {
    return { src: matchedBrandLogo.src, alt: matchedBrandLogo.alt };
  }
  if (categoryValue.includes("hp pascabayar") || titleValue.includes("telkom") || titleValue.includes("indosat") || titleValue.includes("xl") || titleValue.includes("axis") || titleValue.includes("smartfren") || titleValue.includes("tri")) {
    return { src: "/images/providers/logo_telkomsel.png", alt: title || "HP Pascabayar" };
  }
  return { src: "/images/providers/logo_telkomsel.png", alt: title || "Layanan" };
}

export function RetailBillingEntryFlow({
  title,
  description,
  placeholder,
  items,
  mode,
  authToken,
  buyerRole,
  logoSrc,
  logoAlt,
  minDestLength = 5,
  maxDestLength,
}: RetailBillingEntryFlowProps) {
  const [dest, setDest] = React.useState("");
  const [checking, setChecking] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);
  const [checkOrder, setCheckOrder] = React.useState<UserAppBillingCheck | null>(null);
  const [showCheckModal, setShowCheckModal] = React.useState(false);
  const [showPayModal, setShowPayModal] = React.useState(false);
  const [turnstileToken, setTurnstileToken] = React.useState("");
  const [turnstileError, setTurnstileError] = React.useState<string | null>(null);
  const [waitingTurnstile, setWaitingTurnstile] = React.useState(false);
  const [turnstileAppearance, setTurnstileAppearance] = React.useState<"always" | "interaction-only">("interaction-only");
  const [inputInvalid, setInputInvalid] = React.useState(false);

  const cleanDest = normalizeDest(dest);
  const isReady = cleanDest.length >= minDestLength && (typeof maxDestLength !== "number" || cleanDest.length <= maxDestLength);
  const guestNeedsTurnstile = !authToken && TURNSTILE_ENABLED && Boolean(TURNSTILE_SITE_KEY);
  const guestIdentity = React.useMemo(() => buildGuestIdentity(dest), [dest]);
  const checkItem = React.useMemo(() => resolveCheckItem(items), [items]);
  const payItems = React.useMemo(() => items.filter((item) => isBillingPayProduct(item)), [items]);
  const payItem = React.useMemo(() => payItems[0] || null, [payItems]);
  const billingInquiry = checkOrder?.billing_inquiry || null;
  const feeActive = payItem ? getRetailFeeForProduct(payItem, authToken ? buyerRole : "guest") : 0;
  const productAdmin = Number(payItem?.harga_dasar_app || 0);
  const providerBill = Number(billingInquiry?.bill_amount || 0);
  const providerAdmin = Number(billingInquiry?.admin_amount || 0);
  const providerPenalty = Number(billingInquiry?.penalty_amount || 0);
  const providerTotal = Number(billingInquiry?.total_amount || 0);
  const subtotalBayar = providerTotal + productAdmin + feeActive;
  const guestQrisAdminFee = mode === "guest" && subtotalBayar > 0 ? Math.ceil(subtotalBayar * 0.007) : 0;
  const adminFeeDisplay = providerAdmin + productAdmin + feeActive + guestQrisAdminFee;
  const estimatedGuestTotal = subtotalBayar + guestQrisAdminFee;
  const isCheckPending = Boolean(checkOrder && !["success", "failed", "refunded", "expired", "cancelled"].includes(checkOrder.status));
  const isCheckFinal = Boolean(checkOrder && ["success", "failed", "refunded", "expired", "cancelled"].includes(checkOrder.status));
  const billingLogo = React.useMemo(() => resolveBillingLogo(title, items, logoSrc, logoAlt), [items, logoAlt, logoSrc, title]);
  const isPLNFlow = String(title || "").toLowerCase().includes("listrik")
    || String(items[0]?.kategori_nama || "").toLowerCase().includes("pln")
    || String(items[0]?.brand_nama || "").toLowerCase() === "pln";

  React.useEffect(() => {
    if (!checkOrder) {
      setShowCheckModal(false);
    }
  }, [checkOrder]);

  React.useEffect(() => {
    if (!showCheckModal || !isCheckFinal || !billingInquiry?.can_pay || !payItem) return;
    const timer = window.setTimeout(() => {
      setShowCheckModal(false);
      setShowPayModal(true);
    }, 450);
    return () => window.clearTimeout(timer);
  }, [billingInquiry?.can_pay, isCheckFinal, payItem, showCheckModal]);

  React.useEffect(() => {
    if (!waitingTurnstile || !turnstileToken || !checkItem) return;
    void performCheck();
     
  }, [waitingTurnstile, turnstileToken, checkItem]);

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
    setChecking(true);
    setError(null);
    try {
      const orderPayload = {
        produk_id: checkItem!.id,
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
    if (!isReady) {
      setInputInvalid(true);
      return;
    }
    if (!checkItem) {
      setError("Produk cek tagihan belum tersedia.");
      return;
    }
    if (payItems.length === 0) {
      setError("Produk pembayaran tagihan belum tersedia.");
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
    <div className="space-y-4">
      <section className={`${isPLNFlow ? "overflow-hidden rounded-[28px] border border-emerald-950/5 bg-linear-to-br from-white via-emerald-50/80 to-lime-50/70 p-4 shadow-[0_18px_42px_rgba(6,78,59,0.12)]" : "rounded-md bg-white p-4 shadow-[0_14px_34px_rgba(15,23,42,0.12)] ring-1 ring-slate-100"} ${inputInvalid ? "auth-shake" : ""}`}>
        <div className="flex items-center gap-3">
          <div className={isPLNFlow ? "grid h-14 w-14 shrink-0 place-items-center rounded-[20px] bg-white p-2 shadow-[0_12px_24px_rgba(6,78,59,0.12)] ring-1 ring-emerald-100" : "shrink-0"}>
            <Image
              src={billingLogo.src}
              alt={billingLogo.alt}
              width={44}
              height={44}
              className={isPLNFlow ? "h-full w-full object-contain" : "h-11 w-11 object-contain"}
            />
          </div>
          <div className="min-w-0">
            {isPLNFlow ? <p className="text-[10px] font-black uppercase tracking-[0.16em] text-emerald-700">Isiloka PLN</p> : null}
            <h2 className={isPLNFlow ? "mt-0.5 text-lg font-black tracking-tight text-slate-950" : "text-lg font-bold text-slate-900"}>{title}</h2>
            {description ? <p className="mt-1 text-sm text-slate-500">{description}</p> : null}
          </div>
        </div>

        <div className="mt-4 space-y-2">
          <input
            type="tel"
            inputMode="numeric"
            value={dest}
            onChange={(e) => {
              const nextValue = e.target.value.replace(/\D/g, "");
              setDest(typeof maxDestLength === "number" ? nextValue.slice(0, maxDestLength) : nextValue);
              setInputInvalid(false);
              setCheckOrder(null);
              setShowCheckModal(false);
              setShowPayModal(false);
              setError(null);
            }}
            maxLength={maxDestLength}
            placeholder={placeholder}
            className={`w-full border px-4 py-3 text-sm text-slate-900 placeholder:text-slate-400 focus:outline-none ${isPLNFlow ? "h-14 rounded-2xl font-bold focus:ring-4" : "rounded-md focus:ring-1"} ${inputInvalid ? "border-rose-300 bg-rose-50/60 focus:border-rose-400 focus:ring-rose-200" : isPLNFlow ? "border-emerald-500/20 bg-white focus:border-emerald-500 focus:ring-emerald-100" : "border-slate-200 focus:border-sky-500 focus:ring-sky-500"}`}
          />
          <button
            type="button"
            onClick={handleCheck}
            disabled={checking}
            className={isPLNFlow ? "inline-flex h-12 w-full items-center justify-center gap-2 rounded-2xl bg-[#13464b] px-4 text-sm font-black text-white shadow-[0_14px_28px_rgba(5,46,38,0.22)] ring-1 ring-lime-200/20 transition hover:opacity-95" : "inline-flex h-11 w-full items-center justify-center gap-2 rounded-2xl bg-linear-to-r from-sky-500 to-cyan-500 px-4 text-sm font-semibold text-white shadow-[0_8px_20px_rgba(15,111,203,0.24)] transition hover:opacity-95"}
          >
            <FileSearch className="h-4 w-4" />
            {checking ? (
              <span className="inline-flex items-center gap-2">
                <LoaderCircle className="h-4 w-4 animate-spin" />
                Memproses cek...
              </span>
            ) : (
              "Cek Tagihan"
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
      </section>

      {error ? (
        <section className="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">{error}</section>
      ) : null}

      {turnstileError ? (
        <section className="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">{turnstileError}</section>
      ) : null}

      {payItems.length === 0 ? (
        <section className="grid min-h-40 place-items-center rounded-md border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_10px_28px_rgba(15,23,42,0.08)]">
          Produk belum tersedia.
        </section>
      ) : null}

      {showCheckModal && checkOrder ? (
        <div className="fixed inset-0 z-70 bg-slate-950/55">
          <div className="absolute inset-x-0 bottom-0 left-1/2 w-full max-w-md -translate-x-1/2 overflow-hidden rounded-t-[28px] bg-white shadow-[0_26px_70px_rgba(15,23,42,0.38)] md:w-97.5 md:max-w-none">
            <div className="flex items-start justify-between gap-4 border-b border-slate-100 px-5 py-4">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.24em] text-sky-600">Status Cek</p>
                <h3 className="mt-1 text-lg font-bold text-slate-900">{title}</h3>
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
              <div className="rounded-2xl bg-slate-50 px-4 py-3 text-sm text-slate-600">
                <div className="flex items-center justify-between gap-3">
                  <span>Pelanggan</span>
                  <span className="font-semibold text-slate-900">{cleanDest}</span>
                </div>
                <div className="mt-2 flex items-center justify-between gap-3">
                  <span>Status</span>
                  <span className="font-semibold text-slate-900">{statusLabel(checkOrder.status)}</span>
                </div>
                {billingInquiry?.customer_name ? (
                  <div className="mt-2 flex items-center justify-between gap-3">
                    <span>Nama</span>
                    <span className="text-right font-semibold text-slate-900">{billingInquiry.customer_name}</span>
                  </div>
                ) : null}
                {billingInquiry?.usage_label ? (
                  <div className="mt-2 flex items-center justify-between gap-3">
                    <span>Pemakaian</span>
                    <span className="font-semibold text-slate-900">{billingInquiry.usage_label}</span>
                  </div>
                ) : null}
              </div>

              {isCheckPending ? (
                <div className="rounded-2xl border border-sky-100 bg-sky-50 px-4 py-4 text-sm text-sky-700">
                  <div className="flex items-start gap-3">
                    <LoaderCircle className="mt-0.5 h-5 w-5 animate-spin" />
                    <div>
                      <p className="font-semibold">Mohon tunggu.</p>
                      <p className="mt-1 text-sky-700/80">Tagihan sedang dicek.</p>
                    </div>
                  </div>
                </div>
              ) : null}

              {isCheckFinal && billingInquiry ? (
                <div className={`rounded-2xl px-4 py-4 text-sm ${billingInquiry.can_pay ? "border border-emerald-100 bg-emerald-50 text-emerald-900" : "border border-rose-100 bg-rose-50 text-rose-700"}`}>
                  <p className="font-semibold">{billingInquiry.can_pay ? "Hasil cek tagihan sudah siap" : (billingInquiry.display_message || "Cek gagal")}</p>
                  {!billingInquiry.can_pay ? null : (
                    <p className="mt-1 wrap-break-word text-sm opacity-90">{billingInquiry.display_message || billingInquiry.provider_message}</p>
                  )}

                  {billingInquiry.can_pay ? (
                    <div className="mt-4 space-y-2 rounded-2xl bg-white/80 px-4 py-3 text-slate-700 ring-1 ring-emerald-100">
                      {billingInquiry.customer_name ? (
                        <div className="flex items-center justify-between gap-3">
                          <span>Nama</span>
                          <span className="text-right font-semibold text-slate-900">{billingInquiry.customer_name}</span>
                        </div>
                      ) : null}
                      {billingInquiry.meter_type ? (
                        <div className="flex items-center justify-between gap-3">
                          <span>Jenis meter</span>
                          <span className="font-semibold text-slate-900">{billingInquiry.meter_type}</span>
                        </div>
                      ) : null}
                      {billingInquiry.period_label ? (
                        <div className="flex items-center justify-between gap-3">
                          <span>Periode</span>
                          <span className="font-semibold text-slate-900">{billingInquiry.period_label}</span>
                        </div>
                      ) : null}
                      {billingInquiry.meter_range ? (
                        <div className="flex items-center justify-between gap-3">
                          <span>Stand meter</span>
                          <span className="font-semibold text-slate-900">{billingInquiry.meter_range}</span>
                        </div>
                      ) : null}
                      <div className="flex items-center justify-between gap-3">
                        <span>Total tagihan</span>
                        <span className="font-semibold text-slate-900">{new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(providerBill > 0 ? providerBill : providerTotal)}</span>
                      </div>
                      {providerPenalty > 0 ? (
                        <div className="flex items-center justify-between gap-3">
                          <span>Denda</span>
                          <span className="font-semibold text-slate-900">{new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(providerPenalty)}</span>
                        </div>
                      ) : null}
                      <div className="flex items-center justify-between gap-3">
                        <span>Admin</span>
                        <span className="font-semibold text-slate-900">{new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(providerAdmin)}</span>
                      </div>
                      <div className="flex items-center justify-between gap-3">
                        <span>Biaya aplikasi</span>
                        <span className="font-semibold text-slate-900">{new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(adminFeeDisplay)}</span>
                      </div>
                      {billingInquiry.transaction_time ? (
                        <div className="flex items-center justify-between gap-3">
                          <span>Waktu</span>
                          <span className="font-semibold text-slate-900">{billingInquiry.transaction_time}</span>
                        </div>
                      ) : null}
                      <div className="flex items-center justify-between gap-3 border-t border-slate-200 pt-2">
                        <span>{mode === "guest" ? "Estimasi total bayar" : "Subtotal bayar"}</span>
                        <span className="font-bold text-sky-700">{new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(mode === "guest" ? estimatedGuestTotal : subtotalBayar)}</span>
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

      {payItem ? (
        <UserCheckoutModal
          open={showPayModal}
          product={payItem}
          authToken={authToken}
          buyerRole={buyerRole}
          initialDest={cleanDest}
          billingCheckRefId={checkOrder?.ref_id}
          billingInquiry={billingInquiry}
          onClose={() => setShowPayModal(false)}
        />
      ) : null}
    </div>
  );
}
