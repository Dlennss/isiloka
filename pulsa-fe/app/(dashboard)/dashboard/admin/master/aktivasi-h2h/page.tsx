"use client";

import { useEffect, useMemo, useState } from "react";
import { CheckCircle2, Loader2, PencilLine, RefreshCcw, Settings2, ShieldCheck } from "lucide-react";
import EditUserModal from "@/components/dashboard/EditUserModal";
import { AppModal } from "@/components/ui/app-modal";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { alertError, alertSuccess, alertWarning } from "@/components/ui/alerts";
import { roleLabel } from "@/lib/memberRoles";

type MemberRow = {
  id: number;
  email: string;
  nama: string;
  role: string;
  aktif: boolean;
  saldo: number;
  fee_member_rp: number;
  retail_agent_commission_rp: number;
  retail_master_commission_rp: number;
  h2h_agent_commission_rp: number;
  h2h_master_commission_rp: number;
  h2h_agent_member_id?: number | null;
  h2h_master_member_id?: number | null;
  dibuat_pada?: string;
};

type MemberFeeCategoryRow = {
  id: number;
  member_id: number;
  fee_code: string;
  fee_rp: number;
  aktif: boolean;
};

const H2H_FEE_CODES = ["DANA", "GOPAY", "OVO", "LINKAJA", "SHOPEEPAY", "BANK", "LAINNYA"] as const;

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

function fmtIDR(n: number | string): string {
  const num = typeof n === "string" ? Number(String(n).replace(/[^\d-]/g, "")) : n;
  if (!Number.isFinite(num)) return "0";
  return new Intl.NumberFormat("id-ID").format(num);
}

function digitsOnly(s: string): string {
  return String(s || "").replace(/\D+/g, "");
}

function roleSortWeight(role: string): number {
  if (role === "agent_member") return 0;
  if (role === "member") return 1;
  if (role === "master_member") return 2;
  return 9;
}

