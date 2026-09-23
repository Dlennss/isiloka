import Link from "next/link";
import { ChevronLeft, LockKeyhole, PhoneCall, ShieldCheck } from "lucide-react";
import { GuestBottomNav } from "@/components/guest/GuestBottomNav";

const principles = [
  {
    title: "Layanan Isiloka",
    items: [
      "Isiloka menyediakan pulsa, paket data, token listrik, e-wallet, voucher game, PPOB, dan produk digital lain yang tersedia di aplikasi.",
      "Harga, ketersediaan produk, estimasi proses, dan status transaksi dapat berubah mengikuti kondisi mitra layanan dan sistem pembayaran.",
      "Setiap transaksi harus dilakukan menggunakan nomor tujuan dan data yang benar. Kesalahan input menjadi tanggung jawab pengguna.",
    ],
  },
  {
    title: "Akun dan Saldo",
    items: [
      "Satu akun hanya boleh digunakan oleh pemiliknya. Pengguna wajib menjaga password, PIN, kode verifikasi, dan perangkat login.",
      "Saldo utama digunakan untuk pembelian produk dan transaksi yang tersedia. Saldo yang sudah terpotong hanya dapat dikembalikan sesuai hasil pemeriksaan transaksi.",
      "Top up dianggap berhasil setelah pembayaran terverifikasi dan saldo masuk ke akun. Bukti pembayaran wajib disimpan sampai transaksi selesai.",
    ],
  },
  {
    title: "Program Kredit Agent",
    items: [
      "Fasilitas kredit hanya tersedia untuk agent yang mengajukan dan memenuhi pemeriksaan sistem serta dokumen yang ditentukan.",
      "Operator Kredit memiliki keputusan akhir untuk menyetujui, menolak, menonaktifkan, atau mengubah limit kredit agent.",
      "Kredit yang disetujui masuk sebagai saldo utama dan hanya dapat digunakan sesuai aturan aplikasi. Saat agent berhenti menggunakan layanan, seluruh kewajiban kredit harus diselesaikan.",
      "Agent yang tidak bertransaksi selama beberapa hari dapat masuk pemantauan operator untuk ditindaklanjuti bersama marketing.",
    ],
  },
  {
    title: "Peran Marketing dan Operator",
    items: [
      "Marketing membantu memantau agent binaan, melihat dokumen dan aktivitas yang diizinkan, serta menyampaikan informasi kepada operator.",
      "Marketing tidak dapat menyetujui kredit, menolak kredit, atau menaikkan limit agent.",
      "Operator dan admin dapat melakukan pemantauan, pemeriksaan, pembatasan akun, dan penanganan masalah demi keamanan layanan.",
    ],
  },
  {
    title: "Penggunaan yang Dilarang",
    items: [
      "Dilarang menggunakan Isiloka untuk penipuan, pencucian uang, penyalahgunaan identitas, manipulasi transaksi, atau aktivitas yang melanggar hukum.",
      "Isiloka dapat menunda transaksi, membatasi fitur, atau menonaktifkan akun jika ditemukan aktivitas mencurigakan atau pelanggaran ketentuan.",
    ],
  },
];

type PrivacyPolicyPageContentProps = {
  backHref?: string;
};

