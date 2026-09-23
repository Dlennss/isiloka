"use client";

import Image from "next/image";
import Link from "next/link";
import { useMemo, useState } from "react";
import {
  BookOpen,
  ChevronRight,
  Search,
  Zap,
} from "lucide-react";
import { cn } from "@/lib/utils";

type DirectoryMode = "guest" | "user";

type ServiceItem = {
  label: string;
  href: string;
  iconSrc: string;
};

type ServiceGroup = {
  id: string;
  title: string;
  eyebrow: string;
  items: ServiceItem[];
};

type ServiceDirectoryProps = {
  mode?: DirectoryMode;
  role?: string | null;
};

const guestPath = {
  pulsaData: "/pulsa-data",
  pulsa: "/pulsa",
  paketData: "/paket-data",
  ewallet: "/ewallet",
  transferBank: "/kategori?layanan=transfer-bank",
  game: "/game",
  pln: "/listrik",
  tokenPln: "/listrik/token",
  pdam: "/pdam",
  pgn: "/pgn",
  bpjs: "/bpjs",
  tv: "/tv",
  internet: "/internet-pascabayar",
  hpPascabayar: "/hp-pascabayar",
};

const userPath = {
  pulsaData: "/user/pulsa-data",
  pulsa: "/user/pulsa",
  paketData: "/user/paket-data",
  ewallet: "/user/ewallet",
  transferBank: "/user/transfer-bank",
  game: "/user/kategori?layanan=voucher-game",
  pln: "/user/listrik",
  tokenPln: "/user/listrik/token",
  pdam: "/user/kategori?layanan=pdam",
  pgn: "/user/kategori?layanan=gas-pgn",
  bpjs: "/user/kategori?layanan=bpjs",
  tv: "/user/kategori?layanan=tv-kabel",
  internet: "/user/kategori?layanan=internet-wifi",
  hpPascabayar: "/user/kategori/hp-pascabayar",
  esimRoaming: "/user/kategori/esim-roaming",
};

function fallbackPath(mode: DirectoryMode, slug: string) {
  return mode === "user" ? `/user/kategori?layanan=${slug}` : `/kategori?layanan=${slug}`;
}

function isAgentRole(role?: string | null) {
  const normalized = String(role || "").trim().toLowerCase();
  return normalized === "agent" || normalized === "retail_agent" || normalized === "agent_retail";
}

const iconPath = {
  pulsa: "/service-icons/pulsa.png",
  paketData: "/service-icons/paket-data.png",
  hpPascabayar: "/service-icons/hp-pascabayar.png",
  esimRoaming: "/service-icons/esim-roaming.png",
  ewallet: "/service-icons/ewallet.png",
  transferBank: "/service-icons/transfer-bank.png",
  qris: "/service-icons/qris.png",
  uangElektronik: "/service-icons/uang-elektronik.png",
  kartuKredit: "/service-icons/kartu-kredit.png",
  asuransi: "/service-icons/asuransi.png",
  bpjs: "/service-icons/bpjs.png",
  tokenPln: "/service-icons/token-pln.png",
  pdam: "/service-icons/pdam.png",
  gasPgn: "/service-icons/gas-pgn.png",
  internetWifi: "/service-icons/internet-wifi.png",
  tvKabel: "/service-icons/tv-kabel.png",
  voucherGame: "/service-icons/voucher-game.png",
  voucherDigital: "/service-icons/voucher-digital.png",
  streamingMusik: "/service-icons/streaming-musik.png",
  klinikKesehatan: "/service-icons/klinik-kesehatan.png",
  uangSekolah: "/service-icons/uang-sekolah.png",
  cicilanKendaraan: "/service-icons/cicilan-kendaraan.png",
  cicilanMultifinance: "/service-icons/cicilan-multifinance.png",
  pbb: "/service-icons/pbb.png",
  pajakNegara: "/service-icons/pajak-negara.png",
  tiketPerjalanan: "/service-icons/tiket-perjalanan.png",
  saldoKartuTol: "/service-icons/saldo-kartu-tol.png",
  parkirDigital: "/service-icons/parkir-digital.png",
  kurirPengiriman: "/service-icons/kurir-pengiriman.png",
  zakatDonasi: "/service-icons/zakat-donasi.png",
};

