"use client";

import { type FormEvent, type ReactNode, useEffect, useMemo, useState } from "react";
import {
  BadgeCheck,
  Copy,
  Edit3,
  KeyRound,
  Loader2,
  Mail,
  Phone,
  RefreshCcw,
  Search,
  UserRound,
  X,
} from "lucide-react";
import type { AgentCreditApplication } from "@/lib/api.auth";

type AgentRow = {
  id: number;
  email: string;
  nama: string;
  phone?: string;
  role: string;
  aktif: boolean;
  saldo?: number;
  retail_agent_commission_rp?: number;
  retail_master_commission_rp?: number;
  fee_member_rp?: number;
  h2h_agent_commission_rp?: number;
  h2h_master_commission_rp?: number;
};

type MembersResp = {
  ok?: boolean;
  rows?: AgentRow[];
  items?: AgentRow[];
  total_count?: number;
  error?: string;
};

type ApiResp = {
  ok?: boolean;
  error?: string;
};

type AgentFormState = {
  id?: number;
  name: string;
  email: string;
  phone: string;
  password: string;
  commission: string;
  aktif: boolean;
};

type PortfolioFilter = "all" | "survey" | "operator" | "active" | "paid";

type Props = {
  applications: AgentCreditApplication[];
};

const emptyForm = (): AgentFormState => ({
  name: "",
  email: "",
  phone: "",
  password: "",
  commission: "0",
  aktif: true,
});

const inputClassName =
  "h-11 w-full rounded-2xl border border-slate-200 bg-white px-4 text-sm font-bold text-slate-950 outline-none transition placeholder:text-slate-400 focus:border-emerald-400 focus:ring-4 focus:ring-emerald-100";

function authHeader(): Record<string, string> {
  if (typeof window === "undefined") return {};
  const token = window.localStorage.getItem("auth_token") || "";
  return token ? { Authorization: `Bearer ${token}` } : {};
}

function makePassword() {
  const seed = Math.random().toString(36).slice(2, 8).toUpperCase();
  return `Kilat${seed}24`;
}

function formatIDR(value: number) {
  return `Rp ${new Intl.NumberFormat("id-ID").format(Number(value || 0))}`;
}

function applicationTime(item: AgentCreditApplication) {
  return new Date(item.updated_at || item.created_at).getTime();
}

function applicationStage(item?: AgentCreditApplication): Exclude<PortfolioFilter, "all"> | "none" | "rejected" {
  if (!item) return "none";
  const status = String(item.status || "").toLowerCase();
  const loanStatus = String(item.loan_status || "").toLowerCase();
  if (status.includes("reject")) return "rejected";
  if (status === "submitted" || status === "marketing_review") return "survey";
  if (status === "analysis_review" || status === "master_review") return "operator";
  if (status === "approved" && loanStatus === "paid") return "paid";
  if (status === "approved") return "active";
  return "none";
}

const stageMeta = {
  none: { label: "Belum mengajukan", className: "bg-slate-100 text-slate-600" },
  survey: { label: "Perlu survei", className: "bg-amber-100 text-amber-700" },
  operator: { label: "Di operator", className: "bg-sky-100 text-sky-700" },
  active: { label: "Kredit aktif", className: "bg-emerald-100 text-emerald-700" },
  paid: { label: "Siklus selesai", className: "bg-lime-100 text-lime-800" },
  rejected: { label: "Ditolak", className: "bg-rose-100 text-rose-700" },
} as const;

