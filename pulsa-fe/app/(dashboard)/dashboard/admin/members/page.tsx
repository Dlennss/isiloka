import { redirect } from "next/navigation";

export default function LegacyAdminMembersPage() {
  redirect("/dashboard/admin/master/members");
}
