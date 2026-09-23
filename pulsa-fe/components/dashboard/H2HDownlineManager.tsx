"use client";

import { useEffect, useMemo, useState } from "react";
import { KeyRound, Network, ShieldCheck, UserPlus, Users } from "lucide-react";

type H2HDownline = {
  id: number;
  email: string;
  nama: string;
  role: string;
  aktif: boolean;
  saldo: number;
  h2h_agent_nama?: string | null;
  h2h_master_nama?: string | null;
};

type Props = {
  authToken: string;
};

function decodeRoleFromToken(token: string): string {
  try {
    const payload = token.split(".")[1] || "";
    const normalized = payload.replace(/-/g, "+").replace(/_/g, "/");
    const padded = normalized + "=".repeat((4 - (normalized.length % 4)) % 4);
    const parsed = JSON.parse(atob(padded));
    return typeof parsed?.role === "string" ? parsed.role : "";
  } catch {
    return "";
  }
}

function roleLabel(role: string) {
  switch (String(role || "").toLowerCase()) {
    case "master_member":
      return "Master Member";
    case "agent_member":
      return "Agent Member";
    case "member":
      return "Member";
    default:
      return role || "-";
  }
}

function fmtIDR(v: number) {
  return new Intl.NumberFormat("id-ID").format(Number(v || 0));
}

