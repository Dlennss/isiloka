"use client";

import { useEffect, useMemo, useState } from "react";
import { usePathname } from "next/navigation";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableActions, type DataTableColumn } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { AppModal } from "@/components/ui/app-modal";
import { ProductSearchSelect } from "@/components/dashboard/ProductSearchSelect";
import { alertConfirm, alertError, alertSuccess, alertWarning } from "@/components/ui/alerts";
import { Eye, Pencil, Plus, Power, RefreshCw, Search, Trash2 } from "lucide-react";
import { fetchAllAdminProducts } from "@/lib/adminProducts";
import { decodeJwt } from "@/lib/jwt";

type ProdukRow = {
  id: number;
  sku: string;
  nama: string;
  group_name?: string;
  kategori_id?: number;
  brand_id?: number;
  tipe_harga?: "FIXED" | "OPEN_AMOUNT";
  nominal?: number | null;
  maksimal_nominal?: number | null;
  aktif: boolean;
};

type ProviderRow = {
  id: number;
  nama: string;
  aktif: boolean;
};

type MapRow = {
  id: number;
  produk_id: number;
  produk_sku: string;
  produk_nama: string;
  tipe_harga?: "FIXED" | "OPEN_AMOUNT";
  nominal?: number | null;
  minimal_nominal?: number | null;
  maksimal_nominal?: number | null;
  provider: string;
  kode_provider: string;
  special_code?: string | null;
  mode?: string | null;
  fee_rp: number;
  aktif: boolean;
  jam_buka?: string | null;
  jam_tutup?: string | null;
};

type CombinedRow = {
  key: string;
  produk_id: number;
  produk_sku: string;
  produk_nama: string;
  tipe_harga: "FIXED" | "OPEN_AMOUNT";
  nominal: number | null;
  minimal_nominal: number | null;
  maksimal_nominal: number | null;
  provider_id: number | null;
  provider_nama: string;
  map_id: number | null;
  kode_provider: string;
  special_code: string;
  mode: string;
  map_aktif: boolean;
  fee_rp: number;
  jam_buka: string;
  jam_tutup: string;
};

type FormState = {
  produk_id: string;
  provider_id: string;
  tipe_harga: "FIXED" | "OPEN_AMOUNT";
  minimal_nominal: string;
  maksimal_nominal: string;
  kode_provider: string;
  special_code: string;
  mode: string;
  fee_rp: string;
  map_aktif: boolean;
  jam_buka: string;
  jam_tutup: string;
};

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

function normalizeProviderName(value: string | null | undefined) {
  return String(value || "").trim().toLowerCase();
}

function formatRupiah(value: number) {
  return `Rp ${new Intl.NumberFormat("id-ID").format(value || 0)}`;
}

function toNumOrNull(v: string): number | null {
  const n = Number(String(v || "").replace(/[^\d-]/g, ""));
  if (!Number.isFinite(n) || n <= 0) return null;
  return Math.floor(n);
}

function buildMapRowKey(item: {
  produk_id: number;
  provider: string;
  kode_provider?: string;
  minimal_nominal?: number | null;
  maksimal_nominal?: number | null;
  id?: number | null;
}) {
  if (item.id) return `map:${item.id}`;
  return [
    "map",
    item.produk_id,
    normalizeProviderName(item.provider),
    String(item.kode_provider || "").trim().toUpperCase(),
    String(item.minimal_nominal ?? 0),
    String(item.maksimal_nominal ?? 0),
  ].join(":");
}

function minionsFeeByCode(code: string) {
  switch (String(code || "").trim().toUpperCase()) {
    case "HDANA":
      return 55;
    case "HDANAB":
      return 100;
    case "HGOPAY":
      return 975;
    case "HLINK":
      return 625;
    case "HLINKB":
      return 700;
    case "HOVO":
      return 580;
    case "HSHOPEE":
      return 300;
    case "HSHOPEEB":
      return 500;
    default:
      return 0;
  }
}

function usesCodeBasedFee(providerName: string | null | undefined) {
  return normalizeProviderName(providerName) === "minions";
}

const EWALLET_SKUS = ["GOPAY","GOJEK","GPAY","DANA","DNID","OVO","LINKAJA","LAJA","SHOPEE","SHOPEEPAY","SHPAY"];
const BANK_SKUS = ["BCA","BRI","BNI","MANDIRI","BSI","CIMB","DANAMON","PERMATA","BTN","OCBC","MAYBANK","PANIN","MEGA","BUKOPIN","SINARMAS","COMMONWEALTH","UOB","HSBC","BTPN","MUAMALAT","JAGO","NEOCOMMERCE","SEABANK","ALLOBANK","BJB"];

