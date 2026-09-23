import { BadgeCheck, History, ShieldCheck, WalletCards } from "lucide-react";
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

export default async function AnalystApprovedHistoryPage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  const rawApplications = session?.backendToken ? await getAgentCreditApplications(session.backendToken) : [];
  const applications = await attachAgentCreditPaymentsFallback(rawApplications);
  const approvedItems = applications.filter((item) => item.status === "approved");
  const activeApproved = approvedItems.filter((item) => String(item.loan_status || "").toLowerCase() === "active").length;
  const paidApproved = approvedItems.filter((item) => String(item.loan_status || "").toLowerCase() === "paid").length;
  const unusedApproved = approvedItems.filter((item) => String(item.loan_status || "").toLowerCase() === "active" && Number(item.outstanding_amount || 0) <= 0).length;
  const approvedNominal = approvedItems.reduce((total, item) => total + Number(item.approved_amount || 0), 0);

  const stats = [
    { label: "Total ACC", value: String(approvedItems.length), hint: "Disetujui operator", icon: BadgeCheck, tone: "from-emerald-500 to-lime-400" },
    { label: "Limit Aktif", value: String(activeApproved), hint: unusedApproved ? `${unusedApproved} belum dipakai` : "Masih berjalan", icon: ShieldCheck, tone: "from-sky-500 to-cyan-400" },
    { label: "Siklus Selesai", value: String(paidApproved), hint: "Status pengajuan selesai", icon: History, tone: "from-violet-500 to-fuchsia-500" },
    { label: "Nominal ACC", value: formatIDR(approvedNominal), hint: "Total disetujui", icon: WalletCards, tone: "from-amber-500 to-orange-400" },
  ];

  return (
    <main className="-m-2 min-h-screen bg-[#eef7f2] p-3 text-slate-950 sm:p-5 lg:p-7">
      <section className="mx-auto flex w-full max-w-7xl flex-col gap-5">
        <div className="overflow-hidden rounded-[28px] border border-emerald-100 bg-white shadow-[0_24px_60px_rgba(15,23,42,0.08)]">
          <div className="relative isolate overflow-hidden bg-[radial-gradient(circle_at_88%_6%,rgba(163,230,53,0.50),transparent_28%),linear-gradient(135deg,#052e26_0%,#047857_56%,#30c654_115%)] px-5 py-6 text-white sm:px-7 lg:px-9 lg:py-8">
            <div className="absolute -right-14 -top-20 h-56 w-56 rounded-full border border-white/20 bg-white/10" />
            <div className="relative flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between">
              <div className="max-w-2xl">
                <p className="mb-3 inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/12 px-3 py-1 text-[11px] font-black uppercase tracking-[0.24em] text-lime-100">
                  <History className="h-3.5 w-3.5" />
                  Riwayat Operator
                </p>
                <h1 className="text-3xl font-black tracking-normal sm:text-4xl">Riwayat ACC Operator</h1>
                <p className="mt-3 max-w-xl text-sm font-medium leading-6 text-emerald-50/90 sm:text-base">
                  Arsip keputusan operator yang sudah menyetujui pinjaman, lengkap dengan dokumen, nominal, dan riwayat pembayaran.
                </p>
              </div>
              <div className="rounded-3xl border border-white/20 bg-white/12 p-4 shadow-[inset_0_1px_0_rgba(255,255,255,0.25)] backdrop-blur">
                <div className="flex items-center gap-3">
                  <div className="grid h-14 w-14 place-items-center rounded-2xl bg-white text-emerald-700 shadow-lg">
                    <BadgeCheck className="h-7 w-7" />
                  </div>
                  <div>
                    <p className="text-xs font-bold uppercase tracking-[0.18em] text-lime-100">Sudah Final</p>
                    <p className="text-xl font-black">ACC Operator</p>
                  </div>
                </div>
              </div>
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

            <div className="rounded-3xl border border-emerald-100 bg-emerald-50/70 p-4">
              <div className="flex items-start gap-3">
                <span className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-white text-emerald-700 shadow-sm">
                  <WalletCards className="h-5 w-5" />
                </span>
                <div>
                  <p className="text-sm font-black text-slate-950">Riwayat ini tidak perlu aksi lagi.</p>
                  <p className="mt-1 text-xs font-semibold leading-5 text-slate-500">
                    Gunakan untuk audit keputusan operator dan memantau pembayaran agent yang sudah cair.
                  </p>
                </div>
              </div>
            </div>

            <MasterAgentCreditApplicationList
              applications={approvedItems}
              mode="analyst"
              showActions={false}
              eyebrow="Sudah ACC"
              title="Daftar Pinjaman Disetujui"
              emptyTitle="Belum ada ACC operator"
              emptyDescription="Data yang disetujui operator akan masuk ke riwayat ini setelah keputusan final dibuat."
            />
          </div>
        </div>
      </section>
    </main>
  );
}
