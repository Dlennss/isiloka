"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import DashboardProfileCard from "@/components/dashboard/DashboardProfileCard";
import { H2HWithdrawClient } from "@/components/dashboard/H2HWithdrawClient";
import { decodeJwt, type JwtClaims } from "@/lib/jwt";

export default function MemberWithdrawPage() {
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
      <DashboardProfileCard description="Ajukan pencairan fee H2H dan pantau status withdraw anda." />
      <H2HWithdrawClient authToken={authToken} />
    </div>
  );
}
