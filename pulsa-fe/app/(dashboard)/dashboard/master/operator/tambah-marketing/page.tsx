"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { Activity, CircleCheck, Mail, Pencil, Phone, RefreshCw, Search, ShieldCheck, UserPlus, UsersRound, UserX } from "lucide-react";
import RegisterMemberModal from "@/components/dashboard/RegisterMemberModal";
import { AppModal } from "@/components/ui/app-modal";
import { alertSuccess, alertWarning } from "@/components/ui/alerts";

type MarketingAccount = { id: number; nama: string; email: string; phone: string; aktif: boolean; agent_count: number; active_agent_count: number; last_agent_transaction_at?: string; dibuat_pada?: string };
type ManagementResponse = { ok?: boolean; accounts?: MarketingAccount[]; error?: string };
type StatusFilter = "all" | "active" | "inactive";

function authHeader(): Record<string, string> {
  const token = typeof window === "undefined" ? "" : localStorage.getItem("auth_token") || "";
  return token ? { Authorization: `Bearer ${token}` } : {};
}

export default function OperatorTambahMarketingPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const [selected, setSelected] = useState<MarketingAccount | null>(null);
  const [accounts, setAccounts] = useState<MarketingAccount[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState<StatusFilter>("all");
  const [error, setError] = useState("");

  const loadAccounts = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const response = await fetch("/api/operator/marketing/create", { headers: authHeader(), cache: "no-store" });
      const body = (await response.json().catch(() => ({}))) as ManagementResponse;
      if (!response.ok || !body.ok) throw new Error(body.error || "Akun marketing tidak dapat dimuat");
      setAccounts(Array.isArray(body.accounts) ? body.accounts : []);
      setSelected((current) => current ? (body.accounts || []).find((item) => item.id === current.id) || null : null);
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : "Akun marketing tidak dapat dimuat");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { void loadAccounts(); }, [loadAccounts]);

  const visibleAccounts = useMemo(() => {
    const query = search.trim().toLowerCase();
    return accounts.filter((account) => {
      const matchesSearch = !query || [account.nama, account.email, account.phone].some((value) => String(value || "").toLowerCase().includes(query));
      const matchesStatus = status === "all" || (status === "active" ? account.aktif : !account.aktif);
      return matchesSearch && matchesStatus;
    });
  }, [accounts, search, status]);

  const activeCount = accounts.filter((account) => account.aktif).length;
  const inactiveCount = accounts.length - activeCount;
  const totalAgents = accounts.reduce((total, account) => total + Number(account.agent_count || 0), 0);

  return (
    <main className="min-h-screen bg-[#eef7f2] p-3 text-slate-950 sm:p-6 lg:p-8">
      <div className="mx-auto w-full max-w-6xl">
        <header className="overflow-hidden rounded-lg bg-[#075c46] text-white shadow-[0_16px_36px_rgba(4,80,62,0.18)]">
          <div className="flex flex-col gap-5 p-5 sm:flex-row sm:items-center sm:justify-between sm:p-7">
            <div className="flex items-start gap-4"><span className="grid h-12 w-12 shrink-0 place-items-center rounded-lg bg-white text-emerald-800"><UsersRound className="h-6 w-6" /></span><div><p className="text-[10px] font-black uppercase tracking-[0.2em] text-lime-200">Manajemen Tim</p><h1 className="mt-1 text-2xl font-black sm:text-3xl">Akun Marketing</h1><p className="mt-2 max-w-2xl text-xs font-semibold leading-5 text-emerald-50 sm:text-sm">Pantau pembagian agent, aktivitas terbaru, dan status akun marketing dari satu tempat.</p></div></div>
            <button type="button" onClick={() => setCreateOpen(true)} className="inline-flex h-11 shrink-0 items-center justify-center gap-2 rounded-lg bg-lime-300 px-5 text-sm font-black text-emerald-950 shadow-sm transition hover:bg-lime-200"><UserPlus className="h-5 w-5" /> Tambah Marketing</button>
          </div>
        </header>

        <section className="mt-4 grid grid-cols-2 gap-3 lg:grid-cols-4">
          <Summary icon={<UsersRound className="h-5 w-5" />} label="Total Marketing" value={accounts.length} tone="emerald" />
          <Summary icon={<ShieldCheck className="h-5 w-5" />} label="Akun Aktif" value={activeCount} tone="blue" />
          <Summary icon={<UserX className="h-5 w-5" />} label="Akun Nonaktif" value={inactiveCount} tone="amber" />
          <Summary icon={<Activity className="h-5 w-5" />} label="Agent Terhubung" value={totalAgents} tone="lime" />
        </section>

        <section className="mt-4 overflow-hidden rounded-lg border border-emerald-100 bg-white shadow-[0_12px_30px_rgba(15,23,42,0.06)]">
          <div className="border-b border-slate-100 p-4 sm:p-5"><div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between"><div><h2 className="text-lg font-black">Daftar Marketing</h2><p className="mt-1 text-xs font-semibold text-slate-500">{visibleAccounts.length} akun ditampilkan</p></div><div className="flex min-w-0 flex-col gap-2 sm:flex-row"><label className="flex h-10 min-w-0 flex-1 items-center gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 sm:w-72"><Search className="h-4 w-4 shrink-0 text-emerald-700" /><input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Cari nama, email, atau nomor" className="min-w-0 flex-1 bg-transparent text-xs font-semibold outline-none placeholder:text-slate-400" /></label><select value={status} onChange={(event) => setStatus(event.target.value as StatusFilter)} className="h-10 rounded-lg border border-slate-200 bg-white px-3 text-xs font-bold text-slate-700 outline-none"><option value="all">Semua status</option><option value="active">Aktif</option><option value="inactive">Nonaktif</option></select><button type="button" onClick={() => void loadAccounts()} disabled={loading} title="Muat ulang" className="grid h-10 w-10 shrink-0 place-items-center rounded-lg border border-emerald-200 bg-emerald-50 text-emerald-800 disabled:opacity-60"><RefreshCw className={loading ? "h-4 w-4 animate-spin" : "h-4 w-4"} /></button></div></div></div>
          {error ? <div className="m-4 rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-xs font-bold text-rose-700 sm:m-5">{error}</div> : null}
          <div className="hidden overflow-x-auto md:block"><table className="w-full min-w-[900px] text-left text-xs"><thead className="bg-[#e8f7f0] text-[10px] font-black uppercase tracking-[0.12em] text-emerald-900"><tr><th className="px-5 py-3">Marketing</th><th className="px-4 py-3">Kontak</th><th className="px-4 py-3">Agent Binaan</th><th className="px-4 py-3">Aktivitas Terakhir</th><th className="px-4 py-3">Status</th><th className="px-5 py-3 text-right">Aksi</th></tr></thead><tbody>
            {visibleAccounts.map((account) => <tr key={account.id} className="border-t border-slate-100 hover:bg-emerald-50/50"><td className="px-5 py-4"><AccountIdentity account={account} /></td><td className="px-4 py-4"><p className="inline-flex items-center gap-2 font-semibold text-slate-700"><Mail className="h-3.5 w-3.5 text-emerald-600" />{account.email || "-"}</p><p className="mt-1 inline-flex items-center gap-2 text-[10px] font-semibold text-slate-500"><Phone className="h-3 w-3" />{account.phone || "Belum tersedia"}</p></td><td className="px-4 py-4"><p className="font-black text-slate-900">{account.agent_count || 0} agent</p><p className="mt-1 text-[10px] font-semibold text-emerald-700">{account.active_agent_count || 0} aktif</p></td><td className="px-4 py-4 font-semibold text-slate-600">{formatDateTime(account.last_agent_transaction_at)}</td><td className="px-4 py-4"><StatusBadge active={account.aktif} /></td><td className="px-5 py-4 text-right"><button onClick={() => setSelected(account)} className="inline-flex h-9 items-center gap-2 rounded-lg border border-emerald-200 bg-white px-3 font-black text-emerald-800 shadow-sm"><Pencil className="h-3.5 w-3.5" /> Kelola</button></td></tr>)}
            <EmptyRows loading={loading} empty={visibleAccounts.length === 0} colSpan={6} />
          </tbody></table></div>
          <div className="divide-y divide-slate-100 md:hidden">{visibleAccounts.map((account) => <article key={account.id} className="p-4"><div className="flex items-start justify-between gap-3"><AccountIdentity account={account} /><StatusBadge active={account.aktif} /></div><div className="mt-4 grid grid-cols-2 gap-2"><Info label="Agent binaan" value={`${account.agent_count || 0} agent`} /><Info label="Aktivitas terakhir" value={formatDateTime(account.last_agent_transaction_at)} /></div><button onClick={() => setSelected(account)} className="mt-3 inline-flex h-10 w-full items-center justify-center gap-2 rounded-lg bg-emerald-800 text-xs font-black text-white"><Pencil className="h-4 w-4" /> Kelola Akun</button></article>)}{!loading && visibleAccounts.length === 0 ? <p className="px-4 py-14 text-center text-xs font-semibold text-slate-400">Akun marketing tidak ditemukan.</p> : null}{loading ? <p className="px-4 py-14 text-center text-xs font-semibold text-slate-400">Memuat akun marketing...</p> : null}</div>
        </section>
      </div>
      <RegisterMemberModal open={createOpen} onClose={() => setCreateOpen(false)} fixedRole="marketing" title="Tambah Marketing" subtitle="Buat akun marketing Isiloka untuk memantau agent binaan." theme="retail" createEndpoint="/api/operator/marketing/create" onSuccess={loadAccounts} />
      <ManageMarketingModal account={selected} onClose={() => setSelected(null)} onSaved={loadAccounts} />
    </main>
  );
}

