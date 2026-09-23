"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { AppModal } from "@/components/ui/app-modal";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";
import { EmptyState } from "@/components/ui/empty-state";
import { alertError, alertSuccess, alertWarning } from "@/components/ui/alerts";
import { Building2, CheckCircle2, Copy, Landmark, PlusCircle, Ticket, XCircle } from "lucide-react";

type DepositRow = {
  id: number;
  member_id: number;
  ref_id?: string;
  bank_id?: number | null;
  bank_nama?: string;
  bank_nomor_rekening?: string;
  bank_atas_nama?: string;
  amount: number;
  requested_amount?: number;
  unique_code?: number;
  approved_amount?: number;
  metode: string;
  bukti_url: string;
  status: string;
  note: string;
  dibuat_pada: string;
};

type BankOption = {
  id: number;
  nama: string;
  nomor_rekening: string;
  atas_nama: string;
  aktif: boolean;
};

const VA_MIN_AMOUNT = 10_000_000;
const VA_BANKS = [
  { code: "VA24MAN", label: "VA Mandiri" },
  { code: "VA24BRI", label: "VA BRI" },
  { code: "VA24PRMT", label: "VA Permata" },
  { code: "VA24DNMN", label: "VA Danamon" },
  { code: "VA24OCBC", label: "VA OCBC" },
];

const fmtID = (value: number) => Number(value || 0).toLocaleString("id-ID");
const fmtAmountInput = (value: string) => {
  const digits = value.replace(/\D/g, "");
  return digits ? Number(digits).toLocaleString("id-ID") : "";
};

async function apiFetch(path: string, init?: RequestInit) {
  const t = localStorage.getItem("auth_token") || "";
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (t) headers.Authorization = `Bearer ${t}`;
  const res = await fetch(path, { ...init, headers: { ...headers, ...(init?.headers || {}) } });
  const data = await res.json().catch(() => ({}));
  return { ok: res.ok, status: res.status, data };
}

function statusLabel(status: string, metode?: string) {
  const method = String(metode || "").toLowerCase();
  switch (String(status || "").toLowerCase()) {
    case "ticket":
      if (method === "va") return "VA aktif";
      return "Tiket aktif";
    case "pending":
      return "Menunggu admin";
    case "approved":
      return "Approved";
    case "rejected":
      return "Rejected";
    case "cancelled":
      return "Dibatalkan";
    default:
      return status || "-";
  }
}

function isVADeposit(row: DepositRow) {
  return String(row.metode || "").toLowerCase() === "va";
}

function statusClass(status: string) {
  const s = String(status || "").toLowerCase();
  if (s === "approved") return "border-emerald-400/30 bg-emerald-400/10 text-emerald-200";
  if (s === "rejected" || s === "cancelled") return "border-rose-400/30 bg-rose-400/10 text-rose-200";
  if (s === "ticket") return "border-cyan-400/30 bg-cyan-400/10 text-cyan-200";
  return "border-amber-400/30 bg-amber-400/10 text-amber-200";
}

