import { Archive, BadgeCheck, FileX2, History, WalletCards } from "lucide-react";
import { getAppServerSession } from "@/lib/server-auth";
import { getAgentCreditApplications } from "@/lib/api.auth";
import { attachAgentCreditPaymentsFallback } from "@/lib/agent-credit-payment-fallback.server";
import { MasterAgentCreditApplicationList } from "@/components/dashboard/MasterAgentCreditApplicationList";

type SessionShape = {
  backendToken?: string;
};

function formatIDR(value: number) {
  return `Rp ${new Intl.NumberFormat("id-ID").format(Number(value || 0))}`;
}

function isArchivedCreditStatus(status: string) {
  return status === "analysis_rejected" || status === "master_rejected" || status === "rejected";
}

function isFinishedLoan(status: string, loanStatus?: string) {
  const normalizedLoan = String(loanStatus || "").toLowerCase();
  return status === "approved" && (normalizedLoan === "paid" || normalizedLoan === "cancelled");
}

export default async function MasterCreditHistoryPage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  const rawApplications = session?.backendToken ? await getAgentCreditApplications(session.backendToken) : [];
  const applications = await attachAgentCreditPaymentsFallback(rawApplications);
  const historyItems = applications.filter((item) => isArchivedCreditStatus(item.status) || isFinishedLoan(item.status, item.loan_status));

  const rejected = historyItems.filter((item) => isArchivedCreditStatus(item.status)).length;
  const paid = historyItems.filter((item) => String(item.loan_status || "").toLowerCase() === "paid").length;
  const finishedLimit = historyItems.reduce((total, item) => total + Number(item.approved_amount || 0), 0);
  const stats = [
    { label: "Total Riwayat", value: String(historyItems.length), hint: "Data selesai/arsip", icon: Archive, tone: "from-emerald-500 to-lime-400" },
    { label: "Ditolak", value: String(rejected), hint: "Keputusan master", icon: FileX2, tone: "from-rose-500 to-orange-400" },
    { label: "Siklus Selesai", value: String(paid), hint: "Pengajuan selesai", icon: BadgeCheck, tone: "from-sky-500 to-cyan-400" },
    { label: "Nominal Selesai", value: formatIDR(finishedLimit), hint: "Total ACC di arsip", icon: WalletCards, tone: "from-violet-500 to-fuchsia-500" },
  ];

  return (
    <main className="-m-2 min-h-screen bg-[#eef7f2] p-3 text-slate-950 sm:p-5 lg:p-7">
      <section className="mx-auto flex w-full max-w-7xl flex-col gap-5">
        <div className="overflow-hidden rounded-[28px] border border-emerald-100 bg-white shadow-[0_24px_60px_rgba(15,23,42,0.08)]">
          <div className="relative isolate overflow-hidden bg-[radial-gradient(circle_at_90%_10%,rgba(163,230,53,0.55),transparent_28%),linear-gradient(135deg,#052e26_0%,#057a45_48%,#3bd64a_100%)] px-5 py-6 text-white sm:px-7 lg:px-9 lg:py-8">
            <div className="absolute -right-16 -top-20 h-56 w-56 rounded-full border border-white/20 bg-white/10" />
            <div className="relative max-w-2xl">
              <p className="mb-3 inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/12 px-3 py-1 text-[11px] font-black uppercase tracking-[0.24em] text-lime-100">
                <History className="h-3.5 w-3.5" />
                Arsip Kredit
              </p>
              <h1 className="text-3xl font-black tracking-normal sm:text-4xl">Riwayat Pinjaman Agent</h1>
              <p className="mt-3 max-w-xl text-sm font-medium leading-6 text-emerald-50/90 sm:text-base">
                Data yang sudah ditolak atau pinjaman yang sudah selesai disimpan di sini agar panel kerja tetap bersih.
              </p>
            </div>
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
                      <div className={`grid h-12 w-12 place-items-center rounded-2xl bg-linear-to-br ${item.tone} text-white shadow-lg`}>
                        <Icon className="h-6 w-6" />
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>

            <MasterAgentCreditApplicationList applications={historyItems} mode="master" showActions={false} />
          </div>
        </div>
      </section>
    </main>
  );
}
