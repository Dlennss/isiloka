"use client";

import { useEffect, useMemo, useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { getProviders, getSession, signIn } from "next-auth/react";
import { ArrowRight, Eye, EyeOff, LockKeyhole, Mail, Phone, ShieldCheck, UserPlus, UserRound, Zap } from "lucide-react";

const GOOGLE_LOGIN_ENABLED = String(process.env.NEXT_PUBLIC_GOOGLE_LOGIN_ENABLED ?? "false").toLowerCase() === "true";

type RegisterResp = {
  ok?: boolean;
  member_id?: number;
  role?: string;
  auto_refund_claimed?: boolean;
  auto_refund_amount?: number;
  refund_invoice_id?: string;
  error?: string;
};

function cn(...v: Array<string | false | null | undefined>) {
  return v.filter(Boolean).join(" ");
}

function normalizeGoogleNext(raw: string, fallback = "/user") {
  let current = (raw || "").trim() || fallback;
  for (let i = 0; i < 6; i += 1) {
    try {
      const url = current.startsWith("http://") || current.startsWith("https://")
        ? new URL(current)
        : new URL(current, "https://pulsakilat.local");
      if (url.pathname === "/login") {
        const nested = (url.searchParams.get("callbackUrl") || "").trim();
        if (nested) { current = nested; continue; }
      }
      if (url.pathname === "/auth/google/complete") {
        const nested = (url.searchParams.get("next") || "").trim();
        if (nested) { current = nested; continue; }
      }
      return `${url.pathname}${url.search}${url.hash}` || fallback;
    } catch {
      return current.startsWith("/") ? current : fallback;
    }
  }
  return current.startsWith("/") ? current : fallback;
}

function cleanPhoneInput(value: string) {
  return value.replace(/[^\d+]/g, "").slice(0, 16);
}

function normalizePhone(value: string) {
  const cleaned = cleanPhoneInput(value.trim());
  if (cleaned.startsWith("+62")) return `0${cleaned.slice(3)}`;
  if (cleaned.startsWith("62")) return `0${cleaned.slice(2)}`;
  return cleaned;
}

export function RegisterCard() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [nama, setNama] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState("");
  const [shake, setShake] = useState(false);
  const [googleAvailable, setGoogleAvailable] = useState(false);
  const refundContext = useMemo(() => ({
    invoiceId: (searchParams.get("refund_invoice_id") || "").trim(),
    guestEmail: (searchParams.get("guest_email") || "").trim(),
    guestPhone: (searchParams.get("guest_phone") || "").trim(),
  }), [searchParams]);
  const hasRefundContext = Boolean(refundContext.invoiceId && refundContext.guestEmail && refundContext.guestPhone);
  const loginCallbackUrl = useMemo(() => {
    return normalizeGoogleNext((searchParams.get("callbackUrl") || "").trim(), "/user");
  }, [searchParams]);
  const googleCallbackUrl = useMemo(() => {
    const params = new URLSearchParams({ next: loginCallbackUrl });
    if (hasRefundContext) {
      params.set("refund_invoice_id", refundContext.invoiceId);
      params.set("guest_email", refundContext.guestEmail);
      params.set("guest_phone", refundContext.guestPhone);
    }
    return `/auth/google/complete?${params.toString()}`;
  }, [hasRefundContext, loginCallbackUrl, refundContext.guestEmail, refundContext.guestPhone, refundContext.invoiceId]);
  const canUseGoogleLogin = GOOGLE_LOGIN_ENABLED && googleAvailable && !loading;

  useEffect(() => {
    if (!GOOGLE_LOGIN_ENABLED) return;
    void (async () => {
      const providers = await getProviders().catch(() => null);
      setGoogleAvailable(Boolean(providers?.google));
    })();
  }, []);

  async function claimRefundWithSession() {
    const sess = (await getSession().catch(() => null)) as { backendToken?: string } | null;
    const backendToken = String(sess?.backendToken || "").trim();
    if (!backendToken || !hasRefundContext) return false;
    const res = await fetch("/api/app/me/refunds/claim", {
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${backendToken}` },
      body: JSON.stringify({
        invoice_id: refundContext.invoiceId,
        guest_email: refundContext.guestEmail,
        guest_phone: refundContext.guestPhone,
      }),
    });
    return res.ok;
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setErr("");
    const cleanNama = nama.trim();
    const cleanEmail = email.trim().toLowerCase();
    const cleanPhone = normalizePhone(phone);
    const cleanPassword = password.trim();
    const cleanConfirm = confirmPassword.trim();
    if (!cleanNama || !cleanEmail || !cleanPhone || !cleanPassword || !cleanConfirm) { setErr("Lengkapi nama, email, nomor telepon, dan password."); return; }
    if (!/^08\d{8,12}$/.test(cleanPhone)) { setErr("Nomor telepon gunakan format 08 dan 10-14 digit."); return; }
    if (cleanPassword.length < 8) { setErr("Password minimal 8 karakter."); return; }
    if (cleanPassword !== cleanConfirm) { setErr("Konfirmasi password tidak sama."); return; }
    setLoading(true);
    try {
      const r = await fetch("/api/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          nama: cleanNama, email: cleanEmail, phone: cleanPhone, password: cleanPassword,
          refund_invoice_id: refundContext.invoiceId,
          guest_email: refundContext.guestEmail,
          guest_phone: refundContext.guestPhone,
        }),
      });
      const j = (await r.json().catch(() => ({}))) as RegisterResp;
      if (!r.ok || !j?.ok) { setErr(j?.error || "Registrasi gagal"); return; }
      if (hasRefundContext) {
        const loginResult = await signIn("credentials", { redirect: false, email: cleanEmail, password: cleanPassword });
        if (!loginResult || loginResult.error) { router.replace(`/login?registered=1`); return; }
        const refundClaimed = j?.auto_refund_claimed || (await claimRefundWithSession());
        router.replace(refundClaimed ? "/user?registered=1&refund=1" : "/user?registered=1");
        return;
      }
      router.replace(`/login?registered=1`);
    } finally {
      setLoading(false);
    }
  }

  if (err && !shake) setTimeout(() => setShake(true), 0);
  if (!err && shake) setTimeout(() => setShake(false), 0);

  const inputClass = "h-[54px] w-full rounded-2xl border border-slate-200 bg-white px-[52px] text-sm font-semibold text-slate-900 outline-none shadow-[0_10px_24px_rgba(15,23,42,0.05)] transition placeholder:text-slate-400 focus:border-[#10b981] focus:ring-4 focus:ring-emerald-100";
  const inputWithToggle = "h-[54px] w-full rounded-2xl border border-slate-200 bg-white px-[52px] pr-[52px] text-sm font-semibold text-slate-900 outline-none shadow-[0_10px_24px_rgba(15,23,42,0.05)] transition placeholder:text-slate-400 focus:border-[#10b981] focus:ring-4 focus:ring-emerald-100";
  const inputIcon = "pointer-events-none absolute left-5 top-1/2 h-5 w-5 -translate-y-1/2 text-emerald-700/55";
  const toggleBtn = "absolute right-4 top-1/2 -translate-y-1/2 rounded-lg p-1.5 text-slate-400 transition-colors hover:text-slate-600";

  return (
    <section className={cn(
      "auth-card auth-jiggle overflow-hidden rounded-[28px] bg-white/95 shadow-[0_24px_64px_rgba(5,46,38,0.32)]",
      shake && "auth-shake"
    )}>
      <div className="relative overflow-hidden bg-[linear-gradient(135deg,#052e26_0%,#047857_60%,#8ee82d_150%)] px-6 pb-6 pt-7 text-white">
        <div className="pointer-events-none absolute -right-12 -top-12 h-36 w-36 rounded-full bg-white/10" />
        <div className="pointer-events-none absolute left-6 top-6 h-20 w-20 rounded-full border border-lime-200/20" />
        <div className="relative flex items-start justify-between gap-4">
          <div className="min-w-0">
            <Image
              src="/images/logo-pulsakilat-header.svg"
              alt="Isiloka"
              width={250}
              height={58}
              className="h-14 w-auto max-w-[210px] rounded-2xl bg-white/95 px-3 py-2 object-contain shadow-[0_14px_26px_rgba(6,78,59,0.18)]"
              priority
            />
            <h1 className="mt-5 text-2xl font-black text-white">Buat Akun</h1>
            <p className="mt-1.5 max-w-[270px] text-sm font-medium leading-6 text-emerald-50/80">
              Daftar cepat untuk pulsa, data, game, dan e-wallet.
            </p>
          </div>
          <span className="mt-1 grid h-12 w-12 shrink-0 place-items-center rounded-2xl bg-white/15 ring-1 ring-white/20">
            <Zap className="h-6 w-6 fill-lime-300 text-lime-300" />
          </span>
        </div>
      </div>

      <div className="px-6 pb-6 pt-6">
        {err && (
          <div className="mb-5 flex items-start gap-2.5 rounded-xl border border-rose-200 bg-rose-50 px-4 py-3">
            <div className="mt-0.5 h-1.5 w-1.5 shrink-0 rounded-full bg-rose-400" />
            <p className="text-[13px] font-medium text-rose-700">{err}</p>
          </div>
        )}

        {hasRefundContext && (
          <div className="mb-5 flex items-start gap-2.5 rounded-xl border border-sky-200 bg-sky-50 px-4 py-3">
            <div className="mt-0.5 h-1.5 w-1.5 shrink-0 rounded-full bg-sky-400" />
            <p className="text-[13px] font-medium text-sky-700">
              Transaksi gagal dengan invoice <span className="font-bold">{refundContext.invoiceId}</span>. Daftar agar saldo tidak hilang.
            </p>
          </div>
        )}

        <form onSubmit={onSubmit} className="space-y-4">
          <div>
            <label className="mb-3 block text-sm font-bold text-slate-800">Nama Lengkap</label>
            <div className="relative">
              <UserRound className={inputIcon} />
              <input className={inputClass} value={nama} onChange={(e) => setNama(e.target.value)} placeholder="Nama anda" autoComplete="name" />
            </div>
          </div>

          <div>
            <label className="mb-3 block text-sm font-bold text-slate-800">Email</label>
            <div className="relative">
              <Mail className={inputIcon} />
              <input className={inputClass} value={email} onChange={(e) => setEmail(e.target.value)} placeholder="nama@email.com" autoComplete="email" type="email" />
            </div>
          </div>

          <div>
            <label className="mb-3 block text-sm font-bold text-slate-800">Nomor Telepon</label>
            <div className="relative">
              <Phone className={inputIcon} />
              <input
                className={inputClass}
                value={phone}
                onChange={(e) => setPhone(cleanPhoneInput(e.target.value))}
                placeholder="08xxxxxxxxxx"
                autoComplete="tel"
                inputMode="tel"
                type="tel"
              />
            </div>
          </div>

          <div>
            <label className="mb-3 block text-sm font-bold text-slate-800">Password</label>
            <div className="relative">
              <LockKeyhole className={inputIcon} />
              <input className={inputWithToggle} type={showPassword ? "text" : "password"} value={password} onChange={(e) => setPassword(e.target.value)} placeholder="Minimal 8 karakter" autoComplete="new-password" />
              <button type="button" onClick={() => setShowPassword((v) => !v)} className={toggleBtn} tabIndex={-1}>
                {showPassword ? <EyeOff className="h-[18px] w-[18px]" /> : <Eye className="h-[18px] w-[18px]" />}
              </button>
            </div>
          </div>

          <div>
            <label className="mb-3 block text-sm font-bold text-slate-800">Konfirmasi Password</label>
            <div className="relative">
              <ShieldCheck className={inputIcon} />
              <input className={inputWithToggle} type={showConfirmPassword ? "text" : "password"} value={confirmPassword} onChange={(e) => setConfirmPassword(e.target.value)} placeholder="Ulangi password" autoComplete="new-password" />
              <button type="button" onClick={() => setShowConfirmPassword((v) => !v)} className={toggleBtn} tabIndex={-1}>
                {showConfirmPassword ? <EyeOff className="h-[18px] w-[18px]" /> : <Eye className="h-[18px] w-[18px]" />}
              </button>
            </div>
          </div>

          <button
            className="group relative mt-2 flex h-[58px] w-full items-center justify-center gap-3 rounded-2xl bg-linear-to-r from-[#009944] via-[#16b934] to-[#57d735] text-base font-black text-white shadow-[0_14px_30px_rgba(22,185,52,0.34)] transition-all hover:shadow-[0_18px_38px_rgba(22,185,52,0.44)] active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-60 disabled:shadow-none"
            disabled={loading}
            type="submit"
          >
            {loading ? (
              <>
                <div className="h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white" />
                <span>Memproses...</span>
              </>
            ) : (
              <>
                <UserPlus className="h-5 w-5" />
                <span>Daftar Sekarang</span>
                <ArrowRight className="h-4 w-4 opacity-0 transition-all group-hover:translate-x-0.5 group-hover:opacity-100" />
              </>
            )}
          </button>

          {GOOGLE_LOGIN_ENABLED && (
            <>
              <div className="flex items-center gap-4">
                <div className="h-px flex-1 bg-slate-200" />
                <span className="text-sm font-semibold text-slate-400">atau</span>
                <div className="h-px flex-1 bg-slate-200" />
              </div>
              <button
                type="button"
                className="flex h-[54px] w-full items-center justify-center gap-3 rounded-2xl border border-slate-200 bg-white text-sm font-bold text-slate-700 shadow-[0_8px_20px_rgba(15,23,42,0.05)] transition-all hover:border-slate-300 hover:bg-slate-50 hover:shadow-sm active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-60"
                disabled={!canUseGoogleLogin}
                onClick={() => {
                  if (!canUseGoogleLogin) {
                    setErr("Login Google belum dikonfigurasi. Isi GOOGLE_CLIENT_ID dan GOOGLE_CLIENT_SECRET.");
                    return;
                  }
                  void signIn("google", { callbackUrl: googleCallbackUrl });
                }}
              >
                <Image src="/google.svg" alt="" width={18} height={18} aria-hidden="true" />
                {googleAvailable ? "Daftar dengan Google" : "Google belum dikonfigurasi"}
              </button>
            </>
          )}

          <div className="space-y-2 pt-1 text-center">
            <p className="text-sm font-medium text-slate-500">
              Sudah punya akun?{" "}
              <Link href="/login" className="font-black text-emerald-700 hover:underline" style={{ color: "#087e8b" }}>
                Masuk
              </Link>
            </p>
          </div>
        </form>
      </div>
    </section>
  );
}
