"use client";

import Image from "next/image";
import Link from "next/link";
import { useMemo, useState } from "react";
import {
  ArrowLeft,
  BadgePercent,
  BriefcaseBusiness,
  Car,
  CalendarDays,
  Check,
  CheckCircle2,
  ChevronRight,
  Clock3,
  CircleDollarSign,
  Copy,
  CreditCard,
  Gamepad2,
  GraduationCap,
  HandHeart,
  HeartPulse,
  Headphones,
  Home,
  Landmark,
  MonitorPlay,
  Plane,
  QrCode,
  ReceiptText,
  Rocket,
  Search,
  ShieldCheck,
  Signal,
  Smartphone,
  Sparkles,
  Tv,
  UserRound,
  WalletCards,
  Waves,
  Wifi,
  type LucideIcon,
} from "lucide-react";
import type { UserAppOrder } from "@/components/user/types";

type ServiceMeta = {
  title: string;
  subtitle: string;
  inputLabel: string;
  placeholder: string;
  icon: LucideIcon;
  providers: string[];
  products: string[];
  accent: string;
};

type DraftOrder = {
  invoiceId: string;
  service: string;
  destination: string;
  provider: string;
  product: string;
  total: number;
};

type HpPostpaidBill = {
  customerName: string;
  destination: string;
  provider: string;
  period: string;
  nominal: number;
  adminFee: number;
  total: number;
};

type UniversalOrderResponse = {
  ok?: boolean;
  error?: string;
  item?: UserAppOrder;
};

