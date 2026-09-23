"use client";

import { type ChangeEvent, type FormEvent, type ReactNode, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Camera, Check, FileSignature, Loader2, LocateFixed, MapPin, PlusCircle, Save, X } from "lucide-react";

type ApiBody = {
  ok?: boolean;
  error?: string;
};

type AgentMember = {
  id: number;
  nama?: string;
  email?: string;
  phone?: string;
  role?: string;
  aktif?: boolean;
};

type StoredImage = {
  name: string;
  type: string;
  size: number;
  data_url: string;
};

type DocumentKey = "ktp" | "store" | "selfie_ktp" | "selfie_marketing";

type MarketingAgentCreditCreateFormProps = {
  defaultOpen?: boolean;
};

type SurveyLocation = { latitude: number; longitude: number; accuracy: number };

const emptyApplicant = {
  memberId: "",
  agentName: "",
  storeName: "",
  nik: "",
  whatsapp: "",
  email: "",
  monthlyTransactions: "",
  requestedAmount: "500000",
  familyName: "",
  familyRelation: "",
  familyWhatsapp: "",
  homeAddress: "",
  storeAddress: "",
};

const documentOptions: Array<{ key: DocumentKey; label: string; helper: string }> = [
  { key: "ktp", label: "Foto KTP", helper: "KTP asli agent terlihat jelas" },
  { key: "store", label: "Foto Toko", helper: "Tampak depan toko atau usaha" },
  { key: "selfie_ktp", label: "Selfie Agent Pegang KTP", helper: "Wajah agent dan KTP terlihat" },
  { key: "selfie_marketing", label: "Foto Agent & Marketing", helper: "Agent dan marketing dalam satu foto" },
];

function authHeader(): Record<string, string> {
  if (typeof window === "undefined") return {};
  const token = window.localStorage.getItem("auth_token") || "";
  return token ? { Authorization: `Bearer ${token}` } : {};
}

function readImage(file: File): Promise<StoredImage> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = () => reject(new Error(`Gagal membaca ${file.name}`));
    reader.onload = () => resolve({ name: file.name, type: file.type, size: file.size, data_url: String(reader.result || "") });
    reader.readAsDataURL(file);
  });
}

function Field({ label, children }: { label: string; children: ReactNode }) {
  return <label className="block min-w-0"><span className="mb-1.5 block text-[10px] font-black text-slate-600">{label}</span>{children}</label>;
}

const inputClassName = "h-11 w-full rounded-lg border border-slate-200 bg-[#fbfffd] px-3 text-sm font-bold text-slate-950 outline-none placeholder:text-slate-400 focus:border-emerald-400 focus:ring-2 focus:ring-emerald-100";
const textAreaClassName = "w-full resize-none rounded-lg border border-slate-200 bg-[#fbfffd] px-3 py-3 text-sm font-bold text-slate-950 outline-none placeholder:text-slate-400 focus:border-emerald-400 focus:ring-2 focus:ring-emerald-100";

