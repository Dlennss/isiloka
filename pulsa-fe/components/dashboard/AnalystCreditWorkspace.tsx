import {
  Archive,
  BadgeCheck,
  ClipboardList,
  FileSearch,
  FileCheck2,
  type LucideIcon,
  MessageSquareText,
  ReceiptText,
  ShieldAlert,
  ShieldCheck,
  WalletCards,
  XCircle,
} from "lucide-react";
import { getAgentCreditApplications, type AgentCreditApplication } from "@/lib/api.auth";
import { attachAgentCreditPaymentsFallback, getAgentCreditApplicationsDatabaseFallback } from "@/lib/agent-credit-payment-fallback.server";
import { getAppServerSession } from "@/lib/server-auth";
import { MasterAgentCreditApplicationList } from "@/components/dashboard/MasterAgentCreditApplicationList";

type SessionShape = {
  backendToken?: string;
};

export type AnalystCreditWorkspaceView =
  | "decision"
  | "queue"
  | "repayment"
  | "proof"
  | "rejected"
  | "archive";

const viewConfig = {
  decision: {
    eyebrow: "Operator Kredit",
    title: "Keputusan Akhir Kredit Agent",
    desc: "Periksa data agent, dokumen, tanda tangan agent, lalu beri keputusan akhir.",
    listTitle: "Data Siap Diputuskan",
    emptyTitle: "Belum ada data untuk diputuskan",
    emptyDescription: "Pengajuan agent yang sudah lengkap akan muncul di sini.",
    icon: ShieldCheck,
    showActions: true,
  },
  queue: {
    eyebrow: "Antrean Operator",
    title: "Antrean Pengajuan",
    desc: "Urutan pengajuan yang menunggu pengecekan dokumen, risiko, dan keputusan operator.",
    listTitle: "Antrean Siap Dicek",
    emptyTitle: "Antrean masih kosong",
    emptyDescription: "Belum ada pengajuan baru dari agent.",
    icon: ClipboardList,
    showActions: true,
  },
  repayment: {
    eyebrow: "Monitor Aktivitas",
    title: "Pantau Aktivitas Agent",
    desc: "Pantau transaksi terakhir agent dan temukan agent yang perlu di-follow-up.",
    listTitle: "Aktivitas Agent",
    emptyTitle: "Belum ada agent aktif",
    emptyDescription: "Agent yang sudah disetujui operator akan tampil di sini beserta transaksi terakhirnya.",
    icon: WalletCards,
    showActions: false,
  },
  proof: {
    eyebrow: "Aktivitas Modal",
    title: "Riwayat Aktivitas Modal",
    desc: "Pantau aktivitas agent dan keputusan operator dalam satu tempat.",
    listTitle: "Aktivitas Agent",
    emptyTitle: "Belum ada aktivitas modal",
    emptyDescription: "Aktivitas agent akan tampil setelah transaksi tercatat.",
    icon: FileCheck2,
    showActions: false,
  },
  rejected: {
    eyebrow: "Penolakan & Catatan",
    title: "Data Ditolak dan Catatan Risiko",
    desc: "Pengajuan yang perlu diperbaiki agent atau tidak layak dicairkan.",
    listTitle: "Daftar Penolakan",
    emptyTitle: "Belum ada data ditolak",
    emptyDescription: "Keputusan tolak dari operator akan masuk ke daftar ini.",
    icon: XCircle,
    showActions: true,
  },
  archive: {
    eyebrow: "Arsip Keputusan",
    title: "Arsip Semua Keputusan",
    desc: "Riwayat keputusan operator untuk audit, pengecekan ulang, dan monitoring kredit.",
    listTitle: "Arsip Keputusan Operator",
    emptyTitle: "Arsip masih kosong",
    emptyDescription: "Data yang sudah diputuskan operator akan disimpan di sini.",
    icon: Archive,
    showActions: true,
  },
} satisfies Record<AnalystCreditWorkspaceView, {
  eyebrow: string;
  title: string;
  desc: string;
  listTitle: string;
  emptyTitle: string;
  emptyDescription: string;
  icon: LucideIcon;
  showActions: boolean;
}>;

function formatIDR(value: number) {
  return `Rp ${new Intl.NumberFormat("id-ID").format(Number(value || 0))}`;
}

function isRejected(item: AgentCreditApplication) {
  const status = String(item.status || "").toLowerCase();
  return status === "rejected" || status.includes("rejected");
}

