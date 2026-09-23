"use client";

import { BadgeCheck, Shield, UserRound } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

type ProfileResponse = {
  ok?: boolean;
  profile?: {
    id?: number;
    nama?: string;
    email?: string;
  };
};

type Props = {
  name?: string;
  role?: string;
  description?: string;
};

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

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

function roleLabel(role: string): string {
  switch (role) {
    case "admin":
      return "Admin";
    case "staff":
      return "Staff";
    case "member":
      return "Member";
    case "agent_member":
      return "Agent Member";
    case "master_member":
      return "Master Member";
    case "operator_trx":
      return "Operator Transaksi";
    case "operator_wallet":
      return "Operator Wallet";
    case "user":
      return "User";
    case "agent":
      return "Agent";
    case "master":
      return "Master";
    default:
      return role || "-";
  }
}

export default function DashboardProfileCard({
  name: initialName,
  role: initialRole,
  description = "Profil akun yang sedang aktif di dashboard ini.",
}: Props) {
  const [name, setName] = useState(initialName || "");
  const [role] = useState(() => {
    if (initialRole) return initialRole;
    if (typeof window === "undefined") return "";
    const token = localStorage.getItem("auth_token") || "";
    return decodeRoleFromToken(token);
  });

  useEffect(() => {
    if (!initialName) {
      void (async () => {
        try {
          const r = await fetch("/api/me/profile", {
            headers: authHeader(),
            cache: "no-store",
          });
          const j: ProfileResponse = await r.json().catch(() => ({}));
          if (r.ok && j?.ok) {
            if (j.profile?.nama) setName(String(j.profile.nama));
          }
        } catch {
          // noop
        }
      })();
    }
  }, [initialName, initialRole]);

  const computedRoleLabel = useMemo(() => roleLabel(role), [role]);

  return (
    <div className="rounded-2xl border border-white/12 bg-linear-to-br from-slate-900/90 via-slate-900/72 to-cyan-950/28 p-5 shadow-[0_20px_48px_-28px_rgba(34,211,238,0.35)]">
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <div className="text-xs font-semibold uppercase tracking-[0.2em] text-cyan-200/80">Profil Aktif</div>
          <div className="mt-2 truncate text-2xl font-semibold text-white">{name || "Pengguna Dashboard"}</div>
          <div className="mt-2 max-w-2xl text-sm text-white/65">{description}</div>
        </div>
        <div className="grid h-12 w-12 shrink-0 place-items-center rounded-2xl border border-cyan-400/20 bg-cyan-400/10 text-cyan-300">
          <UserRound className="h-5 w-5" />
        </div>
      </div>

      <div className="mt-4 flex flex-wrap items-center gap-3">
        <div className="inline-flex items-center gap-2 rounded-full border border-emerald-400/20 bg-emerald-400/10 px-3 py-1.5 text-sm font-medium text-emerald-200">
          <BadgeCheck className="h-4 w-4" />
          {computedRoleLabel}
        </div>
        <div className="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/4 px-3 py-1.5 text-sm text-slate-200">
          <Shield className="h-4 w-4" />
          Role aktif di sesi ini
        </div>
      </div>
    </div>
  );
}
