import Link from "next/link";
import { getAppServerSession } from "@/lib/server-auth";
import { redirect } from "next/navigation";
import type { UserSession } from "@/components/user/types";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import { UserSaldoMutationHistoryList } from "@/components/user/UserSaldoMutationHistoryList";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

type SearchParams = {
  from?: string;
  to?: string;
};

type MutasiRow = {
  id: number;
  member_id: number;
  ref_id: string;
  arah: string;
  jumlah: number;
  alasan: string;
  catatan?: string | null;
  saldo_sebelum?: number | null;
  saldo_sesudah?: number | null;
  dibuat_pada: string;
};

const PAGE_SIZE = 10;

async function getMutasiFiltered(token: string, arah: string, from: string, to: string, limit: number, offset: number): Promise<MutasiRow[]> {
  const base = (process.env.NEXT_PUBLIC_API_BASE || process.env.API_BASE || "http://127.0.0.1:8083").replace(/\/+$/, "");
  try {
    const qs = new URLSearchParams({ limit: String(limit), offset: String(offset) });
    if (arah) qs.set("arah", arah);
    if (from) qs.set("from", from);
    if (to) qs.set("to", to);
    const res = await fetch(`${base}/v1/history/mutasi?${qs.toString()}`, {
      cache: "no-store",
      headers: { Authorization: `Bearer ${token}` },
    });
    const json = (await res.json().catch(() => ({}))) as { ok?: boolean; rows?: MutasiRow[] };
    if (!res.ok || !json.ok || !Array.isArray(json.rows)) return [];
    return json.rows;
  } catch {
    return [];
  }
}

export default async function UserAccountMutasiPage({ searchParams }: { searchParams?: Promise<SearchParams> }) {
  const session = (await getAppServerSession()) as SessionShape | null;

  if (!session?.backendToken) {
    redirect("/login");
  }

  const params = (await searchParams) || {};
  const today = new Intl.DateTimeFormat("en-CA", {
    timeZone: "Asia/Jakarta",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(new Date());
  const from = String(params.from || today).trim();
  const to = String(params.to || today).trim();
  const fetchedRows = await getMutasiFiltered(session.backendToken, "", from, to, PAGE_SIZE + 1, 0);
  const hasNextPage = fetchedRows.length > PAGE_SIZE;
  const rows = hasNextPage ? fetchedRows.slice(0, PAGE_SIZE) : fetchedRows;

  return (
    <main className="h-dvh overflow-hidden bg-sky-50 px-4 pt-5">
      <div className="mx-auto flex h-full w-full max-w-md flex-col gap-4">
        <section className="flex min-h-0 flex-1 flex-col">
          <h1 className="text-lg font-bold text-neutral-900">Mutasi Saldo</h1>
          <p className="mt-1 text-sm text-neutral-500">Riwayat perubahan saldo akun anda.</p>

          <form className="mt-3 grid grid-cols-2 gap-3" method="GET">
            <label className="grid gap-1">
              <span className="text-xs font-medium text-slate-500">Dari</span>
              <input
                type="date"
                name="from"
                defaultValue={from}
                className="h-10 rounded-md border border-slate-200 bg-white px-3 text-sm text-slate-900 outline-none focus:border-sky-500"
              />
            </label>
            <label className="grid gap-1">
              <span className="text-xs font-medium text-slate-500">Sampai</span>
              <input
                type="date"
                name="to"
                defaultValue={to}
                className="h-10 rounded-md border border-slate-200 bg-white px-3 text-sm text-slate-900 outline-none focus:border-sky-500"
              />
            </label>
            <div className="col-span-2 flex gap-2">
              <button
                type="submit"
                className="inline-flex h-10 flex-1 items-center justify-center rounded-md bg-[#087e8b] px-4 text-sm font-semibold text-white shadow-sm transition hover:bg-[#0a5dad]"
              >
                Terapkan
              </button>
              <Link
                href="/user/account/mutasi"
                className="inline-flex h-10 items-center justify-center rounded-md border border-slate-200 bg-white px-4 text-sm font-semibold text-slate-700 transition hover:bg-slate-50"
              >
                Reset
              </Link>
            </div>
          </form>

          <UserSaldoMutationHistoryList
            initialItems={rows}
            initialHasNextPage={hasNextPage}
            arah=""
            from={from}
            to={to}
            authToken={session.backendToken}
          />
        </section>
      </div>

      <UserBottomNav />
    </main>
  );
}
