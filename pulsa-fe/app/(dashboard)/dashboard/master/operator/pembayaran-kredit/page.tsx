import { OperatorCreditPaymentsPage } from "@/components/dashboard/OperatorCreditPaymentsPage";
import { getAgentCreditApplications } from "@/lib/api.auth";
import { attachAgentCreditPaymentsFallback, getAgentCreditApplicationsDatabaseFallback } from "@/lib/agent-credit-payment-fallback.server";
import { getAppServerSession } from "@/lib/server-auth";

type SessionShape = { backendToken?: string };

export default async function OperatorCreditPaymentPage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  const backendItems = session?.backendToken ? await getAgentCreditApplications(session.backendToken, 200) : [];
  const fallbackItems = backendItems.length ? backendItems : await getAgentCreditApplicationsDatabaseFallback();
  const applications = backendItems.length ? fallbackItems : await attachAgentCreditPaymentsFallback(fallbackItems);
  return <OperatorCreditPaymentsPage applications={applications} />;
}