const serviceMap: Record<string, ServiceMeta> = {
  "hp-pascabayar": {
    title: "HP Pascabayar",
    subtitle: "Cek dan bayar tagihan nomor pascabayar.",
    inputLabel: "Nomor HP",
    placeholder: "Masukkan nomor pascabayar",
    icon: ReceiptText,
    providers: ["Telkomsel Halo", "Indosat Postpaid", "XL Prioritas"],
    products: ["Cek Tagihan", "Bayar Tagihan", "Reminder Bulanan"],
    accent: "from-cyan-500 to-sky-500",
  },
  "esim-roaming": {
    title: "eSIM & Roaming",
    subtitle: "Paket internet perjalanan luar negeri.",
    inputLabel: "Nomor HP",
    placeholder: "Masukkan nomor aktif",
    icon: Smartphone,
    providers: ["Asia Roaming", "Global eSIM", "Umrah & Haji"],
    products: ["1 GB / 3 Hari", "3 GB / 7 Hari", "5 GB / 15 Hari"],
    accent: "from-violet-500 to-indigo-500",
  },
  qris: {
    title: "Pembayaran QRIS",
    subtitle: "Bayar merchant dengan kode QR.",
    inputLabel: "ID / Nama Merchant",
    placeholder: "Masukkan ID merchant",
    icon: QrCode,
    providers: ["QRIS Dinamis", "QRIS Statis", "Merchant Lokal"],
    products: ["Rp 25.000", "Rp 50.000", "Rp 100.000"],
    accent: "from-purple-500 to-fuchsia-500",
  },
  "uang-elektronik": {
    title: "Uang Elektronik",
    subtitle: "Isi saldo kartu transport dan uang elektronik.",
    inputLabel: "Nomor Kartu",
    placeholder: "Masukkan nomor kartu",
    icon: CreditCard,
    providers: ["E-Money", "TapCash", "Brizzi"],
    products: ["Rp 20.000", "Rp 50.000", "Rp 100.000"],
    accent: "from-indigo-500 to-blue-500",
  },
  "kartu-kredit": {
    title: "Kartu Kredit",
    subtitle: "Cek dan bayar tagihan kartu kredit.",
    inputLabel: "Nomor Kartu",
    placeholder: "Masukkan 12-16 digit kartu",
    icon: CreditCard,
    providers: ["BCA Card", "Mandiri Card", "BRI Card"],
    products: ["Cek Tagihan", "Bayar Minimum", "Bayar Penuh"],
    accent: "from-violet-500 to-indigo-500",
  },
  asuransi: {
    title: "Asuransi",
    subtitle: "Bayar premi asuransi dengan cepat.",
    inputLabel: "Nomor Polis",
    placeholder: "Masukkan nomor polis",
    icon: ShieldCheck,
    providers: ["Prudential", "Allianz", "AIA"],
    products: ["Premi Bulanan", "Premi Tahunan", "Cek Polis"],
    accent: "from-emerald-500 to-teal-500",
  },
  bpjs: {
    title: "BPJS",
    subtitle: "Bayar iuran BPJS Kesehatan.",
    inputLabel: "Nomor VA BPJS",
    placeholder: "Masukkan nomor BPJS",
    icon: HeartPulse,
    providers: ["BPJS Kesehatan", "BPJS Ketenagakerjaan"],
    products: ["1 Bulan", "3 Bulan", "6 Bulan"],
    accent: "from-rose-500 to-orange-500",
  },
  pdam: {
    title: "PDAM",
    subtitle: "Cek tagihan air sesuai wilayah.",
    inputLabel: "ID Pelanggan",
    placeholder: "Masukkan ID pelanggan",
    icon: Waves,
    providers: ["PDAM Kota", "PDAM Kabupaten", "Perumda Air"],
    products: ["Cek Tagihan", "Bayar Tagihan", "Simpan ID"],
    accent: "from-cyan-500 to-teal-500",
  },
  "gas-pgn": {
    title: "Gas PGN",
    subtitle: "Bayar tagihan gas rumah tangga.",
    inputLabel: "ID Pelanggan",
    placeholder: "Masukkan ID pelanggan",
    icon: Home,
    providers: ["PGN Rumah", "PGN Bisnis", "Gas Pintar"],
    products: ["Cek Tagihan", "Bayar Tagihan", "Riwayat Meter"],
    accent: "from-orange-500 to-amber-500",
  },
  "internet-wifi": {
    title: "Internet & WiFi",
    subtitle: "Bayar internet rumah dan WiFi.",
    inputLabel: "ID Pelanggan",
    placeholder: "Masukkan nomor pelanggan",
    icon: Wifi,
    providers: ["IndiHome", "Iconnet", "Biznet"],
    products: ["Cek Tagihan", "Bayar Bulanan", "Perpanjang Paket"],
    accent: "from-sky-500 to-blue-500",
  },
  "tv-kabel": {
    title: "TV Kabel",
    subtitle: "Bayar langganan TV kabel.",
    inputLabel: "ID Pelanggan",
    placeholder: "Masukkan ID pelanggan",
    icon: Tv,
    providers: ["K-Vision", "Transvision", "MNC Vision"],
    products: ["Paket Dasar", "Paket Sport", "Paket Family"],
    accent: "from-purple-500 to-violet-500",
  },
  "voucher-game": {
    title: "Voucher Game",
    subtitle: "Top up game dan voucher digital.",
    inputLabel: "User ID / Email",
    placeholder: "Masukkan ID game atau email",
    icon: Gamepad2,
    providers: ["Mobile Legends", "Free Fire", "PUBG Mobile"],
    products: ["86 Diamonds", "172 Diamonds", "Weekly Pass"],
    accent: "from-emerald-600 to-lime-500",
  },
  "voucher-digital": {
    title: "Voucher Digital",
    subtitle: "Beli voucher digital favorit.",
    inputLabel: "Email / Nomor HP",
    placeholder: "Masukkan tujuan voucher",
    icon: BadgePercent,
    providers: ["Google Play", "Apple Gift", "Steam Wallet"],
    products: ["Rp 25.000", "Rp 50.000", "Rp 100.000"],
    accent: "from-red-500 to-orange-500",
  },
  "streaming-musik": {
    title: "Streaming & Musik",
    subtitle: "Langganan hiburan digital.",
    inputLabel: "Email / Nomor HP",
    placeholder: "Masukkan akun tujuan",
    icon: MonitorPlay,
    providers: ["Spotify", "Vidio", "Netflix"],
    products: ["7 Hari", "30 Hari", "90 Hari"],
    accent: "from-fuchsia-500 to-pink-500",
  },
  "klinik-kesehatan": {
    title: "Klinik & Kesehatan",
    subtitle: "Pembayaran layanan kesehatan.",
    inputLabel: "Nomor Pasien",
    placeholder: "Masukkan nomor pasien",
    icon: HeartPulse,
    providers: ["Klinik Umum", "Apotek", "Telemedis"],
    products: ["Registrasi", "Tebus Obat", "Konsultasi"],
    accent: "from-rose-500 to-pink-500",
  },
  "uang-sekolah": {
    title: "Uang Sekolah",
    subtitle: "Bayar pendidikan lebih praktis.",
    inputLabel: "NIS / ID Siswa",
    placeholder: "Masukkan nomor siswa",
    icon: GraduationCap,
    providers: ["SD / SMP", "SMA / SMK", "Kampus"],
    products: ["SPP", "Uang Buku", "Daftar Ulang"],
    accent: "from-blue-500 to-indigo-500",
  },
  "cicilan-kendaraan": {
    title: "Cicilan Kendaraan",
    subtitle: "Bayar angsuran kendaraan.",
    inputLabel: "Nomor Kontrak",
    placeholder: "Masukkan nomor kontrak",
    icon: Car,
    providers: ["FIF", "Adira", "WOM Finance"],
    products: ["Cek Tagihan", "Bayar Cicilan", "Denda / Admin"],
    accent: "from-sky-500 to-cyan-500",
  },
  "cicilan-multifinance": {
    title: "Cicilan Multifinance",
    subtitle: "Bayar angsuran multifinance.",
    inputLabel: "Nomor Kontrak",
    placeholder: "Masukkan nomor kontrak",
    icon: CircleDollarSign,
    providers: ["Home Credit", "Kredivo", "Akulaku"],
    products: ["Cek Tagihan", "Bayar Cicilan", "Pelunasan"],
    accent: "from-amber-500 to-yellow-500",
  },
  pbb: {
    title: "PBB",
    subtitle: "Bayar pajak bumi dan bangunan.",
    inputLabel: "NOP",
    placeholder: "Masukkan nomor objek pajak",
    icon: Home,
    providers: ["PBB Kota", "PBB Kabupaten", "Pajak Daerah"],
    products: ["Cek Tagihan", "Bayar Pajak", "Unduh Bukti"],
    accent: "from-violet-500 to-purple-500",
  },
  "pajak-negara": {
    title: "Pajak & Negara",
    subtitle: "Pembayaran administrasi negara.",
    inputLabel: "Kode Billing",
    placeholder: "Masukkan kode billing",
    icon: Landmark,
    providers: ["MPN", "Samsat", "Pajak Daerah"],
    products: ["Cek Billing", "Bayar Billing", "Simpan Bukti"],
    accent: "from-emerald-500 to-teal-500",
  },
  "tiket-perjalanan": {
    title: "Tiket Perjalanan",
    subtitle: "Pembayaran tiket dan perjalanan.",
    inputLabel: "Kode Booking",
    placeholder: "Masukkan kode booking",
    icon: BriefcaseBusiness,
    providers: ["Kereta", "Travel", "Pesawat"],
    products: ["Cek Booking", "Bayar Tiket", "Asuransi Trip"],
    accent: "from-sky-500 to-blue-500",
  },
  "saldo-kartu-tol": {
    title: "Saldo Kartu Tol",
    subtitle: "Top up kartu tol elektronik.",
    inputLabel: "Nomor Kartu",
    placeholder: "Masukkan nomor kartu",
    icon: CreditCard,
    providers: ["E-Toll", "Flazz", "TapCash"],
    products: ["Rp 50.000", "Rp 100.000", "Rp 200.000"],
    accent: "from-emerald-500 to-lime-500",
  },
  "parkir-digital": {
    title: "Parkir Digital",
    subtitle: "Bayar parkir tanpa ribet.",
    inputLabel: "Nomor Kendaraan",
    placeholder: "Contoh: B1234ABC",
    icon: Car,
    providers: ["Parkir Mall", "Parkir Kota", "Langganan"],
    products: ["1 Jam", "3 Jam", "Harian"],
    accent: "from-blue-500 to-cyan-500",
  },
  "kurir-pengiriman": {
    title: "Kurir & Pengiriman",
    subtitle: "Bayar layanan pengiriman.",
    inputLabel: "Nomor Resi",
    placeholder: "Masukkan nomor resi",
    icon: BriefcaseBusiness,
    providers: ["JNE", "J&T", "SiCepat"],
    products: ["Cek Resi", "Bayar Ongkir", "Asuransi Paket"],
    accent: "from-orange-500 to-amber-500",
  },
  "zakat-donasi": {
    title: "Zakat & Donasi",
    subtitle: "Salurkan bantuan secara digital.",
    inputLabel: "Nama Donatur",
    placeholder: "Masukkan nama donatur",
    icon: HandHeart,
    providers: ["Zakat", "Donasi Sosial", "Sedekah"],
    products: ["Rp 25.000", "Rp 50.000", "Rp 100.000"],
    accent: "from-rose-500 to-orange-500",
  },
};

const defaultService: ServiceMeta = {
  title: "Layanan Isiloka",
  subtitle: "Isi data tujuan untuk melanjutkan transaksi.",
  inputLabel: "Data Tujuan",
  placeholder: "Masukkan data tujuan",
  icon: WalletCards,
  providers: ["Isiloka", "Instan", "Reguler"],
  products: ["Rp 25.000", "Rp 50.000", "Rp 100.000"],
  accent: "from-emerald-600 to-lime-500",
};

const hpPostpaidProviders = [
  { name: "Telkomsel Halo", label: "Telkomsel", logo: "/images/providers/logo_telkomsel.webp" },
  { name: "XL Prioritas", label: "XL", logo: "/images/providers/logo_xl.png" },
  { name: "Indosat Postpaid", label: "Indosat", logo: "/images/providers/logo_im3.webp" },
  { name: "Tri Postpaid", label: "Tri", logo: "/images/providers/logo_tri.webp" },
  { name: "Smartfren Postpaid", label: "Smartfren", logo: "/images/providers/logo_smartfren.webp" },
];

const esimRegions = ["Semua", "Asia", "Eropa", "Amerika", "Afrika", "Oseania"];

