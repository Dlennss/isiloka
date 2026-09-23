"use client";

import { useEffect, useMemo, useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { ArrowLeft, Loader2, Signal, Smartphone, Wifi, X } from "lucide-react";
import { UserProductGrid } from "@/components/user/UserProductGrid";
import type { UserBrandItem, UserCategoryItem, UserProductItem } from "@/components/user/types";

type TabKey = "pulsa" | "data";

type CategoryWithBrands = {
  category: UserCategoryItem;
  brands: UserBrandItem[];
};

type Props = {
  tabs: Partial<Record<TabKey, CategoryWithBrands>>;
  isLoggedIn: boolean;
  authToken?: string;
  buyerRole?: string;
  backHref?: string;
  loginCallbackUrl?: string;
  showBackLink?: boolean;
  initialTab?: TabKey;
  title?: string;
  showTabs?: boolean;
};

type OperatorMatch = {
  label: string;
  brandKeywords: string[];
};

function getOperatorVisual(operator: OperatorMatch | null) {
  const label = operator?.label || "";
  const key = normalizeName(label);
  if (key.includes("by u") || key.includes("byu")) return { short: "by.U", logo: "/images/providers/logo_byu.webp" };
  if (key.includes("telkomsel")) return { short: "Tsel", logo: "/images/providers/logo_telkomsel.webp" };
  if (key.includes("indosat")) return { short: "im3", logo: "/images/providers/logo_im3.webp" };
  if (key === "xl") return { short: "XL", logo: "/images/providers/logo_xl.png" };
  if (key.includes("axis")) return { short: "AXIS", logo: "/images/providers/logo_axis.webp" };
  if (key.includes("tri")) return { short: "3", logo: "/images/providers/logo_tri.webp" };
  if (key.includes("smartfren")) return { short: "SF", logo: "/images/providers/logo_smartfren.webp" };
  return { short: "HP", logo: "" };
}

function normalizeName(value: string) {
  return value.toLowerCase().replace(/[^a-z0-9]+/g, " ").trim();
}

const OPERATOR_PREFIXES: Array<{ prefixes: string[]; match: OperatorMatch }> = [
  { prefixes: ["0851"], match: { label: "by.U", brandKeywords: ["by u", "byu", "by.u"] } },
  { prefixes: ["0811", "0812", "0813", "0821", "0822", "0823", "0852", "0853"], match: { label: "Telkomsel", brandKeywords: ["telkomsel", "simpati", "as"] } },
  { prefixes: ["0814", "0815", "0816", "0855", "0856", "0857", "0858"], match: { label: "Indosat", brandKeywords: ["indosat", "im3", "mentari", "isat"] } },
  { prefixes: ["0817", "0818", "0819", "0859", "0877", "0878"], match: { label: "XL", brandKeywords: ["xl"] } },
  { prefixes: ["0831", "0832", "0833", "0838"], match: { label: "AXIS", brandKeywords: ["axis"] } },
  { prefixes: ["0895", "0896", "0897", "0898", "0899"], match: { label: "Tri", brandKeywords: ["tri", "three", "3"] } },
  { prefixes: ["0881", "0882", "0883", "0884", "0885", "0886", "0887", "0888", "0889"], match: { label: "Smartfren", brandKeywords: ["smartfren"] } },
];

function normalizePhone(value: string) {
  const digits = value.replace(/\D+/g, "");
  if (!digits) return "";
  if (digits.startsWith("62")) return `0${digits.slice(2)}`;
  return digits;
}

function detectOperator(phone: string): OperatorMatch | null {
  const normalized = normalizePhone(phone);
  if (normalized.length < 4) return null;
  for (const item of OPERATOR_PREFIXES) {
    if (item.prefixes.some((prefix) => normalized.startsWith(prefix))) {
      return item.match;
    }
  }
  return null;
}

function findBrandId(brands: UserBrandItem[], operator: OperatorMatch | null) {
  if (!operator) return "";

  const scored = brands
    .map((brand) => {
      const normalizedBrand = normalizeName(brand.nama);
      const score = operator.brandKeywords.reduce((total, keyword) => {
        const normalizedKeyword = normalizeName(keyword);
        if (!normalizedKeyword) return total;
        if (normalizedBrand === normalizedKeyword) return total + 100;
        if (normalizedBrand.startsWith(normalizedKeyword)) return total + 60;
        if (normalizedBrand.includes(normalizedKeyword)) return total + 40;

        const brandTokens = normalizedBrand.split(/\s+/).filter(Boolean);
        const keywordTokens = normalizedKeyword.split(/\s+/).filter(Boolean);
        const tokenHits = keywordTokens.filter((token) => brandTokens.includes(token)).length;
        return total + tokenHits * 15;
      }, 0);

      return { brand, score };
    })
    .sort((a, b) => b.score - a.score);

  const best = scored[0];
  return best && best.score > 0 ? String(best.brand.id) : "";
}

async function fetchProducts(categoryId: string, brandId: string) {
  const url = `/api/app/produk?kategori_id=${encodeURIComponent(categoryId)}&brand_id=${encodeURIComponent(brandId)}`;
  const res = await fetch(url, { cache: "no-store" });
  if (!res.ok) return [] as UserProductItem[];
  const json = (await res.json().catch(() => ({}))) as { items?: UserProductItem[] };
  return Array.isArray(json.items) ? json.items : [];
}

export function UserPulsaDataExplorer({
  tabs,
  isLoggedIn,
  authToken,
  buyerRole,
  backHref = "/user",
  showBackLink = true,
  initialTab,
  title = "Pulsa & Data",
  showTabs = true,
}: Props) {
  const [phone, setPhone] = useState("");
  const [activeTab, setActiveTab] = useState<TabKey>(initialTab && tabs[initialTab] ? initialTab : tabs.pulsa ? "pulsa" : "data");
  const [products, setProducts] = useState<UserProductItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const detectedOperator = useMemo(() => detectOperator(phone), [phone]);
  const operatorVisual = useMemo(() => getOperatorVisual(detectedOperator), [detectedOperator]);
  const activeGroup = tabs[activeTab] ?? null;
  const brandId = useMemo(() => findBrandId(activeGroup?.brands ?? [], detectedOperator), [activeGroup?.brands, detectedOperator]);

  useEffect(() => {
    let cancelled = false;

    async function run() {
      if (!activeGroup) {
        setProducts([]);
        setError("Kategori belum tersedia.");
        return;
      }
      if (!detectedOperator) {
        setProducts([]);
        setError("");
        return;
      }
      if (!brandId) {
        setProducts([]);
        setError(`Brand ${detectedOperator.label} belum tersedia di kategori ${activeGroup.category.nama}.`);
        return;
      }

      setLoading(true);
      setError("");
      try {
        const items = await fetchProducts(String(activeGroup.category.id), brandId);
        if (!cancelled) {
          const activeItems = items
            .filter((item) => item.aktif)
            .sort((a, b) => Number(a.nominal || a.harga_guest_final || a.harga_user_final || 0) - Number(b.nominal || b.harga_guest_final || b.harga_user_final || 0));
          setProducts(activeItems);
          if (!activeItems.length) {
            setError(`Belum ada produk ${activeGroup.category.nama} untuk ${detectedOperator.label}.`);
          }
        }
      } catch {
        if (!cancelled) {
          setProducts([]);
          setError("Gagal memuat produk.");
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    void run();
    return () => {
      cancelled = true;
    };
  }, [activeGroup, brandId, detectedOperator]);

  return (
    <div className="pb-24">
      <section className="relative overflow-hidden rounded-b-[30px] bg-linear-to-br from-[#13464b] via-[#087e8b] to-[#22c55e] px-4 pb-7 pt-4 text-white shadow-[0_18px_42px_rgba(6,78,59,0.20)]">
        <div className="pointer-events-none absolute -right-14 -top-16 h-40 w-40 rounded-full bg-lime-300/30 blur-2xl" />
        <div className="pointer-events-none absolute -bottom-12 left-12 h-28 w-44 rotate-[-10deg] bg-emerald-200/15 blur-2xl" />

        <div className="relative flex h-10 items-center justify-center">
          {showBackLink ? (
            <Link
              href={backHref}
              className="absolute left-0 grid h-9 w-9 place-items-center rounded-2xl bg-white/10 text-white ring-1 ring-white/15"
              aria-label="Kembali"
            >
              <ArrowLeft className="h-5 w-5" />
            </Link>
          ) : null}
          <h1 className="text-lg font-black tracking-tight">{title}</h1>
        </div>

        <div className="relative mt-4 rounded-[22px] bg-white p-2 text-slate-900 shadow-[0_16px_34px_rgba(5,46,38,0.18)]">
          <div className="flex items-center gap-2">
            <div className={`grid h-11 w-11 shrink-0 place-items-center overflow-hidden rounded-2xl ${detectedOperator ? "bg-white ring-1 ring-slate-200" : "bg-emerald-50 text-[#087e8b]"}`}>
              {detectedOperator && operatorVisual.logo ? (
                <Image src={operatorVisual.logo} alt={`Logo ${detectedOperator.label}`} width={44} height={44} className="h-full w-full object-contain p-1.5" />
              ) : detectedOperator ? (
                <span className="text-xs font-black uppercase text-slate-700">{operatorVisual.short}</span>
              ) : (
                <Smartphone className="h-5 w-5" />
              )}
            </div>
            <input
              value={phone}
              onChange={(event) => setPhone(event.target.value.replace(/\D+/g, ""))}
              inputMode="numeric"
              pattern="[0-9]*"
              placeholder="Masukkan nomor HP"
              className="h-11 min-w-0 flex-1 bg-transparent text-lg font-black tracking-tight text-slate-900 outline-none placeholder:text-sm placeholder:font-semibold placeholder:text-slate-400"
            />
            {phone ? (
              <button
                type="button"
                onClick={() => setPhone("")}
                className="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-slate-100 text-slate-500"
                aria-label="Hapus nomor"
              >
                <X className="h-4 w-4" />
              </button>
            ) : null}
          </div>
          {detectedOperator ? (
            <div className="mt-2 inline-flex rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-[#087e8b]">
              {detectedOperator.label}
            </div>
          ) : phone.length >= 4 ? (
            <div className="mt-2 inline-flex rounded-full bg-amber-50 px-3 py-1 text-xs font-black text-amber-700">
              Operator belum dikenali
            </div>
          ) : null}
        </div>

        {showTabs ? (
          <div className="relative mt-4 grid grid-cols-2 gap-1 rounded-2xl bg-[#13464b]/28 p-1 ring-1 ring-white/15">
            {(["pulsa", "data"] as TabKey[]).map((tab) => {
              const enabled = Boolean(tabs[tab]);
              const active = activeTab === tab;
              return (
                <button
                  key={tab}
                  type="button"
                  disabled={!enabled}
                  onClick={() => enabled && setActiveTab(tab)}
                  className={`inline-flex h-10 items-center justify-center gap-2 rounded-xl text-sm font-black transition ${active ? "bg-white text-[#087e8b] shadow-[0_10px_20px_rgba(5,46,38,0.16)]" : "text-white/80"} ${!enabled ? "cursor-not-allowed opacity-40" : ""}`}
                >
                  {tab === "pulsa" ? <Signal className="h-4 w-4" /> : <Wifi className="h-4 w-4" />}
                  {tab === "pulsa" ? "Pulsa" : "Paket Data"}
                </button>
              );
            })}
          </div>
        ) : null}

      </section>

      <section className="-mt-4 space-y-3 rounded-t-[28px] bg-[#f8fafc] px-4 pt-5">
        {!phone.trim() ? (
          <div className="rounded-2xl border border-emerald-100 bg-white px-4 py-5 text-center shadow-[0_10px_28px_rgba(15,23,42,0.08)]">
            <p className="text-sm font-black text-slate-900">Masukkan nomor HP</p>
            <p className="mt-1 text-xs font-semibold leading-5 text-slate-500">
              Produk akan muncul otomatis sesuai operator nomor.
            </p>
          </div>
        ) : !detectedOperator ? (
          <div className="rounded-2xl border border-dashed border-slate-200 bg-white px-4 py-5 text-center text-sm font-semibold text-slate-500 shadow-[0_10px_28px_rgba(15,23,42,0.08)]">
            Ketik minimal 4 digit nomor operator.
          </div>
        ) : loading ? (
          <div className="grid min-h-40 place-items-center rounded-2xl bg-white shadow-[0_10px_28px_rgba(15,23,42,0.08)]">
            <div className="inline-flex items-center gap-2 text-sm text-slate-500">
              <Loader2 className="h-4 w-4 animate-spin" />
              Memuat produk...
            </div>
          </div>
        ) : error ? (
          <div className="grid min-h-40 place-items-center rounded-2xl border border-dashed border-slate-200 bg-white px-4 text-center text-sm text-slate-500 shadow-[0_10px_28px_rgba(15,23,42,0.08)]">
            {error}
          </div>
        ) : (
          <UserProductGrid items={products} isLoggedIn={isLoggedIn} authToken={authToken} buyerRole={buyerRole} initialDest={normalizePhone(phone)} enableGuestHint={false} />
        )}
      </section>
    </div>
  );
}
