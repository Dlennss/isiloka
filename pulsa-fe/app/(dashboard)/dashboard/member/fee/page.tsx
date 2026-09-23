"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import DashboardProfileCard from "@/components/dashboard/DashboardProfileCard";
import { H2HCommissionClient } from "@/components/dashboard/H2HCommissionClient";
import { decodeJwt, type JwtClaims } from "@/lib/jwt";

export default function MemberFeePage() {
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
    if (role !== "agent_member" && role !== "master_member") {
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
      <DashboardProfileCard description="Pantau fee H2H yang masuk dari transaksi downline anda." />
      <H2HCommissionClient authToken={authToken} />
    </div>
  );
}
