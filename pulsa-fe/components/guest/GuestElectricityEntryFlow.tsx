"use client";

import * as React from "react";
import Image from "next/image";
import type { UserAppBillingCheck, UserProductItem } from "@/components/user/types";
import { GuestProductGrid } from "@/components/guest/GuestProductGrid";

type GuestElectricityEntryFlowProps = {
  title: string;
  description: string;
  placeholder: string;
  items: UserProductItem[];
  checkProduct?: UserProductItem | null;
  authToken?: string;
};

function normalizeDigits(value: string) {
  return value.replace(/\D/g, "");
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

function normalizeCheckErrorMessage(value: string) {
  const raw = String(value || "").trim();
  if (!raw) return "tidak valid";
  const lowered = raw.toLowerCase();
  if (lowered.includes("nomor tujuan salah") || lowered.includes("tidak valid")) {
    return "tidak valid";
  }
  return "tidak valid";
}

function isInvalidCheckMessage(value: string) {
  const lowered = String(value || "").trim().toLowerCase();
  return lowered.includes("nomor tujuan salah") || lowered.includes("tidak valid");
}

function normalizeElectricityGroupLabel(item: UserProductItem) {
  const fromGroup = String(item.group_name || "").trim().toUpperCase();
  if (fromGroup.includes("FULL REPLY")) return "FULL REPLY";
  if (fromGroup.includes("STANDAR REPLY")) return "STANDAR REPLY";
  if (fromGroup.includes("PROMO")) return "PROMO";

  const upperName = String(item.nama || "").toUpperCase();
  if (upperName.includes("FULL REPLY")) return "FULL REPLY";
  if (upperName.includes("PROMO")) return "PROMO";
  return "STANDAR REPLY";
}

function getElectricityGroupPriority(label: string) {
  if (label === "PROMO") return 0;
  if (label === "STANDAR REPLY") return 1;
  if (label === "FULL REPLY") return 2;
  return 3;
}

export function GuestElectricityEntryFlow({
  title,
  description,
  placeholder,
  items,
  checkProduct,
  authToken,
}: GuestElectricityEntryFlowProps) {
  const [dest, setDest] = React.useState("");
  const [checking, setChecking] = React.useState(false);
  const [checkError, setCheckError] = React.useState<string | null>(null);
  const [checkOrder, setCheckOrder] = React.useState<UserAppBillingCheck | null>(null);
  const [inputInvalid, setInputInvalid] = React.useState(false);
  const cleanDest = normalizeDigits(dest);
  const guestIdentity = React.useMemo(() => buildGuestIdentity(dest), [dest]);
  const isReadyForCheck = cleanDest.length >= 11 && cleanDest.length <= 12;
  const groupedProducts = React.useMemo(() => {
    const grouped = new Map<string, UserProductItem[]>();
    for (const item of items) {
      const label = normalizeElectricityGroupLabel(item);
      const bucket = grouped.get(label) || [];
      bucket.push(item);
      grouped.set(label, bucket);
    }

    return Array.from(grouped.entries())
      .map(([label, groupItems]) => ({
        label,
        items: [...groupItems].sort((a, b) => Number(a.nominal || 0) - Number(b.nominal || 0)),
      }))
      .sort((a, b) => {
        const priority = getElectricityGroupPriority(a.label) - getElectricityGroupPriority(b.label);
        if (priority !== 0) return priority;
        return a.label.localeCompare(b.label);
      });
  }, [items]);
  const [selectedGroup, setSelectedGroup] = React.useState("");

  React.useEffect(() => {
    if (!groupedProducts.length) {
      setSelectedGroup("");
      return;
    }

    setSelectedGroup((current) => {
      if (current && groupedProducts.some((group) => group.label === current)) return current;
      return groupedProducts[0].label;
    });
  }, [groupedProducts]);

  const activeProducts = React.useMemo(() => {
    if (!groupedProducts.length) return items;
    return groupedProducts.find((group) => group.label === selectedGroup)?.items || groupedProducts[0]?.items || items;
  }, [groupedProducts, items, selectedGroup]);

  React.useEffect(() => {
    setCheckError(null);
    setCheckOrder(null);
    if (!checkProduct || !isReadyForCheck) return;

    let cancelled = false;
    const timer = window.setTimeout(() => {
      void (async () => {
        setChecking(true);
        try {
          const res = await fetch("/api/app/billing-checks", {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              ...(authToken ? { Authorization: `Bearer ${authToken}` } : {}),
            },
            body: JSON.stringify({
              produk_id: checkProduct.id,
              dest: cleanDest,
              ...(!authToken
                ? {
                    guest_nama: guestIdentity.guestNama,
                    guest_email: guestIdentity.guestEmail,
                    guest_phone: guestIdentity.guestPhone,
                  }
                : {}),
            }),
          });
          const json = (await res.json().catch(() => ({}))) as { ok?: boolean; item?: UserAppBillingCheck; error?: string };
          if (cancelled) return;
          if (!res.ok || !json.ok || !json.item) {
            throw new Error(json.error || "Gagal memeriksa nomor meter");
          }
          setCheckOrder(json.item);
          setCheckError(null);
        } catch (err) {
          if (cancelled) return;
          setCheckOrder(null);
          setCheckError(normalizeCheckErrorMessage(err instanceof Error ? err.message : ""));
        } finally {
          if (!cancelled) setChecking(false);
        }
      })();
    }, 550);

    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [authToken, checkProduct, cleanDest, guestIdentity.guestEmail, guestIdentity.guestNama, guestIdentity.guestPhone, isReadyForCheck]);

  React.useEffect(() => {
    if (!inputInvalid) return;
    const timer = window.setTimeout(() => setInputInvalid(false), 420);
    return () => window.clearTimeout(timer);
  }, [inputInvalid]);

  const billingInfo = checkOrder?.billing_inquiry || null;
  const rawCheckDisplay = String(billingInfo?.display_message || "").trim();
  const checkDisplay = isInvalidCheckMessage(rawCheckDisplay) ? "" : rawCheckDisplay;
  const derivedCheckError = !checking && !checkError && isInvalidCheckMessage(rawCheckDisplay) ? "tidak valid" : null;
  const customerName = String(billingInfo?.customer_name || "").trim();
  const destinationLine =
    !checking && (checkError || derivedCheckError)
      ? `Nomor tujuan: ${cleanDest || dest} tidak valid`
      : !checking && checkDisplay
        ? `Nomor tujuan: ${checkDisplay}`
      : cleanDest
        ? `Nomor tujuan: ${cleanDest}`
        : "";
  React.useEffect(() => {
    if (!checkOrder?.ref_id) return;
    if (customerName || checkDisplay) return;
    if (["success", "failed", "refunded", "expired", "cancelled"].includes(String(checkOrder.status || "").toLowerCase())) return;

    let cancelled = false;
    const poll = async () => {
      try {
        const res = await fetch(`/api/app/billing-checks/${encodeURIComponent(checkOrder.ref_id)}`, {
          method: "GET",
          headers: {
            ...(authToken ? { Authorization: `Bearer ${authToken}` } : {}),
          },
        });
        const json = (await res.json().catch(() => ({}))) as { ok?: boolean; item?: UserAppBillingCheck };
        if (!cancelled && res.ok && json.ok && json.item) {
          setCheckOrder(json.item);
        }
      } catch {}
    };

    void poll();
    const timer = window.setInterval(poll, 1200);
    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, [authToken, checkDisplay, checkOrder?.ref_id, checkOrder?.status, customerName]);

  return (
    <div className="space-y-4">
      <section className="overflow-hidden rounded-[28px] border border-emerald-950/5 bg-linear-to-br from-white via-emerald-50/80 to-lime-50/70 p-4 shadow-[0_18px_42px_rgba(6,78,59,0.12)]">
        <div className="flex items-center gap-3">
          <div className="grid h-14 w-14 shrink-0 place-items-center rounded-[20px] bg-white p-2 shadow-[0_12px_24px_rgba(6,78,59,0.12)] ring-1 ring-emerald-100">
            <Image
              src="/images/pln/logo_pln.png"
              alt="PLN"
              width={44}
              height={44}
              className="h-full w-full object-contain"
            />
          </div>
          <div className="min-w-0">
            <p className="text-[10px] font-black uppercase tracking-[0.16em] text-emerald-700">Isiloka PLN</p>
            <h1 className="mt-0.5 text-lg font-black tracking-tight text-slate-950">{title}</h1>
            {description ? <p className="mt-1 text-sm text-slate-500">{description}</p> : null}
          </div>
        </div>

        <div className="mt-4 space-y-2">
          <input
            type="tel"
            inputMode="numeric"
            value={dest}
            onChange={(e) => {
              setDest(normalizeDigits(e.target.value).slice(0, 12));
              setInputInvalid(false);
            }}
            placeholder={placeholder}
            onBlur={() => {
              if (cleanDest && cleanDest.length < 11) {
                setInputInvalid(true);
              }
            }}
            className={`h-14 w-full rounded-2xl px-4 py-3 text-sm font-bold text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-4 ${
              inputInvalid
                ? "auth-shake border border-rose-300 bg-rose-50 focus:border-rose-400 focus:ring-rose-100"
                : "border border-emerald-500/20 bg-white focus:border-emerald-500 focus:ring-emerald-100"
            }`}
          />
          {destinationLine ? <p className="text-xs text-slate-500">{destinationLine}</p> : null}
          {inputInvalid ? <p className="text-xs text-rose-700">Nomor meter PLN harus 11 sampai 12 digit.</p> : null}
          {checking ? <p className="text-xs font-medium text-sky-700">Memeriksa nomor meter PLN...</p> : null}
        </div>
      </section>

      {items.length === 0 ? (
        <section className="grid min-h-40 place-items-center rounded-[28px] border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_10px_28px_rgba(15,23,42,0.08)]">
          Produk belum tersedia.
        </section>
      ) : (
        <div className="space-y-3">
          {groupedProducts.length > 1 ? (
            <div className="mb-1 flex snap-x snap-mandatory gap-1.5 overflow-x-auto pb-1 [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
              {groupedProducts.map((group) => (
                <button
                  key={group.label}
                  type="button"
                  onClick={() => setSelectedGroup(group.label)}
                  className={`shrink-0 snap-start whitespace-nowrap rounded-full px-2.5 py-2 text-[10px] font-semibold leading-tight transition ${
                    selectedGroup === group.label
                      ? "bg-[#13464b] text-lime-100 shadow-[0_8px_18px_rgba(5,46,38,0.20)]"
                      : "bg-white text-slate-700 ring-1 ring-slate-200 hover:bg-slate-50"
                  }`}
                >
                  {group.label}
                </button>
              ))}
            </div>
          ) : null}

          <GuestProductGrid
            items={activeProducts}
            isLoggedIn={Boolean(authToken)}
            authToken={authToken}
            initialDest={cleanDest}
            destLabel="Nomor Meter PLN"
            destPlaceholder={placeholder}
            enableGuestHint={false}
            showBrandBanner={false}
          />
        </div>
      )}
    </div>
  );
}