function getGroups(mode: DirectoryMode, role?: string | null): ServiceGroup[] {
  const path = mode === "user" ? userPath : guestPath;
  const normalizedRole = String(role || "").trim().toLowerCase();
  const canUseAgentCredit = mode === "user" && isAgentRole(normalizedRole);

  return [
    {
      id: "komunikasi",
      title: "Komunikasi",
      eyebrow: "Kebutuhan nomor",
      items: [
        { label: "Pulsa", href: path.pulsaData, iconSrc: iconPath.pulsa },
        { label: "Paket Data", href: path.pulsaData, iconSrc: iconPath.paketData },
        { label: "HP Pascabayar", href: path.hpPascabayar, iconSrc: iconPath.hpPascabayar },
        { label: "eSIM & Roaming", href: mode === "user" ? userPath.esimRoaming : fallbackPath(mode, "esim-roaming"), iconSrc: iconPath.esimRoaming },
      ],
    },
    {
      id: "keuangan",
      title: "Keuangan",
      eyebrow: "Saldo & pembayaran",
      items: [
        { label: "E-Wallet", href: path.ewallet, iconSrc: iconPath.ewallet },
        { label: "Transfer Bank", href: path.transferBank, iconSrc: iconPath.transferBank },
        { label: "Pembayaran QRIS", href: fallbackPath(mode, "qris"), iconSrc: iconPath.qris },
        { label: "Uang Elektronik", href: fallbackPath(mode, "uang-elektronik"), iconSrc: iconPath.uangElektronik },
        ...(canUseAgentCredit ? [{ label: "Kredit Saldo Agent", href: "/user/saldo/kredit-agent", iconSrc: iconPath.cicilanMultifinance }] : []),
        { label: "Kartu Kredit", href: fallbackPath(mode, "kartu-kredit"), iconSrc: iconPath.kartuKredit },
        { label: "Asuransi", href: fallbackPath(mode, "asuransi"), iconSrc: iconPath.asuransi },
      ],
    },
    {
      id: "tagihan",
      title: "Tagihan",
      eyebrow: "Bayar rutin",
      items: [
        { label: "BPJS", href: path.bpjs, iconSrc: iconPath.bpjs },
      ],
    },
    {
      id: "rumah",
      title: "Rumah Tangga",
      eyebrow: "Kebutuhan rumah",
      items: [
        { label: "Token PLN", href: path.tokenPln, iconSrc: iconPath.tokenPln },
        { label: "PDAM", href: path.pdam, iconSrc: iconPath.pdam },
        { label: "Gas PGN", href: path.pgn, iconSrc: iconPath.gasPgn },
        { label: "Internet & WiFi", href: path.internet, iconSrc: iconPath.internetWifi },
      ],
    },
    {
      id: "hiburan",
      title: "Hiburan",
      eyebrow: "Digital fun",
      items: [
        { label: "TV Kabel", href: path.tv, iconSrc: iconPath.tvKabel },
        { label: "Voucher Game", href: path.game, iconSrc: iconPath.voucherGame },
        { label: "Voucher Digital", href: fallbackPath(mode, "voucher-digital"), iconSrc: iconPath.voucherDigital },
        { label: "Streaming & Musik", href: fallbackPath(mode, "streaming-musik"), iconSrc: iconPath.streamingMusik },
      ],
    },
    {
      id: "kesehatan",
      title: "Kesehatan",
      eyebrow: "Layanan sehat",
      items: [
        { label: "Klinik & Kesehatan", href: fallbackPath(mode, "klinik-kesehatan"), iconSrc: iconPath.klinikKesehatan },
      ],
    },
    {
      id: "pendidikan",
      title: "Pendidikan",
      eyebrow: "Biaya sekolah",
      items: [
        { label: "Uang Sekolah", href: fallbackPath(mode, "uang-sekolah"), iconSrc: iconPath.uangSekolah },
      ],
    },
    {
      id: "cicilan",
      title: "Cicilan",
      eyebrow: "Bayar angsuran",
      items: [
        { label: "Cicilan Kendaraan", href: fallbackPath(mode, "cicilan-kendaraan"), iconSrc: iconPath.cicilanKendaraan },
        { label: "Cicilan Multifinance", href: fallbackPath(mode, "cicilan-multifinance"), iconSrc: iconPath.cicilanMultifinance },
      ],
    },
    {
      id: "pajak",
      title: "Pajak",
      eyebrow: "Kewajiban resmi",
      items: [
        { label: "PBB", href: fallbackPath(mode, "pbb"), iconSrc: iconPath.pbb },
        { label: "Pajak & Negara", href: fallbackPath(mode, "pajak-negara"), iconSrc: iconPath.pajakNegara },
      ],
    },
    {
      id: "perjalanan",
      title: "Perjalanan",
      eyebrow: "Mobilitas",
      items: [
        { label: "Tiket Perjalanan", href: fallbackPath(mode, "tiket-perjalanan"), iconSrc: iconPath.tiketPerjalanan },
        { label: "Saldo Kartu Tol", href: fallbackPath(mode, "saldo-kartu-tol"), iconSrc: iconPath.saldoKartuTol },
        { label: "Parkir Digital", href: fallbackPath(mode, "parkir-digital"), iconSrc: iconPath.parkirDigital },
        { label: "Kurir & Pengiriman", href: fallbackPath(mode, "kurir-pengiriman"), iconSrc: iconPath.kurirPengiriman },
      ],
    },
    {
      id: "sosial",
      title: "Sosial",
      eyebrow: "Berbagi",
      items: [
        { label: "Zakat & Donasi", href: fallbackPath(mode, "zakat-donasi"), iconSrc: iconPath.zakatDonasi },
      ],
    },
  ];
}

