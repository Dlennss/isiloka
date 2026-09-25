import Image from "next/image";
import Link from "next/link";
import type { LucideIcon } from "lucide-react";
import type { UserAppOrder } from "@/components/user/types";
import {
  Bell,
  ChevronRight,
  CreditCard,
  Eye,
  FileText,
  Gift,
  Grid2X2,
  PhoneCall,
  Plus,
  ReceiptText,
  Send,
  ShieldCheck,
  Smartphone,
  Sun,
  UserRound,
  Wallet,
  Wifi,
  Zap,
} from "lucide-react";

const services = [
  { href: "/pulsa-data", label: "Pulsa & Data", icon: Smartphone, tone: "bg-[#dcfff4] text-[#12b98a]" },
  { href: "/listrik/token", label: "Token Listrik", icon: Zap, tone: "bg-[#fff1cc] text-[#ffac18]" },
  { href: "/ewallet", label: "E-Wallet", icon: CreditCard, tone: "bg-[#eee3ff] text-[#7654e8]" },
  { href: "/kategori", label: "Tagihan", icon: FileText, tone: "bg-[#e2f4ff] text-[#269be8]" },
  { href: "/internet-pascabayar", label: "Paket Internet", icon: Wifi, tone: "bg-[#ffe3ed] text-[#ee4770]" },
  { href: "/kategori", label: "Lainnya", icon: Grid2X2, tone: "bg-[#dffff2] text-[#15b884]" },
];

const userServices = [
  { href: "/user/pulsa-data", label: "Pulsa & Data", icon: Smartphone, tone: "bg-[#dcfff4] text-[#12b98a]" },
  { href: "/user/listrik/token", label: "Token Listrik", icon: Zap, tone: "bg-[#fff1cc] text-[#ffac18]" },
  { href: "/user/ewallet", label: "E-Wallet", icon: CreditCard, tone: "bg-[#eee3ff] text-[#7654e8]" },
  { href: "/user/kategori", label: "Tagihan", icon: FileText, tone: "bg-[#e2f4ff] text-[#269be8]" },
  { href: "/user/kategori", label: "Paket Internet", icon: Wifi, tone: "bg-[#ffe3ed] text-[#ee4770]" },
  { href: "/user/kategori", label: "Lainnya", icon: Grid2X2, tone: "bg-[#dffff2] text-[#15b884]" },
];

type HomeUser =
  | {
      isLoggedIn: true;
      name: string;
      email: string;
      image?: string | null;
      balance: number;
    }
  | {
      isLoggedIn: false;
    };

type ActivityItem = {
  title: string;
  detail: string;
  amount: string;
  time: string;
  icon: LucideIcon;
  tone: string;
};

function formatIDR(value: number) {
  return `Rp ${new Intl.NumberFormat("id-ID").format(Number.isFinite(value) ? value : 0)}`;
}

function shortName(value?: string) {
  const clean = String(value || "").trim();
  if (!clean) return "Pengguna";
  return clean.split(/\s+/)[0] || clean;
}

function initials(name?: string, email?: string) {
  const source = String(name || email || "U").trim();
  const parts = source.includes("@") ? [source[0]] : source.split(/\s+/);
  return parts.slice(0, 2).map((part) => part[0]?.toUpperCase()).join("") || "U";
}

