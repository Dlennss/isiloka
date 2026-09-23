import { Suspense } from "react";
import { BackgroundAuth } from "@/components/auth/BackgroundAuth";
import GoogleCompleteClient from "@/components/auth/GoogleCompleteClient";

export default function GoogleCompletePage() {
  return (
    <BackgroundAuth>
      <main className="flex min-h-[60vh] items-center justify-center px-6 py-16">
        <Suspense
          fallback={
            <section className="w-full max-w-md rounded-2xl border border-slate-200 bg-white p-6 text-center shadow-[0_20px_48px_rgba(15,23,42,0.12)]">
              <h1 className="text-lg font-semibold text-slate-900">Login Google</h1>
              <p className="mt-2 text-sm text-slate-500">Menyelesaikan login Google...</p>
            </section>
          }
        >
          <GoogleCompleteClient />
        </Suspense>
      </main>
    </BackgroundAuth>
  );
}
