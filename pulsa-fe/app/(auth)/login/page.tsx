import { Suspense } from "react";
import { LoginCard } from "@/components/auth/LoginCard";
import { BackgroundAuth } from "@/components/auth/BackgroundAuth";

// Login bergantung pada parameter URL dan pemeriksaan sesi di browser.
// Jangan prerender/cache halaman ini agar Safari mobile tidak berganti antara
// fallback Suspense dan kartu login saat hydration.
export const dynamic = "force-dynamic";

export default function LoginPage() {
  return (
    <BackgroundAuth>
      <Suspense fallback={<div className="min-h-[620px] w-full" aria-hidden="true" />}>
        <LoginCard />
      </Suspense>
    </BackgroundAuth>
  );
}
