import Link from "next/link";
import { BadgeCheck, Camera, CheckCircle2, Clock3, FileSignature, UserRoundPlus, UsersRound, WalletCards } from "lucide-react";
import { getAppServerSession } from "@/lib/server-auth";
import { getAgentCreditApplications } from "@/lib/api.auth";
import { attachAgentCreditPaymentsFallback } from "@/lib/agent-credit-payment-fallback.server";

type SessionShape = {
  backendToken?: string;
};

function formatIDR(value: number) {
  return `Rp ${new Intl.NumberFormat("id-ID").format(Number(value || 0))}`;
}

function isTruthy(value: unknown) {
  return value === true || value === "true" || value === 1 || value === "1";
}

function hasDocument(item: { document_data?: Record<string, unknown> }, key: string) {
  const value = item.document_data?.[key];
  if (typeof value === "string") return value.length > 10;
  if (value && typeof value === "object") return true;
  return false;
}

export default async function MasterDashboardPage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  const rawApplications = session?.backendToken ? await getAgentCreditApplications(session.backendToken) : [];
  const applications = await attachAgentCreditPaymentsFallback(rawApplications);
  const waiting = applications.filter((item) => item.status === "submitted" || item.status === "marketing_review").length;
  const inAnalysis = applications.filter((item) => item.status === "analysis_review").length;
  const approved = applications.filter((item) => item.status === "approved").length;
  const paidOff = applications.filter((item) => item.loan_status === "paid" || item.status === "paid").length;
  const activeCredit = applications.filter((item) => item.status === "approved" && String(item.loan_status || "").toLowerCase() !== "paid");
  const activeLimit = activeCredit.reduce((total, item) => total + Number(item.outstanding_amount || 0), 0);
  const needsSelfie = applications.filter(
    (item) => (item.status === "submitted" || item.status === "marketing_review") && !hasDocument(item, "selfie_marketing")
  );
  const readyForAnalysis = applications.filter((item) => item.status === "analysis_review");
  const offlineBills = activeCredit.filter((item) => Number(item.outstanding_amount || 0) > 0);

  const stats = [
    { label: "Total Peminjam", value: String(applications.length), hint: "Seluruh pengajuan", icon: UsersRound, tone: "bg-emerald-50 text-emerald-700" },
    { label: "Butuh Review", value: String(waiting), hint: "Menunggu cek marketing", icon: Clock3, tone: "bg-amber-50 text-amber-700" },
    { label: "Sudah Diterima", value: String(approved), hint: "Aktif dipantau", icon: BadgeCheck, tone: "bg-sky-50 text-sky-700" },
    { label: "Siklus Selesai", value: String(paidOff), hint: "Status pengajuan selesai", icon: CheckCircle2, tone: "bg-lime-50 text-lime-700" },
  ];
  const workCards = [
    { label: "Perlu Pendampingan", value: String(needsSelfie.length), hint: "Selfie bersama agent belum ada", icon: Camera },
    { label: "Siap Dicek Operator", value: String(inAnalysis), hint: "Data dan pertemuan sudah lengkap", icon: FileSignature },
    { label: "Agent Perlu Follow-up", value: String(offlineBills.length), hint: "Perlu dihubungi marketing", icon: WalletCards },
  ];
  const closestSelfie = needsSelfie.slice(0, 2);
  const closestAnalysis = readyForAnalysis.slice(0, 2);

  return (
    <main className="-m-2 min-h-screen bg-[#eef7f2] p-3 text-slate-950 sm:p-5 lg:p-7">
      <section className="mx-auto flex w-full max-w-7xl flex-col gap-5">
        <div className="relative overflow-hidden rounded-[30px] bg-[radial-gradient(circle_at_88%_10%,rgba(217,255,117,0.42),transparent_31%),linear-gradient(135deg,#053b2f_0%,#08764f_58%,#35c63f_135%)] p-6 text-[#f5fff8] shadow-[0_24px_60px_rgba(4,120,87,0.18)] sm:p-8">
          <div className="absolute -right-14 -top-20 h-56 w-56 rounded-full border border-white/20 bg-white/10" />
          <div className="absolute right-16 top-12 hidden h-32 w-32 rounded-full border border-white/15 bg-white/8 lg:block" />
          <div className="relative flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <p className="mb-3 text-[11px] font-black uppercase tracking-[0.22em] text-[#d9ff75]">
                Ruang Data Peminjam
              </p>
              <h1 className="max-w-2xl text-3xl font-black leading-tight sm:text-4xl">Monitoring Kredit Agent</h1>
              <p className="mt-3 max-w-3xl text-sm font-semibold leading-6 text-[#dcfce7]">
                Semua pengajuan tersusun dari yang terbaru. Marketing cek data agent, tanda tangan, selfie lapangan, lalu kirim ke operator untuk keputusan akhir.
              </p>
            </div>
            <div className="grid h-20 w-20 shrink-0 place-items-center rounded-[26px] border border-white/20 bg-white/18 text-white shadow-lg backdrop-blur">
              <WalletCards className="h-10 w-10" strokeWidth={2.3} />
            </div>
          </div>
        </div>

        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
          {stats.map((item) => {
            const Icon = item.icon;
            return (
              <div key={item.label} className="rounded-3xl border border-emerald-100 bg-white p-5 shadow-[0_16px_36px_rgba(15,23,42,0.05)]">
                <div className="flex items-center justify-between gap-4">
                  <div>
                    <p className="text-xs font-black text-slate-500">{item.label}</p>
                    <p className="mt-2 text-2xl font-black text-slate-950">{item.value}</p>
                    <p className="mt-1 text-xs font-semibold text-slate-400">{item.hint}</p>
                  </div>
                  <div className={`grid h-12 w-12 place-items-center rounded-2xl ${item.tone}`}>
                    <Icon className="h-6 w-6" />
                  </div>
                </div>
              </div>
            );
          })}
        </div>

        <section className="relative overflow-hidden rounded-[28px] border border-emerald-100 bg-white p-5 shadow-[0_18px_42px_rgba(15,23,42,0.05)]">
          <div className="absolute -right-10 -top-16 h-44 w-44 rounded-full bg-emerald-50" />
          <div className="relative">
            <p className="text-[11px] font-black uppercase tracking-[0.22em] text-emerald-600">Meja Kerja Marketing</p>
            <div className="mt-2 flex flex-col gap-2 lg:flex-row lg:items-end lg:justify-between">
              <div>
                <h2 className="text-2xl font-black text-slate-950">Kerjakan yang paling penting</h2>
                <p className="mt-1 max-w-3xl text-sm font-semibold leading-6 text-slate-500">
                  Daftarkan agent, bantu pengajuan, pantau dokumen dan aktivitas agent. Keputusan kredit tetap dilakukan operator.
                </p>
              </div>
            </div>

            <div className="mt-5 grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
              {workCards.map((item) => {
                const Icon = item.icon;
                return (
                  <div key={item.label} className="rounded-3xl border border-slate-200 bg-white p-5 shadow-[0_12px_28px_rgba(15,23,42,0.04)]">
                    <div className="grid h-11 w-11 place-items-center rounded-2xl bg-emerald-50 text-emerald-700">
                      <Icon className="h-5 w-5" />
                    </div>
                    <p className="mt-5 text-xs font-black text-slate-500">{item.label}</p>
                    <p className="mt-2 text-3xl font-black text-slate-950">{item.value}</p>
                    <p className="mt-1 text-xs font-semibold text-slate-400">{item.hint}</p>
                  </div>
                );
              })}
            </div>

            <div className="mt-6 grid grid-cols-1 gap-4 lg:grid-cols-2">
              <div className="rounded-3xl border border-slate-200 bg-white p-5">
                <div className="flex items-center justify-between gap-3">
                  <div>
                    <p className="text-[11px] font-black uppercase tracking-[0.18em] text-emerald-600">Pekerjaan Terdekat</p>
                    <h3 className="mt-2 text-lg font-black text-slate-950">Perlu selfie pertemuan</h3>
                  </div>
                  <Camera className="h-5 w-5 text-emerald-700" />
                </div>
                <div className="mt-4 space-y-2">
                  {closestSelfie.length ? (
                    closestSelfie.map((item) => (
                      <Link
                        key={item.id}
                        href="/dashboard/master/pinjaman"
                        className="flex items-center justify-between gap-3 rounded-2xl bg-emerald-50 px-4 py-3 text-sm font-bold text-[#063f32] transition hover:bg-emerald-100"
                      >
                        <span className="min-w-0">
                          <span className="block truncate">{item.member_name || "Agent Isiloka"}</span>
                          <span className="block text-xs font-semibold text-slate-500">Selfie bersama marketing belum lengkap</span>
                        </span>
                        <span className="text-xs font-black text-emerald-700">{formatIDR(item.requested_amount)}</span>
                      </Link>
                    ))
                  ) : (
                    <div className="rounded-2xl bg-slate-50 px-4 py-4 text-sm font-semibold text-slate-500">
                      Tidak ada agent yang menunggu pertemuan.
                    </div>
                  )}
                </div>
              </div>

              <div className="rounded-3xl border border-slate-200 bg-white p-5">
                <div className="flex items-center justify-between gap-3">
                  <div>
                    <p className="text-[11px] font-black uppercase tracking-[0.18em] text-emerald-600">Kirim Lanjutan</p>
                    <h3 className="mt-2 text-lg font-black text-slate-950">Siap diperiksa operator</h3>
                  </div>
                  <FileSignature className="h-5 w-5 text-emerald-700" />
                </div>
                <div className="mt-4 space-y-2">
                  {closestAnalysis.length ? (
                    closestAnalysis.map((item) => (
                      <Link
                        key={item.id}
                        href="/dashboard/master/operator"
                        className="flex items-center justify-between gap-3 rounded-2xl bg-emerald-50 px-4 py-3 text-sm font-bold text-[#063f32] transition hover:bg-emerald-100"
                      >
                        <span className="min-w-0">
                          <span className="block truncate">{item.member_name || "Agent Isiloka"}</span>
                          <span className="block text-xs font-semibold text-slate-500">
                            {isTruthy(item.applicant_data?.terms_accepted) ? "Data dan persetujuan lengkap" : "Perlu cek persetujuan agent"}
                          </span>
                        </span>
                        <span className="text-xs font-black text-emerald-700">{formatIDR(item.requested_amount)}</span>
                      </Link>
                    ))
                  ) : (
                    <div className="rounded-2xl bg-slate-50 px-4 py-4 text-sm font-semibold text-slate-500">
                      Belum ada pengajuan yang siap dicek operator.
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        </section>

        <section className="grid grid-cols-1 gap-3 md:grid-cols-3">
          <Link href="/dashboard/master/tambah-agent" className="rounded-3xl border border-emerald-100 bg-white p-5 shadow-sm transition hover:-translate-y-0.5 hover:shadow-lg">
            <UserRoundPlus className="h-6 w-6 text-emerald-700" />
            <h3 className="mt-4 text-base font-black text-slate-950">Daftar Agent</h3>
            <p className="mt-1 text-sm font-semibold leading-6 text-slate-500">Tambahkan akun agent baru dari satu panel.</p>
          </Link>
          <Link href="/dashboard/master/input-pinjaman-manual" className="rounded-3xl border border-emerald-100 bg-white p-5 shadow-sm transition hover:-translate-y-0.5 hover:shadow-lg">
            <FileSignature className="h-6 w-6 text-emerald-700" />
            <h3 className="mt-4 text-base font-black text-slate-950">Bantu Pengajuan</h3>
            <p className="mt-1 text-sm font-semibold leading-6 text-slate-500">Masukkan pengajuan manual bila agent dibantu marketing.</p>
          </Link>
          <div className="rounded-3xl border border-emerald-100 bg-[#063f32] p-5 text-[#f1fff7] shadow-sm">
            <WalletCards className="h-6 w-6 text-[#d9ff75]" />
            <h3 className="mt-4 text-base font-black">Total Limit Berjalan</h3>
            <p className="mt-2 text-2xl font-black">{formatIDR(activeLimit)}</p>
            <p className="mt-1 text-sm font-semibold leading-6 text-emerald-100">Pantau modal aktif dan aktivitas transaksi agent.</p>
          </div>
        </section>
      </section>
    </main>
  );
}
