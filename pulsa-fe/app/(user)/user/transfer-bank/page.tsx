import { redirect } from "next/navigation";
import { getCategoryIdByName } from "@/lib/api.products";

export default async function CategoryEntryPage() {
  const id = await getCategoryIdByName("Bank Transfer");
  redirect(id ? `/user/kategori/${id}` : "/user/kategori");
}
