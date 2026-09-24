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

type GuestConceptHomeProps = {
  isLoggedIn?: boolean;
};

export function GuestConceptHome({ isLoggedIn = false }: GuestConceptHomeProps) {
  return (
    <div className="mx-auto w-full max-w-md px-5 pb-24 pt-1">
      <section className="mb-5">
        <h1 className="text-[27px] font-black leading-none text-[#071d38]">
          {isLoggedIn ? "Halo!" : "Halo, Dinda!"}
        </h1>
        <p className="mt-2 text-[19px] font-medium leading-tight text-[#65748f]">
          Semoga harimu menyenangkan
          <span className="ml-1 align-middle text-xl">☀</span>
        </p>
      </section>

      <section className="space-y-4">
        <Link href={isLoggedIn ? "/user/saldo" : "/login"} prefetch={false} className="block transition hover:-translate-y-0.5">
          <Image
            src="/isiloka-concept/saldo_card_full.png"
            alt="Saldo Isiloka Rp 250.000"
            width={683}
            height={238}
            className="h-auto w-full rounded-[24px] shadow-[0_22px_46px_rgba(5,120,104,0.18)]"
            priority
          />
        </Link>

        <Link href="/pulsa-data" prefetch={false} className="block transition hover:-translate-y-0.5">
          <Image
            src="/isiloka-concept/hero_banner_full.png"
            alt="Semua kebutuhan dalam satu aplikasi"
            width={681}
            height={269}
            className="h-auto w-full rounded-[22px] shadow-[0_18px_38px_rgba(26,129,113,0.10)]"
            priority
          />
        </Link>
        <Image src="/isiloka-concept/hero_banner_dots.png" alt="" width={82} height={23} className="mx-auto -mt-1 h-3 w-auto" />
      </section>

      <section className="mt-4">
        <div className="mb-3 flex items-center justify-between gap-3">
          <h2 className="text-[21px] font-black leading-none text-[#071d38]">Layanan Favorit</h2>
          <Link href="/kategori" prefetch={false} className="text-[15px] font-black text-[#0aa67f]">
            Lihat Semua ›
          </Link>
        </div>
        <div className="grid grid-cols-3 gap-3">
          {services.map((service) => (
            <Link key={service.alt} href={service.href} prefetch={false} className="block transition hover:-translate-y-0.5">
              <Image
                src={service.src}
                alt={service.alt}
                width={217}
                height={160}
                className="h-auto w-full rounded-[18px] shadow-[0_14px_28px_rgba(27,78,94,0.08)]"
              />
            </Link>
          ))}
        </div>
      </section>

      <section className="mt-5">
        <Link href={isLoggedIn ? "/user/transaksi" : "/transaksi"} prefetch={false} className="block transition hover:-translate-y-0.5">
          <Image
            src="/isiloka-concept/section_aktivitas_terakhir.png"
            alt="Aktivitas terakhir"
            width={681}
            height={280}
            className="h-auto w-full rounded-[22px] shadow-[0_18px_38px_rgba(27,78,94,0.08)]"
          />
        </Link>
      </section>

      <section className="mt-5">
        <Link href="/artikel" prefetch={false} className="block transition hover:-translate-y-0.5">
          <Image
            src="/isiloka-concept/promo_banner_full.png"
            alt="Banyak promo setiap hari"
            width={680}
            height={111}
            className="h-auto w-full rounded-[22px] shadow-[0_18px_38px_rgba(27,78,94,0.08)]"
          />
        </Link>
      </section>
    </div>
  );
}
