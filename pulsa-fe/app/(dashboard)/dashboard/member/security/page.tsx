"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { Copy, KeyRound, LockKeyhole, ShieldCheck, ShieldEllipsis } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";
import { AppModal } from "@/components/ui/app-modal";
import { alertConfirm, alertError, alertSuccess, alertWarning } from "@/components/ui/alerts";

type IPRow = {
  id: number;
  ip: string;
  label: string;
  webhook_url?: string;
  aktif: boolean;
  dibuat_pada: string;
};

type ApiKeyRow = { id: number; api_key: string; aktif: boolean; dibuat_pada?: string };

type ProfileResponse = {
  ok: boolean;
  profile?: {
    charge_receiver?: boolean;
  };
  api_keys?: ApiKeyRow[];
  error?: string;
};

type ModalKind = "api" | "password" | "pin" | "ip";

async function api(path: string, init?: RequestInit) {
  const t = localStorage.getItem("auth_token") || "";
  const headers: Record<string, string> = {};
  if (t) headers.Authorization = `Bearer ${t}`;
  if (init?.method && init.method !== "GET") headers["Content-Type"] = "application/json";

  const res = await fetch(path, { ...init, headers: { ...headers, ...(init?.headers || {}) }, cache: "no-store" });
  const data = await res.json().catch(() => ({}));
  return { ok: res.ok, status: res.status, data };
}

function maskKey(k: string) {
  const s = (k || "").trim();
  if (!s) return "";
  if (s.length <= 12) return "********";
  return `${s.slice(0, 6)}...${s.slice(-6)}`;
}

async function copyToClipboard(text: string) {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    return false;
  }
}

function isValidHttpUrl(s: string) {
  const v = (s || "").trim();
  if (!v) return false;
  try {
    const u = new URL(v);
    return u.protocol === "http:" || u.protocol === "https:";
  } catch {
    return false;
  }
}

function SettingActionCard({
  icon,
  title,
  description,
  actionLabel,
  onClick,
  tone = "cyan",
}: {
  icon: React.ReactNode;
  title: string;
  description: string;
  actionLabel: string;
  onClick: () => void;
  tone?: "cyan" | "emerald" | "amber";
}) {
  const toneClass =
    tone === "emerald"
      ? "border-emerald-400/20 bg-emerald-400/10 text-emerald-200"
      : tone === "amber"
        ? "border-amber-400/20 bg-amber-400/10 text-amber-200"
        : "border-cyan-400/20 bg-cyan-400/10 text-cyan-200";

  return (
    <div className="rounded-2xl border border-white/10 bg-card/80 p-4 shadow-[0_16px_40px_-26px_rgba(0,0,0,0.8)]">
      <div className="flex items-start justify-between gap-4">
        <div className={`grid h-11 w-11 shrink-0 place-items-center rounded-2xl border ${toneClass}`}>{icon}</div>
        <div className="flex-1 space-y-1">
          <div className="text-sm font-semibold text-white">{title}</div>
          <div className="text-xs leading-5 text-white/60">{description}</div>
        </div>
      </div>
      <div className="mt-4">
        <Button onClick={onClick} variant="outline" className="w-full rounded-xl border-white/10 bg-white/[0.04] text-white hover:bg-white/10">
          {actionLabel}
        </Button>
      </div>
    </div>
  );
}

