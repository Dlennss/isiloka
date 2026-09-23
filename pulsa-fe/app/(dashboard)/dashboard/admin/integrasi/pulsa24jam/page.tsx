import Link from "next/link";
import {
  Activity,
  ArrowUpRight,
  CheckCircle2,
  ClipboardList,
  ExternalLink,
  KeyRound,
  Link2,
  MapPinned,
  PlugZap,
  ShieldCheck,
  Webhook,
  XCircle,
  Zap,
} from "lucide-react";

function configured(value?: string) {
  return Boolean(String(value || "").trim());
}

function mask(value?: string) {
  const text = String(value || "").trim();
  if (!text) return "Belum diisi";
  if (text.length <= 8) return "****";
  return `${text.slice(0, 4)}****${text.slice(-4)}`;
}

function appBaseURL() {
  return (
    process.env.NEXTAUTH_URL ||
    process.env.NEXT_PUBLIC_SITE_URL ||
    process.env.APP_URL ||
    "https://domain-pulsakilat-kamu.com"
  ).replace(/\/$/, "");
}

function backendBaseURL() {
  return (
    process.env.API_BASE ||
    process.env.NEXT_PUBLIC_API_BASE ||
    appBaseURL()
  ).replace(/\/$/, "");
}

const envItems = [
  {
    key: "PULSA24JAM_BASE_URL",
    label: "Base URL",
    value: process.env.PULSA24JAM_BASE_URL,
    desc: "Isi https://api.pulsa24jam.net atau base URL resmi provider.",
  },
  {
    key: "PULSA24JAM_API_KEY",
    label: "API Key",
    value: process.env.PULSA24JAM_API_KEY,
    desc: "Dikirim dari backend melalui header X-Api-Key.",
  },
  {
    key: "PULSA24JAM_PIN",
    label: "PIN",
    value: process.env.PULSA24JAM_PIN,
    desc: "Dikirim di body request sebagai pin.",
  },
  {
    key: "PULSA24JAM_PASSWORD",
    label: "Password",
    value: process.env.PULSA24JAM_PASSWORD,
    desc: "Disimpan di server sebagai credential akun Pulsa24Jam.",
  },
  {
    key: "PULSA24JAM_CALLBACK_IP",
    label: "Callback IP",
    value: process.env.PULSA24JAM_CALLBACK_IP,
    desc: "IP callback provider yang diizinkan masuk ke webhook.",
  },
  {
    key: "PULSA24JAM_CALLBACK_TOKEN",
    label: "Callback Token",
    value: process.env.PULSA24JAM_CALLBACK_TOKEN,
    desc: "Validasi status transaksi yang masuk.",
  },
];

const commandRows = [
  {
    command: "SALDO",
    panel: "Admin / Wallet",
    fungsi: "Cek saldo deposit induk Pulsa24Jam.",
    wajib: "pin",
    catatan: "Tidak perlu product, dest, qty, atau refid.",
  },
  {
    command: "PRODUK",
    panel: "Admin Produk",
    fungsi: "Sinkron referensi produk, harga, dan status aktif.",
    wajib: "pin",
    catatan: "Boleh pakai product untuk filter kode tertentu.",
  },
  {
    command: "PAY",
    panel: "Retail & H2H",
    fungsi: "Kirim transaksi utama setelah order valid.",
    wajib: "product, dest, qty, refid, pin",
    catatan: "Refid wajib unik dari sistem Isiloka.",
  },
  {
    command: "INQ",
    panel: "Retail PPOB & H2H",
    fungsi: "Cek tagihan sebelum pembayaran postpaid.",
    wajib: "product, dest, refid, pin",
    catatan: "Qty otomatis 1 untuk inquiry.",
  },
  {
    command: "STATUS-PAY",
    panel: "Operator / Admin",
    fungsi: "Cek ulang transaksi pending atau belum final.",
    wajib: "product, dest, qty, refid, pin",
    catatan: "Dipakai untuk retry status dan fallback provider.",
  },
];