function isPaid(item: AgentCreditApplication) {
  const loanStatus = String(item.loan_status || "").toLowerCase();
  return loanStatus === "paid";
}

function hasPaymentProof(item: AgentCreditApplication) {
  return (item.payments || []).some((payment) => {
    const src = payment.payment_proof?.data_url;
    return typeof src === "string" && src.startsWith("data:image/");
  });
}

function isPendingCreditApplication(item: AgentCreditApplication) {
  return [
    "draft",
    "submitted",
    "marketing_review",
    "analysis_review",
    "master_review",
    "ready_to_disburse",
  ].includes(String(item.status || "").toLowerCase());
}

function inactiveDays(item: AgentCreditApplication) {
  const source = item.last_transaction_at || item.loan_approved_at;
  if (!source) return 999;
  const timestamp = new Date(source).getTime();
  if (!Number.isFinite(timestamp)) return 999;
  return Math.max(0, Math.floor((Date.now() - timestamp) / 86400000));
}

function getItemsForView(view: AnalystCreditWorkspaceView, applications: AgentCreditApplication[]) {
  const analysisItems = applications.filter(isPendingCreditApplication);
  const approvedItems = applications.filter((item) => item.status === "approved");
  const activeCredits = approvedItems.filter((item) => !isPaid(item));
  const usedCredits = activeCredits;
  const followUpCredits = activeCredits.filter((item) => inactiveDays(item) >= 2);
  const proofItems = applications.filter((item) => Number(item.payment_count || 0) > 0 || (item.payments || []).length > 0 || hasPaymentProof(item));
  const rejectedItems = applications.filter(isRejected);
  const archiveItems = applications.filter((item) => item.status === "approved" || isRejected(item) || isPaid(item));

  if (view === "repayment") return activeCredits;
  if (view === "proof") return proofItems;
  if (view === "rejected") return rejectedItems;
  if (view === "archive") return archiveItems;
  return analysisItems;
}

function getAgentName(item: AgentCreditApplication) {
  const applicantName = item.applicant_data?.nama_lengkap;
  return item.agent_name || (typeof applicantName === "string" && applicantName.trim() ? applicantName : "") || item.agent_email || "Agent Isiloka";
}

function getAgentStore(item: AgentCreditApplication) {
  const storeName = item.applicant_data?.nama_toko;
  const storeAddress = item.applicant_data?.alamat_toko;
  return item.store_name ||
    (typeof storeName === "string" && storeName.trim() ? storeName : "") ||
    (typeof storeAddress === "string" && storeAddress.trim() ? storeAddress : "") ||
    "Retail Isiloka";
}