function ManageMarketingModal({ account, onClose, onSaved }: { account: MarketingAccount | null; onClose: () => void; onSaved: () => Promise<void> }) {
  const [nama, setNama] = useState(""); const [phone, setPhone] = useState(""); const [password, setPassword] = useState(""); const [active, setActive] = useState(true); const [saving, setSaving] = useState(false);
  useEffect(() => { if (account) { setNama(account.nama || ""); setPhone(account.phone || ""); setPassword(""); setActive(account.aktif); } }, [account]);
  async function patch(payload: Record<string, unknown>) { const response = await fetch("/api/operator/marketing/create", { method: "PATCH", headers: { "Content-Type": "application/json", ...authHeader() }, body: JSON.stringify(payload) }); const body = (await response.json().catch(() => ({}))) as { ok?: boolean; error?: string }; if (!response.ok || !body.ok) throw new Error(body.error || "Perubahan tidak dapat disimpan"); }
  async function saveAccount() { if (!account) return; if (!nama.trim() || !phone.trim()) return alertWarning("Nama dan nomor telepon marketing wajib diisi."); if (password && password.length < 8) return alertWarning("Password baru minimal 8 karakter."); setSaving(true); try { await patch({ action: "update_account", id: account.id, nama: nama.trim(), phone: phone.trim(), new_password: password, aktif: active }); await onSaved(); await alertSuccess("Akun marketing berhasil diperbarui."); } catch (err) { await alertWarning(err instanceof Error ? err.message : "Perubahan tidak dapat disimpan"); } finally { setSaving(false); } }
  return <AppModal open={Boolean(account)} onClose={onClose} title="Kelola Akun Marketing" subtitle={account ? `${account.nama || "Marketing"} · ${account.email}` : ""} theme="retail" maxWidthClassName="max-w-3xl" footer={<div className="flex w-full flex-col-reverse gap-2 sm:flex-row sm:justify-end"><button onClick={onClose} className="h-11 rounded-lg border border-slate-200 px-5 text-sm font-black">Tutup</button><button onClick={() => void saveAccount()} disabled={saving} className="h-11 rounded-lg bg-emerald-800 px-5 text-sm font-black text-white disabled:opacity-60">{saving ? "Menyimpan..." : "Simpan Perubahan"}</button></div>}>
    <div className="grid gap-4 sm:grid-cols-2"><Field label="Nama Marketing" value={nama} onChange={setNama} /><Field label="Nomor Telepon" value={phone} onChange={setPhone} type="tel" /><Field label="Password Baru" value={password} onChange={setPassword} type="password" placeholder="Kosongkan jika tidak diubah" /><label className="grid gap-2"><span className="text-xs font-black text-slate-700">Status Akun</span><select value={active ? "active" : "inactive"} onChange={(event) => setActive(event.target.value === "active")} className="h-11 rounded-lg border border-slate-200 bg-white px-3 text-sm font-bold outline-none"><option value="active">Aktif</option><option value="inactive">Nonaktif</option></select></label></div>
  </AppModal>;
}