export default function Pulsa24JamIntegrationPage() {
  const readyCount = envItems.filter((item) => configured(item.value)).length;
  const callbackURL = `${backendBaseURL()}/v1/webhook/pulsa24jam`;
  const trxURL = `${String(process.env.PULSA24JAM_BASE_URL || "https://api.pulsa24jam.net").replace(/\/$/, "")}/v1/trx`;
  const isReady = readyCount >= 4;
  const steps = [
    {
      title: "1. IP Whitelist",
      desc: "Daftarkan IP publik server Isiloka di Pulsa24Jam. Tanpa ini, transaksi H2H akan ditolak.",
      icon: ShieldCheck,
    },
    {
      title: "2. Webhook Callback",
      desc: "Pasang URL callback Isiloka agar status pending, sukses, gagal, SN, dan keterangan final masuk otomatis.",
      icon: Webhook,
    },
    {
      title: "3. API Key",
      desc: "Simpan API key hanya di server. Backend mengirimnya lewat header X-Api-Key, bukan dari browser.",
      icon: KeyRound,
    },
    {
      title: "4. PIN Transaksi",
      desc: "PIN akun H2H dipakai di body request untuk semua command transaksi.",
      icon: ShieldCheck,
    },
    {
      title: "5. Kirim Transaksi",
      desc: "Backend mengirim command ke /v1/trx dengan refid unik, lalu menunggu callback final.",
      icon: MapPinned,
    },
  ];

  return (
    <main className="-m-2 min-h-screen bg-[#f0f8f8] p-3 text-slate-950 sm:p-5 lg:p-7">
      <section className="mx-auto flex w-full max-w-7xl flex-col gap-5">
        <header className="overflow-hidden rounded-[28px] border border-emerald-200 bg-white shadow-[0_24px_60px_rgba(6,78,59,0.10)]">
          <div className="grid gap-0 lg:grid-cols-[minmax(0,1fr)_360px]">
            <div className="bg-[linear-gradient(135deg,#064e3b_0%,#047857_62%,#55c72f_100%)] px-5 py-7 text-white sm:px-7 lg:px-9">
              <p className="inline-flex items-center gap-2 rounded-full border border-white bg-[#13464b] px-3 py-1 text-[11px] font-black uppercase tracking-[0.22em] text-white">
                <Zap className="h-3.5 w-3.5 fill-white text-white" />
                Isiloka Gateway
              </p>
              <h1 className="mt-5 max-w-2xl text-3xl font-black tracking-normal sm:text-4xl">
                Integrasi Provider
              </h1>
              <p className="mt-3 max-w-2xl text-sm font-semibold leading-6 text-emerald-50">
                Halaman ini adalah pusat kontrol alur H2H Pulsa24Jam: whitelist IP, webhook, API key, PIN, command transaksi, callback final, dan monitoring semua panel.
              </p>
              <div className="mt-6 flex flex-wrap gap-2">
                <span className="rounded-full border border-[#13464b] bg-white px-3 py-1.5 text-xs font-black text-[#13464b]">
                  Provider: Pulsa24Jam
                </span>
                <span className="rounded-full border border-white bg-[#13464b] px-3 py-1.5 text-xs font-black text-white">
                  Endpoint: POST /v1/trx
                </span>
              </div>
            </div>

            <aside className="flex flex-col justify-center gap-4 bg-[#f8fffb] p-5 sm:p-7">
              <div className={isReady ? "rounded-[24px] border-2 border-[#13464b] bg-white p-5" : "rounded-[24px] border-2 border-amber-700 bg-white p-5"}>
                <div className="flex items-center gap-3">
                  <span className={isReady ? "grid h-12 w-12 place-items-center rounded-2xl bg-[#13464b] text-white shadow-[0_12px_24px_rgba(0,168,120,0.20)]" : "grid h-12 w-12 place-items-center rounded-2xl bg-white text-amber-800 ring-2 ring-amber-800 shadow-[0_12px_24px_rgba(251,191,36,0.22)]"}>
                    {isReady ? <CheckCircle2 className="h-6 w-6" /> : <Activity className="h-6 w-6" />}
                  </span>
                  <div>
                    <p className="text-[10px] font-black uppercase tracking-[0.16em] text-slate-500">Status Integrasi</p>
                    <p className="mt-1 text-xl font-black">{isReady ? "Siap diuji" : "Belum lengkap"}</p>
                  </div>
                </div>
                <div className="mt-4 h-3 overflow-hidden rounded-full bg-slate-200">
                  <div className="h-full rounded-full bg-[#13464b]" style={{ width: `${(readyCount / envItems.length) * 100}%` }} />
                </div>
                <p className="mt-3 text-xs font-bold text-slate-600">{readyCount}/{envItems.length} konfigurasi server terisi</p>
              </div>
            </aside>
          </div>
        </header>

        <section className="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
          {envItems.map((item) => {
            const ok = configured(item.value);
            const sensitive =
              item.key.includes("API_KEY") ||
              item.key.includes("PASSWORD") ||
              item.key.includes("PIN") ||
              item.key.includes("SECRET") ||
              item.key.includes("TOKEN");
            return (
              <div key={item.key} className="rounded-[22px] border border-emerald-100 bg-white p-4 shadow-[0_14px_34px_rgba(6,78,59,0.07)]">
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <p className="text-[10px] font-black uppercase tracking-[0.16em] text-[#13464b]">{item.label}</p>
                    <p className="mt-2 truncate text-sm font-black text-slate-950">{sensitive || item.key.includes("API_KEY") ? mask(item.value) : item.value || "Belum diisi"}</p>
                  </div>
                  <span className={ok ? "grid h-8 w-8 shrink-0 place-items-center rounded-xl bg-white text-[#13464b] ring-2 ring-[#13464b]" : "grid h-8 w-8 shrink-0 place-items-center rounded-xl bg-white text-rose-700 ring-2 ring-rose-700"}>
                    {ok ? <CheckCircle2 className="h-4 w-4" /> : <XCircle className="h-4 w-4" />}
                  </span>
                </div>
                <p className="mt-3 text-xs font-semibold leading-5 text-slate-500">{item.desc}</p>
              </div>
            );
          })}
        </section>

        <section className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_390px]">
          <div className="rounded-[28px] border border-emerald-100 bg-white p-5 shadow-[0_20px_46px_rgba(6,78,59,0.08)] sm:p-6">
            <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
              <div className="flex items-start gap-3">
                <span className="grid h-12 w-12 shrink-0 place-items-center rounded-2xl bg-[#e8fff4] text-[#008f6b] ring-1 ring-emerald-100">
                  <PlugZap className="h-6 w-6" />
                </span>
                <div className="min-w-0">
                  <p className="text-[11px] font-black uppercase tracking-[0.20em] text-[#13464b]">Koneksi Transaksi</p>
                  <h2 className="mt-1 text-xl font-black text-slate-950">Jalur hit Isiloka ke Pulsa24Jam</h2>
                  <p className="mt-2 text-sm font-semibold leading-6 text-slate-500">
                    Frontend dan panel hanya membuat order atau monitoring. Backend Isiloka yang menyimpan API key, mengirim command, mencatat refid, lalu menerima callback final.
                  </p>
                </div>
              </div>
              <span className="w-fit rounded-full bg-slate-100 px-3 py-1.5 text-xs font-black text-slate-600">
                Backend first
              </span>
            </div>

            <div className="mt-5 rounded-[22px] border border-dashed border-emerald-300 bg-[#f7fffb] p-4">
              <p className="text-[10px] font-black uppercase tracking-[0.16em] text-slate-500">Callback URL</p>
              <p className="mt-2 break-all text-sm font-black text-[#13464b]">{callbackURL}</p>
            </div>
            <div className="mt-3 rounded-[22px] border border-dashed border-emerald-300 bg-white p-4">
              <p className="text-[10px] font-black uppercase tracking-[0.16em] text-slate-500">Transaction URL</p>
              <p className="mt-2 break-all text-sm font-black text-[#13464b]">{trxURL}</p>
              <p className="mt-2 text-xs font-semibold leading-5 text-slate-500">
                Header wajib: X-Api-Key dan Content-Type application/json. Body wajib berisi commands dan pin.
              </p>
            </div>

            <div className="mt-5 grid gap-3 sm:grid-cols-2">
              <Link href="/dashboard/admin/master/produk/provider/map" className="inline-flex min-h-11 items-center justify-center gap-2 rounded-2xl border-2 border-[#13464b] bg-white px-4 py-3 text-sm font-black text-[#13464b] shadow-[0_12px_24px_rgba(6,78,59,0.10)] outline-none transition hover:bg-[#f8fffb] focus-visible:ring-4 focus-visible:ring-emerald-200">
                Buka Mapping
                <ArrowUpRight className="h-4 w-4" />
              </Link>
              <Link href="/dashboard/admin/transaksi/provider" className="inline-flex min-h-11 items-center justify-center gap-2 rounded-2xl border-2 border-[#13464b] bg-white px-4 py-3 text-sm font-black text-[#13464b] outline-none transition hover:bg-slate-50 focus-visible:ring-4 focus-visible:ring-emerald-200">
                Log Provider
                <ExternalLink className="h-4 w-4" />
              </Link>
            </div>
          </div>

          <div className="rounded-[28px] border border-emerald-100 bg-white p-5 shadow-[0_20px_46px_rgba(6,78,59,0.08)] sm:p-6">
            <div className="flex items-center gap-3">
              <span className="grid h-12 w-12 place-items-center rounded-2xl bg-lime-100 text-[#3a8f00]">
                <ClipboardList className="h-6 w-6" />
              </span>
              <div>
                <p className="text-[11px] font-black uppercase tracking-[0.20em] text-[#13464b]">Setup Live</p>
                <h2 className="mt-1 text-xl font-black text-slate-950">Checklist</h2>
              </div>
            </div>
            <div className="mt-5 space-y-3">
              {steps.map((step, index) => {
                const Icon = step.icon;
                return (
                  <div key={step.title} className="flex gap-3 rounded-[20px] border border-slate-100 bg-[#fbfffd] p-3">
                    <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-[#13464b] text-white">
                      <Icon className="h-4 w-4" />
                    </span>
                    <div>
                      <p className="text-sm font-black text-slate-950">{step.title}</p>
                      <p className="mt-1 text-xs font-semibold leading-5 text-slate-500">{step.desc}</p>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </section>

        <section className="rounded-[28px] border border-emerald-100 bg-white p-5 shadow-[0_20px_46px_rgba(6,78,59,0.08)] sm:p-6">
          <div className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
            <div>
              <p className="text-[11px] font-black uppercase tracking-[0.20em] text-[#13464b]">Command H2H</p>
              <h2 className="mt-1 text-xl font-black text-slate-950">Perintah yang Dipakai Semua Panel</h2>
              <p className="mt-2 text-sm font-semibold leading-6 text-slate-500">
                Semua command dikirim dari backend ke Pulsa24Jam. Panel hanya mengatur data, menampilkan status, dan menjalankan monitoring.
              </p>
            </div>
            <span className="w-fit rounded-full bg-[#13464b] px-3 py-1.5 text-xs font-black text-white">
              Jangan taruh API key di frontend
            </span>
          </div>

          <div className="mt-5 overflow-hidden rounded-[22px] border border-slate-200">
            <div className="hidden grid-cols-[120px_160px_minmax(0,1fr)_220px_minmax(0,1fr)] gap-3 bg-[#13464b] px-4 py-3 text-[10px] font-black uppercase tracking-[0.12em] text-white md:grid">
              <span>Command</span>
              <span>Panel</span>
              <span>Fungsi</span>
              <span>Parameter Wajib</span>
              <span>Catatan</span>
            </div>
            <div className="divide-y divide-slate-100">
              {commandRows.map((row) => (
                <div key={row.command} className="grid gap-3 px-4 py-4 text-sm md:grid-cols-[120px_160px_minmax(0,1fr)_220px_minmax(0,1fr)] md:items-center">
                  <div>
                    <span className="inline-flex rounded-xl bg-emerald-50 px-3 py-1 text-xs font-black text-[#087e8b] ring-1 ring-emerald-100">
                      {row.command}
                    </span>
                  </div>
                  <p className="font-black text-slate-950">{row.panel}</p>
                  <p className="font-semibold leading-6 text-slate-600">{row.fungsi}</p>
                  <p className="break-words font-mono text-xs font-bold text-slate-800">{row.wajib}</p>
                  <p className="font-semibold leading-6 text-slate-500">{row.catatan}</p>
                </div>
              ))}
            </div>
          </div>
        </section>

        <section className="rounded-[26px] border border-amber-200 bg-[#fff9e8] p-5 shadow-[0_14px_32px_rgba(180,83,9,0.06)]">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-start">
            <span className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-amber-400 text-slate-950">
              <Link2 className="h-5 w-5" />
            </span>
            <div>
              <h2 className="text-base font-black text-slate-950">Catatan sebelum live</h2>
              <p className="mt-2 text-sm font-semibold leading-6 text-slate-600">
                Alur live sekarang harus berurutan: ambil produk, pastikan IP terdaftar, pastikan webhook aktif, kirim command dengan API key dan PIN, simpan refid, lalu biarkan callback menjadi sumber status final.
              </p>
            </div>
          </div>
        </section>
      </section>
    </main>
  );
}