export default function AdminH2HActivationPage() {
  const [items, setItems] = useState<MemberRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [editOpen, setEditOpen] = useState(false);
  const [editMember, setEditMember] = useState<MemberRow | null>(null);
  const [feeOpen, setFeeOpen] = useState(false);
  const [feeMember, setFeeMember] = useState<MemberRow | null>(null);
  const [feeLoading, setFeeLoading] = useState(false);
  const [feeSaving, setFeeSaving] = useState(false);
  const [approvingID, setApprovingID] = useState<number | null>(null);
  const [feeValues, setFeeValues] = useState<Record<string, string>>({
    DANA: "0",
    GOPAY: "0",
    OVO: "0",
    LINKAJA: "0",
    SHOPEEPAY: "0",
    LAINNYA: "0",
  });

  async function load() {
    setLoading(true);
    try {
      const qs = new URLSearchParams({
        scope: "h2h",
        limit: "500",
        offset: "0",
      });
      const r = await fetch(`/api/admin/members?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) {
        await alertError(j?.error || "Gagal memuat antrian aktivasi H2H.");
        setItems([]);
        return;
      }

      const rows: MemberRow[] = Array.isArray(j.items) ? j.items : [];
      const pendingRows = rows
        .filter((item) => !item.aktif)
        .sort((a, b) => {
          const roleCmp = roleSortWeight(a.role) - roleSortWeight(b.role);
          if (roleCmp !== 0) return roleCmp;
          return Number(b.id || 0) - Number(a.id || 0);
        });
      setItems(pendingRows);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  const filteredItems = useMemo(() => {
    const needle = search.trim().toLowerCase();
    if (!needle) return items;
    return items.filter((item) =>
      [item.email, item.nama, item.role, String(item.id)].some((value) => String(value || "").toLowerCase().includes(needle)),
    );
  }, [items, search]);

  async function openFeeModal(member: MemberRow) {
    setFeeMember(member);
    setFeeOpen(true);
    setFeeLoading(true);
    setFeeValues({
      DANA: "0",
      GOPAY: "0",
      OVO: "0",
      LINKAJA: "0",
      SHOPEEPAY: "0",
      LAINNYA: "0",
    });
    try {
      const r = await fetch(`/api/admin/members/fee/categories?member_id=${member.id}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) {
        await alertError(j?.error || "Gagal memuat fee kategori H2H.");
        return;
      }

      const rows: MemberFeeCategoryRow[] = Array.isArray(j.items) ? j.items : [];
      const nextValues: Record<string, string> = {
        DANA: "0",
        GOPAY: "0",
        OVO: "0",
        LINKAJA: "0",
        SHOPEEPAY: "0",
        LAINNYA: "0",
      };
      for (const row of rows) {
        const code = String(row.fee_code || "").trim().toUpperCase();
        if (code in nextValues) nextValues[code] = String(Math.floor(Number(row.fee_rp || 0)));
      }
      setFeeValues(nextValues);
    } finally {
      setFeeLoading(false);
    }
  }

  async function saveFees() {
    if (!feeMember) return;
    for (const code of H2H_FEE_CODES) {
      const value = Number(digitsOnly(feeValues[code] || "0"));
      if (!Number.isFinite(value) || value < 0) {
        await alertWarning(`Fee ${code} tidak valid.`);
        return;
      }
    }

    setFeeSaving(true);
    try {
      for (const code of H2H_FEE_CODES) {
        const feeRP = Number(digitsOnly(feeValues[code] || "0"));
        const r = await fetch("/api/admin/members/fee/categories/upsert", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            ...authHeader(),
          },
          body: JSON.stringify({
            member_id: feeMember.id,
            fee_code: code,
            fee_rp: Math.floor(feeRP),
            aktif: true,
          }),
        });
        const j = await r.json().catch(() => ({}));
        if (!r.ok || !j?.ok) {
          await alertError(j?.error || `Gagal menyimpan fee ${code}.`);
          return;
        }
      }
      await alertSuccess("Fee kategori H2H berhasil disimpan.");
      setFeeOpen(false);
      setFeeMember(null);
    } finally {
      setFeeSaving(false);
    }
  }

  async function approveMember(member: MemberRow) {
    setApprovingID(member.id);
    try {
      const r = await fetch(`/api/admin/members/${member.id}`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          ...authHeader(),
        },
        body: JSON.stringify({
          email: member.email,
          nama: member.nama,
          role: member.role,
          aktif: true,
          fee_member_rp: Number(member.fee_member_rp ?? 0),
          retail_agent_commission_rp: Number(member.retail_agent_commission_rp ?? 0),
          retail_master_commission_rp: Number(member.retail_master_commission_rp ?? 0),
          h2h_agent_commission_rp: Number(member.h2h_agent_commission_rp ?? 0),
          h2h_master_commission_rp: Number(member.h2h_master_commission_rp ?? 0),
        }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) {
        await alertError(j?.error || "Gagal mengaktifkan akun H2H.");
        return;
      }
      await alertSuccess(`Akun H2H #${member.id} berhasil diaktifkan.`);
      setItems((prev) => prev.filter((item) => item.id !== member.id));
    } finally {
      setApprovingID(null);
    }
  }

  return (
    <div className="space-y-6 p-6">
      <section className="rounded-2xl border border-white/10 bg-slate-950/40 p-5">
        <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div>
            <div className="flex items-center gap-2 text-white">
              <ShieldCheck className="h-5 w-5 text-cyan-300" />
              <h1 className="text-lg font-semibold">Aktivasi H2H</h1>
            </div>
            <p className="mt-2 max-w-3xl text-sm text-white/60">
              Semua akun H2H yang didaftarkan master atau agent masuk ke antrian ini dalam status nonaktif. Admin dapat koreksi profil, cek fee kategori H2H, lalu mengaktifkan akun ketika sudah benar.
            </p>
          </div>
          <div className="flex items-center gap-3">
            <div className="rounded-2xl border border-cyan-400/25 bg-cyan-500/10 px-4 py-3 text-center">
              <div className="text-xs uppercase tracking-wide text-cyan-200/80">Pending</div>
              <div className="text-2xl font-bold text-cyan-200">{items.length}</div>
            </div>
            <Button variant="outline" onClick={() => void load()} disabled={loading}>
              {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <RefreshCcw className="mr-2 h-4 w-4" />}
              Refresh
            </Button>
          </div>
        </div>
      </section>

      <section className="rounded-2xl border border-white/10 bg-slate-950/40 p-5">
        <div className="mb-4 flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div>
            <h2 className="text-base font-semibold text-white">Antrian Aktivasi</h2>
            <p className="text-sm text-white/55">Fokus khusus akun H2H yang masih menunggu persetujuan admin.</p>
          </div>
          <Input
            className="max-w-md"
            placeholder="Cari email, nama, role, atau ID"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>

        {loading ? <div className="py-8 text-sm text-white/60">Memuat antrian aktivasi H2H...</div> : null}
        {!loading && filteredItems.length === 0 ? (
          <div className="rounded-2xl border border-emerald-400/20 bg-emerald-500/10 px-4 py-6 text-sm text-emerald-100">
            Tidak ada akun H2H pending. Semua pendaftaran yang masuk sudah aktif atau belum ada permintaan baru.
          </div>
        ) : null}

        {!loading && filteredItems.length > 0 ? (
          <div className="space-y-3">
            {filteredItems.map((member) => (
              <div key={member.id} className="rounded-2xl border border-white/10 bg-slate-900/60 p-4">
                <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                  <div className="min-w-0 space-y-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="text-base font-semibold text-white">{member.nama || member.email}</span>
                      <span className="rounded-full border border-amber-400/30 bg-amber-500/10 px-2 py-0.5 text-xs font-semibold text-amber-200">
                        Pending Admin
                      </span>
                    </div>
                    <div className="text-sm text-white/70">{member.email}</div>
                    <div className="text-xs text-white/50">
                      #{member.id} • {roleLabel(member.role)}
                      {member.h2h_agent_member_id ? ` • Agent #${member.h2h_agent_member_id}` : ""}
                      {member.h2h_master_member_id ? ` • Master #${member.h2h_master_member_id}` : ""}
                      {member.dibuat_pada ? ` • Dibuat ${new Date(member.dibuat_pada).toLocaleString("id-ID")}` : ""}
                    </div>
                    <div className="pt-1 text-xs text-white/50">
                      Saldo dompet: Rp {fmtIDR(member.saldo || 0)} • Komisi Agent: Rp {fmtIDR(member.h2h_agent_commission_rp || 0)} • Komisi Master: Rp {fmtIDR(member.h2h_master_commission_rp || 0)}
                    </div>
                  </div>

                  <div className="flex flex-wrap items-center gap-2">
                    <Button
                      variant="outline"
                      onClick={() => {
                        setEditMember(member);
                        setEditOpen(true);
                      }}
                    >
                      <PencilLine className="mr-2 h-4 w-4" />
                      Edit Profil
                    </Button>
                    <Button variant="outline" onClick={() => void openFeeModal(member)}>
                      <Settings2 className="mr-2 h-4 w-4" />
                      Fee H2H
                    </Button>
                    <Button
                      variant="success"
                      onClick={() => void approveMember(member)}
                      disabled={approvingID === member.id}
                    >
                      {approvingID === member.id ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <CheckCircle2 className="mr-2 h-4 w-4" />}
                      Aktifkan
                    </Button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : null}
      </section>

      <EditUserModal
        open={editOpen}
        user={editMember}
        onClose={() => {
          setEditOpen(false);
          setEditMember(null);
        }}
        scope="h2h"
        onSuccess={async () => {
          await load();
        }}
      />

      <AppModal
        open={feeOpen && !!feeMember}
        onClose={() => {
          if (feeSaving) return;
          setFeeOpen(false);
          setFeeMember(null);
        }}
        title="Koreksi Fee H2H"
        subtitle={feeMember ? `Member #${feeMember.id} ${feeMember.email}` : ""}
        maxWidthClassName="max-w-3xl"
        footer={
          <div className="flex justify-end gap-2">
            <Button
              variant="outline"
              onClick={() => {
                if (feeSaving) return;
                setFeeOpen(false);
                setFeeMember(null);
              }}
              disabled={feeSaving}
            >
              Tutup
            </Button>
            <Button variant="success" onClick={() => void saveFees()} disabled={feeSaving || feeLoading}>
              {feeSaving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Settings2 className="mr-2 h-4 w-4" />}
              Simpan Fee
            </Button>
          </div>
        }
      >
        <div className="space-y-4">
          <div className="rounded-xl border border-cyan-400/20 bg-cyan-500/10 px-4 py-3 text-sm text-cyan-100">
            Koreksi fee kategori H2H sebelum akun diaktifkan. Jalur transaksi H2H membaca kategori DANA, GOPAY, OVO, LINKAJA, SHOPEEPAY, BANK, dan LAINNYA.
          </div>
          {feeLoading ? <div className="py-8 text-sm text-white/60">Memuat fee kategori H2H...</div> : null}
          {!feeLoading ? (
            <div className="grid gap-3 md:grid-cols-2">
              {H2H_FEE_CODES.map((code) => (
                <label key={code} className="grid gap-2">
                  <span className="text-sm text-slate-200">{code}</span>
                  <Input
                    inputMode="numeric"
                    value={fmtIDR(feeValues[code] || "0")}
                    onChange={(e) =>
                      setFeeValues((prev) => ({
                        ...prev,
                        [code]: digitsOnly(e.target.value),
                      }))
                    }
                    placeholder="0"
                  />
                </label>
              ))}
            </div>
          ) : null}
        </div>
      </AppModal>
    </div>
  );
}
