"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Coins, FileText, KeyRound, Landmark, Network, ShieldCheck, Wallet } from "lucide-react";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";
import { AppModal } from "@/components/ui/app-modal";
import { Button } from "@/components/ui/button";
import { OverviewCard } from "@/components/ui/card/overview-card";
import DashboardProfileCard from "@/components/dashboard/DashboardProfileCard";
import { decodeJwt, type JwtClaims } from "@/lib/jwt";

type ProfileResp = {
  ok: boolean;
  profile?: {
    id: number;
    email: string;
    nama: string;
    aktif: boolean;
    saldo: number;
    dibuat_pada: string;
  };
  api_keys?: Array<{ id: number; api_key: string; aktif: boolean }>;
  ip_whitelist?: Array<{ id: number; ip: string; aktif: boolean }>;
  error?: string;
};

type StatsResp = {
  ok: boolean;
  rows?: StatRow[];
  overall?: OverallStat;
  error?: string;
};

type TrxRow = {
  id: number;
  member_id: number;
  ref_id: string;
  perintah: string;
  kode_produk: string;
  tujuan: string;
  qty: number;
  qty_provider?: number;
  charge_receiver_applied?: boolean;
  fee_member_rp?: number;
  status: string;
  keterangan?: string | null;
  biaya_perkiraan: number;
  biaya_aktual: number;
  dibuat_pada: string;
  diperbarui_pada: string;
};

type TrxResp = {
  ok: boolean;
  rows?: TrxRow[];
  error?: string;
};

type H2HSummaryResp = {
  ok: boolean;
  item?: {
    total_earned: number;
    total_pending_withdraw: number;
    total_approved_withdraw: number;
    total_rejected_withdraw: number;
    available_saldo: number;
  };
  error?: string;
};

type H2HDownlineResp = {
  ok: boolean;
  items?: Array<{
    id: number;
    email: string;
    nama: string;
    role: string;
  }>;
  error?: string;
};

type StatRow = {
    year: number;
    month: number;
    deposit_count: number;
    deposit_sum: number;
    trx_count: number;
    trx_sum: number;
    success_count?: number;
    success_sum?: number;
    failed_count?: number;
    failed_sum?: number;
};

type OverallStat = {
  deposit_count: number;
  deposit_sum: number;
  success_count: number;
  success_sum: number;
  failed_count: number;
  failed_sum: number;
  other_mutation_net?: number;
  ledger_balance?: number;
  saldo_reconciled?: boolean;
};

function statusLabel(status: string): string {
  const s = String(status || "").toLowerCase();
  if (s === "success") return "Berhasil";
  if (s === "failed") return "Gagal";
  if (s === "pending") return "Pending";
  return status || "-";
}

function statusTone(status: string): string {
  const s = String(status || "").toLowerCase();
  if (s === "success") return "bg-emerald-500/15 text-emerald-200 ring-1 ring-emerald-400/30";
  if (s === "failed") return "bg-rose-500/15 text-rose-200 ring-1 ring-rose-400/30";
  return "bg-amber-500/15 text-amber-200 ring-1 ring-amber-400/30";
}

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

