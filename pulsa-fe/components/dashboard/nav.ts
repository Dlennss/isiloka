export type NavLinkItem = {
  href: string;
  label: string;
};

export type NavSection = {
  title?: string;
  items: NavLinkItem[];
};

export type H2HRole = "member" | "agent_member" | "master_member";

export function getMemberNavSections(role: H2HRole): NavSection[] {
  const baseItems: NavLinkItem[] = [
    { href: "/dashboard/member", label: "Dashboard" },
    { href: "/dashboard/member/produk", label: "Produk H2H" },
    { href: "/dashboard/member/deposit", label: "Request Deposit" },
    { href: "/dashboard/member/security", label: "Pengaturan" },
  ];

  if (role === "agent_member" || role === "master_member") {
    baseItems.splice(1, 0, { href: "/dashboard/member/downline", label: "Jaringan H2H" });
    baseItems.splice(2, 0, { href: "/dashboard/member/fee", label: "Fee H2H" });
    baseItems.splice(3, 0, { href: "/dashboard/member/withdraw", label: "Withdraw H2H" });
  }

  return [
    {
      items: baseItems,
    },
    {
      title: "History",
      items: [
        { href: "/dashboard/member/history/mutasi", label: "Mutasi" },
        { href: "/dashboard/member/history/transaksi", label: "Transaksi" },
        { href: "/dashboard/member/history/deposit", label: "Deposit" },
      ],
    },
  ];
}

export const operatorNavSections: NavSection[] = [
  {
    items: [
      { href: "/dashboard/operator", label: "Dashboard" },
      { href: "/dashboard/operator/deposits", label: "Deposit Request" },
      { href: "/dashboard/operator/bank", label: "Bank" },
      { href: "/dashboard/operator/master/produk/provider/map", label: "Toggle Mapping Provider" },
    ],
  },
  {
    title: "Transaksi",
    items: [
      { href: "/dashboard/operator/transaksi/aplikasi", label: "Transaksi Retail" },
      { href: "/dashboard/operator/transaksi/provider", label: "Transaksi Provider" },
      { href: "/dashboard/operator/transaksi/member", label: "Transaksi Member" },
      { href: "/dashboard/operator/transaksi/deposit-va", label: "Deposit VA" },
      { href: "/dashboard/operator/transaksi/guest-refund", label: "Guest Refund Pending" },
    ],
  },
  {
    title: "Audit",
    items: [
      { href: "/dashboard/operator/transaksi/member-status-logs", label: "Audit Log Status Saya" },
      { href: "/dashboard/operator/transaksi/member-status-logs-manual", label: "Audit Transaksi Diubah" },
      { href: "/dashboard/operator/transaksi/status-mismatch", label: "Audit Status Mismatch" },
      { href: "/dashboard/operator/transaksi/provider-empty-response", label: "Audit Indikasi Timeout" },
      { href: "/dashboard/operator/transaksi/provider-wallet-missing-debit", label: "Audit Potong Saldo Provider" },
      { href: "/dashboard/operator/transaksi/transaksi-suspect", label: "Transaksi Provider Suspect" },
    ],
  },
];

export const walletNavSections: NavSection[] = [
  {
    items: [
      { href: "/dashboard/wallet", label: "Dashboard" },
      { href: "/dashboard/wallet/deposits", label: "Deposit Request" },
      { href: "/dashboard/wallet/deposit-va", label: "Deposit VA" },
      { href: "/dashboard/wallet/history/members", label: "Riwayat Deposit Member" },
      { href: "/dashboard/wallet/mutasi-bank", label: "Mutasi Bank" },
    ],
  },
  {
    title: "Wallet",
    items: [
      { href: "/dashboard/wallet/bank", label: "Bank" },
      { href: "/dashboard/wallet/members", label: "Saldo Member" },
      { href: "/dashboard/wallet/provider-wallets", label: "Saldo Provider" },
      { href: "/dashboard/wallet/provider-accounts", label: "Rekening Provider" },
      { href: "/dashboard/wallet/qrtp-transfer", label: "Transfer QRTP ke Provider" },
      { href: "/dashboard/wallet/loketbayar-transfer", label: "Transfer LoketBayar ke Provider" },
      { href: "/dashboard/wallet/retail-withdraws", label: "Retail Withdraw" },
      { href: "/dashboard/wallet/h2h-withdraws", label: "H2H Withdraw" },
      { href: "/dashboard/wallet/history/providers", label: "Riwayat Deposit Provider" },
      { href: "/dashboard/wallet/internal-finance", label: "Pengeluaran Internal" },
    ],
  },
  {
    title: "Transaksi",
    items: [
      { href: "/dashboard/wallet/transaksi/provider", label: "Transaksi Provider" },
      { href: "/dashboard/wallet/transaksi/member", label: "Transaksi Member" },
      { href: "/dashboard/wallet/transaksi/guest-refund", label: "Guest Refund Pending" },
    ],
  },
  {
    title: "Audit",
    items: [
      { href: "/dashboard/wallet/transaksi/member-status-logs", label: "Audit Log Status Member" },
      { href: "/dashboard/wallet/transaksi/member-status-logs-manual", label: "Audit Transaksi Diubah" },
      { href: "/dashboard/wallet/transaksi/status-mismatch", label: "Audit Status Mismatch" },
      { href: "/dashboard/wallet/transaksi/provider-empty-response", label: "Audit Indikasi Timeout" },
      { href: "/dashboard/wallet/transaksi/provider-wallet-missing-debit", label: "Audit Potong Saldo Provider" },
      { href: "/dashboard/wallet/transaksi/transaksi-suspect", label: "Transaksi Provider Suspect" },
    ],
  },
];

