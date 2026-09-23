import type { UserCategoryItem } from "@/components/user/types";

type CategoryRoute = {
  id: string;
  name: string;
  guestPath: string;
};

const CATEGORY_ROUTES: CategoryRoute[] = [
  { id: "1", name: "Pulsa", guestPath: "/pulsa" },
  { id: "2", name: "Paket Data", guestPath: "/paket-data" },
  { id: "3", name: "E-Wallet", guestPath: "/ewallet" },
  { id: "4", name: "PLN", guestPath: "/listrik" },
  { id: "5", name: "Game", guestPath: "/game" },
  { id: "7", name: "TV", guestPath: "/tv" },
  { id: "8", name: "Aktivasi Perdana", guestPath: "/aktivasi-perdana" },
  { id: "9", name: "Masa Aktif", guestPath: "/masa-aktif" },
  { id: "10", name: "Paket Telepon", guestPath: "/paket-telepon" },
  { id: "11", name: "Listrik", guestPath: "/listrik" },
  { id: "17", name: "PDAM", guestPath: "/pdam" },
  { id: "18", name: "HP Pascabayar", guestPath: "/hp-pascabayar" },
  { id: "19", name: "BPJS", guestPath: "/bpjs" },
  { id: "20", name: "Gas Negara", guestPath: "/pgn" },
];

const CATEGORY_ROUTE_BY_NAME: Record<string, string> = {
  pulsa: "/pulsa",
  "paket data": "/paket-data",
  "e-money": "/ewallet",
  "e-wallet": "/ewallet",
  pln: "/listrik",
  listrik: "/listrik",
  game: "/game",
  tv: "/tv",
  "aktivasi perdana": "/aktivasi-perdana",
  "masa aktif": "/masa-aktif",
  "paket telepon": "/paket-telepon",
  pdam: "/pdam",
  "internet pascabayar": "/internet-pascabayar",
  "hp pascabayar": "/hp-pascabayar",
  bpjs: "/bpjs",
  "gas negara": "/pgn",
};

function normalizeName(name: string) {
  return name.trim().toLowerCase();
}

export function getGuestCategoryPathById(id: string) {
  return CATEGORY_ROUTES.find((item) => item.id === String(id))?.guestPath ?? null;
}

export function getGuestCategoryPath(item: Pick<UserCategoryItem, "id" | "nama">) {
  return CATEGORY_ROUTE_BY_NAME[normalizeName(item.nama)] ?? getGuestCategoryPathById(String(item.id)) ?? `/kategori/${item.id}?name=${encodeURIComponent(item.nama)}`;
}