function escapeHtml(value: string): string {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

export default function DashboardPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [role, setRole] = useState<"member" | "agent_member" | "master_member">("member");
  const [p, setP] = useState<ProfileResp | null>(null);
  const [s, setS] = useState<StatsResp | null>(null);
  const [trx, setTrx] = useState<TrxResp | null>(null);
  const [h2hSummary, setH2HSummary] = useState<H2HSummaryResp | null>(null);
  const [h2hDownlines, setH2HDownlines] = useState<H2HDownlineResp | null>(null);
  const [selectedReceipt, setSelectedReceipt] = useState<TrxRow | null>(null);
  const [err, setErr] = useState<string>("");

  useEffect(() => {
    const t = localStorage.getItem("auth_token");
    if (!t) {
      router.replace("/login");
      return;
    }

    const claims = decodeJwt(t) as JwtClaims | null;
    const rawRole = String(claims?.role || "").toLowerCase();
    setRole(rawRole === "agent_member" || rawRole === "master_member" ? rawRole : "member");

    let alive = true;

    (async () => {
      try {
        const headers = authHeader();
        const loadJson = async <T,>(url: string): Promise<T> => {
          const r = await fetch(url, { headers, cache: "no-store" });
          return r.json() as Promise<T>;
        };
        const safeLoad = async <T extends { ok: boolean; error?: string },>(url: string, fallback: T): Promise<T> => {
          try {
            return await loadJson<T>(url);
          } catch (e: unknown) {
            return { ...fallback, error: e instanceof Error ? e.message : "fetch error" };
          }
        };

        const pp = await loadJson<ProfileResp>("/api/me/profile");

        if (!alive) return;

        if (!pp?.ok) {
          setErr(pp?.error || "Gagal load profile");
          setP(pp);
          return;
        }

        setP(pp);
        setLoading(false);

        const [ss, tt, hs, hd] = await Promise.all([
          safeLoad<StatsResp>("/api/me/stats", { ok: false }),
          safeLoad<TrxResp>("/api/me/history/transaksi?limit=5", { ok: false }),
          safeLoad<H2HSummaryResp>("/api/me/h2h/commissions/summary", { ok: false }),
          safeLoad<H2HDownlineResp>("/api/me/h2h/downlines", { ok: false }),
        ]);

        if (!alive) return;

        setS(ss);
        setTrx(tt);
        setH2HSummary(hs);
        setH2HDownlines(hd);
      } catch (e: unknown) {
        if (!alive) return;
        setErr(e instanceof Error ? e.message : "fetch error");
      } finally {
        if (alive) setLoading(false);
      }
    })();

    return () => {
      alive = false;
    };
  }, [router]);

  if (loading) return <div className="text-white/70">Loading...</div>;
  if (err) return <div className="text-red-400">Error: {err}</div>;
  if (!p?.ok) return <div className="text-red-400">Error: {p?.error || "failed"}</div>;

  const saldo = Number(p.profile?.saldo || 0);
  const apiKeyAktif = (p.api_keys || []).filter((x) => x.aktif).length;
  const wlAktif = (p.ip_whitelist || []).filter((x) => x.aktif).length;
  const rows = s?.rows || [];
  const overall = s?.overall;
  const trxRows = trx?.rows || [];
  const downlines = h2hDownlines?.items || [];
  const feeSummary = h2hSummary?.item;
  const canManageNetwork = role === "agent_member" || role === "master_member";
  const statsLoaded = s !== null;
  const statsReady = !!s?.ok;
  const trxLoaded = trx !== null;
  const totalDeposit = rows.reduce((a, b) => a + Number(b.deposit_sum || 0), 0);
  const totalSuccess = rows.reduce((a, b) => a + Number((b.success_sum ?? b.trx_sum) || 0), 0);
  const totalFailed = rows.reduce((a, b) => a + Number(b.failed_sum || 0), 0);
  const overallSuccessSum = Number(overall?.success_sum || 0);
  const overallFailedSum = Number(overall?.failed_sum || 0);
  const overallDepositSum = Number(overall?.deposit_sum || 0);
  const formatRp = (value: number) => `Rp ${value.toLocaleString("id-ID")}`;
  const statsValue = (value: number) => (statsReady ? formatRp(value) : "Memuat...");
  const columns: DataTableColumn<StatRow>[] = [
    {
      id: "bulan",
      header: "Bulan",
      render: (r) => (
        <span className="text-white/85">
          {String(r.month).padStart(2, "0")}/{r.year}
        </span>
      ),
    },
    { id: "deposit_count", header: "Deposit", render: (r) => r.deposit_count },
    {
      id: "deposit_sum",
      header: "Deposit (Rp)",
      render: (r) => `Rp ${Number(r.deposit_sum || 0).toLocaleString("id-ID")}`,
    },
    { id: "success_count", header: "Berhasil", render: (r) => Number((r.success_count ?? r.trx_count) || 0) },
    {
      id: "success_sum",
      header: "Berhasil (Rp)",
      render: (r) => `Rp ${Number((r.success_sum ?? r.trx_sum) || 0).toLocaleString("id-ID")}`,
    },
    { id: "failed_count", header: "Gagal", render: (r) => Number(r.failed_count || 0) },
    {
      id: "failed_sum",
      header: "Gagal (Rp)",
      render: (r) => `Rp ${Number(r.failed_sum || 0).toLocaleString("id-ID")}`,
    },
  ];
  const trxColumns: DataTableColumn<TrxRow>[] = [
    {
      id: "waktu",
      header: "Waktu",
      render: (r) => new Date(r.dibuat_pada).toLocaleString("id-ID"),
    },
    {
      id: "produk",
      header: "Produk",
      render: (r) => (
        <div className="min-w-36">
          <div className="font-medium text-white/90">{r.kode_produk}</div>
          <div className="text-xs text-white/50">{r.perintah}</div>
        </div>
      ),
    },
    {
      id: "tujuan",
      header: "Tujuan",
      render: (r) => r.tujuan,
    },
    {
      id: "status",
      header: "Status",
      render: (r) => (
        <span className={`inline-flex rounded-full px-2 py-1 text-xs font-medium ${statusTone(r.status)}`}>
          {statusLabel(r.status)}
        </span>
      ),
    },
    {
      id: "qty",
      header: "Nominal",
      render: (r) => `Rp ${Number(r.qty || 0).toLocaleString("id-ID")}`,
    },
    {
      id: "biaya_aktual",
      header: "Biaya Aktual",
      render: (r) => `Rp ${Number(r.biaya_aktual || r.biaya_perkiraan || 0).toLocaleString("id-ID")}`,
    },
    {
      id: "ref_id",
      header: "Ref ID",
      render: (r) => <span className="font-mono text-xs text-cyan-200">{r.ref_id}</span>,
    },
    {
      id: "aksi",
      header: "Aksi",
      render: (r) => (
        <Button variant="outline" className="h-8 px-3" onClick={() => setSelectedReceipt(r)}>
          <FileText className="h-4 w-4" />
          Struk
        </Button>
      ),
    },
  ];

  function printReceipt(row: TrxRow) {
    const biayaAktual = Number(row.biaya_aktual || row.biaya_perkiraan || 0);
    const qty = Number(row.qty || 0);
    const printedAt = new Date(row.dibuat_pada).toLocaleString("id-ID");
    const html = `<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>Struk ${escapeHtml(row.ref_id)}</title>
    <style>
      @page { size: 58mm auto; margin: 2mm; }
      * { box-sizing: border-box; }
      body {
        margin: 0;
        background: #fff;
        color: #000;
        font-family: "Courier New", Courier, monospace;
        font-size: 12px;
        line-height: 1.45;
      }
      .receipt {
        width: 54mm;
        max-width: 54mm;
        margin: 0;
        padding: 2mm 0 4mm;
      }
      .title { font-size: 16px; font-weight: 700; text-align: center; }
      .muted { color: #222; font-size: 11px; text-align: center; }
      .line {
        border-top: 1px dashed #000;
        margin: 8px 0;
      }
      .row {
        display: flex;
        gap: 8px;
        justify-content: space-between;
        align-items: flex-start;
        margin: 3px 0;
      }
      .label { width: 28mm; }
      .value {
        flex: 1;
        text-align: right;
        word-break: break-word;
      }
      .value.left { text-align: left; }
      .footer { margin-top: 10px; text-align: center; font-size: 11px; }
    </style>
  </head>
  <body>
    <div class="receipt">
      <div class="title">Bukti Transaksi</div>
      <div class="muted">Transaksi sukses</div>
      <div class="line"></div>
      <div class="row"><div class="label">Waktu</div><div class="value">${escapeHtml(printedAt)}</div></div>
      <div class="row"><div class="label">Ref ID</div><div class="value">${escapeHtml(row.ref_id)}</div></div>
      <div class="row"><div class="label">Perintah</div><div class="value">${escapeHtml(row.perintah)}</div></div>
      <div class="row"><div class="label">Produk</div><div class="value">${escapeHtml(row.kode_produk)}</div></div>
      <div class="row"><div class="label">Tujuan</div><div class="value">${escapeHtml(row.tujuan)}</div></div>
      <div class="row"><div class="label">Qty</div><div class="value">Rp ${qty.toLocaleString("id-ID")}</div></div>
      <div class="row"><div class="label">Biaya Aktual</div><div class="value">Rp ${biayaAktual.toLocaleString("id-ID")}</div></div>
      <div class="row"><div class="label">Status</div><div class="value">${escapeHtml(String(row.status).toUpperCase())}</div></div>
      <div class="line"></div>
      <div class="row"><div class="label">Keterangan / SN</div></div>
      <div class="row"><div class="value left">${escapeHtml(row.keterangan || "-")}</div></div>
      <div class="line"></div>
      <div class="footer">Terima kasih</div>
    </div>
    <script>
      window.onload = function() {
        window.print();
        setTimeout(function() { window.close(); }, 250);
      };
    </script>
  </body>
</html>`;

    const win = window.open("", "_blank", "width=420,height=720");
    if (!win) return;
    win.document.open();
    win.document.write(html);
    win.document.close();
  }

  return (
    <div className="space-y-6 p-2">
      <DashboardProfileCard
        name={p.profile?.nama}
        role="member"
        description="Akun member aktif untuk transaksi API, monitoring saldo, keamanan akun, dan histori transaksi."
      />

      <div className="grid gap-3 md:grid-cols-3">
        <OverviewCard
          title="Saldo"
          value={`Rp ${saldo.toLocaleString("id-ID")}`}
          sub="Saldo tersedia saat ini"
          icon={<Wallet size={16} />}
          tone="emerald"
        />
        <OverviewCard
          title="API Key Aktif"
          value={`${apiKeyAktif}`}
          sub="Untuk transaksi via API"
          icon={<KeyRound size={16} />}
          tone="sky"
        />
        <OverviewCard
          title="IP Whitelist Aktif"
          value={`${wlAktif}`}
          sub="Batasi akses transaksi"
          icon={<ShieldCheck size={16} />}
          tone="amber"
        />
      </div>

      <div className="grid gap-3 md:grid-cols-3">
        <OverviewCard
          title="Total Deposit"
          value={statsValue(overallDepositSum)}
          icon={<Wallet size={16} />}
          tone="sky"
        />
        <OverviewCard
          title="Transaksi Berhasil"
          value={statsValue(overallSuccessSum)}
          icon={<FileText size={16} />}
          tone="emerald"
        />
        <OverviewCard
          title="Transaksi Gagal"
          value={statsValue(overallFailedSum)}
          icon={<FileText size={16} />}
          tone="rose"
        />
      </div>

      <div className={`grid gap-3 ${canManageNetwork ? "md:grid-cols-3" : "md:grid-cols-1"}`}>
        {canManageNetwork ? (
          <OverviewCard
            title="Downline H2H"
            value={`${downlines.length}`}
            sub="Agent/member di bawah akun anda"
            icon={<Network size={16} />}
            tone="sky"
          />
        ) : null}
        {canManageNetwork ? (
          <OverviewCard
            title="Fee H2H"
            value={`Rp ${Number(feeSummary?.total_earned || 0).toLocaleString("id-ID")}`}
            sub="Akumulasi fee yang sudah masuk"
            icon={<Coins size={16} />}
            tone="amber"
          />
        ) : null}
        {canManageNetwork ? (
          <OverviewCard
            title="Withdraw H2H"
            value={`Rp ${Number(feeSummary?.total_pending_withdraw || 0).toLocaleString("id-ID")}`}
            sub="Nominal withdraw yang masih pending"
            icon={<Landmark size={16} />}
            tone="amber"
          />
        ) : null}
      </div>

      {canManageNetwork ? (
        <div className="overflow-hidden rounded-2xl border border-white/10 bg-card/80 shadow-[0_16px_40px_-24px_rgba(0,0,0,0.8)]">
          <div className="flex flex-wrap items-center justify-between gap-3 border-b border-white/10 bg-linear-to-r from-white/10 via-white/5 to-transparent px-4 py-3">
            <div>
              <div className="text-lg font-semibold">Jaringan H2H</div>
              <div className="text-xs text-white/55">Ringkasan downline dan shortcut untuk fee serta withdraw H2H.</div>
            </div>
            <div className="flex flex-wrap gap-2">
              <Button variant="outline" className="h-9" onClick={() => router.push("/dashboard/member/downline")}>
                Downline
              </Button>
              <Button variant="outline" className="h-9" onClick={() => router.push("/dashboard/member/fee")}>
                Fee
              </Button>
              <Button variant="outline" className="h-9" onClick={() => router.push("/dashboard/member/withdraw")}>
                Withdraw
              </Button>
            </div>
          </div>
          <div className="space-y-3 px-4 py-4">
            {downlines.length === 0 ? <div className="text-sm text-white/60">Belum ada downline H2H.</div> : null}
            {downlines.slice(0, 5).map((item) => (
              <div key={item.id} className="rounded-xl border border-white/10 bg-slate-950/35 px-4 py-3">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <div className="font-medium text-white/90">{item.nama || "-"}</div>
                    <div className="text-xs text-white/55">{item.email}</div>
                  </div>
                  <div className="rounded-full border border-cyan-400/20 bg-cyan-400/10 px-2 py-1 text-xs text-cyan-200">
                    {item.role}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      ) : null}

      <div className="overflow-hidden rounded-2xl border border-white/10 bg-card/80 shadow-[0_16px_40px_-24px_rgba(0,0,0,0.8)]">
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-white/10 bg-linear-to-r from-white/10 via-white/5 to-transparent px-4 py-3">
          <div>
            <div className="text-lg font-semibold">Statistik 3 Bulan</div>
            <div className="text-xs text-white/55">Deposit dan transaksi berhasil/gagal per bulan. Ini bukan saldo berjalan.</div>
          </div>
          <div className="flex gap-2 text-xs">
            <div className="rounded-lg border border-emerald-400/20 bg-emerald-400/10 px-2 py-1 text-emerald-200">
              Deposit: {statsReady ? formatRp(totalDeposit) : "Memuat..."}
            </div>
            <div className="rounded-lg border border-sky-400/20 bg-sky-400/10 px-2 py-1 text-sky-200">
              Berhasil: {statsReady ? formatRp(totalSuccess) : "Memuat..."}
            </div>
            <div className="rounded-lg border border-rose-400/20 bg-rose-400/10 px-2 py-1 text-rose-200">
              Gagal: {statsReady ? formatRp(totalFailed) : "Memuat..."}
            </div>
          </div>
        </div>

        {!statsLoaded ? (
          <div className="px-4 py-4 text-sm text-white/60">Memuat statistik...</div>
        ) : !s?.ok ? (
          <div className="px-4 py-4 text-sm text-red-400">Error: {s?.error || "failed"}</div>
        ) : (
          <div className="px-4 pb-4">
            <DataTable<StatRow>
              columns={columns}
              rows={rows}
              rowKey={(r) => `${r.year}-${r.month}`}
              emptyText="Belum ada data"
              minWidthClassName="min-w-[980px]"
              showRowNumber={false}
              wrapperClassName="mt-4 overflow-auto rounded-md border border-white/10 bg-slate-950/35"
            />
          </div>
        )}
      </div>

      <div className="overflow-hidden rounded-2xl border border-white/10 bg-card/80 shadow-[0_16px_40px_-24px_rgba(0,0,0,0.8)]">
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-white/10 bg-linear-to-r from-white/10 via-white/5 to-transparent px-4 py-3">
          <div>
            <div className="text-lg font-semibold">Transaksi Terakhir</div>
            <div className="text-xs text-white/55">Menampilkan status berhasil, gagal, dan pending agar tidak ambigu.</div>
          </div>
          <Button variant="outline" className="h-9" onClick={() => router.push("/dashboard/member/history/transaksi")}>
            Lihat Semua
          </Button>
        </div>

        {!trxLoaded ? (
          <div className="px-4 py-4 text-sm text-white/60">Memuat transaksi terakhir...</div>
        ) : !trx?.ok ? (
          <div className="px-4 py-4 text-sm text-red-400">Error: {trx?.error || "failed"}</div>
        ) : (
          <div className="px-4 pb-4">
            <DataTable<TrxRow>
              columns={trxColumns}
              rows={trxRows}
              rowKey={(r) => r.id}
              emptyText="Belum ada transaksi."
              minWidthClassName="min-w-[980px]"
              showRowNumber={false}
              wrapperClassName="mt-4 overflow-auto rounded-md border border-white/10 bg-slate-950/35"
            />
          </div>
        )}
      </div>

      <AppModal
        open={!!selectedReceipt}
        onClose={() => setSelectedReceipt(null)}
        title="Struk Transaksi"
        subtitle="Format ringkas yang siap dicetak ke kertas thermal."
      >
        {selectedReceipt ? (
          <div className="space-y-4 text-sm text-slate-200">
            <div className="mx-auto w-full max-w-[320px] rounded-2xl border border-dashed border-white/15 bg-white px-4 py-5 font-mono text-black shadow-[0_20px_50px_-25px_rgba(0,0,0,0.8)]">
              <div className="text-center">
                <div className="text-lg font-bold tracking-wide"> RESI TRANSAKSI</div>
                <div className="text-[11px] uppercase tracking-[0.2em] text-slate-500">Struk Transaksi Sukses</div>
              </div>

              <div className="my-3 border-t border-dashed border-slate-400" />

              <div className="space-y-1.5 text-[12px] leading-5">
                <div className="flex items-start justify-between gap-3">
                  <span>Waktu</span>
                  <span className="max-w-40 text-right">{new Date(selectedReceipt.dibuat_pada).toLocaleString("id-ID")}</span>
                </div>
                <div className="flex items-start justify-between gap-3">
                  <span>Ref ID</span>
                  <span className="max-w-40 break-all text-right">{selectedReceipt.ref_id}</span>
                </div>
                <div className="flex items-start justify-between gap-3">
                  <span>Perintah</span>
                  <span className="text-right">{selectedReceipt.perintah}</span>
                </div>
                <div className="flex items-start justify-between gap-3">
                  <span>Produk</span>
                  <span className="text-right">{selectedReceipt.kode_produk}</span>
                </div>
                <div className="flex items-start justify-between gap-3">
                  <span>Tujuan</span>
                  <span className="text-right">{selectedReceipt.tujuan}</span>
                </div>
                <div className="flex items-start justify-between gap-3">
                  <span>Qty</span>
                  <span className="text-right">Rp {Number(selectedReceipt.qty || 0).toLocaleString("id-ID")}</span>
                </div>
                <div className="flex items-start justify-between gap-3">
                  <span>Biaya Aktual</span>
                  <span className="text-right">Rp {Number(selectedReceipt.biaya_aktual || selectedReceipt.biaya_perkiraan || 0).toLocaleString("id-ID")}</span>
                </div>
                <div className="flex items-start justify-between gap-3">
                  <span>Status</span>
                  <span className="text-right font-bold uppercase">{selectedReceipt.status}</span>
                </div>
              </div>

              <div className="my-3 border-t border-dashed border-slate-400" />

              <div className="text-[11px] uppercase tracking-wide text-slate-500">Keterangan / SN</div>
              <div className="mt-1 wrap-break-word text-[12px] leading-5">{selectedReceipt.keterangan || "-"}</div>

              <div className="my-3 border-t border-dashed border-slate-400" />
              <div className="text-center text-[11px] text-slate-500">Terima kasih</div>
            </div>

            <div className="flex justify-center">
              <Button variant="primary" className="h-10 px-4" onClick={() => printReceipt(selectedReceipt)}>
                <FileText className="h-4 w-4" />
                Cetak
              </Button>
            </div>
          </div>
        ) : null}
      </AppModal>

    </div>
  );
}