export const adminNavSections: NavSection[] = [
  {
    title: "Ringkasan",
    items: [
      { href: "/dashboard/admin", label: "Dashboard" },
      { href: "/dashboard/admin/komisi", label: "Laporan Bisnis" },
    ],
  },
  {
    title: "Pengguna & Kredit",
    items: [
      { href: "/dashboard/admin/master/members", label: "Akun Pengguna" },
      { href: "/dashboard/admin/pemantauan-tim", label: "Pemantauan Tim" },
      { href: "/dashboard/admin/kredit/pengajuan", label: "Pengajuan Kredit" },
    ],
  },
  {
    title: "Transaksi",
    items: [
      { href: "/dashboard/admin/transaksi/aplikasi", label: "Transaksi Pelanggan" },
      { href: "/dashboard/admin/transaksi/aplikasi/provider", label: "Proses Pulsa24Jam" },
      { href: "/dashboard/admin/transaksi/guest-refund", label: "Refund Pelanggan" },
    ],
  },
  {
    title: "Keuangan",
    items: [
      { href: "/dashboard/admin/deposits", label: "Top Up Pengguna" },
      { href: "/dashboard/admin/retail-withdraws", label: "Penarikan Agent" },
      { href: "/dashboard/admin/bank", label: "Rekening Bank" },
    ],
  },
  {
    title: "Produk & Aplikasi",
    items: [
      { href: "/dashboard/admin/master/produk", label: "Produk & Harga" },
      { href: "/dashboard/admin/master/fee-kategori-aplikasi", label: "Biaya Layanan" },
      { href: "/dashboard/admin/master/iklan", label: "Banner Aplikasi" },
      { href: "/dashboard/admin/integrasi/pulsa24jam", label: "Koneksi Pulsa24Jam" },
    ],
  },
];

const staffHiddenAdminHrefs = new Set([
  "/dashboard/admin/master/members",
  "/dashboard/admin/pemantauan-tim",
  "/dashboard/admin/master/aktivasi-h2h",
  "/dashboard/admin/members",
  "/dashboard/admin/provider-merchant-ids",
]);

export const staffNavSections: NavSection[] = adminNavSections
  .map((section) => ({
    ...section,
    items: section.items.filter((item) => !staffHiddenAdminHrefs.has(item.href)),
  }))
  .filter((section) => section.items.length > 0);

export const auditorNavSections: NavSection[] = [
  {
    title: "Ringkasan",
    items: [
      { href: "/dashboard/auditor", label: "Dashboard Auditor" },
      { href: "/dashboard/auditor/rugi-laba", label: "Rugi Laba" },
    ],
  },
  {
    title: "Transaksi",
    items: [
      { href: "/dashboard/auditor/transaksi", label: "Transaksi Harian" },
      { href: "/dashboard/auditor/transaksi-jual-beli", label: "Transaksi Jual Beli" },
      { href: "/dashboard/auditor/transaksi-keuangan", label: "Transaksi Keuangan" },
    ],
  },
];

export const masterNavSections: NavSection[] = [
  {
    items: [
      { href: "/dashboard/master", label: "Dashboard" },
    ],
  },
  {
    title: "Operasional Lapangan",
    items: [
      { href: "/dashboard/master/tambah-agent", label: "Tambah Agent" },
      { href: "/dashboard/master/pinjaman", label: "Antrean Survei" },
    ],
  },
  {
    title: "Portofolio Agent",
    items: [
      { href: "/dashboard/master/akun-agent", label: "Agent Binaan" },
      { href: "/dashboard/master/profil-agent", label: "Aktivitas Agent" },
      { href: "/dashboard/master/riwayat-pinjaman", label: "Penagihan Kredit" },
    ],
  },
  {
    title: "Administrasi",
    items: [
      { href: "/dashboard/master/laporan", label: "Laporan" },
    ],
  },
];

export const analystNavSections: NavSection[] = [
  {
    items: [
      { href: "/dashboard/master/operator", label: "Pengajuan Kredit" },
    ],
  },
  {
    title: "Kelola Agent",
    items: [
      { href: "/dashboard/master/operator/tambah-marketing", label: "Akun Marketing" },
      { href: "/dashboard/master/operator/input-data-agent", label: "Migrasi Data Agent" },
      { href: "/dashboard/master/operator/kenaikan-limit", label: "Limit Agent" },
    ],
  },
  {
    title: "Pemantauan",
    items: [
      { href: "/dashboard/master/operator/pembayaran-kredit", label: "Pembayaran Kredit" },
      { href: "/dashboard/master/operator/transaksi-agent", label: "Aktivitas Agent" },
      { href: "/dashboard/master/operator/konter-tidak-transaksi", label: "Agent Tidak Aktif" },
      { href: "/dashboard/master/operator/arsip-keputusan", label: "Riwayat Keputusan" },
    ],
  },
];
