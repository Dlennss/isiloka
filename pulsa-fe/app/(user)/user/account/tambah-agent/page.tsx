import { MasterCreateAgentForm } from "@/components/dashboard/MasterCreateAgentForm";

export default function MarketingAddAgentPage() {
  return (
    <main className="bg-sky-50 px-4 py-4">
      <section className="mx-auto w-full max-w-md space-y-4">
        <header className="rounded-[22px] bg-[linear-gradient(135deg,#13464b,#087e8b)] p-5 text-white shadow-[0_16px_36px_rgba(6,78,59,0.18)]">
          <p className="text-[10px] font-black uppercase tracking-[0.22em] text-lime-200">Menu Marketing</p>
          <h1 className="mt-2 text-2xl font-black">Tambah Agent</h1>
          <p className="mt-1 text-xs font-semibold leading-5 text-emerald-50/80">Daftarkan agent baru saat kunjungan lapangan.</p>
        </header>
        <MasterCreateAgentForm useRetailEndpoint />
      </section>
    </main>
  );
}
