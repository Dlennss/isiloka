import Image from "next/image";
import Link from "next/link";
import {
  ChevronRight,
  Eye,
  Plus,
  ReceiptText,
  Send,
  ShieldCheck,
  Smartphone,
  Sun,
  Wallet,
  Zap,
} from "lucide-react";

const services = [
  { href: "/pulsa-data", label: "Pulsa & Data", icon: "/isiloka-concept/pulsa_data_icon.png" },
  { href: "/listrik/token", label: "Token Listrik", icon: "/isiloka-concept/token_listrik_icon.png" },
  { href: "/ewallet", label: "E-Wallet", icon: "/isiloka-concept/ewallet_icon.png" },
  { href: "/kategori", label: "Tagihan", icon: "/isiloka-concept/tagihan_icon.png" },
  { href: "/internet-pascabayar", label: "Paket Internet", icon: "/isiloka-concept/paket_internet_icon.png" },
  { href: "/kategori", label: "Lainnya", icon: "/isiloka-concept/lainnya_icon.png" },
];

const activities = [
  {
    title: "Isi Pulsa Telkomsel",
    detail: "+62 812 3456 7890",
    amount: "- Rp 50.000",
    time: "Hari ini, 08:24",
    icon: Smartphone,
    tone: "bg-[#dcfff4] text-[#12b98a]",
  },
  {
    title: "Token Listrik PLN",
    detail: "No. Meter 1234 5678 90",
    amount: "- Rp 100.000",
    time: "Kemarin, 19:12",
    icon: Zap,
    tone: "bg-[#fff1cc] text-[#ffac18]",
  },
  {
    title: "Top Up DANA",
    detail: "+62 812 3456 7890",
    amount: "- Rp 75.000",
    time: "12 Mar 2024, 14:03",
    icon: Wallet,
    tone: "bg-[#eee3ff] text-[#7654e8]",
  },
];

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

