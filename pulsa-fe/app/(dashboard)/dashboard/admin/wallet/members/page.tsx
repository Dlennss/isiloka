import { redirect } from "next/navigation";

export default function AdminWalletMembersRedirect() {
  redirect("/dashboard/admin/master/members?scope=retail");
}
