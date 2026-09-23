import { getAppServerSession } from "@/lib/server-auth";
import type { UserSession } from "@/components/user/types";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import { UserAuthClientSync } from "@/components/user/UserAuthClientSync";
import { ServiceDirectory } from "@/components/shared/ServiceDirectory";
import { UserUniversalServicePageContent } from "@/components/user/UserUniversalServicePageContent";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

type PageProps = {
  searchParams?: Promise<{ layanan?: string }>;
};

export default async function UserAllCategoryPage({ searchParams }: PageProps) {
  const session = (await getAppServerSession()) as SessionShape | null;
  const backendToken = session?.backendToken;
  const role = String(session?.user?.role || "").trim().toLowerCase();
  const resolvedSearchParams = searchParams ? await searchParams : undefined;
  const serviceSlug = String(resolvedSearchParams?.layanan || "").trim().toLowerCase();

  if (serviceSlug) {
    return (
      <>
        {backendToken ? <UserAuthClientSync backendToken={backendToken} /> : null}
        <UserUniversalServicePageContent serviceSlug={serviceSlug} />
        <UserBottomNav />
      </>
    );
  }

  return (
    <main className="bg-sky-50">
      {backendToken ? <UserAuthClientSync backendToken={backendToken} /> : null}

      <div className="space-y-4 px-4 pt-4">
        <ServiceDirectory mode="user" role={role} />
      </div>

      <UserBottomNav />
    </main>
  );
}