export function H2HDownlineManager({ authToken }: Props) {
  const role = useMemo(() => decodeRoleFromToken(authToken), [authToken]);
  const [items, setItems] = useState<H2HDownline[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [err, setErr] = useState("");
  const [ok, setOk] = useState("");
  const [email, setEmail] = useState("");
  const [nama, setNama] = useState("");
  const [password, setPassword] = useState("");
  const [pin, setPin] = useState("");
  const [targetRole, setTargetRole] = useState("member");
  const [feeDana, setFeeDana] = useState("");
  const [feeGopay, setFeeGopay] = useState("");
  const [feeLinkAja, setFeeLinkAja] = useState("");
  const [feeOvo, setFeeOvo] = useState("");
  const [feeShopee, setFeeShopee] = useState("");
  const [feeBank, setFeeBank] = useState("");
  const [feeLainnya, setFeeLainnya] = useState("");

  const canCreate = role === "master_member" || role === "agent_member";
  const roleOptions = useMemo(
    () =>
      role === "master_member"
        ? [
            { value: "member", label: "Member" },
            { value: "agent_member", label: "Agent Member" },
          ]
        : [{ value: "member", label: "Member" }],
    [role],
  );

  async function load() {
    setLoading(true);
    setErr("");
    try {
      const r = await fetch("/api/me/h2h/downlines", {
        headers: { Authorization: `Bearer ${authToken}` },
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) {
        setErr(j?.error || "Gagal memuat jaringan H2H.");
        setItems([]);
        return;
      }
      setItems(Array.isArray(j.items) ? j.items : []);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setErr("");
    setOk("");
    if (!email.trim() || !nama.trim() || password.trim().length < 8) {
      setErr("Email, nama, dan password minimal 8 karakter wajib diisi.");
      return;
    }
    if (pin.trim().length < 4 || pin.trim().length > 12) {
      setErr("PIN harus 4-12 karakter.");
      return;
    }
    const requiredFees = [
      { label: "Fee DANA", value: feeDana },
      { label: "Fee GOPAY", value: feeGopay },
      { label: "Fee LINKAJA", value: feeLinkAja },
      { label: "Fee OVO", value: feeOvo },
      { label: "Fee SHOPEEPAY", value: feeShopee },
      { label: "Fee BANK", value: feeBank },
      { label: "Fee Lainnya", value: feeLainnya },
    ];
    for (const item of requiredFees) {
      if (item.value.trim() === "") {
        setErr(`${item.label} wajib diisi.`);
        return;
      }
      const parsed = Number(item.value);
      if (!Number.isFinite(parsed) || parsed < 0) {
        setErr(`${item.label} harus berupa angka >= 0.`);
        return;
      }
    }

    setSaving(true);
    try {
      const r = await fetch("/api/me/h2h/downlines", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${authToken}`,
        },
        body: JSON.stringify({
          email: email.trim(),
          nama: nama.trim(),
          password: password.trim(),
          pin: pin.trim(),
          role: targetRole,
          fee_dana: Math.floor(Number(feeDana)),
          fee_gopay: Math.floor(Number(feeGopay)),
          fee_linkaja: Math.floor(Number(feeLinkAja)),
          fee_ovo: Math.floor(Number(feeOvo)),
          fee_shopee: Math.floor(Number(feeShopee)),
          fee_bank: Math.floor(Number(feeBank)),
          fee_lainnya: Math.floor(Number(feeLainnya)),
        }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) {
        setErr(j?.error || "Gagal membuat akun H2H.");
        return;
      }
      setOk(`Permintaan akun H2H berhasil dikirim. Member #${j.member_id} menunggu aktivasi admin sebelum bisa login atau transaksi.`);
      setEmail("");
      setNama("");
      setPassword("");
      setPin("");
      setTargetRole(role === "master_member" ? "member" : "member");
      setFeeDana("");
      setFeeGopay("");
      setFeeLinkAja("");
      setFeeOvo("");
      setFeeShopee("");
      setFeeBank("");
      setFeeLainnya("");
      await load();
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="space-y-4">
      <section className="rounded-2xl border border-white/10 bg-slate-900/60 p-5 shadow-[0_18px_44px_-28px_rgba(34,211,238,0.28)]">
        <div className="flex items-center gap-3">
          <span className="grid h-11 w-11 place-items-center rounded-2xl bg-cyan-500/15 text-cyan-300">
            <Network className="h-5 w-5" />
          </span>
          <div>
            <h1 className="text-lg font-bold text-white">Jaringan H2H</h1>
            <p className="text-sm text-white/60">
              {role === "master_member"
                ? "Master Member dapat menambahkan Agent Member atau Member."
                : role === "agent_member"
                  ? "Agent Member dapat menambahkan Member."
                  : "Daftar jaringan H2H yang terhubung ke akun ini."}
            </p>
          </div>
        </div>
      </section>

      {canCreate ? (
        <section className="rounded-2xl border border-white/10 bg-slate-900/60 p-5 shadow-[0_18px_44px_-28px_rgba(34,211,238,0.28)]">
          <div className="mb-4 flex items-center gap-3">
            <span className="grid h-10 w-10 place-items-center rounded-2xl bg-emerald-500/15 text-emerald-300">
              <UserPlus className="h-5 w-5" />
            </span>
            <div>
              <h2 className="font-bold text-white">Tambah Akun H2H</h2>
              <p className="text-sm text-white/60">Akun baru masuk sebagai downline H2H pending dan harus diaktifkan admin lebih dulu.</p>
            </div>
          </div>

          {err ? <div className="mb-3 rounded-2xl border border-rose-500/30 bg-rose-500/10 px-3 py-2 text-sm text-rose-200">{err}</div> : null}
          {ok ? <div className="mb-3 rounded-2xl border border-emerald-500/30 bg-emerald-500/10 px-3 py-2 text-sm text-emerald-200">{ok}</div> : null}

          <form className="grid gap-3 md:grid-cols-2" onSubmit={submit}>
            <div className="grid gap-1.5 md:col-span-2">
              <label className="text-sm font-medium text-slate-200">Role</label>
              <select
                className="h-11 rounded-xl border border-white/10 bg-slate-950/60 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400"
                value={targetRole}
                onChange={(e) => setTargetRole(e.target.value)}
              >
                {roleOptions.map((item) => (
                  <option key={item.value} value={item.value}>{item.label}</option>
                ))}
              </select>
            </div>
            <input className="h-11 rounded-xl border border-white/10 bg-slate-950/60 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400" placeholder="Nama" value={nama} onChange={(e) => setNama(e.target.value)} />
            <input className="h-11 rounded-xl border border-white/10 bg-slate-950/60 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400" placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} />
            <div className="relative">
              <KeyRound className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
              <input className="h-11 w-full rounded-xl border border-white/10 bg-slate-950/60 pl-9 pr-3 text-sm text-slate-100 outline-none focus:border-cyan-400" placeholder="Password minimal 8 karakter" type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
            </div>
            <div className="relative">
              <ShieldCheck className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
              <input className="h-11 w-full rounded-xl border border-white/10 bg-slate-950/60 pl-9 pr-3 text-sm text-slate-100 outline-none focus:border-cyan-400" placeholder="PIN 4-12 karakter" value={pin} onChange={(e) => setPin(e.target.value)} />
            </div>
            <input className="h-11 rounded-xl border border-white/10 bg-slate-950/60 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400" placeholder="Fee DANA" inputMode="numeric" value={feeDana} onChange={(e) => setFeeDana(e.target.value.replace(/[^\d]/g, ""))} />
            <input className="h-11 rounded-xl border border-white/10 bg-slate-950/60 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400" placeholder="Fee GOPAY" inputMode="numeric" value={feeGopay} onChange={(e) => setFeeGopay(e.target.value.replace(/[^\d]/g, ""))} />
            <input className="h-11 rounded-xl border border-white/10 bg-slate-950/60 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400" placeholder="Fee LINKAJA" inputMode="numeric" value={feeLinkAja} onChange={(e) => setFeeLinkAja(e.target.value.replace(/[^\d]/g, ""))} />
            <input className="h-11 rounded-xl border border-white/10 bg-slate-950/60 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400" placeholder="Fee OVO" inputMode="numeric" value={feeOvo} onChange={(e) => setFeeOvo(e.target.value.replace(/[^\d]/g, ""))} />
            <input className="h-11 rounded-xl border border-white/10 bg-slate-950/60 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400" placeholder="Fee SHOPEEPAY" inputMode="numeric" value={feeShopee} onChange={(e) => setFeeShopee(e.target.value.replace(/[^\d]/g, ""))} />
            <input className="h-11 rounded-xl border border-white/10 bg-slate-950/60 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400" placeholder="Fee BANK" inputMode="numeric" value={feeBank} onChange={(e) => setFeeBank(e.target.value.replace(/[^\d]/g, ""))} />
            <input className="h-11 rounded-xl border border-white/10 bg-slate-950/60 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400" placeholder="Fee Lainnya" inputMode="numeric" value={feeLainnya} onChange={(e) => setFeeLainnya(e.target.value.replace(/[^\d]/g, ""))} />
            <button className="h-11 rounded-xl bg-cyan-500 text-sm font-semibold text-slate-950 hover:bg-cyan-400 disabled:opacity-70 md:col-span-2" disabled={saving} type="submit">
              {saving ? "Menyimpan..." : "Tambah Akun H2H"}
            </button>
          </form>
        </section>
      ) : null}

      <section className="rounded-2xl border border-white/10 bg-slate-900/60 p-5 shadow-[0_18px_44px_-28px_rgba(34,211,238,0.28)]">
        <div className="mb-4 flex items-center gap-3">
            <span className="grid h-10 w-10 place-items-center rounded-2xl bg-sky-500/15 text-sky-300">
              <Users className="h-5 w-5" />
            </span>
            <div>
              <h2 className="font-bold text-white">Daftar Agent/Member</h2>
            <p className="text-sm text-white/60">Downline H2H yang terhubung ke akun ini.</p>
            </div>
          </div>

        <div className="space-y-3">
          {loading ? <div className="text-sm text-white/60">Memuat jaringan H2H...</div> : null}
          {!loading && items.length === 0 ? <div className="text-sm text-white/60">Belum ada agent atau member di bawah akun ini.</div> : null}
          {items.map((item) => (
            <div key={item.id} className="rounded-2xl border border-white/10 bg-slate-950/40 px-4 py-3">
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <div className="truncate text-sm font-semibold text-white">{item.nama || item.email}</div>
                  <div className="truncate text-xs text-white/55">{item.email}</div>
                  <div className="mt-1 text-xs text-white/50">
                    {roleLabel(item.role)}
                    {item.h2h_agent_nama ? ` • Agent: ${item.h2h_agent_nama}` : ""}
                    {item.h2h_master_nama ? ` • Master: ${item.h2h_master_nama}` : ""}
                  </div>
                </div>
                <div className="text-right">
                  <div className="text-xs uppercase tracking-wide text-white/45">{item.aktif ? "Aktif" : "Pending Admin"}</div>
                  <div className="mt-1 text-sm font-semibold text-cyan-300">Rp {fmtIDR(item.saldo)}</div>
                </div>
              </div>
              {!item.aktif ? (
                <div className="mt-3 rounded-xl border border-amber-400/20 bg-amber-500/10 px-3 py-2 text-xs text-amber-100">
                  Akun ini belum aktif. Admin harus memeriksa fee H2H dan menyetujui aktivasinya terlebih dulu.
                </div>
              ) : null}
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}
