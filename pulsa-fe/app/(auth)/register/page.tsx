import { getServerSession } from "next-auth";
import { redirect } from "next/navigation";
import { authOptions } from "@/lib/nextauth";
import { isJwtValid } from "@/lib/jwt";
import { RegisterCard } from "@/components/auth/RegisterCard";
import { BackgroundAuth } from "@/components/auth/BackgroundAuth";

type SessionShape = { backendToken?: string; user?: { role?: string } };

export default async function RegisterPage() {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const isValid = isJwtValid(session?.backendToken);
  if (session?.backendToken && isValid) {
    const role = (session.user?.role || "").toLowerCase();
    if (role === "admin" || role === "staff") redirect("/dashboard/admin");
    if (role === "member" || role === "agent_member" || role === "master_member") redirect("/dashboard/member");
    if (role === "analis" || role === "analyst") redirect("/dashboard/master/operator");
    if (role === "master" || role === "marketing") redirect("/dashboard/master");
    if (role === "operator_trx") redirect("/dashboard/operator");
    if (role === "operator_wallet") redirect("/dashboard/wallet");
    if (role === "user" || role === "agent") redirect("/user");
    redirect("/user");
  }

  return (
    <BackgroundAuth>
      <RegisterCard />
    </BackgroundAuth>
  );
}