export function PrivacyPolicyPageContent({ backHref = "/" }: PrivacyPolicyPageContentProps) {
  return (
    <main className="relative overflow-hidden bg-sky-50 text-slate-900">
      <div className="pointer-events-none absolute inset-x-0 top-0 h-72 bg-[radial-gradient(620px_260px_at_10%_0%,rgba(15,111,203,0.18),transparent_62%),radial-gradient(760px_320px_at_100%_0%,rgba(45,212,191,0.14),transparent_56%),linear-gradient(180deg,#e9f5ff_0%,rgba(233,245,255,0.36)_38%,rgba(233,245,255,0)_100%)]" />

      <section className="relative space-y-5 px-4 pb-5 pt-4">
        <Link
          href={backHref}
          className="inline-flex items-center gap-1 rounded-full border border-sky-200 bg-white/85 px-3 py-1.5 text-sm font-semibold text-sky-700 shadow-[0_10px_24px_rgba(15,111,203,0.08)]"
          aria-label="Kembali"
        >
          <ChevronLeft className="h-4 w-4" />
          <span>Kembali</span>
        </Link>

        <div className="overflow-hidden rounded-[28px] border border-white/70 bg-linear-to-br from-[#087e8b] via-[#087e8b] to-[#087e8b] px-5 py-6 text-white shadow-[0_24px_60px_rgba(15,111,203,0.22)]">
          <div className="max-w-[20rem]">
            <p className="text-[11px] font-semibold uppercase tracking-[0.22em] text-white/72">Syarat & Ketentuan</p>
            <h1 className="mt-3 text-balance text-[30px] font-black leading-[1.02] tracking-tight">
              Aturan penggunaan layanan Isiloka.
            </h1>
            <p className="mt-4 text-sm leading-6 text-white/90 text-justify">
              Ketentuan ini menjelaskan cara menggunakan produk digital, saldo, akun agent, program kredit, dan fitur pendampingan marketing di Isiloka.
            </p>
            <p className="mt-4 text-xs font-semibold uppercase tracking-[0.16em] text-white/72">Berlaku sejak 6 April 2026</p>
          </div>
        </div>
      </section>

      <section className="relative px-4 pb-4">
        <div className="rounded-[28px] border border-sky-100 bg-white p-5 shadow-[0_18px_40px_rgba(15,23,42,0.08)]">
          <div className="flex items-start gap-3">
            <div className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-sky-100 text-sky-700">
              <ShieldCheck className="h-5 w-5" />
            </div>
            <div>
              <p className="text-[10px] font-semibold uppercase tracking-[0.18em] text-sky-700">Ringkasan</p>
              <h2 className="mt-1 text-xl font-black tracking-tight text-slate-950">
                Isiloka menjaga transaksi tetap aman, jelas, dan bertanggung jawab
              </h2>
            </div>
          </div>

          <div className="mt-5 grid gap-4">
            {principles.map((section) => (
              <div key={section.title} className="rounded-2xl border border-slate-200 bg-slate-50/80 p-4">
                <h3 className="text-sm font-black tracking-tight text-slate-950">{section.title}</h3>
                <ul className="mt-3 space-y-2 text-sm leading-6 text-slate-600">
                  {section.items.map((item) => (
                    <li key={item} className="flex gap-2">
                      <span className="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-sky-500" />
                      <span>{item}</span>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="relative px-4 pb-4">
        <div className="rounded-[28px] border border-slate-200 bg-white p-5 shadow-[0_18px_40px_rgba(15,23,42,0.08)]">
          <div className="flex items-start gap-3">
            <div className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-sky-100 text-sky-700">
              <LockKeyhole className="h-5 w-5" />
            </div>
            <div>
            <p className="text-[10px] font-semibold uppercase tracking-[0.18em] text-sky-700">Persetujuan & Perubahan</p>
            <h2 className="mt-1 text-xl font-black tracking-tight text-slate-950">
                Dengan memakai layanan Isiloka, Anda menyetujui ketentuan penggunaan ini
              </h2>
            </div>
          </div>

          <div className="mt-4 space-y-3 text-sm leading-6 text-slate-600 text-justify">
            <p>
              Isiloka dapat memperbarui ketentuan ini mengikuti perubahan fitur, proses layanan, dan kebutuhan keamanan.
              Versi terbaru akan dipublikasikan pada halaman ini.
            </p>
            <p>
              Pengguna dianggap menyetujui perubahan ketentuan apabila tetap menggunakan akun dan layanan Isiloka setelah pembaruan dipublikasikan.
            </p>
          </div>
        </div>
      </section>

      <section className="relative px-4 pb-10">
        <aside className="rounded-[28px] border border-sky-200 bg-linear-to-r from-[#087e8b] via-[#087e8b] to-[#087e8b] p-5 text-white shadow-[0_24px_60px_rgba(15,111,203,0.22)]">
          <div className="flex items-start gap-3">
            <div className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-white/12">
              <PhoneCall className="h-5 w-5" />
            </div>
            <div>
              <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-white/70">Kontak Resmi</p>
              <h2 className="mt-2 text-xl font-black tracking-tight">
                Hubungi Isiloka jika Anda membutuhkan bantuan terkait akun, transaksi, atau ketentuan layanan
              </h2>
            </div>
          </div>

          <div className="mt-6 grid gap-4 text-sm leading-7 text-white/92">
            <div className="rounded-2xl border border-white/16 bg-white/10 px-4 py-3">
              <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-white/64">Badan Usaha</p>
              <p className="mt-2 font-semibold">PT Pulsa Mitra Nasional</p>
            </div>
            <div className="rounded-2xl border border-white/16 bg-white/10 px-4 py-3">
              <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-white/64">WhatsApp & Telepon</p>
              <Link href="tel:+6282219107558" className="mt-2 inline-block text-base font-bold text-cyan-100 hover:text-white">
                0822-1910-7558
              </Link>
            </div>
          </div>

          <div className="mt-6 flex gap-3">
            <Link
              href="https://wa.me/6282219107558"
              className="inline-flex h-12 flex-1 items-center justify-center rounded-full bg-white px-4 text-center text-sm font-semibold text-[#087e8b]! visited:text-[#087e8b]! hover:bg-sky-50 hover:text-[#087e8b]!"
            >
              WhatsApp
            </Link>
            <Link
              href="/tentang"
              className="inline-flex h-12 flex-1 items-center justify-center rounded-full border border-white/30 bg-white/10 px-4 text-center text-sm font-semibold text-white transition hover:bg-white/15"
            >
              Tentang Kami
            </Link>
          </div>
        </aside>
      </section>

      <GuestBottomNav />
    </main>
  );
}