function formatOrderTime(value?: string | null) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return new Intl.DateTimeFormat("id-ID", {
    day: "2-digit",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

function statusLabel(status?: string) {
  const value = String(status || "").toLowerCase();
  if (value === "success" || value === "paid" || value === "completed") return "Berhasil";
  if (value === "failed" || value === "cancelled" || value === "expired") return "Gagal";
  if (value === "pending") return "Diproses";
  return status || "Transaksi";
}

function mapOrderToActivity(order: UserAppOrder): ActivityItem {
  const name = order.produk_nama_snapshot || order.produk_sku_snapshot || "Transaksi";
  const status = statusLabel(order.status);
  return {
    title: name,
    detail: `${status}${order.dest ? ` • ${order.dest}` : ""}`,
    amount: `- ${formatIDR(Number(order.harga_final || order.nominal || 0))}`,
    time: formatOrderTime(order.dibuat_pada),
    icon: order.status === "success" ? Smartphone : ReceiptText,
    tone: order.status === "success" ? "bg-[#dcfff4] text-[#12b98a]" : "bg-[#e2f4ff] text-[#269be8]",
  };
}

function SectionTitle({ title, href }: { title: string; href: string }) {
  return (
    <div className="mb-3 flex items-center justify-between">
      <h2 className="text-[20px] font-black leading-tight text-[#071d38]">{title}</h2>
      <Link
        href={href}
        prefetch={false}
        className="flex items-center gap-1 text-[14px] font-extrabold text-[#079c7f]"
      >
        Lihat Semua
        <ChevronRight className="h-5 w-5" strokeWidth={3} />
      </Link>
    </div>
  );
}

function ServiceCard({
  href,
  label,
  icon: Icon,
  tone,
}: {
  href: string;
  label: string;
  icon: LucideIcon;
  tone: string;
}) {
  return (
    <Link
      href={href}
      prefetch={false}
      className="flex min-h-[112px] flex-col items-center justify-center rounded-[18px] bg-white px-2 text-center text-[#0a1e38]! shadow-[0_13px_28px_rgba(15,78,81,0.09)] visited:text-[#0a1e38]!"
    >
      <span className={`grid h-12 w-12 place-items-center rounded-[16px] ${tone}`}>
        <Icon className="h-7 w-7" strokeWidth={2.25} />
      </span>
      <span className="mt-2 block min-h-[32px] text-[13px] font-black leading-tight">{label}</span>
    </Link>
  );
}

export function GuestConceptHome({
  user = { isLoggedIn: false },
  recentOrders = [],
  userMode = false,
}: {
  user?: HomeUser;
  recentOrders?: UserAppOrder[];
  userMode?: boolean;
}) {
  const displayName = user.isLoggedIn ? shortName(user.name || user.email) : "";
  const activities = recentOrders.slice(0, 3).map(mapOrderToActivity);
  const serviceItems = userMode ? userServices : services;
  const transactionHref = userMode ? "/user/transaksi" : "/transaksi";
  const categoryHref = userMode ? "/user/kategori" : "/kategori";
  const primaryOrderHref = userMode ? "/user/pulsa-data" : "/pulsa-data";

  return (
    <main className="isiloka-home mx-auto min-h-dvh w-full max-w-[430px] overflow-hidden bg-[radial-gradient(circle_at_50%_0%,#fafffe_0%,#f1fffb_44%,#e7f8f4_100%)] px-4 pb-28 pt-4 text-[#071d38] shadow-[0_20px_70px_rgba(8,91,84,0.14)] md:rounded-[34px]">
      <div className="mx-auto w-full max-w-[398px]">
        <header className="flex items-start justify-between gap-3 pt-1">
          <div className="flex items-center gap-3">
            <Image
              src="/isiloka-concept/logo_symbol.png"
              alt="Isiloka"
              width={48}
              height={48}
              className="h-12 w-12 rounded-xl"
              priority
            />
            <div className="min-w-0 leading-none">
              <div className="text-[25px] font-black tracking-normal text-[#084f55]">Isiloka</div>
              <div className="mt-1 whitespace-nowrap text-[10px] font-extrabold tracking-[0.03em] text-[#536a7d]">ISI HARI, DARI SINI</div>
            </div>
          </div>

          <div className="flex shrink-0 items-center gap-2">
            <Link
              href={user.isLoggedIn ? transactionHref : "/login"}
              prefetch={false}
              aria-label="Notifikasi"
              className="relative grid h-11 w-11 place-items-center rounded-[17px] bg-white text-[#087e8b]! shadow-[0_14px_32px_rgba(12,68,75,0.11)] visited:text-[#087e8b]!"
            >
              <Bell className="h-5 w-5" strokeWidth={2.4} />
              <span className="absolute right-2.5 top-2.5 h-2.5 w-2.5 rounded-full border-2 border-white bg-[#ff315f]" />
            </Link>
            <Link
              href={user.isLoggedIn ? "/user/account" : "/login"}
              prefetch={false}
              aria-label={user.isLoggedIn ? "Akun" : "Masuk"}
              className={
                user.isLoggedIn
                  ? "grid h-11 w-11 shrink-0 place-items-center overflow-hidden rounded-full bg-white text-sm font-black text-[#087e8b]! shadow-[0_14px_32px_rgba(12,68,75,0.09)] visited:text-[#087e8b]!"
                  : "flex h-11 shrink-0 items-center gap-1.5 rounded-[17px] bg-white px-3 text-[13px] font-black text-[#087e8b]! shadow-[0_14px_32px_rgba(12,68,75,0.09)] visited:text-[#087e8b]!"
              }
            >
              {user.isLoggedIn && user.image ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img src={user.image} alt="" className="h-full w-full object-cover" />
              ) : user.isLoggedIn ? (
                <span>{initials(user.name, user.email)}</span>
              ) : (
                <>
                  <UserRound className="h-4 w-4" strokeWidth={2.4} />
                  <span>Masuk</span>
                </>
              )}
            </Link>
          </div>
        </header>

        <section className="mt-6">
          <h1 className="text-[28px] font-black leading-none text-[#061d38]">
            {user.isLoggedIn ? `Halo, ${displayName}!` : "Halo!"}
          </h1>
          <p className="mt-2 flex items-center gap-2 text-[17px] font-semibold leading-tight text-[#62728b]">
            {user.isLoggedIn ? "Semoga harimu menyenangkan" : "Masuk untuk melihat saldo dan aktivitas akun"}
            <Sun className="h-5 w-5 fill-[#ffbf24] text-[#ffbf24]" strokeWidth={2.3} />
          </p>
        </section>

        <section className="relative mt-5 overflow-hidden rounded-[24px] bg-[linear-gradient(135deg,#02786e_0%,#0aa889_52%,#7ee486_100%)] p-4 text-white shadow-[0_18px_40px_rgba(0,125,105,0.24)]">
          <div className="absolute -right-10 top-4 h-36 w-64 rotate-[-12deg] rounded-full bg-white/13" />
          <div className="absolute -bottom-14 right-6 h-32 w-56 rounded-full bg-[#0b745f]/18" />
          <div className="relative z-10">
            <div className="flex items-start justify-between gap-3">
              <div className="flex items-center gap-3">
                <Wallet className="h-8 w-8 text-white" strokeWidth={2.2} />
                <div className="text-[16px] font-extrabold">Saldo Isiloka</div>
              </div>
              <div className="flex items-center gap-1.5 rounded-full bg-white/72 px-3 py-1 text-[12px] font-bold text-[#0a695a]">
                <ShieldCheck className="h-4 w-4 text-[#0bb77f]" />
                Aman & Terpercaya
              </div>
            </div>

            <div className="mt-4 flex items-center gap-3">
              <div className="text-[38px] font-black leading-none tracking-normal">
                {user.isLoggedIn ? formatIDR(user.balance) : "Masuk dulu"}
              </div>
              <button
                type="button"
                aria-label="Lihat saldo"
                className="grid h-9 w-9 place-items-center rounded-full bg-white/16 text-white backdrop-blur"
              >
                <Eye className="h-5 w-5" strokeWidth={2.4} />
              </button>
            </div>

            <div className="mt-7 grid grid-cols-3 gap-2">
              <Link href={user.isLoggedIn ? "/user/account/topup" : "/login"} prefetch={false} className="flex h-[60px] items-center justify-center gap-2 rounded-xl bg-white px-2 text-[13px] font-black text-[#075862]! shadow-[0_10px_24px_rgba(6,77,70,0.12)] visited:text-[#075862]!">
                <span className="grid h-8 w-8 shrink-0 place-items-center rounded-full bg-[#0a7d76] text-white">
                  <Plus className="h-5 w-5" strokeWidth={3} />
                </span>
                <span>Isi Saldo</span>
              </Link>
              <Link href={user.isLoggedIn ? "/user/transfer-bank" : "/login"} prefetch={false} className="flex h-[60px] items-center justify-center gap-2 rounded-xl bg-white px-2 text-[13px] font-black text-[#075862]! shadow-[0_10px_24px_rgba(6,77,70,0.12)] visited:text-[#075862]!">
                <Send className="h-7 w-7 shrink-0 fill-[#0a7d76] text-[#0a7d76]" strokeWidth={1.8} />
                <span>Transfer</span>
              </Link>
              <Link href={user.isLoggedIn ? transactionHref : "/login"} prefetch={false} className="flex h-[60px] items-center justify-center gap-2 rounded-xl bg-white px-2 text-[13px] font-black text-[#075862]! shadow-[0_10px_24px_rgba(6,77,70,0.12)] visited:text-[#075862]!">
                <ReceiptText className="h-7 w-7 shrink-0 fill-[#0a7d76] text-white" strokeWidth={2.2} />
                <span>Riwayat</span>
              </Link>
            </div>
          </div>
        </section>

        <section className="relative mt-4 overflow-hidden rounded-[23px] bg-[#def8f2] p-5 shadow-[0_16px_36px_rgba(22,102,95,0.10)]">
          <div className="absolute -bottom-10 left-28 h-36 w-56 rounded-full bg-[#b7efd9]" />
          <div className="absolute -right-8 top-4 h-32 w-32 rounded-full bg-white/45" />
          <div className="relative z-10 min-h-[186px]">
            <div className="relative z-20 max-w-[176px] min-[390px]:max-w-[184px]">
              <h2 className="text-[20px] font-black leading-[1.18] text-[#084f55] min-[390px]:text-[21px]">Semua Kebutuhan Dalam Satu Aplikasi</h2>
              <p className="mt-2 text-[13px] font-semibold leading-snug text-[#5a6f81] min-[390px]:text-[14px]">
                Isi pulsa, paket data, token listrik dan berbagai pembayaran lainnya.
              </p>
              <Link
                href={primaryOrderHref}
                prefetch={false}
                className="mt-4 inline-flex h-11 items-center gap-2 rounded-full bg-[#079c7f] px-5 text-[15px] font-extrabold text-white! shadow-[0_12px_24px_rgba(0,141,111,0.22)] visited:text-white!"
              >
                Isi Sekarang
                <ChevronRight className="h-5 w-5" strokeWidth={3} />
              </Link>
            </div>
            <div className="absolute bottom-2 -right-2 z-10 h-[150px] w-[158px] min-[390px]:right-0 min-[390px]:h-[164px] min-[390px]:w-[178px]">
              <div className="absolute bottom-0 right-7 h-[132px] w-[78px] rotate-[7deg] rounded-[18px] border-[5px] border-[#07515a] bg-white shadow-[0_18px_32px_rgba(7,81,90,0.20)]">
                <div className="mx-auto mt-2 h-1 w-7 rounded-full bg-[#07515a]/20" />
                <div className="mt-7 grid place-items-center text-center">
                  <Image
                    src="/isiloka-concept/logo_symbol.png"
                    alt=""
                    width={34}
                    height={34}
                    className="h-8 w-8 rounded-lg"
                  />
                  <div className="mt-1 text-[13px] font-black leading-none text-[#07515a]">Isiloka</div>
                  <div className="mt-0.5 text-[4px] font-black uppercase text-[#07515a]/70">Isi hari, dari sini</div>
                </div>
              </div>
              <span className="absolute left-8 top-8 grid h-10 w-10 rotate-[-10deg] place-items-center rounded-xl bg-[#4fb7f4] text-white shadow-[0_10px_18px_rgba(34,138,203,0.20)]">
                <PhoneCall className="h-5 w-5" strokeWidth={2.4} />
              </span>
              <span className="absolute right-0 top-3 grid h-11 w-11 rotate-[8deg] place-items-center rounded-xl bg-[#ffd45a] text-white shadow-[0_10px_18px_rgba(235,174,31,0.22)]">
                <Zap className="h-6 w-6 fill-white" strokeWidth={2.1} />
              </span>
              <span className="absolute bottom-7 left-5 grid h-10 w-10 rotate-[-8deg] place-items-center rounded-xl bg-[#f05f89] text-white shadow-[0_10px_18px_rgba(217,70,120,0.20)]">
                <Wifi className="h-5 w-5" strokeWidth={2.5} />
              </span>
              <span className="absolute bottom-2 right-1 grid h-10 w-10 rotate-[9deg] place-items-center rounded-xl bg-[#31c781] text-white shadow-[0_10px_18px_rgba(30,154,101,0.22)]">
                <Wallet className="h-5 w-5" strokeWidth={2.3} />
              </span>
            </div>
          </div>
        </section>

        <div className="mt-2 flex justify-center gap-1.5">
          <span className="h-2.5 w-4 rounded-full bg-[#0aa889]" />
          <span className="h-2.5 w-2.5 rounded-full bg-[#b8ddd8]" />
          <span className="h-2.5 w-2.5 rounded-full bg-[#b8ddd8]" />
        </div>

        <section className="mt-5">
          <SectionTitle title="Layanan Favorit" href={categoryHref} />
          <div className="grid grid-cols-3 gap-3">
            {serviceItems.map((service) => (
              <ServiceCard key={service.label} {...service} />
            ))}
          </div>
        </section>

        <section className="mt-5 rounded-[22px] bg-white px-4 py-4 shadow-[0_16px_36px_rgba(15,78,81,0.10)]">
          <SectionTitle title="Aktivitas Terakhir" href={transactionHref} />
          {user.isLoggedIn && activities.length > 0 ? (
            <div className="divide-y divide-[#e5eef0]">
              {activities.map((activity) => {
                const Icon = activity.icon;
                return (
                  <Link key={`${activity.title}-${activity.time}`} href={transactionHref} prefetch={false} className="grid grid-cols-[44px_1fr_auto] items-center gap-3 py-3 first:pt-1 last:pb-0">
                    <span className={`grid h-11 w-11 place-items-center rounded-full ${activity.tone}`}>
                      <Icon className="h-6 w-6" strokeWidth={2.3} />
                    </span>
                    <span className="min-w-0">
                      <span className="block truncate text-[15px] font-black leading-tight text-[#0a1e38]">{activity.title}</span>
                      <span className="mt-1 block truncate text-[13px] font-semibold text-[#657790]">{activity.detail}</span>
                    </span>
                    <span className="text-right">
                      <span className="block text-[15px] font-black leading-tight text-[#0a1e38]">{activity.amount}</span>
                      <span className="mt-1 block text-[12px] font-semibold text-[#657790]">{activity.time}</span>
                    </span>
                  </Link>
                );
              })}
            </div>
          ) : (
            <div className="grid min-h-[96px] place-items-center rounded-[18px] bg-[#f2fbf8] px-4 text-center">
              <p className="text-sm font-bold leading-relaxed text-[#52756b]">
                {user.isLoggedIn ? "Belum ada aktivitas transaksi." : "Masuk untuk melihat aktivitas transaksi akun."}
              </p>
            </div>
          )}
        </section>

        <section className="mt-5 grid grid-cols-[72px_1fr] items-center gap-3 rounded-[22px] bg-[#dff8ef] px-4 py-4 shadow-[0_16px_34px_rgba(15,78,81,0.09)] min-[410px]:grid-cols-[78px_1fr_auto]">
          <div className="relative grid h-[72px] w-[72px] shrink-0 place-items-center rounded-[20px] bg-[linear-gradient(135deg,#0f9d83,#36d091)] text-white shadow-[0_14px_26px_rgba(7,128,107,0.18)]">
            <Gift className="h-10 w-10" strokeWidth={2.2} />
            <span className="absolute -right-1 top-2 h-3 w-3 rounded-full bg-[#ffd34e]" />
            <span className="absolute bottom-2 left-2 h-2 w-2 rounded-full bg-[#fff6bd]" />
          </div>
          <div className="min-w-0 flex-1">
            <p className="text-[11px] font-black uppercase tracking-[0.08em] text-[#168b75]">Spesial Untuk Kamu</p>
            <h2 className="mt-1 text-[18px] font-black leading-tight text-[#084f55]">Banyak Promo Setiap Hari</h2>
            <p className="mt-1 text-[13px] font-semibold leading-snug text-[#416477]">Dapatkan diskon dan cashback menarik untuk transaksi pilihan.</p>
          </div>
          <Link
            href="/artikel"
            prefetch={false}
            className="col-span-2 flex h-11 shrink-0 items-center justify-center gap-1 rounded-full bg-[#079c7f] px-4 text-[13px] font-black text-white! shadow-[0_12px_24px_rgba(0,141,111,0.18)] visited:text-white! min-[410px]:col-span-1"
          >
            Lihat Promo
            <ChevronRight className="h-4 w-4" strokeWidth={3} />
          </Link>
        </section>
      </div>
    </main>
  );
}
