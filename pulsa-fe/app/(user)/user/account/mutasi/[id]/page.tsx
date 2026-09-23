import { getAppServerSession } from "@/lib/server-auth";
import { redirect } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, ArrowDownLeft, ArrowUpRight } from "lucide-react";
import type { UserSession } from "@/components/user/types";
import { UserBottomNav } from "@/components/user/UserBottomNav";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
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

function formatRupiah(value?: number | null) {
  return `Rp ${new Intl.NumberFormat("id-ID").format(Number(value || 0))}`;
}

function formatDateTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString("id-ID", {
    day: "2-digit",
    month: "long",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function formatReason(row: MutasiRow) {
  if (row.alasan === "APP_ORDER_REFUND") return "Refund transaksi aplikasi";
  if (row.alasan === "GUEST_REFUND_CLAIM") return "Klaim refund transaksi guest";
  return row.alasan || "Mutasi saldo";
}

async function getMutasiDetail(token: string, id: string): Promise<MutasiRow | null> {
  const base = (process.env.NEXT_PUBLIC_API_BASE || process.env.API_BASE || "http://127.0.0.1:8083").replace(/\/+$/, "");
  try {
    const res = await fetch(`${base}/v1/history/mutasi/${encodeURIComponent(id)}`, {
      cache: "no-store",
      headers: { Authorization: `Bearer ${token}` },
    });
    const json = (await res.json().catch(() => ({}))) as { ok?: boolean; item?: MutasiRow };
    if (!res.ok || !json.ok || !json.item) return null;
    return json.item;
  } catch {
    return null;
  }
}

type PageProps = { params: Promise<{ id: string }> };

export default async function UserAccountMutasiDetailPage({ params }: PageProps) {
  const session = (await getAppServerSession()) as SessionShape | null;
  if (!session?.backendToken) redirect("/login");

  const { id } = await params;
  const item = await getMutasiDetail(session.backendToken, id);
  if (!item) redirect("/user/account/mutasi");

  const arah = String(item.arah || "").toUpperCase();
  const isCredit = arah === "CREDIT";

  return (
    <main className="h-dvh overflow-hidden bg-sky-50 px-4 pt-5">
      <div className="mx-auto flex h-full w-full max-w-md flex-col gap-4">
        <section className="flex min-h-0 flex-1 flex-col rounded-md bg-white p-5 shadow-[0_8px_22px_rgba(15,23,42,0.13)]">
          <div className="flex items-center gap-3">
            <Link href="/user/account/mutasi" className="grid h-9 w-9 place-items-center rounded-full bg-slate-100 text-slate-700">
              <ArrowLeft className="h-4 w-4" />
            </Link>
            <div>
              <h1 className="text-lg font-bold text-neutral-900">Detail Mutasi</h1>
              <p className="text-sm text-neutral-500">Rincian perubahan saldo akun anda.</p>
            </div>
          </div>

          <div className="mt-5 flex items-start gap-3 rounded-md border border-slate-200 bg-slate-50 p-4">
            <span className={`grid h-10 w-10 shrink-0 place-items-center rounded-full ${isCredit ? "bg-emerald-100 text-emerald-700" : "bg-rose-100 text-rose-700"}`}>
              {isCredit ? <ArrowDownLeft className="h-5 w-5" /> : <ArrowUpRight className="h-5 w-5" />}
            </span>
            <div className="min-w-0">
              <p className="text-sm font-semibold text-slate-900">{formatReason(item)}</p>
              <p className={`mt-1 text-xl font-bold ${isCredit ? "text-emerald-700" : "text-rose-700"}`}>
                {isCredit ? "+" : "-"}{formatRupiah(item.jumlah)}
              </p>
              <p className="mt-1 text-xs text-slate-500">{formatDateTime(item.dibuat_pada)}</p>
            </div>
          </div>

          <div className="mt-5 flex-1 space-y-3 overflow-y-auto pb-[calc(5rem+env(safe-area-inset-bottom))]">
            {[
              ["Ref ID", item.ref_id || "-"],
              ["Arah", isCredit ? "Masuk" : "Keluar"],
              ["Saldo Sebelum", formatRupiah(item.saldo_sebelum)],
              ["Saldo Sesudah", formatRupiah(item.saldo_sesudah)],
            ].map(([label, value]) => (
              <div key={label} className="rounded-md border border-slate-200 bg-white px-4 py-3">
                <p className="text-[11px] font-medium uppercase tracking-[0.14em] text-slate-400">{label}</p>
                <p className="mt-1 break-all text-sm font-semibold text-slate-900">{value}</p>
              </div>
            ))}

            {item.catatan ? (
              <div className="rounded-md border border-slate-200 bg-white px-4 py-3">
                <p className="text-[11px] font-medium uppercase tracking-[0.14em] text-slate-400">Catatan</p>
                <p className="mt-1 whitespace-pre-wrap text-sm text-slate-700">{item.catatan}</p>
              </div>
            ) : null}
          </div>
        </section>
      </div>

      <UserBottomNav />
    </main>
  );
}
