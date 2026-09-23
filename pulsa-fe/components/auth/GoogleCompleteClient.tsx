"use client";

import { useEffect, useMemo, useState } from "react";
import { getSession } from "next-auth/react";
import { useRouter, useSearchParams } from "next/navigation";

type SessionShape = {
  backendToken?: string;
  user?: { role?: string };
};

function normalizeGoogleNext(raw: string, fallback = "/user") {
  let current = (raw || "").trim() || fallback;

  for (let i = 0; i < 6; i += 1) {
    try {
      const url = current.startsWith("http://") || current.startsWith("https://")
        ? new URL(current)
        : new URL(current, "https://pulsakilat.local");

      if (url.pathname === "/login") {
        const nested = (url.searchParams.get("callbackUrl") || "").trim();
        if (nested) {
          current = nested;
          continue;
        }
      }

      if (url.pathname === "/auth/google/complete") {
        const nested = (url.searchParams.get("next") || "").trim();
        if (nested) {
          current = nested;
          continue;
        }
      }

      return `${url.pathname}${url.search}${url.hash}` || fallback;
    } catch {
      return current.startsWith("/") ? current : fallback;
    }
  }

  return current.startsWith("/") ? current : fallback;
}

function toDashboardByRole(role?: string | null, fallbackNext?: string) {
  const next = (fallbackNext || "").trim();
  if (next.startsWith("/")) return next;
  const r = (role || "").toLowerCase();
  if (r === "admin" || r === "staff") return "/dashboard/admin";
  if (r === "analis" || r === "analyst") return "/dashboard/master/operator";
  if (r === "master" || r === "marketing") return "/dashboard/master";
  if (r === "user" || r === "member" || r === "agent" || r === "agent_member" || r === "master_member") return "/user";
  if (r === "operator_trx") return "/dashboard/operator";
  if (r === "operator_wallet") return "/dashboard/wallet";
  return "/user";
}

export default function GoogleCompleteClient() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [message, setMessage] = useState("Menyelesaikan login Google...");
  const [error, setError] = useState("");
  const refundContext = useMemo(
    () => ({
      invoiceId: (searchParams.get("refund_invoice_id") || "").trim(),
      guestEmail: (searchParams.get("guest_email") || "").trim(),
      guestPhone: (searchParams.get("guest_phone") || "").trim(),
      next: normalizeGoogleNext((searchParams.get("next") || "").trim(), "/user"),
    }),
    [searchParams],
  );

  useEffect(() => {
    let cancelled = false;

    const finish = async () => {
      for (let attempt = 0; attempt < 12; attempt += 1) {
        const sess = (await getSession().catch(() => null)) as SessionShape | null;
        const backendToken = String(sess?.backendToken || "").trim();
        if (backendToken) {
          localStorage.setItem("auth_token", backendToken);
          localStorage.setItem("auth_source", "google");
          await fetch("/api/auth/persist", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ token: backendToken }),
          }).catch(() => undefined);

          let refundClaimed = false;
          if (refundContext.invoiceId && refundContext.guestEmail && refundContext.guestPhone) {
            setMessage("Menghubungkan refund ke akun Google anda...");
            try {
              const res = await fetch("/api/app/me/refunds/claim", {
                method: "POST",
                headers: {
                  "Content-Type": "application/json",
                  Authorization: `Bearer ${backendToken}`,
                },
                body: JSON.stringify({
                  invoice_id: refundContext.invoiceId,
                  guest_email: refundContext.guestEmail,
                  guest_phone: refundContext.guestPhone,
                }),
              });
              refundClaimed = res.ok;
            } catch {}
          }

          if (!cancelled) {
            const target = new URL(toDashboardByRole(sess?.user?.role || "user", refundContext.next), window.location.origin);
            if (refundClaimed) {
              target.searchParams.set("refund", "1");
            }
            router.replace(`${target.pathname}${target.search}`);
          }
          return;
        }
        await new Promise((resolve) => window.setTimeout(resolve, 500));
      }

      if (!cancelled) {
        setError("Login Google belum selesai. Coba ulangi sekali lagi.");
      }
    };

    void finish();
    return () => {
      cancelled = true;
    };
  }, [refundContext.guestEmail, refundContext.guestPhone, refundContext.invoiceId, refundContext.next, router]);

  return (
    <section className="w-full max-w-md rounded-2xl border border-slate-200 bg-white p-6 text-center shadow-[0_20px_48px_rgba(15,23,42,0.12)]">
      <h1 className="text-lg font-semibold text-slate-900">Login Google</h1>
      <p className="mt-2 text-sm text-slate-500">{error || message}</p>
    </section>
  );
}