export function ServiceDirectory({ mode = "guest", role }: ServiceDirectoryProps) {
  const groups = useMemo(() => getGroups(mode, role), [mode, role]);
  const [activeGroup, setActiveGroup] = useState("semua");
  const [query, setQuery] = useState("");

  const normalizedQuery = query.trim().toLowerCase();
  const filteredGroups = groups
    .filter((group) => activeGroup === "semua" || group.id === activeGroup)
    .map((group) => ({
      ...group,
      items: group.items.filter((item) => item.label.toLowerCase().includes(normalizedQuery)),
    }))
    .filter((group) => group.items.length > 0);

  return (
    <section className="space-y-3 pb-24">
      <div className="sticky top-0 z-20 -mx-4 bg-[#f0f8f8]/92 px-4 pb-3 pt-3 backdrop-blur-xl">
        <label className="flex h-13 items-center gap-3 rounded-[22px] border border-emerald-950/10 bg-white px-4 shadow-[0_12px_30px_rgba(6,78,59,0.08)]">
          <Search className="h-5 w-5 shrink-0 text-[#087e8b]" />
          <input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Cari pulsa, tagihan, voucher..."
            className="h-full min-w-0 flex-1 bg-transparent text-sm font-semibold text-slate-800 outline-none placeholder:text-slate-400"
          />
        </label>

        <div className="-mx-4 mt-3 overflow-x-auto px-4 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
          <div className="flex w-max gap-2">
            <button
              type="button"
              onClick={() => setActiveGroup("semua")}
              className={cn(
                "h-9 rounded-full px-4 text-xs font-black transition",
                activeGroup === "semua"
                  ? "bg-[#13464b] text-white shadow-[0_10px_20px_rgba(5,46,38,0.22)]"
                  : "border border-emerald-950/10 bg-white text-slate-600"
              )}
            >
              Semua
            </button>
            {groups.map((group) => (
              <button
                key={group.id}
                type="button"
                onClick={() => setActiveGroup(group.id)}
                className={cn(
                  "h-9 rounded-full px-4 text-xs font-black transition",
                  activeGroup === group.id
                    ? "bg-[#13464b] text-white shadow-[0_10px_20px_rgba(5,46,38,0.22)]"
                    : "border border-emerald-950/10 bg-white text-slate-600"
                )}
              >
                {group.title}
              </button>
            ))}
          </div>
        </div>
      </div>

      <div className="overflow-hidden rounded-[24px] border border-lime-200/70 bg-linear-to-r from-[#fff8e7] via-white to-[#f0f8f8] px-4 py-3 shadow-[0_12px_28px_rgba(6,78,59,0.08)]">
        <div className="flex items-center gap-3">
          <div className="grid h-10 w-10 shrink-0 place-items-center rounded-2xl bg-orange-100 text-orange-500">
            <Zap className="h-5 w-5" />
          </div>
          <div>
            <p className="text-sm font-black text-[#13464b]">Transaksi makin praktis</p>
            <p className="mt-0.5 text-[11px] font-semibold text-slate-500">Pilih layanan, isi data, lalu selesaikan pembayaran.</p>
          </div>
        </div>
      </div>

      {filteredGroups.length > 0 ? (
        filteredGroups.map((group) => (
          <div
            key={group.id}
            className="overflow-hidden rounded-[26px] border border-emerald-950/[0.08] bg-white p-4 shadow-[0_14px_34px_rgba(6,78,59,0.08)]"
          >
            <div className="mb-4 flex items-center gap-2">
              <h2 className="text-[17px] font-black tracking-tight text-slate-950">{group.title}</h2>
              <span className="rounded-full bg-lime-100 px-2 py-1 text-[10px] font-black text-[#087e8b]">
                {group.items.length} layanan
              </span>
            </div>

            <div className="grid grid-cols-4 gap-x-2 gap-y-5">
              {group.items.map((item) => {
                return (
                  <Link
                    key={`${group.id}-${item.label}`}
                    href={item.href}
                    prefetch={false}
                    className="group flex min-h-[82px] flex-col items-center justify-start gap-2 text-center"
                  >
                    <span className="relative grid h-14 w-14 place-items-center overflow-hidden rounded-[20px] bg-white shadow-[0_12px_24px_rgba(6,78,59,0.12)] ring-1 ring-slate-200/80 transition duration-300 group-hover:-translate-y-0.5 group-hover:shadow-[0_16px_30px_rgba(6,78,59,0.18)]">
                      <Image
                        src={item.iconSrc}
                        alt=""
                        fill
                        sizes="56px"
                        className="object-contain p-0.5"
                      />
                    </span>
                    <span className="line-clamp-2 max-w-[76px] text-[10px] font-black leading-tight text-slate-950">
                      {item.label}
                    </span>
                  </Link>
                );
              })}
            </div>
          </div>
        ))
      ) : (
        <div className="rounded-[26px] border border-dashed border-emerald-200 bg-white p-8 text-center shadow-[0_14px_34px_rgba(6,78,59,0.08)]">
          <BookOpen className="mx-auto h-8 w-8 text-[#087e8b]" />
          <p className="mt-3 text-sm font-black text-slate-900">Layanan tidak ditemukan</p>
          <p className="mt-1 text-xs font-semibold text-slate-500">Coba kata kunci yang lebih singkat.</p>
        </div>
      )}

      <Link
        href={mode === "user" ? "/user" : "/"}
        prefetch={false}
        className="group flex items-center justify-between rounded-[24px] border border-emerald-300/30 bg-[#13464b] px-4 py-4 text-white shadow-[0_16px_34px_rgba(5,46,38,0.22)]"
      >
        <span>
          <span className="block text-sm font-black !text-white">Kembali ke beranda</span>
          <span className="mt-0.5 block text-xs font-semibold !text-emerald-50">Lihat promo dan produk utama.</span>
        </span>
        <span className="grid h-10 w-10 place-items-center rounded-2xl border border-white/70 bg-lime-300 text-[#13464b] shadow-[0_8px_18px_rgba(163,230,53,0.28)] transition group-hover:translate-x-0.5">
          <ChevronRight className="h-5 w-5" strokeWidth={2.6} />
        </span>
      </Link>
    </section>
  );
}
