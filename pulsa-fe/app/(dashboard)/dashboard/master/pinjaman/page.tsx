import { Sparkles, UsersRound } from "lucide-react";
import { getAppServerSession } from "@/lib/server-auth";
import { getAgentCreditApplications } from "@/lib/api.auth";
import { attachAgentCreditPaymentsFallback } from "@/lib/agent-credit-payment-fallback.server";
import { MasterAgentCreditApplicationList } from "@/components/dashboard/MasterAgentCreditApplicationList";

type SessionShape = {
  backendToken?: string;
  user?: { role?: string };
};

export default async function MasterDashboardPage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  const rawApplications = session?.backendToken ? await getAgentCreditApplications(session.backendToken) : [];
  const applications = await attachAgentCreditPaymentsFallback(rawApplications);
  const role = String(session?.user?.role || "").trim().toLowerCase();
  const masterItems = applications.filter((item) => {
    if (role === "marketing") return true;
    if (item.status === "submitted" || item.status === "marketing_review" || item.status === "master_review") return true;
    if (item.status !== "approved") return false;
    const loanStatus = String(item.loan_status || "").toLowerCase();
    return loanStatus === "active" || loanStatus === "overdue" || Number(item.outstanding_amount || 0) > 0;
  });
  const reviewMode = role === "marketing" ? "marketing" : "master";

  return (
    <main className="-m-2 min-h-screen bg-[#eef7f2] p-3 text-slate-950 sm:p-5 lg:p-7">
      <section className="mx-auto flex w-full max-w-7xl flex-col gap-5">
        <div className="overflow-hidden rounded-[30px] border border-emerald-100 bg-white shadow-[0_24px_60px_rgba(15,23,42,0.08)]">
          <div className="relative isolate overflow-hidden bg-[radial-gradient(circle_at_92%_10%,rgba(190,242,100,0.55),transparent_28%),linear-gradient(135deg,#053a2f_0%,#05824c_55%,#45d63f_100%)] px-5 py-7 text-white sm:px-7 lg:px-9 lg:py-9">
            <div className="absolute -right-14 -top-20 h-56 w-56 rounded-full border border-white/25 bg-white/10" />
            <div className="absolute bottom-0 right-28 h-32 w-32 rounded-full bg-lime-300/20 blur-2xl" />
            <div className="relative flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between">
              <div className="max-w-2xl">
                <p className="mb-3 inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/12 px-3 py-1 text-[11px] font-black uppercase tracking-[0.24em] text-lime-100">
                  <Sparkles className="h-3.5 w-3.5" />
                  Pertemuan & Selfie
                </p>
                <h1 className="text-3xl font-black tracking-normal sm:text-4xl">Pendampingan lapangan marketing</h1>
                <p className="mt-3 max-w-xl text-sm font-medium leading-6 text-emerald-50/90 sm:text-base">
                  Pantau seluruh status agent binaan, lengkapi dokumen yang diperlukan, lalu ikuti proses keputusan operator.
                </p>
              </div>
              <div className="rounded-3xl border border-white/20 bg-white/12 p-4 shadow-[inset_0_1px_0_rgba(255,255,255,0.25)] backdrop-blur">
                <div className="flex items-center gap-3">
                  <div className="grid h-14 w-14 place-items-center rounded-2xl bg-white text-emerald-700 shadow-lg">
                    <UsersRound className="h-7 w-7" />
                  </div>
                  <div>
                    <p className="text-xs font-bold uppercase tracking-[0.18em] text-lime-100">Mode Marketing</p>
                    <p className="text-xl font-black">Dampingi Agent</p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div className="p-4 sm:p-6 lg:p-7">
            <div className="grid gap-5">
              <MasterAgentCreditApplicationList
                applications={masterItems}
                mode={reviewMode}
                eyebrow="Pengajuan Masuk"
                title="Daftar pendampingan agent"
                emptyTitle="Belum ada pengajuan"
                emptyDescription="Data agent binaan akan muncul di sini setelah pengajuan dibuat dan tetap terlihat setelah operator mengambil keputusan."
              />
            </div>
          </div>
        </div>
      </section>
    </main>
  );
}
