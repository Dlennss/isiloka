import { UserPlus, UsersRound, WalletCards } from "lucide-react";
import { MasterCreateAgentForm } from "@/components/dashboard/MasterCreateAgentForm";

export default function MasterTambahAgentPage() {
  return (
    <main className="-m-2 min-h-screen bg-[#eef7f2] p-3 text-slate-950 sm:p-5 lg:p-7">
      <section className="mx-auto flex w-full max-w-6xl flex-col gap-5">
        <div className="relative isolate overflow-hidden rounded-[30px] bg-[radial-gradient(circle_at_92%_0%,rgba(163,230,53,0.54),transparent_26%),linear-gradient(135deg,#052e26_0%,#057a45_58%,#3bd64a_110%)] px-5 py-6 text-white shadow-[0_24px_60px_rgba(4,120,87,0.22)] sm:px-7 lg:px-9 lg:py-8">
          <div className="pointer-events-none absolute -right-16 -top-20 h-56 w-56 rounded-full border border-white/20 bg-white/10" />
          <div className="pointer-events-none absolute bottom-0 right-28 h-32 w-32 rounded-full bg-lime-300/20 blur-2xl" />
          <div className="relative flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
            <div className="max-w-2xl">
              <p className="mb-3 inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/12 px-3 py-1 text-[11px] font-black uppercase tracking-[0.24em] text-lime-100">
                <UserPlus className="h-3.5 w-3.5" />
                Marketing Agent
              </p>
              <h1 className="text-3xl font-black tracking-normal sm:text-4xl">Tambah Agent Baru</h1>
              <p className="mt-3 max-w-xl text-sm font-medium leading-6 text-emerald-50/90 sm:text-base">
                Buat akun agent retail Isiloka dengan cepat, rapi, dan siap dipakai untuk login.
              </p>
            </div>
            <div className="grid grid-cols-2 gap-3 sm:w-[340px]">
              <div className="rounded-[22px] bg-white/12 p-4 ring-1 ring-white/15 backdrop-blur">
                <UsersRound className="h-6 w-6 text-lime-100" />
                <p className="mt-3 text-xl font-black">Agent</p>
                <p className="mt-1 text-[11px] font-semibold text-white/70">Role otomatis</p>
              </div>
              <div className="rounded-[22px] bg-white p-4 text-emerald-800 shadow-lg">
                <WalletCards className="h-6 w-6" />
                <p className="mt-3 text-xl font-black">Start</p>
                <p className="mt-1 text-[11px] font-semibold text-emerald-700/70">Level awal</p>
              </div>
            </div>
          </div>
        </div>

        <MasterCreateAgentForm useRetailEndpoint />
      </section>
    </main>
  );
}
