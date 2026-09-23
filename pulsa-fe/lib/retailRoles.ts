import type { UserProductItem } from "@/components/user/types";

export type RetailRole = "guest" | "user" | "agent" | "master";

export function normalizeRetailRole(role?: string | null): RetailRole {
  const value = String(role || "").trim().toLowerCase();
  if (value === "master") return "master";
  if (value === "agent") return "agent";
  if (value === "user") return "user";
  return "guest";
}

export function getRetailFeeForProduct(item: UserProductItem, role?: string | null) {
  switch (normalizeRetailRole(role)) {
    case "master":
      return Number(item.fee_master || 0);
    case "agent":
      return Number(item.fee_agent || 0);
    case "user":
      return Number(item.fee_user || 0);
    default:
      return Number(item.fee_guest || 0);
  }
}

