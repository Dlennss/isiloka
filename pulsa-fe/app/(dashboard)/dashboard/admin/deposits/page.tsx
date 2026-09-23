import DepositRequestsPage from "@/components/dashboard/DepositRequestsPage";

export const dynamic = "force-dynamic";
export const revalidate = 0;

export default function AdminDepositsPage() {
  return (
    <DepositRequestsPage
      title="Deposit Requests"
      description="List pending untuk approve / reject."
    />
  );
}
