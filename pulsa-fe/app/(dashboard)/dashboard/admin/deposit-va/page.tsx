import DepositVARequestsPage from "@/components/dashboard/DepositVARequestsPage";

export const dynamic = "force-dynamic";
export const revalidate = 0;

export default function AdminDepositVAPage() {
  return <DepositVARequestsPage subtitle="Tiket VA LoketBayar untuk admin." />;
}
