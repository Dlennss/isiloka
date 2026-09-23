"use client";

import * as React from "react";
import { LoaderCircle, QrCode, X } from "lucide-react";
import { useRouter } from "next/navigation";
import TurnstileWidget from "@/components/TurnstileWidget";
import type { UserAppBillingInquiry, UserAppOrder, UserAppOrderPayment, UserProductItem } from "@/components/user/types";
import { getRetailFeeForProduct } from "@/lib/retailRoles";
import { saveGuestTransaction } from "@/lib/guest-transaction-storage";

type UserCheckoutModalProps = {
  open: boolean;
  product: UserProductItem | null;
  authToken?: string;
  buyerRole?: string;
  initialDest?: string;
  destLabel?: string;
  destPlaceholder?: string;
  destMode?: "single" | "ml_id_server" | "alphanumeric";
  billingCheckRefId?: string;
  billingInquiry?: UserAppBillingInquiry | null;
  turnstileToken?: string;
  turnstileRequired?: boolean;
  onClose: () => void;
};

type ApiErrorResponse = {
  ok?: boolean;
  error?: string;
};

type ApiItemResponse<T> = {
  ok?: boolean;
  item?: T;
  error?: string;
};

type ApiItemsResponse<T> = {
  ok?: boolean;
  items?: T[];
  error?: string;
};

type CheckoutQrisItem = {
  ref_id: string;
  amount: number;
  fee_admin: number;
  gross_amount: number;
  status: string;
  payment_type?: string;
  transaction_id?: string;
  qr_url?: string;
  expired_at?: string;
};

const TURNSTILE_ENABLED = /^(1|true|yes|on)$/i.test(process.env.NEXT_PUBLIC_TURNSTILE_ENABLED || "");
const TURNSTILE_SITE_KEY = process.env.NEXT_PUBLIC_TURNSTILE_SITE_KEY || "";

function formatRupiah(value: number) {
  return `Rp ${new Intl.NumberFormat("id-ID").format(value || 0)}`;
}

function normalizeGuestEmail(value: string) {
  return value.trim().toLowerCase();
}

function normalizeGuestPhone(value: string) {
  return value.replace(/\D/g, "");
}

function normalizeDestInput(value: string, maxLength = 13) {
  return value.replace(/\D/g, "").slice(0, maxLength);
}

function normalizeAlphaNumericDestInput(value: string, maxLength = 32) {
  return value.replace(/[^a-zA-Z0-9]/g, "").slice(0, maxLength);
}

function normalizeServerInput(value: string, maxLength = 4) {
  return value.replace(/\D/g, "").slice(0, maxLength);
}

function isMobilePhoneDestValid(value: string) {
  return /^08\d{6,11}$/.test(value);
}

function buildGuestIdentity(dest: string) {
  const phone = normalizeGuestPhone(dest);
  return {
    guestNama: "Guest Isiloka",
    guestEmail: phone ? `guest-${phone}@pulsakilat.local` : "guest@pulsakilat.local",
    guestPhone: phone,
  };
}

function normalizeRetailRole(role?: string | null) {
	const value = String(role || "").trim().toLowerCase();
	if (value === "master") return "master";
	if (value === "agent") return "agent";
	if (value === "user") return "user";
	return "guest";
}

function getProductSubtotalPriceForRole(product: UserProductItem | null, role?: string | null) {
	if (!product) return 0;
	switch (normalizeRetailRole(role)) {
		case "master":
			return Number(product.harga_master_final || 0);
		case "agent":
			return Number(product.harga_agent_final || 0);
		case "user":
			return Number(product.harga_user_final || 0);
		default:
			return Number(product.harga_guest_final || 0);
	}
}

function extractRequestedNominalLabel(product: UserProductItem, nominalValue: number) {
  if (product.tipe_harga === "OPEN_AMOUNT") {
    return nominalValue > 0 ? formatRupiah(nominalValue) : "-";
  }

  const upper = String(product.nama || "").toUpperCase();
  const dotted = upper.match(/(\d{1,3}(?:\.\d{3})+)/);
  if (dotted) return `Rp ${dotted[1]}`;

  const compact = upper.match(/(\d+)\s*K\b/);
  if (compact) return formatRupiah(Number.parseInt(compact[1], 10) * 1000);

  return formatRupiah(Number(product.nominal || 0));
}

function getSummaryPrimaryLabel(product: UserProductItem) {
  const category = String(product.kategori_nama || "").toUpperCase();
  if (category.includes("PASCABAYAR") || category.includes("PDAM") || category.includes("BPJS") || category.includes("GAS")) {
    return "Tagihan";
  }
  if (category.includes("E-MONEY") || category.includes("PAKET") || category.includes("GAME") || category.includes("TV")) {
    return "Produk";
  }
  return "Nominal pulsa";
}

function shouldShowSummaryPrimaryRow(product: UserProductItem) {
  const category = String(product.kategori_nama || "").toUpperCase();
  const upperName = String(product.nama || "").toUpperCase();
  return !category.includes("E-MONEY") && !category.includes("PAKET DATA") && !category.includes("GAME") && !(category.includes("LISTRIK") && upperName.includes("TOKEN LISTRIK"));
}

function getBillingDetailConfig(product: UserProductItem | null) {
  const category = String(product?.kategori_nama || "").toUpperCase();
  const name = String(product?.nama || "").toUpperCase();
  if (category.includes("GAS") || name.includes("PGN")) {
    return {
      title: "Detail tagihan PGN",
      accountLabel: "ID pelanggan",
      usageLabel: "Pemakaian",
      adminLabel: "Admin PGN",
      showMeterType: false,
      showPeriod: false,
      showMeterRange: false,
    };
  }
  return {
    title: "Detail tagihan PLN",
    accountLabel: "IDPEL / nomor meter",
    usageLabel: "Pemakaian",
    adminLabel: "Admin PLN",
    showMeterType: true,
    showPeriod: true,
    showMeterRange: true,
  };
}

function formatRemaining(expiredAt?: string | null) {
  if (!expiredAt) return "Menunggu pembayaran";
  const diff = new Date(expiredAt).getTime() - Date.now();
  if (Number.isNaN(diff) || diff <= 0) return "QR sudah kedaluwarsa";
  const totalSeconds = Math.floor(diff / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `Berlaku ${minutes}:${String(seconds).padStart(2, "0")}`;
}

function formatDateTime(value?: string | null) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString("id-ID");
}

function parsePaymentBreakdown(rawRequest?: string | null) {
  if (!rawRequest) return { walletDebit: 0, creditDebit: 0, qrisAmount: 0, totalAmount: 0 };
  try {
    const parsed = JSON.parse(rawRequest) as {
      wallet_debit?: number;
      credit_debit?: number;
      qris_amount?: number;
      total_amount?: number;
    };
    return {
      walletDebit: Number(parsed.wallet_debit || 0),
      creditDebit: Number(parsed.credit_debit || 0),
      qrisAmount: Number(parsed.qris_amount || 0),
      totalAmount: Number(parsed.total_amount || 0),
    };
  } catch {
    return { walletDebit: 0, creditDebit: 0, qrisAmount: 0, totalAmount: 0 };
  }
}

function isEwalletProduct(product: UserProductItem | null) {
  const category = String(product?.kategori_nama || "").toUpperCase();
  return category.includes("E-MONEY") || category.includes("E-WALLET");
}

