import WalletActivityPage from "@/components/dashboard/WalletActivityPage";
import DashboardProfileCard from "@/components/dashboard/DashboardProfileCard";

export default function WalletDashboardPage() {
  return (
    <div className="space-y-4 p-2">
      <DashboardProfileCard
        role="operator_wallet"
        description="Akun operator wallet aktif untuk koreksi saldo member/provider, melihat mutasi, dan audit aktivitas wallet."
      />
      <WalletActivityPage
        title="Dashboard Operator"
        description="Lihat aktivitas koreksi saldo member dan provider berdasarkan actor admin/operator wallet."
        showHeader={false}
      />
    </div>
  );
}