export async function AnalystCreditWorkspace({ view }: { view: AnalystCreditWorkspaceView }) {
  const config = viewConfig[view];
  const Icon = config.icon;
  const session = (await getAppServerSession()) as SessionShape | null;
  const backendApplications = session?.backendToken ? await getAgentCreditApplications(session.backendToken, 50) : [];
  const rawApplications = backendApplications.length ? backendApplications : await getAgentCreditApplicationsDatabaseFallback();
  const applications = backendApplications.length ? rawApplications : await attachAgentCreditPaymentsFallback(rawApplications);
  const analysisItems = applications.filter(isPendingCreditApplication);
  const approvedItems = applications.filter((item) => item.status === "approved");
  const activeCredits = approvedItems.filter((item) => !isPaid(item));
  const usedCredits = activeCredits;
  const followUpCredits: AgentCreditApplication[] = activeCredits.filter((item) => inactiveDays(item) >= 2);
  const rejectedItems = applications.filter(isRejected);
  const items = getItemsForView(view, applications);
  const nominalApproved = approvedItems.reduce((total, item) => total + Number(item.approved_amount || 0), 0);
  const decisionReadyItems = analysisItems.filter((item) => Number(item.recommended_amount || item.requested_amount || 0) > 0);
  const proofItems = applications.filter((item) => Number(item.payment_count || 0) > 0 || (item.payments || []).length > 0 || hasPaymentProof(item));
  const priorityReviewItems = analysisItems.slice(0, 3);
  const activePriorityItems = activeCredits.slice(0, 3);
  const usedPriorityItems = followUpCredits.slice(0, 3);
  const paidCredits = approvedItems.filter(isPaid);
  const totalPaidAmount = approvedItems.reduce((total, item) => {
    const paymentsTotal = (item.payments || []).reduce((sum, payment) => sum + Number(payment.amount || 0), 0);
    return total + Number(item.paid_amount || paymentsTotal || 0);
  }, 0);
  const totalOutstandingAmount = activeCredits.reduce((total, item) => total + Number(item.outstanding_amount || 0), 0);

  const stats = [
    { label: view === "decision" ? "Berkas Masuk" : "Perlu Keputusan", value: String(analysisItems.length), hint: "Dikirim agent", icon: ShieldCheck, tone: "from-emerald-500 to-lime-400" },
    { label: "Kredit Diterima", value: String(approvedItems.length), hint: formatIDR(nominalApproved), icon: BadgeCheck, tone: "from-sky-500 to-cyan-400" },
    { label: "Perlu Follow-up", value: String(followUpCredits.length), hint: "Tidak transaksi 2 hari atau lebih", icon: WalletCards, tone: "from-amber-500 to-orange-400" },
    { label: "Ditolak", value: String(rejectedItems.length), hint: "Perlu catatan", icon: ShieldAlert, tone: "from-rose-500 to-orange-500" },
  ];

  const decisionCards = [
    { label: "Berkas Masuk", value: analysisItems.length, hint: "Menunggu pemeriksaan akhir", icon: FileSearch },
    { label: "Siap Keputusan", value: decisionReadyItems.length, hint: "Nominal dan dokumen siap dicek", icon: ShieldCheck },
    { label: "Bukti Bayar", value: proofItems.length, hint: "Pembayaran agent tercatat", icon: ReceiptText },
    { label: "Catatan Risiko", value: rejectedItems.length, hint: "Data yang perlu evaluasi ulang", icon: MessageSquareText },
  ];
  const queueCards = [
    { label: "Masuk Antrean", value: analysisItems.length, hint: "Dikirim dari agent", icon: ClipboardList },
    { label: "Siap Dicek", value: decisionReadyItems.length, hint: "Nominal dan berkas tersedia", icon: FileSearch },
    { label: "Sudah Diterima", value: approvedItems.length, hint: "Keputusan operator selesai", icon: BadgeCheck },
    { label: "Perlu Catatan", value: rejectedItems.length, hint: "Pengajuan ditolak", icon: MessageSquareText },
  ];
  const repaymentCards = [
    { label: "Agent Aktif", value: Math.max(0, activeCredits.length - followUpCredits.length), hint: "Transaksi hari ini atau kemarin", icon: ShieldCheck },
    { label: "Perlu Follow-up", value: followUpCredits.length, hint: "Tidak transaksi 2 hari atau lebih", icon: ShieldAlert },
    { label: "Belum Pernah Transaksi", value: activeCredits.filter((item) => !item.last_transaction_at).length, hint: "Perlu dihubungi", icon: ReceiptText },
    { label: "Total Agent", value: activeCredits.length, hint: "Modal aktif dipantau operator", icon: WalletCards },
  ];

  return (
    <main className="min-h-screen overflow-x-hidden bg-[#eef7f2] p-0 text-slate-950 sm:p-5 lg:p-7">
      <section className="mx-auto flex w-full max-w-7xl min-w-0 flex-col gap-3 sm:gap-5">
        {view === "decision" ? (
          <div className="rounded-[26px] border border-emerald-100 bg-white/80 p-5 shadow-[0_18px_50px_rgba(15,23,42,0.06)] backdrop-blur sm:p-6">
            <p className="text-[11px] font-black uppercase tracking-[0.26em] text-emerald-700">Operator Kredit</p>
            <h1 className="mt-2 text-2xl font-black tracking-normal text-slate-950 sm:text-3xl">Keputusan Akhir Kredit Agent</h1>
            <p className="mt-2 max-w-2xl text-sm font-semibold leading-6 text-slate-500">
              Operator menjadi pintu akhir. Cek data agent, dokumen, tanda tangan marketing, riwayat pembayaran, lalu putuskan diterima atau ditolak.
            </p>
          </div>
        ) : null}
        <div className="min-w-0 overflow-hidden border border-emerald-100 bg-white shadow-[0_24px_60px_rgba(15,23,42,0.08)] sm:rounded-[28px]">
          <div className="relative isolate overflow-hidden bg-[radial-gradient(circle_at_88%_6%,rgba(190,242,100,0.52),transparent_28%),linear-gradient(135deg,#052e26_0%,#047857_56%,#31c653_115%)] px-4 py-5 text-white sm:px-7 lg:px-9 lg:py-8">
            <div className="absolute -right-14 -top-20 h-56 w-56 rounded-full border border-white/20 bg-white/10" />
            {view === "decision" ? <div className="absolute bottom-0 right-28 h-32 w-32 rounded-full border border-white/15 bg-white/8" /> : null}
            <div className="relative flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between">
              <div className="max-w-2xl">
                <p className="mb-3 inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/12 px-3 py-1 text-[11px] font-black uppercase tracking-[0.24em] text-lime-100">
                  <Icon className="h-3.5 w-3.5" />
                  {config.eyebrow}
                </p>
                <h1 className="text-3xl font-black tracking-normal sm:text-4xl">{config.title}</h1>
                <p className="mt-3 max-w-xl text-sm font-medium leading-6 text-emerald-50/90 sm:text-base">{config.desc}</p>
              </div>
              <div className="rounded-3xl border border-white/20 bg-white/12 p-4 shadow-[inset_0_1px_0_rgba(255,255,255,0.25)] backdrop-blur">
                <div className="flex items-center gap-3">
                  <div className="grid h-14 w-14 place-items-center rounded-2xl bg-white text-emerald-700 shadow-lg">
                    <Icon className="h-7 w-7" />
                  </div>
                  <div>
                    <p className="text-xs font-bold uppercase tracking-[0.18em] text-lime-100">{view === "decision" ? "Ruang Keputusan" : "Panel Operator"}</p>
                    <p className="text-xl font-black">{view === "decision" ? "Final Kredit" : "Isiloka"}</p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div className="min-w-0 space-y-4 p-2 sm:space-y-5 sm:p-6 lg:p-7">
            {view === "decision" ? (
              <div className="hidden grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
                {stats.map((item) => {
                  const StatIcon = item.icon;
                  return (
                    <div key={item.label} className="rounded-3xl border border-slate-200 bg-white p-4 shadow-[0_14px_32px_rgba(15,23,42,0.06)]">
                      <div className="flex items-center justify-between gap-4">
                        <div>
                          <p className="text-xs font-bold text-slate-500">{item.label}</p>
                          <p className="mt-1 text-2xl font-black text-slate-950">{item.value}</p>
                          <p className="mt-1 text-xs font-semibold text-slate-400">{item.hint}</p>
                        </div>
                        <div className={`grid h-12 w-12 place-items-center rounded-2xl bg-linear-to-br ${item.tone} text-white shadow-lg`}>
                          <StatIcon className="h-6 w-6" />
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            ) : null}

            {view === "decision" ? (
              <div className="hidden overflow-hidden rounded-[28px] border border-emerald-200 bg-[linear-gradient(135deg,#ffffff_0%,#f2fff8_62%,#e4f8ee_100%)] p-4 shadow-[0_16px_36px_rgba(6,78,59,0.06)] sm:p-5">
                <div className="relative">
                  <div className="absolute -right-16 -top-20 hidden h-44 w-44 rounded-full bg-emerald-100/70 sm:block" />
                  <div className="relative">
                    <p className="text-[11px] font-black uppercase tracking-[0.26em] text-emerald-700">Meja Keputusan Operator</p>
                    <h2 className="mt-2 text-2xl font-black tracking-normal text-slate-950">Kontrol Kelayakan & Keputusan Akhir</h2>
                    <p className="mt-2 max-w-3xl text-sm font-semibold leading-6 text-slate-500">
                      Semua pengajuan di sini sudah melewati pendampingan marketing. Periksa identitas, dokumen inti, selfie pertemuan, tanda tangan, nominal, dan catatan lapangan sebelum memberi keputusan.
                    </p>
                  </div>
                </div>

                <div className="mt-5 grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
                  {decisionCards.map((card) => {
                    const CardIcon = card.icon;
                    return (
                      <div key={card.label} className="rounded-3xl border border-slate-200 bg-white p-4 shadow-[0_14px_30px_rgba(15,23,42,0.05)]">
                        <div className="grid h-11 w-11 place-items-center rounded-2xl bg-emerald-50 text-emerald-700">
                          <CardIcon className="h-5 w-5" />
                        </div>
                        <p className="mt-4 text-xs font-black text-slate-500">{card.label}</p>
                        <p className="mt-1 text-2xl font-black text-slate-950">{card.value}</p>
                        <p className="mt-1 text-xs font-semibold leading-5 text-slate-400">{card.hint}</p>
                      </div>
                    );
                  })}
                </div>

                <div className="mt-6">
                  <div className="mb-3 flex flex-col gap-1 sm:flex-row sm:items-center sm:gap-3">
                    <p className="text-[11px] font-black uppercase tracking-[0.24em] text-emerald-700">Berkas Prioritas</p>
                    <p className="text-xs font-semibold text-slate-400">Periksa yang paling siap lebih dulu.</p>
                  </div>
                  <div className="grid grid-cols-1 gap-3 lg:grid-cols-2">
                    <div className="rounded-3xl border border-slate-200 bg-white p-4">
                      <div className="flex items-center justify-between gap-3">
                        <h3 className="text-base font-black text-slate-950">Menunggu keputusan operator</h3>
                        <span className="rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-700">{priorityReviewItems.length} data</span>
                      </div>
                      <div className="mt-4 space-y-2">
                        {priorityReviewItems.length ? (
                          priorityReviewItems.map((item) => (
                            <div key={item.id} className="flex items-center justify-between gap-3 rounded-2xl bg-slate-50 p-3">
                              <div className="min-w-0">
                                <p className="truncate text-sm font-black text-slate-950">{getAgentName(item)}</p>
                                <p className="truncate text-xs font-semibold text-slate-500">{getAgentStore(item)}</p>
                              </div>
                              <p className="shrink-0 text-sm font-black text-emerald-700">{formatIDR(Number(item.requested_amount || item.recommended_amount || 0))}</p>
                            </div>
                          ))
                        ) : (
                          <div className="rounded-2xl bg-slate-50 p-4 text-sm font-semibold text-slate-400">Belum ada pengajuan yang perlu diputuskan.</div>
                        )}
                      </div>
                    </div>

                    <div className="rounded-3xl border border-slate-200 bg-white p-4">
                      <div className="flex items-center justify-between gap-3">
                        <h3 className="text-base font-black text-slate-950">Kredit aktif dipantau</h3>
                        <span className="rounded-full bg-lime-50 px-3 py-1 text-xs font-black text-lime-700">{activePriorityItems.length} aktif</span>
                      </div>
                      <div className="mt-4 space-y-2">
                        {activePriorityItems.length ? (
                          activePriorityItems.map((item) => {
                            const idleDays = inactiveDays(item);
                            return (
                            <div key={item.id} className="flex items-center justify-between gap-3 rounded-2xl bg-slate-50 p-3">
                              <div className="min-w-0">
                                <p className="truncate text-sm font-black text-slate-950">{getAgentName(item)}</p>
                                <p className="truncate text-xs font-semibold text-slate-500">{idleDays >= 999 ? "Belum pernah transaksi" : `Tidak transaksi ${idleDays} hari`}</p>
                              </div>
                              <p className="shrink-0 text-sm font-black text-amber-700">Follow-up</p>
                            </div>
                          );
                          })
                        ) : (
                          <div className="rounded-2xl bg-slate-50 p-4 text-sm font-semibold text-slate-400">Belum ada kredit aktif yang perlu dipantau.</div>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            ) : view === "queue" ? (
              <div className="space-y-4">
                <div className="relative overflow-hidden rounded-[28px] border border-emerald-200 bg-[linear-gradient(135deg,#ffffff_0%,#f5fff9_68%,#e6fbef_100%)] p-5 shadow-[0_16px_36px_rgba(6,78,59,0.06)]">
                  <div className="absolute -right-12 -top-16 h-40 w-40 rounded-full bg-lime-100/70" />
                  <div className="relative">
                    <p className="text-[11px] font-black uppercase tracking-[0.26em] text-emerald-700">Antrean Berkas</p>
                    <h2 className="mt-2 text-2xl font-black tracking-normal text-slate-950">Berkas Siap Dicek Operator</h2>
                    <p className="mt-2 max-w-3xl text-sm font-semibold leading-6 text-slate-500">
                      Cek antrean dari marketing: data agent, dokumen KTP, foto toko, selfie bersama marketing, tanda tangan, nominal kredit, dan catatan lapangan.
                    </p>
                  </div>
                </div>

                <div className="rounded-[26px] border border-emerald-100 bg-slate-50 p-4">
                  <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
                    <span className="grid h-14 w-14 shrink-0 place-items-center rounded-2xl bg-emerald-50 text-emerald-700">
                      <ClipboardList className="h-7 w-7" />
                    </span>
                    <div>
                      <p className="text-[11px] font-black uppercase tracking-[0.2em] text-emerald-700">Fokus Operator</p>
                      <h3 className="text-xl font-black text-slate-950">{analysisItems.length} pengajuan perlu ditangani</h3>
                      <p className="mt-1 text-xs font-semibold leading-5 text-slate-500">
                        Mulai dari pengajuan yang paling lengkap. Setelah dicek, buka detail dan beri keputusan akhir melalui tombol setuju atau tolak.
                      </p>
                    </div>
                  </div>
                </div>

                <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
                  {queueCards.map((card) => {
                    const CardIcon = card.icon;
                    return (
                      <div key={card.label} className="rounded-3xl border border-slate-200 bg-white p-4 shadow-[0_14px_30px_rgba(15,23,42,0.05)]">
                        <div className="flex items-start justify-between gap-3">
                          <div>
                            <p className="text-xs font-black text-slate-500">{card.label}</p>
                            <p className="mt-1 text-2xl font-black text-slate-950">{card.value}</p>
                            <p className="mt-1 text-xs font-semibold leading-5 text-slate-400">{card.hint}</p>
                          </div>
                          <span className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-emerald-50 text-emerald-700">
                            <CardIcon className="h-5 w-5" />
                          </span>
                        </div>
                      </div>
                    );
                  })}
                </div>

                <div className="rounded-[28px] border border-slate-200 bg-white p-4">
                  <div className="mb-3 flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
                    <div>
                      <p className="text-[11px] font-black uppercase tracking-[0.24em] text-emerald-700">Prioritas Antrean</p>
                      <h3 className="mt-1 text-lg font-black text-slate-950">Pengajuan paling baru dari marketing</h3>
                    </div>
                    <span className="w-fit rounded-full bg-emerald-800 px-4 py-2 text-xs font-black uppercase tracking-[0.08em] text-white">Operator</span>
                  </div>
                  <div className="space-y-2">
                    {priorityReviewItems.length ? (
                      priorityReviewItems.map((item) => (
                        <div key={item.id} className="flex flex-col gap-3 rounded-3xl border border-slate-100 bg-slate-50 p-4 sm:flex-row sm:items-center sm:justify-between">
                          <div className="min-w-0">
                            <p className="text-[11px] font-black uppercase tracking-[0.14em] text-emerald-700">KSA-{item.id}</p>
                            <p className="mt-1 truncate text-base font-black text-slate-950">{getAgentName(item)}</p>
                            <p className="mt-1 truncate text-xs font-semibold text-slate-500">{getAgentStore(item)}</p>
                            <p className="mt-3 text-xs font-bold text-slate-500">Menunggu keputusan operator</p>
                          </div>
                          <p className="shrink-0 text-lg font-black text-emerald-700">{formatIDR(Number(item.requested_amount || item.recommended_amount || 0))}</p>
                        </div>
                      ))
                    ) : (
                      <div className="rounded-2xl bg-slate-50 p-4 text-sm font-semibold text-slate-400">Belum ada antrean dari marketing.</div>
                    )}
                  </div>
                </div>
              </div>
            ) : view === "repayment" ? (
              <div className="space-y-4">
                <div className="relative overflow-hidden rounded-[28px] border border-emerald-200 bg-[linear-gradient(135deg,#ffffff_0%,#f7fff9_66%,#e8fbef_100%)] p-5 shadow-[0_16px_36px_rgba(6,78,59,0.06)]">
                  <div className="absolute -right-12 -top-16 h-40 w-40 rounded-full bg-lime-100/80" />
                  <div className="relative">
                    <p className="text-[11px] font-black uppercase tracking-[0.26em] text-emerald-700">Aktivitas Agent</p>
                    <h2 className="mt-2 text-2xl font-black tracking-normal text-slate-950">Monitor Aktivitas Modal</h2>
                    <p className="mt-2 max-w-3xl text-sm font-semibold leading-6 text-slate-500">
                      Pantau status aktivitas agent, transaksi terakhir, dan agent yang perlu di-follow-up.
                    </p>
                  </div>
                </div>

                <div className="rounded-[28px] border border-emerald-100 bg-slate-50 p-4 sm:p-5">
                  <div className="mb-5 flex flex-col gap-3 sm:flex-row sm:items-start">
                    <span className="grid h-14 w-14 shrink-0 place-items-center rounded-2xl bg-emerald-50 text-emerald-700">
                      <ReceiptText className="h-7 w-7" />
                    </span>
                    <div>
                      <p className="text-[11px] font-black uppercase tracking-[0.24em] text-emerald-700">Monitor Saldo Kredit</p>
                      <h3 className="text-2xl font-black tracking-normal text-slate-950">Aktivitas Agent</h3>
                      <p className="mt-1 max-w-3xl text-sm font-semibold leading-6 text-slate-500">
                        Modal berjalan selama agent aktif menjadi mitra. Buka detail untuk melihat aktivitas dan status agent.
                      </p>
                    </div>
                  </div>

                  <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
                    {repaymentCards.map((card) => {
                      const CardIcon = card.icon;
                      return (
                        <div key={card.label} className="rounded-3xl border border-slate-200 bg-white p-4 shadow-[0_12px_28px_rgba(15,23,42,0.04)]">
                          <div className="flex items-start justify-between gap-3">
                            <div>
                              <p className="text-xs font-black text-slate-500">{card.label}</p>
                              <p className="mt-1 text-2xl font-black text-slate-950">{card.value}</p>
                              <p className="mt-1 text-xs font-semibold leading-5 text-slate-400">{card.hint}</p>
                            </div>
                            <span className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-emerald-50 text-emerald-700">
                              <CardIcon className="h-5 w-5" />
                            </span>
                          </div>
                        </div>
                      );
                    })}
                  </div>

                  <div className="mt-5 space-y-2">
                    {usedPriorityItems.length ? (
                      usedPriorityItems.slice(0, 6).map((item: AgentCreditApplication) => (
                        <div key={item.id} className="flex flex-col gap-3 rounded-3xl border border-slate-100 bg-white p-4 shadow-[0_10px_24px_rgba(15,23,42,0.04)] sm:flex-row sm:items-center sm:justify-between">
                          <div className="min-w-0">
                            <p className="truncate text-base font-black text-slate-950">{getAgentName(item)}</p>
                            <p className="mt-1 truncate text-xs font-semibold text-slate-500">{getAgentStore(item)}</p>
                            <p className="mt-2 text-xs font-bold text-amber-700">
                              {inactiveDays(item) >= 999 ? "Belum pernah transaksi" : `Tidak transaksi ${inactiveDays(item)} hari`}
                            </p>
                          </div>
                          <div className="text-left sm:text-right">
                            <p className="text-lg font-black text-amber-700">Perlu follow-up</p>
                            <p className="mt-1 text-xs font-semibold text-slate-400">
                              Transaksi terakhir: {item.last_transaction_at ? new Intl.DateTimeFormat("id-ID", { day: "2-digit", month: "short", year: "numeric", hour: "2-digit", minute: "2-digit" }).format(new Date(item.last_transaction_at)) : "Belum ada"}
                            </p>
                          </div>
                        </div>
                      ))
                    ) : (
                      <div className="rounded-2xl bg-white p-4 text-sm font-semibold text-slate-400">Belum ada aktivitas agent yang perlu dipantau.</div>
                    )}
                  </div>
                </div>
              </div>
            ) : (
              <div className="rounded-3xl border border-emerald-100 bg-emerald-50/70 p-4">
                <div className="flex items-start gap-3">
                  <span className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-white text-emerald-700 shadow-sm">
                    <ClipboardList className="h-5 w-5" />
                  </span>
                  <div>
                    <p className="text-sm font-black text-slate-950">Operator menjadi keputusan akhir kredit.</p>
                    <p className="mt-1 text-xs font-semibold leading-5 text-slate-500">
                      Data yang masuk ke panel ini diisi langsung oleh agent. Operator cek kelayakan, lalu setujui atau tolak.
                    </p>
                  </div>
                </div>
              </div>
            )}

            <MasterAgentCreditApplicationList
              applications={items}
              mode="analyst"
              showActions={config.showActions}
              enableReportActions={view === "archive"}
              eyebrow={config.eyebrow}
              title={config.listTitle}
              emptyTitle={config.emptyTitle}
              emptyDescription={config.emptyDescription}
            />
          </div>
        </div>
      </section>
    </main>
  );
}
