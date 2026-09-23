import Link from "next/link";

const cards = [
  { href: "/dashboard/auditor/transaksi", title: "Transaksi Harian", desc: "Ringkasan penjualan harian per scope dan margin." },
  { href: "/dashboard/auditor/transaksi-jual-beli", title: "Transaksi Jual Beli", desc: "Detail ref id, harga beli, harga jual, komisi, dan margin." },
  { href: "/dashboard/auditor/transaksi-keuangan", title: "Transaksi Keuangan", desc: "Arus kas bank: deposit, topup provider, withdraw, dan pengeluaran internal." },
  { href: "/dashboard/auditor/rugi-laba", title: "Rugi Laba", desc: "Laporan rugi laba dan neraca ringkas per bulan." },
];

export default function AuditorDashboardPage() {
  return (
    <div className="space-y-4 p-2">
      <div>
        <div className="text-lg font-semibold tracking-tight">Dashboard Auditor</div>
        <div className="text-sm text-muted-foreground">Pusat laporan audit transaksi, keuangan, dan rugi laba.</div>
      </div>
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        {cards.map((card) => (
          <Link key={card.href} href={card.href} className="rounded-md border border-white/15 bg-slate-950/50 p-4 transition hover:border-sky-400/40">
            <div className="text-base font-semibold text-slate-100">{card.title}</div>
            <div className="mt-2 text-sm text-slate-400">{card.desc}</div>
          </Link>
        ))}
      </div>
    </div>
  );
}
