"use client";

import { useEffect, useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { alertError, alertSuccess } from "@/components/ui/alerts";
import { authHeader, fmtDateTime, fmtIDR, todayJakarta } from "@/components/dashboard/auditorHelpers";

type BankRow = {
  id: number;
  nama: string;
  saldo: number;
};

type InternalFinanceRow = {
  id: number;
  ref_id: string;
  entry_type: string;
  category: string;
  direction: string;
  bank_id: number;
  bank_nama: string;
  amount: number;
  fee: number;
  total_amount: number;
  counterparty: string;
  note: string;
  occurred_at: string;
  created_nama?: string | null;
};

const ENTRY_OPTIONS = [
  ["purchase", "Pembelian"],
  ["salary", "Pembayaran Gaji"],
  ["other_expense", "Pengeluaran Lain"],
  ["other_income", "Pemasukan Lain"],
  ["bank_admin", "Biaya Admin Bank"],
  ["bank_interest", "Bunga Bank"],
  ["bank_interest_tax", "Pajak Bunga Bank"],
  ["rent_expense", "Biaya Sewa"],
  ["audit_expense", "Biaya Audit"],
  ["printing_expense", "Biaya Cetak"],
  ["event_expense", "Biaya Event"],
  ["tax_income_expense", "Biaya Pajak"],
  ["operational_expense", "Biaya Operasional"],
] as const;

export function InternalFinanceClient() {
  const [banks, setBanks] = useState<BankRow[]>([]);
  const [items, setItems] = useState<InternalFinanceRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  const [bankID, setBankID] = useState("");
  const [entryType, setEntryType] = useState("purchase");
  const [counterparty, setCounterparty] = useState("");
  const [amount, setAmount] = useState("");
  const [fee, setFee] = useState("");
  const [note, setNote] = useState("");
  const [occurredAt, setOccurredAt] = useState(todayJakarta());
  const [filterBankID, setFilterBankID] = useState("");
  const [filterType, setFilterType] = useState("");

  const selectedBank = useMemo(() => banks.find((item) => String(item.id) === bankID) || null, [banks, bankID]);
  const debitPreview = useMemo(() => Number(amount || 0) + Number(fee || 0), [amount, fee]);

  async function loadBanks() {
    const r = await fetch("/api/admin/master/bank", { headers: authHeader(), cache: "no-store" });
    const j = await r.json().catch(() => ({}));
    if (!r.ok || !j?.ok) throw new Error(j?.error || "Gagal memuat bank");
    setBanks(Array.isArray(j.items) ? j.items : []);
  }

  async function loadItems(nextBankID = filterBankID, nextType = filterType) {
    setLoading(true);
    try {
      const qs = new URLSearchParams({ limit: "100" });
      if (nextBankID) qs.set("bank_id", nextBankID);
      if (nextType) qs.set("entry_type", nextType);
      const r = await fetch(`/api/admin/wallet/internal-finance?${qs.toString()}`, { headers: authHeader(), cache: "no-store" });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) throw new Error(j?.error || "Gagal memuat transaksi internal");
      setItems(Array.isArray(j.items) ? j.items : []);
    } catch (err) {
      await alertError(err instanceof Error ? err.message : "Gagal memuat transaksi internal");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void loadBanks();
    void loadItems();
  }, []);

  async function submit() {
    if (!bankID || Number(bankID) <= 0) {
      await alertError("Pilih rekening sumber dulu.");
      return;
    }
    if (!amount || Number(amount) <= 0) {
      await alertError("Nominal wajib diisi.");
      return;
    }
    setSaving(true);
    try {
      const r = await fetch("/api/admin/wallet/internal-finance", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({
          bank_id: Number(bankID),
          entry_type: entryType,
          amount: Number(amount || 0),
          fee: Number(fee || 0),
          counterparty,
          note,
          occurred_at: occurredAt ? new Date(`${occurredAt}T12:00:00+07:00`).toISOString() : undefined,
        }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) throw new Error(j?.error || "Gagal menyimpan transaksi internal");
      await alertSuccess("Transaksi internal berhasil dicatat.");
      setAmount("");
      setFee("");
      setCounterparty("");
      setNote("");
      await loadBanks();
      await loadItems();
    } catch (err) {
      await alertError(err instanceof Error ? err.message : "Gagal menyimpan transaksi internal");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="space-y-4 p-2">
      <div>
        <div className="text-lg font-semibold tracking-tight">Pengeluaran Internal</div>
        <div className="text-sm text-muted-foreground">
          Catat pembelian, gaji, biaya operasional, atau pemasukan lain. Saldo rekening sumber akan langsung ikut terpotong atau bertambah, dan mutasinya masuk ke transaksi bank.
        </div>
      </div>

      <section className="rounded-md border border-white/15 bg-slate-950/50 p-4">
        <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
          <label className="space-y-1 text-sm">
            <div className="text-xs uppercase tracking-wide text-slate-400">Rekening Sumber</div>
            <select className="h-10 w-full rounded-md border border-white/15 bg-slate-950 px-3" value={bankID} onChange={(e) => setBankID(e.target.value)}>
              <option value="">Pilih rekening</option>
              {banks.map((bank) => (
                <option key={bank.id} value={bank.id}>
                  {bank.nama} | saldo {fmtIDR(bank.saldo)}
                </option>
              ))}
            </select>
          </label>

          <label className="space-y-1 text-sm">
            <div className="text-xs uppercase tracking-wide text-slate-400">Jenis Transaksi</div>
            <select className="h-10 w-full rounded-md border border-white/15 bg-slate-950 px-3" value={entryType} onChange={(e) => setEntryType(e.target.value)}>
              {ENTRY_OPTIONS.map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </select>
          </label>

          <label className="space-y-1 text-sm">
            <div className="text-xs uppercase tracking-wide text-slate-400">Tanggal WIB</div>
            <Input className="h-10" type="date" value={occurredAt} onChange={(e) => setOccurredAt(e.target.value)} />
          </label>

          <label className="space-y-1 text-sm">
            <div className="text-xs uppercase tracking-wide text-slate-400">Nominal</div>
            <Input className="h-10" inputMode="numeric" value={amount} onChange={(e) => setAmount(e.target.value.replace(/\D+/g, ""))} placeholder="0" />
          </label>

          <label className="space-y-1 text-sm">
            <div className="text-xs uppercase tracking-wide text-slate-400">Fee</div>
            <Input className="h-10" inputMode="numeric" value={fee} onChange={(e) => setFee(e.target.value.replace(/\D+/g, ""))} placeholder="0" />
          </label>

          <label className="space-y-1 text-sm">
            <div className="text-xs uppercase tracking-wide text-slate-400">Pihak / Tujuan</div>
            <Input className="h-10" value={counterparty} onChange={(e) => setCounterparty(e.target.value)} placeholder="Vendor / Karyawan / Keterangan pihak" />
          </label>
        </div>

        <label className="mt-3 block space-y-1 text-sm">
          <div className="text-xs uppercase tracking-wide text-slate-400">Catatan</div>
          <textarea className="min-h-24 w-full rounded-md border border-white/15 bg-slate-950 px-3 py-2" value={note} onChange={(e) => setNote(e.target.value)} placeholder="Catatan internal" />
        </label>

        <div className="mt-3 rounded-md border border-white/10 bg-slate-900/60 p-3 text-sm">
          <div>Bank sumber: <span className="font-medium text-slate-100">{selectedBank?.nama || "-"}</span></div>
          <div>Total debit/kredit bank: <span className="font-medium text-slate-100">{fmtIDR(debitPreview)}</span></div>
        </div>

        <div className="mt-3 flex justify-end">
          <Button className="h-10" onClick={() => void submit()} disabled={saving}>
            {saving ? "Menyimpan..." : "Simpan Transaksi Internal"}
          </Button>
        </div>
      </section>

      <section className="rounded-md border border-white/15 bg-slate-950/50 p-4">
        <div className="mb-3 flex flex-wrap items-end gap-2">
          <label className="space-y-1 text-sm">
            <div className="text-xs uppercase tracking-wide text-slate-400">Filter Bank</div>
            <select className="h-10 rounded-md border border-white/15 bg-slate-950 px-3" value={filterBankID} onChange={(e) => setFilterBankID(e.target.value)}>
              <option value="">Semua bank</option>
              {banks.map((bank) => (
                <option key={bank.id} value={bank.id}>{bank.nama}</option>
              ))}
            </select>
          </label>
          <label className="space-y-1 text-sm">
            <div className="text-xs uppercase tracking-wide text-slate-400">Filter Jenis</div>
            <select className="h-10 rounded-md border border-white/15 bg-slate-950 px-3" value={filterType} onChange={(e) => setFilterType(e.target.value)}>
              <option value="">Semua jenis</option>
              {ENTRY_OPTIONS.map(([value, label]) => (
                <option key={value} value={value}>{label}</option>
              ))}
            </select>
          </label>
          <Button variant="outline" className="h-10" onClick={() => void loadItems(filterBankID, filterType)} disabled={loading}>
            {loading ? "Memuat..." : "Terapkan"}
          </Button>
        </div>

        <div className="overflow-x-auto">
          <table className="min-w-full text-sm">
            <thead className="text-left text-slate-400">
              <tr className="border-b border-white/10">
                <th className="px-3 py-2">Ref</th>
                <th className="px-3 py-2">Waktu</th>
                <th className="px-3 py-2">Bank</th>
                <th className="px-3 py-2">Jenis</th>
                <th className="px-3 py-2">Pihak</th>
                <th className="px-3 py-2 text-right">Nominal</th>
                <th className="px-3 py-2 text-right">Fee</th>
                <th className="px-3 py-2 text-right">Total</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.id} className="border-b border-white/5">
                  <td className="px-3 py-2 font-mono text-xs">{item.ref_id}</td>
                  <td className="px-3 py-2">{fmtDateTime(item.occurred_at)}</td>
                  <td className="px-3 py-2">{item.bank_nama}</td>
                  <td className="px-3 py-2">{item.entry_type}</td>
                  <td className="px-3 py-2">{item.counterparty || item.note || "-"}</td>
                  <td className="px-3 py-2 text-right">{fmtIDR(item.amount)}</td>
                  <td className="px-3 py-2 text-right">{fmtIDR(item.fee)}</td>
                  <td className="px-3 py-2 text-right">{fmtIDR(item.total_amount)}</td>
                </tr>
              ))}
              {items.length === 0 ? (
                <tr>
                  <td className="px-3 py-6 text-center text-slate-400" colSpan={8}>Belum ada transaksi internal.</td>
                </tr>
              ) : null}
            </tbody>
          </table>
        </div>
      </section>
    </div>
  );
}
