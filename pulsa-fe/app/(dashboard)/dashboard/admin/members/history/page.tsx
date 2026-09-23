import { redirect } from "next/navigation";

type Props = {
  searchParams?: {
    member_id?: string;
    tab?: string;
  };
};

export default function LegacyAdminMembersHistoryPage({ searchParams }: Props) {
  const qs = new URLSearchParams();
  if (searchParams?.member_id) qs.set("member_id", searchParams.member_id);
  if (searchParams?.tab) qs.set("tab", searchParams.tab);

  const suffix = qs.toString();
  redirect(`/dashboard/admin/master/members/history${suffix ? `?${suffix}` : ""}`);
}
