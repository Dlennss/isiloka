"use client";

import { useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableActions, type DataTableColumn } from "@/components/ui/data-table";
import RegisterMemberModal from "@/components/dashboard/RegisterMemberModal";
import EditUserModal from "@/components/dashboard/EditUserModal";
import { ProductSearchSelect } from "@/components/dashboard/ProductSearchSelect";
import { fetchAllAdminProducts } from "@/lib/adminProducts";
import { alertConfirm, alertError, alertSuccess, alertWarning } from "@/components/ui/alerts";
import { Input } from "@/components/ui/input";
import { AppModal } from "@/components/ui/app-modal";
import { ArrowDownCircle, ArrowUpCircle, Clock3, Coins, Eye, Loader2, MoreHorizontal, Plus, Save, Search, ShieldCheck, Trash2, UserCog, Users, Wallet } from "lucide-react";
import { decodeJwt, type JwtClaims } from "@/lib/jwt";
import { createRolesForScope, roleLabel, type AccountScope, type ManageableRole } from "@/lib/memberRoles";

const H2H_FEE_CATEGORY_NAMES = ["DANA", "GOPAY", "OVO", "LINKAJA", "SHOPEEPAY", "BANK", "LAINNYA"] as const;
const H2H_FEE_CATEGORY_OPTIONS = H2H_FEE_CATEGORY_NAMES.map((name) => ({
  id: name,
  sku: name,
  nama: name,
  aktif: true,
}));

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
  retail_agent_id?: number | null;
  retail_master_id?: number | null;
  h2h_agent_member_id?: number | null;
  h2h_master_member_id?: number | null;
  dibuat_pada: string;
};

type HierarchyTargetPreview = {
  member_id: number;
  email: string;
  nama: string;
  role: string;
  commission_rp: number;
  calculated_total: number;
  calculated_count: number;
};

type HierarchyPreview = {
  scope: string;
  member_id: number;
  member_email: string;
  member_nama: string;
  member_role: string;
  current_agent_id?: number | null;
  current_master_id?: number | null;
  next_agent_id?: number | null;
  next_master_id?: number | null;
  derived_master_from_agent: boolean;
  transaction_count: number;
  agent?: HierarchyTargetPreview | null;
  master?: HierarchyTargetPreview | null;
};

type MemberFeeProductRow = {
  id: number;
  member_id: number;
  produk_id?: number;
  kode_produk: string;
  nama_produk?: string;
  fee_persen?: number | null;
  fee_rp?: number | null;
};

type MemberFeeCategoryRow = {
  id: number;
  member_id: number;
  fee_code: string;
  kategori_nama: string;
  fee_rp: number;
  aktif: boolean;
};

type ProdukOption = {
  id: number;
  sku: string;
  nama: string;
  aktif: boolean;
};

const PAGE_SIZE = 10;

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  if (!t) return {};
  return { Authorization: `Bearer ${t}` };
}

/** format angka jadi 1.234.567 */
function fmtIDR(n: number | string): string {
  const num = typeof n === "string" ? Number(String(n).replace(/[^\d-]/g, "")) : n;
  if (!Number.isFinite(num)) return "0";
  return new Intl.NumberFormat("id-ID").format(num);
}

function fmtDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return new Intl.DateTimeFormat("id-ID", { day: "2-digit", month: "short", year: "numeric" }).format(date);
}

/** ambil hanya digit */
function digitsOnly(s: string): string {
  return String(s || "").replace(/\D+/g, "");
}

/** untuk input amount: simpan sebagai digit-only, tampilkan dengan pemisah ribuan */
function normalizeAmountInput(raw: string): { digits: string; pretty: string } {
  const digits = digitsOnly(raw);
  const pretty = digits ? fmtIDR(digits) : "";
  return { digits, pretty };
}