function friendlyCheckoutError(error: unknown) {
  const message = error instanceof Error ? error.message : String(error || "");
  if (/produk (?:tidak aktif|tidak tersedia|retail wajib memakai provider Pulsa24Jam)|provider retail tidak valid/i.test(message)) {
    return "Produk ini sudah tidak tersedia untuk pembelian. Muat ulang halaman agar daftar produk memakai katalog Pulsa24Jam terbaru.";
  }
  if (/pulsa24jam client belum tersedia|koneksi katalog Pulsa24Jam belum dikonfigurasi|PULSA24JAM_.*belum|missing required env: PULSA24JAM/i.test(message)) {
    return "Koneksi Pulsa24Jam belum aktif. Isi credential Pulsa24Jam di backend dulu agar transaksi bisa dikirim ke provider.";
  }
  if (/deposit qris|commands tidak didukung|internal error|\{\s*"?ok"?\s*:/i.test(message)) {
    return "Layanan pembayaran sedang tidak tersedia. Transaksi belum diproses dan saldo Anda tidak terpotong.";
  }
  return message || "Terjadi kesalahan saat memproses checkout.";
}

function checkoutQrisPayment(order: UserAppOrder, qris: CheckoutQrisItem): UserAppOrderPayment {
  return {
    id: 0,
    app_order_id: order.id,
    order_id: order.invoice_id,
    transaction_id: qris.transaction_id || qris.ref_id,
    gross_amount: Number(qris.gross_amount || qris.amount || order.harga_final || 0),
    payment_type: qris.payment_type || "qris",
    transaction_status: qris.status || "pending",
    fraud_status: null,
    acquirer: "pulsa24jam",
    qr_url: qris.qr_url || null,
    raw_request: JSON.stringify({
      wallet_debit: 0,
      credit_debit: 0,
      qris_amount: Number(qris.gross_amount || qris.amount || order.harga_final || 0),
      total_amount: Number(qris.gross_amount || qris.amount || order.harga_final || 0),
      deposit_ref_id: qris.ref_id,
    }),
    expired_at: qris.expired_at || null,
  };
}

async function ensureFreshProductAvailable(product: UserProductItem) {
  const params = new URLSearchParams();
  if (product.kategori_id) params.set("kategori_id", String(product.kategori_id));
  if (product.brand_id) params.set("brand_id", String(product.brand_id));
  if (!params.has("kategori_id") || !params.has("brand_id")) {
    params.set("q", product.sku || product.nama || String(product.id));
  }

  const res = await fetch(`/api/app/produk?${params.toString()}`, { cache: "no-store" });
  const json = (await res.json().catch(() => ({}))) as ApiItemsResponse<UserProductItem>;
  if (!res.ok || !json.ok || !Array.isArray(json.items)) {
    throw new Error(json.error || "Gagal memuat ulang katalog produk.");
  }

  const freshProduct = json.items.find((item) => Number(item.id) === Number(product.id));
  if (!freshProduct || !freshProduct.aktif) {
    throw new Error("produk tidak tersedia");
  }
}

export function UserCheckoutModal({
  open,
  product,
  authToken,
  buyerRole,
  initialDest = "",
  destLabel = "Nomor Tujuan",
  destPlaceholder = "Masukkan nomor tujuan",
  destMode = "single",
  billingCheckRefId,
  billingInquiry,
  turnstileToken,
  turnstileRequired,
  onClose,
}: UserCheckoutModalProps) {
  const router = useRouter();
  const [dest, setDest] = React.useState("");
  const [destServer, setDestServer] = React.useState("");
  const [nominal, setNominal] = React.useState("");
  const [error, setError] = React.useState<string | null>(null);
  const [loading, setLoading] = React.useState(false);
  const [checkingStatus, setCheckingStatus] = React.useState(false);
  const [order, setOrder] = React.useState<UserAppOrder | null>(null);
  const [payment, setPayment] = React.useState<UserAppOrderPayment | null>(null);
  const [checkoutQris, setCheckoutQris] = React.useState<CheckoutQrisItem | null>(null);
  const [retailSaldo, setRetailSaldo] = React.useState<number | null>(null);
  const [guestTurnstileToken, setGuestTurnstileToken] = React.useState("");
  const [waitingTurnstile, setWaitingTurnstile] = React.useState(false);
  const [turnstileError, setTurnstileError] = React.useState<string | null>(null);
  const [turnstileHint, setTurnstileHint] = React.useState<string | null>(null);
  const [turnstileAppearance, setTurnstileAppearance] = React.useState<"always" | "interaction-only">("interaction-only");
  const [, forceTick] = React.useReducer((v) => v + 1, 0);
  const refundRecoveryAttemptedRef = React.useRef(false);
  const checkoutQrisCompletingRef = React.useRef(false);
  const effectiveAuthToken = authToken?.trim() || "";
  const activeTurnstileToken = turnstileToken || guestTurnstileToken;
  const guestNeedsTurnstile = !effectiveAuthToken && TURNSTILE_ENABLED && Boolean(TURNSTILE_SITE_KEY);

  React.useEffect(() => {
    if (!open) return;
    setDest(initialDest);
    setDestServer("");
    setNominal("");
    setError(null);
    setLoading(false);
    setOrder(null);
    setPayment(null);
    setCheckoutQris(null);
    setRetailSaldo(null);
    setGuestTurnstileToken("");
    setWaitingTurnstile(false);
    setTurnstileError(null);
    setTurnstileHint(null);
    setTurnstileAppearance("interaction-only");
    refundRecoveryAttemptedRef.current = false;
    checkoutQrisCompletingRef.current = false;
  }, [open, initialDest, product?.id]);

  React.useEffect(() => {
    if (!open || !effectiveAuthToken) return;

    let cancelled = false;
    const loadSaldo = async () => {
      try {
        const profileRes = await fetch("/api/me/profile", {
          method: "GET",
          headers: { Authorization: `Bearer ${effectiveAuthToken}` },
        });
        const profileJson = (await profileRes.json().catch(() => ({}))) as { ok?: boolean; profile?: { saldo?: number } };
        if (!cancelled) {
          setRetailSaldo(profileRes.ok && profileJson.ok ? Number(profileJson.profile?.saldo || 0) : 0);
        }
      } catch {
        if (!cancelled) {
          setRetailSaldo(0);
        }
      }
    };

    void loadSaldo();
    return () => {
      cancelled = true;
    };
  }, [open, effectiveAuthToken]);

  React.useEffect(() => {
    if (!open) return;

    const prevBodyOverflow = document.body.style.overflow;
    const prevHtmlOverflow = document.documentElement.style.overflow;
    const prevBodyTouchAction = document.body.style.touchAction;

    document.body.style.overflow = "hidden";
    document.documentElement.style.overflow = "hidden";
    document.body.style.touchAction = "none";

    return () => {
      document.body.style.overflow = prevBodyOverflow;
      document.documentElement.style.overflow = prevHtmlOverflow;
      document.body.style.touchAction = prevBodyTouchAction;
    };
  }, [open, initialDest]);

  React.useEffect(() => {
    if (!payment?.expired_at) return;
    const timer = window.setInterval(() => forceTick(), 1000);
    return () => window.clearInterval(timer);
  }, [payment?.expired_at]);

  React.useEffect(() => {
    if (!waitingTurnstile || !activeTurnstileToken) return;
    const form = document.getElementById("guest-checkout-form") as HTMLFormElement | null;
    if (!form) return;
    setTurnstileHint(null);
    form.requestSubmit();
    setWaitingTurnstile(false);
  }, [waitingTurnstile, activeTurnstileToken]);

  React.useEffect(() => {
    if (!guestNeedsTurnstile || !waitingTurnstile || activeTurnstileToken) return;
    const timer = window.setTimeout(() => {
      setTurnstileAppearance("always");
      setTurnstileHint("Selesaikan verifikasi keamanan lalu lanjutkan pembayaran.");
    }, 1800);
    return () => window.clearTimeout(timer);
  }, [guestNeedsTurnstile, waitingTurnstile, activeTurnstileToken]);

  const isFixed = product?.tipe_harga === "FIXED";
  const billingTotalAmount = Number(billingInquiry?.total_amount || 0);
  const billingBillAmount = Number(billingInquiry?.bill_amount || 0);
  const billingAdminAmount = Number(billingInquiry?.admin_amount || 0);
  const isBillingPayment = Boolean(product && billingCheckRefId && billingTotalAmount > 0);
  const effectiveRole = effectiveAuthToken ? (buyerRole || "user") : "guest";
  const feeActive = product ? getRetailFeeForProduct(product, effectiveRole) : 0;
  const backendFixedSubtotal = getProductSubtotalPriceForRole(product, effectiveRole);
  const baseCharge = Number(product?.harga_dasar_app || 0);
  const nominalValue = isFixed ? Number(product?.nominal || 0) : Number.parseInt(nominal.replace(/\D/g, ""), 10) || 0;
  const hargaSebelumFeeAdmin = isBillingPayment
    ? billingTotalAmount + baseCharge + feeActive
    : isFixed
    ? backendFixedSubtotal > 0
      ? backendFixedSubtotal
      : nominalValue > 0
        ? nominalValue + feeActive
        : 0
    : nominalValue > 0
      ? nominalValue + baseCharge + feeActive
      : 0;
  const walletSaldo = Math.max(0, retailSaldo || 0);
  const estimatedWalletDebit = effectiveAuthToken ? Math.min(hargaSebelumFeeAdmin, walletSaldo) : 0;
  const estimatedQrisAmount = effectiveAuthToken ? Math.max(hargaSebelumFeeAdmin - estimatedWalletDebit, 0) : 0;
  const needsTopup = Boolean(effectiveAuthToken) && estimatedQrisAmount > 0;
  const feeAdminQris = 0;
  const totalBayar = hargaSebelumFeeAdmin;
  const ewalletProduct = isEwalletProduct(product);
  const ewalletNominal = ewalletProduct ? Math.max(0, isFixed ? Number(product?.nominal || 0) : nominalValue) : 0;
  const ewalletServiceFee = ewalletProduct ? Math.max(0, totalBayar - ewalletNominal) : 0;
  const billingDisplayTotalTagihan = billingBillAmount > 0 ? billingBillAmount : billingTotalAmount;
  const billingDisplayAdminFee = billingTotalAmount > 0
    ? Math.max(totalBayar - billingDisplayTotalTagihan, 0)
    : billingAdminAmount + baseCharge + feeActive + feeAdminQris;
  const billingAccountNumber = (order?.dest || dest || initialDest || "").trim();
  const billingDetailConfig = getBillingDetailConfig(product);
  const guestIdentity = buildGuestIdentity(dest);
  const requiresMobilePhoneDest = React.useMemo(() => {
    const category = String(product?.kategori_nama || "").toUpperCase();
    return category.includes("PULSA")
      || category.includes("DATA")
      || category.includes("PAKET")
      || category.includes("E-MONEY")
      || category.includes("E-WALLET");
  }, [product?.kategori_nama]);

  React.useEffect(() => {
    if (!open || !order || order.status !== "failed" || order.buyer_type !== "guest") return;
    if (refundRecoveryAttemptedRef.current) return;

    const guestEmail = normalizeGuestEmail(String(order.guest_email || guestIdentity.guestEmail || ""));
    const guestPhone = normalizeGuestPhone(String(order.guest_phone || guestIdentity.guestPhone || ""));
    if (!order.invoice_id || !guestEmail || !guestPhone) return;
    refundRecoveryAttemptedRef.current = true;

    if (effectiveAuthToken) {
      const timer = window.setTimeout(() => {
        void (async () => {
          try {
            const res = await fetch("/api/app/me/refunds/claim", {
              method: "POST",
              headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${effectiveAuthToken}`,
              },
              body: JSON.stringify({
                invoice_id: order.invoice_id,
                guest_email: guestEmail,
                guest_phone: guestPhone,
              }),
            });
            if (res.ok) {
              router.replace("/user?refund=1");
              return;
            }
          } catch {}

          refundRecoveryAttemptedRef.current = false;
        })();
      }, 150);
      return () => window.clearTimeout(timer);
    }

    const target = new URLSearchParams({
      refund_invoice_id: order.invoice_id,
      guest_email: guestEmail,
      guest_phone: guestPhone,
    });
    const timer = window.setTimeout(() => {
      window.location.assign(`/register?${target.toString()}`);
    }, 150);
    return () => window.clearTimeout(timer);
  }, [effectiveAuthToken, guestIdentity.guestEmail, guestIdentity.guestPhone, open, order, router]);

  async function fetchLatestOrder(invoiceId: string) {
    const res = await fetch(`/api/app/orders/${encodeURIComponent(invoiceId)}`, {
      method: "GET",
      headers: {
        ...(effectiveAuthToken ? { Authorization: `Bearer ${effectiveAuthToken}` } : {}),
        ...(!effectiveAuthToken
          ? {
              "X-Guest-Email": normalizeGuestEmail(guestIdentity.guestEmail),
              "X-Guest-Phone": normalizeGuestPhone(guestIdentity.guestPhone),
            }
          : {}),
      } satisfies Record<string, string>,
    });
    const json = (await res.json().catch(() => ({}))) as ApiItemResponse<UserAppOrder> & ApiErrorResponse;
    if (!res.ok || !json.ok || !json.item) {
      throw new Error(json.error || "Gagal mengecek status pembayaran.");
    }
    return json.item;
  }

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const currentProduct = product;
    if (!currentProduct) return;
    setError(null);

    const cleanDest = dest.trim();
    const cleanServer = destServer.trim();
    if (!cleanDest) {
      setError(`${destLabel} wajib diisi.`);
      return;
    }
    if (destMode === "ml_id_server") {
      if (!cleanServer) {
        setError("Server wajib diisi.");
        return;
      }
      if (cleanDest.length > 15) {
        setError(`${destLabel} maksimal 15 digit.`);
        return;
      }
      if (cleanServer.length > 6) {
        setError("Server maksimal 6 digit.");
        return;
      }
    }
    if (requiresMobilePhoneDest && !isMobilePhoneDestValid(cleanDest)) {
      setError(`${destLabel} harus diawali 08 dan maksimal 13 digit.`);
      return;
    }
    if (!isFixed && nominalValue <= 0) {
      setError("Nominal wajib diisi.");
      return;
    }
    const finalDest = destMode === "ml_id_server" ? `${cleanDest}${cleanServer}` : cleanDest;

    setLoading(true);
    try {
      await ensureFreshProductAvailable(currentProduct);

      const orderPayload = {
        produk_id: currentProduct.id,
        dest: finalDest,
        qty: isFixed ? 1 : nominalValue,
        ...(isBillingPayment && billingCheckRefId ? { source_check_ref_id: billingCheckRefId } : {}),
        ...(effectiveAuthToken
          ? {}
          : {
              guest_nama: guestIdentity.guestNama,
              guest_email: guestIdentity.guestEmail,
              guest_phone: guestIdentity.guestPhone,
            }),
      };

      const orderRes = await fetch("/api/app/orders", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(effectiveAuthToken ? { Authorization: `Bearer ${effectiveAuthToken}` } : {}),
          ...(!effectiveAuthToken && (turnstileRequired || guestNeedsTurnstile) ? { "X-Turnstile-Required": "1" } : {}),
          ...(!effectiveAuthToken && activeTurnstileToken ? { "X-Turnstile-Token": activeTurnstileToken } : {}),
        },
        body: JSON.stringify(orderPayload),
      });
      const orderJson = (await orderRes.json().catch(() => ({}))) as ApiItemResponse<UserAppOrder> & ApiErrorResponse;
      if (!orderRes.ok || !orderJson.ok || !orderJson.item) {
        throw new Error(orderJson.error || "Gagal membuat order.");
      }

      const guestHeaders: Record<string, string> = !effectiveAuthToken
        ? {
            "X-Guest-Email": normalizeGuestEmail(guestIdentity.guestEmail),
            "X-Guest-Phone": normalizeGuestPhone(guestIdentity.guestPhone),
          }
        : {};

      const payRes = await fetch(`/api/app/orders/${encodeURIComponent(orderJson.item.invoice_id)}/pay`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(effectiveAuthToken ? { Authorization: `Bearer ${effectiveAuthToken}` } : {}),
          ...guestHeaders,
        } satisfies Record<string, string>,
        body: "{}",
      });
      const payJson = (await payRes.json().catch(() => ({}))) as ApiItemResponse<UserAppOrderPayment> & ApiErrorResponse;
      if (!payRes.ok || !payJson.ok || !payJson.item) {
        const paymentError = payJson.error || "Gagal membuat pembayaran.";
        if (/saldo utama(?: dan saldo kredit)? tidak cukup/i.test(paymentError)) {
          const requiredTopup = Math.max(100_000, Math.ceil(estimatedQrisAmount || Number(orderJson.item.harga_final || 0)));
          onClose();
          router.push(`/user/account/topup?amount=${requiredTopup}`);
          return;
        }
        throw new Error(paymentError);
      }

      setOrder(orderJson.item);
      setPayment(payJson.item);
      try {
        const latestOrder = await fetchLatestOrder(orderJson.item.invoice_id);
        setOrder(latestOrder);
      } catch {}
      // Simpan transaksi guest ke localStorage
      if (buyerRole === "guest" || !authToken) {
        try {
          const o = orderJson.item;
          saveGuestTransaction({
            invoice_id: o.invoice_id,
            guest_email: String(o.guest_email || guestIdentity.guestEmail || "").trim().toLowerCase(),
            guest_phone: String(o.guest_phone || guestIdentity.guestPhone || "").replace(/\D/g, ""),
            dest: o.dest || "",
            title: o.produk_nama_snapshot || product?.nama || "Transaksi Guest",
            created_at: o.dibuat_pada || new Date().toISOString(),
            updated_at: o.diubah_pada || o.dibuat_pada || new Date().toISOString(),
            status: o.status || "pending_payment",
            amount: o.harga_final || 0,
            serial_number: o.sn || "",
          });
        } catch {}
      }
      setTurnstileHint(null);
    } catch (err) {
      if (err instanceof Error && /turnstile/i.test(err.message)) {
        setGuestTurnstileToken("");
      }
      setError(friendlyCheckoutError(err));
    } finally {
      setLoading(false);
    }
  }

  async function handleCheckStatus() {
    if (!order?.invoice_id) return;
    setCheckingStatus(true);
    setError(null);
    try {
      const latestOrder = await fetchLatestOrder(order.invoice_id);
      setOrder(latestOrder);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gagal mengecek status pembayaran.");
    } finally {
      setCheckingStatus(false);
    }
  }

  React.useEffect(() => {
    if (!order?.invoice_id || !checkoutQris?.ref_id || !effectiveAuthToken) return;
    const currentStatus = String(checkoutQris.status || "").toLowerCase();
    if (currentStatus === "approved" || currentStatus === "rejected") return;

    let cancelled = false;
    const refreshQris = async () => {
      try {
        const params = new URLSearchParams({ ref_id: checkoutQris.ref_id, refresh: "1" });
        const statusRes = await fetch(`/api/me/deposit/request/qris/status?${params.toString()}`, {
          headers: { Authorization: `Bearer ${effectiveAuthToken}` },
          cache: "no-store",
        });
        const statusJson = (await statusRes.json().catch(() => ({}))) as ApiItemResponse<CheckoutQrisItem> & ApiErrorResponse;
        if (!statusRes.ok || !statusJson.ok || !statusJson.item) {
          throw new Error(statusJson.error || "Gagal mengecek status QRIS Pulsa24Jam.");
        }
        if (cancelled) return;

        const nextQris = statusJson.item;
        const nextStatus = String(nextQris.status || "").toLowerCase();
        setCheckoutQris(nextQris);
        setPayment(checkoutQrisPayment(order, nextQris));

        if (nextStatus !== "approved" || checkoutQrisCompletingRef.current) return;
        checkoutQrisCompletingRef.current = true;

        const payRes = await fetch(`/api/app/orders/${encodeURIComponent(order.invoice_id)}/pay`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${effectiveAuthToken}`,
          },
          body: "{}",
        });
        const payJson = (await payRes.json().catch(() => ({}))) as ApiItemResponse<UserAppOrderPayment> & ApiErrorResponse;
        if (!payRes.ok || !payJson.ok || !payJson.item) {
          checkoutQrisCompletingRef.current = false;
          throw new Error(payJson.error || "QRIS berhasil, tetapi transaksi belum dapat dilanjutkan.");
        }
        if (cancelled) return;
        setPayment(payJson.item);
        setCheckoutQris(null);
        setOrder(await fetchLatestOrder(order.invoice_id));
      } catch (err) {
        if (!cancelled) setError(err instanceof Error ? err.message : "Gagal mengecek pembayaran QRIS.");
      }
    };

    void refreshQris();
    const timer = window.setInterval(refreshQris, 4000);
    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, [checkoutQris?.ref_id, checkoutQris?.status, effectiveAuthToken, order?.invoice_id]);

  React.useEffect(() => {
    if (!order?.invoice_id) return;
    if (order.status === "success" || order.status === "failed" || order.status === "refunded" || order.status === "expired" || order.status === "cancelled") {
      return;
    }

    let cancelled = false;
    const guestHeaders: Record<string, string> = !effectiveAuthToken
      ? {
          "X-Guest-Email": normalizeGuestEmail(guestIdentity.guestEmail),
          "X-Guest-Phone": normalizeGuestPhone(guestIdentity.guestPhone),
        }
      : {};

    const poll = async () => {
      try {
        const res = await fetch(`/api/app/orders/${encodeURIComponent(order.invoice_id)}`, {
          method: "GET",
          headers: {
            ...(effectiveAuthToken ? { Authorization: `Bearer ${effectiveAuthToken}` } : {}),
            ...guestHeaders,
          } satisfies Record<string, string>,
        });
        const json = (await res.json().catch(() => ({}))) as ApiItemResponse<UserAppOrder> & ApiErrorResponse;
        if (!cancelled && res.ok && json.ok && json.item) {
          setOrder((prev) => {
            if (!prev) return json.item!;
            if (prev.status === json.item!.status && prev.sn === json.item!.sn && prev.diubah_pada === json.item!.diubah_pada) {
              return prev;
            }
            return json.item!;
          });
        }
      } catch {}
    };

    void poll();
    const timer = window.setInterval(poll, 3000);
    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, [effectiveAuthToken, guestIdentity.guestEmail, guestIdentity.guestPhone, order?.invoice_id, order?.status]);

  if (!open || !product) return null;

  const orderStatus = order?.status || "pending_payment";
  const isAwaitingPayment = orderStatus === "pending_payment";
  const isPaymentReceived = orderStatus === "paid";
  const isProviderProcessing = orderStatus === "processing_provider";
  const isOrderSuccess = orderStatus === "success";
  const isOrderRefunded = orderStatus === "refunded";
  const isOrderFailed = orderStatus === "failed";
  const isOrderExpired = orderStatus === "expired" || orderStatus === "cancelled";
  const checkoutQrisStatus = String(checkoutQris?.status || "").toLowerCase();
  const isCheckoutQrisRejected = checkoutQrisStatus === "rejected";
  const isFinalState = isOrderSuccess || isOrderRefunded || isOrderFailed || isOrderExpired || isCheckoutQrisRejected;
  const showQRCode = isAwaitingPayment && !isOrderExpired && !isCheckoutQrisRejected && Boolean(payment?.qr_url);
  const breakdown = parsePaymentBreakdown(payment?.raw_request);
  const walletUsed = breakdown.walletDebit;
  const qrisAmount = breakdown.qrisAmount || payment?.gross_amount || 0;
  const qrisAdminFeeActual = order ? Math.max(Number(order.fee || 0) - feeActive, 0) : feeAdminQris;
  const totalAmount = walletUsed + qrisAmount;
  const isWalletOnly = ["wallet", "balance"].includes((payment?.payment_type || "").toLowerCase());
  const usesQris = qrisAmount > 0;
  const paymentMethodLabel = usesQris
    ? walletUsed > 0
      ? "Saldo + QRIS"
      : "QRIS"
    : "Saldo Utama";
  const isGuestFailed = isOrderFailed && order?.buyer_type === "guest";

  let statusText = payment?.transaction_status || "pending";
  let statusTone = "text-amber-600";
  let statusTitle = "Menunggu pembayaran";
  let statusDescription = "Selesaikan pembayaran untuk melanjutkan transaksi.";

  if (isPaymentReceived) {
    statusText = "Pembayaran diterima";
    statusTone = "text-sky-600";
    statusTitle = "Pembayaran berhasil";
    statusDescription = "Transaksi anda sedang diproses.";
  } else if (isProviderProcessing) {
    statusText = "Diproses provider";
    statusTone = "text-violet-600";
    statusTitle = "Pembayaran berhasil";
    statusDescription = "Transaksi anda sedang diproses.";
  } else if (isOrderSuccess) {
    statusText = "Transaksi berhasil";
    statusTone = "text-emerald-600";
    statusTitle = "Transaksi berhasil";
    statusDescription = "Pembayaran berhasil dan transaksi sukses diproses.";
  } else if (isOrderRefunded) {
    statusText = "Dana dikembalikan";
    statusTone = "text-amber-700";
    statusTitle = "Dana dikembalikan ke saldo";
    statusDescription = "Transaksi gagal di provider dan dana sudah dikembalikan ke saldo akun anda.";
  } else if (isOrderFailed) {
    statusText = "Transaksi gagal";
    statusTone = "text-rose-600";
    statusTitle = "Transaksi gagal";
    statusDescription = isGuestFailed
      ? "Transaksi gagal. Anda akan diarahkan ke halaman daftar agar saldo refund otomatis masuk ke akun baru."
      : "Transaksi tidak dapat diproses oleh provider.";
  } else if (isOrderExpired) {
    statusText = orderStatus === "expired" ? "Pembayaran kedaluwarsa" : "Pembayaran dibatalkan";
    statusTone = "text-rose-600";
    statusTitle = statusText;
    statusDescription = usesQris ? "QRIS tidak lagi aktif. Buat order baru untuk melanjutkan transaksi." : "Pembayaran tidak lagi aktif. Buat order baru untuk melanjutkan transaksi.";
  } else if (isCheckoutQrisRejected) {
    statusText = "QRIS kedaluwarsa";
    statusTone = "text-rose-600";
    statusTitle = "Pembayaran tidak selesai";
    statusDescription = "QRIS sudah tidak aktif. Buat transaksi baru untuk melanjutkan.";
  }
  if (isAwaitingPayment && walletUsed > 0) {
    statusDescription = "Saldo sudah terpotong. Selesaikan pembayaran untuk sisa tagihan.";
  }

  function handleDownloadProof() {
    void (async () => {
      if (!order || !payment) return;

      const [{ default: jsPDF }, { default: autoTable }] = await Promise.all([import("jspdf"), import("jspdf-autotable")]);
      const doc = new jsPDF({ unit: "pt", format: "a4" });
      const pageWidth = doc.internal.pageSize.getWidth();
      const centerX = pageWidth / 2;

      doc.setFillColor(15, 111, 203);
      doc.rect(0, 0, pageWidth, 96, "F");
      doc.setFillColor(47, 146, 223);
      doc.rect(0, 96, pageWidth, 14, "F");
      doc.setTextColor(255, 255, 255);
      doc.setFont("helvetica", "bold");
      doc.setFontSize(22);
      doc.text("Bukti Pembayaran", centerX, 42, { align: "center" });
      doc.setFont("helvetica", "normal");
      doc.setFontSize(11);
      doc.text("Isiloka - QRIS Payment Receipt", centerX, 61, { align: "center" });
      doc.setFontSize(10);
      doc.text(formatDateTime(payment.paid_at || payment.settlement_time || order.diubah_pada), centerX, 79, { align: "center" });

      doc.setTextColor(31, 41, 55);
      doc.setFillColor(248, 250, 252);
      doc.roundedRect(40, 132, pageWidth - 80, 88, 18, 18, "F");
      doc.setFont("helvetica", "bold");
      doc.setFontSize(13);
      doc.setTextColor(100, 116, 139);
      doc.text("TOTAL PEMBAYARAN", 56, 158);
      doc.setFontSize(24);
      doc.setTextColor(15, 23, 42);
      doc.text(formatRupiah(payment.gross_amount), 56, 188);
      doc.setFont("helvetica", "normal");
      doc.setFontSize(11);
      doc.setTextColor(71, 85, 105);
      doc.text(order.produk_nama_snapshot, 56, 208);

      doc.setFillColor(236, 253, 245);
      doc.roundedRect(pageWidth - 178, 146, 122, 26, 13, 13, "F");
      doc.setFont("helvetica", "bold");
      doc.setFontSize(10);
      doc.setTextColor(5, 150, 105);
      doc.text("PEMBAYARAN BERHASIL", pageWidth - 166, 163);

      doc.setFont("helvetica", "normal");
      doc.setFontSize(11);
      doc.setTextColor(71, 85, 105);
      doc.text(`Invoice: ${order.invoice_id}`, 40, 248);

      autoTable(doc, {
        startY: 268,
        theme: "plain",
        styles: {
          fontSize: 10.5,
          cellPadding: 9,
          textColor: [31, 41, 55],
          lineColor: [226, 232, 240],
          lineWidth: 0.6,
        },
        headStyles: {
          fillColor: [241, 245, 249],
          textColor: [15, 23, 42],
          fontStyle: "bold",
        },
        body: [
          ["Tujuan", order.dest],
          ["Nama", order.member_nama || order.guest_nama || "-"],
          ["Total", formatRupiah(payment.gross_amount)],
          ["Status Order", order.status],
          ["Status Payment", payment.transaction_status || "-"],
          ["Transaction ID", payment.transaction_id || "-"],
          ["Acquirer", payment.acquirer || "-"],
          ["Dibuat", formatDateTime(order.dibuat_pada)],
          ["Dibayar", formatDateTime(payment.paid_at || payment.settlement_time || order.diubah_pada)],
        ],
        columnStyles: {
          0: { cellWidth: 136, fontStyle: "bold", textColor: [100, 116, 139] },
          1: { cellWidth: pageWidth - 216 },
        },
        didDrawCell: (data) => {
          if (data.section === "body") {
            doc.setDrawColor(226, 232, 240);
            doc.line(data.cell.x, data.cell.y + data.cell.height, data.cell.x + data.cell.width, data.cell.y + data.cell.height);
          }
        },
      });

      const finalY = (doc as unknown as { lastAutoTable?: { finalY?: number } }).lastAutoTable?.finalY || 360;
      doc.setFillColor(248, 250, 252);
      doc.roundedRect(40, finalY + 20, pageWidth - 80, 54, 14, 14, "F");
      doc.setFont("helvetica", "bold");
      doc.setFontSize(10);
      doc.setTextColor(100, 116, 139);
      doc.text("Catatan", 56, finalY + 41);
      doc.setFont("helvetica", "normal");
      doc.text("Dokumen ini dibuat otomatis oleh sistem Isiloka dan valid sebagai bukti pembayaran.", 56, finalY + 58);

      doc.save(`bukti-pembayaran-${order.invoice_id}.pdf`);
    })();
  }


  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-slate-950/40 px-4 py-6">
      <div className="max-h-[88vh] w-full max-w-md overflow-y-auto rounded-[28px] bg-white p-5 shadow-[0_18px_46px_rgba(15,23,42,0.24)] md:w-97.5 md:max-w-none">
        <div className="flex items-start justify-between gap-3">
          <div>
            <p className="text-xs font-semibold tracking-wide text-slate-500 uppercase">Checkout</p>
            <h2 className="mt-1 text-lg font-bold text-slate-900">{product.nama}</h2>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="grid h-10 w-10 place-items-center rounded-2xl bg-slate-100 text-slate-500"
            aria-label="Tutup modal"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        {payment && order ? (
          <div className="mt-5 space-y-4">
            {showQRCode ? (
              <div className="rounded-md border border-sky-100 bg-sky-50 p-1 text-center">
                {payment.qr_url ? (
                   
                  <img src={payment.qr_url} alt={`QRIS ${order.invoice_id}`} className="mx-auto h-56 w-56 bg-white" />
                ) : (
                  <div className="mx-auto grid h-56 w-56 place-items-center rounded-2xl bg-white">
                    <QrCode className="h-8 w-8 text-sky-600" />
                  </div>
                )}
                <p className="mt-1 text-xs text-slate-500">{formatRemaining(payment.expired_at)}</p>
                {walletUsed > 0 ? <p className="mt-1 text-xs font-medium text-sky-700">Saldo terpakai {formatRupiah(walletUsed)}</p> : null}
              </div>
            ) : (
              <div className={`rounded-3xl border p-5 text-center ${isOrderSuccess ? "border-emerald-100 bg-emerald-50" : isOrderRefunded ? "border-amber-100 bg-amber-50" : isOrderFailed || isOrderExpired ? "border-rose-100 bg-rose-50" : isProviderProcessing ? "border-violet-100 bg-violet-50" : "border-sky-100 bg-sky-50"}`}>
                <div className={`mx-auto grid h-14 w-14 place-items-center rounded-2xl bg-white shadow-sm ${isOrderSuccess ? "text-emerald-600" : isOrderRefunded ? "text-amber-600" : isOrderFailed || isOrderExpired ? "text-rose-600" : isProviderProcessing ? "text-violet-600" : "text-sky-600"}`}>
                  <QrCode className="h-6 w-6" />
                </div>
                <p className={`mt-3 text-base font-bold ${statusTone}`}>{statusTitle}</p>
                <p className="mt-1 text-sm text-slate-600">{statusDescription}</p>
              </div>
            )}

            <div className="rounded-3xl bg-slate-50 p-4 text-sm text-slate-600">
              <div className="flex items-center justify-between gap-3">
                <span>Invoice</span>
                <div className="flex items-center gap-2">
                  <span className="text-xs font-medium text-slate-900">{order.invoice_id}</span>
                </div>
              </div>
              <div className="mt-2 flex items-center justify-between gap-3">
                <span>Tujuan</span>
              <span className="font-semibold text-slate-900">{order.dest}</span>
              </div>
              <div className="mt-2 flex items-center justify-between gap-3">
                <span>Status</span>
                <span className={`font-semibold ${statusTone}`}>
                  {statusText}
                </span>
              </div>
              {isBillingPayment && billingInquiry ? (
                <div className="mt-3 rounded-2xl border border-slate-200 bg-white px-4 py-3">
                  <p className="text-xs font-semibold tracking-wide text-slate-500 uppercase">{billingDetailConfig.title}</p>
                  {billingInquiry.customer_name ? (
                    <div className="mt-2 flex items-center justify-between gap-3 text-sm">
                      <span>Nama pelanggan</span>
                      <span className="text-right font-semibold text-slate-900">{billingInquiry.customer_name}</span>
                    </div>
                  ) : null}
                  {billingAccountNumber ? (
                    <div className="mt-2 flex items-center justify-between gap-3 text-sm">
                      <span>{billingDetailConfig.accountLabel}</span>
                      <span className="text-right font-semibold text-slate-900">{billingAccountNumber}</span>
                    </div>
                  ) : null}
                  {billingInquiry.usage_label ? (
                    <div className="mt-2 flex items-center justify-between gap-3 text-sm">
                      <span>{billingDetailConfig.usageLabel}</span>
                      <span className="text-right font-semibold text-slate-900">{billingInquiry.usage_label}</span>
                    </div>
                  ) : null}
                  {billingDetailConfig.showMeterType && billingInquiry.meter_type ? (
                    <div className="mt-2 flex items-center justify-between gap-3 text-sm">
                      <span>Tarif / daya</span>
                      <span className="text-right font-semibold text-slate-900">{billingInquiry.meter_type}</span>
                    </div>
                  ) : null}
                  {billingDetailConfig.showPeriod && billingInquiry.period_label ? (
                    <div className="mt-2 flex items-center justify-between gap-3 text-sm">
                      <span>Periode</span>
                      <span className="text-right font-semibold text-slate-900">{billingInquiry.period_label}</span>
                    </div>
                  ) : null}
                  {billingDetailConfig.showMeterRange && billingInquiry.meter_range ? (
                    <div className="mt-2 flex items-center justify-between gap-3 text-sm">
                      <span>Stand meter</span>
                      <span className="text-right font-semibold text-slate-900">{billingInquiry.meter_range}</span>
                    </div>
                  ) : null}
                  <div className="mt-2 flex items-center justify-between gap-3 text-sm">
                    <span>Nilai tagihan</span>
                    <span className="text-right font-semibold text-slate-900">{formatRupiah(billingDisplayTotalTagihan)}</span>
                  </div>
                  <div className="mt-2 flex items-center justify-between gap-3 text-sm">
                    <span>{billingDetailConfig.adminLabel}</span>
                    <span className="text-right font-semibold text-slate-900">{formatRupiah(billingAdminAmount)}</span>
                  </div>
                  <div className="mt-2 flex items-center justify-between gap-3 border-t border-slate-100 pt-2 text-sm">
                    <span>Total tagihan</span>
                    <span className="text-right font-bold text-slate-900">{formatRupiah(billingTotalAmount)}</span>
                  </div>
                </div>
              ) : null}
              {order?.sn ? (
                <div className="mt-2 flex items-center justify-between gap-3">
                  <span>SN</span>
                  <span className="text-right font-semibold text-slate-900">{order.sn}</span>
                </div>
              ) : null}
              {qrisAdminFeeActual > 0 ? (
                <div className="mt-2 flex items-center justify-between gap-3">
                  <span>Fee admin</span>
                  <span className="font-semibold text-slate-900">{formatRupiah(qrisAdminFeeActual)}</span>
                </div>
              ) : null}
              {walletUsed > 0 ? (
                <div className="mt-2 flex items-center justify-between gap-3">
                  <span>Pakai saldo</span>
                  <span className="font-semibold text-slate-900">{formatRupiah(walletUsed)}</span>
                </div>
              ) : null}
              {usesQris ? (
                <div className="mt-2 flex items-center justify-between gap-3">
                  <span>Bayar QRIS</span>
                  <span className="font-semibold text-slate-900">{formatRupiah(qrisAmount)}</span>
                </div>
              ) : null}
              <div className="mt-2 flex items-center justify-between gap-3">
                <span>Total bayar</span>
                <span className="font-semibold text-sky-700">{formatRupiah(totalAmount)}</span>
              </div>
              <div className="mt-2 flex items-center justify-between gap-3">
                <span>Metode bayar</span>
                <span className="font-semibold text-slate-900">{paymentMethodLabel}</span>
              </div>
            </div>

            <div className={`grid gap-3 ${showQRCode ? "grid-cols-2" : "grid-cols-1"}`}>
              {showQRCode ? (
                <a
                  href={payment.qr_url || "#"}
                  target="_blank"
                  rel="noreferrer"
                  className="inline-flex h-11 w-full items-center justify-center rounded-2xl bg-linear-to-r from-[#087e8b] to-[#087e8b] px-4 text-sm font-semibold text-white shadow-[0_8px_20px_rgba(15,111,203,0.24)]"
                >
                  Buka QRIS
                </a>
              ) : null}
              {isOrderSuccess || isOrderRefunded ? (
                <button
                  type="button"
                  onClick={handleDownloadProof}
                  className={`inline-flex h-11 w-full items-center justify-center rounded-2xl px-4 text-sm font-semibold text-white ${isOrderRefunded ? "bg-linear-to-r from-amber-500 to-orange-500 shadow-[0_8px_20px_rgba(245,158,11,0.24)]" : "bg-linear-to-r from-emerald-500 to-teal-500 shadow-[0_8px_20px_rgba(16,185,129,0.24)]"}`}
                >
                  Download Bukti Pembayaran
                </button>
              ) : !isFinalState ? (
                <button
                  type="button"
                  onClick={handleCheckStatus}
                  disabled={checkingStatus}
                  className="inline-flex h-11 w-full items-center justify-center gap-2 rounded-2xl border border-slate-200 bg-white px-4 text-sm font-semibold text-slate-700 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {checkingStatus ? <LoaderCircle className="h-4 w-4 animate-spin" /> : null}
                  {checkingStatus ? "Mengecek..." : isWalletOnly ? "Cek Status Provider" : "Cek Status"}
                </button>
              ) : null}
            </div>
          </div>
        ) : (
          <form id="guest-checkout-form" onSubmit={handleSubmit} className="mt-5 space-y-4">
            <div className="rounded-3xl bg-slate-50 p-4 text-sm text-slate-600">
              {isBillingPayment && billingInquiry ? (
                <div className="rounded-2xl border border-slate-200 bg-white px-4 py-3">
                  <p className="text-xs font-semibold tracking-wide text-slate-500 uppercase">{billingDetailConfig.title}</p>
                  {billingInquiry.customer_name ? (
                    <div className="mt-2 flex items-center justify-between gap-3">
                      <span>Nama pelanggan</span>
                      <span className="text-right font-semibold text-slate-900">{billingInquiry.customer_name}</span>
                    </div>
                  ) : null}
                  {billingAccountNumber ? (
                    <div className="mt-2 flex items-center justify-between gap-3">
                      <span>{billingDetailConfig.accountLabel}</span>
                      <span className="text-right font-semibold text-slate-900">{billingAccountNumber}</span>
                    </div>
                  ) : null}
                  {billingInquiry.usage_label ? (
                    <div className="mt-2 flex items-center justify-between gap-3">
                      <span>{billingDetailConfig.usageLabel}</span>
                      <span className="text-right font-semibold text-slate-900">{billingInquiry.usage_label}</span>
                    </div>
                  ) : null}
                  {billingDetailConfig.showMeterType && billingInquiry.meter_type ? (
                    <div className="mt-2 flex items-center justify-between gap-3">
                      <span>Tarif / daya</span>
                      <span className="text-right font-semibold text-slate-900">{billingInquiry.meter_type}</span>
                    </div>
                  ) : null}
                  {billingDetailConfig.showPeriod && billingInquiry.period_label ? (
                    <div className="mt-2 flex items-center justify-between gap-3">
                      <span>Periode</span>
                      <span className="text-right font-semibold text-slate-900">{billingInquiry.period_label}</span>
                    </div>
                  ) : null}
                  {billingDetailConfig.showMeterRange && billingInquiry.meter_range ? (
                    <div className="mt-2 flex items-center justify-between gap-3">
                      <span>Stand meter</span>
                      <span className="text-right font-semibold text-slate-900">{billingInquiry.meter_range}</span>
                    </div>
                  ) : null}
                  <div className="mt-2 flex items-center justify-between gap-3">
                    <span>Nilai tagihan</span>
                    <span className="text-right font-semibold text-slate-900">{formatRupiah(billingDisplayTotalTagihan)}</span>
                  </div>
                  <div className="mt-2 flex items-center justify-between gap-3">
                    <span>{billingDetailConfig.adminLabel}</span>
                    <span className="text-right font-semibold text-slate-900">{formatRupiah(billingAdminAmount)}</span>
                  </div>
                  <div className="mt-2 flex items-center justify-between gap-3 border-t border-slate-100 pt-2">
                    <span>Total tagihan</span>
                    <span className="text-right font-bold text-slate-900">{formatRupiah(billingTotalAmount)}</span>
                  </div>
                </div>
              ) : null}
              {shouldShowSummaryPrimaryRow(product) ? (
                <div className={`${isBillingPayment && billingInquiry ? "mt-3" : ""} flex items-center justify-between gap-3`}>
                  <span>{isBillingPayment ? "Total tagihan" : getSummaryPrimaryLabel(product)}</span>
                  <span className="font-semibold text-slate-900">
                    {isBillingPayment ? formatRupiah(billingDisplayTotalTagihan) : extractRequestedNominalLabel(product, nominalValue)}
                  </span>
                </div>
              ) : null}
              {isBillingPayment ? (
                <div className="mt-2 flex items-center justify-between gap-3">
                  <span>Admin fee</span>
                  <span className="font-semibold text-slate-900">{formatRupiah(billingDisplayAdminFee)}</span>
                </div>
              ) : ewalletProduct ? (
                <>
                  <div className="mt-2 flex items-center justify-between gap-3">
                    <span>Nominal masuk ke {product.brand_nama || "e-wallet"}</span>
                    <span className="font-semibold text-slate-900">{formatRupiah(ewalletNominal)}</span>
                  </div>
                  <div className="mt-2 flex items-center justify-between gap-3">
                    <span>Biaya layanan</span>
                    <span className="font-semibold text-slate-900">{formatRupiah(ewalletServiceFee)}</span>
                  </div>
                </>
              ) : (
                <div className="mt-2 flex items-center justify-between gap-3">
                  <span>Harga</span>
                  <span className="font-semibold text-slate-900">{hargaSebelumFeeAdmin > 0 ? formatRupiah(hargaSebelumFeeAdmin) : "-"}</span>
                </div>
              )}
              {estimatedWalletDebit > 0 ? (
                <div className="mt-2 flex items-center justify-between gap-3">
                  <span>Dibayar dari saldo utama</span>
                  <span className="font-semibold text-slate-900">{formatRupiah(estimatedWalletDebit)}</span>
                </div>
              ) : null}
              {needsTopup ? (
                <div className="mt-2 flex items-center justify-between gap-3">
                  <span>Kekurangan saldo utama</span>
                  <span className="font-semibold text-amber-700">{formatRupiah(estimatedQrisAmount)}</span>
                </div>
              ) : null}
              {!isBillingPayment && feeAdminQris > 0 ? (
                <div className="mt-2 flex items-center justify-between gap-3">
                  <span>Fee admin</span>
                  <span className="font-semibold text-slate-900">{formatRupiah(feeAdminQris)}</span>
                </div>
              ) : null}
              <div className="mt-2 flex items-center justify-between gap-3">
                <span>Total bayar</span>
                <span className="font-semibold text-sky-700">{formatRupiah(totalBayar)}</span>
              </div>
            </div>

            <label className="block space-y-2 text-sm">
              <span className="font-semibold text-slate-700">{destLabel}</span>
                  {destMode === "ml_id_server" ? (
                    <div className="grid grid-cols-[minmax(0,1fr)_92px] gap-2">
                      <input
                        value={dest}
                        onChange={(e) => setDest(normalizeDestInput(e.target.value, 15))}
                        placeholder={destPlaceholder}
                        type="tel"
                        inputMode="numeric"
                        maxLength={15}
                        className="h-12 w-full rounded-2xl border border-slate-200 px-4 text-base text-slate-900 outline-none ring-0 transition focus:border-sky-300"
                      />
                      <input
                        value={destServer}
                        onChange={(e) => setDestServer(normalizeServerInput(e.target.value, 6))}
                        placeholder="Server"
                        type="tel"
                        inputMode="numeric"
                        maxLength={6}
                        className="h-12 w-full rounded-2xl border border-slate-200 px-4 text-base text-slate-900 outline-none ring-0 transition focus:border-sky-300"
                      />
                    </div>
                  ) : (
                  <input
                    value={dest}
                    onChange={(e) => setDest(destMode === "alphanumeric" ? normalizeAlphaNumericDestInput(e.target.value, 32) : normalizeDestInput(e.target.value))}
                    placeholder={destPlaceholder}
                    type={destMode === "alphanumeric" ? "text" : "tel"}
                    inputMode={destMode === "alphanumeric" ? "text" : "numeric"}
                    autoCapitalize={destMode === "alphanumeric" ? "off" : undefined}
                    autoCorrect={destMode === "alphanumeric" ? "off" : undefined}
                    spellCheck={destMode === "alphanumeric" ? false : undefined}
                    maxLength={destMode === "alphanumeric" ? 32 : 13}
                    className="h-12 w-full rounded-2xl border border-slate-200 px-4 text-base text-slate-900 outline-none ring-0 transition focus:border-sky-300"
                  />
                  )}
                </label>

            {!isFixed && !isBillingPayment ? (
              <label className="block space-y-2 text-sm">
                <span className="font-semibold text-slate-700">Nominal</span>
                  <input
                    value={nominal}
                    onChange={(e) => setNominal(e.target.value.replace(/[^\d]/g, ""))}
                    placeholder="Contoh: 10000"
                    inputMode="numeric"
                    className="h-12 w-full rounded-2xl border border-slate-200 px-4 text-base text-slate-900 outline-none ring-0 transition focus:border-sky-300"
                  />
                </label>
            ) : null}

            {!effectiveAuthToken && guestNeedsTurnstile && (waitingTurnstile || turnstileError) ? (
              <div className="space-y-3">
                <TurnstileWidget
                  key={`${turnstileAppearance}-${waitingTurnstile}-${turnstileError ? "error" : "ready"}`}
                  siteKey={TURNSTILE_SITE_KEY}
                  appearance={turnstileAppearance}
                  onToken={(token) => {
                    setGuestTurnstileToken(token);
                    setTurnstileError(null);
                    setTurnstileHint(null);
                  }}
                  onExpire={() => {
                    setGuestTurnstileToken("");
                    setTurnstileHint("Verifikasi keamanan kedaluwarsa. Klik Bayar lagi untuk melanjutkan.");
                  }}
                  onError={() => {
                    setGuestTurnstileToken("");
                    setWaitingTurnstile(false);
                    setTurnstileAppearance("always");
                    setTurnstileError("Verifikasi keamanan belum berhasil. Klik verifikasi lalu lanjutkan pembayaran.");
                  }}
                />
              </div>
            ) : null}

            {needsTopup ? (
              <div className="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
                Saldo utama tidak mencukupi. Isi saldo sebesar {formatRupiah(estimatedQrisAmount)} untuk melanjutkan {ewalletProduct ? `top up ${product.brand_nama || "e-wallet"}` : "transaksi"}.
              </div>
            ) : null}
            {error ? <div className="rounded-2xl border border-rose-100 bg-rose-50 px-4 py-3 text-sm text-rose-700">{error}</div> : null}
            {turnstileHint ? <div className="rounded-2xl border border-sky-100 bg-sky-50 px-4 py-3 text-sm text-sky-700">{turnstileHint}</div> : null}
            {turnstileError ? <div className="rounded-2xl border border-amber-100 bg-amber-50 px-4 py-3 text-sm text-amber-700">{turnstileError}</div> : null}

            {needsTopup ? (
              <button
                type="button"
                onClick={() => {
                  onClose();
                  router.push(`/user/account/topup?amount=${Math.ceil(estimatedQrisAmount)}`);
                }}
                className="inline-flex h-13 w-full items-center justify-center rounded-2xl bg-[#087e8b] px-5 text-xs font-black text-white shadow-[0_10px_20px_rgba(4,120,87,0.22)] transition hover:bg-[#036b4d]"
              >
                Isi Saldo
              </button>
            ) : (
              <button
                type="submit"
                disabled={loading}
                className="inline-flex h-13 w-full items-center justify-center rounded-2xl bg-[#13464b] px-11 text-xs font-black text-white shadow-[0_10px_20px_rgba(5,46,38,0.18)] transition group-hover:bg-[#087e8b]"
              >
                {loading ? <LoaderCircle className="h-4 w-4 animate-spin" /> : null}
                {loading ? "Memproses..." : "Bayar"}
              </button>
            )}
          </form>
        )}
      </div>
    </div>
  );
}
