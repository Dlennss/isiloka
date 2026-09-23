/**
 * Guest Transaction Local Storage
 * Menyimpan riwayat transaksi guest di localStorage browser,
 * mirip dengan GuestTransactionStorage di aplikasi mobile.
 */

const STORAGE_KEY = 'guest_transaction_history_v1';
const MAX_ENTRIES = 40;

export type GuestTransactionEntry = {
  invoice_id: string;
  guest_email: string;
  guest_phone: string;
  dest: string;
  title: string;
  created_at: string;
  updated_at: string;
  status: string;
  amount: number;
  serial_number: string;
};

function readEntries(): GuestTransactionEntry[] {
  if (typeof window === 'undefined') return [];
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed.filter((e: unknown): e is GuestTransactionEntry => {
      if (!e || typeof e !== 'object') return false;
      const entry = e as Partial<GuestTransactionEntry>;
      return typeof entry.invoice_id === 'string' && Boolean(entry.invoice_id.trim());
    });
  } catch {
    return [];
  }
}

function persistEntries(entries: GuestTransactionEntry[]) {
  if (typeof window === 'undefined') return;
  try {
    const sorted = [...entries].sort((a, b) =>
      (b.updated_at || b.created_at).localeCompare(a.updated_at || a.created_at)
    );
    localStorage.setItem(STORAGE_KEY, JSON.stringify(sorted.slice(0, MAX_ENTRIES)));
  } catch {
    // Ignore storage errors
  }
}

export function loadGuestTransactions(): GuestTransactionEntry[] {
  return readEntries();
}

export function saveGuestTransaction(entry: GuestTransactionEntry) {
  const entries = readEntries();
  const idx = entries.findIndex((e) => e.invoice_id === entry.invoice_id);
  if (idx >= 0) {
    const current = entries[idx];
    entries[idx] = {
      ...current,
      guest_email: entry.guest_email || current.guest_email,
      guest_phone: entry.guest_phone || current.guest_phone,
      dest: entry.dest || current.dest,
      title: entry.title || current.title,
      created_at: current.created_at || entry.created_at,
      updated_at: entry.updated_at || current.updated_at,
      status: entry.status || current.status,
      amount: entry.amount > 0 ? entry.amount : current.amount,
      serial_number: entry.serial_number || current.serial_number,
    };
  } else {
    entries.push(entry);
  }
  persistEntries(entries);
}

export function removeGuestTransaction(invoiceId: string) {
  const entries = readEntries().filter((e) => e.invoice_id !== invoiceId);
  persistEntries(entries);
}

export function clearGuestTransactions() {
  if (typeof window === 'undefined') return;
  localStorage.removeItem(STORAGE_KEY);
}