export function MarketingAgentCreditCreateForm({ defaultOpen = false }: MarketingAgentCreditCreateFormProps) {
  const router = useRouter();
  const [open, setOpen] = useState(defaultOpen);
  const [busy, setBusy] = useState(false);
  const [loadingAgents, setLoadingAgents] = useState(true);
  const [agents, setAgents] = useState<AgentMember[]>([]);
  const [applicant, setApplicant] = useState(emptyApplicant);
  const [documents, setDocuments] = useState<Partial<Record<DocumentKey, StoredImage>>>({});
  const [surveyLocation, setSurveyLocation] = useState<SurveyLocation | null>(null);
  const [locating, setLocating] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  useEffect(() => {
    let cancelled = false;
    async function loadAgents() {
      try {
        const response = await fetch("/api/me/retail/downlines", { headers: authHeader(), cache: "no-store" });
        const body = (await response.json().catch(() => ({}))) as { ok?: boolean; items?: AgentMember[]; rows?: AgentMember[]; error?: string };
        if (!response.ok || !body.ok) throw new Error(body.error || "Data agent gagal dimuat");
        if (!cancelled) setAgents((Array.isArray(body.items) ? body.items : body.rows || []).filter((item) => item.aktif !== false && String(item.role || "").toLowerCase() === "agent"));
      } catch (error) {
        if (!cancelled) setMessage({ type: "error", text: error instanceof Error ? error.message : "Data agent gagal dimuat" });
      } finally {
        if (!cancelled) setLoadingAgents(false);
      }
    }
    void loadAgents();
    return () => { cancelled = true; };
  }, []);

  function updateApplicant(key: keyof typeof emptyApplicant, value: string) {
    setApplicant((current) => ({ ...current, [key]: value }));
  }

  function selectAgent(memberId: string) {
    const selected = agents.find((item) => String(item.id) === memberId);
    setApplicant((current) => ({
      ...current,
      memberId,
      agentName: selected?.nama || "",
      email: selected?.email || "",
      whatsapp: selected?.phone || "",
    }));
  }

  async function captureDocument(key: DocumentKey, event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file) return;
    if (!file.type.startsWith("image/")) {
      setMessage({ type: "error", text: "Dokumen harus berupa foto." });
      return;
    }
    if (file.size > 6 * 1024 * 1024) {
      setMessage({ type: "error", text: "Ukuran setiap foto maksimal 6 MB." });
      return;
    }
    try {
      const image = await readImage(file);
      setDocuments((current) => ({ ...current, [key]: image }));
      setMessage(null);
    } catch (error) {
      setMessage({ type: "error", text: error instanceof Error ? error.message : "Foto gagal dibaca" });
    }
  }

  function captureLocation() {
    if (!navigator.geolocation) {
      setMessage({ type: "error", text: "Perangkat ini tidak mendukung lokasi." });
      return;
    }
    setLocating(true);
    setMessage(null);
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setSurveyLocation({ latitude: position.coords.latitude, longitude: position.coords.longitude, accuracy: position.coords.accuracy });
        setLocating(false);
      },
      () => {
        setMessage({ type: "error", text: "Lokasi belum dapat diambil. Izinkan akses lokasi lalu coba lagi." });
        setLocating(false);
      },
      { enableHighAccuracy: true, timeout: 15000, maximumAge: 0 },
    );
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const requiredText = [applicant.memberId, applicant.agentName, applicant.storeName, applicant.nik, applicant.whatsapp, applicant.monthlyTransactions, applicant.familyName, applicant.familyRelation, applicant.familyWhatsapp, applicant.homeAddress, applicant.storeAddress];
    if (requiredText.some((value) => !value.trim())) {
      setMessage({ type: "error", text: "Lengkapi seluruh data wajib agent, usaha, dan kontak keluarga." });
      return;
    }
    if (!/^\d{16}$/.test(applicant.nik.replace(/\D/g, ""))) {
      setMessage({ type: "error", text: "NIK wajib terdiri dari 16 angka." });
      return;
    }
    const phone = applicant.whatsapp.replace(/\D/g, "").replace(/^62/, "0");
    const familyPhone = applicant.familyWhatsapp.replace(/\D/g, "").replace(/^62/, "0");
    if (!/^08\d{8,12}$/.test(phone) || !/^08\d{8,12}$/.test(familyPhone)) {
      setMessage({ type: "error", text: "Nomor WhatsApp agent dan keluarga harus valid dan diawali 08." });
      return;
    }
    if (phone === familyPhone) {
      setMessage({ type: "error", text: "Nomor WhatsApp keluarga harus berbeda dari nomor agent." });
      return;
    }
    const requestedAmount = Number(applicant.requestedAmount.replace(/[^\d]/g, ""));
    if (!Number.isFinite(requestedAmount) || requestedAmount < 100000) {
      setMessage({ type: "error", text: "Nominal kredit minimal Rp100.000 dan tidak boleh kosong." });
      return;
    }
    if (documentOptions.some((item) => !documents[item.key])) {
      setMessage({ type: "error", text: "Empat dokumen foto lapangan wajib dilengkapi marketing." });
      return;
    }

    setBusy(true);
    setMessage(null);
    try {
      const response = await fetch("/api/agent-credit/manual-applications", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({
          member_id: Number(applicant.memberId),
          requested_amount: requestedAmount,
          applicant_data: {
            agent_name: applicant.agentName.trim(),
            store_name: applicant.storeName.trim(),
            nik: applicant.nik.replace(/\D/g, ""),
            whatsapp: phone,
            email: applicant.email.trim(),
            monthly_transactions: Number(applicant.monthlyTransactions.replace(/[^\d]/g, "")),
            family_name: applicant.familyName.trim(),
            family_relation: applicant.familyRelation.trim(),
            family_whatsapp: familyPhone,
            home_address: applicant.homeAddress.trim(),
            store_address: applicant.storeAddress.trim(),
            input_by: "marketing",
            survey_taken_at: new Date().toISOString(),
            survey_location: surveyLocation,
          },
          document_data: documents,
          agent_signature: "",
          terms_accepted: false,
        }),
      });
      const body = (await response.json().catch(() => ({}))) as ApiBody;
      if (!response.ok || !body.ok) throw new Error(body.error || "Pengajuan gagal disimpan");
      setApplicant(emptyApplicant);
      setDocuments({});
      setSurveyLocation(null);
      setMessage({ type: "success", text: "Pengajuan dan dokumen lapangan berhasil disimpan. Lanjutkan verifikasi melalui Antrean Survei." });
      router.refresh();
    } catch (error) {
      setMessage({ type: "error", text: error instanceof Error ? error.message : "Pengajuan gagal disimpan" });
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="overflow-hidden rounded-lg border border-emerald-200 bg-white shadow-[0_18px_42px_rgba(6,78,59,0.08)]">
      <header className="flex flex-col gap-4 bg-[linear-gradient(135deg,#13464b,#087e8b)] px-4 py-5 text-white sm:flex-row sm:items-center sm:justify-between sm:px-6">
        <div><p className="text-[9px] font-black uppercase tracking-[0.18em] text-lime-200">Operasional Lapangan</p><h1 className="mt-1 text-2xl font-black">Pengajuan & Dokumen Kredit</h1><p className="mt-1 text-xs font-semibold text-emerald-100/75">Isi data agent dan ambil dokumen langsung saat kunjungan marketing.</p></div>
        <span className="grid h-11 w-11 shrink-0 place-items-center rounded-lg bg-white text-emerald-700"><FileSignature className="h-5 w-5" /></span>
      </header>

      <div className="p-3 sm:p-4">
        <button type="button" onClick={() => setOpen((value) => !value)} className="inline-flex h-11 items-center gap-2 rounded-lg bg-emerald-700 px-4 text-xs font-black text-white transition hover:bg-emerald-800">
          {open ? <X className="h-4 w-4" /> : <PlusCircle className="h-4 w-4" />}{open ? "Tutup Form Pengajuan" : "Buka Form Pengajuan"}
        </button>

        {open ? (
          <form onSubmit={submit} className="mt-4 space-y-4">
            <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
              <Field label="Nama Agent"><select value={applicant.memberId} onChange={(event) => selectAgent(event.target.value)} disabled={loadingAgents} className={inputClassName}><option value="">{loadingAgents ? "Memuat agent..." : "Pilih akun agent"}</option>{agents.map((agent) => <option key={agent.id} value={agent.id}>{agent.nama || agent.email} - ID {agent.id}</option>)}</select></Field>
              <Field label="Nama Toko"><input value={applicant.storeName} onChange={(event) => updateApplicant("storeName", event.target.value)} placeholder="Nama toko/usaha" className={inputClassName} /></Field>
              <Field label="NIK"><input value={applicant.nik} onChange={(event) => updateApplicant("nik", event.target.value.replace(/\D/g, "").slice(0, 16))} placeholder="16 digit NIK" inputMode="numeric" className={inputClassName} /></Field>
              <Field label="Nomor WhatsApp"><input value={applicant.whatsapp} onChange={(event) => updateApplicant("whatsapp", event.target.value)} placeholder="08xxxxxxxxxx" inputMode="tel" className={inputClassName} /></Field>
              <Field label="Email"><input value={applicant.email} onChange={(event) => updateApplicant("email", event.target.value)} placeholder="Opsional" type="email" className={inputClassName} /></Field>
              <Field label="Transaksi per Bulan"><input value={applicant.monthlyTransactions} onChange={(event) => updateApplicant("monthlyTransactions", event.target.value.replace(/\D/g, ""))} placeholder="Contoh: 150" inputMode="numeric" className={inputClassName} /></Field>
              <Field label="Nominal Kredit Diajukan"><input value={applicant.requestedAmount} onChange={(event) => updateApplicant("requestedAmount", event.target.value.replace(/\D/g, ""))} placeholder="500000" inputMode="numeric" className={inputClassName} /></Field>
              <Field label="Nama Kontak Keluarga"><input value={applicant.familyName} onChange={(event) => updateApplicant("familyName", event.target.value)} placeholder="Nama keluarga" className={inputClassName} /></Field>
              <Field label="Hubungan Keluarga"><input value={applicant.familyRelation} onChange={(event) => updateApplicant("familyRelation", event.target.value)} placeholder="Orang tua / saudara" className={inputClassName} /></Field>
              <Field label="WhatsApp Keluarga"><input value={applicant.familyWhatsapp} onChange={(event) => updateApplicant("familyWhatsapp", event.target.value)} placeholder="08xxxxxxxxxx" inputMode="tel" className={inputClassName} /></Field>
            </div>

            <div className="grid gap-4 xl:grid-cols-2">
              <div className="grid gap-3">
                <Field label="Alamat Rumah"><textarea value={applicant.homeAddress} onChange={(event) => updateApplicant("homeAddress", event.target.value)} rows={3} placeholder="Alamat rumah lengkap" className={textAreaClassName} /></Field>
                <Field label="Alamat Toko"><textarea value={applicant.storeAddress} onChange={(event) => updateApplicant("storeAddress", event.target.value)} rows={4} placeholder="Alamat toko atau usaha" className={textAreaClassName} /></Field>
              </div>

              <fieldset className="rounded-lg border border-emerald-200 bg-emerald-50/35 p-3">
                <legend className="px-2 text-xs font-black text-slate-950">Dokumen Survei Marketing</legend>
                <p className="mb-3 text-[11px] font-semibold leading-5 text-slate-500">Foto wajib diambil langsung saat kunjungan dan harus terlihat jelas.</p>
                <div className="grid gap-2 sm:grid-cols-2">
                  {documentOptions.map((item) => {
                    const saved = documents[item.key];
                    return <label key={item.key} className={saved ? "flex min-h-20 cursor-pointer items-center gap-2 overflow-hidden rounded-lg border border-emerald-300 bg-white p-2" : "flex min-h-20 cursor-pointer items-center gap-3 rounded-lg border border-dashed border-emerald-300 bg-white p-3 transition hover:bg-emerald-50"}>
                      <input type="file" accept="image/*" capture={item.key === "selfie_ktp" || item.key === "selfie_marketing" ? "user" : "environment"} className="sr-only" onChange={(event) => void captureDocument(item.key, event)} />
                      {saved ? <img src={saved.data_url} alt={`Preview ${item.label}`} className="h-16 w-20 shrink-0 rounded-lg bg-slate-100 object-contain" /> : <span className="grid h-9 w-9 shrink-0 place-items-center rounded-lg bg-emerald-50 text-emerald-700"><Camera className="h-4 w-4" /></span>}
                      <span className="min-w-0 flex-1"><span className="block text-xs font-black text-slate-950">{item.label}</span><span className="mt-1 block text-[10px] font-semibold text-slate-500">{saved ? "Foto siap disimpan" : item.helper}</span><span className="mt-1 block text-[10px] font-black text-emerald-700">{saved ? "Ketuk untuk ganti" : "Buka kamera"}</span></span>
                      {saved ? <Check className="h-4 w-4 shrink-0 text-emerald-700" /> : null}
                    </label>;
                  })}
                </div>
                <div className="mt-3 flex flex-col gap-2 rounded-lg border border-emerald-200 bg-white p-3 sm:flex-row sm:items-center sm:justify-between">
                  <div className="flex min-w-0 items-center gap-2"><MapPin className="h-4 w-4 shrink-0 text-emerald-700" /><div className="min-w-0"><p className="text-[10px] font-black text-slate-700">Lokasi kunjungan (opsional)</p><p className="truncate text-[9px] font-semibold text-slate-500">{surveyLocation ? `${surveyLocation.latitude.toFixed(6)}, ${surveyLocation.longitude.toFixed(6)} · akurasi ${Math.round(surveyLocation.accuracy)} m` : "Belum diambil"}</p></div></div>
                  <button type="button" onClick={captureLocation} disabled={locating} className="inline-flex h-9 shrink-0 items-center justify-center gap-2 rounded-lg bg-emerald-50 px-3 text-[10px] font-black text-emerald-800 disabled:opacity-50">{locating ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <LocateFixed className="h-3.5 w-3.5" />}{surveyLocation ? "Ambil Ulang" : "Ambil Lokasi"}</button>
                </div>
              </fieldset>
            </div>

            {message ? <div className={message.type === "success" ? "rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-xs font-black text-emerald-700" : "rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-xs font-black text-rose-600"}>{message.text}</div> : null}
            <button type="submit" disabled={busy || loadingAgents} className="inline-flex h-12 w-full items-center justify-center gap-2 rounded-lg bg-[linear-gradient(135deg,#087e8b,#65a30d)] text-sm font-black text-white shadow-[0_14px_28px_rgba(5,150,105,0.20)] disabled:opacity-50">{busy ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}{busy ? "Menyimpan Pengajuan..." : "Simpan Pengajuan & Dokumen"}</button>
          </form>
        ) : null}
      </div>
    </section>
  );
}