export function GuestConceptHome() {
  return (
    <main className="mx-auto min-h-dvh w-full max-w-[430px] overflow-hidden bg-[radial-gradient(circle_at_50%_0%,#fafffe_0%,#f1fffb_44%,#e7f8f4_100%)] px-4 pb-28 pt-4 text-[#071d38] shadow-[0_20px_70px_rgba(8,91,84,0.14)] md:rounded-[34px]">
      <div className="mx-auto w-full max-w-[398px]">
        <header className="flex items-start justify-between pt-1">
          <div className="flex items-center gap-3">
            <Image
              src="/isiloka-concept/logo_symbol.png"
              alt="Isiloka"
              width={48}
              height={48}
              className="h-12 w-12 rounded-xl"
              priority
            />
            <div className="leading-none">
              <div className="text-[26px] font-black tracking-normal text-[#084f55]">Isiloka</div>
              <div className="mt-1 text-[11px] font-extrabold tracking-[0.03em] text-[#536a7d]">ISI HARI, DARI SINI</div>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <Link
              href="/transaksi"
              prefetch={false}
              aria-label="Notifikasi"
              className="relative grid h-12 w-12 place-items-center rounded-[18px] bg-white shadow-[0_14px_32px_rgba(12,68,75,0.11)]"
            >
              <Image src="/isiloka-concept/bell_notification.png" alt="" width={31} height={31} className="h-8 w-8" />
            </Link>
            <Link href="/login" prefetch={false} aria-label="Akun">
              <Image
                src="/isiloka-concept/avatar_profile.png"
                alt=""
                width={48}
                height={48}
                className="h-12 w-12 rounded-full object-cover"
              />
            </Link>
          </div>
        </header>

        <section className="mt-5">
          <h1 className="text-[28px] font-black leading-none text-[#061d38]">Halo, Dinda!</h1>
          <p className="mt-2 flex items-center gap-2 text-[17px] font-semibold leading-tight text-[#62728b]">
            Semoga harimu menyenangkan
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
              <div className="text-[38px] font-black leading-none tracking-normal">Rp 250.000</div>
              <button
                type="button"
                aria-label="Lihat saldo"
                className="grid h-9 w-9 place-items-center rounded-full bg-white/16 text-white backdrop-blur"
              >
                <Eye className="h-5 w-5" strokeWidth={2.4} />
              </button>
            </div>

            <div className="mt-7 grid grid-cols-3 gap-2">
              <Link href="/login" prefetch={false} className="flex h-[62px] items-center justify-center gap-2 rounded-xl bg-white px-2 text-[13px] font-black text-[#075862]! shadow-[0_10px_24px_rgba(6,77,70,0.12)] visited:text-[#075862]!">
                <span className="grid h-8 w-8 shrink-0 place-items-center rounded-full bg-[#0a7d76] text-white">
                  <Plus className="h-5 w-5" strokeWidth={3} />
                </span>
                <span>Isi Saldo</span>
              </Link>
              <Link href="/user/transfer-bank" prefetch={false} className="flex h-[62px] items-center justify-center gap-2 rounded-xl bg-white px-2 text-[13px] font-black text-[#075862]! shadow-[0_10px_24px_rgba(6,77,70,0.12)] visited:text-[#075862]!">
                <Send className="h-7 w-7 shrink-0 fill-[#0a7d76] text-[#0a7d76]" strokeWidth={1.8} />
                <span>Transfer</span>
              </Link>
              <Link href="/transaksi" prefetch={false} className="flex h-[62px] items-center justify-center gap-2 rounded-xl bg-white px-2 text-[13px] font-black text-[#075862]! shadow-[0_10px_24px_rgba(6,77,70,0.12)] visited:text-[#075862]!">
                <ReceiptText className="h-7 w-7 shrink-0 fill-[#0a7d76] text-white" strokeWidth={2.2} />
                <span>Riwayat</span>
              </Link>
            </div>
          </div>
        </section>

        <section className="relative mt-4 overflow-hidden rounded-[23px] bg-[#def8f2] px-5 pb-4 pt-5 shadow-[0_16px_36px_rgba(22,102,95,0.10)]">
          <div className="absolute -bottom-12 left-32 h-32 w-52 rounded-full bg-[#b7efd9]" />
          <div className="absolute -right-5 bottom-0 h-32 w-32 rounded-full bg-white/40" />
          <div className="relative z-10 grid min-h-[205px] grid-cols-[1.1fr_0.9fr] gap-2 min-[390px]:grid-cols-[1.05fr_0.95fr]">
            <div className="flex flex-col items-start">
              <h2 className="text-[20px] font-black leading-[1.18] text-[#084f55] min-[390px]:text-[25px]">Semua Kebutuhan Dalam Satu Aplikasi</h2>
              <p className="mt-2 text-[14px] font-semibold leading-snug text-[#5a6f81]">
                Isi pulsa, paket data, token listrik dan berbagai pembayaran lainnya.
              </p>
              <Link
                href="/pulsa-data"
                prefetch={false}
                className="mt-4 flex h-11 items-center gap-2 rounded-full bg-[#079c7f] px-5 text-[15px] font-extrabold text-white shadow-[0_12px_24px_rgba(0,141,111,0.22)]"
              >
                Isi Sekarang
                <ChevronRight className="h-5 w-5" strokeWidth={3} />
              </Link>
            </div>
            <div className="relative min-h-[190px]">
              <Image
                src="/isiloka-concept/hero_banner_phone_illustration.png"
                alt=""
                width={250}
                height={222}
                className="absolute bottom-[-2px] right-[-8px] h-auto w-[158px] max-w-none min-[390px]:right-[-4px] min-[390px]:w-[200px]"
                sizes="(min-width: 390px) 200px, 158px"
              />
              <div className="absolute right-0 top-3 hidden max-w-[78px] rotate-[-5deg] text-center text-[18px] font-black italic leading-[1.02] text-[#07515a] min-[390px]:block">
                Lebih Mudah Lebih Dekat Untukmu
              </div>
            </div>
          </div>
        </section>

        <div className="mt-2 flex justify-center gap-1.5">
          <span className="h-2.5 w-4 rounded-full bg-[#0aa889]" />
          <span className="h-2.5 w-2.5 rounded-full bg-[#b8ddd8]" />
          <span className="h-2.5 w-2.5 rounded-full bg-[#b8ddd8]" />
        </div>

        <section className="mt-5">
          <SectionTitle title="Layanan Favorit" href="/kategori" />
          <div className="grid grid-cols-3 gap-3">
            {services.map((service) => (
              <Link
                key={service.label}
                href={service.href}
                prefetch={false}
                className="flex aspect-[1.42] min-h-[102px] flex-col items-center justify-center rounded-[18px] bg-white px-2 text-center shadow-[0_13px_28px_rgba(15,78,81,0.09)]"
              >
                <Image src={service.icon} alt="" width={54} height={54} className="h-[54px] w-[54px] object-contain" />
                <span className="mt-2 text-[13px] font-black leading-tight text-[#0a1e38]">{service.label}</span>
              </Link>
            ))}
          </div>
        </section>

        <section className="mt-5 rounded-[22px] bg-white px-4 py-4 shadow-[0_16px_36px_rgba(15,78,81,0.10)]">
          <SectionTitle title="Aktivitas Terakhir" href="/transaksi" />
          <div className="divide-y divide-[#e5eef0]">
            {activities.map((activity) => {
              const Icon = activity.icon;
              return (
                <Link key={activity.title} href="/transaksi" prefetch={false} className="grid grid-cols-[44px_1fr_auto] items-center gap-3 py-3 first:pt-1 last:pb-0">
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
        </section>

        <section className="mt-5 flex items-center gap-3 rounded-[22px] bg-[#dff8ef] px-4 py-4 shadow-[0_16px_34px_rgba(15,78,81,0.09)]">
          <Image
            src="/isiloka-concept/promo_gift_illustration.png"
            alt=""
            width={95}
            height={82}
            className="h-[78px] w-[90px] shrink-0 object-contain"
          />
          <div className="min-w-0 flex-1">
            <p className="text-[11px] font-black uppercase tracking-[0.08em] text-[#168b75]">Spesial Untuk Kamu</p>
            <h2 className="mt-1 text-[18px] font-black leading-tight text-[#084f55]">Banyak Promo Setiap Hari</h2>
            <p className="mt-1 text-[13px] font-semibold leading-snug text-[#416477]">Dapatkan diskon dan cashback menarik untuk transaksi pilihan.</p>
          </div>
          <Link
            href="/artikel"
            prefetch={false}
            className="hidden h-11 shrink-0 items-center gap-1 rounded-full bg-[#079c7f] px-4 text-[13px] font-black text-white shadow-[0_12px_24px_rgba(0,141,111,0.18)] min-[390px]:flex"
          >
            Lihat Promo
            <ChevronRight className="h-4 w-4" strokeWidth={3} />
          </Link>
        </section>
      </div>
    </main>
  );
}
