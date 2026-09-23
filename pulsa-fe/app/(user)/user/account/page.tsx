import Link from "next/link";
import { redirect } from "next/navigation";
import {
  Bell,
  BriefcaseBusiness,
  ClipboardList,
  ChevronRight,
  FileText,
  HelpCircle,
  LockKeyhole,
  Mail,
  Phone,
  ShieldCheck,
  UserPlus,
  UserRound,
  UsersRound,
} from "lucide-react";
import { getAppServerSession } from "@/lib/server-auth";
import { getUserProfile } from "@/lib/api.auth";
import { getInitials } from "@/components/user/helpers";
import type { UserSession } from "@/components/user/types";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import { UserLogoutButton } from "@/components/user/UserLogoutButton";
import { UserProfilePhotoUploader } from "@/components/user/UserProfilePhotoUploader";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

function normalizeRole(role?: string | null) {
  const value = String(role || "").trim().toLowerCase();
  return ["analyst", "operator", "operator kredit", "operator_kredit", "operator_credit", "operator-credit"].includes(value) ? "analis" : value;
}

function panelPathByRole(role: string) {
  if (role === "admin" || role === "staff") return "/dashboard/admin";
  if (role === "auditor") return "/dashboard/auditor";
  if (role === "member" || role === "agent_member" || role === "master_member") return "/dashboard/member";
  if (role === "analis") return "/dashboard/master/operator";
  if (role === "master" || role === "marketing") return "/dashboard/master";
  if (role === "operator_trx") return "/dashboard/operator";
  if (role === "operator_wallet") return "/dashboard/wallet";
  return "/user";
}

function panelDescriptionByRole(role: string) {
  if (role === "admin" || role === "staff") return "Masuk ke panel admin";
  if (role === "auditor") return "Masuk ke panel audit";
  if (role === "analis") return "Masuk ke panel operator kredit";
  if (role === "marketing") return "Menu kerja marketing tersedia di Akun";
  if (role === "master") return "Masuk ke panel master";
  if (role === "operator_trx") return "Masuk ke panel transaksi";
  if (role === "operator_wallet") return "Masuk ke panel wallet";
  if (role === "member" || role === "agent_member" || role === "master_member") return "Masuk ke panel H2H";
  return "Masuk ke aplikasi agent";
}

