import type { UserBrandItem } from "@/components/user/types";

const PREFIX_ALIASES: Array<{ aliases: string[]; prefixes: string[] }> = [
  { aliases: ["by.u", "byu"], prefixes: ["085154", "085155", "085156", "085157", "085158", "08514", "08515"] },
  { aliases: ["telkomsel", "simpati", "kartu as", "kartuas", "loop", "halo"], prefixes: ["0811", "0812", "0813", "0821", "0822", "0823", "0851", "0852", "0853"] },
  { aliases: ["indosat", "im3", "mentari", "matrix"], prefixes: ["0814", "0815", "0816", "0855", "0856", "0857", "0858"] },
  { aliases: ["xl", "axis"], prefixes: ["0817", "0818", "0819", "0859", "0877", "0878"] },
  { aliases: ["axis", "xl"], prefixes: ["0831", "0832", "0833", "0837", "0838"] },
  { aliases: ["tri", "3"], prefixes: ["0895", "0896", "0897", "0898", "0899"] },
  { aliases: ["smartfren", "smart"], prefixes: ["0881", "0882", "0883", "0884", "0885", "0886", "0887", "0888", "0889"] },
];

export function normalizeOperatorDigits(value: string) {
  const digits = value.replace(/\D/g, "");
  if (digits.startsWith("62")) return `0${digits.slice(2)}`;
  return digits;
}

export function findDetectedOperatorBrand(phone: string, brands: UserBrandItem[]): UserBrandItem | null {
  const digits = normalizeOperatorDigits(phone);
  if (digits.length < 4) return null;

  const byuRule = PREFIX_ALIASES.find((rule) => rule.aliases.includes("by.u"));
  if (byuRule && digits.length >= 5 && byuRule.prefixes.some((prefix) => digits.startsWith(prefix))) {
    const found = brands.find((brand) => {
      const lower = brand.nama.trim().toLowerCase();
      return byuRule.aliases.some((alias) => lower === alias || lower.includes(alias));
    });
    if (found) return found;
  }

  const prefix4 = digits.slice(0, 4);
  for (const rule of PREFIX_ALIASES) {
    if (rule.aliases.includes("by.u")) continue;
    if (!rule.prefixes.some((prefix) => prefix.length === 4 && prefix4 === prefix)) continue;
    const found = brands.find((brand) => {
      const lower = brand.nama.trim().toLowerCase();
      return rule.aliases.some((alias) => lower === alias || lower.includes(alias));
    });
    if (found) return found;
  }

  return null;
}
