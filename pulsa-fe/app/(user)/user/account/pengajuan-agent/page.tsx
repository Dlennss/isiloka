import { getAppServerSession } from "@/lib/server-auth";
import { getAgentCreditApplications } from "@/lib/api.auth";
import { MasterAgentCreditApplicationList } from "@/components/dashboard/MasterAgentCreditApplicationList";

type SessionShape = { backendToken?: string };

export default async function MarketingApplicationsPage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  const applications = session?.backendToken ? await getAgentCreditApplications(session.backendToken, 50) : [];
  const items = applications.filter((item) => item.status === "submitted" || item.status === "marketing_review");

  return (
    <main className="bg-sky-50 px-4 py-4">
      <section className="mx-auto w-full max-w-md">
        <MasterAgentCreditApplicationList
          applications={items}
          mode="marketing"
          eyebrow="Menu Marketing"
          title="Pengajuan & Dokumen"
          emptyTitle="Belum ada pengajuan"
          emptyDescription="Pengajuan agent akan muncul di sini untuk dipantau marketing."
        />
      </section>
    </main>
  );
}