const esimPopularCountries = [
  { name: "Jepang", code: "JP", region: "Asia", flag: "/images/flags/japan.svg" },
  { name: "Korea Selatan", code: "KR", region: "Asia", flag: "/images/flags/south-korea.svg" },
  { name: "Singapura", code: "SG", region: "Asia", flag: "/images/flags/singapore.svg" },
  { name: "Malaysia", code: "MY", region: "Asia", flag: "/images/flags/malaysia.svg" },
  { name: "Thailand", code: "TH", region: "Asia", flag: "/images/flags/thailand.svg" },
  { name: "Lainnya", code: "ALL", region: "Semua", flag: null },
];

const esimPackages = [
  {
    country: "Jepang",
    code: "JP",
    region: "Asia",
    quota: "10 GB",
    duration: "30 Hari",
    network: "4G/5G",
    price: 85000,
    badge: "Populer",
  },
  {
    country: "Korea Selatan",
    code: "KR",
    region: "Asia",
    quota: "15 GB",
    duration: "30 Hari",
    network: "4G/5G",
    price: 95000,
    badge: "Populer",
  },
  {
    country: "Singapura",
    code: "SG",
    region: "Asia",
    quota: "8 GB",
    duration: "30 Hari",
    network: "4G/5G",
    price: 65000,
    badge: "Populer",
  },
  {
    country: "Malaysia",
    code: "MY",
    region: "Asia",
    quota: "12 GB",
    duration: "30 Hari",
    network: "4G/5G",
    price: 75000,
    badge: "Populer",
  },
  {
    country: "Eropa",
    code: "EU",
    region: "Eropa",
    quota: "20 GB",
    duration: "30 Hari",
    network: "4G/5G",
    price: 145000,
    badge: "Multi Negara",
  },
  {
    country: "Amerika Serikat",
    code: "US",
    region: "Amerika",
    quota: "10 GB",
    duration: "15 Hari",
    network: "4G/5G",
    price: 125000,
    badge: "Favorit",
  },
];

function formatSlug(slug: string) {
  return slug
    .split("-")
    .filter(Boolean)
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

function productPrice(product: string, index: number) {
  const match = product.match(/Rp\s?([\d.]+)/i);
  if (match?.[1]) return Number(match[1].replace(/\./g, ""));
  const base = [12500, 25000, 50000, 75000, 100000];
  return base[index] || 25000;
}

function formatCurrency(value: number) {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    maximumFractionDigits: 0,
  }).format(value);
}

function hpPostpaidBillAmount(destinationValue: string, provider: string) {
  const digits = destinationValue.replace(/\D/g, "");
  const digitTotal = digits.split("").reduce((sum, digit) => sum + Number(digit), 0);
  const providerOffset = provider.length * 250;
  return 45000 + ((digitTotal * 1250 + providerOffset) % 85000);
}

function currentBillingPeriod() {
  return new Intl.DateTimeFormat("id-ID", {
    month: "long",
    year: "numeric",
  }).format(new Date());
}

function toLocalAppOrder(order: DraftOrder): UserAppOrder {
  const now = new Date().toISOString();
  const subtotal = order.total > 100000 ? order.total - 1500 : order.total - 1000;
  return {
    id: Date.now() * -1,
    invoice_id: order.invoiceId,
    member_id: null,
    member_nama: "Isiloka User",
    produk_id: 0,
    produk_sku_snapshot: `LOCAL-${order.service.toUpperCase().replace(/[^A-Z0-9]+/g, "-")}`,
    produk_nama_snapshot: `${order.service} - ${order.product}`,
    dest: order.destination,
    qty: 1,
    nominal: subtotal,
    buyer_type: "user",
    harga_dasar: subtotal,
    fee: order.total - subtotal,
    harga_final: order.total,
    status: "pending_payment",
    sn: null,
    dibuat_pada: now,
    diubah_pada: now,
  };
}

function saveLocalOrder(order: DraftOrder) {
  const key = "pulsakilat_local_service_orders";
  const current = JSON.parse(window.localStorage.getItem(key) || "[]") as UserAppOrder[];
  window.localStorage.setItem(key, JSON.stringify([toLocalAppOrder(order), ...current].slice(0, 30)));
}

function draftFromServerOrder(item: UserAppOrder, fallback: Omit<DraftOrder, "invoiceId">): DraftOrder {
  return {
    invoiceId: item.invoice_id,
    service: fallback.service,
    destination: item.dest || fallback.destination,
    provider: fallback.provider,
    product: fallback.product,
    total: Number(item.harga_final || fallback.total),
  };
}

