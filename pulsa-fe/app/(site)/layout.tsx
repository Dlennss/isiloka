import type { Metadata } from "next";
import { getServerSession } from "next-auth";
import { SiteShell } from "@/components/site/SiteShell";
import { authOptions } from "@/lib/nextauth";
import type { UserSession } from "@/components/user/types";
import { AppTopHeader } from "@/components/shared/AppTopHeader";

export const metadata: Metadata = {
  title: "Isiloka",
  description: "Topup & PPOB cepat",
};

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

export default async function SiteLayout({ children }: { children: React.ReactNode }) {
  const session = (await getServerSession(authOptions)) as SessionShape | null;

  return (
    <div className="min-h-dvh bg-[radial-gradient(circle_at_50%_10%,#effffc_0%,#e6f8f5_38%,#f7fffc_100%)] text-neutral-900 md:grid md:place-items-start md:py-4">
      <div className="relative mx-auto min-h-dvh w-full max-w-md overflow-hidden bg-[#f7fffc] shadow-[0_26px_90px_rgba(23,89,86,0.16)] md:min-h-[calc(100dvh-2rem)] md:w-97.5 md:max-w-none md:rounded-[42px] md:border md:border-white/80">
        <AppTopHeader isLoggedIn={Boolean(session?.backendToken)} />
        <SiteShell>{children}</SiteShell>
      </div>
    </div>
  );
}
