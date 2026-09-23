import { redirect } from "next/navigation";

export default function LegacyProviderTrxPage() {
  redirect("/dashboard/admin/history/transaksi-provider");
}