export function UserUniversalServicePageContent({ serviceSlug }: { serviceSlug: string }) {
  const service = useMemo(() => {
    const meta = serviceMap[serviceSlug] || defaultService;
    return serviceMap[serviceSlug] ? meta : { ...meta, title: formatSlug(serviceSlug) || meta.title };
  }, [serviceSlug]);
  const Icon = service.icon;
  const [destination, setDestination] = useState("");
  const [selectedProvider, setSelectedProvider] = useState(service.providers[0]);
  const [selectedProduct, setSelectedProduct] = useState(service.products[0]);
  const [completedOrder, setCompletedOrder] = useState<DraftOrder | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState("");
  const [saveNumber, setSaveNumber] = useState(true);
  const [hpBill, setHpBill] = useState<HpPostpaidBill | null>(null);
  const [esimQuery, setEsimQuery] = useState("");
  const [selectedEsimRegion, setSelectedEsimRegion] = useState("Semua");
  const [selectedEsimPackage, setSelectedEsimPackage] = useState(esimPackages[0]);
  const [esimContact, setEsimContact] = useState("");

  const selectedProductIndex = Math.max(0, service.products.indexOf(selectedProduct));
  const subtotal = productPrice(selectedProduct, selectedProductIndex);
  const adminFee = subtotal >= 100000 ? 1500 : 1000;
  const total = subtotal + adminFee;
  const canContinue = destination.trim().length >= 3;
  const isHpPostpaid = serviceSlug === "hp-pascabayar";
  const isEsimRoaming = serviceSlug === "esim-roaming";
  const filteredEsimPackages = esimPackages.filter((item) => {
    const matchesRegion = selectedEsimRegion === "Semua" || item.region === selectedEsimRegion || item.country === selectedEsimRegion;
    const matchesSearch = !esimQuery.trim() || item.country.toLowerCase().includes(esimQuery.trim().toLowerCase());
    return matchesRegion && matchesSearch;
  });

  async function createOrder(override?: Partial<Omit<DraftOrder, "invoiceId">>) {
    if ((!override && !canContinue) || isCreating) return;
    const orderData = {
      service: override?.service || service.title,
      destination: override?.destination || destination.trim(),
      provider: override?.provider || selectedProvider,
      product: override?.product || selectedProduct,
      total: override?.total || total,
    };
    setIsCreating(true);
    setError("");
    try {
      const res = await fetch("/api/app/universal-orders", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(orderData),
      });
      const body = (await res.json().catch(() => ({}))) as UniversalOrderResponse;
      if (res.ok && body.ok && body.item) {
        const serverOrder = draftFromServerOrder(body.item, orderData);
        setCompletedOrder(serverOrder);
        return;
      }
      throw new Error(body.error || "Transaksi gagal disimpan ke server");
    } catch (err) {
      const order: DraftOrder = {
      invoiceId: `PK${Date.now().toString().slice(-9)}`,
        ...orderData,
      };
      saveLocalOrder(order);
      setCompletedOrder(order);
      setError(err instanceof Error ? `${err.message}. Riwayat sementara disimpan di perangkat ini.` : "Riwayat sementara disimpan di perangkat ini.");
    } finally {
      setIsCreating(false);
    }
  }

  function checkHpPostpaidBill() {
    if (!canContinue || isCreating) return;
    const nominal = hpPostpaidBillAmount(destination, selectedProvider);
    const nextBill = {
      customerName: "Pelanggan Isiloka",
      destination: destination.trim(),
      provider: selectedProvider,
      period: currentBillingPeriod(),
      nominal,
      adminFee: 1500,
      total: nominal + 1500,
    };
    setError("");
    setHpBill(nextBill);
  }

  function payHpPostpaidBill() {
    if (!hpBill) return;
    void createOrder({
      service: "HP Pascabayar",
      destination: hpBill.destination,
      provider: hpBill.provider,
      product: `Bayar Tagihan ${hpBill.period}`,
      total: hpBill.total,
    });
  }

  function buyEsimPackage() {
    if (!selectedEsimPackage || isCreating) return;
    void createOrder({
      service: "eSIM & Roaming",
      destination: esimContact.trim() || selectedEsimPackage.country,
      provider: selectedEsimPackage.country,
      product: `${selectedEsimPackage.quota} / ${selectedEsimPackage.duration} ${selectedEsimPackage.network}`,
      total: selectedEsimPackage.price,
    });
  }

  if (isEsimRoaming && !completedOrder) {
    return (
      <main className="min-h-screen bg-[#f4fbf7] pb-36 text-[#07112e]">
        <section className="relative overflow-hidden bg-[radial-gradient(circle_at_82%_2%,rgba(255,255,255,0.18),transparent_34%),linear-gradient(145deg,#006f38_0%,#009f4d_55%,#006934_100%)] px-5 pb-20 pt-8 text-white shadow-[0_18px_42px_rgba(0,105,54,0.22)]">
          <div className="mx-auto flex w-full max-w-md items-center justify-between">
            <Link href="/user/kategori" className="grid h-11 w-11 shrink-0 place-items-center rounded-full text-white transition hover:bg-white/12">
              <ArrowLeft className="h-8 w-8" strokeWidth={2.8} />
            </Link>
            <h1 className="min-w-0 flex-1 px-3 text-center text-[24px] font-extrabold leading-tight text-white min-[390px]:text-[29px]">
              eSIM & Roaming
            </h1>
            <Link href="/user/transaksi" className="grid h-11 w-11 shrink-0 place-items-center rounded-full text-white transition hover:bg-white/12">
              <Clock3 className="h-8 w-8" strokeWidth={2.8} />
            </Link>
          </div>
        </section>

        <div className="relative z-10 mx-auto -mt-12 w-full max-w-md space-y-5 px-4">
          <section className="overflow-hidden rounded-[30px] border border-emerald-100 bg-white shadow-[0_18px_44px_rgba(15,23,42,0.12)]">
            <div className="relative min-h-[220px] overflow-hidden bg-[radial-gradient(circle_at_78%_42%,rgba(0,122,61,0.16),transparent_28%),linear-gradient(135deg,#f8fffb,#e8f8ee)] px-5 py-6">
              <div className="pointer-events-none absolute right-6 top-8 hidden h-28 w-44 rounded-full border border-emerald-200/80 min-[380px]:block" />
              <div className="pointer-events-none absolute right-12 top-14 hidden h-20 w-32 rounded-full border border-emerald-200/70 min-[380px]:block" />
              <div className="relative z-10 max-w-[210px]">
                <h2 className="text-[30px] font-extrabold leading-tight text-[#063f2e]">
                  eSIM Global, Internet di Mana Saja
                </h2>
                <p className="mt-4 text-base font-medium leading-relaxed text-slate-700">
                  Aktifkan eSIM instan, tanpa kartu fisik, praktis & terpercaya.
                </p>
              </div>
              <div className="absolute bottom-5 right-4 grid h-[132px] w-[118px] place-items-center rounded-[28px] border-[7px] border-[#063f2e] bg-white shadow-[0_18px_34px_rgba(6,78,59,0.20)]">
                <Image src="/service-icons/esim-roaming.png" alt="eSIM" width={72} height={72} className="h-16 w-16 object-contain" />
                <span className="absolute -right-3 -top-3 grid h-[52px] w-[52px] place-items-center rounded-full bg-[#009f4d] text-white shadow-lg">
                  <Plane className="h-7 w-7" strokeWidth={2.7} />
                </span>
              </div>
            </div>

            <div className="space-y-5 p-4">
              <div className="flex h-[58px] items-center overflow-hidden rounded-[20px] border border-slate-200 bg-white shadow-[0_6px_18px_rgba(15,23,42,0.04)] focus-within:border-[#008a42] focus-within:ring-4 focus-within:ring-emerald-100">
                <span className="grid h-full w-14 shrink-0 place-items-center text-slate-400">
                  <Search className="h-7 w-7" strokeWidth={2.3} />
                </span>
                <input
                  value={esimQuery}
                  onChange={(event) => setEsimQuery(event.target.value)}
                  placeholder="Cari negara tujuan"
                  className="min-w-0 flex-1 bg-transparent text-base font-semibold text-[#07112e] outline-none placeholder:text-slate-400"
                />
              </div>

              <div>
                <div className="flex items-center justify-between gap-3">
                  <h3 className="text-[21px] font-extrabold leading-none text-[#07112e]">Negara Populer</h3>
                  <button
                    type="button"
                    onClick={() => {
                      setSelectedEsimRegion("Semua");
                      setEsimQuery("");
                    }}
                    className="flex shrink-0 items-center gap-1 rounded-full px-2 py-1 text-sm font-black text-[#007a3d] transition hover:bg-emerald-50"
                  >
                    Lihat Semua
                    <ChevronRight className="h-5 w-5" strokeWidth={2.8} />
                  </button>
                </div>
                <div className="-mx-1 mt-4 flex snap-x gap-3 overflow-x-auto px-1 pb-2 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
                  {esimPopularCountries.map((country) => {
                    const active = esimQuery === country.name || selectedEsimRegion === country.name;
                    return (
                      <button
                        key={country.name}
                        type="button"
                        onClick={() => {
                          setEsimQuery(country.name === "Lainnya" ? "" : country.name);
                          setSelectedEsimRegion(country.name === "Lainnya" ? "Semua" : country.region);
                        }}
                        className={[
                          "flex min-h-[126px] w-[92px] shrink-0 snap-start flex-col items-center justify-center rounded-[20px] border bg-white px-2 py-3 text-center shadow-[0_10px_22px_rgba(15,23,42,0.06)] transition min-[420px]:w-[96px]",
                          active ? "border-[#009f4d] bg-[#f2fff8] ring-2 ring-emerald-100" : "border-slate-200 hover:border-emerald-200",
                        ].join(" ")}
                      >
                        <span className="grid h-[54px] w-[66px] place-items-center overflow-hidden rounded-[16px] bg-[#f8fafc] text-sm font-black text-[#007a3d] shadow-[inset_0_0_0_1px_rgba(15,23,42,0.05),0_8px_18px_rgba(15,23,42,0.08)]">
                          {country.flag ? (
                            <Image
                              src={country.flag}
                              alt={`Bendera ${country.name}`}
                              width={72}
                              height={54}
                              className="h-full w-full object-cover"
                            />
                          ) : (
                            <span className="grid h-11 w-11 grid-cols-2 place-items-center gap-1 rounded-full bg-emerald-50 p-2">
                              <span className="h-3 w-3 rounded-full bg-[#008a42]" />
                              <span className="h-3 w-3 rounded-full bg-[#008a42]" />
                              <span className="h-3 w-3 rounded-full bg-[#008a42]" />
                              <span className="h-3 w-3 rounded-full bg-[#008a42]" />
                            </span>
                          )}
                        </span>
                        <span className="mt-3 line-clamp-2 min-h-[30px] text-[12px] font-bold leading-tight text-[#07112e]">{country.name}</span>
                      </button>
                    );
                  })}
                </div>
              </div>

              <div className="flex gap-2 overflow-x-auto rounded-[20px] border border-slate-200 bg-white p-1.5">
                {esimRegions.map((region) => {
                  const active = selectedEsimRegion === region && !esimQuery;
                  return (
                    <button
                      key={region}
                      type="button"
                      onClick={() => {
                        setSelectedEsimRegion(region);
                        setEsimQuery("");
                      }}
                      className={[
                        "h-10 shrink-0 rounded-[15px] px-5 text-sm font-black transition",
                        active ? "bg-[linear-gradient(135deg,#009f4d,#007a3d)] text-white shadow-[0_8px_16px_rgba(0,122,61,0.18)]" : "text-[#07112e] hover:bg-emerald-50",
                      ].join(" ")}
                    >
                      {region}
                    </button>
                  );
                })}
              </div>

              <div>
                <div className="flex items-center justify-between">
                  <h3 className="text-xl font-extrabold text-[#07112e]">Paket eSIM</h3>
                  <button type="button" className="rounded-full border border-slate-200 bg-white px-4 py-2 text-sm font-bold text-[#07112e]">
                    Terpopuler
                  </button>
                </div>

                <div className="mt-3 space-y-3">
                  {filteredEsimPackages.map((pkg) => {
                    const active = selectedEsimPackage.country === pkg.country;
                    return (
                      <button
                        key={pkg.country}
                        type="button"
                        onClick={() => setSelectedEsimPackage(pkg)}
                        className={[
                          "grid w-full grid-cols-[minmax(0,1fr)_auto] items-center gap-3 rounded-[22px] border bg-white p-4 text-left shadow-[0_10px_24px_rgba(15,23,42,0.06)] transition max-[360px]:grid-cols-1",
                          active ? "border-[#009f4d] ring-2 ring-emerald-100" : "border-slate-200 hover:border-emerald-200",
                        ].join(" ")}
                      >
                        <span className="min-w-0">
                          <span className="flex items-center gap-2">
                            <span className="rounded-xl bg-emerald-50 px-2.5 py-1 text-xs font-black text-[#007a3d]">{pkg.code}</span>
                            <span className="block text-lg font-extrabold leading-tight text-[#07112e]">{pkg.country}</span>
                          </span>
                          <span className="mt-3 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs font-semibold text-[#07112e]">
                            <span className="inline-flex items-center gap-1"><Wifi className="h-4 w-4 text-[#008a42]" />{pkg.quota}</span>
                            <span className="inline-flex items-center gap-1"><CalendarDays className="h-4 w-4 text-[#008a42]" />{pkg.duration}</span>
                            <span className="inline-flex items-center gap-1"><Signal className="h-4 w-4 text-[#008a42]" />{pkg.network}</span>
                          </span>
                          <span className="mt-2 inline-flex rounded-full bg-emerald-100 px-3 py-1 text-[11px] font-black text-[#007a3d]">{pkg.badge}</span>
                        </span>
                        <span className="text-right max-[360px]:flex max-[360px]:items-center max-[360px]:justify-between">
                          <span className="block text-xs font-semibold text-slate-500">Mulai dari</span>
                          <span className="block text-xl font-extrabold text-[#007a3d]">{formatCurrency(pkg.price)}</span>
                          <ChevronRight className="ml-auto mt-2 h-7 w-7 text-[#007a3d] max-[360px]:mt-0" strokeWidth={2.8} />
                        </span>
                      </button>
                    );
                  })}
                  {filteredEsimPackages.length === 0 ? (
                    <div className="rounded-[22px] border border-dashed border-emerald-200 bg-[#f7fffb] px-4 py-8 text-center text-sm font-semibold text-slate-500">
                      Paket untuk negara ini belum tersedia.
                    </div>
                  ) : null}
                </div>
              </div>

              <section className="rounded-[24px] border border-emerald-100 bg-[linear-gradient(135deg,#f4fff8,#ffffff)] p-4 shadow-[0_12px_26px_rgba(0,122,61,0.08)]">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="text-[11px] font-black uppercase tracking-[0.18em] text-[#008a42]">Checkout eSIM</p>
                    <h3 className="mt-1 text-xl font-extrabold text-[#07112e]">{selectedEsimPackage.country}</h3>
                    <p className="mt-1 text-sm font-semibold text-slate-500">
                      {selectedEsimPackage.quota} / {selectedEsimPackage.duration} / {selectedEsimPackage.network}
                    </p>
                  </div>
                  <span className="text-right text-xl font-extrabold text-[#007a3d]">{formatCurrency(selectedEsimPackage.price)}</span>
                </div>
                <div className="mt-4 flex h-[56px] items-center overflow-hidden rounded-[18px] border border-slate-200 bg-white focus-within:border-[#008a42] focus-within:ring-4 focus-within:ring-emerald-100">
                  <span className="grid h-full w-[52px] shrink-0 place-items-center text-[#008a42]">
                    <Smartphone className="h-6 w-6" strokeWidth={2.5} />
                  </span>
                  <input
                    value={esimContact}
                    onChange={(event) => setEsimContact(event.target.value)}
                    placeholder="Email atau WhatsApp penerima"
                    className="min-w-0 flex-1 bg-transparent text-sm font-semibold text-[#07112e] outline-none placeholder:text-slate-400"
                  />
                </div>
                <button
                  type="button"
                  disabled={isCreating}
                  onClick={buyEsimPackage}
                  className="mt-4 flex h-[58px] w-full items-center justify-center gap-2 rounded-[18px] bg-[linear-gradient(135deg,#13464b,#008a42,#65c82f)] text-base font-extrabold text-white shadow-[0_16px_28px_rgba(0,122,61,0.24)] disabled:bg-none disabled:bg-slate-200 disabled:text-slate-500"
                >
                  <Rocket className="h-5 w-5" strokeWidth={2.5} />
                  {isCreating ? "MEMPROSES..." : "BELI PAKET eSIM"}
                </button>
              </section>
            </div>
          </section>

          <section className="grid grid-cols-2 gap-2 rounded-[24px] border border-emerald-100 bg-white p-3 shadow-[0_12px_28px_rgba(15,23,42,0.06)] min-[390px]:grid-cols-4">
            {[
              ["Instan", "Aktif setelah pembelian", Rocket],
              ["Aman", "Transaksi terlindungi", ShieldCheck],
              ["Global", "Tersedia banyak negara", Plane],
              ["24/7", "Bantuan kapan saja", Headphones],
            ].map(([title, desc, BenefitIcon]) => (
              <div key={title as string} className="flex items-center gap-2 rounded-[18px] bg-[#f4fff8] p-2 min-[390px]:block min-[390px]:text-center">
                <span className="grid h-10 w-10 shrink-0 place-items-center rounded-full bg-[#008a42] text-white min-[390px]:mx-auto">
                  <BenefitIcon className="h-5 w-5" strokeWidth={2.5} />
                </span>
                <span className="min-w-0">
                  <span className="block text-sm font-black text-[#007a3d]">{title as string}</span>
                  <span className="mt-0.5 block text-[11px] font-medium leading-snug text-[#07112e]">{desc as string}</span>
                </span>
              </div>
            ))}
          </section>
        </div>
      </main>
    );
  }

  if (isHpPostpaid && !completedOrder) {
    return (
      <main className="min-h-screen overflow-x-hidden bg-[#edf7f2] pb-32 text-slate-950">
        <header className="relative overflow-hidden bg-[linear-gradient(140deg,#052e26_0%,#047857_64%,#38b94d_115%)] px-4 pb-10 pt-5 text-white shadow-[0_16px_36px_rgba(5,46,38,0.22)]">
          <div className="mx-auto w-full max-w-md">
            <div className="grid grid-cols-[44px_minmax(0,1fr)_44px] items-center gap-2">
              <Link href="/user/kategori" aria-label="Kembali" className="grid h-11 w-11 place-items-center rounded-2xl bg-white/10 ring-1 ring-white/15 transition hover:bg-white/15">
                <ArrowLeft className="h-5 w-5" strokeWidth={2.5} />
              </Link>
              <h1 className="truncate text-center text-xl font-black">HP Pascabayar</h1>
              <Link href="/user/transaksi" aria-label="Riwayat transaksi" className="grid h-11 w-11 place-items-center rounded-2xl bg-white/10 ring-1 ring-white/15 transition hover:bg-white/15">
                <Clock3 className="h-5 w-5" strokeWidth={2.5} />
              </Link>
            </div>

            <div className="mt-7 flex items-center gap-4">
              <span className="grid h-14 w-14 shrink-0 place-items-center rounded-2xl bg-white text-[#087e8b] shadow-[0_12px_26px_rgba(0,0,0,0.16)]">
                <ReceiptText className="h-7 w-7" strokeWidth={2.5} />
              </span>
              <div className="min-w-0">
                <p className="text-[11px] font-black uppercase tracking-[0.18em] text-lime-100">Tagihan Seluler</p>
                <p className="mt-1 text-sm font-semibold leading-5 text-emerald-50">Cek nomor pelanggan dan selesaikan tagihan dalam satu proses.</p>
              </div>
            </div>
          </div>
        </header>

        <div className="relative z-10 mx-auto -mt-4 w-full max-w-md space-y-4 px-3 min-[390px]:px-4">
          <section className="overflow-hidden rounded-[24px] border border-emerald-100 bg-white shadow-[0_16px_38px_rgba(15,23,42,0.09)]">
            <div className="border-b border-slate-100 px-4 py-5">
              <div className="flex items-center justify-between gap-3">
                <div>
                  <p className="text-[10px] font-black uppercase tracking-[0.18em] text-[#087e8b]">Langkah 1</p>
                  <h2 className="mt-1 text-lg font-black">Pilih operator</h2>
                </div>
                <span className="rounded-full bg-emerald-50 px-3 py-1.5 text-[10px] font-black text-[#087e8b]">{hpPostpaidProviders.length} tersedia</span>
              </div>

              <div className="mt-4 grid grid-cols-3 gap-2">
                {hpPostpaidProviders.map((provider) => {
                  const active = selectedProvider === provider.name || (provider.name === "Telkomsel Halo" && selectedProvider === "Telkomsel");
                  return (
                    <button
                      key={provider.name}
                      type="button"
                      onClick={() => {
                        setSelectedProvider(provider.name);
                        setHpBill(null);
                      }}
                      aria-pressed={active}
                      className={[
                        "relative flex min-h-[88px] min-w-0 flex-col items-center justify-center rounded-[16px] border px-2 py-3 text-center transition",
                        active
                          ? "border-[#087e8b] bg-emerald-50 shadow-[0_8px_18px_rgba(4,120,87,0.10)] ring-1 ring-[#087e8b]"
                          : "border-slate-200 bg-white hover:border-emerald-300",
                      ].join(" ")}
                    >
                      {active ? <CheckCircle2 className="absolute right-2 top-2 h-4 w-4 text-[#087e8b]" strokeWidth={2.8} /> : null}
                      <span className="grid h-9 w-full place-items-center overflow-hidden">
                        <Image src={provider.logo} alt={`Logo ${provider.label}`} width={60} height={36} className="max-h-8 max-w-[58px] object-contain" />
                      </span>
                      <span className="mt-2 line-clamp-1 w-full text-[10px] font-black text-slate-800 min-[390px]:text-[11px]">{provider.label}</span>
                    </button>
                  );
                })}
              </div>
            </div>

            <div className="px-4 py-5">
              <p className="text-[10px] font-black uppercase tracking-[0.18em] text-[#087e8b]">Langkah 2</p>
              <label htmlFor="hp-postpaid-number" className="mt-1 block text-lg font-black">Nomor pelanggan</label>
              <div className="mt-3 flex min-h-14 items-center overflow-hidden rounded-[16px] border border-slate-200 bg-slate-50 transition focus-within:border-[#087e8b] focus-within:bg-white focus-within:ring-4 focus-within:ring-emerald-100">
                <span className="grid h-14 w-12 shrink-0 place-items-center text-[#087e8b]">
                  <Smartphone className="h-5 w-5" strokeWidth={2.4} />
                </span>
                <input
                  id="hp-postpaid-number"
                  value={destination}
                  onChange={(event) => {
                    setDestination(event.target.value.replace(/\D+/g, ""));
                    setHpBill(null);
                  }}
                  inputMode="numeric"
                  pattern="[0-9]*"
                  placeholder="Contoh: 081234567890"
                  className="h-14 min-w-0 flex-1 bg-transparent pr-3 text-base font-bold text-slate-950 outline-none placeholder:text-sm placeholder:font-semibold placeholder:text-slate-400"
                />
              </div>

              <div className="mt-4 flex min-h-11 items-center justify-between gap-4 rounded-[14px] bg-slate-50 px-3 py-2">
                <div className="min-w-0">
                  <p className="text-sm font-bold text-slate-800">Simpan nomor</p>
                  <p className="truncate text-[10px] font-semibold text-slate-400">Untuk transaksi berikutnya</p>
                </div>
                <button
                  type="button"
                  aria-label="Simpan nomor pelanggan"
                  aria-pressed={saveNumber}
                  onClick={() => setSaveNumber((value) => !value)}
                  className={[
                    "relative h-7 w-12 shrink-0 rounded-full transition",
                    saveNumber ? "bg-[#087e8b]" : "bg-slate-300",
                  ].join(" ")}
                >
                  <span className={["absolute top-1 h-5 w-5 rounded-full bg-white shadow-sm transition", saveNumber ? "left-6" : "left-1"].join(" ")} />
                </button>
              </div>

              {error ? <p role="alert" className="mt-4 rounded-[14px] border border-rose-200 bg-rose-50 px-3 py-3 text-xs font-bold leading-5 text-rose-700">{error}</p> : null}

              <button
                type="button"
                disabled={!canContinue || isCreating}
                onClick={checkHpPostpaidBill}
                className="mt-5 flex h-14 w-full items-center justify-center gap-2 rounded-[16px] bg-[#087e8b] px-4 text-sm font-black text-white shadow-[0_12px_24px_rgba(4,120,87,0.20)] transition hover:bg-[#036b4e] disabled:bg-slate-200 disabled:text-slate-400 disabled:shadow-none"
              >
                <Search className="h-4 w-4" strokeWidth={2.7} />
                {isCreating ? "Memeriksa..." : "Cek Tagihan"}
              </button>
            </div>
          </section>

          {hpBill ? (
            <section className="overflow-hidden rounded-[24px] border border-emerald-100 bg-white shadow-[0_16px_38px_rgba(15,23,42,0.09)]">
              <div className="flex items-center justify-between gap-3 border-b border-slate-100 px-4 py-4">
                <div className="min-w-0">
                  <p className="text-[10px] font-black uppercase tracking-[0.18em] text-[#087e8b]">Tagihan Ditemukan</p>
                  <h2 className="mt-1 truncate text-lg font-black">{hpBill.provider}</h2>
                  <p className="mt-0.5 text-xs font-semibold text-slate-400">Periode {hpBill.period}</p>
                </div>
                <span className="grid h-11 w-11 shrink-0 place-items-center rounded-[14px] bg-emerald-50 text-[#087e8b]">
                  <ReceiptText className="h-5 w-5" strokeWidth={2.5} />
                </span>
              </div>

              <dl className="divide-y divide-slate-100 px-4">
                {[
                  ["Nama pelanggan", hpBill.customerName],
                  ["Nomor pelanggan", hpBill.destination],
                  ["Nominal tagihan", formatCurrency(hpBill.nominal)],
                  ["Biaya admin", formatCurrency(hpBill.adminFee)],
                ].map(([label, value]) => (
                  <div key={label} className="grid min-h-12 grid-cols-[minmax(0,1fr)_minmax(0,1.25fr)] items-center gap-3 py-3">
                    <dt className="text-xs font-semibold text-slate-500">{label}</dt>
                    <dd className="break-words text-right text-sm font-black text-slate-900">{value}</dd>
                  </div>
                ))}
              </dl>

              <div className="bg-[#13464b] px-4 py-4 text-white">
                <div className="flex items-end justify-between gap-3">
                  <span className="text-xs font-bold text-emerald-100">Total pembayaran</span>
                  <span className="text-xl font-black text-white min-[390px]:text-2xl">{formatCurrency(hpBill.total)}</span>
                </div>
              </div>

              <div className="p-4">
                <div className="flex items-center gap-3 rounded-[16px] border border-emerald-200 bg-emerald-50 px-3 py-3">
                  <span className="grid h-10 w-10 shrink-0 place-items-center rounded-[12px] bg-[#087e8b] text-white">
                    <WalletCards className="h-5 w-5" strokeWidth={2.5} />
                  </span>
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-black text-slate-900">Saldo Isiloka</p>
                    <p className="text-[10px] font-semibold text-slate-500">Pembayaran dari saldo utama</p>
                  </div>
                  <CheckCircle2 className="h-5 w-5 shrink-0 text-[#087e8b]" strokeWidth={2.7} />
                </div>

                <button
                  type="button"
                  disabled={isCreating}
                  onClick={payHpPostpaidBill}
                  className="mt-4 flex h-14 w-full items-center justify-center rounded-[16px] bg-[#087e8b] px-4 text-sm font-black text-white shadow-[0_12px_24px_rgba(4,120,87,0.20)] transition hover:bg-[#036b4e] disabled:bg-slate-200 disabled:text-slate-400 disabled:shadow-none"
                >
                  {isCreating ? "Memproses..." : "Bayar Sekarang"}
                </button>
              </div>
            </section>
          ) : null}
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-[#f0f8f8] pb-24">
      <section className="relative overflow-hidden bg-[linear-gradient(135deg,#052e26_0%,#047857_58%,#a3e635_150%)] px-4 pb-9 pt-5 text-white">
        <div className="pointer-events-none absolute -right-14 -top-16 h-44 w-44 rounded-full bg-lime-300/20 blur-2xl" />
        <div className="pointer-events-none absolute -bottom-16 left-8 h-36 w-56 rounded-full bg-emerald-300/15 blur-2xl" />
        <div className="relative mx-auto flex w-full max-w-md items-center gap-3">
          <Link href="/user/kategori" className="grid h-10 w-10 shrink-0 place-items-center rounded-2xl bg-white/12 ring-1 ring-white/15">
            <ArrowLeft className="h-5 w-5" strokeWidth={2.4} />
          </Link>
          <div className="min-w-0 flex-1">
            <h1 className="truncate text-lg font-black">{service.title}</h1>
            <p className="mt-0.5 truncate text-[11px] font-semibold text-white/75">{service.subtitle}</p>
          </div>
          <span className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-white text-[#087e8b] shadow-[0_12px_26px_rgba(0,0,0,0.16)]">
            <Icon className="h-5 w-5" strokeWidth={2.5} />
          </span>
        </div>

        <div className="relative mx-auto mt-5 grid w-full max-w-md grid-cols-3 gap-2">
          {["Data", "Produk", "Bayar"].map((step, index) => (
            <div key={step} className="rounded-2xl bg-white/10 px-3 py-2 ring-1 ring-white/12">
              <p className="text-[10px] font-black text-lime-100">0{index + 1}</p>
              <p className="mt-0.5 text-xs font-black">{step}</p>
            </div>
          ))}
        </div>
      </section>

      <div className="mx-auto -mt-5 w-full max-w-md space-y-4 px-4">
        {completedOrder ? (
          <section className="overflow-hidden rounded-[30px] border border-emerald-100 bg-white shadow-[0_22px_50px_rgba(6,78,59,0.14)]">
            <div className="bg-[linear-gradient(135deg,#13464b,#087e8b,#84cc16)] px-5 py-6 text-center text-white">
              <div className="mx-auto grid h-18 w-18 place-items-center rounded-full bg-white text-[#087e8b] shadow-lg">
                <Check className="h-9 w-9" strokeWidth={3} />
              </div>
              <h2 className="mt-4 text-xl font-black">Transaksi Dibuat</h2>
              <p className="mt-1 text-xs font-semibold text-white/75">Invoice siap dibayar.</p>
              {error ? <p className="mx-auto mt-3 max-w-[260px] rounded-2xl bg-white/12 px-3 py-2 text-[10px] font-bold text-white/85">{error}</p> : null}
            </div>
            <div className="space-y-3 p-5">
              {[
                ["Invoice", completedOrder.invoiceId],
                ["Layanan", completedOrder.service],
                ["Tujuan", completedOrder.destination],
                ["Produk", completedOrder.product],
                ["Total", formatCurrency(completedOrder.total)],
              ].map(([label, value]) => (
                <div key={label} className="flex items-center justify-between gap-3 rounded-2xl bg-emerald-50 px-4 py-3">
                  <span className="text-[11px] font-bold text-slate-500">{label}</span>
                  <span className="max-w-[190px] truncate text-right text-sm font-black text-slate-950">{value}</span>
                </div>
              ))}
              <button
                type="button"
                onClick={() => navigator.clipboard?.writeText(completedOrder.invoiceId)}
                className="flex h-12 w-full items-center justify-center gap-2 rounded-[18px] border-2 border-[#087e8b] bg-white text-sm font-black !text-[#087e8b] shadow-[0_10px_20px_rgba(6,78,59,0.08)]"
                style={{ color: "#087e8b" }}
              >
                <Copy className="h-4 w-4" />
                Salin Invoice
              </button>
              <Link
                href="/user/transaksi"
                className="flex h-13 w-full items-center justify-center gap-2 rounded-[20px] border-2 border-[#13464b] bg-[linear-gradient(135deg,#064e3b_0%,#047857_72%,#55c72f_145%)] text-sm font-black !text-white shadow-[0_14px_28px_rgba(6,78,59,0.20)]"
                style={{ color: "#ffffff" }}
              >
                <span className="!text-white" style={{ color: "#ffffff" }}>Lihat Riwayat</span>
                <ChevronRight className="h-4 w-4 !text-white" strokeWidth={2.6} />
              </Link>
            </div>
          </section>
        ) : (
          <>
            <section className="rounded-[28px] border border-emerald-950/5 bg-white p-4 shadow-[0_18px_42px_rgba(6,78,59,0.10)]">
              <div className="mb-3 flex items-center gap-3">
                <span className={`grid h-12 w-12 place-items-center rounded-2xl bg-gradient-to-br ${service.accent} text-white shadow-[0_14px_28px_rgba(6,78,59,0.16)]`}>
                  <Icon className="h-6 w-6" strokeWidth={2.5} />
                </span>
                <div>
                  <p className="text-sm font-black text-slate-950">{service.inputLabel}</p>
                  <p className="text-[11px] font-semibold text-slate-400">Pastikan data tujuan sudah benar.</p>
                </div>
              </div>
              <div className="flex h-14 overflow-hidden rounded-[20px] border border-emerald-200 bg-[#fbfffd] focus-within:border-[#087e8b] focus-within:ring-4 focus-within:ring-emerald-100">
                <span className="grid w-14 shrink-0 place-items-center border-r border-emerald-100 text-[#087e8b]">
                  ID
                </span>
                <input
                  value={destination}
                  onChange={(event) => setDestination(event.target.value)}
                  placeholder={service.placeholder}
                  className="min-w-0 flex-1 bg-transparent px-4 text-sm font-black text-slate-950 outline-none placeholder:text-slate-400"
                />
              </div>
            </section>

            <section className="rounded-[28px] border border-emerald-950/5 bg-white p-4 shadow-[0_18px_42px_rgba(6,78,59,0.10)]">
              <div className="mb-3 flex items-center justify-between">
                <h2 className="text-base font-black text-slate-950">Pilih Penyedia</h2>
                <span className="rounded-full bg-lime-100 px-3 py-1 text-[10px] font-black text-[#087e8b]">Tersedia</span>
              </div>
              <div className="grid grid-cols-3 gap-2.5">
                {service.providers.map((provider) => {
                  const active = selectedProvider === provider;
                  return (
                    <button
                      key={provider}
                      type="button"
                      onClick={() => setSelectedProvider(provider)}
                      className={
                        active
                          ? "relative min-h-[104px] rounded-[24px] border border-[#087e8b] bg-emerald-50 p-2.5 text-center shadow-[0_16px_30px_rgba(4,120,87,0.15)]"
                          : "min-h-[104px] rounded-[24px] border border-slate-200 bg-white p-2.5 text-center shadow-[0_10px_22px_rgba(15,23,42,0.04)] transition hover:border-emerald-200"
                      }
                    >
                      {active ? <CheckCircle2 className="absolute right-2 top-2 h-4 w-4 text-[#087e8b]" /> : null}
                      <span className={`mx-auto grid h-12 w-12 place-items-center rounded-2xl bg-gradient-to-br ${service.accent} text-white`}>
                        <Icon className="h-5 w-5" strokeWidth={2.4} />
                      </span>
                      <span className="mt-2 line-clamp-2 block text-[10px] font-black leading-tight text-slate-950">{provider}</span>
                    </button>
                  );
                })}
              </div>
            </section>

            <section className="rounded-[28px] border border-emerald-950/5 bg-white p-4 shadow-[0_18px_42px_rgba(6,78,59,0.10)]">
              <h2 className="text-base font-black text-slate-950">Pilih Produk</h2>
              <div className="mt-3 grid grid-cols-2 gap-3">
                {service.products.map((product, index) => {
                  const active = selectedProduct === product;
                  return (
                    <button
                      key={product}
                      type="button"
                      onClick={() => setSelectedProduct(product)}
                      className={
                        active
                          ? "relative overflow-hidden rounded-[22px] border border-[#087e8b] bg-[#ecfdf5] px-3 py-4 text-left shadow-[0_14px_26px_rgba(4,120,87,0.13)]"
                          : "rounded-[22px] border border-slate-200 bg-white px-3 py-4 text-left shadow-sm"
                      }
                    >
                      {active ? <Sparkles className="absolute right-3 top-3 h-4 w-4 text-[#087e8b]" /> : null}
                      <span className="block pr-5 text-sm font-black text-slate-950">{product}</span>
                      <span className="mt-2 block text-xs font-black text-[#087e8b]">{formatCurrency(productPrice(product, index))}</span>
                    </button>
                  );
                })}
              </div>
            </section>

            <section className="rounded-[28px] border border-emerald-950/5 bg-white p-4 shadow-[0_18px_42px_rgba(6,78,59,0.10)]">
              <h2 className="text-base font-black text-slate-950">Ringkasan</h2>
              <div className="mt-3 space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="font-semibold text-slate-500">Produk</span>
                  <span className="font-black text-slate-950">{formatCurrency(subtotal)}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="font-semibold text-slate-500">Biaya admin</span>
                  <span className="font-black text-slate-950">{formatCurrency(adminFee)}</span>
                </div>
                <div className="h-px bg-slate-100" />
                <div className="flex justify-between">
                  <span className="text-sm font-black text-slate-950">Total bayar</span>
                  <span className="text-lg font-black text-[#087e8b]">{formatCurrency(total)}</span>
                </div>
              </div>
            </section>

            <button
              type="button"
              disabled={!canContinue || isCreating}
              className="flex h-14 w-full items-center justify-center gap-2 rounded-[22px] bg-[linear-gradient(135deg,#13464b,#087e8b,#84cc16)] text-sm font-black text-white shadow-[0_18px_34px_rgba(4,120,87,0.22)] transition hover:brightness-105 disabled:bg-none disabled:bg-slate-200 disabled:text-slate-400 disabled:shadow-none"
              onClick={() => createOrder()}
            >
              {isCreating ? "Menyimpan..." : "Buat Transaksi"}
              <ChevronRight className="h-4 w-4" strokeWidth={2.6} />
            </button>
          </>
        )}
      </div>
    </main>
  );
}
