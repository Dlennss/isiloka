export function fmtID(n: number | string | null | undefined): string {
  if (n === null || n === undefined) return "0";
  const raw = typeof n === "number" ? String(n) : String(n);

  // ambil tanda minus + digit saja
  const neg = raw.trim().startsWith("-");
  const digits = raw.replace(/[^\d]/g, "");
  if (!digits) return "0";

  const out = new Intl.NumberFormat("id-ID").format(Number(digits));
  return neg ? `-${out}` : out;
}

// untuk input uang: simpan digit-only, tampilkan pretty
export function moneyInput(raw: string): { digits: string; pretty: string; value: number } {
  const digits = String(raw || "").replace(/\D+/g, "");
  const value = digits ? Number(digits) : 0;
  return { digits, pretty: digits ? fmtID(digits) : "", value };
}
