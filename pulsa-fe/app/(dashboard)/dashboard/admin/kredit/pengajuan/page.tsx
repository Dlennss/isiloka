import { BadgeCheck, ClipboardList, ShieldCheck, WalletCards } from "lucide-react";
import { MasterAgentCreditApplicationList } from "@/components/dashboard/MasterAgentCreditApplicationList";
import { attachAgentCreditPaymentsFallback, getAgentCreditApplicationsDatabaseFallback } from "@/lib/agent-credit-payment-fallback.server";
import { getAgentCreditApplications } from "@/lib/api.auth";
import { getAppServerSession } from "@/lib/server-auth";

type SessionShape = {
  backendToken?: string;
};

function formatIDR(value: number) {
  return `Rp ${new Intl.NumberFormat("id-ID").format(Number(value || 0))}`;
}

function isRejected(status: string) {
  const normalized = String(status || "").toLowerCase();
  return normalized === "rejected" || normalized.includes("rejected");
}

export default async function AdminCreditApplicationsPage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  const backendApplications = session?.backendToken ? await getAgentCreditApplications(session.backendToken) : [];
  const rawApplications = backendApplications.length ? backendApplications : await getAgentCreditApplicationsDatabaseFallback();
  const applications = await attachAgentCreditPaymentsFallback(rawApplications);
  const actionable = applications.filter((item) =>
    ["submitted", "marketing_review", "analysis_review", "master_review", "ready_to_disburse"].includes(String(item.status || "").toLowerCase())
  );
  const approved = applications.filter((item) => item.status === "approved");
  const outstanding = approved.reduce((total, item) => total + Number(item.outstanding_amount || 0), 0);

  const stats = [
    { label: "Perlu Keputusan", value: String(actionable.length), hint: "Bisa approve/reject manual", icon: ClipboardList },
    { label: "Kredit Diterima", value: String(approved.length), hint: formatIDR(approved.reduce((sum, item) => sum + Number(item.approved_amount || 0), 0)), icon: BadgeCheck },
    { label: "Piutang Aktif", value: formatIDR(outstanding), hint: "Pinjaman aktif agent", icon: WalletCards },
    { label: "Ditolak", value: String(applications.filter((item) => isRejected(item.status)).length), hint: "Arsip keputusan", icon: ShieldCheck },
  ];

  return (
    <main className="-m-2 min-h-screen bg-[#eef7f2] p-3 text-slate-950 sm:p-5 lg:p-7">
      <section className="mx-auto flex w-full max-w-7xl flex-col gap-5">
        <div className="overflow-hidden rounded-[30px] border border-emerald-100 bg-white shadow-[0_24px_60px_rgba(15,23,42,0.08)]">
          <div className="relative isolate overflow-hidden bg-[radial-gradient(circle_at_92%_10%,rgba(190,242,100,0.55),transparent_28%),linear-gradient(135deg,#053a2f_0%,#05824c_55%,#45d63f_100%)] px-5 py-7 text-white sm:px-7 lg:px-9 lg:py-9">
            <p className="mb-3 inline-flex rounded-full border border-white/15 bg-white/12 px-3 py-1 text-[11px] font-black uppercase tracking-[0.24em] text-lime-100">
              Admin Kredit
            </p>
            <h1 className="text-3xl font-black tracking-normal sm:text-4xl">Approval Kredit Agent</h1>
            <p className="mt-3 max-w-2xl text-sm font-medium leading-6 text-emerald-50/90 sm:text-base">
              Admin bisa mengambil keputusan manual, memantau limit, aktivitas modal, dokumen survey marketing, dan riwayat transaksi agent.
            </p>
          </div>

          <div className="space-y-5 p-4 sm:p-6 lg:p-7">
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
              {stats.map((item) => {
                const Icon = item.icon;
                return (
                  <div key={item.label} className="rounded-3xl border border-slate-200 bg-white p-4 shadow-[0_14px_32px_rgba(15,23,42,0.06)]">
                    <div className="flex items-center justify-between gap-4">
                      <div>
                        <p className="text-xs font-bold text-slate-500">{item.label}</p>
                        <p className="mt-1 text-2xl font-black text-slate-950">{item.value}</p>
                        <p className="mt-1 text-xs font-semibold text-slate-400">{item.hint}</p>
                      </div>
                      <div className="grid h-12 w-12 place-items-center rounded-2xl bg-emerald-50 text-emerald-700">
                        <Icon className="h-6 w-6" />
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>

            <MasterAgentCreditApplicationList
              applications={applications}
              mode="admin"
              eyebrow="Approval Kredit"
              title="Semua Pengajuan Kredit Agent"
              emptyTitle="Belum ada pengajuan kredit"
              emptyDescription="Data pengajuan dari agent atau marketing akan tampil di sini."
            />
          </div>
        </div>
      </section>
    </main>
  );
}
