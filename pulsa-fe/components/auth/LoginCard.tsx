"use client";

import { useEffect, useState, type FormEvent } from "react";
import Image from "next/image";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { getProviders, signIn } from "next-auth/react";
import { Eye, EyeOff, LockKeyhole, UserRound, Zap } from "lucide-react";
import TurnstileWidget from "@/components/TurnstileWidget";
import { decodeJwt } from "@/lib/jwt";

const SITE_KEY = process.env.NEXT_PUBLIC_TURNSTILE_SITE_KEY || "";
const TURNSTILE_ENABLED = /^(1|true|yes|on)$/i.test(process.env.NEXT_PUBLIC_TURNSTILE_ENABLED || "");
// Ketersediaan akhir tetap ditentukan provider dari server (`getProviders`).
// Default aktif mencegah build lama mengunci tombol ketika konfigurasi server baru dipasang.
const GOOGLE_LOGIN_ENABLED = String(process.env.NEXT_PUBLIC_GOOGLE_LOGIN_ENABLED ?? "true").toLowerCase() === "true";

type PasswordLoginResp = { ok?: boolean; token?: string; role?: string; error?: string };

function cn(...v: Array<string | false | null | undefined>) {
  return v.filter(Boolean).join(" ");
}

async function persistLoginToken(token: string) {
  const response = await fetch("/api/auth/persist", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ token }),
    cache: "no-store",
  }).catch(() => null);

  return Boolean(response?.ok);
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

function toDashboardByRole(role?: string | null) {
  const r = (role || "").toLowerCase();
  if (r === "admin" || r === "staff") return "/dashboard/admin";
  if (r === "auditor") return "/dashboard/auditor";
  if (r === "analis" || r === "analyst") return "/dashboard/master/operator";
  if (r === "master") return "/dashboard/master";
  if (r === "user" || r === "member" || r === "agent" || r === "agent_member" || r === "master_member") return "/user";
  if (r === "operator_trx") return "/dashboard/operator";
  if (r === "operator_wallet") return "/dashboard/wallet";
  return "/user";
}

