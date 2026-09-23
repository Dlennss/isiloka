import { AgentCreditRiskPage } from "@/components/dashboard/credit-risk/AgentCreditRiskPage";
import { attachAgentCreditPaymentsFallback, getAgentCreditApplicationsDatabaseFallback } from "@/lib/agent-credit-payment-fallback.server";
import { getAgentCreditApplications } from "@/lib/api.auth";
import { getAppServerSession } from "@/lib/server-auth";

type SessionShape = {
  backendToken?: string;
};

export default async function MarketingAgentProfilePage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  const backendApplications = session?.backendToken ? await getAgentCreditApplications(session.backendToken) : [];
  const rawApplications = backendApplications.length ? backendApplications : await getAgentCreditApplicationsDatabaseFallback();
  const applications = await attachAgentCreditPaymentsFallback(rawApplications);
  return <AgentCreditRiskPage applications={applications} mode="marketing" />;
}
