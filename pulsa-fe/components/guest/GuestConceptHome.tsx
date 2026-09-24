import Image from "next/image";
import Link from "next/link";

type Hotspot = {
  href: string;
  label: string;
  className: string;
};

const hotspots: Hotspot[] = [
  { href: "/login", label: "Buka akun", className: "left-[73%] top-[5%] h-[6%] w-[14%]" },
  { href: "/login", label: "Isi saldo", className: "left-[17%] top-[23%] h-[5%] w-[22%]" },
  { href: "/user/transfer-bank", label: "Transfer", className: "left-[40%] top-[23%] h-[5%] w-[22%]" },
  { href: "/transaksi", label: "Riwayat", className: "left-[61%] top-[23%] h-[5%] w-[22%]" },
  { href: "/pulsa-data", label: "Isi sekarang", className: "left-[16%] top-[40%] h-[4%] w-[22%]" },
  { href: "/pulsa-data", label: "Pulsa dan data", className: "left-[16%] top-[49%] h-[8%] w-[22%]" },
  { href: "/listrik/token", label: "Token listrik", className: "left-[39%] top-[49%] h-[8%] w-[23%]" },
  { href: "/ewallet", label: "E-Wallet", className: "left-[62%] top-[49%] h-[8%] w-[22%]" },
  { href: "/kategori", label: "Tagihan", className: "left-[16%] top-[57%] h-[8%] w-[22%]" },
  { href: "/internet-pascabayar", label: "Paket internet", className: "left-[39%] top-[57%] h-[8%] w-[23%]" },
  { href: "/kategori", label: "Lainnya", className: "left-[62%] top-[57%] h-[8%] w-[22%]" },
  { href: "/transaksi", label: "Aktivitas terakhir", className: "left-[15%] top-[67%] h-[16%] w-[70%]" },
  { href: "/artikel", label: "Promo", className: "left-[15%] top-[85%] h-[7%] w-[70%]" },
  { href: "/", label: "Beranda", className: "left-[18%] top-[94%] h-[5%] w-[12%]" },
  { href: "/transaksi", label: "Riwayat", className: "left-[37%] top-[94%] h-[5%] w-[12%]" },
  { href: "/artikel", label: "Promo", className: "left-[56%] top-[94%] h-[5%] w-[12%]" },
  { href: "/login", label: "Akun", className: "left-[73%] top-[94%] h-[5%] w-[12%]" },
];

export function GuestConceptHome() {
  return (
    <div className="mx-auto w-full max-w-[941px]">
      <div className="relative">
        <Image
          src="/isiloka-concept/dashboard_full.png"
          alt="Isiloka"
          width={941}
          height={1672}
          className="h-auto w-full select-none"
          priority
        />
        {hotspots.map((hotspot) => (
          <Link
            key={`${hotspot.href}-${hotspot.label}`}
            href={hotspot.href}
            prefetch={false}
            aria-label={hotspot.label}
            className={`absolute rounded-2xl focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#0aa67f] ${hotspot.className}`}
          />
        ))}
      </div>
    </div>
  );
}
