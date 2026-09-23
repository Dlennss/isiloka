import { redirect } from "next/navigation";

export default async function UserSaldoTagihanPage() {
  // There is no recurring payment flow for an active agent partnership.
  redirect("/user/saldo/kredit-agent");
}
