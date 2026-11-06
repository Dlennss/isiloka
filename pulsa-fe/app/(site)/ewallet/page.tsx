import { redirect } from "next/navigation";
import { getCategoryIdByName } from "@/lib/api.products";

export default async function CategoryEntryPage() {
  const id = await getCategoryIdByName("E-Wallet");
  redirect(id ? `/kategori/${id}` : "/kategori");
}
