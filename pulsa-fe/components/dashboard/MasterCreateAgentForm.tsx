"use client";

import { type FormEvent, type ReactNode, useState } from "react";
import { useRouter } from "next/navigation";
import {
  BadgeCheck,
  Copy,
  KeyRound,
  Loader2,
  Mail,
  Phone,
  RefreshCcw,
  Store,
  UserPlus,
  UserRound,
  type LucideIcon,
} from "lucide-react";

type CreateResp = {
  ok?: boolean;
  error?: string;
  member_id?: number;
};

function authHeader(): Record<string, string> {
  if (typeof window === "undefined") return {};
  const token = window.localStorage.getItem("auth_token") || "";
  return token ? { Authorization: `Bearer ${token}` } : {};
}

function makePassword() {
  const seed = Math.random().toString(36).slice(2, 8).toUpperCase();
  return `Kilat${seed}24`;
}

function Field({
  label,
  icon: Icon,
  children,
  helper,
}: {
  label: string;
  icon: LucideIcon;
  children: ReactNode;
  helper?: string;
}) {
  return (
    <label className="block min-w-0">
      <span className="mb-2 flex items-center gap-2 text-[10px] font-black uppercase tracking-[0.12em] text-slate-600">
        <Icon className="h-4 w-4 shrink-0 text-emerald-600" strokeWidth={2.2} />
        {label}
      </span>
      {children}
      {helper ? <span className="mt-2 block text-[11px] font-medium leading-[1.45] text-slate-500">{helper}</span> : null}
    </label>
  );
}

