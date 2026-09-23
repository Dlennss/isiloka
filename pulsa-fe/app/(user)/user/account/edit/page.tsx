import { getAppServerSession } from "@/lib/server-auth";
import { redirect } from "next/navigation";
import { UserAccountEditForm } from "@/components/user/UserAccountEditForm";
import { getUserProfile } from "@/lib/api.auth";
import type { UserSession } from "@/components/user/types";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

export default async function UserAccountEditPage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  if (!session?.backendToken) {
    redirect("/login");
  }

  const profile = await getUserProfile(session.backendToken);
  const nama = profile?.nama || session.user?.name || "";
  const email = profile?.email || session.user?.email || "";
  const phone = profile?.phone || "";
  const profilePhotoURL = profile?.profile_photo_url || session.user?.image || "";

  return <UserAccountEditForm nama={nama} email={email} phone={phone} profilePhotoURL={profilePhotoURL} />;
}
