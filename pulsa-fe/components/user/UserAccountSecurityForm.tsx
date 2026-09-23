"use client";

import Link from "next/link";
import { useState, type FormEvent } from "react";
import { ArrowLeft, CheckCircle2, Eye, EyeOff, KeyRound, LockKeyhole, Save, ShieldCheck } from "lucide-react";

type Props = {
  authToken: string;
};

type PasswordResponse = {
  ok?: boolean;
  error?: string;
};

function errorMessage(value?: string) {
  const message = String(value || "").trim();
  if (message === "old_password salah") return "Password lama tidak sesuai.";
  if (message.includes("minimal 6")) return "Password baru terlalu pendek.";
  if (message.includes("required")) return "Password lama dan password baru wajib diisi.";
  return message || "Password belum dapat diperbarui.";
}

function PasswordField({
  id,
  label,
  value,
  onChange,
  visible,
  onToggle,
  autoComplete,
  disabled,
}: {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
  visible: boolean;
  onToggle: () => void;
  autoComplete: string;
  disabled: boolean;
}) {
  return (
    <label htmlFor={id} className="block">
      <span className="mb-2 block text-[10px] font-black uppercase tracking-[0.16em] text-emerald-800">{label}</span>
      <span className="flex h-13 items-center rounded-2xl border border-slate-200 bg-slate-50 px-4 transition focus-within:border-emerald-500 focus-within:bg-white focus-within:ring-4 focus-within:ring-emerald-100">
        <KeyRound className="h-4 w-4 shrink-0 text-slate-400" />
        <input
          id={id}
          type={visible ? "text" : "password"}
          value={value}
          onChange={(event) => onChange(event.target.value)}
          autoComplete={autoComplete}
          disabled={disabled}
          className="min-w-0 flex-1 bg-transparent px-3 text-sm font-bold text-slate-900 outline-none placeholder:text-slate-400"
          placeholder="Masukkan password"
        />
        <button type="button" onClick={onToggle} disabled={disabled} className="grid h-9 w-9 shrink-0 place-items-center rounded-xl text-slate-500 hover:bg-emerald-50 hover:text-emerald-700" aria-label={visible ? `Sembunyikan ${label.toLowerCase()}` : `Tampilkan ${label.toLowerCase()}`}>
          {visible ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
        </button>
      </span>
    </label>
  );
}

export function UserAccountSecurityForm({ authToken }: Props) {
  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [visible, setVisible] = useState({ old: false, next: false, confirm: false });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (saving) return;
    setError("");
    setSuccess("");

    if (!oldPassword || !newPassword || !confirmPassword) {
      setError("Semua kolom password wajib diisi.");
      return;
    }
    if (newPassword.length < 8) {
      setError("Password baru minimal 8 karakter.");
      return;
    }
    if (newPassword !== confirmPassword) {
      setError("Konfirmasi password baru tidak sama.");
      return;
    }
    if (oldPassword === newPassword) {
      setError("Password baru harus berbeda dari password lama.");
      return;
    }

    setSaving(true);
    try {
      const response = await fetch("/api/me/password", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${authToken}`,
        },
        body: JSON.stringify({ old_password: oldPassword, new_password: newPassword }),
      });
      const data = (await response.json().catch(() => ({}))) as PasswordResponse;
      if (!response.ok || !data.ok) {
        setError(errorMessage(data.error));
        return;
      }

      setOldPassword("");
      setNewPassword("");
      setConfirmPassword("");
      setSuccess("Password berhasil diperbarui.");
    } catch {
      setError("Koneksi terputus. Silakan coba kembali.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="space-y-4">
      <section className="overflow-hidden rounded-[28px] bg-[linear-gradient(135deg,#13464b,#047857_62%,#4dcf38)] p-5 text-white shadow-[0_18px_40px_rgba(6,78,59,0.18)]">
        <div className="flex items-center gap-3">
          <Link href="/user/account" className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-white/12 ring-1 ring-white/20" aria-label="Kembali ke akun">
            <ArrowLeft className="h-5 w-5" />
          </Link>
          <span className="grid h-12 w-12 shrink-0 place-items-center rounded-2xl bg-white text-emerald-700">
            <ShieldCheck className="h-6 w-6" />
          </span>
          <div className="min-w-0">
            <p className="text-[10px] font-black uppercase tracking-[0.2em] text-lime-100">Keamanan Akun</p>
            <h1 className="mt-1 text-2xl font-black">Ganti Password</h1>
          </div>
        </div>
      </section>

      <form onSubmit={submit} className="overflow-hidden rounded-[28px] border border-emerald-100 bg-white shadow-[0_16px_38px_rgba(6,78,59,0.08)]">
        <div className="flex items-center gap-3 border-b border-slate-100 px-5 py-4">
          <span className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-emerald-50 text-emerald-700">
            <LockKeyhole className="h-5 w-5" />
          </span>
          <div>
            <h2 className="text-base font-black text-slate-950">Password Login</h2>
            <p className="mt-0.5 text-xs font-semibold text-slate-500">Perbarui password akun Isiloka.</p>
          </div>
        </div>

        <div className="space-y-4 p-5">
          <PasswordField id="old-password" label="Password Lama" value={oldPassword} onChange={setOldPassword} visible={visible.old} onToggle={() => setVisible((current) => ({ ...current, old: !current.old }))} autoComplete="current-password" disabled={saving} />
          <PasswordField id="new-password" label="Password Baru" value={newPassword} onChange={setNewPassword} visible={visible.next} onToggle={() => setVisible((current) => ({ ...current, next: !current.next }))} autoComplete="new-password" disabled={saving} />
          <PasswordField id="confirm-password" label="Konfirmasi Password Baru" value={confirmPassword} onChange={setConfirmPassword} visible={visible.confirm} onToggle={() => setVisible((current) => ({ ...current, confirm: !current.confirm }))} autoComplete="new-password" disabled={saving} />

          {error ? <div role="alert" className="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-xs font-bold text-rose-700">{error}</div> : null}
          {success ? <div className="flex items-center gap-2 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-xs font-bold text-emerald-700"><CheckCircle2 className="h-4 w-4 shrink-0" />{success}</div> : null}

          <button type="submit" disabled={saving} className="inline-flex h-13 w-full items-center justify-center gap-2 rounded-2xl bg-[linear-gradient(90deg,#087e8b,#16a34a)] px-5 text-sm font-black text-white shadow-[0_12px_24px_rgba(5,150,105,0.22)] disabled:cursor-not-allowed disabled:opacity-60">
            <Save className="h-4 w-4" />
            {saving ? "Menyimpan..." : "Simpan Password Baru"}
          </button>
        </div>
      </form>
    </div>
  );
}