export default function SecurityPage() {
  const [chargeReceiver, setChargeReceiver] = useState(false);
  const [initialChargeReceiver, setInitialChargeReceiver] = useState(false);
  const [savingChargeReceiver, setSavingChargeReceiver] = useState(false);

  const [oldPass, setOldPass] = useState("");
  const [newPass, setNewPass] = useState("");
  const [savingPassword, setSavingPassword] = useState(false);

  const [oldPin, setOldPin] = useState("");
  const [newPin, setNewPin] = useState("");
  const [savingPIN, setSavingPIN] = useState(false);

  const [apiKeyNew, setApiKeyNew] = useState<string>("");
  const [apiKeys, setApiKeys] = useState<ApiKeyRow[]>([]);
  const [resettingApiKey, setResettingApiKey] = useState(false);

  const [ip, setIp] = useState("");
  const [label, setLabel] = useState("");
  const [webhookUrl, setWebhookUrl] = useState("");
  const [ips, setIps] = useState<IPRow[]>([]);
  const [savingIP, setSavingIP] = useState(false);
  const [activeModal, setActiveModal] = useState<ModalKind | null>(null);

  const loadProfile = useCallback(async () => {
    const r = await api("/api/me/profile");
    const d = r.data as ProfileResponse;
    if (r.ok && d?.ok) {
      setApiKeys(d.api_keys || []);
      const currentChargeReceiver = Boolean(d.profile?.charge_receiver);
      setChargeReceiver(currentChargeReceiver);
      setInitialChargeReceiver(currentChargeReceiver);
    }
  }, []);

  const loadIPs = useCallback(async () => {
    const r = await api("/api/me/ip-whitelist");
    if (r.ok && r.data?.ok) setIps(r.data.rows || []);
  }, []);

  useEffect(() => {
    const t = setTimeout(() => {
      void loadProfile();
      void loadIPs();
    }, 0);
    return () => clearTimeout(t);
  }, [loadProfile, loadIPs]);

  const changePassword = async () => {
    if (savingPassword) return;
    setSavingPassword(true);
    try {
      const r = await api("/api/me/password", {
        method: "POST",
        body: JSON.stringify({ old_password: oldPass, new_password: newPass }),
      });
      if (!r.ok || !r.data?.ok) return alertError(r.data?.error || "Gagal update password");
      setOldPass("");
      setNewPass("");
      setActiveModal(null);
      await alertSuccess("Password berhasil diupdate.");
    } finally {
      setSavingPassword(false);
    }
  };

  const changePIN = async () => {
    if (savingPIN) return;
    setSavingPIN(true);
    try {
      const r = await api("/api/me/pin", {
        method: "POST",
        body: JSON.stringify({ old_pin: oldPin, new_pin: newPin }),
      });
      if (!r.ok || !r.data?.ok) return alertError(r.data?.error || "Gagal update PIN");
      setOldPin("");
      setNewPin("");
      setActiveModal(null);
      await alertSuccess("PIN berhasil diupdate.");
    } finally {
      setSavingPIN(false);
    }
  };

  const resetApiKey = async () => {
    if (resettingApiKey) return;
    const ok = await alertConfirm({
      title: "Reset API Key?",
      text: "API key lama akan diganti dan tidak bisa dipakai lagi.",
      confirmButtonText: "Ya, reset",
    });
    if (!ok) return;

    setResettingApiKey(true);
    try {
      const r = await api("/api/me/api-key/reset", { method: "POST" });
      if (!r.ok || !r.data?.ok) return alertError(r.data?.error || "Gagal reset API key");
      setApiKeyNew(r.data.api_key || "");
      await alertSuccess("API key berhasil direset. Simpan key baru sekarang.");
      await loadProfile();
    } finally {
      setResettingApiKey(false);
    }
  };

  const saveChargeReceiver = async () => {
    if (savingChargeReceiver) return;
    setSavingChargeReceiver(true);
    try {
      const r = await api("/api/me/charge-receiver", {
        method: "PUT",
        body: JSON.stringify({ charge_receiver: chargeReceiver }),
      });
      if (!r.ok || !r.data?.ok) return alertError(r.data?.error || "Gagal menyimpan pengaturan charge receiver");
      setInitialChargeReceiver(Boolean(r.data?.charge_receiver));
      setChargeReceiver(Boolean(r.data?.charge_receiver));
      await alertSuccess("Pengaturan charge receiver berhasil diperbarui.");
    } finally {
      setSavingChargeReceiver(false);
    }
  };

  const addIP = async () => {
    if (savingIP) return;
    const ipVal = (ip || "").trim();
    const labelVal = (label || "").trim();
    const wh = (webhookUrl || "").trim();

    if (!ipVal) return alertWarning("IP wajib diisi");
    if (!wh) return alertWarning("Webhook URL wajib diisi");
    if (!isValidHttpUrl(wh)) return alertWarning("Webhook URL tidak valid (harus http/https)");

    setSavingIP(true);
    try {
      const r = await api("/api/me/ip-whitelist", {
        method: "POST",
        body: JSON.stringify({ ip: ipVal, label: labelVal, webhook_url: wh }),
      });
      if (!r.ok || !r.data?.ok) return alertError(r.data?.error || "Gagal tambah IP whitelist");

      setIp("");
      setLabel("");
      setWebhookUrl("");
      setActiveModal(null);
      await loadIPs();
      await alertSuccess("IP whitelist berhasil ditambahkan.");
    } finally {
      setSavingIP(false);
    }
  };

  const delIP = async (id: number) => {
    const ok = await alertConfirm({
      title: "Hapus IP whitelist?",
      text: "Data IP whitelist akan dihapus permanen.",
      confirmButtonText: "Ya, hapus",
    });
    if (!ok) return;
    const r = await api(`/api/me/ip-whitelist?id=${encodeURIComponent(String(id))}`, { method: "DELETE" });
    if (!r.ok || !r.data?.ok) return alertError(r.data?.error || "Gagal hapus IP whitelist");
    await loadIPs();
    await alertSuccess("IP whitelist berhasil dihapus.");
  };

  const activeKey = (apiKeys || []).find((k) => k.aktif)?.api_key || "";
  const columns: DataTableColumn<IPRow>[] = useMemo(
    () => [
      { id: "id", header: "ID", render: (x) => x.id },
      { id: "ip", header: "IP", render: (x) => <span className="font-mono">{x.ip}</span> },
      { id: "label", header: "Label", render: (x) => x.label || "-" },
      {
        id: "webhook",
        header: "Webhook",
        render: (x) => <span className="block whitespace-normal break-all">{x.webhook_url || "-"}</span>,
      },
      {
        id: "aktif",
        header: "Aktif",
        render: (x) =>
          x.aktif ? (
            <span className="rounded-md border border-emerald-400/30 bg-emerald-400/10 px-2 py-1 text-xs text-emerald-200">yes</span>
          ) : (
            <span className="rounded-md border border-slate-400/30 bg-slate-400/10 px-2 py-1 text-xs text-slate-200">no</span>
          ),
      },
    ],
    []
  );

  return (
    <div className="space-y-6 p-2">
      <div>
        <div className="bg-linear-to-r from-white via-white/90 to-white/60 bg-clip-text text-xl font-semibold text-transparent">Pengaturan</div>
        <div className="text-sm text-white/60">Kelola akses API, keamanan akun, charge receiver, dan whitelist IP.</div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <div className="rounded-2xl border border-white/10 bg-card/80 p-4 shadow-[0_16px_40px_-26px_rgba(0,0,0,0.8)]">
          <div className="flex items-start justify-between gap-4">
            <div className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl border border-sky-400/20 bg-sky-400/10 text-sky-200">
              <ShieldCheck className="h-5 w-5" />
            </div>
            <div className="flex-1 space-y-1">
              <div className="text-sm font-semibold text-white">Charge Receiver</div>
              <div className="text-xs leading-5 text-white/60">Jika diaktifkan, beban fee akan dikenakan kepada penerima.</div>
            </div>
          </div>
          <div className="mt-4 flex flex-wrap items-center gap-3">
            <label className="inline-flex cursor-pointer items-center gap-3 rounded-xl border border-white/10 bg-white/[0.04] px-3 py-2 text-sm text-white/85">
              <button
                type="button"
                role="switch"
                aria-checked={chargeReceiver}
                onClick={() => {
                  if (savingChargeReceiver) return;
                  setChargeReceiver((v) => !v);
                }}
                className={`relative h-6 w-11 rounded-full transition ${chargeReceiver ? "bg-cyan-500" : "bg-slate-700"} ${
                  savingChargeReceiver ? "cursor-not-allowed opacity-60" : ""
                }`}
                disabled={savingChargeReceiver}
            >
              <span className={`absolute top-0.5 h-5 w-5 rounded-full bg-white transition ${chargeReceiver ? "left-[22px]" : "left-0.5"}`} />
            </button>
          </label>
            <Button onClick={saveChargeReceiver} variant="info" className="rounded-xl font-semibold" disabled={savingChargeReceiver || chargeReceiver === initialChargeReceiver}>
              {savingChargeReceiver ? "Menyimpan..." : "Save"}
            </Button>
          </div>
        </div>

        <SettingActionCard
          icon={<KeyRound className="h-5 w-5" />}
          title="API Key"
          description="Lihat API key aktif, copy key, dan reset bila terindikasi bocor."
          actionLabel={resettingApiKey ? "Memproses..." : "Kelola API Key"}
          onClick={() => setActiveModal("api")}
          tone="cyan"
        />
        <SettingActionCard
          icon={<ShieldEllipsis className="h-5 w-5" />}
          title="Password"
          description="Perbarui password akun member tanpa memenuhi halaman dashboard."
          actionLabel="Ganti Password"
          onClick={() => setActiveModal("password")}
          tone="emerald"
        />
        <SettingActionCard
          icon={<LockKeyhole className="h-5 w-5" />}
          title="PIN"
          description="Ubah PIN transaksi yang dipakai untuk autentikasi request H2H."
          actionLabel="Ganti PIN"
          onClick={() => setActiveModal("pin")}
          tone="amber"
        />
        <SettingActionCard
          icon={<ShieldCheck className="h-5 w-5" />}
          title="IP Whitelist"
          description="Kelola IP yang diizinkan beserta webhook callback URL untuk akses transaksi API."
          actionLabel="Kelola IP Whitelist"
          onClick={() => setActiveModal("ip")}
          tone="cyan"
        />
      </div>

      <section className="rounded-2xl border border-cyan-400/15 bg-linear-to-br from-cyan-500/10 via-slate-950/50 to-slate-950/70 p-5 shadow-[0_16px_40px_-26px_rgba(0,0,0,0.8)]">
        <div className="space-y-2">
          <div className="text-lg font-semibold text-white">Cara Transaksi H2H</div>
          <div className="text-sm leading-6 text-white/70">
            Setelah pengaturan di atas selesai, ikuti urutan ini agar aplikasi Anda bisa langsung terhubung ke API H2H
            Isiloka dengan benar: daftarkan IP, pasang webhook callback, simpan API key, siapkan PIN, lalu kirim transaksi.
          </div>
        </div>

        <div className="mt-4 grid gap-4 xl:grid-cols-2">
          <div className="rounded-2xl border border-white/10 bg-black/20 p-4">
            <div className="text-sm font-semibold text-cyan-200">1. Tambahkan IP whitelist</div>
            <div className="mt-2 text-sm leading-6 text-white/70">
              Server Anda wajib memakai IP publik yang didaftarkan di menu <span className="font-semibold text-white">IP Whitelist</span>.
              Jika IP server belum terdaftar, transaksi H2H akan ditolak.
            </div>
            <div className="mt-3 text-xs leading-6 text-white/55">
              Isi:
              <br />• <span className="text-white">IP</span>: IP publik server/aplikasi Anda
              <br />• <span className="text-white">Label</span>: nama bebas agar mudah dikenali
              <br />• <span className="text-white">Webhook URL</span>: URL callback yang menerima status transaksi
            </div>
          </div>

          <div className="rounded-2xl border border-white/10 bg-black/20 p-4">
            <div className="text-sm font-semibold text-cyan-200">2. Siapkan webhook callback</div>
            <div className="mt-2 text-sm leading-6 text-white/70">
              Setelah transaksi dikirim, hasil final akan dikirim ke webhook URL yang Anda daftarkan.
              Endpoint ini harus bisa menerima <span className="font-semibold text-white">HTTP POST</span> dari server kami.
            </div>
            <div className="mt-3 text-xs leading-6 text-white/55">
              Simpan minimal data ini dari callback:
              <br />• <span className="text-white">refid</span>: nomor referensi transaksi Anda
              <br />• <span className="text-white">status</span>: sukses / gagal / pending
              <br />• <span className="text-white">sn</span> atau keterangan final transaksi
            </div>
          </div>

          <div className="rounded-2xl border border-white/10 bg-black/20 p-4">
            <div className="text-sm font-semibold text-cyan-200">3. Simpan API key aktif</div>
            <div className="mt-2 text-sm leading-6 text-white/70">
              API key dipakai sebagai identitas akun H2H Anda. Ambil dari menu <span className="font-semibold text-white">API Key</span>,
              lalu simpan di server Anda. Jangan ditaruh di frontend publik atau aplikasi client yang bisa dibaca orang lain.
            </div>
            <div className="mt-3 text-xs leading-6 text-white/55">
              Header yang dipakai:
              <br />• <span className="font-mono text-white">X-Api-Key: API_KEY_AKTIF_ANDA</span>
              <br />• <span className="font-mono text-white">Content-Type: application/json</span>
            </div>
          </div>

          <div className="rounded-2xl border border-white/10 bg-black/20 p-4">
            <div className="text-sm font-semibold text-cyan-200">4. Siapkan PIN transaksi</div>
            <div className="mt-2 text-sm leading-6 text-white/70">
              PIN dipakai untuk autentikasi request transaksi H2H. Jadi walaupun API key benar, transaksi tetap harus memakai PIN yang valid.
            </div>
            <div className="mt-3 text-xs leading-6 text-white/55">
              Parameter request:
              <br />• <span className="font-mono text-white">pin=PIN_AKUN_ANDA</span>
            </div>
          </div>
        </div>

        <div className="mt-4 rounded-2xl border border-white/10 bg-black/25 p-4">
          <div className="text-sm font-semibold text-cyan-200">5. Kirim request transaksi</div>
          <div className="mt-2 text-sm leading-6 text-white/70">
            Gunakan <span className="font-semibold text-white">kode produk</span> dari menu Produk H2H.
            Setiap transaksi wajib punya <span className="font-semibold text-white">refid unik</span> dari sistem Anda sendiri.
          </div>

          <div className="mt-4 rounded-2xl border border-white/10 bg-black/20 p-4">
            <div className="text-sm font-semibold text-cyan-200">Daftar Command</div>
            <div className="mt-2 text-sm leading-6 text-white/70">
              Semua request dikirim ke <span className="font-mono text-white">POST https://api.pulsakilat.net/v1/trx</span> dengan
              header <span className="font-mono text-white">X-Api-Key</span> dan <span className="font-mono text-white">Content-Type: application/json</span>.
            </div>

            <div className="mt-4 overflow-auto">
              <table className="w-full text-xs text-white/80">
                <thead>
                  <tr className="border-b border-white/10 text-left text-white/50">
                    <th className="pb-2 pr-4 font-semibold">Command</th>
                    <th className="pb-2 pr-4 font-semibold">Fungsi</th>
                    <th className="pb-2 pr-4 font-semibold">Parameter Wajib</th>
                    <th className="pb-2 font-semibold">Keterangan</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/5">
                  <tr>
                    <td className="py-3 pr-4 font-mono font-semibold text-emerald-300">SALDO</td>
                    <td className="py-3 pr-4">Cek saldo akun</td>
                    <td className="py-3 pr-4 font-mono">pin</td>
                    <td className="py-3">Tidak perlu product, dest, qty, atau refid. Cukup kirim commands dan pin.</td>
                  </tr>
                  <tr>
                    <td className="py-3 pr-4 font-mono font-semibold text-sky-300">PAY</td>
                    <td className="py-3 pr-4">Kirim transaksi pembayaran</td>
                    <td className="py-3 pr-4 font-mono">product, dest, qty, refid, pin</td>
                    <td className="py-3">Transaksi utama. Saldo akan terpotong sesuai nominal + fee. Hasil final dikirim ke webhook.</td>
                  </tr>
                  <tr>
                    <td className="py-3 pr-4 font-mono font-semibold text-amber-300">INQ</td>
                    <td className="py-3 pr-4">Inquiry / cek tagihan</td>
                    <td className="py-3 pr-4 font-mono">product, dest, refid, pin</td>
                    <td className="py-3">Untuk produk postpaid (PLN tagihan, BPJS, PDAM, dll). Cek nominal tagihan sebelum bayar. qty otomatis 1.</td>
                  </tr>
                  <tr>
                    <td className="py-3 pr-4 font-mono font-semibold text-violet-300">STATUS-PAY</td>
                    <td className="py-3 pr-4">Cek status transaksi</td>
                    <td className="py-3 pr-4 font-mono">product, dest, qty, refid, pin</td>
                    <td className="py-3">Cek status transaksi yang sudah dikirim sebelumnya. Jika transaksi masih pending, sistem akan otomatis retry atau fallback ke provider lain.</td>
                  </tr>
                  <tr>
                    <td className="py-3 pr-4 font-mono font-semibold text-pink-300">PRODUK</td>
                    <td className="py-3 pr-4">Lihat daftar produk + harga</td>
                    <td className="py-3 pr-4 font-mono">pin</td>
                    <td className="py-3">Opsional: tambahkan product untuk filter produk tertentu. Menampilkan kode produk, nama, harga, dan status aktif.</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div className="mt-4 rounded-2xl border border-white/10 bg-black/20 p-4">
            <div className="text-sm font-semibold text-cyan-200">Contoh Request per Command</div>

            <div className="mt-4 grid gap-4 xl:grid-cols-2">
              <div>
                <div className="mb-2 text-xs font-semibold uppercase tracking-wide text-emerald-300/70">SALDO — Cek Saldo</div>
                <pre className="overflow-x-auto rounded-xl border border-white/10 bg-slate-950/60 p-4 text-xs leading-6 text-white/80">{`curl -X POST &apos;https://api.pulsakilat.net/v1/trx&apos; \
  -H &apos;X-Api-Key: API_KEY_ANDA&apos; \
  -H &apos;Content-Type: application/json&apos; \
  -d &apos;{"commands":"SALDO","pin":"1234"}&apos;`}</pre>
                <pre className="mt-2 overflow-x-auto rounded-xl border border-emerald-400/20 bg-emerald-500/10 p-3 text-xs leading-6 text-emerald-100/90">{`{
  "ok": true,
  "command": "SALDO",
  "balance": 5000000
}`}</pre>
              </div>

              <div>
                <div className="mb-2 text-xs font-semibold uppercase tracking-wide text-pink-300/70">PRODUK — Lihat Daftar Produk</div>
                <pre className="overflow-x-auto rounded-xl border border-white/10 bg-slate-950/60 p-4 text-xs leading-6 text-white/80">{`curl -X POST &apos;https://api.pulsakilat.net/v1/trx&apos; \
  -H &apos;X-Api-Key: API_KEY_ANDA&apos; \
  -H &apos;Content-Type: application/json&apos; \
  -d &apos;{"commands":"PRODUK","pin":"1234"}&apos;`}</pre>
              </div>
            </div>
          </div>

          <div className="mt-4 grid gap-4 xl:grid-cols-2">
            <div>
              <div className="mb-2 text-xs font-semibold uppercase tracking-wide text-white/45">Contoh Request Transaksi (PAY)</div>
              <pre className="overflow-x-auto rounded-xl border border-white/10 bg-slate-950/60 p-4 text-xs leading-6 text-white/80">{`curl -X POST 'https://api.pulsakilat.net/v1/trx' \\
  -H 'X-Api-Key: API_KEY_AKTIF_ANDA' \\
  -H 'Content-Type: application/json' \\
  -d '{
  "commands": "PAY",
  "product": "DANA",
  "dest": "081234567890",
  "qty": 10000,
  "refid": "INV123456789",
  "pin": "1234"
}'`}</pre>
            </div>

            <div>
              <div className="mb-2 text-xs font-semibold uppercase tracking-wide text-white/45">Arti Parameter</div>
              <div className="rounded-xl border border-white/10 bg-slate-950/60 p-4 text-xs leading-6 text-white/75">
                • <span className="text-white font-semibold">commands</span>: perintah transaksi (SALDO, PAY, INQ, STATUS-PAY, PRODUK)
                <br />• <span className="text-white font-semibold">product</span>: kode produk dari menu Produk H2H
                <br />• <span className="text-white font-semibold">dest</span>: tujuan transaksi (nomor HP, ID pelanggan)
                <br />• <span className="text-white font-semibold">qty</span>: nominal / quantity (angka tanpa titik)
                <br />• <span className="text-white font-semibold">refid</span>: kode unik dari sistem Anda
                <br />• <span className="text-white font-semibold">pin</span>: PIN akun H2H Anda
              </div>
            </div>
          </div>

          <div className="mt-4 grid gap-4 xl:grid-cols-3">
            <div className="rounded-xl border border-sky-400/20 bg-sky-500/10 p-4">
              <div className="mb-2 text-xs font-semibold uppercase tracking-wide text-sky-200">Contoh Balasan: Diproses</div>
              <pre className="overflow-x-auto text-xs leading-6 text-sky-100/90">{`{
  "ok": true,
  "refid": "INV123456789",
  "status": "pending",
  "msg": "Transaksi sedang diproses"
}`}</pre>
            </div>

            <div className="rounded-xl border border-emerald-400/20 bg-emerald-500/10 p-4">
              <div className="mb-2 text-xs font-semibold uppercase tracking-wide text-emerald-200">Contoh Balasan: Berhasil</div>
              <pre className="overflow-x-auto text-xs leading-6 text-emerald-100/90">{`{
  "ok": true,
  "refid": "INV123456789",
  "status": "success",
  "sn": "2026040910121481030100166446398744601",
  "msg": "Transaksi berhasil"
}`}</pre>
            </div>

            <div className="rounded-xl border border-rose-400/20 bg-rose-500/10 p-4">
              <div className="mb-2 text-xs font-semibold uppercase tracking-wide text-rose-200">Contoh Balasan: Gagal</div>
              <pre className="overflow-x-auto text-xs leading-6 text-rose-100/90">{`{
  "ok": false,
  "refid": "INV123456789",
  "status": "failed",
  "msg": "Nomor tujuan salah"
}`}</pre>
            </div>
          </div>

          <div className="mt-4 grid gap-4 xl:grid-cols-2">
            <div className="rounded-xl border border-white/10 bg-slate-950/60 p-4 text-xs leading-6 text-white/75">
              <div className="mb-2 text-xs font-semibold uppercase tracking-wide text-white/45">Alur Paling Aman</div>
              1. Ambil kode produk dari menu Produk H2H.
              <br />2. Pastikan IP server Anda sudah masuk whitelist.
              <br />3. Pastikan webhook URL aktif dan bisa menerima POST.
              <br />4. Kirim transaksi dengan API key aktif + PIN yang benar.
              <br />5. Simpan <span className="text-white font-semibold">refid</span> di sistem Anda.
              <br />6. Tunggu callback final ke webhook Anda.
            </div>

            <div className="rounded-xl border border-amber-400/20 bg-amber-500/10 p-4 text-xs leading-6 text-amber-100/90">
              <div className="mb-2 text-xs font-semibold uppercase tracking-wide text-amber-200">Catatan Penting</div>
              • Jangan memakai API key di frontend/browser publik.
              <br />• Jangan kirim transaksi dari IP yang belum didaftarkan.
              <br />• Gunakan <span className="text-white font-semibold">refid berbeda</span> untuk tiap transaksi baru.
              <br />• Untuk produk bebas nominal, harga yang tampil di menu berarti <span className="text-white font-semibold">nominal + fee member</span>.
            </div>
          </div>
        </div>
      </section>

      <AppModal
        open={activeModal === "api"}
        onClose={() => {
          if (resettingApiKey) return;
          setActiveModal(null);
        }}
        title="Kelola API Key"
        subtitle="Lihat API key aktif, salin key, atau reset bila terindikasi bocor."
        maxWidthClassName="max-w-3xl"
        footer={
          <div className="flex justify-end gap-2">
            <Button
              variant="outline"
              className="rounded-xl border-white/10 bg-white/[0.04] text-white hover:bg-white/10"
              onClick={() => setActiveModal(null)}
              disabled={resettingApiKey}
            >
              Tutup
            </Button>
            <Button variant="primary" className="rounded-xl font-semibold" onClick={() => void resetApiKey()} disabled={resettingApiKey}>
              {resettingApiKey ? "Memproses..." : "Reset API Key"}
            </Button>
          </div>
        }
      >
        <div className="grid gap-3 md:grid-cols-2">
          <div className="rounded-xl border border-white/10 bg-black/20 p-3">
            <div className="text-xs text-white/60">API Key Aktif (masked)</div>
            <div className="mt-1 flex items-center justify-between gap-3">
              <div className="font-mono text-sm break-all">{activeKey ? maskKey(activeKey) : "-"}</div>
              <Button
                variant="outline"
                className="rounded-lg border-white/15 bg-transparent text-white/85 hover:bg-white/10 hover:text-white"
                onClick={async () => {
                  if (!activeKey) return;
                  const ok = await copyToClipboard(activeKey);
                  if (ok) await alertSuccess("API key berhasil disalin.");
                  else await alertError("Gagal copy API key.");
                }}
                disabled={!activeKey || resettingApiKey}
              >
                <Copy className="h-4 w-4" />
                Copy
              </Button>
            </div>
          </div>

          <div className="rounded-xl border border-white/10 bg-black/20 p-3">
            <div className="text-xs text-white/60">API Key Baru (tampil sekali)</div>
            <div className="mt-1 flex items-center justify-between gap-3">
              <div className="font-mono text-sm break-all">{apiKeyNew ? apiKeyNew : "-"}</div>
              <Button
                variant="outline"
                className="rounded-lg border-white/15 bg-transparent text-white/85 hover:bg-white/10 hover:text-white"
                onClick={async () => {
                  if (!apiKeyNew) return;
                  const ok = await copyToClipboard(apiKeyNew);
                  if (ok) await alertSuccess("API key baru berhasil disalin.");
                  else await alertError("Gagal copy API key baru.");
                }}
                disabled={!apiKeyNew || resettingApiKey}
              >
                <Copy className="h-4 w-4" />
                Copy
              </Button>
            </div>
          </div>
        </div>
      </AppModal>

      <AppModal
        open={activeModal === "password"}
        onClose={() => {
          if (savingPassword) return;
          setActiveModal(null);
        }}
        title="Ganti Password"
        subtitle="Masukkan password lama dan password baru akun member."
        maxWidthClassName="max-w-xl"
        footer={
          <div className="flex justify-end gap-2">
            <Button variant="outline" className="rounded-xl border-white/10 bg-white/[0.04] text-white hover:bg-white/10" onClick={() => setActiveModal(null)} disabled={savingPassword}>
              Batal
            </Button>
            <Button variant="success" className="rounded-xl font-semibold" onClick={changePassword} disabled={savingPassword}>
              {savingPassword ? "Menyimpan..." : "Update Password"}
            </Button>
          </div>
        }
      >
        <div className="space-y-3">
          <Input value={oldPass} onChange={(e) => setOldPass(e.target.value)} placeholder="old password" type="password" disabled={savingPassword} />
          <Input value={newPass} onChange={(e) => setNewPass(e.target.value)} placeholder="new password" type="password" disabled={savingPassword} />
        </div>
      </AppModal>

      <AppModal
        open={activeModal === "pin"}
        onClose={() => {
          if (savingPIN) return;
          setActiveModal(null);
        }}
        title="Ganti PIN"
        subtitle="Masukkan PIN lama dan PIN baru untuk autentikasi transaksi."
        maxWidthClassName="max-w-xl"
        footer={
          <div className="flex justify-end gap-2">
            <Button variant="outline" className="rounded-xl border-white/10 bg-white/[0.04] text-white hover:bg-white/10" onClick={() => setActiveModal(null)} disabled={savingPIN}>
              Batal
            </Button>
            <Button variant="success" className="rounded-xl font-semibold" onClick={changePIN} disabled={savingPIN}>
              {savingPIN ? "Menyimpan..." : "Update PIN"}
            </Button>
          </div>
        }
      >
        <div className="space-y-3">
          <Input value={oldPin} onChange={(e) => setOldPin(e.target.value)} placeholder="old pin" type="password" disabled={savingPIN} />
          <Input value={newPin} onChange={(e) => setNewPin(e.target.value)} placeholder="new pin" type="password" disabled={savingPIN} />
        </div>
      </AppModal>

      <AppModal
        open={activeModal === "ip"}
        onClose={() => {
          if (savingIP) return;
          setActiveModal(null);
        }}
        title="Tambah IP Whitelist"
        subtitle="Daftarkan IP dan webhook URL yang diizinkan mengakses transaksi API."
        maxWidthClassName="max-w-5xl"
        footer={
          <div className="flex justify-end gap-2">
            <Button variant="outline" className="rounded-xl border-white/10 bg-white/[0.04] text-white hover:bg-white/10" onClick={() => setActiveModal(null)} disabled={savingIP}>
              Batal
            </Button>
            <Button variant="info" className="rounded-xl font-semibold" onClick={addIP} disabled={savingIP}>
              {savingIP ? "Menyimpan..." : "Simpan IP"}
            </Button>
          </div>
        }
      >
        <div className="space-y-4">
          <div className="grid gap-3 md:grid-cols-3">
            <Input value={ip} onChange={(e) => setIp(e.target.value)} placeholder="1.2.3.4" disabled={savingIP} />
            <Input value={label} onChange={(e) => setLabel(e.target.value)} placeholder="label (optional)" disabled={savingIP} />
            <Input value={webhookUrl} onChange={(e) => setWebhookUrl(e.target.value)} placeholder="webhook callback URL (wajib)" disabled={savingIP} />
          </div>

          <DataTable<IPRow>
            columns={columns}
            rows={ips}
            rowKey={(x) => x.id}
            emptyText="Belum ada IP whitelist"
            minWidthClassName="min-w-170"
            showRowNumber={false}
            wrapperClassName="overflow-auto rounded-md border border-white/10 bg-slate-950/35"
            actions={{
              header: "Aksi",
              align: "right",
              render: (x) => (
                <Button variant="danger" className="rounded-lg font-semibold" onClick={() => delIP(x.id)}>
                  Delete
                </Button>
              ),
            }}
          />
        </div>
      </AppModal>
    </div>
  );
}
