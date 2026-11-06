import type { UserCategoryItem } from "@/components/user/types";

export function getGuestCategoryPath(item: Pick<UserCategoryItem, "id" | "nama">) {
  return `/kategori/${item.id}`;
}
