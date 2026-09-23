"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import DashboardProfileCard from "@/components/dashboard/DashboardProfileCard";
import { H2HProductClient } from "@/components/dashboard/H2HProductClient";
import { decodeJwt, type JwtClaims } from "@/lib/jwt";

export default function MemberProdukPage() {
  const router = useRouter();
  const [authToken, setAuthToken] = useState("");
  const [allowed, setAllowed] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem("auth_token") || "";
    if (!token) {
      router.replace("/login");
      return;
    }
    const claims = decodeJwt(token) as JwtClaims | null;
    const role = String(claims?.role || "").toLowerCase();
    if (role !== "member" && role !== "agent_member" && role !== "master_member") {
      router.replace("/dashboard/member");
      return;
    }
    queueMicrotask(() => {
      setAuthToken(token);
      setAllowed(true);
    });
  }, [router]);

  if (!authToken || !allowed) return <div className="text-white/70">Loading...</div>;

  return (
    <div className="space-y-6">
      <DashboardProfileCard description="Lihat daftar SKU internal H2H beserta harga jual member yang sudah memuat fee kategori akun anda." />
      <H2HProductClient authToken={authToken} />
    </div>
  );
}