export default function AdminMembersPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [currentRole, setCurrentRole] = useState<string>("");

  const [accountScope, setAccountScope] = useState<AccountScope>("retail");
  const [items, setItems] = useState<MemberRow[]>([]);
  const [search, setSearch] = useState("");
  const [roleFilter, setRoleFilter] = useState("");
  const [loading, setLoading] = useState(false);
  const [offset, setOffset] = useState(0);
  const [hasNext, setHasNext] = useState(false);
  const [totalSaldo, setTotalSaldo] = useState(0);
  const [registerOpen, setRegisterOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [editMember, setEditMember] = useState<MemberRow | null>(null);

  const [adjOpen, setAdjOpen] = useState(false);
  const [adjMember, setAdjMember] = useState<MemberRow | null>(null);
  const [direction, setDirection] = useState<"credit" | "debit">("credit");
  const [amountDigits, setAmountDigits] = useState("10000"); // disimpan digit-only
  const [amountPretty, setAmountPretty] = useState(fmtIDR("10000")); // ditampilkan pretty
  const [note, setNote] = useState("");
  const [adjustSubmitting, setAdjustSubmitting] = useState(false);

  const [feeOpen, setFeeOpen] = useState(false);
  const [feeMember, setFeeMember] = useState<MemberRow | null>(null);
  const [newFee, setNewFee] = useState("");
  const [savingFee, setSavingFee] = useState(false);
  const [commissionOpen, setCommissionOpen] = useState(false);
  const [commissionMember, setCommissionMember] = useState<MemberRow | null>(null);
  const [commissionValue, setCommissionValue] = useState("");
  const [savingCommission, setSavingCommission] = useState(false);
  const [feeProductOpen, setFeeProductOpen] = useState(false);
  const [feeProductMember, setFeeProductMember] = useState<MemberRow | null>(null);
  const [feeProducts, setFeeProducts] = useState<MemberFeeProductRow[]>([]);
  const [feeCategories, setFeeCategories] = useState<MemberFeeCategoryRow[]>([]);
  const [feeProductLoading, setFeeProductLoading] = useState(false);
  const [feeCategoryLoading, setFeeCategoryLoading] = useState(false);
  const [feeProductProdukID, setFeeProductProdukID] = useState("");
  const [feeCategoryID, setFeeCategoryID] = useState("");
  const [feeProductOptions, setFeeProductOptions] = useState<ProdukOption[]>([]);
  const [feeProductOptionsLoading, setFeeProductOptionsLoading] = useState(false);
  const [feeProductMode, setFeeProductMode] = useState<"rp" | "persen">("rp");
  const [feeProductValue, setFeeProductValue] = useState("");
  const [feeCategoryValue, setFeeCategoryValue] = useState("");
  const [feeCategoryAktif, setFeeCategoryAktif] = useState(true);
  const [feeProductSaving, setFeeProductSaving] = useState(false);
  const [feeCategorySaving, setFeeCategorySaving] = useState(false);
  const [feeProductDeletingID, setFeeProductDeletingID] = useState<number | null>(null);
  const [feeCategoryDeletingID, setFeeCategoryDeletingID] = useState<number | null>(null);
  const [actionMenu, setActionMenu] = useState<{ key: string; top: number; left: number; member: MemberRow } | null>(null);
  const [hierarchyOpen, setHierarchyOpen] = useState(false);
  const [hierarchyMember, setHierarchyMember] = useState<MemberRow | null>(null);
  const [hierarchyAgentValue, setHierarchyAgentValue] = useState("");
  const [hierarchyMasterValue, setHierarchyMasterValue] = useState("");
  const [hierarchySearch, setHierarchySearch] = useState("");
  const [hierarchySearchLoading, setHierarchySearchLoading] = useState(false);
  const [hierarchySearchResults, setHierarchySearchResults] = useState<MemberRow[]>([]);
  const [hierarchyTargetMember, setHierarchyTargetMember] = useState<MemberRow | null>(null);
  const [hierarchyTargetRole, setHierarchyTargetRole] = useState("");
  const [hierarchyPreview, setHierarchyPreview] = useState<HierarchyPreview | null>(null);
  const [hierarchyLoading, setHierarchyLoading] = useState(false);
  const [hierarchySaving, setHierarchySaving] = useState(false);
  const [downlineViewerOpen, setDownlineViewerOpen] = useState(false);
  const [downlineViewerMember, setDownlineViewerMember] = useState<MemberRow | null>(null);
  const [downlineViewerLoading, setDownlineViewerLoading] = useState(false);
  const [downlineViewerItems, setDownlineViewerItems] = useState<MemberRow[]>([]);

  const createRoleOptions = createRolesForScope(accountScope);

  function scopeLabel(scope: AccountScope): string {
    switch (scope) {
      case "h2h":
        return "Akun H2H";
      case "retail":
        return "Pengguna Aplikasi";
      case "internal":
        return "Tim Operasional";
      default:
        return "Pengguna";
    }
  }

  function effectiveCommission(member: MemberRow, scope: AccountScope): number {
    if (scope === "retail") {
      if (member.role === "agent") return Number(member.retail_agent_commission_rp || 0);
      if (member.role === "master") return Number(member.retail_master_commission_rp || 0);
      return 0;
    }
    if (scope === "h2h") {
      if (member.role === "agent_member") return Number(member.h2h_agent_commission_rp || 0);
      if (member.role === "master_member") return Number(member.h2h_master_commission_rp || 0);
      return 0;
    }
    return 0;
  }

  function commissionTitleForScope(scope: AccountScope): string {
    if (scope === "retail") return "Komisi Retail";
    if (scope === "h2h") return "Komisi H2H";
    return "Komisi";
  }

  function commissionLabel(member: MemberRow, scope: AccountScope): string {
    if (scope === "h2h") {
      if (member.role === "agent_member") return "Komisi Agent Member / trx";
      if (member.role === "master_member") return "Komisi Master Member / trx";
    }
    if (scope === "retail") {
      if (member.role === "agent") return "Komisi Agent Retail / trx";
      if (member.role === "master") return "Komisi Master Retail / trx";
    }
    return "Komisi / trx";
  }

  function hierarchyAllowedTargetRoles(manager: MemberRow, scope: AccountScope): string[] {
    if (scope === "retail") {
      if (manager.role === "master") return ["agent", "user"];
      if (manager.role === "agent") return ["user"];
      return [];
    }
    if (scope === "h2h") {
      if (manager.role === "master_member") return ["agent_member", "member"];
      if (manager.role === "agent_member") return ["member"];
      return [];
    }
    return [];
  }

  function hierarchyActionLabel(role: string): string {
    if (role === "agent" || role === "agent_member") return "Jadikan Agent";
    if (role === "user" || role === "member") return "Jadikan Member";
    return "Pilih";
  }

  function hierarchySearchPlaceholder(manager: MemberRow, scope: AccountScope): string {
    const allowed = hierarchyAllowedTargetRoles(manager, scope);
    if (allowed.includes("agent") || allowed.includes("agent_member")) {
      return "Cari nama atau email member/agent";
    }
    return "Cari nama atau email member";
  }

  async function openDownlineViewer(member: MemberRow) {
    if (isInternalScope) {
      await alertWarning("Akun internal tidak punya downline.");
      return;
    }
    setDownlineViewerMember(member);
    setDownlineViewerItems([]);
    setDownlineViewerOpen(true);
    setDownlineViewerLoading(true);
    try {
      const qs = new URLSearchParams({
        scope: accountScope,
        limit: "500",
        offset: "0",
      });
      const r = await fetch(`/api/admin/members?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) {
        await alertError(j?.error || "Gagal memuat downline.");
        return;
      }
      const rows: MemberRow[] = Array.isArray(j.items) ? j.items : [];
      const filtered = rows.filter((item) => {
        if (accountScope === "retail") {
          return member.role === "master"
            ? item.retail_master_id === member.id || item.retail_agent_id === member.id
            : item.retail_agent_id === member.id;
        }
        return member.role === "master_member"
          ? item.h2h_master_member_id === member.id || item.h2h_agent_member_id === member.id
          : item.h2h_agent_member_id === member.id;
      });
      setDownlineViewerItems(filtered);
    } finally {
      setDownlineViewerLoading(false);
    }
  }

  function openCommissionModal(member: MemberRow) {
    setCommissionMember(member);
    setCommissionValue(String(effectiveCommission(member, accountScope)));
    setCommissionOpen(true);
  }

  async function load(nextOffset = offset, nextRole = roleFilter, nextScope = accountScope) {
    setLoading(true);
    try {
      const qs = new URLSearchParams();
      if (search) qs.set("search", search);
      qs.set("limit", String(PAGE_SIZE + 1));
      qs.set("offset", String(nextOffset));
      qs.set("scope", nextScope);
      if (nextRole) qs.set("role", nextRole);

      const r = await fetch(`/api/admin/members?${qs.toString()}`, {
        headers: authHeader(),
      });

      const j = await r.json().catch(() => ({}));
      const all: MemberRow[] = Array.isArray(j.items) ? j.items : [];
      setHasNext(all.length > PAGE_SIZE);
      setItems(all.slice(0, PAGE_SIZE));
      setTotalSaldo(Number(j.total_saldo || 0));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    const token = localStorage.getItem("auth_token") || "";
    const claims = decodeJwt(token) as JwtClaims | null;
    setCurrentRole(String(claims?.role || "").toLowerCase());
  }, []);

  useEffect(() => {
    const requestedScope = (searchParams.get("scope") || "").toLowerCase();
    if (requestedScope === "retail" || requestedScope === "internal") {
      setAccountScope((prev) => (prev === requestedScope ? prev : (requestedScope as AccountScope)));
    }
  }, [searchParams]);

  useEffect(() => {
    load(offset, roleFilter, accountScope);
     
  }, [offset, accountScope]);

  useEffect(() => {
    function handleOutsideClick(e: MouseEvent) {
      const target = e.target as HTMLElement | null;
      if (!target) return;
      if (target.closest("[data-action-dropdown]")) return;
      setActionMenu(null);
    }

    function handleEsc(e: KeyboardEvent) {
      if (e.key === "Escape") setActionMenu(null);
    }

    document.addEventListener("mousedown", handleOutsideClick);
    document.addEventListener("keydown", handleEsc);
    return () => {
      document.removeEventListener("mousedown", handleOutsideClick);
      document.removeEventListener("keydown", handleEsc);
    };
  }, []);

  async function doAdjust() {
    if (!adjMember) return;
    if (adjustSubmitting) return;

    const amt = Number(amountDigits || "0");
    if (!amt || amt <= 0) return alertWarning("Amount tidak valid.");
    if (!note.trim()) return alertWarning("Note wajib diisi.");

    setAdjustSubmitting(true);
    try {
      const r = await fetch("/api/admin/wallet/adjust", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({
          member_id: adjMember.id,
          amount: amt,
          direction,
          reason: `admin manual ${direction}`,
          note: note.trim(),
        }),
      });

      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal adjust");
        return;
      }

      setAdjOpen(false);
      setAdjMember(null);
      setNote("");
      await load();
      await alertSuccess(`Adjust berhasil. ref_id=${j.ref_id || "-"}`);
    } finally {
      setAdjustSubmitting(false);
    }
  }

  async function doSetFee() {
    if (!feeMember) return;
    const feeRaw = Number(newFee || "0");
    const fee = Number.isFinite(feeRaw) && feeRaw >= 0 ? Math.floor(feeRaw) : 0;

    setSavingFee(true);
    try {
      const r = await fetch("/api/admin/members/fee", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({
          member_id: feeMember.id,
          fee_member_rp: fee,
        }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) return alertError(j.error || "Gagal update fee");

      setItems((prev) => prev.map((m) => (m.id === feeMember.id ? { ...m, fee_member_rp: fee } : m)));
      setFeeOpen(false);
      setFeeMember(null);
      setNewFee("");
      await alertSuccess("Fee berhasil diupdate.");
    } finally {
      setSavingFee(false);
    }
  }

  async function doSaveCommission() {
    if (!commissionMember) return;
    const value = Number(digitsOnly(commissionValue || "0"));
    if (!Number.isFinite(value) || value < 0) return alertWarning("Komisi tidak valid.");

    const isAgentRole = commissionMember.role === "agent" || commissionMember.role === "agent_member";
    const isMasterRole = commissionMember.role === "master" || commissionMember.role === "master_member";
    if (!isAgentRole && !isMasterRole) return alertWarning("Komisi hanya berlaku untuk akun agent atau master.");

    setSavingCommission(true);
    try {
      const body =
        accountScope === "h2h"
          ? {
              email: commissionMember.email,
              nama: commissionMember.nama,
              role: commissionMember.role,
              aktif: commissionMember.aktif,
              fee_member_rp: Number(commissionMember.fee_member_rp ?? 0),
              retail_agent_commission_rp: Number(commissionMember.retail_agent_commission_rp ?? 0),
              retail_master_commission_rp: Number(commissionMember.retail_master_commission_rp ?? 0),
              h2h_agent_commission_rp: isAgentRole ? Math.floor(value) : Number(commissionMember.h2h_agent_commission_rp ?? 0),
              h2h_master_commission_rp: isMasterRole ? Math.floor(value) : Number(commissionMember.h2h_master_commission_rp ?? 0),
            }
          : {
              email: commissionMember.email,
              nama: commissionMember.nama,
              role: commissionMember.role,
              aktif: commissionMember.aktif,
              fee_member_rp: Number(commissionMember.fee_member_rp ?? 0),
              retail_agent_commission_rp: isAgentRole ? Math.floor(value) : Number(commissionMember.retail_agent_commission_rp ?? 0),
              retail_master_commission_rp: isMasterRole ? Math.floor(value) : Number(commissionMember.retail_master_commission_rp ?? 0),
              h2h_agent_commission_rp: Number(commissionMember.h2h_agent_commission_rp ?? 0),
              h2h_master_commission_rp: Number(commissionMember.h2h_master_commission_rp ?? 0),
            };

      const r = await fetch(`/api/admin/members/${commissionMember.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify(body),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) return alertError(j.error || "Gagal update komisi");

      setItems((prev) =>
        prev.map((item) =>
          item.id !== commissionMember.id
            ? item
            : accountScope === "h2h"
              ? {
                  ...item,
                  h2h_agent_commission_rp: isAgentRole ? Math.floor(value) : item.h2h_agent_commission_rp,
                  h2h_master_commission_rp: isMasterRole ? Math.floor(value) : item.h2h_master_commission_rp,
                }
              : {
                  ...item,
                  retail_agent_commission_rp: isAgentRole ? Math.floor(value) : item.retail_agent_commission_rp,
                  retail_master_commission_rp: isMasterRole ? Math.floor(value) : item.retail_master_commission_rp,
                },
        ),
      );
      setCommissionOpen(false);
      setCommissionMember(null);
      await alertSuccess("Komisi berhasil diupdate.");
    } finally {
      setSavingCommission(false);
    }
  }

  async function loadFeeProducts(memberID: number) {
    setFeeProductLoading(true);
    try {
      const qs = new URLSearchParams({ member_id: String(memberID) });
      const r = await fetch(`/api/admin/members/fee/products?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal mengambil fee produk");
        setFeeProducts([]);
        return;
      }
      setFeeProducts(Array.isArray(j.items) ? j.items : []);
    } finally {
      setFeeProductLoading(false);
    }
  }

  async function loadFeeCategories(memberID: number) {
    setFeeCategoryLoading(true);
    try {
      const qs = new URLSearchParams({ member_id: String(memberID) });
      const r = await fetch(`/api/admin/members/fee/categories?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal mengambil fee kategori");
        setFeeCategories([]);
        return;
      }
      setFeeCategories(Array.isArray(j.items) ? j.items : []);
    } finally {
      setFeeCategoryLoading(false);
    }
  }

  async function loadFeeProductOptions() {
    setFeeProductOptionsLoading(true);
    try {
      const raw = await fetchAllAdminProducts(authHeader());
      const mapped: ProdukOption[] = raw
        .map((x: Partial<ProdukOption>) => ({
          id: Number(x.id || 0),
          sku: String(x.sku || "").toUpperCase(),
          nama: String(x.nama || ""),
          aktif: Boolean(x.aktif),
        }))
        .filter((x: ProdukOption) => x.id > 0 && x.sku);
      setFeeProductOptions(mapped);
    } catch {
      await alertError("Gagal mengambil daftar produk");
      setFeeProductOptions([]);
    } finally {
      setFeeProductOptionsLoading(false);
    }
  }

  async function openFeeProducts(member: MemberRow) {
    setFeeProductMember(member);
    setFeeProductOpen(true);
    setFeeProductProdukID("");
    setFeeCategoryID("");
    setFeeProductMode("rp");
    setFeeProductValue("");
    setFeeCategoryValue("");
    setFeeCategoryAktif(true);
    await loadFeeCategories(member.id);
  }

  async function doUpsertFeeProduct() {
    if (!feeProductMember) return;

    const produkID = Number(feeProductProdukID || "0");
    if (!produkID) return alertWarning("Produk wajib dipilih.");
    const produk = feeProductOptions.find((x) => x.id === produkID);
    if (!produk) return alertWarning("Produk tidak valid.");

    const body: Record<string, number | string> = {
      member_id: feeProductMember.id,
      produk_id: produkID,
      kode_produk: produk.sku,
    };

    if (feeProductMode === "rp") {
      const feeRP = Number(digitsOnly(feeProductValue || "0"));
      if (!Number.isFinite(feeRP) || feeRP < 0) return alertWarning("Fee Rp tidak valid.");
      body.fee_rp = Math.floor(feeRP);
    } else {
      const raw = String(feeProductValue || "").replace(",", ".").trim();
      const feePersen = Number(raw);
      if (!Number.isFinite(feePersen) || feePersen < 0) return alertWarning("Fee persen tidak valid.");
      body.fee_persen = feePersen;
    }

    setFeeProductSaving(true);
    try {
      const r = await fetch("/api/admin/members/fee/products/upsert", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify(body),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) return alertError(j.error || "Gagal simpan fee produk");

      await loadFeeProducts(feeProductMember.id);
      setFeeProductProdukID("");
      setFeeProductValue("");
      await alertSuccess("Fee produk berhasil disimpan.");
    } finally {
      setFeeProductSaving(false);
    }
  }

  async function doDeleteFeeProduct(row: MemberFeeProductRow) {
    if (!feeProductMember) return;

    const ok = await alertConfirm({
      title: "Hapus fee produk?",
      text: `Hapus konfigurasi ${row.kode_produk} untuk member ini?`,
      confirmButtonText: "Ya, hapus",
      cancelButtonText: "Batal",
    });
    if (!ok) return;

    setFeeProductDeletingID(row.id);
    try {
      const r = await fetch("/api/admin/members/fee/products/delete", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({
          member_id: feeProductMember.id,
          produk_id: row.produk_id ?? 0,
          kode_produk: row.kode_produk,
        }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) return alertError(j.error || "Gagal hapus fee produk");

      await loadFeeProducts(feeProductMember.id);
      await alertSuccess("Fee produk berhasil dihapus.");
    } finally {
      setFeeProductDeletingID(null);
    }
  }

  async function doUpsertFeeCategory() {
    if (!feeProductMember) return;

    const feeCode = String(feeCategoryID || "").trim().toUpperCase();
    if (!H2H_FEE_CATEGORY_NAMES.includes(feeCode as typeof H2H_FEE_CATEGORY_NAMES[number])) return alertWarning("Kategori fee H2H wajib dipilih.");
    const feeRP = Number(digitsOnly(feeCategoryValue || "0"));
    if (!Number.isFinite(feeRP) || feeRP < 0) return alertWarning("Fee kategori tidak valid.");

    setFeeCategorySaving(true);
    try {
      const r = await fetch("/api/admin/members/fee/categories/upsert", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({
          member_id: feeProductMember.id,
          fee_code: feeCode,
          fee_rp: Math.floor(feeRP),
          aktif: feeCategoryAktif,
        }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) return alertError(j.error || "Gagal simpan fee kategori");

      await loadFeeCategories(feeProductMember.id);
      setFeeCategoryID("");
      setFeeCategoryValue("");
      setFeeCategoryAktif(true);
      await alertSuccess("Fee kategori berhasil disimpan.");
    } finally {
      setFeeCategorySaving(false);
    }
  }

  async function doDeleteFeeCategory(row: MemberFeeCategoryRow) {
    if (!feeProductMember) return;

    const ok = await alertConfirm({
      title: "Hapus fee kategori?",
      text: `Hapus konfigurasi ${row.kategori_nama} untuk member ini?`,
      confirmButtonText: "Ya, hapus",
      cancelButtonText: "Batal",
    });
    if (!ok) return;

    setFeeCategoryDeletingID(row.id);
    try {
      const r = await fetch("/api/admin/members/fee/categories/delete", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({
          member_id: feeProductMember.id,
          fee_code: row.fee_code,
        }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) return alertError(j.error || "Gagal hapus fee kategori");

      await loadFeeCategories(feeProductMember.id);
      await alertSuccess("Fee kategori berhasil dihapus.");
    } finally {
      setFeeCategoryDeletingID(null);
    }
  }

  function openWalletActionMenu(member: MemberRow, element: HTMLElement) {
    const rect = element.getBoundingClientRect();
    const menuWidth = 176;
    const viewportPadding = 12;
    const preferredLeft = rect.right - menuWidth;
    const left = Math.max(viewportPadding, Math.min(preferredLeft, window.innerWidth - menuWidth - viewportPadding));
    const top = Math.min(rect.bottom + 8, window.innerHeight - 220);
    setActionMenu({ key: `manage-${member.id}`, top, left, member });
  }

  function openAdjust(m: MemberRow) {
    setAdjMember(m);
    setAdjOpen(true);
    setDirection("credit");

    const initial = "10000";
    const norm = normalizeAmountInput(initial);
    setAmountDigits(norm.digits || "0");
    setAmountPretty(norm.pretty || "0");

    setNote("");
  }

  function openHierarchy(member: MemberRow) {
    if (hierarchyAllowedTargetRoles(member, accountScope).length === 0) {
      void alertWarning("Menu downline hanya berlaku untuk akun agent atau master.");
      return;
    }
    setHierarchyMember(member);
    setHierarchyPreview(null);
    setHierarchyAgentValue("");
    setHierarchyMasterValue("");
    setHierarchySearch("");
    setHierarchySearchResults([]);
    setHierarchyTargetMember(null);
    setHierarchyTargetRole("");
    setHierarchyOpen(true);
  }

  async function searchHierarchyCandidates() {
    if (!hierarchyMember) return;
    const q = hierarchySearch.trim();
    if (!q) {
      setHierarchySearchResults([]);
      return;
    }
    setHierarchySearchLoading(true);
    try {
      const qs = new URLSearchParams({
        scope: accountScope,
        search: q,
        limit: "20",
        offset: "0",
      });
      const r = await fetch(`/api/admin/members?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) {
        await alertError(j?.error || "Gagal mencari akun.");
        return;
      }
      const rows: MemberRow[] = Array.isArray(j.items) ? j.items : [];
      setHierarchySearchResults(rows.filter((item) => item.id !== hierarchyMember.id));
    } finally {
      setHierarchySearchLoading(false);
    }
  }

  function selectHierarchyTarget(member: MemberRow, targetRole: string) {
    if (!hierarchyMember) return;
    setHierarchyTargetMember(member);
    setHierarchyTargetRole(targetRole);
    if (accountScope === "h2h") {
      if (hierarchyMember.role === "master_member") {
        setHierarchyAgentValue("");
        setHierarchyMasterValue(String(hierarchyMember.id));
      } else {
        setHierarchyAgentValue(String(hierarchyMember.id));
        setHierarchyMasterValue(hierarchyMember.h2h_master_member_id ? String(hierarchyMember.h2h_master_member_id) : "");
      }
    } else {
      if (hierarchyMember.role === "master") {
        setHierarchyAgentValue("");
        setHierarchyMasterValue(String(hierarchyMember.id));
      } else {
        setHierarchyAgentValue(String(hierarchyMember.id));
        setHierarchyMasterValue(hierarchyMember.retail_master_id ? String(hierarchyMember.retail_master_id) : "");
      }
    }
    setHierarchyPreview(null);
  }

  async function previewHierarchy() {
    if (!hierarchyMember || !hierarchyTargetMember || !hierarchyTargetRole) {
      await alertWarning("Cari lalu pilih dulu akun downline yang ingin ditambahkan.");
      return;
    }
    setHierarchyLoading(true);
    try {
      const agentID = hierarchyAgentValue ? Number(hierarchyAgentValue) : null;
      const masterID = hierarchyMasterValue ? Number(hierarchyMasterValue) : null;
      const r = await fetch("/api/admin/members/hierarchy/preview", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({
          scope: accountScope,
          member_id: hierarchyTargetMember.id,
          agent_id: agentID || null,
          master_id: masterID || null,
          target_role: hierarchyTargetRole,
        }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) {
        await alertError(j?.error || "Gagal preview hierarki.");
        return;
      }
      setHierarchyPreview(j.item || null);
      if (j?.item?.next_agent_id) setHierarchyAgentValue(String(j.item.next_agent_id));
      if (j?.item?.next_master_id) setHierarchyMasterValue(String(j.item.next_master_id));
    } finally {
      setHierarchyLoading(false);
    }
  }

  async function applyHierarchy(applyHistorical: boolean) {
    if (!hierarchyMember || !hierarchyTargetMember || !hierarchyTargetRole) {
      await alertWarning("Cari lalu pilih dulu akun downline yang ingin ditambahkan.");
      return;
    }
    setHierarchySaving(true);
    try {
      const agentID = hierarchyAgentValue ? Number(hierarchyAgentValue) : null;
      const masterID = hierarchyMasterValue ? Number(hierarchyMasterValue) : null;
      const r = await fetch("/api/admin/members/hierarchy/apply", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({
          scope: accountScope,
          member_id: hierarchyTargetMember.id,
          agent_id: agentID || null,
          master_id: masterID || null,
          target_role: hierarchyTargetRole,
          apply_historical: applyHistorical,
        }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) {
        await alertError(j?.error || "Gagal menyimpan hierarki.");
        return;
      }
      const preview: HierarchyPreview | null = j?.item?.preview || null;
      setHierarchyPreview(preview);
      setItems((prev) =>
        prev.map((item) =>
          item.id !== hierarchyTargetMember.id
            ? item
            : {
                ...item,
                role: hierarchyTargetRole,
                retail_agent_id: accountScope === "retail" ? preview?.next_agent_id ?? null : item.retail_agent_id,
                retail_master_id: accountScope === "retail" ? preview?.next_master_id ?? null : item.retail_master_id,
                h2h_agent_member_id: accountScope === "h2h" ? preview?.next_agent_id ?? null : item.h2h_agent_member_id,
                h2h_master_member_id: accountScope === "h2h" ? preview?.next_master_id ?? null : item.h2h_master_member_id,
              },
        ),
      );
      await alertSuccess(
        applyHistorical
          ? `Hierarki dan komisi historis berhasil diterapkan. Agent: Rp ${fmtIDR(Number(j?.item?.agent_applied_total || 0))}, Master: Rp ${fmtIDR(Number(j?.item?.master_applied_total || 0))}.`
          : "Hierarki berhasil disimpan tanpa komisi historis.",
      );
      setHierarchyOpen(false);
      setHierarchyMember(null);
      setHierarchyTargetMember(null);
      setHierarchyTargetRole("");
      setHierarchyPreview(null);
    } finally {
      setHierarchySaving(false);
    }
  }

  const columns: DataTableColumn<MemberRow>[] = [
    {
      id: "account",
      header: "Akun",
      tdClassName: "min-w-56",
      render: (m) => (
        <div>
          <p className="font-bold text-slate-950">{m.nama || "Nama belum diisi"}</p>
          <p className="mt-0.5 text-xs text-slate-500">{m.email}</p>
        </div>
      ),
    },
    {
      id: "role",
      header: "Role",
      tdClassName: "whitespace-nowrap text-slate-700",
      render: (m) => roleLabel(m.role),
    },
    {
      id: "status",
      header: "Status",
      tdClassName: "whitespace-nowrap",
      render: (m) => (
        <span className={`inline-flex rounded-full px-2.5 py-1 text-xs font-bold ${m.aktif ? "bg-emerald-50 text-emerald-700" : "bg-rose-50 text-rose-700"}`}>
          {m.aktif ? "Aktif" : "Nonaktif"}
        </span>
      ),
    },
    {
      id: "saldo",
      header: "Saldo Utama",
      tdClassName: "whitespace-nowrap font-semibold text-emerald-700",
      render: (m) => `Rp ${fmtIDR(m.saldo)}`,
    },
    {
      id: "registered",
      header: "Terdaftar",
      tdClassName: "whitespace-nowrap text-slate-600",
      render: (m) => fmtDate(m.dibuat_pada),
    },
  ];

  const canManageWallet = currentRole === "admin" || currentRole === "operator_wallet";
  const canManageSecurity = currentRole === "admin";
  const isInternalScope = accountScope === "internal";

  const actions: DataTableActions<MemberRow> = {
    header: "Aksi",
    align: "right",
    render: (m) => (
      <div className="flex items-center justify-end gap-2">
        <Button
          className="h-9 gap-1.5 border-emerald-200 bg-emerald-50 px-3 font-bold text-emerald-800 hover:bg-emerald-100"
          variant="outline"
          onClick={() => {
            setActionMenu(null);
            setEditMember(m);
            setEditOpen(true);
          }}
        >
          <Eye className="h-4 w-4" />
          Detail
        </Button>
        <div className="inline-flex" data-action-dropdown>
          <button
            type="button"
            className="flex h-9 w-9 cursor-pointer items-center justify-center rounded-lg border border-slate-200 bg-white text-slate-600 transition hover:border-emerald-300 hover:bg-emerald-50 hover:text-emerald-800"
            onClick={(e) => {
              const target = e.currentTarget as HTMLElement;
              if (actionMenu?.key === `manage-${m.id}`) {
                setActionMenu(null);
                return;
              }
              openWalletActionMenu(m, target);
            }}
            aria-label={`Tindakan untuk ${m.nama || m.email}`}
            title="Tindakan lainnya"
          >
            <MoreHorizontal className="h-4 w-4" />
          </button>
        </div>
      </div>
    ),
  };

  return (
    <>
      {actionMenu ? (
        <div className="fixed inset-0 z-40" onMouseDown={() => setActionMenu(null)}>
          <div
            className="fixed z-50 w-52 rounded-xl border border-slate-200 bg-white p-2 shadow-[0_18px_45px_rgba(15,23,42,0.18)]"
            style={{ top: actionMenu.top, left: actionMenu.left }}
            onMouseDown={(e) => e.stopPropagation()}
            data-action-dropdown
          >
            {canManageSecurity ? <Button
              className="h-9 w-full justify-start border-0 bg-white text-slate-700 hover:bg-emerald-50 hover:text-emerald-800"
              onClick={() => {
                const member = actionMenu.member;
                setActionMenu(null);
                setEditMember(member);
                setEditOpen(true);
              }}
            >
              <UserCog className="mr-1.5 h-4 w-4" /> Edit Akun
            </Button> : null}
            {canManageWallet && ["user", "agent"].includes(actionMenu.member.role) ? <Button
              className="h-9 w-full justify-start border-0 bg-white text-slate-700 hover:bg-emerald-50 hover:text-emerald-800"
              onClick={() => {
                const member = actionMenu.member;
                setActionMenu(null);
                openAdjust(member);
              }}
            >
              <Wallet className="mr-1.5 h-4 w-4" /> Kelola Saldo
            </Button> : null}
            {(["agent", "master", "agent_member", "master_member"].includes(actionMenu.member.role)) ? <Button
              className="h-9 w-full justify-start border-0 bg-white text-slate-700 hover:bg-emerald-50 hover:text-emerald-800"
              onClick={() => {
                const member = actionMenu.member;
                setActionMenu(null);
                openCommissionModal(member);
              }}
            >
              <Coins className="mr-1.5 h-4 w-4" /> Atur Komisi
            </Button> : null}
            {(["agent", "master", "agent_member", "master_member"].includes(actionMenu.member.role)) ? <Button
              className="h-9 w-full justify-start border-0 bg-white text-slate-700 hover:bg-emerald-50 hover:text-emerald-800"
              onClick={() => {
                const member = actionMenu.member;
                setActionMenu(null);
                void openDownlineViewer(member);
              }}
            >
              <Users className="mr-1.5 h-4 w-4" /> Lihat Jaringan
            </Button> : null}
            {(["agent", "master", "agent_member", "master_member"].includes(actionMenu.member.role)) ? <Button
              className="h-9 w-full justify-start border-0 bg-white text-slate-700 hover:bg-emerald-50 hover:text-emerald-800"
              onClick={() => {
                const member = actionMenu.member;
                setActionMenu(null);
                openHierarchy(member);
              }}
            >
              <ShieldCheck className="mr-1.5 h-4 w-4" /> Atur Jaringan
            </Button> : null}
            <Button
              className="h-9 w-full justify-start border-0 bg-white text-slate-700 hover:bg-emerald-50 hover:text-emerald-800"
              onClick={() => {
                const member = actionMenu.member;
                setActionMenu(null);
                router.push(`/dashboard/admin/master/members/history?member_id=${member.id}`);
              }}
            >
              <Clock3 className="mr-1.5 h-4 w-4" /> Riwayat Aktivitas
            </Button>
          </div>
        </div>
      ) : null}

    <div className="-m-2 min-h-screen bg-[#f0f8f8] p-3 sm:p-5 lg:p-6">
      <div className="mx-auto max-w-7xl space-y-4">
        <section className="rounded-lg border border-emerald-100 bg-white p-4 shadow-sm sm:p-5">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex min-w-0 items-start gap-3">
              <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-emerald-50 text-emerald-700">
                <Users className="h-5 w-5" />
              </div>
              <div className="min-w-0">
                <div className="text-[11px] font-bold uppercase text-emerald-700">Direktori Akun</div>
                <h1 className="mt-1 text-xl font-bold text-slate-950 sm:text-2xl">Akun Isiloka</h1>
                <p className="mt-1 text-sm text-slate-600">Kelola pelanggan, agent, marketing, dan operator dalam satu tempat.</p>
              </div>
            </div>
            <div className="rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 sm:min-w-56">
              <div className="text-[11px] font-bold uppercase text-emerald-700">Total Saldo</div>
              <div className="mt-1 text-xl font-bold text-emerald-950">Rp {fmtIDR(totalSaldo)}</div>
              <div className="mt-0.5 text-xs text-emerald-700">{roleFilter ? roleLabel(roleFilter) : scopeLabel(accountScope)}</div>
            </div>
          </div>
        </section>

        <section className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
          <div className="flex flex-wrap gap-2 border-b border-slate-100 pb-4">
            {[
              { label: "Semua Akun", scope: "retail" as AccountScope, role: "" },
              { label: "Pelanggan", scope: "retail" as AccountScope, role: "user" },
              { label: "Agent", scope: "retail" as AccountScope, role: "agent" },
              { label: "Marketing", scope: "retail" as AccountScope, role: "marketing" },
              { label: "Operator Kredit", scope: "retail" as AccountScope, role: "analis" },
              { label: "Tim Internal", scope: "internal" as AccountScope, role: "" },
            ].map((tab) => (
              <button
                key={`${tab.scope}-${tab.role || "all"}`}
                type="button"
                className={`rounded-md px-4 py-2 text-sm font-semibold transition ${
                  accountScope === tab.scope && roleFilter === tab.role
                    ? "bg-emerald-700 text-white"
                    : "border border-slate-200 bg-white text-slate-600 hover:border-emerald-200 hover:bg-emerald-50 hover:text-emerald-800"
                }`}
                onClick={() => {
                  setAccountScope(tab.scope);
                  setRoleFilter(tab.role);
                  setOffset(0);
                  void load(0, tab.role, tab.scope);
                }}
              >
                {tab.label}
              </button>
            ))}
          </div>

          <div className="mt-4 grid gap-3 md:grid-cols-[auto_minmax(240px,1fr)]">
            {currentRole === "admin" ? (
              <Button
                className="h-11 border-0 bg-emerald-700 px-4 text-white hover:bg-emerald-800"
                onClick={() => setRegisterOpen(true)}
              >
                <Plus className="mr-1.5 h-4 w-4" />
                Tambah Pengguna
              </Button>
            ) : null}
            <div className="flex min-w-0 gap-2">
              <Input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    e.preventDefault();
                    setOffset(0);
                    void load(0);
                  }
                }}
                placeholder="Cari nama atau email"
                className="h-11 min-w-0 flex-1 border-slate-200 bg-white text-slate-900 placeholder:text-slate-400 focus-visible:ring-emerald-500"
              />
              <Button
                className="h-11 w-11 shrink-0 border-0 bg-emerald-700 p-0 text-white hover:bg-emerald-800"
                onClick={() => {
                  setOffset(0);
                  void load(0);
                }}
                disabled={loading}
                aria-label={loading ? "Memuat pencarian" : "Cari pengguna"}
                title="Cari"
              >
                {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Search className="h-4 w-4" />}
              </Button>
            </div>
          </div>
        </section>

        <DataTable
          columns={columns}
          rows={items}
          rowKey={(m) => m.id}
          rowNumberStart={offset + 1}
          minWidthClassName="min-w-[820px]"
          emptyText={`Belum ada ${scopeLabel(accountScope).toLowerCase()}.`}
          actions={actions}
          loading={loading}
          variant="light"
          pagination={{
            page: Math.floor(offset / PAGE_SIZE) + 1,
            totalPages: hasNext ? Math.floor(offset / PAGE_SIZE) + 2 : Math.floor(offset / PAGE_SIZE) + 1,
            onPrev: () => setOffset((v) => Math.max(0, v - PAGE_SIZE)),
            onNext: () => setOffset((v) => v + PAGE_SIZE),
            onPageChange: (nextPage) => setOffset((nextPage - 1) * PAGE_SIZE),
            disablePrev: loading || offset === 0,
            disableNext: loading || !hasNext,
          }}
        />
      </div>
    </div>

      {/* Adjust modal */}
      <RegisterMemberModal
        open={registerOpen}
        onClose={() => setRegisterOpen(false)}
        scope={accountScope}
        allowedRoles={createRoleOptions as ManageableRole[]}
        onSuccess={async ({ member }) => {
          setOffset(0);
          await load(0);
          if (member.role === "member" || member.role === "agent_member" || member.role === "master_member") {
            await openFeeProducts({
              id: member.id,
              email: member.email,
              nama: member.nama,
              role: member.role,
              aktif: true,
              saldo: 0,
              fee_member_rp: 0,
              retail_agent_commission_rp: 0,
              retail_master_commission_rp: 0,
              h2h_agent_commission_rp: 0,
              h2h_master_commission_rp: 0,
              dibuat_pada: "",
            });
          }
        }}
      />

      <EditUserModal
        open={editOpen}
        user={editMember}
        scope={accountScope}
        onClose={() => {
          setEditOpen(false);
          setEditMember(null);
        }}
        onSuccess={async () => {
          setOffset(0);
          await load(0);
        }}
      />

      {/* Adjust modal */}
      {adjOpen && adjMember ? (
        <AppModal
          open={adjOpen}
          onClose={() => {
            if (adjustSubmitting) return;
            setAdjOpen(false);
            setAdjMember(null);
            setNote("");
          }}
          title="Adjust Saldo"
          subtitle={`Member #${adjMember.id} ${adjMember.email}`}
          maxWidthClassName="max-w-xl"
          footer={
            <div className="flex justify-end gap-2">
              <Button
                variant="outline"
                disabled={adjustSubmitting}
                onClick={() => {
                  if (adjustSubmitting) return;
                  setAdjOpen(false);
                  setAdjMember(null);
                  setNote("");
                }}
              >
                Batal
              </Button>
              <Button
                className="h-10 border-0 bg-emerald-700 text-white shadow-[0_12px_30px_-15px_rgba(4,120,87,0.45)] hover:bg-emerald-600"
                onClick={doAdjust}
                disabled={adjustSubmitting}
              >
                {adjustSubmitting ? <Loader2 className="mr-1.5 h-4 w-4 animate-spin" /> : <ShieldCheck className="mr-1.5 h-4 w-4" />}
                {adjustSubmitting ? "Memproses..." : "Proses"}
              </Button>
            </div>
          }
        >
          <div className="space-y-4">
            <div className="rounded-xl border border-emerald-200 bg-emerald-50 p-3">
              <div className="text-base font-semibold text-slate-100">Adjust Saldo</div>
              <div className="mt-1 text-sm text-slate-300">
                Member #{adjMember.id} {adjMember.email}
              </div>
              <div className="mt-1 text-xs text-slate-400">
                Saldo saat ini: <span className="font-medium text-emerald-700">{fmtIDR(adjMember.saldo)}</span>
              </div>
            </div>

            <div className="grid gap-3 rounded-xl border border-white/12 bg-slate-900/55 p-4">
              <div className="grid gap-2">
                <div className="text-xs font-semibold uppercase tracking-wider text-slate-300">Direction</div>
                <div className="flex gap-2">
                  <Button
                    className={`h-10 border-0 ${direction === "credit" ? "bg-emerald-700 text-white shadow-[0_12px_28px_-15px_rgba(4,120,87,0.45)]" : "bg-slate-800/80 text-slate-300 hover:bg-slate-700"}`}
                    onClick={() => setDirection("credit")}
                    disabled={adjustSubmitting}
                  >
                    <ArrowUpCircle className="mr-1.5 h-4 w-4" />
                    Credit
                  </Button>
                  <Button
                    className={`h-10 border-0 ${direction === "debit" ? "bg-linear-to-r from-rose-500 to-orange-500 text-white shadow-[0_12px_28px_-15px_rgba(244,63,94,0.9)]" : "bg-slate-800/80 text-slate-300 hover:bg-slate-700"}`}
                    onClick={() => setDirection("debit")}
                    disabled={adjustSubmitting}
                  >
                    <ArrowDownCircle className="mr-1.5 h-4 w-4" />
                    Debit
                  </Button>
                </div>
              </div>

              <div className="grid gap-2">
                <div className="text-xs font-semibold uppercase tracking-wider text-slate-300">Amount</div>
                <Input
                  value={amountPretty}
                  onChange={(e) => {
                    const norm = normalizeAmountInput(e.target.value);
                    setAmountDigits(norm.digits);
                    setAmountPretty(norm.pretty);
                  }}
                  className="h-11 border-white/15 bg-slate-950/70 text-slate-100 placeholder:text-slate-500"
                  inputMode="numeric"
                  placeholder="misal: 10.000"
                  disabled={adjustSubmitting}
                />
                <div className="text-xs text-muted-foreground">
                  Nilai terkirim: <span className="font-medium">{amountDigits ? fmtIDR(amountDigits) : "0"}</span>
                </div>
              </div>

              <div className="grid gap-2">
                <div className="text-xs font-semibold uppercase tracking-wider text-slate-300">Note (wajib)</div>
                <Input
                  value={note}
                  onChange={(e) => setNote(e.target.value)}
                  className="h-11 border-white/15 bg-slate-950/70 text-slate-100 placeholder:text-slate-500"
                  placeholder="contoh: koreksi manual"
                  disabled={adjustSubmitting}
                />
              </div>
            </div>
          </div>
        </AppModal>
      ) : null}

      {hierarchyOpen && hierarchyMember ? (
        <AppModal
          open={hierarchyOpen}
          onClose={() => {
            if (hierarchySaving) return;
            setHierarchyOpen(false);
            setHierarchyMember(null);
            setHierarchyPreview(null);
          }}
          title={accountScope === "h2h" ? "Atur Upline H2H" : "Atur Upline Retail"}
          subtitle={`Akun #${hierarchyMember.id} ${hierarchyMember.email}`}
          maxWidthClassName="max-w-2xl"
          footer={
            <div className="flex flex-wrap justify-end gap-2">
              <Button
                variant="outline"
                onClick={() => {
                  if (hierarchySaving) return;
                  setHierarchyOpen(false);
                  setHierarchyMember(null);
                  setHierarchyPreview(null);
                }}
                disabled={hierarchySaving}
              >
                Batal
              </Button>
              <Button
                variant="outline"
                onClick={() => void previewHierarchy()}
                disabled={hierarchyLoading || hierarchySaving}
              >
                {hierarchyLoading ? <Loader2 className="mr-1.5 h-4 w-4 animate-spin" /> : <Search className="mr-1.5 h-4 w-4" />}
                Preview
              </Button>
              <Button
                className="border-0 bg-slate-700 text-white hover:bg-slate-600"
                onClick={() => void applyHierarchy(false)}
                disabled={hierarchySaving}
              >
                <Save className="mr-1.5 h-4 w-4" />
                Simpan Relasi
              </Button>
              <Button
                className="border-0 bg-emerald-700 text-white hover:bg-emerald-600"
                onClick={() => void applyHierarchy(true)}
                disabled={hierarchySaving}
              >
                {hierarchySaving ? <Loader2 className="mr-1.5 h-4 w-4 animate-spin" /> : <Wallet className="mr-1.5 h-4 w-4" />}
                Simpan + Komisi Historis
              </Button>
            </div>
          }
        >
          <div className="space-y-4">
            <div className="rounded-xl border border-emerald-300 bg-emerald-50 px-4 py-3 text-sm font-medium text-emerald-900">
              Gunakan modal ini untuk <span className="font-semibold">menambahkan akun ini ke agent atau master yang sudah ada</span>. Pilih upline di bawah, lalu simpan. Jika komisi historis ingin langsung masuk ke dompet upline, gunakan <span className="font-semibold">Simpan + Komisi Historis</span>.
            </div>

            <div className="grid gap-3">
              <div className="text-sm text-slate-200">
                Cari akun existing yang ingin dijadikan downline dari <span className="font-semibold text-slate-100">{hierarchyMember.nama}</span>.
              </div>
              <div className="flex flex-col gap-2 sm:flex-row">
                <Input
                  value={hierarchySearch}
                  onChange={(e) => setHierarchySearch(e.target.value)}
                  className="h-11"
                  placeholder={hierarchySearchPlaceholder(hierarchyMember, accountScope)}
                  disabled={hierarchySaving}
                />
                <Button
                  variant="outline"
                  onClick={() => void searchHierarchyCandidates()}
                  disabled={hierarchySearchLoading || hierarchySaving || !hierarchySearch.trim()}
                >
                  {hierarchySearchLoading ? <Loader2 className="mr-1.5 h-4 w-4 animate-spin" /> : <Search className="mr-1.5 h-4 w-4" />}
                  Cari Akun
                </Button>
              </div>
              <div className="rounded-xl border border-white/12 bg-slate-950/45">
                {hierarchySearchResults.length === 0 ? (
                  <div className="px-4 py-3 text-sm text-slate-400">Belum ada hasil pencarian. Cari berdasarkan nama atau email.</div>
                ) : (
                  <div className="divide-y divide-white/10">
                    {hierarchySearchResults.map((item) => (
                      <div key={item.id} className="flex flex-col gap-3 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
                        <div className="space-y-1">
                          <div className="text-sm font-semibold text-slate-100">#{item.id} {item.nama}</div>
                          <div className="text-xs text-slate-400">{item.email} • {roleLabel(item.role)}</div>
                        </div>
                        <div className="flex flex-wrap gap-2">
                          {hierarchyAllowedTargetRoles(hierarchyMember, accountScope).map((role) => (
                            <Button
                              key={`${item.id}-${role}`}
                              variant={hierarchyTargetMember?.id === item.id && hierarchyTargetRole === role ? "default" : "outline"}
                              className={hierarchyTargetMember?.id === item.id && hierarchyTargetRole === role ? "border-0 bg-emerald-700 text-white" : ""}
                              onClick={() => selectHierarchyTarget(item, role)}
                              disabled={hierarchySaving}
                            >
                              {hierarchyActionLabel(role)}
                            </Button>
                          ))}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>

            <div className="rounded-xl border border-white/12 bg-slate-900/55 p-4 text-sm text-slate-300">
              <div>
                Upline terpilih: <span className="font-semibold text-slate-100">{roleLabel(hierarchyMember.role)}</span>
              </div>
              <div className="mt-1">
                Akun ini akan menjadi downline dari agent/master yang dipilih. Komisi tetap memakai nilai flat per transaksi dari akun upline itu. Harga jual produk tetap mengikuti harga dasar + fee retail/H2H yang sudah berlaku.
              </div>
              {hierarchyTargetMember ? (
                <div className="mt-3 rounded-lg border border-emerald-300 bg-emerald-50 px-3 py-2 text-sm font-medium text-emerald-900">
                  Downline terpilih: <span className="font-semibold">#{hierarchyTargetMember.id} {hierarchyTargetMember.email}</span> akan dijadikan <span className="font-semibold">{roleLabel(hierarchyTargetRole)}</span>.
                </div>
              ) : null}
            </div>

            {hierarchyPreview ? (
              <div className="space-y-3 rounded-xl border border-emerald-300/20 bg-emerald-500/10 p-4">
                <div className="text-sm font-bold text-emerald-900">Preview Histori</div>
                <div className="text-sm text-emerald-50">
                  Total transaksi sukses historis downline ini: <span className="font-semibold">{fmtIDR(hierarchyPreview.transaction_count)}</span>
                </div>
                {hierarchyPreview.derived_master_from_agent ? (
                  <div className="text-xs font-medium text-emerald-800">Master otomatis mengikuti master yang terpasang pada agent terpilih.</div>
                ) : null}
                {hierarchyPreview.agent ? (
                  <div className="rounded-lg border border-emerald-300/20 bg-slate-950/35 p-3 text-sm text-slate-200">
                    <div className="font-semibold text-slate-100">Agent</div>
                    <div className="mt-1">{hierarchyPreview.agent.email}</div>
                    <div className="mt-1 text-xs text-slate-400">
                      Komisi / trx: Rp {fmtIDR(hierarchyPreview.agent.commission_rp)}. Estimasi histori: Rp {fmtIDR(hierarchyPreview.agent.calculated_total)} dari {fmtIDR(hierarchyPreview.agent.calculated_count)} transaksi.
                    </div>
                  </div>
                ) : null}
                {hierarchyPreview.master ? (
                  <div className="rounded-lg border border-emerald-300/20 bg-slate-950/35 p-3 text-sm text-slate-200">
                    <div className="font-semibold text-slate-100">Master</div>
                    <div className="mt-1">{hierarchyPreview.master.email}</div>
                    <div className="mt-1 text-xs text-slate-400">
                      Komisi / trx: Rp {fmtIDR(hierarchyPreview.master.commission_rp)}. Estimasi histori: Rp {fmtIDR(hierarchyPreview.master.calculated_total)} dari {fmtIDR(hierarchyPreview.master.calculated_count)} transaksi.
                    </div>
                  </div>
                ) : null}
              </div>
            ) : null}
          </div>
        </AppModal>
      ) : null}

      {downlineViewerOpen && downlineViewerMember ? (
        <AppModal
          open={downlineViewerOpen}
          onClose={() => {
            setDownlineViewerOpen(false);
            setDownlineViewerMember(null);
            setDownlineViewerItems([]);
          }}
          title={accountScope === "h2h" ? "Daftar Downline H2H" : "Daftar Downline Retail"}
          subtitle={`Akun #${downlineViewerMember.id} ${downlineViewerMember.email}`}
          maxWidthClassName="max-w-2xl"
          footer={
            <div className="flex justify-end">
              <Button
                variant="outline"
                onClick={() => {
                  setDownlineViewerOpen(false);
                  setDownlineViewerMember(null);
                  setDownlineViewerItems([]);
                }}
              >
                Tutup
              </Button>
            </div>
          }
        >
          <div className="space-y-4">
            <div className="rounded-xl border border-emerald-300 bg-emerald-50 px-4 py-3 text-sm font-medium text-emerald-900">
              Menampilkan akun yang saat ini berada langsung di bawah <span className="font-semibold">{downlineViewerMember.nama}</span>.
            </div>

            {downlineViewerLoading ? (
              <div className="text-sm text-slate-300">Memuat downline...</div>
            ) : downlineViewerItems.length === 0 ? (
              <div className="rounded-xl border border-white/12 bg-slate-900/55 px-4 py-3 text-sm text-slate-300">
                Belum ada downline langsung di bawah akun ini.
              </div>
            ) : (
              <div className="space-y-3">
                {downlineViewerItems.map((item) => (
                  <div key={item.id} className="rounded-xl border border-white/12 bg-slate-900/55 px-4 py-3">
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0">
                        <div className="text-sm font-semibold text-slate-100">#{item.id} {item.nama}</div>
                        <div className="truncate text-xs text-slate-400">{item.email}</div>
                        <div className="mt-1 text-xs text-slate-400">
                          Role: {roleLabel(item.role)}
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="text-xs uppercase tracking-wide text-slate-500">{item.aktif ? "Aktif" : "Nonaktif"}</div>
                        <div className="mt-1 text-sm font-semibold text-emerald-700">Rp {fmtIDR(item.saldo)}</div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </AppModal>
      ) : null}

      {commissionOpen && commissionMember ? (
        <AppModal
          open={commissionOpen}
          onClose={() => {
            if (savingCommission) return;
            setCommissionOpen(false);
            setCommissionMember(null);
          }}
          title={commissionTitleForScope(accountScope)}
          subtitle={`Akun #${commissionMember.id} ${commissionMember.email}`}
          maxWidthClassName="max-w-lg"
          footer={
            <div className="flex justify-end gap-2">
              <Button
                variant="outline"
                onClick={() => {
                  if (savingCommission) return;
                  setCommissionOpen(false);
                  setCommissionMember(null);
                }}
                disabled={savingCommission}
              >
                Batal
              </Button>
              <Button
                className="border-0 bg-linear-to-r from-amber-500 to-orange-500 text-white hover:from-amber-400 hover:to-orange-400"
                onClick={doSaveCommission}
                disabled={savingCommission}
              >
                {savingCommission ? "Saving..." : "Simpan"}
              </Button>
            </div>
          }
        >
          <div className="space-y-4">
            <div className="rounded-xl border border-amber-400 bg-amber-50 px-4 py-3 text-sm font-medium text-amber-900">
              Komisi dibayar <span className="font-semibold">flat per transaksi sukses</span>. Setiap akun hanya memakai <span className="font-semibold">satu nilai komisi</span> sesuai role akun saat ini.
            </div>
            <div className="grid gap-4">
              <label className="grid gap-2">
                <span className="text-sm text-slate-200">{commissionLabel(commissionMember, accountScope)}</span>
                <Input
                  value={commissionValue}
                  onChange={(e) => setCommissionValue(e.target.value)}
                  className="h-11"
                  inputMode="numeric"
                  placeholder="contoh: 100"
                />
              </label>
              <div className="text-xs text-slate-400">
                Role akun saat ini: <span className="font-semibold text-slate-200">{roleLabel(commissionMember.role)}</span>. Nilai komisi aktif hanya mengikuti role akun itu.
              </div>
            </div>
          </div>
        </AppModal>
      ) : null}

      {/* Fee Product modal */}
      {feeProductOpen && feeProductMember ? (
        <AppModal
          open={feeProductOpen}
          onClose={() => {
            setFeeProductOpen(false);
            setFeeProductMember(null);
            setFeeProducts([]);
            setFeeCategories([]);
          }}
          title="Fee Kategori Member H2H"
          subtitle={`Member #${feeProductMember.id} ${feeProductMember.email}`}
          maxWidthClassName="max-w-3xl"
          footer={
            <div className="flex justify-end gap-2">
              <Button
                variant="outline"
                onClick={() => {
                  setFeeProductOpen(false);
                  setFeeProductMember(null);
                  setFeeProducts([]);
                  setFeeCategories([]);
                }}
              >
                Tutup
              </Button>
            </div>
          }
        >
          <div className="space-y-4">
            <div className="rounded-xl border border-amber-400 bg-amber-50 px-4 py-3 text-sm font-medium text-amber-900">
              Kategori fee H2H yang dipakai transaksi adalah <span className="font-semibold">DANA, GOPAY, OVO, LINKAJA, SHOPEEPAY, BANK, dan LAINNYA</span>. Retail memakai fee terpisah dan tidak ikut terbaca di jalur ini.
            </div>

            <div className="grid gap-3 rounded-xl border border-emerald-300/20 bg-slate-900/55 p-3 sm:p-4">
              <div className="flex items-center justify-between gap-2">
                <div>
                  <div className="text-sm font-semibold text-slate-100">Fee Kategori H2H</div>
                  <div className="text-xs text-slate-400">Pilih kategori fee H2H yang sesuai untuk transaksi member.</div>
                </div>
              </div>

              <div className="grid gap-2 sm:grid-cols-[1fr_180px_auto_auto] sm:items-end">
                <div className="grid gap-1.5">
                  <div className="text-xs font-semibold uppercase tracking-wider text-slate-300">Kategori</div>
                  <ProductSearchSelect
                    items={H2H_FEE_CATEGORY_OPTIONS}
                    value={feeCategoryID}
                    onChange={setFeeCategoryID}
                    loading={false}
                    placeholder="Pilih kategori H2H"
                    disabled={feeCategorySaving}
                    itemLabelMode="single"
                  />
                </div>

                <div className="grid gap-1.5">
                  <div className="text-xs font-semibold uppercase tracking-wider text-slate-300">Fee Rp</div>
                  <Input
                    value={feeCategoryValue}
                    onChange={(e) => setFeeCategoryValue(e.target.value)}
                    placeholder="contoh: 500"
                    className="h-11 border-white/15 bg-slate-950/70 text-slate-100 placeholder:text-slate-500"
                    inputMode="numeric"
                  />
                </div>

                <label className="flex h-11 items-center gap-2 rounded-lg border border-white/15 bg-slate-950/70 px-3 text-sm text-slate-200">
                  <input
                    type="checkbox"
                    checked={feeCategoryAktif}
                    onChange={(e) => setFeeCategoryAktif(e.target.checked)}
                    className="h-4 w-4 rounded border-white/20 bg-slate-900 text-emerald-400"
                  />
                  Aktif
                </label>

                <Button
                  className="h-11 border-0 bg-emerald-700 text-white shadow-[0_12px_30px_-15px_rgba(4,120,87,0.45)] hover:bg-emerald-600"
                  onClick={doUpsertFeeCategory}
                  disabled={feeCategorySaving}
                >
                  {feeCategorySaving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
                  Simpan
                </Button>
              </div>

              <div className="overflow-x-auto rounded-xl border border-white/10">
                <table className="min-w-full divide-y divide-white/10 text-sm">
                  <thead className="bg-slate-950/70 text-left text-xs uppercase tracking-wider text-slate-400">
                    <tr>
                      <th className="px-3 py-2">Kategori</th>
                      <th className="px-3 py-2">Fee Rp</th>
                      <th className="px-3 py-2">Aktif</th>
                      <th className="px-3 py-2 text-right">Aksi</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-white/8 bg-slate-950/35 text-slate-200">
                    {feeCategoryLoading ? (
                      <tr>
                        <td colSpan={4} className="px-3 py-6 text-center text-slate-400">
                          Memuat fee kategori...
                        </td>
                      </tr>
                    ) : feeCategories.length === 0 ? (
                      <tr>
                        <td colSpan={4} className="px-3 py-6 text-center text-slate-400">
                          Belum ada fee kategori H2H.
                        </td>
                      </tr>
                    ) : (
                      feeCategories.map((row) => (
                        <tr key={row.id}>
                          <td className="px-3 py-2">{row.kategori_nama}</td>
                          <td className="px-3 py-2">{fmtIDR(row.fee_rp ?? 0)}</td>
                          <td className="px-3 py-2">{row.aktif ? "aktif" : "nonaktif"}</td>
                          <td className="px-3 py-2 text-right">
                            <Button
                              variant="outline"
                              size="sm"
                              className="border-rose-400 bg-rose-50 text-rose-800 hover:bg-rose-100"
                              onClick={() => void doDeleteFeeCategory(row)}
                              disabled={feeCategoryDeletingID === row.id}
                            >
                              {feeCategoryDeletingID === row.id ? <Loader2 className="h-4 w-4 animate-spin" /> : <Trash2 className="h-4 w-4" />}
                            </Button>
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </AppModal>
      ) : null}

      {/* Edit Fee modal */}
      {feeOpen && feeMember ? (
        <AppModal
          open={feeOpen}
          onClose={() => {
            setFeeOpen(false);
            setFeeMember(null);
            setNewFee("");
          }}
          title="Edit Default Fee Member"
          subtitle={`Member #${feeMember.id} ${feeMember.email}`}
          maxWidthClassName="max-w-lg"
          footer={
            <div className="flex justify-end gap-2">
              <Button
                variant="outline"
                onClick={() => {
                  setFeeOpen(false);
                  setFeeMember(null);
                  setNewFee("");
                }}
                disabled={savingFee}
              >
                Batal
              </Button>
              <Button
                className="border-0 bg-emerald-700 text-white hover:bg-emerald-600"
                onClick={doSetFee}
                disabled={savingFee}
              >
                {savingFee ? "Saving..." : "Simpan"}
              </Button>
            </div>
          }
        >
          <div className="grid gap-2">
            <div className="text-sm">Fee Default Baru (Rp)</div>
            <Input
              value={newFee}
              onChange={(e) => setNewFee(e.target.value)}
              className="h-11"
              inputMode="numeric"
            />
            <div className="text-xs text-muted-foreground">Preview: Rp {fmtIDR(newFee || "0")}</div>
          </div>
        </AppModal>
      ) : null}
    </>
  );
}
