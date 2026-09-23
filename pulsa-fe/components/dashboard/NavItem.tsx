"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  Activity,
  AlertTriangle,
  Archive,
  ArrowUpCircle,
  BarChart3,
  Camera,
  CheckCircle2,
  CircleDollarSign,
  ClipboardList,
  CreditCard,
  Download,
  FileCheck2,
  FilePlus2,
  FileText,
  Image,
  Landmark,
  LayoutDashboard,
  Package,
  PlugZap,
  ReceiptText,
  RefreshCcw,
  ShieldCheck,
  Tags,
  UserPlus,
  Users,
  Wallet,
  WalletCards,
  XCircle,
} from "lucide-react";

type Props = {
  href: string;
  label: string;
  onClick?: () => void;
  variant?: "light" | "dark";
};

const iconByHref = {
  "/dashboard/admin": LayoutDashboard,
  "/dashboard/admin/komisi": BarChart3,
  "/dashboard/admin/master/members": Users,
  "/dashboard/admin/pemantauan-tim": Activity,
  "/dashboard/admin/kredit/pengajuan": CreditCard,
  "/dashboard/admin/transaksi/aplikasi": ReceiptText,
  "/dashboard/admin/transaksi/aplikasi/provider": PlugZap,
  "/dashboard/admin/transaksi/guest-refund": RefreshCcw,
  "/dashboard/admin/deposits": CircleDollarSign,
  "/dashboard/admin/retail-withdraws": Download,
  "/dashboard/admin/bank": Landmark,
  "/dashboard/admin/master/produk": Package,
  "/dashboard/admin/master/fee-kategori-aplikasi": Tags,
  "/dashboard/admin/master/iklan": Image,
  "/dashboard/admin/integrasi/pulsa24jam": PlugZap,
  "/dashboard/operator": Activity,
  "/dashboard/wallet": WalletCards,
  "/dashboard/auditor": ShieldCheck,
  "/dashboard/master": LayoutDashboard,
  "/dashboard/master/tambah-agent": UserPlus,
  "/dashboard/master/input-pinjaman-manual": FileText,
  "/dashboard/master/pinjaman": Camera,
  "/dashboard/master/akun-agent": Users,
  "/dashboard/master/profil-agent": Activity,
  "/dashboard/master/riwayat-pinjaman": ReceiptText,
  "/dashboard/master/laporan": FileCheck2,
  "/dashboard/master/analis": ShieldCheck,
  "/dashboard/master/analis/antrean": ClipboardList,
  "/dashboard/master/analis/monitor-pelunasan": WalletCards,
  "/dashboard/master/analis/bukti-pelunasan": FileCheck2,
  "/dashboard/master/analis/penolakan-catatan": XCircle,
  "/dashboard/master/analis/arsip-keputusan": Archive,
  "/dashboard/master/operator": ShieldCheck,
  "/dashboard/master/operator/tambah-marketing": UserPlus,
  "/dashboard/master/operator/input-data-agent": FilePlus2,
  "/dashboard/master/operator/kenaikan-limit": ArrowUpCircle,
  "/dashboard/master/operator/pembayaran-kredit": ReceiptText,
  "/dashboard/master/operator/monitor-pelunasan": WalletCards,
  "/dashboard/master/operator/transaksi-agent": ReceiptText,
  "/dashboard/master/operator/konter-tidak-transaksi": AlertTriangle,
  "/dashboard/master/operator/arsip-keputusan": Archive,
  "/dashboard/master/riwayat-acc-analis": FileText,
} as const;

export function NavItem({ href, label, onClick, variant = "light" }: Props) {
  const pathname = usePathname();
  const hrefPath = href.split("?")[0] || href;
  const Icon = iconByHref[hrefPath as keyof typeof iconByHref];
  const exactOnlyHrefs = new Set([
    "/dashboard/admin",
    "/dashboard/admin/komisi",
    "/dashboard/admin/transaksi/aplikasi",
    "/dashboard/member",
    "/dashboard/operator",
    "/dashboard/wallet",
    "/dashboard/master",
    "/dashboard/master/operator",
  ]);
  const active = exactOnlyHrefs.has(hrefPath) ? pathname === hrefPath : pathname === hrefPath || pathname.startsWith(hrefPath + "/");
  const isDark = variant === "dark";

  return (
    <Link
      href={href}
      onClick={onClick}
      className={[
        "group flex min-h-12 items-center justify-between gap-3 rounded-2xl border px-3 py-2.5 text-[13px] font-black uppercase tracking-[0.08em] outline-none transition focus-visible:ring-4 focus-visible:ring-emerald-200",
        active
          ? isDark
            ? "border-white bg-white !text-[#13464b] shadow-[0_14px_28px_rgba(0,0,0,0.16)]"
            : "border-[#13464b] bg-white text-[#13464b] shadow-[0_10px_20px_rgba(6,78,59,0.10)]"
          : isDark
            ? "border-white/12 bg-white/8 !text-emerald-50 hover:border-lime-200/55 hover:bg-white/16 hover:!text-white"
            : "border-transparent text-slate-800 hover:border-slate-300 hover:bg-white hover:text-slate-950",
      ].join(" ")}
      aria-current={active ? "page" : undefined}
    >
      <span className="flex min-w-0 items-center gap-3">
        {Icon ? (
          <span
            className={[
              "grid h-8 w-8 shrink-0 place-items-center rounded-xl border transition",
              active
                ? isDark
                  ? "border-emerald-100 bg-emerald-50 text-emerald-800"
                  : "border-emerald-200 bg-emerald-50 text-emerald-700"
                : isDark
                  ? "border-lime-200/25 bg-emerald-950/35 text-lime-200 group-hover:border-lime-100/55 group-hover:bg-emerald-900/45"
                  : "border-slate-200 bg-white text-emerald-700",
            ].join(" ")}
          >
            <Icon className="h-4.5 w-4.5" strokeWidth={2.4} />
          </span>
        ) : null}
        <span className={`min-w-0 truncate ${isDark && !active ? "!text-emerald-50" : ""}`}>{label}</span>
      </span>
      {active ? (
        <span className={`inline-flex h-6 w-6 items-center justify-center rounded-full ${isDark ? "bg-[#13464b] text-white" : "bg-[#13464b] text-white"}`}>
          <CheckCircle2 className="h-4 w-4" />
        </span>
      ) : null}
    </Link>
  );
}