function classifyProduct(sku: string): "ewallet" | "bank" | "lainnya" {
  const up = sku.trim().toUpperCase();
  if (EWALLET_SKUS.includes(up)) return "ewallet";
  if (BANK_SKUS.includes(up) || up.startsWith("BANK")) return "bank";
  return "lainnya";
}

export default function AdminProviderProductConfigPage() {
  const [produk, setProduk] = useState<ProdukRow[]>([]);
  const [providers, setProviders] = useState<ProviderRow[]>([]);
  const [maps, setMaps] = useState<MapRow[]>([]);
  const [searchInput, setSearchInput] = useState("");
  const [search, setSearch] = useState("");
  const [tab, setTab] = useState<"semua" | "ewallet" | "bank" | "lainnya">("semua");
  const [loading, setLoading] = useState(false);
  const [produkLoading, setProdukLoading] = useState(false);

  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<CombinedRow | null>(null);
  const [saving, setSaving] = useState(false);
  const [openActionKey, setOpenActionKey] = useState<string | null>(null);
  const [currentRole, setCurrentRole] = useState("");
  const pathname = usePathname();
  const [form, setForm] = useState<FormState>({
    produk_id: "",
    provider_id: "",
    tipe_harga: "FIXED",
    minimal_nominal: "",
    maksimal_nominal: "",
    kode_provider: "",
    special_code: "",
    mode: "",
    fee_rp: "0",
    map_aktif: true,
    jam_buka: "",
    jam_tutup: "",
  });

  const activeProviders = useMemo(() => providers.filter((item) => item.aktif), [providers]);

  async function loadProduk(): Promise<ProdukRow[]> {
    setProdukLoading(true);
    try {
      const items = await fetchAllAdminProducts(authHeader());
      const nextItems = Array.isArray(items) ? items : [];
      setProduk(nextItems);
      return nextItems;
    } catch {
      setProduk([]);
      return [];
    } finally {
      setProdukLoading(false);
    }
  }

  async function ensureProdukLoaded(): Promise<ProdukRow[]> {
    if (produk.length > 0) return produk;
    if (produkLoading) return produk;
    return loadProduk();
  }

  async function loadProviders() {
    const r = await fetch("/api/admin/master/provider", { headers: authHeader(), cache: "no-store" });
    const j = await r.json().catch(() => ({}));
    if (!r.ok || !j.ok) {
      setProviders([]);
      return;
    }
    setProviders(Array.isArray(j.items) ? j.items : []);
  }

  async function loadMaps() {
    const r = await fetch("/api/admin/master/produk-provider-map", {
      headers: authHeader(),
      cache: "no-store",
    });
    const j = await r.json().catch(() => ({}));
    if (!r.ok || !j.ok) {
      throw new Error(j.error || "Gagal memuat mapping produk provider");
    }
    setMaps(Array.isArray(j.items) ? j.items : []);
  }

  async function refreshAll() {
    setLoading(true);
    try {
      await Promise.all([loadProviders(), loadMaps()]);
    } catch (error) {
      const message = error instanceof Error ? error.message : "Gagal memuat konfigurasi provider";
      await alertError(message);
      setMaps([]);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void refreshAll();
  }, []);

  useEffect(() => {
    const token = localStorage.getItem("auth_token") || "";
    const role = String(decodeJwt(token)?.role || "").trim().toLowerCase();
    setCurrentRole(role);
  }, []);

  useEffect(() => {
    function handleOutsideClick(e: MouseEvent) {
      const target = e.target as HTMLElement | null;
      if (!target) return;
      if (target.closest("[data-action-dropdown]")) return;
      setOpenActionKey(null);
    }

    function handleEsc(e: KeyboardEvent) {
      if (e.key === "Escape") setOpenActionKey(null);
    }

    document.addEventListener("mousedown", handleOutsideClick);
    document.addEventListener("keydown", handleEsc);
    return () => {
      document.removeEventListener("mousedown", handleOutsideClick);
      document.removeEventListener("keydown", handleEsc);
    };
  }, []);

  const combinedItems = useMemo<CombinedRow[]>(() => {
    const out: CombinedRow[] = [];
    for (const item of maps) {
      const providerKey = normalizeProviderName(item.provider);
      const providerMatch = providers.find((provider) => normalizeProviderName(provider.nama) === providerKey) || null;
      const key = buildMapRowKey({
        id: item.id,
        produk_id: item.produk_id,
        provider: item.provider,
        kode_provider: item.kode_provider,
        minimal_nominal: item.minimal_nominal,
        maksimal_nominal: item.maksimal_nominal,
      });
      const derivedFee = usesCodeBasedFee(providerMatch?.nama || item.provider)
        ? minionsFeeByCode(item.kode_provider || "")
        : null;

      out.push({
        key,
        produk_id: item.produk_id,
        produk_sku: item.produk_sku,
        produk_nama: item.produk_nama,
        tipe_harga: item.tipe_harga || "FIXED",
        nominal: item.nominal != null ? Number(item.nominal) : null,
        minimal_nominal: item.minimal_nominal != null ? Number(item.minimal_nominal) : null,
        maksimal_nominal: item.maksimal_nominal != null ? Number(item.maksimal_nominal) : null,
        provider_id: providerMatch?.id ?? null,
        provider_nama: providerMatch?.nama || item.provider,
        map_id: item.id,
        kode_provider: item.kode_provider || "",
        special_code: (item.special_code || "").trim().toUpperCase(),
        mode: (item.mode || "").trim().toUpperCase(),
        map_aktif: Boolean(item.aktif),
        fee_rp: derivedFee != null ? derivedFee : Number(item.fee_rp || 0),
        jam_buka: item.jam_buka || "",
        jam_tutup: item.jam_tutup || "",
      });
    }

    return out.sort((a, b) => {
      const skuCompare = a.produk_sku.localeCompare(b.produk_sku);
      if (skuCompare !== 0) return skuCompare;
      const providerCompare = a.provider_nama.localeCompare(b.provider_nama);
      if (providerCompare !== 0) return providerCompare;
      return (a.maksimal_nominal ?? 0) - (b.maksimal_nominal ?? 0);
    });
  }, [maps, providers]);

  const filteredItems = useMemo(() => {
    let items = combinedItems;
    if (tab !== "semua") {
      items = items.filter((item) => classifyProduct(item.produk_sku) === tab);
    }
    const q = search.trim().toLowerCase();
    if (!q) return items;
    return items.filter((item) =>
      [
        item.produk_sku,
        item.produk_nama,
        item.provider_nama,
        item.kode_provider,
        item.special_code,
        item.tipe_harga,
        String(item.nominal || 0),
        String(item.minimal_nominal || 0),
        String(item.maksimal_nominal || 0),
        String(item.fee_rp || 0),
        item.jam_buka || "",
        item.jam_tutup || "",
      ]
        .join(" ")
        .toLowerCase()
        .includes(q),
    );
  }, [combinedItems, search, tab]);

  async function openCreate() {
    const loadedProduk = await ensureProdukLoaded();
    const produkSource = loadedProduk.length ? loadedProduk : produk;
    const activeProdukSource = produkSource.filter((item) => item.aktif);
    setEditing(null);
    setForm({
      produk_id: activeProdukSource[0]?.id ? String(activeProdukSource[0].id) : "",
      provider_id: activeProviders[0]?.id ? String(activeProviders[0].id) : "",
      tipe_harga: activeProdukSource[0]?.tipe_harga || "FIXED",
      minimal_nominal: "",
      maksimal_nominal: "",
      kode_provider: "",
      special_code: "",
      mode: "",
      fee_rp: "0",
      map_aktif: true,
      jam_buka: "",
      jam_tutup: "",
    });
    setOpen(true);
  }

  async function openEdit(item: CombinedRow) {
    await ensureProdukLoaded();
    setEditing(item);
    setForm({
      produk_id: String(item.produk_id),
      provider_id: item.provider_id ? String(item.provider_id) : "",
      tipe_harga: item.tipe_harga,
      minimal_nominal: item.minimal_nominal != null ? String(item.minimal_nominal) : "",
      maksimal_nominal: item.maksimal_nominal != null ? String(item.maksimal_nominal) : "",
      kode_provider: item.kode_provider || "",
      special_code: (item.special_code || "").trim().toUpperCase(),
      mode: (item.mode || "").trim().toUpperCase(),
      fee_rp: String(item.fee_rp ?? 0),
      map_aktif: Boolean(item.map_aktif),
      jam_buka: (item.jam_buka || "").slice(0, 5),
      jam_tutup: (item.jam_tutup || "").slice(0, 5),
    });
    setOpen(true);
  }

  async function save() {
    const produkID = Number(form.produk_id || "0");
    const providerID = Number(form.provider_id || "0");
    const feeRp = Number(form.fee_rp || "0");
    const minimalNominal = toNumOrNull(form.minimal_nominal);
    const maksimalNominal = toNumOrNull(form.maksimal_nominal);

    if (!produkID) return alertWarning("Produk wajib dipilih.");
    if (!providerID) return alertWarning("Provider wajib dipilih.");
    if (!form.kode_provider.trim()) return alertWarning("Kode provider wajib diisi.");
    if (Number.isNaN(feeRp) || feeRp < 0) return alertWarning("Fee harus angka >= 0.");
    if (form.tipe_harga === "OPEN_AMOUNT" && !maksimalNominal) {
      return alertWarning("Untuk OPEN_AMOUNT, maksimal nominal wajib diisi.");
    }

    const provider = providers.find((item) => item.id === providerID);
    if (!provider) return alertWarning("Provider tidak ditemukan.");
    const useDerivedFee = usesCodeBasedFee(provider.nama);

    setSaving(true);
    try {
      const mapPayload = {
        produk_id: produkID,
        provider: normalizeProviderName(provider.nama),
        kode_provider: form.kode_provider.trim().toUpperCase(),
        special_code: form.special_code.trim().toUpperCase() || null,
        mode: form.mode.trim().toUpperCase() || null,
        minimal_nominal: form.tipe_harga === "OPEN_AMOUNT" ? minimalNominal : null,
        maksimal_nominal: form.tipe_harga === "OPEN_AMOUNT" ? maksimalNominal : null,
        fee_rp: useDerivedFee ? minionsFeeByCode(form.kode_provider || "") : feeRp,
        aktif: form.map_aktif,
        jam_buka: form.jam_buka || null,
        jam_tutup: form.jam_tutup || null,
      };
      const mapURL = editing?.map_id ? `/api/admin/master/produk-provider-map/${editing.map_id}` : "/api/admin/master/produk-provider-map";
      const mapMethod = editing?.map_id ? "PUT" : "POST";
      const mapResp = await fetch(mapURL, {
        method: mapMethod,
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify(mapPayload),
      });
      const mapJson = await mapResp.json().catch(() => ({}));
      if (!mapResp.ok || !mapJson.ok) return alertError(mapJson.error || "Gagal menyimpan mapping provider");

      setOpen(false);
      await refreshAll();
      await alertSuccess(editing ? "Konfigurasi provider berhasil diperbarui." : "Konfigurasi provider berhasil ditambahkan.");
    } finally {
      setSaving(false);
    }
  }

  async function remove(item: CombinedRow) {
    const ok = await alertConfirm({
      title: "Hapus Konfigurasi Provider",
      text: `Yakin hapus mapping ${item.produk_sku} @ ${item.provider_nama}?`,
      confirmButtonText: "Ya, hapus",
    });
    if (!ok) return;

    if (!item.map_id) return alertWarning("Data mapping belum tersimpan.");

    const resp = await fetch(`/api/admin/master/produk-provider-map/${item.map_id}`, {
      method: "DELETE",
      headers: authHeader(),
    });
    const payload = await resp.json().catch(() => ({}));
    if (!resp.ok || !payload?.ok) {
      return alertError(payload?.error || "Gagal menghapus konfigurasi provider");
    }

    await refreshAll();
    await alertSuccess("Konfigurasi provider berhasil dihapus.");
  }

  const isOperatorToggleOnly = currentRole === "operator_trx" && (pathname?.startsWith("/dashboard/operator") ?? false);
  const canManageFull = !isOperatorToggleOnly;

  async function toggleActive(item: CombinedRow) {
    if (!item.map_id) return alertWarning("Data mapping belum tersimpan.");

    const nextAktif = !item.map_aktif;
    const ok = await alertConfirm({
      title: nextAktif ? "Aktifkan Mapping" : "Nonaktifkan Mapping",
      text: `${nextAktif ? "Aktifkan" : "Nonaktifkan"} mapping ${item.produk_sku} @ ${item.provider_nama}?`,
      confirmButtonText: nextAktif ? "Ya, aktifkan" : "Ya, nonaktifkan",
    });
    if (!ok) return;

    const resp = await fetch(`/api/admin/master/produk-provider-map/${item.map_id}/toggle-active`, {
      method: "POST",
      headers: { "Content-Type": "application/json", ...authHeader() },
      body: JSON.stringify({ aktif: nextAktif }),
    });
    const payload = await resp.json().catch(() => ({}));
    if (!resp.ok || !payload?.ok) {
      return alertError(payload?.error || "Gagal mengubah status mapping provider");
    }

    await refreshAll();
    await alertSuccess(`Mapping provider berhasil ${nextAktif ? "diaktifkan" : "dinonaktifkan"}.`);
  }

  const columns: DataTableColumn<CombinedRow>[] = [
    {
      id: "produk",
      header: "Produk",
      tdClassName: "whitespace-nowrap",
      render: (item) => (
        <div>
          <div className="font-medium text-slate-100">{item.produk_sku}</div>
          <div className="text-xs text-slate-400">{item.produk_nama}</div>
        </div>
      ),
    },
    {
      id: "provider",
      header: "Provider",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (item) => item.provider_nama,
    },
    {
      id: "tipe_harga",
      header: "Tipe Harga",
      tdClassName: "whitespace-nowrap",
      render: (item) => (
        <span className="rounded-full border border-cyan-400/30 bg-cyan-500/10 px-2 py-0.5 text-xs font-semibold text-cyan-200">
          {item.tipe_harga}
        </span>
      ),
    },
    {
      id: "limit_produk",
      header: "Limit Mapping",
      tdClassName: "whitespace-nowrap font-semibold text-slate-200",
      render: (item) =>
        item.tipe_harga === "OPEN_AMOUNT"
          ? item.maksimal_nominal
            ? `${item.minimal_nominal ? `Min ${formatRupiah(item.minimal_nominal)} · ` : ""}Max ${formatRupiah(item.maksimal_nominal)}`
            : "-"
          : item.nominal
            ? formatRupiah(item.nominal)
            : "-",
    },
    {
      id: "kode_provider",
      header: "Kode Provider",
      tdClassName: "whitespace-nowrap font-mono text-xs text-cyan-200",
      render: (item) => item.kode_provider || "-",
    },
    {
      id: "special_code",
      header: "Special Code",
      tdClassName: "whitespace-nowrap font-mono text-xs text-fuchsia-200",
      render: (item) => item.special_code || <span className="text-slate-400 italic">-</span>,
    },
    {
      id: "mode",
      header: "Mode",
      tdClassName: "whitespace-nowrap font-mono text-xs text-amber-200",
      render: (item) => item.mode || <span className="text-slate-400 italic">Legacy</span>,
    },
    {
      id: "fee_rp",
      header: "Fee Provider",
      tdClassName: "whitespace-nowrap font-semibold text-cyan-200",
      render: (item) => formatRupiah(item.fee_rp),
    },
    {
      id: "jam_online",
      header: "Jam Online",
      tdClassName: "whitespace-nowrap text-xs text-slate-300",
      render: (item) => {
        if (!item.jam_buka && !item.jam_tutup) return <span className="text-slate-400 italic">Ikut Produk</span>;
        const buka = (item.jam_buka || "").slice(0, 5);
        const tutup = (item.jam_tutup || "").slice(0, 5);
        if (!buka && !tutup) return <span className="text-slate-400 italic">Ikut Produk</span>;
        const is24h = buka === "00:00" && tutup === "23:59";
        return is24h
          ? <span className="rounded-full border border-emerald-400/30 bg-emerald-500/10 px-2 py-0.5 text-xs font-semibold text-emerald-300">24 Jam</span>
          : <span>{buka || "?"} - {tutup || "?"}</span>;
      },
    },
    {
      id: "map_status",
      header: "Status Mapping",
      tdClassName: "whitespace-nowrap",
      render: (item) =>
        item.map_id ? (
          item.map_aktif ? (
            <span className="rounded-full border border-emerald-400/40 bg-emerald-500/15 px-2 py-0.5 text-xs font-semibold text-emerald-300">Aktif</span>
          ) : (
            <span className="rounded-full border border-rose-400/40 bg-rose-500/15 px-2 py-0.5 text-xs font-semibold text-rose-300">Nonaktif</span>
          )
        ) : (
          <span className="rounded-full border border-amber-400/40 bg-amber-500/15 px-2 py-0.5 text-xs font-semibold text-amber-200">Belum ada</span>
        ),
    },
  ];

  const actions: DataTableActions<CombinedRow> = {
    header: "Aksi",
    align: "right",
    tdClassName: "whitespace-nowrap",
    render: (item, index) => {
      if (isOperatorToggleOnly) {
        return (
          <Button
            size="sm"
            className={`h-8 ${item.map_aktif ? "bg-rose-600 hover:bg-rose-500" : "bg-emerald-600 hover:bg-emerald-500"} text-white`}
            onClick={() => void toggleActive(item)}
          >
            <Power className="mr-1.5 h-3.5 w-3.5" />
            {item.map_aktif ? "Nonaktifkan" : "Aktifkan"}
          </Button>
        );
      }

      return (
        <div className="relative inline-flex" data-action-dropdown>
          <Button
            size="sm"
            className="h-8 w-8 px-0 bg-linear-to-r from-sky-500 to-cyan-500 text-white hover:opacity-90"
            onClick={() => setOpenActionKey((prev) => (prev === item.key ? null : item.key))}
            aria-label="Aksi"
          >
            <Eye className="h-4 w-4" />
          </Button>

          {openActionKey === item.key ? (
            <div
              className={`absolute right-0 z-20 w-36 rounded-xl border border-white/12 bg-slate-900/95 p-1.5 shadow-2xl backdrop-blur ${
                index >= filteredItems.length - 2 ? "bottom-10" : "top-10"
              }`}
            >
              <button
                type="button"
                className="flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm text-slate-100 transition hover:bg-white/10"
                onClick={() => {
                  setOpenActionKey(null);
                  void openEdit(item);
                }}
              >
                <Pencil className="h-3.5 w-3.5" />
                Edit
              </button>
              <button
                type="button"
                className="mt-1 flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm text-rose-300 transition hover:bg-rose-500/15"
                onClick={() => {
                  setOpenActionKey(null);
                  void remove(item);
                }}
              >
                <Trash2 className="h-3.5 w-3.5" />
                Hapus
              </button>
            </div>
          ) : null}
        </div>
      );
    },
  };

  return (
    <div className="space-y-4 p-2">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <div className="text-lg font-semibold tracking-tight">Master Mapping & Fee Provider</div>
          <div className="text-sm text-muted-foreground">Kelola mapping provider, fee provider, dan limit nominal bebas per provider dalam satu halaman. Produk tetap menjadi master identitas, sedangkan limit OPEN_AMOUNT disimpan di mapping provider.</div>
          {isOperatorToggleOnly ? <div className="mt-2 text-xs text-amber-300">Mode operator transaksi: hanya bisa mengaktifkan dan menonaktifkan mapping. Perubahan field lain tetap admin-only.</div> : null}
        </div>
        <div className="flex gap-2">
          <Button variant="outline" className="h-10" onClick={() => void refreshAll()} disabled={loading}>
            <RefreshCw className={`mr-2 h-4 w-4 ${loading ? "animate-spin" : ""}`} />
            Refresh
          </Button>
          {canManageFull ? (
            <Button variant="primary" className="h-10" onClick={() => void openCreate()}>
              <Plus className="mr-2 h-4 w-4" />
              Tambah Konfigurasi
            </Button>
          ) : null}
        </div>
      </div>

      <div className="flex flex-wrap gap-2">
        <Input
          value={searchInput}
          onChange={(e) => setSearchInput(e.target.value)}
          placeholder="Cari sku / nama produk / provider / tipe harga / limit / kode provider / fee"
          className="h-10 w-full max-w-xl"
          onKeyDown={(e) => {
            if (e.key === "Enter") setSearch(searchInput);
          }}
        />
        <Button variant="primary" className="h-10" onClick={() => setSearch(searchInput)}>
          <Search className="mr-2 h-4 w-4" />
          Cari
        </Button>
      </div>

      <div className="flex gap-2">
        {(["semua","ewallet","bank","lainnya"] as const).map((t) => (
          <button key={t} type="button"
            className={`rounded-full px-4 py-1.5 text-sm font-medium transition ${tab === t ? "bg-cyan-500/20 text-cyan-200 border border-cyan-400/40" : "text-slate-400 hover:text-slate-200 border border-transparent"}`}
            onClick={() => setTab(t)}>
            {t === "semua" ? "Semua" : t === "ewallet" ? "E-Wallet" : t === "bank" ? "Bank" : "Lainnya"}
            <span className="ml-1.5 text-xs opacity-60">
              {t === "semua" ? combinedItems.length : combinedItems.filter(i => classifyProduct(i.produk_sku) === t).length}
            </span>
          </button>
        ))}
      </div>

      <DataTable
        columns={columns}
        rows={filteredItems}
        rowKey={(item) => item.key}
        minWidthClassName="min-w-[1180px]"
        emptyText="Belum ada konfigurasi provider."
        actions={actions}
      />

      <AppModal
        open={open}
        onClose={() => setOpen(false)}
        title={editing ? "Edit Mapping & Fee Provider" : "Tambah Mapping & Fee Provider"}
        subtitle="Satu form untuk mengelola kode provider dan fee provider tanpa mengubah logika transaksi yang sudah berjalan."
        maxWidthClassName="max-w-3xl"
        footer={
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={() => setOpen(false)} disabled={saving}>
              Batal
            </Button>
            <Button variant="success" onClick={() => void save()} disabled={saving}>
              {saving ? "Saving..." : "Simpan"}
            </Button>
          </div>
        }
      >
        <div className="grid gap-4 md:grid-cols-2">
          <div className="grid gap-1.5 md:col-span-2">
            <div className="text-xs font-semibold uppercase tracking-wider text-slate-300">Produk Internal</div>
            <ProductSearchSelect
              items={produk}
              value={form.produk_id}
              loading={produkLoading}
              onChange={(nextValue) => {
                const selected = produk.find((item) => String(item.id) === nextValue);
                setForm((state) => ({
                  ...state,
                  produk_id: nextValue,
                  tipe_harga: selected?.tipe_harga || "FIXED",
                  minimal_nominal: "",
                  maksimal_nominal: "",
                }));
              }}
              placeholder="Pilih produk"
            />
          </div>

          <div className="grid gap-1.5">
            <div className="text-xs font-semibold uppercase tracking-wider text-slate-300">Tipe Harga Produk</div>
            <select
              value={form.tipe_harga}
              disabled
              className="h-11 rounded-lg border border-white/15 bg-slate-950/70 px-3 text-sm text-slate-100 outline-hidden ring-0 transition focus:border-cyan-400/60"
            >
              <option value="FIXED">FIXED</option>
              <option value="OPEN_AMOUNT">OPEN_AMOUNT</option>
            </select>
          </div>

          <div className="grid gap-1.5">
            <div className="text-xs font-semibold uppercase tracking-wider text-slate-300">Provider</div>
            <select
              value={form.provider_id}
              onChange={(e) => setForm((state) => ({ ...state, provider_id: e.target.value }))}
              className="h-11 rounded-lg border border-white/15 bg-slate-950/70 px-3 text-sm text-slate-100 outline-hidden ring-0 transition focus:border-cyan-400/60"
            >
              <option value="" disabled>
                Pilih provider
              </option>
              {providers.map((provider) => (
                <option key={provider.id} value={provider.id}>
                  {provider.nama} {!provider.aktif ? "(nonaktif)" : ""}
                </option>
              ))}
            </select>
          </div>

          <div className="grid gap-1.5">
            <div className="text-xs font-semibold uppercase tracking-wider text-slate-300">Kode Provider</div>
            <Input
              value={form.kode_provider}
              onChange={(e) => setForm((state) => ({ ...state, kode_provider: e.target.value.toUpperCase() }))}
              className="h-11 border-white/15 bg-slate-950/70 text-slate-100 placeholder:text-slate-500"
              placeholder="contoh: DNID / DANA / UDDANA10"
            />
          </div>

          <div className="grid gap-1.5">
            <div className="text-xs font-semibold uppercase tracking-wider text-slate-300">Special Code</div>
            <Input
              value={form.special_code}
              onChange={(e) => setForm((state) => ({ ...state, special_code: e.target.value.toUpperCase() }))}
              className="h-11 border-white/15 bg-slate-950/70 text-slate-100 placeholder:text-slate-500"
              placeholder="contoh: BIFASTOPEN"
            />
          </div>

          <div className="grid gap-1.5">
            <div className="text-xs font-semibold uppercase tracking-wider text-slate-300">Mode Provider</div>
            <select
              value={form.mode}
              onChange={(e) => setForm((state) => ({ ...state, mode: e.target.value.toUpperCase() }))}
              className="h-11 rounded-lg border border-white/15 bg-slate-950/70 px-3 text-sm text-slate-100 outline-hidden ring-0 transition focus:border-cyan-400/60"
            >
              <option value="">Otomatis / Legacy</option>
              <option value="DIRECT">DIRECT</option>
              <option value="PPOB">PPOB</option>
              <option value="WALLET_PPOB">WALLET_PPOB</option>
              <option value="ELEKTRIK">ELEKTRIK</option>
            </select>
          </div>

          <div className="grid gap-1.5">
            <div className="text-xs font-semibold uppercase tracking-wider text-slate-300">Fee Provider (Rp)</div>
            <Input
              value={form.fee_rp}
              onChange={(e) => setForm((state) => ({ ...state, fee_rp: e.target.value.replace(/[^0-9]/g, "") }))}
              disabled={form.provider_id ? usesCodeBasedFee(providers.find((item) => String(item.id) === form.provider_id)?.nama) : false}
              className="h-11 border-white/15 bg-slate-950/70 text-slate-100 placeholder:text-slate-500"
              placeholder="contoh: 70"
            />
            {form.provider_id && usesCodeBasedFee(providers.find((item) => String(item.id) === form.provider_id)?.nama) ? (
              <div className="text-xs text-amber-300">Fee Minions mengikuti kode provider dan dipakai langsung oleh backend.</div>
            ) : null}
          </div>

          {form.tipe_harga === "OPEN_AMOUNT" ? (
            <>
              <div className="grid gap-1.5">
                <div className="text-xs font-semibold uppercase tracking-wider text-slate-300">Minimal Nominal Mapping</div>
                <Input
                  value={form.minimal_nominal}
                  onChange={(e) => setForm((state) => ({ ...state, minimal_nominal: e.target.value.replace(/[^0-9]/g, "") }))}
                  className="h-11 border-white/15 bg-slate-950/70 text-slate-100 placeholder:text-slate-500"
                  placeholder="contoh: 1000"
                />
              </div>

              <div className="grid gap-1.5">
                <div className="text-xs font-semibold uppercase tracking-wider text-slate-300">Maksimal Nominal Mapping</div>
                <Input
                  value={form.maksimal_nominal}
                  onChange={(e) => setForm((state) => ({ ...state, maksimal_nominal: e.target.value.replace(/[^0-9]/g, "") }))}
                  className="h-11 border-white/15 bg-slate-950/70 text-slate-100 placeholder:text-slate-500"
                  placeholder="contoh: 5000000"
                />
              </div>
            </>
          ) : (
            <div className="rounded-xl border border-white/10 bg-slate-950/40 p-3 text-sm text-slate-300 md:col-span-2">
              Nominal produk FIXED mengikuti master produk. Halaman ini hanya mengatur mapping kode provider, fee provider, dan limit OPEN_AMOUNT per provider.
            </div>
          )}

          <label className="mt-1 inline-flex items-center gap-2 text-sm text-slate-200">
            <input
              type="checkbox"
              checked={form.map_aktif}
              onChange={(e) => setForm((state) => ({ ...state, map_aktif: e.target.checked }))}
              className="h-4 w-4 rounded border-white/20 bg-slate-900"
            />
            Mapping aktif
          </label>

          <div className="grid gap-1.5">
            <div className="text-xs font-semibold uppercase tracking-wider text-slate-300">Jam Buka</div>
            <Input
              type="time"
              value={form.jam_buka}
              onChange={(e) => setForm((state) => ({ ...state, jam_buka: e.target.value }))}
              className="h-11 border-white/15 bg-slate-950/70 text-slate-100"
            />
          </div>
          <div className="grid gap-1.5">
            <div className="text-xs font-semibold uppercase tracking-wider text-slate-300">Jam Tutup</div>
            <Input
              type="time"
              value={form.jam_tutup}
              onChange={(e) => setForm((state) => ({ ...state, jam_tutup: e.target.value }))}
              className="h-11 border-white/15 bg-slate-950/70 text-slate-100"
            />
          </div>

          <div className="mt-1 rounded-xl border border-emerald-400/20 bg-emerald-500/10 px-3 py-2 text-sm text-emerald-200">
            Fee provider selalu aktif. Yang bisa diaktifkan atau dinonaktifkan hanya mapping produk ke provider.
          </div>
        </div>
      </AppModal>
    </div>
  );
}
