import Image from "next/image";
import Link from "next/link";

const services = [
  { href: "/pulsa-data", src: "/isiloka-concept/pulsa_data_tile.png", alt: "Pulsa & Data" },
  { href: "/listrik/token", src: "/isiloka-concept/token_listrik_tile.png", alt: "Token Listrik" },
  { href: "/ewallet", src: "/isiloka-concept/ewallet_tile.png", alt: "E-Wallet" },
  { href: "/kategori", src: "/isiloka-concept/tagihan_tile.png", alt: "Tagihan" },
  { href: "/internet-pascabayar", src: "/isiloka-concept/paket_internet_tile.png", alt: "Paket Internet" },
  { href: "/kategori", src: "/isiloka-concept/lainnya_tile.png", alt: "Lainnya" },
];

export function GuestConceptHome() {
  return (
    <div className="mx-auto w-full max-w-[430px] px-4 pb-28 pt-5">
      <header className="flex items-center justify-between gap-3">
        <Link href="/" prefetch={false} className="flex min-w-0 items-center gap-3">
          <Image src="/isiloka-concept/logo_symbol.png" alt="" width={71} height={73} className="h-10 w-auto shrink-0 min-[390px]:h-12" priority />
          <Image src="/isiloka-concept/logo_wordmark_tagline.png" alt="Isiloka - Isi hari, dari sini" width={196} height={73} className="h-10 w-auto min-w-0 object-contain min-[390px]:h-12" priority />
        </Link>
        <div className="flex shrink-0 items-center gap-2">
          <Link href="/transaksi" prefetch={false} aria-label="Riwayat transaksi" className="grid h-11 w-11 place-items-center rounded-[18px] bg-white shadow-[0_12px_28px_rgba(12,74,76,0.08)]">
            <Image src="/isiloka-concept/bell_notification.png" alt="" width={59} height={72} className="h-9 w-auto" />
          </Link>
          <Link href="/login" prefetch={false} aria-label="Akun" className="grid h-11 w-11 place-items-center overflow-hidden rounded-full bg-white shadow-[0_12px_28px_rgba(12,74,76,0.08)]">
            <Image src="/isiloka-concept/avatar_profile.png" alt="" width={58} height={67} className="h-full w-full object-cover" />
          </Link>
        </div>
      </header>

      <section className="mt-5">
        <h1 className="text-[25px] font-black leading-none text-[#071d38] min-[390px]:text-[28px]">Halo, Dinda!</h1>
        <p className="mt-2 text-[17px] font-medium leading-tight text-[#65748f] min-[390px]:text-[19px]">
          Semoga harimu menyenangkan
          <span className="ml-1 align-middle text-xl">☀</span>
        </p>
      </section>

      <section className="mt-5 space-y-4">
        <Link href="/login" prefetch={false} className="block">
          <Image
            src="/isiloka-concept/saldo_card_full.png"
            alt="Saldo Isiloka Rp 250.000"
            width={683}
            height={238}
            className="h-auto w-full rounded-[22px] shadow-[0_18px_38px_rgba(5,120,104,0.18)]"
            priority
          />
        </Link>

        <Link href="/pulsa-data" prefetch={false} className="block">
          <Image
            src="/isiloka-concept/hero_banner_full.png"
            alt="Semua kebutuhan dalam satu aplikasi"
            width={681}
            height={269}
            className="h-auto w-full rounded-[20px] shadow-[0_16px_32px_rgba(26,129,113,0.10)]"
            priority
          />
        </Link>
        <Image src="/isiloka-concept/hero_banner_dots.png" alt="" width={82} height={23} className="mx-auto -mt-2 h-3 w-auto" />
      </section>

      <section className="mt-4">
        <div className="mb-3 flex items-center justify-between gap-3">
          <h2 className="text-[19px] font-black leading-none text-[#071d38] min-[390px]:text-[21px]">Layanan Favorit</h2>
          <Link href="/kategori" prefetch={false} className="text-[13px] font-black text-[#0aa67f] min-[390px]:text-[15px]">
            Lihat Semua ›
          </Link>
        </div>
        <div className="grid grid-cols-3 gap-2.5 min-[390px]:gap-3">
          {services.map((service) => (
            <Link key={service.alt} href={service.href} prefetch={false} className="block">
              <Image
                src={service.src}
                alt={service.alt}
                width={217}
                height={160}
                className="h-auto w-full rounded-[18px] shadow-[0_12px_24px_rgba(27,78,94,0.07)]"
              />
            </Link>
          ))}
        </div>
      </section>

      <section className="mt-5">
        <Link href="/transaksi" prefetch={false} className="block">
          <Image
            src="/isiloka-concept/section_aktivitas_terakhir.png"
            alt="Aktivitas terakhir"
            width={681}
            height={280}
            className="h-auto w-full rounded-[20px] shadow-[0_16px_32px_rgba(27,78,94,0.08)]"
          />
        </Link>
      </section>

      <section className="mt-5">
        <Link href="/artikel" prefetch={false} className="block">
          <Image
            src="/isiloka-concept/promo_banner_full.png"
            alt="Banyak promo setiap hari"
            width={680}
            height={111}
            className="h-auto w-full rounded-[20px] shadow-[0_16px_32px_rgba(27,78,94,0.08)]"
          />
        </Link>
      </section>
    </div>
  );
}