export function LoginCard() {
  const searchParams = useSearchParams();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [remember, setRemember] = useState(true);
  const [turnstileToken, setTurnstileToken] = useState("");
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState("");
  const [googleAvailable, setGoogleAvailable] = useState(false);
  const [shake, setShake] = useState(false);
  const loginCallbackUrl = normalizeGoogleNext((searchParams.get("callbackUrl") || "").trim(), "/user");
  const googleCallbackUrl = `/auth/google/complete?${new URLSearchParams({ next: loginCallbackUrl }).toString()}`;
  const canUseGoogleLogin = GOOGLE_LOGIN_ENABLED && googleAvailable && !loading;

  useEffect(() => {
    setTurnstileToken(TURNSTILE_ENABLED ? "" : "dev-bypass");
  }, []);

  useEffect(() => {
    if (!GOOGLE_LOGIN_ENABLED) return;
    void (async () => {
      const providers = await getProviders().catch(() => null);
      setGoogleAvailable(Boolean(providers?.google));
    })();
  }, []);

  useEffect(() => {
    if (!err) return;
    setShake(true);
    const t = setTimeout(() => setShake(false), 520);
    return () => clearTimeout(t);
  }, [err]);

  useEffect(() => {
    const googleStatus = (searchParams.get("google") || "").trim();
    const authError = (searchParams.get("error") || "").trim();
    if (googleStatus === "not_configured") {
      setErr("Login Google belum tersambung. Isi Client ID dan Client Secret Google terlebih dulu.");
      return;
    }
    if (authError) {
      setErr("Login Google gagal atau dibatalkan. Coba masuk ulang.");
    }
  }, [searchParams]);

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setErr("");
    if (TURNSTILE_ENABLED && !turnstileToken) {
      setErr("Selesaikan verifikasi keamanan.");
      return;
    }
    setLoading(true);
    try {
      const loginResponse = await fetch("/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ email, password, turnstileToken }),
        cache: "no-store",
      });
      const loginBody = (await loginResponse.json().catch(() => ({}))) as PasswordLoginResp;
      const backendToken = String(loginBody.token || "").trim();
      if (!loginResponse.ok || !loginBody.ok || !backendToken) {
        setErr("Email atau password salah.");
        return;
      }

      if (!(await persistLoginToken(backendToken))) {
        setErr("Sesi login belum tersimpan. Silakan coba lagi.");
        return;
      }

      localStorage.setItem("auth_token", backendToken);
      localStorage.setItem("auth_source", "password");

      // `/api/auth/login` sudah membuat cookie HttpOnly dan mengembalikan token
      // backend. Jangan jalankan login NextAuth kedua karena dua perubahan sesi
      // bersamaan memicu kedipan/navigasi ganda terutama di Safari mobile.
      window.location.replace(toDashboardByRole(loginBody.role || decodeJwt(backendToken)?.role || "member"));
    } catch {
      setErr("Koneksi login gagal. Periksa jaringan lalu coba lagi.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <section className={cn(
      "auth-card overflow-hidden rounded-[28px] bg-white/95 shadow-[0_24px_64px_rgba(5,46,38,0.32)]",
      shake && "auth-shake"
    )}>
      <div className="px-7 py-8">
        <div className="mb-6 text-center">
          <Image
            src="/images/logo-pulsakilat-header.svg"
            alt="Isiloka"
            width={340}
            height={78}
            className="mx-auto h-[64px] w-auto max-w-full object-contain"
            priority
          />
          <h1 className="mt-4 text-[22px] font-black tracking-tight text-slate-900">
            Masuk Akun
          </h1>
          <p className="mt-2 text-sm font-medium leading-6 text-slate-500">
            Lanjutkan ke Isiloka.
          </p>
        </div>

        {err && (
          <div className="mb-5 flex items-start gap-2.5 rounded-xl border border-rose-200 bg-rose-50 px-4 py-3">
            <div className="mt-0.5 h-1.5 w-1.5 shrink-0 rounded-full bg-rose-400" />
            <p className="text-[13px] font-medium text-rose-700">{err}</p>
          </div>
        )}

        <form onSubmit={onSubmit} className="space-y-4">
          <div>
            <label className="mb-3 block text-sm font-bold text-slate-800">
              Email / No HP
            </label>
            <div className="relative">
              <UserRound className="pointer-events-none absolute left-5 top-1/2 h-5 w-5 -translate-y-1/2 text-slate-400" />
              <input
                className="h-[54px] w-full rounded-2xl border border-slate-200 bg-white px-[52px] text-sm font-medium text-slate-900 outline-none shadow-[0_8px_20px_rgba(15,23,42,0.04)] transition placeholder:text-slate-400 focus:border-[#10b981] focus:ring-4 focus:ring-emerald-100"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="Email / No HP"
                autoComplete="username"
                type="text"
              />
            </div>
          </div>

          <div>
            <div className="mb-3 flex items-center justify-between gap-4">
              <label className="block text-sm font-bold text-slate-800">Password</label>
              <Link
                href="#"
                target="_blank"
                rel="noreferrer"
                className="text-xs font-black text-emerald-600 hover:underline"
                style={{ color: "#059669" }}
              >
                Lupa?
              </Link>
            </div>
            <div className="relative">
              <LockKeyhole className="pointer-events-none absolute left-5 top-1/2 h-5 w-5 -translate-y-1/2 text-slate-400" />
              <input
                className="h-[54px] w-full rounded-2xl border border-slate-200 bg-white px-[52px] pr-[52px] text-sm font-medium text-slate-900 outline-none shadow-[0_8px_20px_rgba(15,23,42,0.04)] transition placeholder:text-slate-400 focus:border-[#10b981] focus:ring-4 focus:ring-emerald-100"
                type={showPassword ? "text" : "password"}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Masukkan password"
                autoComplete="current-password"
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-4 top-1/2 -translate-y-1/2 rounded-lg p-1.5 text-slate-400 transition-colors hover:text-slate-600"
                tabIndex={-1}
                aria-label={showPassword ? "Sembunyikan password" : "Tampilkan password"}
              >
                {showPassword ? <EyeOff className="h-5 w-5" /> : <Eye className="h-5 w-5" />}
              </button>
            </div>
          </div>

          <label className="flex items-center gap-3 text-sm font-semibold text-slate-700">
            <input
              type="checkbox"
              checked={remember}
              onChange={(event) => setRemember(event.target.checked)}
              className="h-5 w-5 rounded border-slate-300 accent-[#087e8b]"
            />
            Ingat
          </label>

          {/* Turnstile - interaction-only, tidak tampil kalau sudah verified */}
          {TURNSTILE_ENABLED && !turnstileToken && (
            <div className="overflow-hidden rounded-xl">
              <TurnstileWidget
                siteKey={SITE_KEY}
                onToken={setTurnstileToken}
                onExpire={() => setTurnstileToken("")}
                onError={() => setTurnstileToken("")}
                appearance="interaction-only"
              />
            </div>
          )}

          {/* Login Button */}
          <button
            className="group relative flex h-[58px] w-full items-center justify-center gap-3 rounded-2xl bg-linear-to-r from-[#009944] via-[#16b934] to-[#57d735] text-lg font-black uppercase tracking-wide text-white shadow-[0_14px_30px_rgba(22,185,52,0.34)] transition-all hover:shadow-[0_18px_38px_rgba(22,185,52,0.44)] active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-60 disabled:shadow-none"
            disabled={(TURNSTILE_ENABLED && !turnstileToken) || loading}
            type="submit"
          >
            {loading ? (
              <>
                <div className="h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white" />
                <span>Memproses...</span>
              </>
            ) : (
              <>
                <Zap className="h-6 w-6 fill-white" />
                <span>Masuk</span>
              </>
            )}
          </button>

          {/* Divider */}
          <div className="flex items-center gap-4 pt-1">
            <div className="h-px flex-1 bg-slate-200" />
            <span className="text-sm font-semibold text-slate-400">atau</span>
            <div className="h-px flex-1 bg-slate-200" />
          </div>

          {/* Google Button */}
          <button
            type="button"
            className="flex h-[54px] w-full items-center justify-center gap-3 rounded-2xl border border-slate-200 bg-white text-sm font-bold text-slate-700 shadow-[0_8px_20px_rgba(15,23,42,0.05)] transition-all hover:border-slate-300 hover:bg-slate-50 hover:shadow-sm active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-60"
            disabled={loading || !canUseGoogleLogin}
            onClick={() => {
              if (!canUseGoogleLogin) {
                setErr(GOOGLE_LOGIN_ENABLED ? "Login Google belum dikonfigurasi." : "Login Google belum aktif.");
                return;
              }
              void signIn("google", { callbackUrl: googleCallbackUrl });
            }}
          >
            <Image src="/google.svg" alt="" width={22} height={22} aria-hidden="true" />
            {!GOOGLE_LOGIN_ENABLED ? "Google belum aktif" : googleAvailable ? "Masuk dengan Google" : "Google belum dikonfigurasi"}
          </button>

          {/* Footer */}
          <div className="space-y-2 pt-1 text-center">
            <p className="text-sm font-medium text-slate-500">
              Belum punya akun?{" "}
              <Link href="/register" className="font-black text-emerald-700 hover:underline" style={{ color: "#087e8b" }}>
                Daftar
              </Link>
            </p>
          </div>
        </form>
      </div>
    </section>
  );
}
