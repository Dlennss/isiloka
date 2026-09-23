"use client";

import { useEffect, useMemo, useState } from "react";
import { Eye, EyeOff, KeyRound, Mail, ShieldCheck, ToggleLeft, Trash2, UserRound } from "lucide-react";
import { AppModal } from "@/components/ui/app-modal";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { alertConfirm, alertError, alertSuccess, alertWarning } from "@/components/ui/alerts";
import { ALL_MEMBER_ROLES, roleLabel, type AccountScope, type MemberRole } from "@/lib/memberRoles";

type UserRow = {
  id: number;
  email: string;
  nama: string;
  role: string;
  aktif: boolean;
  fee_member_rp: number;
  retail_agent_commission_rp: number;
  retail_master_commission_rp: number;
  h2h_agent_commission_rp: number;
  h2h_master_commission_rp: number;
};

type Props = {
  open: boolean;
  user: UserRow | null;
  onClose: () => void;
  scope?: AccountScope;
  allowedRoles?: MemberRole[];
  onSuccess?: () => Promise<void> | void;
};

type RoleOption = MemberRole;

function isH2HRole(role: string) {
  return role === "member" || role === "agent_member" || role === "master_member";
}

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export default function EditUserModal({ open, user, onClose, onSuccess, allowedRoles: allowedRolesProp }: Props) {
  const allowedRoles = useMemo(
    () => allowedRolesProp ?? ALL_MEMBER_ROLES,
    [allowedRolesProp],
  );
  const [email, setEmail] = useState("");
  const [nama, setNama] = useState("");
  const [role, setRole] = useState<RoleOption>("member");
  const [aktif, setAktif] = useState(true);
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [newPIN, setNewPIN] = useState("");
  const [confirmPIN, setConfirmPIN] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [showPIN, setShowPIN] = useState(false);
  const [showConfirmPIN, setShowConfirmPIN] = useState(false);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const isMember = useMemo(() => isH2HRole(role), [role]);
  const isConvertingToH2H = Boolean(user && isMember && !isH2HRole(user.role));

  useEffect(() => {
    if (!open || !user) return;
    setEmail(user.email || "");
    setNama(user.nama || "");
    setRole((user.role || "member") as RoleOption);
    setAktif(Boolean(user.aktif));
    setNewPassword("");
    setConfirmPassword("");
    setNewPIN("");
    setConfirmPIN("");
    setShowPassword(false);
    setShowConfirmPassword(false);
    setShowPIN(false);
    setShowConfirmPIN(false);
    setDeleting(false);
  }, [open, user]);

  if (!open || !user) return null;

  async function submit() {
    if (!user) return;
    const normalizedEmail = email.trim().toLowerCase();
    if (!normalizedEmail) return alertWarning("Email wajib diisi.");
    if (!EMAIL_RE.test(normalizedEmail)) return alertWarning("Format email tidak valid.");
    if (!nama.trim()) return alertWarning("Nama wajib diisi.");
    if (newPassword.trim() && newPassword.trim().length < 8) return alertWarning("Password minimal 8 karakter.");
    if (newPassword.trim() && newPassword !== confirmPassword) return alertWarning("Konfirmasi password baru tidak sama.");
    if (isMember && newPIN.trim() && !/^\d{4,12}$/.test(newPIN.trim())) return alertWarning("PIN harus 4-12 digit angka.");
    if (isMember && newPIN.trim() && newPIN !== confirmPIN) return alertWarning("Konfirmasi PIN baru tidak sama.");
    if (isConvertingToH2H && !newPIN.trim()) return alertWarning("PIN baru wajib diisi saat mengubah akun menjadi H2H.");

    setSaving(true);
    try {
      const r = await fetch(`/api/admin/members/${user.id}`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          ...authHeader(),
        },
        body: JSON.stringify({
          email: normalizedEmail,
          nama: nama.trim(),
          role,
          aktif,
          fee_member_rp: Number(user.fee_member_rp ?? 0),
          retail_agent_commission_rp: Number(user.retail_agent_commission_rp ?? 0),
          retail_master_commission_rp: Number(user.retail_master_commission_rp ?? 0),
          h2h_agent_commission_rp: Number(user.h2h_agent_commission_rp ?? 0),
          h2h_master_commission_rp: Number(user.h2h_master_commission_rp ?? 0),
          new_password: newPassword.trim() || undefined,
          new_pin: isMember ? newPIN.trim() || undefined : undefined,
        }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal update pengguna");
        return;
      }
      await alertSuccess("Profil pengguna berhasil diupdate.");
      if (onSuccess) await onSuccess();
      onClose();
    } finally {
      setSaving(false);
    }
  }

  async function removeAccount() {
    if (!user || saving || deleting) return;
    const confirmed = await alertConfirm({
      title: "Hapus akun permanen?",
      text: `Akun ${user.nama || user.email} dan data yang terkait akan dihapus dari database. Tindakan ini tidak dapat dibatalkan.`,
      confirmButtonText: "Ya, hapus akun",
      cancelButtonText: "Batal",
    });
    if (!confirmed) return;

    setDeleting(true);
    try {
      const r = await fetch(`/api/admin/members/${user.id}`, {
        method: "DELETE",
        headers: authHeader(),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal menghapus akun");
        return;
      }
      await alertSuccess("Akun berhasil dihapus dari database.");
      if (onSuccess) await onSuccess();
      onClose();
    } finally {
      setDeleting(false);
    }
  }

  return (
    <AppModal
      open={open}
      onClose={onClose}
      title="Update Pengguna"
      subtitle={`Edit profil pengguna #${user.id}`}
      maxWidthClassName="max-w-2xl"
      footer={
        <div className="flex flex-col-reverse gap-2 sm:flex-row sm:items-center sm:justify-between">
          <Button
            type="button"
            variant="danger"
            className="h-11 justify-center px-4"
            onClick={removeAccount}
            disabled={saving || deleting}
          >
            <Trash2 className="mr-2 h-4 w-4" />
            {deleting ? "Menghapus..." : "Hapus Akun"}
          </Button>
          <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={onClose} disabled={saving || deleting}>
            Tutup
          </Button>
          <Button
            variant="success"
            className="h-11 px-5"
            onClick={submit}
            disabled={saving || deleting}
          >
            {saving ? "Saving..." : "Simpan Perubahan"}
          </Button>
          </div>
        </div>
      }
    >
      <div className="grid gap-4 md:grid-cols-2">
        <label className="grid gap-2 md:col-span-2">
          <span className="text-sm text-slate-200">Email</span>
          <div className="relative">
            <Mail className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input className="h-11 pl-9" value={email} onChange={(e) => setEmail(e.target.value)} placeholder="email@domain.com" />
          </div>
        </label>

        <label className="grid gap-2">
          <span className="text-sm text-slate-200">Nama</span>
          <div className="relative">
            <UserRound className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input className="h-11 pl-9" value={nama} onChange={(e) => setNama(e.target.value)} placeholder="Nama pengguna" />
          </div>
        </label>

        <label className="grid gap-2">
          <span className="text-sm text-slate-200">Role</span>
          <select
            value={role}
            onChange={(e) => setRole(e.target.value as RoleOption)}
            className="h-11 rounded-lg border border-white/15 bg-slate-950/70 px-3 text-sm text-slate-100 outline-hidden ring-0 transition focus:border-cyan-400/60"
          >
            {allowedRoles.map((roleValue) => (
              <option key={roleValue} value={roleValue}>
                {roleLabel(roleValue)}
              </option>
            ))}
          </select>
        </label>

        {!isMember ? (
          <label className="grid gap-2">
            <span className="text-sm text-slate-200">Status Aktif</span>
            <div className="relative">
              <ToggleLeft className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <select
                value={aktif ? "true" : "false"}
                onChange={(e) => setAktif(e.target.value === "true")}
                className="h-11 w-full rounded-lg border border-white/15 bg-slate-950/70 px-9 text-sm text-slate-100 outline-hidden ring-0 transition focus:border-cyan-400/60"
              >
                <option value="true">Aktif</option>
                <option value="false">Nonaktif</option>
              </select>
            </div>
          </label>
        ) : null}

        <div className="grid gap-4 md:col-span-2 md:grid-cols-2">
          <label className="grid gap-2">
            <span className="text-sm text-slate-200">Password Baru</span>
            <div className="relative">
              <KeyRound className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                className="h-11 pl-9 pr-11"
                type={showPassword ? "text" : "password"}
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                placeholder="Kosongkan jika tidak diubah"
              />
              <button
                type="button"
                className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-200"
                onClick={() => setShowPassword((v) => !v)}
                aria-label={showPassword ? "Hide password" : "Show password"}
              >
                {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </button>
            </div>
          </label>

          <label className="grid gap-2">
            <span className="text-sm text-slate-200">Konfirmasi Password Baru</span>
            <div className="relative">
              <KeyRound className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                className="h-11 pl-9 pr-11"
                type={showConfirmPassword ? "text" : "password"}
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder="Ulangi password baru"
              />
              <button
                type="button"
                className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-200"
                onClick={() => setShowConfirmPassword((v) => !v)}
                aria-label={showConfirmPassword ? "Hide password confirmation" : "Show password confirmation"}
              >
                {showConfirmPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </button>
            </div>
          </label>
        </div>

        {isMember ? (
          <>
            <div className="grid gap-4 md:col-span-2 md:grid-cols-2">
              <label className="grid gap-2">
                <span className="text-sm text-slate-200">PIN Baru</span>
                <div className="relative">
                  <ShieldCheck className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    className="h-11 pl-9 pr-11"
                    type={showPIN ? "text" : "password"}
                    value={newPIN}
                    onChange={(e) => setNewPIN(e.target.value)}
                    placeholder="Kosongkan jika tidak diubah"
                  />
                  <button
                    type="button"
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-200"
                    onClick={() => setShowPIN((v) => !v)}
                    aria-label={showPIN ? "Hide PIN" : "Show PIN"}
                  >
                    {showPIN ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </button>
                </div>
              </label>

              <label className="grid gap-2">
                <span className="text-sm text-slate-200">Konfirmasi PIN Baru</span>
                <div className="relative">
                  <ShieldCheck className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    className="h-11 pl-9 pr-11"
                    type={showConfirmPIN ? "text" : "password"}
                    value={confirmPIN}
                    onChange={(e) => setConfirmPIN(e.target.value)}
                    placeholder="Ulangi PIN baru"
                  />
                  <button
                    type="button"
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-200"
                    onClick={() => setShowConfirmPIN((v) => !v)}
                    aria-label={showConfirmPIN ? "Hide PIN confirmation" : "Show PIN confirmation"}
                  >
                    {showConfirmPIN ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </button>
                </div>
              </label>
            </div>
            <div className="grid gap-4 md:col-span-2 md:grid-cols-2">
              <label className="grid gap-2">
                <span className="text-sm text-slate-200">Status Aktif</span>
                <div className="relative">
                  <ToggleLeft className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <select
                    value={aktif ? "true" : "false"}
                    onChange={(e) => setAktif(e.target.value === "true")}
                    className="h-11 w-full rounded-lg border border-white/15 bg-slate-950/70 px-9 text-sm text-slate-100 outline-hidden ring-0 transition focus:border-cyan-400/60"
                  >
                    <option value="true">Aktif</option>
                    <option value="false">Nonaktif</option>
                  </select>
                </div>
              </label>

              <div className="rounded-xl border border-amber-400/25 bg-amber-500/10 px-3 py-3 text-sm text-amber-100">
                {isConvertingToH2H
                  ? "Isi PIN baru untuk mengaktifkan akses H2H. Fee transaksi H2H bisa diatur lewat menu Fee Kategori setelah akun tersimpan."
                  : "Akun H2H tidak memakai fee flat. Fee transaksi H2H wajib diatur per produk."}
              </div>
            </div>
          </>
        ) : null}
      </div>
    </AppModal>
  );
}