export function MasterAgentAccountsPanel({ applications }: Props) {
  const [items, setItems] = useState<AgentRow[]>([]);
  const [query, setQuery] = useState("");
  const [portfolioFilter, setPortfolioFilter] = useState<PortfolioFilter>("all");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);
  const [form, setForm] = useState<AgentFormState>(() => emptyForm());

  const latestApplicationByMember = useMemo(() => {
    const latest = new Map<number, AgentCreditApplication>();
    applications.forEach((application) => {
      const memberId = Number(application.member_id || 0);
      const current = latest.get(memberId);
      if (memberId && (!current || applicationTime(application) > applicationTime(current))) latest.set(memberId, application);
    });
    return latest;
  }, [applications]);

  const filteredItems = useMemo(() => {
    const q = query.trim().toLowerCase();
    return items.filter((item) => {
      const application = latestApplicationByMember.get(item.id);
      const stage = applicationStage(application);
      const matchesFilter = portfolioFilter === "all" || stage === portfolioFilter;
      const storeName = String(application?.store_name || application?.applicant_data?.store_name || "");
      const matchesQuery = !q || [item.nama, item.email, item.phone, storeName, String(item.id)].some((value) => String(value || "").toLowerCase().includes(q));
      return matchesFilter && matchesQuery;
    });
  }, [items, latestApplicationByMember, portfolioFilter, query]);

  const isEditing = Boolean(form.id);

  async function loadAgents() {
    setLoading(true);
    try {
      const qs = new URLSearchParams({ scope: "retail", role: "agent", limit: "200", offset: "0" });
      const response = await fetch(`/api/admin/members?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const body = (await response.json().catch(() => ({}))) as MembersResp;
      if (!response.ok || !body.ok) {
        throw new Error(body.error || "Gagal mengambil data agent");
      }
      setItems(Array.isArray(body.items) ? body.items : Array.isArray(body.rows) ? body.rows : []);
    } catch (err) {
      setMessage({ type: "error", text: err instanceof Error ? err.message : "Gagal mengambil data agent" });
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void loadAgents();
  }, []);

  function resetForm() {
    setForm(emptyForm());
    setMessage(null);
  }

  function editAgent(item: AgentRow) {
    setMessage(null);
    setForm({
      id: item.id,
      name: item.nama || "",
      email: item.email || "",
      phone: item.phone || "",
      password: "",
      commission: String(item.retail_agent_commission_rp ?? 0),
      aktif: Boolean(item.aktif),
    });
  }

  async function copyPassword() {
    if (!form.password) return;
    await navigator.clipboard?.writeText(form.password).catch(() => undefined);
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!form.name.trim() || !form.email.trim()) {
      setMessage({ type: "error", text: "Nama dan email agent wajib diisi." });
      return;
    }
    if (!isEditing) return;

    setSaving(true);
    setMessage(null);
    try {
      const response = await fetch(`/api/admin/members/${form.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({
          email: form.email.trim(),
          nama: form.name.trim(),
          phone: form.phone.trim(),
          role: "agent",
          aktif: form.aktif,
          fee_member_rp: 0,
          retail_agent_commission_rp: Math.max(0, Math.floor(Number(form.commission || 0))),
          retail_master_commission_rp: 0,
          h2h_agent_commission_rp: 0,
          h2h_master_commission_rp: 0,
          new_password: form.password.trim() || undefined,
        }),
      });
      const body = (await response.json().catch(() => ({}))) as ApiResp;
      if (!response.ok || !body.ok) {
        throw new Error(body.error || "Data agent gagal disimpan");
      }
      setMessage({ type: "success", text: "Agent berhasil diperbarui." });
      setForm(emptyForm());
      await loadAgents();
    } catch (err) {
      setMessage({ type: "error", text: err instanceof Error ? err.message : "Data agent gagal disimpan" });
    } finally {
      setSaving(false);
    }
  }

  async function toggleAgent(item: AgentRow) {
    setSaving(true);
    setMessage(null);
    try {
      const response = await fetch(`/api/admin/members/${item.id}`, {
        method: item.aktif ? "DELETE" : "PUT",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: item.aktif
          ? undefined
          : JSON.stringify({
              email: item.email,
              nama: item.nama,
              phone: item.phone || "",
              role: "agent",
              aktif: true,
              fee_member_rp: Number(item.fee_member_rp || 0),
              retail_agent_commission_rp: Number(item.retail_agent_commission_rp || 0),
              retail_master_commission_rp: Number(item.retail_master_commission_rp || 0),
              h2h_agent_commission_rp: Number(item.h2h_agent_commission_rp || 0),
              h2h_master_commission_rp: Number(item.h2h_master_commission_rp || 0),
            }),
      });
      const body = (await response.json().catch(() => ({}))) as ApiResp;
      if (!response.ok || !body.ok) {
        throw new Error(body.error || "Status agent gagal diubah");
      }
      setMessage({ type: "success", text: item.aktif ? "Agent berhasil dihapus dari akun aktif." : "Agent diaktifkan kembali." });
      await loadAgents();
    } catch (err) {
      setMessage({ type: "error", text: err instanceof Error ? err.message : "Status agent gagal diubah" });
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className={isEditing ? "grid gap-5 xl:grid-cols-[minmax(0,1fr)_380px]" : "block"}>
      <section className="overflow-hidden rounded-lg border border-emerald-200 bg-white shadow-[0_16px_36px_rgba(6,78,59,0.08)]">
        <div className="flex items-center justify-between gap-4 bg-[linear-gradient(135deg,#13464b,#075f46)] px-4 py-4 text-white sm:px-5">
          <div className="min-w-0">
            <p className="text-[9px] font-black uppercase tracking-[0.18em] text-lime-200">Portofolio Kredit</p>
            <h1 className="mt-1 text-xl font-black sm:text-2xl">Daftar Kredit Agent Binaan</h1>
            <p className="mt-1 text-xs font-semibold leading-5 text-emerald-100/75">Pantau akun, status kredit, saldo, dan tindak lanjut setiap agent.</p>
          </div>
          <div className="shrink-0 text-center">
            <p className="text-2xl font-black text-lime-200">{items.length}</p>
            <p className="text-[9px] font-bold text-emerald-100/70">Agent</p>
          </div>
        </div>

        <div className="border-b border-emerald-100 bg-[#f8fffb] p-3 sm:p-4">
          <div className="flex flex-col gap-2 lg:flex-row">
            <label className="flex h-11 min-w-0 flex-1 items-center gap-2 rounded-lg border border-slate-200 bg-white px-3 text-sm font-bold text-slate-500 focus-within:border-emerald-400 focus-within:ring-2 focus-within:ring-emerald-100">
              <Search className="h-4 w-4 text-emerald-600" />
              <input value={query} onChange={(event) => setQuery(event.target.value)} className="w-full bg-transparent outline-none" placeholder="Cari agent, toko, WhatsApp, atau ID..." />
            </label>
            <div className="flex gap-2 overflow-x-auto pb-1 lg:pb-0">
              {([
                ["all", "Semua"],
                ["survey", "Survei"],
                ["operator", "Operator"],
                ["active", "Aktif"],
                ["paid", "Siklus selesai"],
              ] as const).map(([value, label]) => (
                <button
                  key={value}
                  type="button"
                  onClick={() => setPortfolioFilter(value)}
                  className={portfolioFilter === value
                    ? "h-11 shrink-0 rounded-lg border border-emerald-700 bg-emerald-700 px-4 text-xs font-black text-white"
                    : "h-11 shrink-0 rounded-lg border border-slate-200 bg-white px-4 text-xs font-black text-slate-600 transition hover:border-emerald-300 hover:text-emerald-700"}
                >
                  {label}
                </button>
              ))}
            </div>
          </div>
        </div>

        {message ? (
          <div className={message.type === "success" ? "m-3 rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-xs font-black text-emerald-700" : "m-3 rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-xs font-black text-rose-600"}>
            {message.text}
          </div>
        ) : null}

        <div className="p-3 sm:p-4">
          {loading ? (
            <div className="space-y-2">
              {Array.from({ length: 4 }).map((_, index) => <div key={index} className="h-16 animate-pulse rounded-lg bg-slate-100" />)}
            </div>
          ) : filteredItems.length ? (
            <div className="overflow-hidden rounded-lg border border-slate-200">
              <div className="hidden grid-cols-[minmax(220px,1.5fr)_130px_135px_165px_160px] gap-3 border-b border-slate-200 bg-slate-50 px-4 py-3 text-[10px] font-black uppercase tracking-[0.1em] text-slate-500 lg:grid">
                <span>Agent</span><span>Status Kredit</span><span>Saldo Utama</span><span>Pengajuan / Modal</span><span className="text-center">Aksi</span>
              </div>
              {filteredItems.map((item) => {
                const application = latestApplicationByMember.get(item.id);
                const stage = applicationStage(application);
                const status = stageMeta[stage];
                const requested = Number(application?.approved_amount || application?.requested_amount || 0);
                const outstanding = Number(application?.outstanding_amount || 0);
                const storeName = String(application?.store_name || application?.applicant_data?.store_name || "Toko belum diisi");

                return (
                  <article key={item.id} className="border-b border-slate-100 bg-white p-4 last:border-b-0 hover:bg-emerald-50/30">
                    <div className="grid gap-3 lg:grid-cols-[minmax(220px,1.5fr)_130px_135px_165px_160px] lg:items-center">
                      <div className="flex min-w-0 items-center gap-3">
                        <span className="grid h-10 w-10 shrink-0 place-items-center rounded-lg bg-emerald-950 text-xs font-black text-lime-300">{(item.nama || "AG").slice(0, 2).toUpperCase()}</span>
                        <div className="min-w-0">
                          <div className="flex items-center gap-2"><p className="truncate text-sm font-black text-slate-950">{item.nama || "Agent"}</p>{!item.aktif ? <span className="rounded bg-rose-100 px-1.5 py-0.5 text-[9px] font-black text-rose-700">Nonaktif</span> : null}</div>
                          <p className="mt-1 truncate text-xs font-semibold text-slate-500">{storeName}</p>
                          <p className="mt-0.5 truncate text-[11px] font-bold text-slate-400">{item.phone || "-"} - ID {item.id}</p>
                        </div>
                      </div>
                      <div><span className={`inline-flex rounded-md px-2.5 py-1.5 text-[10px] font-black ${status.className}`}>{status.label}</span></div>
                      <p className="text-sm font-black text-slate-950">{formatIDR(Number(item.saldo || 0))}</p>
                      <div><p className="text-sm font-black text-slate-950">{formatIDR(requested)}</p><p className={outstanding > 0 ? "mt-1 text-[11px] font-bold text-amber-700" : "mt-1 text-[11px] font-bold text-slate-400"}>Modal berjalan {formatIDR(outstanding)}</p></div>
                      <div className="grid grid-cols-2 gap-2">
                        <button type="button" onClick={() => editAgent(item)} className="inline-flex h-9 items-center justify-center gap-1.5 rounded-lg border border-emerald-200 bg-emerald-50 text-[11px] font-black text-emerald-700"><Edit3 className="h-3.5 w-3.5" />Edit</button>
                        <button type="button" onClick={() => toggleAgent(item)} disabled={saving} className={item.aktif ? "h-9 rounded-lg border border-rose-200 bg-rose-50 text-[11px] font-black text-rose-600" : "h-9 rounded-lg bg-emerald-950 text-[11px] font-black text-white"}>{item.aktif ? "Nonaktifkan" : "Aktifkan"}</button>
                      </div>
                    </div>
                  </article>
                );
              })}
            </div>
          ) : (
            <div className="grid min-h-28 place-items-center rounded-lg border border-dashed border-emerald-300 bg-emerald-50/35 px-4 text-center text-sm font-semibold text-slate-500">Agent binaan tidak ditemukan.</div>
          )}
        </div>
      </section>

      {isEditing ? (
      <form onSubmit={submit} className="rounded-[28px] border border-emerald-100 bg-[linear-gradient(180deg,#ffffff_0%,#f3fff8_100%)] p-4 shadow-[0_18px_42px_rgba(6,78,59,0.08)] sm:p-5">
        <div className="flex items-start justify-between gap-3">
          <div>
            <p className="text-[11px] font-black uppercase tracking-[0.18em] text-emerald-600">Edit Agent</p>
            <h2 className="mt-1 text-xl font-black text-slate-950">{form.name || "Edit Agent"}</h2>
          </div>
          <button type="button" onClick={resetForm} className="grid h-10 w-10 place-items-center rounded-2xl bg-white text-slate-500 ring-1 ring-slate-200">
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="mt-5 space-y-3">
          <InputShell label="Nama Agent" icon={<UserRound className="h-4 w-4" />}>
            <input value={form.name} onChange={(event) => setForm((state) => ({ ...state, name: event.target.value }))} placeholder="Nama agent" className={inputClassName} />
          </InputShell>
          <InputShell label="Email Login" icon={<Mail className="h-4 w-4" />}>
            <input value={form.email} onChange={(event) => setForm((state) => ({ ...state, email: event.target.value }))} placeholder="agent@pulsakilat.local" type="email" className={inputClassName} />
          </InputShell>
          <InputShell label="Nomor HP" icon={<Phone className="h-4 w-4" />}>
            <input value={form.phone} onChange={(event) => setForm((state) => ({ ...state, phone: event.target.value }))} placeholder="08xxxxxxxxxx" inputMode="tel" className={inputClassName} />
          </InputShell>
          <InputShell label="Password Baru" icon={<KeyRound className="h-4 w-4" />}>
            <div className="flex gap-2">
              <input value={form.password} onChange={(event) => setForm((state) => ({ ...state, password: event.target.value }))} placeholder="Kosongkan jika tidak diubah" className={`${inputClassName} min-w-0 flex-1`} />
              <button type="button" onClick={() => setForm((state) => ({ ...state, password: makePassword() }))} className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-emerald-50 text-emerald-700 ring-1 ring-emerald-100">
                <RefreshCcw className="h-4 w-4" />
              </button>
              <button type="button" onClick={copyPassword} className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-white text-slate-600 ring-1 ring-slate-200">
                <Copy className="h-4 w-4" />
              </button>
            </div>
          </InputShell>
          <label className="flex items-center justify-between rounded-2xl border border-slate-200 bg-white px-4 py-3">
            <span>
              <span className="block text-xs font-black text-slate-950">Status akun</span>
              <span className="text-[11px] font-semibold text-slate-400">Aktifkan atau nonaktifkan agent</span>
            </span>
            <input type="checkbox" checked={form.aktif} onChange={(event) => setForm((state) => ({ ...state, aktif: event.target.checked }))} className="h-5 w-5 rounded border-slate-300 text-emerald-600" />
          </label>
        </div>

        {message ? (
          <div className={message.type === "success" ? "mt-4 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-xs font-black text-emerald-700" : "mt-4 rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-xs font-black text-rose-600"}>
            {message.text}
          </div>
        ) : null}

        <button type="submit" disabled={saving} className="mt-4 flex h-12 w-full items-center justify-center gap-2 rounded-[20px] bg-[linear-gradient(135deg,#13464b,#087e8b,#84cc16)] text-sm font-black text-white shadow-[0_16px_30px_rgba(5,150,105,0.20)] disabled:opacity-50">
          {saving ? <Loader2 className="h-5 w-5 animate-spin" /> : <BadgeCheck className="h-5 w-5" />}
          {saving ? "Menyimpan..." : "Simpan Agent"}
        </button>
      </form>
      ) : null}
    </div>
  );
}

function InputShell({ label, icon, children }: { label: string; icon: ReactNode; children: ReactNode }) {
  return (
    <label className="block">
      <span className="mb-2 flex items-center gap-2 text-[10px] font-black uppercase tracking-[0.12em] text-slate-500">
        <span className="text-emerald-600">{icon}</span>
        {label}
      </span>
      {children}
    </label>
  );
}