function Field({ label, value, onChange, type = "text", placeholder = "" }: { label: string; value: string; onChange: (value: string) => void; type?: string; placeholder?: string }) { return <label className="grid gap-2"><span className="text-xs font-black text-slate-700">{label}</span><input value={value} onChange={(event) => onChange(event.target.value)} type={type} placeholder={placeholder} className="h-11 rounded-lg border border-slate-200 px-3 text-sm font-bold outline-none focus:border-emerald-500" /></label>; }
function AccountIdentity({ account }: { account: MarketingAccount }) { return <div className="flex min-w-0 items-center gap-3"><span className="grid h-9 w-9 shrink-0 place-items-center rounded-lg bg-emerald-100 font-black text-emerald-800">{(account.nama || "M").slice(0, 1).toUpperCase()}</span><div className="min-w-0"><p className="truncate font-black">{account.nama || "Marketing"}</p><p className="mt-0.5 text-[10px] font-semibold text-slate-400">ID #{account.id} · {formatDate(account.dibuat_pada)}</p></div></div>; }
function StatusBadge({ active }: { active: boolean }) { return <span className={`inline-flex shrink-0 items-center gap-1.5 rounded-full px-2.5 py-1 text-[10px] font-black ${active ? "bg-emerald-100 text-emerald-800" : "bg-rose-100 text-rose-700"}`}><CircleCheck className="h-3.5 w-3.5" />{active ? "Aktif" : "Nonaktif"}</span>; }
function Summary({ icon, label, value, tone }: { icon: React.ReactNode; label: string; value: number; tone: "emerald" | "blue" | "amber" | "lime" }) { const tones = { emerald: "bg-emerald-100 text-emerald-800", blue: "bg-sky-100 text-sky-700", amber: "bg-amber-100 text-amber-700", lime: "bg-lime-100 text-lime-800" }; return <div className="flex min-h-24 items-center justify-between rounded-lg border border-slate-200 bg-white p-4 shadow-[0_8px_22px_rgba(15,23,42,0.04)]"><div><p className="text-[9px] font-black uppercase tracking-[0.08em] text-slate-500 sm:text-[10px]">{label}</p><p className="mt-1 text-2xl font-black">{value}</p></div><span className={`grid h-10 w-10 shrink-0 place-items-center rounded-lg ${tones[tone]}`}>{icon}</span></div>; }
function Info({ label, value }: { label: string; value: string }) { return <div className="rounded-lg bg-slate-50 p-3"><p className="text-[9px] font-black uppercase text-slate-400">{label}</p><p className="mt-1 text-xs font-black text-slate-800">{value}</p></div>; }
function EmptyRows({ loading, empty, colSpan }: { loading: boolean; empty: boolean; colSpan: number }) { if (!loading && !empty) return null; return <tr><td colSpan={colSpan} className="px-5 py-14 text-center font-semibold text-slate-400">{loading ? "Memuat akun marketing..." : "Akun marketing tidak ditemukan."}</td></tr>; }
function formatDate(value?: string) { return value ? new Intl.DateTimeFormat("id-ID", { day: "2-digit", month: "short", year: "numeric" }).format(new Date(value)) : "-"; }
function formatDateTime(value?: string) { return value ? new Intl.DateTimeFormat("id-ID", { day: "2-digit", month: "short", year: "numeric", hour: "2-digit", minute: "2-digit" }).format(new Date(value)) : "Belum ada transaksi"; }
