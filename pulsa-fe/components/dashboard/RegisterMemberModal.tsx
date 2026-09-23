"use client";

import { useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { alertError, alertSuccess, alertWarning } from "@/components/ui/alerts";
import { Input } from "@/components/ui/input";
import { AppModal } from "@/components/ui/app-modal";
import { KeyRound, Mail, Phone, ShieldCheck, UserRound } from "lucide-react";
import { createRolesForScope, type AccountScope, type ManageableRole } from "@/lib/memberRoles";

type CreateResp = {
  ok?: boolean;
  error?: string;
  member_id?: number;
  api_key?: string;
};

type RegisterSuccessPayload = {
  member: {
    id: number;
    email: string;
    nama: string;
    role: ManageableRole;
  };
};

type RegisterMemberModalProps = {
  open: boolean;
  onClose: () => void;
  scope?: AccountScope;
  allowedRoles?: ManageableRole[];
  fixedRole?: ManageableRole;
  title?: string;
  subtitle?: string;
  theme?: "dark" | "retail";
  createEndpoint?: string;
  onSuccess?: (payload: RegisterSuccessPayload) => Promise<void> | void;
};

type RoleOption = "" | ManageableRole;

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

function isH2HRole(role: RoleOption) {
  return role === "member" || role === "agent_member" || role === "master_member";
}

export default function RegisterMemberModal(props: RegisterMemberModalProps) {
  const { open, onClose, onSuccess } = props;
  const fixedRole = props.fixedRole;
  const scope = props.scope ?? "h2h";
  const allowedRoles = useMemo(
    () => props.allowedRoles ?? createRolesForScope(scope),
    [props.allowedRoles, scope],
  );

  const [role, setRole] = useState<RoleOption>(fixedRole ?? "");
  const [email, setEmail] = useState("");
  const [nama, setNama] = useState("");
  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const [pin, setPin] = useState("");
  const [feeDana, setFeeDana] = useState("");
  const [feeGopay, setFeeGopay] = useState("");
  const [feeLinkAja, setFeeLinkAja] = useState("");
  const [feeOvo, setFeeOvo] = useState("");
  const [feeShopee, setFeeShopee] = useState("");
  const [feeBank, setFeeBank] = useState("");
  const [feeLainnya, setFeeLainnya] = useState("");
  const [loading, setLoading] = useState(false);

  const isMember = useMemo(() => isH2HRole(role), [role]);
  const roleChosen = role !== "";

  if (!open) return null;

  function resetForm() {
    setRole(fixedRole ?? "");
    setEmail("");
    setNama("");
    setPhone("");
    setPassword("");
    setPin("");
    setFeeDana("");
    setFeeGopay("");
    setFeeLinkAja("");
    setFeeOvo("");
    setFeeShopee("");
    setFeeBank("");
    setFeeLainnya("");
  }

  async function submit() {
    if (!roleChosen) return alertWarning("Role wajib dipilih.");
    if (!email.trim()) return alertWarning("Email wajib diisi.");
    if (!nama.trim()) return alertWarning("Nama wajib diisi.");
    if (role === "marketing" && !phone.trim()) return alertWarning("Nomor telepon marketing wajib diisi.");
    if (password.trim().length < 8) return alertWarning("Password minimal 8 karakter.");
    if (isMember && (pin.trim().length < 4 || pin.trim().length > 12)) return alertWarning("PIN harus 4-12 karakter.");
    if (isMember) {
      const requiredFees = [
        { label: "Fee DANA", value: feeDana },
        { label: "Fee GOPAY", value: feeGopay },
        { label: "Fee LINKAJA", value: feeLinkAja },
        { label: "Fee OVO", value: feeOvo },
        { label: "Fee SHOPEE", value: feeShopee },
        { label: "Fee BANK", value: feeBank },
        { label: "Fee lainnya", value: feeLainnya },
      ];
      for (const item of requiredFees) {
        if (item.value.trim() === "") {
          return alertWarning(`${item.label} wajib diisi.`);
        }
        const parsed = Number(item.value);
        if (!Number.isFinite(parsed) || parsed < 0) {
          return alertWarning(`${item.label} harus berupa angka >= 0.`);
        }
      }
    }

    setLoading(true);
    try {
      const r = await fetch(props.createEndpoint ?? "/api/admin/members/create", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({
          email: email.trim(),
          nama: nama.trim(),
          phone: phone.trim(),
          password: password.trim(),
          pin: isMember ? pin.trim() : undefined,
          role,
          fee_dana: isMember ? Math.floor(Number(feeDana)) : undefined,
          fee_gopay: isMember ? Math.floor(Number(feeGopay)) : undefined,
          fee_linkaja: isMember ? Math.floor(Number(feeLinkAja)) : undefined,
          fee_ovo: isMember ? Math.floor(Number(feeOvo)) : undefined,
          fee_shopee: isMember ? Math.floor(Number(feeShopee)) : undefined,
          fee_bank: isMember ? Math.floor(Number(feeBank)) : undefined,
          fee_lainnya: isMember ? Math.floor(Number(feeLainnya)) : undefined,
        }),
      });

      const j: CreateResp = await r.json().catch(() => ({} as CreateResp));
      if (!r.ok || !j.ok || !j.member_id) {
        await alertError(j.error || `Gagal register (HTTP ${r.status})`);
        return;
      }

      const successText = isMember
        ? `Akun H2H berhasil dibuat. member_id=${j.member_id}, api_key=${j.api_key || "-"}. Fee kategori awal sudah terpasang.`
        : role === "operator_wallet"
          ? `Operator wallet berhasil dibuat. member_id=${j.member_id}`
          : role === "operator_trx"
            ? `Operator transaksi berhasil dibuat. member_id=${j.member_id}`
            : `Akun retail berhasil dibuat. member_id=${j.member_id}`;
      await alertSuccess(successText);
      if (onSuccess) {
        await onSuccess({
          member: {
            id: j.member_id,
            email: email.trim(),
            nama: nama.trim(),
            role,
          },
        });
      }
      resetForm();
      onClose();
    } finally {
      setLoading(false);
    }
  }

  return (
    <AppModal
      open={open}
      onClose={() => {
        resetForm();
        onClose();
      }}
      title={props.title ?? "Tambah Pengguna"}
      subtitle={props.subtitle ?? "Pilih role lalu isi data yang diperlukan."}
      theme={props.theme}
      maxWidthClassName="max-w-2xl"
      footer={
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => { resetForm(); onClose(); }} disabled={loading}>
            Tutup
          </Button>
          <Button
            variant="success"
            className="h-11 px-5"
            onClick={submit}
            disabled={loading}
          >
            {loading ? "Saving..." : isMember ? "Simpan Akun H2H" : role === "operator_wallet" ? "Simpan Operator Wallet" : role === "operator_trx" ? "Simpan Operator" : role === "marketing" ? "Simpan Akun Marketing" : "Simpan Akun Retail"}
          </Button>
        </div>
      }
    >
      <div className="grid gap-4 md:grid-cols-2">
        {!fixedRole ? <label className="grid gap-2 md:col-span-2">
          <span className={`text-sm ${props.theme === "retail" ? "text-slate-700" : "text-slate-200"}`}>Role</span>
          <select
            value={role}
            onChange={(e) => setRole(e.target.value as RoleOption)}
            className="h-11 rounded-lg border border-white/15 bg-slate-950/70 px-3 text-sm text-slate-100 outline-hidden ring-0 transition focus:border-cyan-400/60"
          >
            <option value="" disabled>Pilih role</option>
            {allowedRoles.map((roleValue) => (
              <option key={roleValue} value={roleValue}>
                {roleValue === "member"
                  ? "Member"
                  : roleValue === "agent_member"
                    ? "Agent Member"
                    : roleValue === "master_member"
                      ? "Master Member"
                      : roleValue === "user"
                        ? "User"
                        : roleValue === "agent"
                          ? "Agent"
                          : roleValue === "master"
                            ? "Master"
                            : roleValue === "operator_trx"
                              ? "Operator Transaksi"
                              : roleValue === "operator_wallet"
                                ? "Operator Wallet"
                                : roleValue === "marketing"
                                  ? "Marketing"
                                  : roleValue === "analis"
                                    ? "Operator Kredit"
                                    : roleValue}
              </option>
            ))}
          </select>
        </label> : null}

        {roleChosen ? (
          <>
            <label className="grid gap-2">
              <span className={`text-sm ${props.theme === "retail" ? "text-slate-700" : "text-slate-200"}`}>Email</span>
              <div className="relative">
                <Mail className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input className="h-11 pl-9" value={email} onChange={(e) => setEmail(e.target.value)} placeholder="email@domain.com" />
              </div>
            </label>

            <label className="grid gap-2">
              <span className={`text-sm ${props.theme === "retail" ? "text-slate-700" : "text-slate-200"}`}>Nama</span>
              <div className="relative">
                <UserRound className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  className="h-11 pl-9"
                  value={nama}
                  onChange={(e) => setNama(e.target.value)}
                  placeholder={isMember ? "Nama akun H2H" : role === "operator_wallet" ? "Nama operator wallet" : role === "operator_trx" ? "Nama operator" : "Nama akun retail"}
                />
              </div>
            </label>

            {role === "marketing" ? (
              <label className="grid gap-2 md:col-span-2">
                <span className={`text-sm ${props.theme === "retail" ? "text-slate-700" : "text-slate-200"}`}>Nomor Telepon</span>
                <div className="relative">
                  <Phone className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <Input className="h-11 pl-9" value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="08xxxxxxxxxx" inputMode="tel" />
                </div>
              </label>
            ) : null}

            <label className="grid gap-2 md:col-span-2">
              <span className={`text-sm ${props.theme === "retail" ? "text-slate-700" : "text-slate-200"}`}>Password</span>
              <div className="relative">
                <KeyRound className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input className="h-11 pl-9" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="min 8 char" type="password" />
              </div>
            </label>

            {isMember ? (
              <>
                <label className="grid gap-2">
                  <span className={`text-sm ${props.theme === "retail" ? "text-slate-700" : "text-slate-200"}`}>PIN</span>
                  <div className="relative">
                    <ShieldCheck className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                    <Input className="h-11 pl-9" value={pin} onChange={(e) => setPin(e.target.value)} placeholder="4-12 char" />
                  </div>
                </label>

                <label className="grid gap-2">
                  <span className={`text-sm ${props.theme === "retail" ? "text-slate-700" : "text-slate-200"}`}>Fee DANA</span>
                  <Input className="h-11" value={feeDana} onChange={(e) => setFeeDana(e.target.value)} inputMode="numeric" />
                </label>
                <label className="grid gap-2">
                  <span className={`text-sm ${props.theme === "retail" ? "text-slate-700" : "text-slate-200"}`}>Fee GOPAY</span>
                  <Input className="h-11" value={feeGopay} onChange={(e) => setFeeGopay(e.target.value)} inputMode="numeric" />
                </label>
                <label className="grid gap-2">
                  <span className={`text-sm ${props.theme === "retail" ? "text-slate-700" : "text-slate-200"}`}>Fee LINKAJA</span>
                  <Input className="h-11" value={feeLinkAja} onChange={(e) => setFeeLinkAja(e.target.value)} inputMode="numeric" />
                </label>
                <label className="grid gap-2">
                  <span className={`text-sm ${props.theme === "retail" ? "text-slate-700" : "text-slate-200"}`}>Fee OVO</span>
                  <Input className="h-11" value={feeOvo} onChange={(e) => setFeeOvo(e.target.value)} inputMode="numeric" />
                </label>
                <label className="grid gap-2">
                  <span className={`text-sm ${props.theme === "retail" ? "text-slate-700" : "text-slate-200"}`}>Fee SHOPEE</span>
                  <Input className="h-11" value={feeShopee} onChange={(e) => setFeeShopee(e.target.value)} inputMode="numeric" />
                </label>
                <label className="grid gap-2">
                  <span className={`text-sm ${props.theme === "retail" ? "text-slate-700" : "text-slate-200"}`}>Fee BANK</span>
                  <Input className="h-11" value={feeBank} onChange={(e) => setFeeBank(e.target.value)} inputMode="numeric" />
                </label>
                <label className="grid gap-2">
                  <span className={`text-sm ${props.theme === "retail" ? "text-slate-700" : "text-slate-200"}`}>Fee Lainnya</span>
                  <Input className="h-11" value={feeLainnya} onChange={(e) => setFeeLainnya(e.target.value)} inputMode="numeric" />
                </label>

                <div className="rounded-xl border border-amber-400/25 bg-amber-500/10 px-3 py-3 text-sm text-amber-100 md:col-span-2">
                  Semua field fee H2H wajib diisi sebelum akun dibuat. Sistem akan langsung memasang fee kategori H2H untuk DANA, GOPAY, LINKAJA, OVO, SHOPEEPAY, BANK, dan LAINNYA.
                </div>
              </>
            ) : null}
          </>
        ) : null}
      </div>
    </AppModal>
  );
}
