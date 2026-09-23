import { Bolt, Gamepad2, Grid2x2, Signal, Wallet, Wifi } from "lucide-react";
import type { UserMenuItem } from "@/components/user/types";

export type UserMenuKey = "pulsa" | "paket_data" | "listrik_prabayar" | "e_wallet" | "gaming" | "semua";

export type UserMenuCatalogItem = UserMenuItem & {
  key: UserMenuKey;
};

export const USER_MENU_ITEMS: UserMenuCatalogItem[] = [
  { key: "pulsa", title: "Pulsa", icon: Signal, href: "#" },
  { key: "paket_data", title: "Paket Data", icon: Wifi, href: "#" },
  { key: "listrik_prabayar", title: "Listrik Prabayar", icon: Bolt, href: "#" },
  { key: "e_wallet", title: "E-Wallet", icon: Wallet, href: "#" },
  { key: "gaming", title: "Gaming", icon: Gamepad2, href: "#" },
  { key: "semua", title: "Semua", icon: Grid2x2, href: "#" },
];

