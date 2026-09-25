import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { getAppServerSession } from "@/lib/server-auth";
import type { UserSession } from "@/components/user/types";
import { AppTopHeader } from "@/components/shared/AppTopHeader";
import { SiteFrame } from "@/components/site/SiteFrame";
import { SiteShell } from "@/components/site/SiteShell";

export const metadata: Metadata = {
  title: "User Area - Isiloka",
  description: "Aplikasi user untuk pembelian produk digital langsung.",
};

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

export default async function UserLayout({ children }: { children: React.ReactNode }) {
  const session = (await getAppServerSession()) as SessionShape | null;
  const role = String(session?.user?.role || "").trim().toLowerCase();
  const isRetailRole = role === "user" || role === "agent" || role === "master" || role === "marketing";

  if (!session?.backendToken) {
    redirect("/login");
  }

  if (session?.backendToken && role && !isRetailRole) {
    if (role === "admin" || role === "staff") redirect("/dashboard/admin");
    if (role === "member" || role === "agent_member" || role === "master_member") redirect("/dashboard/member");
    if (role === "operator_trx") redirect("/dashboard/operator");
    if (role === "operator_wallet") redirect("/dashboard/wallet");
    redirect("/dashboard");
  }

  return (
    <SiteFrame>
      <div className="site-frame-shell relative mx-auto min-h-dvh w-full max-w-md overflow-hidden bg-[#f7fffc] shadow-[0_26px_90px_rgba(23,89,86,0.16)] md:min-h-[calc(100dvh-2rem)] md:w-97.5 md:max-w-none md:rounded-[42px] md:border md:border-white/80">
        <AppTopHeader
          isLoggedIn={Boolean(session?.backendToken)}
          role={role}
        />
        <SiteShell>{children}</SiteShell>
      </div>
    </SiteFrame>
  );
}
