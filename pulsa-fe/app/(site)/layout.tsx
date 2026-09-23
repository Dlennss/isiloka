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
    <div className="min-h-dvh bg-[#f0f8f8] text-neutral-900 md:grid md:place-items-start md:py-4">
      <div className="relative mx-auto w-full max-w-md md:w-97.5 md:max-w-none md:border md:border-[#13464b]/10 md:bg-[#f0f8f8] md:shadow-[0_24px_80px_rgba(6,78,59,0.18)]">
        <AppTopHeader isLoggedIn={Boolean(session?.backendToken)} />
        <SiteShell>{children}</SiteShell>
      </div>
    </div>
  );
}