export default function DepositRequestPage() {
  const router = useRouter();
  const [openModal, setOpenModal] = useState(false);
  const [openVAModal, setOpenVAModal] = useState(false);
  const [amount, setAmount] = useState("");
  const [bankID, setBankID] = useState("");
  const [vaAmount, setVAAmount] = useState("");
  const [vaBank, setVABank] = useState("VA24MAN");
  const [loading, setLoading] = useState(false);
  const [vaLoading, setVALoading] = useState(false);
  const [confirmingID, setConfirmingID] = useState<number | null>(null);
  const [cancelingID, setCancelingID] = useState<number | null>(null);

  const [rows, setRows] = useState<DepositRow[]>([]);
  const [banks, setBanks] = useState<BankOption[]>([]);
  const [err, setErr] = useState<string>("");

  useEffect(() => {
    const token = localStorage.getItem("auth_token") || "";
    if (!token) {
      router.replace("/login");
    }
  }, [router]);

  const load = useCallback(async () => {
    const [historyRes, bankRes] = await Promise.all([
      apiFetch("/api/me/history/deposit?limit=50"),
      apiFetch("/api/me/deposit/banks"),
    ]);
    if (!historyRes.ok || !historyRes.data?.ok) {
      setErr(historyRes.data?.error || "failed");
      return;
    }
    setRows(historyRes.data.rows || []);
    if (bankRes.ok && bankRes.data?.ok) {
      setBanks((bankRes.data.items || []).filter((item: BankOption) => item.aktif));
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const selectedBank = banks.find((bank) => String(bank.id) === bankID) || null;
  const selectedVABank = VA_BANKS.find((bank) => bank.code === vaBank) || VA_BANKS[0];
  const activeTickets = useMemo(() => rows.filter((row) => String(row.status || "").toLowerCase() === "ticket"), [rows]);

  const copyText = async (text: string, label: string) => {
    if (!text) {
      await alertWarning(`${label} belum tersedia.`);
      return;
    }
    try {
      await navigator.clipboard.writeText(text);
      await alertSuccess(`${label} berhasil disalin.`);
    } catch {
      await alertError(`Gagal menyalin ${label.toLowerCase()}.`);
    }
  };

  const submit = async () => {
    setErr("");
    const amt = Number(amount);
    if (!amt || amt <= 0) return setErr("Nominal tidak valid");
    if (!bankID) return setErr("Rekening tujuan wajib dipilih");

    setLoading(true);
    try {
      const r = await apiFetch("/api/me/deposit/request", {
        method: "POST",
        body: JSON.stringify({ amount: amt, bank_id: Number(bankID) }),
      });
      if (!r.ok || !r.data?.ok) {
        const message = r.data?.error || "request failed";
        setErr(message);
        await alertError(message);
        return;
      }
      setAmount("");
      setBankID("");
      setOpenModal(false);
      await load();
      await alertSuccess("Tiket deposit berhasil dibuat.");
    } finally {
      setLoading(false);
    }
  };

  const submitVA = async () => {
    setErr("");
    const amt = Number(vaAmount);
    if (!amt || amt < VA_MIN_AMOUNT) return setErr("Minimal deposit VA Rp 10.000.000");
    if (!selectedVABank?.code) return setErr("Bank VA wajib dipilih");

    setVALoading(true);
    try {
      const r = await apiFetch("/api/me/deposit/request/va", {
        method: "POST",
        body: JSON.stringify({ amount: amt, bank: selectedVABank.code }),
      });
      if (!r.ok || !r.data?.ok) {
        const message = r.data?.error || "request failed";
        setErr(message);
        await alertError(message);
        return;
      }
      setVAAmount("");
      setVABank("VA24MAN");
      setOpenVAModal(false);
      await load();
      await alertSuccess("Tiket deposit VA berhasil dibuat.");
    } finally {
      setVALoading(false);
    }
  };

  const confirmTransfer = async (row: DepositRow) => {
    setConfirmingID(row.id);
    try {
      const r = await apiFetch("/api/me/deposit/request/confirm-transfer", {
        method: "POST",
        body: JSON.stringify({ id: row.id }),
      });
      if (!r.ok || !r.data?.ok) {
        await alertError(r.data?.error || "Gagal mengirim tiket ke admin.");
        return;
      }
      await load();
      await alertSuccess("Tiket masuk ke pending admin.");
    } finally {
      setConfirmingID(null);
    }
  };

  const cancelTicket = async (row: DepositRow) => {
    setCancelingID(row.id);
    try {
      const r = await apiFetch("/api/me/deposit/request/cancel-ticket", {
        method: "POST",
        body: JSON.stringify({ id: row.id }),
      });
      if (!r.ok || !r.data?.ok) {
        await alertError(r.data?.error || "Gagal membatalkan tiket.");
        return;
      }
      await load();
      await alertSuccess("Tiket berhasil dibatalkan.");
    } finally {
      setCancelingID(null);
    }
  };

  const columns: DataTableColumn<DepositRow>[] = [
    { id: "id", header: "ID", render: (x) => x.id },
    {
      id: "tanggal",
      header: "Tanggal",
      render: (x) => new Date(x.dibuat_pada).toLocaleString("id-ID"),
    },
    {
      id: "bank",
      header: "Rekening Tujuan",
      render: (x) => (
        <div className="text-white/85">
          <div>{x.bank_nama || "-"}</div>
          <div className="text-xs text-white/55">{x.bank_atas_nama || "-"}</div>
          <div className="font-mono text-xs text-cyan-200">{x.bank_nomor_rekening || "-"}</div>
        </div>
      ),
    },
    {
      id: "amount",
      header: "Nominal",
      render: (x) => <span className="font-semibold text-white">Rp {fmtID(x.amount)}</span>,
    },
    {
      id: "status",
      header: "Status",
      render: (x) => (
        <span className={`rounded-md border px-2 py-1 text-xs font-medium uppercase tracking-wide ${statusClass(x.status)}`}>
          {statusLabel(x.status, x.metode)}
        </span>
      ),
    },
    {
      id: "note",
      header: "Note",
      render: (x) => {
        const note = String(x.note || "").trim();
        if (String(x.status || "").toLowerCase() !== "rejected" || !note) {
          return <span className="text-white/45">-</span>;
        }
        return <span className="text-white/80">{note}</span>;
      },
    },
  ];

  return (
    <div className="space-y-6 p-2">
      <section className="rounded-2xl border border-white/10 bg-card/80 p-5 shadow-[0_16px_40px_-26px_rgba(0,0,0,0.8)]">
        <div className="flex items-center gap-3">
          <span className="grid h-11 w-11 place-items-center rounded-2xl border border-cyan-400/20 bg-cyan-400/10 text-cyan-200">
            <Landmark className="h-5 w-5" />
          </span>
          <div>
            <h1 className="text-lg font-bold text-white">Deposit Member</h1>
            <p className="text-sm text-white/60">Minta tiket deposit, transfer sesuai nominal yang tertera, lalu kirim ke admin.</p>
          </div>
        </div>
      </section>

      <div className="flex flex-wrap justify-end gap-2">
        <Button
          onClick={() => {
            setErr("");
            setOpenVAModal(true);
          }}
          disabled={activeTickets.length >= 5}
          variant="info"
          className="h-10 rounded-xl px-4"
        >
          <Building2 className="h-4 w-4" />
          Deposit VA
        </Button>
        <Button
          onClick={() => {
            setErr("");
            setOpenModal(true);
          }}
          disabled={activeTickets.length >= 5}
          variant="primary"
          className="h-10 rounded-xl px-4"
        >
          <PlusCircle className="h-4 w-4" />
          Minta Tiket
        </Button>
      </div>

      {activeTickets.length > 0 ? (
        <section className="space-y-3 rounded-2xl border border-cyan-400/20 bg-cyan-400/5 p-4 shadow-[0_16px_40px_-26px_rgba(0,0,0,0.8)]">
          <div className="flex items-center gap-2 text-sm font-semibold text-cyan-100">
            <Ticket className="h-4 w-4" />
            Tiket Aktif ({activeTickets.length}/5)
          </div>
          <div className="grid gap-3 lg:grid-cols-2">
            {activeTickets.map((ticket) => {
              const vaTicket = isVADeposit(ticket);
              return (
                <div key={ticket.id} className="rounded-xl border border-white/10 bg-slate-950/45 p-4">
                  <div className="grid gap-3 sm:grid-cols-2">
                    <div>
                      <div className="text-xs uppercase tracking-wide text-white/45">{vaTicket ? "VA tujuan" : "Rekening tujuan"}</div>
                      <div className="mt-1 font-medium text-white">{ticket.bank_nama || "-"}</div>
                      <div className="text-sm text-white/60">{ticket.bank_atas_nama || "-"}</div>
                      <button
                        type="button"
                        onClick={() => void copyText(ticket.bank_nomor_rekening || "", vaTicket ? "Nomor VA" : "Nomor rekening")}
                        className="mt-1 inline-flex items-center gap-1 font-mono text-sm text-cyan-200"
                      >
                        {ticket.bank_nomor_rekening || "-"}
                        <Copy className="h-3.5 w-3.5" />
                      </button>
                      {vaTicket && ticket.ref_id ? <div className="mt-2 font-mono text-xs text-white/45">Tiket {ticket.ref_id}</div> : null}
                    </div>
                    <div>
                      <div className="text-xs uppercase tracking-wide text-white/45">{vaTicket ? "Nominal VA" : "Jumlah transfer"}</div>
                      <button
                        type="button"
                        onClick={() => void copyText(String(ticket.amount || ""), vaTicket ? "Nominal VA" : "Jumlah transfer")}
                        className="mt-1 inline-flex items-center gap-2 text-left text-2xl font-bold text-white"
                      >
                        Rp {fmtID(ticket.amount)}
                        <Copy className="h-4 w-4 text-cyan-200" />
                      </button>
                      {!vaTicket ? <div className="mt-1 text-sm text-white/55">Transfer sesuai nominal yang tertera</div> : null}
                    </div>
                  </div>
                  <div className="mt-4 flex flex-wrap justify-end gap-2">
                    <Button
                      type="button"
                      variant="danger"
                      disabled={cancelingID === ticket.id || confirmingID === ticket.id}
                      onClick={() => void cancelTicket(ticket)}
                      className="rounded-xl"
                    >
                      <XCircle className="h-4 w-4" />
                      {cancelingID === ticket.id ? "Membatalkan..." : "Batalkan Tiket"}
                    </Button>
                    {!vaTicket ? (
                      <Button
                        type="button"
                        variant="success"
                        disabled={confirmingID === ticket.id || cancelingID === ticket.id}
                        onClick={() => void confirmTransfer(ticket)}
                        className="rounded-xl"
                      >
                        <CheckCircle2 className="h-4 w-4" />
                        {confirmingID === ticket.id ? "Mengirim..." : "Sudah Transfer"}
                      </Button>
                    ) : null}
                  </div>
                </div>
              );
            })}
          </div>
        </section>
      ) : null}

      <AppModal
        open={openModal}
        onClose={() => (!loading ? setOpenModal(false) : null)}
        title="Minta Tiket Deposit"
        subtitle="Pilih rekening tujuan dan nominal awal"
      >
        <div className="space-y-3">
          <div>
            <div className="mb-1 text-sm text-white/80">Rekening Tujuan</div>
            <select
              value={bankID}
              onChange={(e) => setBankID(e.target.value)}
              className="h-10 w-full rounded-md border border-white/10 bg-background px-3 text-sm text-white"
            >
              <option value="">Pilih rekening aktif</option>
              {banks.map((bank) => (
                <option key={bank.id} value={bank.id}>
                  {bank.nama} - {bank.atas_nama} - {bank.nomor_rekening}
                </option>
              ))}
            </select>
          </div>
          <div className="rounded-xl border border-white/10 bg-white/5 p-3">
            <div className="text-xs font-medium uppercase tracking-wide text-white/55">Rekening Dipilih</div>
            {selectedBank ? (
              <div className="mt-2 space-y-1 text-sm text-white/85">
                <div>{selectedBank.nama}</div>
                <div>{selectedBank.atas_nama}</div>
                <div className="font-mono text-white">{selectedBank.nomor_rekening}</div>
              </div>
            ) : (
              <div className="mt-2 text-sm text-white/55">Pilih rekening aktif untuk melihat detail tujuan.</div>
            )}
            <div className="mt-3">
              <Button
                type="button"
                variant="info"
                onClick={() => void copyText(selectedBank?.nomor_rekening || "", "Nomor rekening")}
                disabled={!selectedBank}
                className="rounded-xl"
              >
                <Copy className="h-4 w-4" />
                Copy Nomor Rekening
              </Button>
            </div>
          </div>
          <div>
            <div className="mb-1 text-sm text-white/80">Nominal Deposit</div>
            <Input value={fmtAmountInput(amount)} onChange={(e) => setAmount(e.target.value.replace(/\D/g, ""))} placeholder="1.000.000" inputMode="numeric" />
          </div>

          {err ? <div className="rounded-md border border-rose-400/25 bg-rose-500/10 px-3 py-2 text-sm text-rose-300">{err}</div> : null}

          <div className="flex justify-end gap-2 pt-2">
            <Button variant="outline" onClick={() => setOpenModal(false)} disabled={loading} className="rounded-xl">
              Batal
            </Button>
            <Button onClick={submit} disabled={loading} variant="success" className="rounded-xl">
              {loading ? "Processing..." : "Buat Tiket"}
            </Button>
          </div>
        </div>
      </AppModal>

      <AppModal
        open={openVAModal}
        onClose={() => (!vaLoading ? setOpenVAModal(false) : null)}
        title="Deposit VA"
        subtitle="Pilih bank VA dan nominal deposit"
      >
        <div className="space-y-3">
          <div>
            <div className="mb-1 text-sm text-white/80">Bank VA</div>
            <select
              value={vaBank}
              onChange={(e) => setVABank(e.target.value)}
              className="h-10 w-full rounded-md border border-white/10 bg-background px-3 text-sm text-white"
            >
              {VA_BANKS.map((bank) => (
                <option key={bank.code} value={bank.code}>
                  {bank.label}
                </option>
              ))}
            </select>
          </div>
          <div>
            <div className="mb-1 text-sm text-white/80">Nominal Deposit</div>
            <Input value={fmtAmountInput(vaAmount)} onChange={(e) => setVAAmount(e.target.value.replace(/\D/g, ""))} placeholder="10.000.000" inputMode="numeric" />
          </div>

          {err ? <div className="rounded-md border border-rose-400/25 bg-rose-500/10 px-3 py-2 text-sm text-rose-300">{err}</div> : null}

          <div className="flex justify-end gap-2 pt-2">
            <Button variant="outline" onClick={() => setOpenVAModal(false)} disabled={vaLoading} className="rounded-xl">
              Batal
            </Button>
            <Button onClick={submitVA} disabled={vaLoading} variant="success" className="rounded-xl">
              {vaLoading ? "Processing..." : "Buat Tiket VA"}
            </Button>
          </div>
        </div>
      </AppModal>

      <div className="rounded-2xl border border-white/10 bg-card/80 p-4 shadow-[0_16px_40px_-26px_rgba(0,0,0,0.8)]">
        {rows.length > 0 ? (
          <>
            <div className="text-lg font-semibold text-white">History Deposit</div>
            <div className="text-xs text-white/55">Tiket dan request deposit member</div>
            <div className="mt-2">
              <DataTable<DepositRow>
                columns={columns}
                rows={rows}
                rowKey={(x) => x.id}
                emptyText="Belum ada data"
                minWidthClassName="min-w-180"
                showRowNumber={false}
                wrapperClassName="mt-3 overflow-auto rounded-md border border-white/10 bg-slate-950/35"
              />
            </div>
          </>
        ) : (
          <EmptyState
            title="Belum ada data"
            description="Tiket deposit yang kamu buat akan tampil di sini."
            icon={<Landmark className="h-6 w-6" />}
          />
        )}
      </div>
    </div>
  );
}