export default async function UserAccountPage() {
  const session = (await getAppServerSession()) as SessionShape | null;

  if (!session?.backendToken) {
    redirect("/login");
  }

  const user = session.user ?? null;
  const profile = session.backendToken ? await getUserProfile(session.backendToken) : null;
  const displayName = profile?.nama || user?.name || "User";
  const displayEmail = profile?.email || user?.email || "-";
  const profileWithPhone = profile as typeof profile & { phone?: string; no_hp?: string; nomor_hp?: string; telepon?: string };
  const phone = profileWithPhone?.phone || profileWithPhone?.no_hp || profileWithPhone?.nomor_hp || profileWithPhone?.telepon || "-";
  const username = displayEmail !== "-" ? `@${displayEmail.split("@")[0]}` : "@pulsakilat";
  const initials = getInitials(displayName, displayEmail);
  const profilePhotoURL = profile?.profile_photo_url || user?.image || "";
  const role = normalizeRole(profile?.role || user?.role);
  const canManageRetailNetwork = role === "master" || role === "agent";
  const canOpenWorkPanel = role !== "user" && role !== "agent" && role !== "marketing";

  const personalItems = [
    {
      label: "Nama lengkap",
      value: displayName,
      icon: UserRound,
    },
    {
      label: "Nomor handphone",
      value: phone,
      icon: Phone,
    },
    {
      label: "Email / Gmail",
      value: displayEmail,
      icon: Mail,
    },
  ];

  const settingItems = [
    ...(role === "marketing"
      ? [
          { href: "/user/account/tambah-agent", label: "Tambah Agent", desc: "Daftarkan agent baru dari lapangan", icon: UserPlus },
          { href: "/user/account/pengajuan-agent", label: "Pengajuan & Dokumen", desc: "Pantau dokumen pengajuan agent", icon: ClipboardList },
          { href: "/user/account/agent-binaan", label: "Agent Binaan", desc: "Lihat saldo dan aktivitas agent", icon: UsersRound },
        ]
      : []),
    ...(canOpenWorkPanel
      ? [
          {
            href: panelPathByRole(role),
            label: "Panel",
            desc: panelDescriptionByRole(role),
            icon: BriefcaseBusiness,
          },
        ]
      : []),
    ...(canManageRetailNetwork
      ? [
          {
            href: "/user/account/downline",
            label: role === "agent" ? "Tambah Member" : "Jaringan Retail",
            desc: role === "master" ? "Kelola agent dan user bawahan" : "Tambahkan member/user bawahan",
            icon: UsersRound,
          },
        ]
      : []),
    ...(role !== "marketing"
      ? [
          {
            href: "/user/account/security",
            label: "Keamanan Akun",
            desc: "Ganti password akun",
            icon: LockKeyhole,
          },
          {
            href: "/user/account",
            label: "Notifikasi",
            desc: "Atur informasi transaksi",
            icon: Bell,
          },
          {
            href: "/user/account",
            label: "Pusat Bantuan",
            desc: "FAQ dan layanan pelanggan",
            icon: HelpCircle,
          },
        ]
      : []),
    {
      href: "/kebijakan-privasi?from=account",
      label: "Syarat & Kebijakan",
      desc: "Ketentuan penggunaan Isiloka",
      icon: FileText,
    },
  ];

  return (
    <main className="min-h-screen bg-[#f3f7f5] pb-24">
      <section className="relative overflow-hidden rounded-b-[32px] bg-[linear-gradient(135deg,#052e26_0%,#047857_58%,#84cc16_145%)] px-4 pb-8 pt-7 text-white shadow-[0_20px_44px_rgba(4,120,87,0.24)]">
        <div className="pointer-events-none absolute -left-14 -top-16 h-40 w-40 rounded-full border border-white/10 bg-white/8" />
        <div className="pointer-events-none absolute -right-10 top-7 h-32 w-32 rounded-full bg-white/10" />
        <div className="mx-auto flex w-full max-w-md flex-col items-center text-center">
          <UserProfilePhotoUploader
            name={displayName}
            email={displayEmail}
            phone={phone}
            initials={initials}
            profilePhotoURL={profilePhotoURL}
          />
          <h1 className="mt-4 max-w-full truncate text-lg font-black tracking-tight">{displayName}</h1>
          <p className="mt-0.5 max-w-full truncate text-[11px] font-bold text-white/75">{username}</p>
          <div className="mt-4 inline-flex items-center gap-2 rounded-full bg-white/12 px-3 py-1.5 text-[10px] font-black text-white ring-1 ring-white/15">
            <ShieldCheck className="h-3.5 w-3.5" strokeWidth={2.4} />
            Akun Isiloka aktif
          </div>
        </div>
      </section>

      <div className="mx-auto -mt-4 w-full max-w-md space-y-3.5 px-4">
        <section className="overflow-hidden rounded-[22px] border border-emerald-950/5 bg-white shadow-[0_16px_36px_rgba(6,78,59,0.08)]">
          <div className="flex items-center justify-between border-b border-slate-100 px-4 py-3.5">
            <h2 className="text-sm font-black text-slate-950">Informasi Pribadi</h2>
            <Link href="/user/account/edit" className="text-[10px] font-black text-[#087e8b]">Edit</Link>
          </div>
          <div className="divide-y divide-slate-100">
            {personalItems.map((item) => {
              const Icon = item.icon;
              return (
                <div key={item.label} className="flex items-center gap-3 px-4 py-3.5">
                  <Icon className="h-5 w-5 shrink-0 text-[#087e8b]" strokeWidth={1.9} />
                  <div className="min-w-0">
                    <p className="text-[10px] font-semibold text-slate-400">{item.label}</p>
                    <p className="mt-0.5 truncate text-xs font-black text-slate-950">{item.value}</p>
                  </div>
                </div>
              );
            })}
          </div>
        </section>

        <section className="overflow-hidden rounded-[22px] border border-emerald-950/5 bg-white shadow-[0_16px_36px_rgba(6,78,59,0.08)]">
          <div className="divide-y divide-slate-100">
            {settingItems.map((item) => {
              const Icon = item.icon;
              return (
                <Link
                  key={item.label}
                  href={item.href}
                  className="flex items-center gap-3 px-4 py-3.5 transition hover:bg-emerald-50/50"
                >
                  <span className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-emerald-50 text-[#087e8b]">
                    <Icon className="h-5 w-5" strokeWidth={2.2} />
                  </span>
                  <span className="min-w-0 flex-1">
                    <span className="block text-xs font-black text-slate-950">{item.label}</span>
                    <span className="mt-0.5 block truncate text-[10px] font-semibold text-slate-400">{item.desc}</span>
                  </span>
                  <ChevronRight className="h-4 w-4 shrink-0 text-slate-500" />
                </Link>
              );
            })}
          </div>
        </section>

        <section className="rounded-[18px] border border-rose-200 bg-white p-3 shadow-[0_10px_24px_rgba(15,23,42,0.04)]">
          <UserLogoutButton className="h-12 w-full rounded-2xl border border-rose-200 bg-white text-xs font-black text-rose-600 shadow-none hover:bg-rose-50 hover:text-rose-700" />
        </section>

        <p className="pt-2 text-center text-[10px] font-semibold text-slate-400">Isiloka versi 1.0.0</p>
      </div>

      <UserBottomNav />
    </main>
  );
}
