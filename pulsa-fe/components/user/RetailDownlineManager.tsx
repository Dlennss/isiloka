"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { ArrowLeft, ChevronRight, UserPlus, Users, X } from "lucide-react";

type RetailDownline = {
  id: number;
  email: string;
  nama: string;
  phone?: string;
  store_name?: string;
  role: string;
  aktif: boolean;
  saldo: number;
  retail_agent_nama?: string | null;
  retail_master_nama?: string | null;
  marketing_nama?: string | null;
  marketing_email?: string | null;
};

type Props = {
  authToken: string;
  role: string;
  allowCreate?: boolean;
};

function roleLabel(role: string) {
  switch (String(role || "").toLowerCase()) {
    case "master":
      return "Master";
    case "agent":
      return "Agent";
    case "user":
      return "User";
    default:
      return role || "-";
  }
}

function fmtIDR(v: number) {
  return new Intl.NumberFormat("id-ID").format(Number(v || 0));
}

export function RetailDownlineManager({ authToken, role, allowCreate = true }: Props) {
  const [items, setItems] = useState<RetailDownline[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [err, setErr] = useState("");
  const [ok, setOk] = useState("");
  const [email, setEmail] = useState("");
  const [nama, setNama] = useState("");
  const [password, setPassword] = useState("");
  const [targetRole, setTargetRole] = useState(role === "marketing" ? "agent" : "user");

  const canCreate = allowCreate && (role === "master" || role === "agent" || role === "marketing");
  const activeCount = items.filter((item) => item.aktif).length;
  const agentCount = items.filter((item) => String(item.role || "").toLowerCase() === "agent").length;
  const roleOptions = useMemo(
    () => (role === "master" ? [
      { value: "user", label: "User" },
      { value: "agent", label: "Agent" },
    ] : role === "marketing" ? [{ value: "agent", label: "Agent" }] : [{ value: "user", label: "User" }]),
    [role],
  );

  const load = useCallback(async () => {
    setLoading(true);
    setErr("");
    try {
      const r = await fetch("/api/me/retail/downlines", {
        headers: { Authorization: `Bearer ${authToken}` },
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) {
        setErr(j?.error || "Gagal memuat jaringan retail.");
        setItems([]);
        return;
      }
      setItems(Array.isArray(j.items) ? j.items : []);
    } finally {
      setLoading(false);
    }
  }, [authToken]);

  useEffect(() => {
    void load();
  }, [load]);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setErr("");
    setOk("");
    if (!email.trim() || !nama.trim() || password.trim().length < 8) {
      setErr("Email, nama, dan password minimal 8 karakter wajib diisi.");
      return;
    }
    setSaving(true);
    try {
      const r = await fetch("/api/me/retail/downlines", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${authToken}`,
        },
        body: JSON.stringify({
          email: email.trim(),
          nama: nama.trim(),
          password: password.trim(),
          role: targetRole,
        }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) {
        setErr(j?.error || "Gagal membuat akun retail.");
        return;
      }
      setOk(`${role === "marketing" ? "Agent binaan" : role === "agent" ? "Member" : "Akun retail"} berhasil dibuat dan langsung terhubung ke akun anda. member_id=${j.member_id}`);
      setEmail("");
      setNama("");
      setPassword("");
      setTargetRole(role === "marketing" ? "agent" : "user");
      setShowCreateModal(false);
      await load();
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="space-y-4">
      <section className="relative overflow-hidden rounded-[30px] bg-[linear-gradient(135deg,#052e26_0%,#047857_58%,#84cc16_145%)] p-5 text-white shadow-[0_20px_44px_rgba(4,120,87,0.22)]">
        <div className="pointer-events-none absolute -right-12 -top-16 h-40 w-40 rounded-full bg-white/10" />
        <div className="pointer-events-none absolute right-10 top-8 h-24 w-24 rounded-full border border-white/10" />
        <div className="relative flex items-center gap-3">
          <Link href="/user/account" className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-white/12 text-white ring-1 ring-white/15">
            <ArrowLeft className="h-5 w-5" />
          </Link>
          <span className="grid h-12 w-12 shrink-0 place-items-center rounded-2xl bg-white text-[#087e8b] shadow-[0_10px_22px_rgba(0,0,0,0.10)]">
            <Users className="h-5 w-5" />
          </span>
          <div className="min-w-0">
            <div className="text-[10px] font-black uppercase tracking-[0.22em] text-lime-100">{role === "marketing" ? "Agent Binaan" : role === "agent" ? "Member Bawahan" : "Downline Retail"}</div>
            <h1 className="mt-1 text-xl font-black">{role === "marketing" ? "Agent Binaan" : role === "agent" ? "Tambah Member" : "Jaringan Retail"}</h1>
            <p className="mt-1 text-xs font-semibold leading-5 text-white/75">
              {role === "master" ? "Master dapat menambahkan agent atau user." : role === "agent" ? "Agent dapat menambahkan user." : "Daftar jaringan di atas akun retail."}
            </p>
          </div>
        </div>
        <div className="relative mt-5 grid grid-cols-3 gap-2">
          {[
            ["Total", items.length],
            ["Aktif", activeCount],
            ["Agent", agentCount],
          ].map(([label, value]) => (
            <div key={label} className="rounded-2xl bg-white/12 px-3 py-2 ring-1 ring-white/14">
              <p className="text-[9px] font-black uppercase tracking-[0.12em] text-white/55">{label}</p>
              <p className="mt-1 text-lg font-black text-white">{value}</p>
            </div>
          ))}
        </div>
      </section>
      <section className="overflow-hidden rounded-[28px] border border-emerald-950/5 bg-white shadow-[0_18px_42px_rgba(6,78,59,0.08)]">
        <div className="flex items-center justify-between gap-3 border-b border-slate-100 bg-[linear-gradient(135deg,#f8fffb,#eefbf4)] px-4 py-4">
          <div>
            <h2 className="text-sm font-black text-slate-950">{role === "marketing" ? "Daftar Agent Binaan" : role === "agent" ? "Member Terdaftar" : "Daftar Agent/User"}</h2>
            <p className="mt-1 text-xs font-semibold text-slate-500">{role === "marketing" ? "Agent yang terhubung ke akun marketing ini." : role === "agent" ? "Member yang terhubung ke akun agent ini." : "Downline retail yang terhubung ke akun ini."}</p>
          </div>
          {canCreate ? (
            <button
              type="button"
              onClick={() => {
                setErr("");
                setOk("");
                setShowCreateModal(true);
              }}
              className="inline-flex h-10 shrink-0 items-center gap-2 rounded-2xl bg-[#13464b] px-3 text-xs font-black text-white shadow-[0_10px_20px_rgba(6,78,59,0.18)] transition hover:bg-[#13464b]"
            >
              <UserPlus className="h-4 w-4" />
              Tambah
            </button>
          ) : null}
        </div>

        <div className="space-y-3 p-4">
          {loading ? <div className="text-sm font-semibold text-slate-500">Memuat jaringan retail...</div> : null}
          {!loading && items.length === 0 ? (
            <div className="grid min-h-[180px] place-items-center rounded-[24px] border border-dashed border-emerald-200 bg-emerald-50/50 px-4 text-center">
              <div>
                <div className="mx-auto grid h-14 w-14 place-items-center rounded-2xl bg-white text-[#087e8b] shadow-sm">
                  <Users className="h-7 w-7" />
                </div>
                <p className="mt-3 text-sm font-black text-slate-950">Belum ada jaringan</p>
                <p className="mt-1 text-xs font-semibold leading-5 text-slate-500">{role === "marketing" ? "Agent yang kamu tambahkan akan tampil di sini." : role === "agent" ? "Member yang kamu tambahkan akan tampil di sini." : "Agent atau user yang kamu tambahkan akan tampil di sini."}</p>
              </div>
            </div>
          ) : null}
          {items.map((item) => (
            <div key={item.id} className="rounded-[22px] border border-slate-200 bg-[linear-gradient(180deg,#ffffff_0%,#f8fffb_100%)] px-4 py-3 shadow-[0_8px_20px_rgba(15,23,42,0.04)]">
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <div className="truncate text-sm font-black text-slate-950">{item.nama || item.email}</div>
                  <div className="truncate text-xs font-semibold text-slate-500">{item.email}</div>
                  {item.store_name ? <div className="mt-1 truncate text-[11px] font-bold text-emerald-700">Toko: {item.store_name}</div> : null}
                  {item.phone ? <div className="mt-1 text-[11px] font-semibold text-slate-500">WA {item.phone}</div> : null}
                  <div className="mt-1 text-[11px] font-semibold text-slate-500">
                    {roleLabel(item.role)}
                    {item.retail_agent_nama ? ` • Agent: ${item.retail_agent_nama}` : ""}
                    {item.retail_master_nama ? ` • Master: ${item.retail_master_nama}` : ""}
                  </div>
                  {item.marketing_nama || item.marketing_email ? (
                    <div className="mt-2 rounded-xl bg-emerald-50 px-2.5 py-2 text-[11px] font-bold text-emerald-800">
                      Marketing pembina: {item.marketing_nama || item.marketing_email}
                      {item.marketing_nama && item.marketing_email ? <span className="block font-semibold text-emerald-700/75">{item.marketing_email}</span> : null}
                    </div>
                  ) : null}
                </div>
                <div className="text-right">
                  <div className={`inline-flex rounded-full px-2.5 py-1 text-[9px] font-black uppercase tracking-[0.12em] ${item.aktif ? "bg-emerald-100 text-emerald-700" : "bg-slate-100 text-slate-500"}`}>
                    {item.aktif ? "Aktif" : "Nonaktif"}
                  </div>
                  <div className="mt-2 text-sm font-black text-[#087e8b]">Rp {fmtIDR(item.saldo)}</div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </section>

      {showCreateModal ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/45 p-4 backdrop-blur-[2px]">
          <div className="max-h-[calc(100dvh-2rem)] w-full max-w-md overflow-y-auto rounded-[30px] bg-white p-5 shadow-[0_18px_46px_rgba(15,23,42,0.24)] md:w-[390px] md:max-w-none">
            <div className="flex items-start justify-between gap-4 border-b border-slate-100 pb-4">
              <div>
                <div className="text-[10px] font-black uppercase tracking-[0.22em] text-[#087e8b]">Tambah Downline</div>
                <h3 className="mt-1 text-xl font-black text-slate-950">{role === "marketing" ? "Buat Agent Binaan" : role === "agent" ? "Buat Member" : "Buat Akun Retail"}</h3>
                <p className="mt-1 text-xs font-semibold leading-5 text-slate-500">Akun baru akan langsung terhubung sebagai member anda.</p>
              </div>
              <button
                type="button"
                onClick={() => setShowCreateModal(false)}
                className="inline-flex h-11 w-11 items-center justify-center rounded-2xl border border-slate-200 text-slate-500 transition hover:bg-slate-50"
                aria-label="Tutup"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <div className="space-y-4 pt-5">
              {err ? <div className="rounded-2xl border border-rose-200 bg-rose-50 px-3 py-2 text-xs font-bold text-rose-700">{err}</div> : null}

              <form className="grid gap-3" onSubmit={submit}>
                {role === "master" ? (
                  <div className="grid gap-1.5">
                    <label className="text-xs font-black text-slate-700">Role</label>
                    <select
                      className="h-12 rounded-2xl border border-slate-200 px-3 text-sm font-bold text-slate-900 outline-none focus:border-emerald-500 focus:ring-4 focus:ring-emerald-100"
                      value={targetRole}
                      onChange={(e) => setTargetRole(e.target.value)}
                    >
                      {roleOptions.map((item) => (
                        <option key={item.value} value={item.value}>{item.label}</option>
                      ))}
                    </select>
                  </div>
                ) : null}
                <input className="h-12 rounded-2xl border border-slate-200 px-3 text-sm font-bold outline-none focus:border-emerald-500 focus:ring-4 focus:ring-emerald-100" placeholder="Nama" value={nama} onChange={(e) => setNama(e.target.value)} />
                <input className="h-12 rounded-2xl border border-slate-200 px-3 text-sm font-bold outline-none focus:border-emerald-500 focus:ring-4 focus:ring-emerald-100" placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} />
                <input className="h-12 rounded-2xl border border-slate-200 px-3 text-sm font-bold outline-none focus:border-emerald-500 focus:ring-4 focus:ring-emerald-100" placeholder="Password minimal 8 karakter" type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
                <button className="mt-1 inline-flex h-12 items-center justify-center gap-2 rounded-2xl bg-[#13464b] text-sm font-black text-white shadow-[0_12px_24px_rgba(6,78,59,0.18)] hover:bg-[#13464b] disabled:opacity-70" disabled={saving} type="submit">
                  {saving ? "Menyimpan..." : "Tambah Akun"}
                  {!saving ? <ChevronRight className="h-4 w-4" /> : null}
                </button>
              </form>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