export function MasterCreateAgentForm({ useRetailEndpoint = false }: { useRetailEndpoint?: boolean }) {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [storeName, setStoreName] = useState("");
  const [password, setPassword] = useState(makePassword);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  const canSubmit = email.trim() && name.trim() && password.trim().length >= 8 && !loading;

  async function copyPassword() {
    await navigator.clipboard?.writeText(password).catch(() => undefined);
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!canSubmit) {
      setMessage({ type: "error", text: "Nama, email, dan password minimal 8 karakter wajib diisi." });
      return;
    }

    setLoading(true);
    setMessage(null);
    try {
      const response = await fetch(useRetailEndpoint ? "/api/me/retail/downlines" : "/api/admin/members/create", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({
          email: email.trim(),
          nama: name.trim(),
          phone: phone.trim(),
          store_name: storeName.trim(),
          password: password.trim(),
          role: "agent",
          retail_agent_commission_rp: 0,
        }),
      });
      const body = (await response.json().catch(() => ({}))) as CreateResp;
      if (!response.ok || !body.ok || !body.member_id) {
        throw new Error(body.error || "Akun agent gagal dibuat");
      }

      setMessage({ type: "success", text: `Agent berhasil dibuat. ID Agent: ${body.member_id}` });
      setEmail("");
      setName("");
      setPhone("");
      setStoreName("");
      setPassword(makePassword());
      router.refresh();
    } catch (err) {
      setMessage({ type: "error", text: err instanceof Error ? err.message : "Akun agent gagal dibuat" });
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={submit} className="w-full min-w-0">
      <section className="w-full min-w-0 overflow-hidden rounded-[24px] border border-emerald-100 bg-white p-5 shadow-[0_18px_42px_rgba(6,78,59,0.08)] sm:p-6">
        <div className="flex items-start justify-between gap-4 border-b border-slate-100 pb-5">
          <div className="min-w-0">
            <p className="text-[10px] font-black uppercase tracking-[0.18em] text-emerald-700">Form Agent</p>
            <h2 className="mt-1 text-2xl font-black leading-[1.15] tracking-normal text-slate-950 sm:text-[28px]">Identitas Agent Baru</h2>
            <p className="mt-2 max-w-xl text-xs font-semibold leading-5 text-slate-500 sm:text-sm">Akun baru langsung aktif sebagai agent Isiloka.</p>
          </div>
          <span className="grid h-12 w-12 shrink-0 place-items-center rounded-2xl bg-emerald-50 text-emerald-700 ring-1 ring-inset ring-emerald-200 sm:h-14 sm:w-14">
            <UserPlus className="h-6 w-6" strokeWidth={2.2} />
          </span>
        </div>

        <div className="mt-5 grid min-w-0 grid-cols-[repeat(auto-fit,minmax(min(100%,280px),1fr))] gap-x-4 gap-y-5">
          <Field label="Nama Agent" icon={UserRound}>
            <input
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder="Contoh: Agent Isiloka 1"
              autoComplete="name"
              className="h-12 w-full rounded-2xl border border-slate-200 bg-slate-50 px-4 text-sm font-bold text-slate-950 outline-none transition placeholder:font-semibold placeholder:text-slate-400 focus:border-emerald-500 focus:bg-white focus:ring-4 focus:ring-emerald-100"
            />
          </Field>
          <Field label="Email Login" icon={Mail}>
            <input
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              placeholder="agent@pulsakilat.local"
              type="email"
              autoComplete="email"
              className="h-12 w-full rounded-2xl border border-slate-200 bg-slate-50 px-4 text-sm font-bold text-slate-950 outline-none transition placeholder:font-semibold placeholder:text-slate-400 focus:border-emerald-500 focus:bg-white focus:ring-4 focus:ring-emerald-100"
            />
          </Field>
          <Field label="Nomor WA" icon={Phone} helper="Nomor yang dapat dihubungi oleh marketing dan operator.">
            <input
              value={phone}
              onChange={(event) => setPhone(event.target.value)}
              placeholder="08xxxxxxxxxx"
              inputMode="tel"
              autoComplete="tel"
              className="h-12 w-full rounded-2xl border border-slate-200 bg-slate-50 px-4 text-sm font-bold text-slate-950 outline-none transition placeholder:font-semibold placeholder:text-slate-400 focus:border-emerald-500 focus:bg-white focus:ring-4 focus:ring-emerald-100"
            />
          </Field>
          <Field label="Nama Toko" icon={Store} helper="Nama konter akan tersimpan pada profil agent.">
            <input
              value={storeName}
              onChange={(event) => setStoreName(event.target.value)}
              placeholder="Nama konter/toko"
              autoComplete="organization"
              className="h-12 w-full rounded-2xl border border-slate-200 bg-slate-50 px-4 text-sm font-bold text-slate-950 outline-none transition placeholder:font-semibold placeholder:text-slate-400 focus:border-emerald-500 focus:bg-white focus:ring-4 focus:ring-emerald-100"
            />
          </Field>
          <div className="min-w-0 [grid-column:1/-1]">
            <Field label="Password Awal" icon={KeyRound} helper="Berikan password ini ke agent, lalu sarankan diganti setelah login.">
              <div className="flex min-w-0 gap-2">
                <input
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  autoComplete="new-password"
                  className="h-12 min-w-0 flex-1 rounded-2xl border border-slate-200 bg-slate-50 px-3 text-sm font-bold text-slate-950 outline-none transition focus:border-emerald-500 focus:bg-white focus:ring-4 focus:ring-emerald-100 sm:px-4"
                />
                <button
                  type="button"
                  onClick={() => setPassword(makePassword())}
                  className="grid h-12 w-12 shrink-0 place-items-center rounded-2xl bg-emerald-50 text-emerald-700 ring-1 ring-inset ring-emerald-200 transition hover:bg-emerald-100"
                  aria-label="Buat password baru"
                  title="Buat password baru"
                >
                  <RefreshCcw className="h-4 w-4" />
                </button>
                <button
                  type="button"
                  onClick={copyPassword}
                  className="grid h-12 w-12 shrink-0 place-items-center rounded-2xl bg-slate-50 text-slate-600 ring-1 ring-inset ring-slate-200 transition hover:bg-slate-100"
                  aria-label="Salin password"
                  title="Salin password"
                >
                  <Copy className="h-4 w-4" />
                </button>
              </div>
            </Field>
          </div>
        </div>

        {message ? (
          <div className={message.type === "success" ? "mt-5 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm font-black text-emerald-700" : "mt-5 rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm font-black text-rose-600"}>
            {message.text}
          </div>
        ) : null}

        <button
          type="submit"
          disabled={!canSubmit}
          className="mt-5 flex h-12 w-full items-center justify-center gap-2 rounded-2xl bg-emerald-700 px-4 text-sm font-black text-white shadow-[0_14px_28px_rgba(4,120,87,0.22)] transition hover:bg-emerald-800 disabled:cursor-not-allowed disabled:bg-slate-300 disabled:shadow-none"
        >
          {loading ? <Loader2 className="h-5 w-5 animate-spin" /> : <BadgeCheck className="h-5 w-5" />}
          {loading ? "Membuat Agent..." : "Buat Akun Agent"}
        </button>
      </section>

    </form>
  );
}
