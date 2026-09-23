"use client";

import MasterSimpleCrud from "@/components/dashboard/MasterSimpleCrud";

export default function AdminMasterProviderPage() {
  return <MasterSimpleCrud title="Master Provider" endpoint="/api/admin/master/provider" emptyLabel="Belum ada provider." showKeterangan />;
}
