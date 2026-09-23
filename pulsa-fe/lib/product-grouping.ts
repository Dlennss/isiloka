import type { UserProductItem } from "@/components/user/types";

function normalizeSpaces(value: string) {
  return value.replace(/\s+/g, " ").trim();
}

export function getProductGroupLabel(item: UserProductItem, fallback = "Lainnya") {
  const fromGroup = normalizeSpaces(String(item.group_name || ""));
  return fromGroup || fallback;
}
