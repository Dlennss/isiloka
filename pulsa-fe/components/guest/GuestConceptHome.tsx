import Image from "next/image";
import Link from "next/link";

type Hotspot = {
  href: string;
  label: string;
  className: string;
};

const topHotspots: Hotspot[] = [
  { href: "/transaksi", label: "Riwayat transaksi", className: "left-[73%] top-[1%] h-[12%] w-[11%]" },
  { href: "/login", label: "Akun", className: "left-[84%] top-[1%] h-[12%] w-[10%]" },
  { href: "/login", label: "Isi Saldo", className: "left-[4%] top-[43%] h-[13%] w-[30%]" },
  { href: "/user/transfer-bank", label: "Transfer", className: "left-[35%] top-[43%] h-[13%] w-[30%]" },
  { href: "/transaksi", label: "Riwayat", className: "left-[66%] top-[43%] h-[13%] w-[30%]" },
  { href: "/pulsa-data", label: "Isi Sekarang", className: "left-[5%] top-[82%] h-[10%] w-[26%]" },
];

const serviceHotspots: Hotspot[] = [
  { href: "/kategori", label: "Lihat semua layanan", className: "left-[76%] top-[1%] h-[10%] w-[22%]" },
  { href: "/pulsa-data", label: "Pulsa dan Data", className: "left-[1%] top-[16%] h-[38%] w-[32%]" },
  { href: "/listrik/token", label: "Token Listrik", className: "left-[34%] top-[16%] h-[38%] w-[32%]" },
  { href: "/ewallet", label: "E-Wallet", className: "left-[67%] top-[16%] h-[38%] w-[32%]" },
  { href: "/kategori", label: "Tagihan", className: "left-[1%] top-[57%] h-[40%] w-[32%]" },
  { href: "/internet-pascabayar", label: "Paket Internet", className: "left-[34%] top-[57%] h-[40%] w-[32%]" },
  { href: "/kategori", label: "Lainnya", className: "left-[67%] top-[57%] h-[40%] w-[32%]" },
];

const activityHotspots: Hotspot[] = [
  { href: "/transaksi", label: "Aktivitas terakhir", className: "left-[0%] top-[0%] h-[64%] w-full" },
  { href: "/artikel", label: "Promo", className: "left-[0%] top-[70%] h-[28%] w-full" },
];

function HotspotLinks({ items }: { items: Hotspot[] }) {
  return (
    <>
      {items.map((item) => (
        <Link
          key={`${item.href}-${item.label}`}
          href={item.href}
          prefetch={false}
          aria-label={item.label}
          className={`absolute rounded-2xl focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#0aa67f] ${item.className}`}
        />
      ))}
    </>
  );
}

export function GuestConceptHome() {
  return (
    <div className="mx-auto w-full max-w-[430px] px-4 pb-28 pt-4">
      <section className="relative">
        <Image
          src="/isiloka-concept/top_content_stack.png"
          alt="Isiloka, saldo, dan banner utama"
          width={682}
          height={562}
          className="h-auto w-full"
          priority
        />
        <HotspotLinks items={topHotspots} />
      </section>

      <section className="relative mt-4">
        <Image
          src="/isiloka-concept/services_grid_full.png"
          alt="Layanan favorit"
          width={681}
          height={394}
          className="h-auto w-full"
          priority
        />
        <HotspotLinks items={serviceHotspots} />
      </section>

      <section className="relative mt-4">
        <Image
          src="/isiloka-concept/activity_and_promo_block.png"
          alt="Aktivitas terakhir dan promo"
          width={681}
          height={267}
          className="h-auto w-full"
        />
        <HotspotLinks items={activityHotspots} />
      </section>
    </div>
  );
}
