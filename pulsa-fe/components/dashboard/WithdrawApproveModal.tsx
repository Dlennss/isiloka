"use client";

import { useEffect, useMemo, useState } from "react";
import { AppModal } from "@/components/ui/app-modal";
import { alertError } from "@/components/ui/alerts";

type BankRow = {
  id: number;
  nama: string;
  nomor_rekening: string;
  atas_nama: string;
  saldo: number;
  aktif: boolean;
};

type SubmitPayload = {
  bank_id: number;
  fee: number;
  note: string;
};

type WithdrawApproveModalProps = {
  open: boolean;
  title: string;
  amount: number;
  sourceAccountNumber?: string;
  sourceAccountLabel?: string;
  submitting?: boolean;
  banksEndpoint?: string;
  onClose: () => void;
  onSubmit: (payload: SubmitPayload) => Promise<void>;
};

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

function fmtIDR(v: number) {
  return new Intl.NumberFormat("id-ID").format(Number(v || 0));
}

function digitsOnly(v: string) {
  return String(v || "").replace(/\D/g, "");
}

export default function WithdrawApproveModal({
  open,
  title,
  amount,
  sourceAccountNumber = "",
  sourceAccountLabel = "",
  submitting = false,
  banksEndpoint = "/api/admin/master/bank",
  onClose,
  onSubmit,
}: WithdrawApproveModalProps) {
  const [banks, setBanks] = useState<BankRow[]>([]);
  const [loadingBanks, setLoadingBanks] = useState(false);
  const [bankID, setBankID] = useState("");
  const [fee, setFee] = useState("");
  const [note, setNote] = useState("");

  const sourceAccountDigits = useMemo(() => digitsOnly(sourceAccountNumber), [sourceAccountNumber]);
  const activeBanks = useMemo(() => {
    const items = banks.filter((item) => item.aktif);
    if (!sourceAccountDigits) return items;
    return items.filter((item) => digitsOnly(item.nomor_rekening) === sourceAccountDigits);
  }, [banks, sourceAccountDigits]);
  const feeValue = Number(fee || "0");
  const totalDebit = amount + (Number.isFinite(feeValue) ? feeValue : 0);
  const selectedBank = activeBanks.find((item) => item.id === Number(bankID));

  useEffect(() => {
    if (!open) return;
    setFee("");
    setNote("");
    setBankID("");
  }, [open]);

  useEffect(() => {
    if (!open) return;
    let cancelled = false;

    async function loadBanks() {
      setLoadingBanks(true);
      try {
        const r = await fetch(banksEndpoint, {
          headers: authHeader(),
          cache: "no-store",
        });
        const j = await r.json().catch(() => ({}));
        if (!r.ok || !j?.ok) {
          throw new Error(j?.error || "Gagal memuat rekening sumber.");
        }
        const items = Array.isArray(j?.items) ? j.items : [];
        if (cancelled) return;
        setBanks(items);
        const sourceItems = sourceAccountDigits
          ? items.filter((item: BankRow) => digitsOnly(item.nomor_rekening) === sourceAccountDigits)
          : items;
        const firstActive = sourceItems.find((item: BankRow) => item.aktif);
        setBankID(firstActive ? String(firstActive.id) : "");
        if (sourceAccountDigits && !firstActive) {
          await alertError(`Rekening sumber ${sourceAccountLabel || sourceAccountNumber} tidak aktif atau belum terdaftar.`);
        }
      } catch (err) {
        if (!cancelled) {
          await alertError(err instanceof Error ? err.message : "Gagal memuat rekening sumber.");
        }
      } finally {
        if (!cancelled) setLoadingBanks(false);
      }
    }

    void loadBanks();
    return () => {
      cancelled = true;
    };
  }, [banksEndpoint, open, sourceAccountDigits, sourceAccountLabel, sourceAccountNumber]);

  async function handleSubmit() {
    const parsedBankID = Number(bankID || "0");
    const parsedFee = Number(fee || "0");
    if (!parsedBankID) {
      await alertError("Pilih rekening sumber terlebih dulu.");
      return;
    }
    if (!Number.isFinite(parsedFee) || parsedFee < 0) {
      await alertError("Fee harus angka 0 atau lebih.");
      return;
    }
    await onSubmit({
      bank_id: parsedBankID,
      fee: parsedFee,
      note: note.trim(),
    });
  }

  return (
    <AppModal
      open={open}
      onClose={() => {
        if (!submitting) onClose();
      }}
      title={title}
      subtitle={sourceAccountLabel ? `Rekening sumber ${sourceAccountLabel}. Isi fee jika ada, lalu approve withdraw.` : "Pilih rekening sumber, isi fee jika ada, lalu approve withdraw."}
      maxWidthClassName="max-w-2xl"
      footer={
        <div className="flex flex-col gap-2 sm:flex-row sm:justify-end">
          <button
            type="button"
            className="h-11 rounded-xl border border-white/10 bg-white/5 px-4 text-sm font-semibold text-white disabled:cursor-not-allowed disabled:opacity-60"
            onClick={onClose}
            disabled={submitting}
          >
            Batal
          </button>
          <button
            type="button"
            className="h-11 rounded-xl bg-emerald-600 px-4 text-sm font-semibold text-white disabled:cursor-not-allowed disabled:opacity-60"
            onClick={() => void handleSubmit()}
            disabled={submitting || loadingBanks || activeBanks.length === 0}
          >
            {submitting ? "Memproses..." : "Approve Withdraw"}
          </button>
        </div>
      }
    >
      <div className="space-y-4">
        <div className="grid gap-4 md:grid-cols-2">
          <label className="space-y-2">
            <span className="text-sm font-medium text-slate-100">Rekening sumber</span>
            <select
              className="h-11 w-full rounded-xl border border-white/10 bg-slate-950/60 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400"
              value={bankID}
              onChange={(e) => setBankID(e.target.value)}
              disabled={loadingBanks || submitting || Boolean(sourceAccountDigits)}
            >
              <option value="">{loadingBanks ? "Memuat rekening..." : sourceAccountDigits ? "Rekening sumber tidak tersedia" : "Pilih rekening sumber"}</option>
              {activeBanks.map((bank) => (
                <option key={bank.id} value={bank.id}>
                  {bank.nama} • {bank.nomor_rekening} • Saldo Rp {fmtIDR(bank.saldo)}
                </option>
              ))}
            </select>
          </label>

          <label className="space-y-2">
            <span className="text-sm font-medium text-slate-100">Fee transfer</span>
            <input
              className="h-11 w-full rounded-xl border border-white/10 bg-slate-950/60 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400"
              inputMode="numeric"
              placeholder="0"
              value={fee}
              onChange={(e) => setFee(e.target.value.replace(/[^\d]/g, ""))}
              disabled={submitting}
            />
          </label>
        </div>

        <label className="block space-y-2">
          <span className="text-sm font-medium text-slate-100">Catatan admin</span>
          <textarea
            className="min-h-[96px] w-full rounded-xl border border-white/10 bg-slate-950/60 px-3 py-3 text-sm text-slate-100 outline-none focus:border-cyan-400"
            placeholder="Opsional. Misalnya waktu transfer, nama rekening pengirim, atau catatan lain."
            value={note}
            onChange={(e) => setNote(e.target.value)}
            disabled={submitting}
          />
        </label>

        <div className="rounded-2xl border border-white/10 bg-slate-950/50 p-4">
          <div className="grid gap-3 text-sm text-slate-300 sm:grid-cols-2">
            <div>
              <div className="text-xs uppercase tracking-wide text-slate-500">Nominal withdraw</div>
              <div className="mt-1 font-semibold text-white">Rp {fmtIDR(amount)}</div>
            </div>
            <div>
              <div className="text-xs uppercase tracking-wide text-slate-500">Fee</div>
              <div className="mt-1 font-semibold text-white">Rp {fmtIDR(feeValue)}</div>
            </div>
            <div className="sm:col-span-2">
              <div className="text-xs uppercase tracking-wide text-slate-500">Total debit rekening sumber</div>
              <div className="mt-1 text-base font-bold text-emerald-400">Rp {fmtIDR(totalDebit)}</div>
            </div>
            <div className="sm:col-span-2">
              <div className="text-xs uppercase tracking-wide text-slate-500">Rekening terpilih</div>
              <div className="mt-1 text-sm text-white">
                {selectedBank ? `${selectedBank.nama} • ${selectedBank.nomor_rekening} • ${selectedBank.atas_nama}` : "-"}
              </div>
            </div>
          </div>
        </div>
      </div>
    </AppModal>
  );
}
