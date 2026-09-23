import { MarketingReportCenter } from "@/components/dashboard/MarketingReportCenter";
import { attachAgentCreditPaymentsFallback, getAgentCreditApplicationsDatabaseFallback } from "@/lib/agent-credit-payment-fallback.server";
import { getAgentCreditApplications } from "@/lib/api.auth";
import { getAppServerSession } from "@/lib/server-auth";

type SessionShape = {
  backendToken?: string;
};

export default async function MarketingReportPage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  const backendApplications = session?.backendToken ? await getAgentCreditApplications(session.backendToken) : [];
  const rawApplications = backendApplications.length ? backendApplications : await getAgentCreditApplicationsDatabaseFallback();
  const applications = await attachAgentCreditPaymentsFallback(rawApplications);

  return (
    <main className="-m-2 min-h-screen bg-[#eef7f2] p-3 text-slate-950 sm:p-5 lg:p-7">
      <MarketingReportCenter applications={applications} />
    </main>
  );
}
