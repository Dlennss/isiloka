"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { Check, ChevronDown, Search } from "lucide-react";
import { Input } from "@/components/ui/input";

type ProductOption = {
  id: number | string;
  sku: string;
  nama: string;
  aktif?: boolean;
};

type ProductSearchSelectProps = {
  items: ProductOption[];
  value: string;
  onChange: (value: string) => void;
  loading?: boolean;
  placeholder?: string;
  disabled?: boolean;
  className?: string;
  itemLabelMode?: "product" | "single";
};

type DropdownPosition = {
  left: number;
  top: number;
  width: number;
  maxHeight: number;
};

function cn(...parts: Array<string | false | null | undefined>) {
  return parts.filter(Boolean).join(" ");
}

export function ProductSearchSelect({
  items,
  value,
  onChange,
  loading = false,
  placeholder = "Pilih produk",
  disabled = false,
  className,
  itemLabelMode = "product",
}: ProductSearchSelectProps) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [position, setPosition] = useState<DropdownPosition | null>(null);
  const rootRef = useRef<HTMLDivElement | null>(null);

  const selected = useMemo(
    () => items.find((item) => String(item.id) === value) ?? null,
    [items, value],
  );

  const filteredItems = useMemo(() => {
    const keyword = query.trim().toLowerCase();
    if (!keyword) return items;
    return items.filter((item) => {
      const haystack = `${item.sku} ${item.nama}`.toLowerCase();
      return haystack.includes(keyword);
    });
  }, [items, query]);

  const isSingleLine = itemLabelMode === "single";

  const updatePosition = useCallback(() => {
    const rect = rootRef.current?.getBoundingClientRect();
    if (!rect) return;

    const viewportHeight = window.innerHeight;
    const viewportWidth = window.innerWidth;
    const estimatedDropdownHeight = 360;
    const gap = 8;
    const spaceBelow = viewportHeight - rect.bottom;
    const spaceAbove = rect.top;
    const openUpward = spaceBelow < estimatedDropdownHeight && spaceAbove > spaceBelow;
    const maxWidth = Math.min(rect.width, viewportWidth - 16);
    const left = Math.min(Math.max(rect.left, 8), viewportWidth - maxWidth - 8);
    const top = openUpward
      ? Math.max(8, rect.top - estimatedDropdownHeight - gap)
      : Math.min(viewportHeight - gap, rect.bottom + gap);
    const maxHeight = openUpward
      ? Math.max(160, Math.min(360, rect.top - 12))
      : Math.max(160, Math.min(360, viewportHeight - top - 12));

    setPosition({
      left,
      top,
      width: maxWidth,
      maxHeight,
    });
  }, []);

  useEffect(() => {
    if (!open) return;

    const rafId = window.requestAnimationFrame(updatePosition);

    function onPointerDown(event: MouseEvent) {
      const target = event.target as Node | null;
      if (!rootRef.current || !target) return;
      if (!rootRef.current.contains(target)) {
        const panel = document.getElementById("product-search-select-portal");
        if (panel?.contains(target)) return;
        setOpen(false);
      }
    }

    function onKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setOpen(false);
      }
    }

    function onViewportChange() {
      updatePosition();
    }

    document.addEventListener("mousedown", onPointerDown);
    document.addEventListener("keydown", onKeyDown);
    window.addEventListener("resize", onViewportChange);
    window.addEventListener("scroll", onViewportChange, true);
    return () => {
      window.cancelAnimationFrame(rafId);
      document.removeEventListener("mousedown", onPointerDown);
      document.removeEventListener("keydown", onKeyDown);
      window.removeEventListener("resize", onViewportChange);
      window.removeEventListener("scroll", onViewportChange, true);
    };
  }, [open, updatePosition]);

  return (
    <div ref={rootRef} className={cn("relative", className)}>
      <button
        type="button"
        disabled={disabled}
        onClick={() => {
          if (disabled) return;
          if (!open) {
            updatePosition();
          }
          setOpen((prev) => !prev);
          setQuery("");
        }}
        className={cn(
          "flex h-11 w-full items-center justify-between rounded-lg border border-white/15 bg-slate-950/70 px-3 text-left text-sm text-slate-100 outline-hidden ring-0 transition",
          "focus:border-cyan-400/60 disabled:cursor-not-allowed disabled:opacity-60",
        )}
      >
        <span className={selected ? "truncate text-slate-100" : "truncate text-slate-500"}>
          {loading
            ? "Loading produk..."
            : selected
              ? isSingleLine
                ? `${selected.nama}${selected.aktif === false ? " (nonaktif)" : ""}`
                : `${selected.sku} - ${selected.nama}${selected.aktif === false ? " (nonaktif)" : ""}`
              : placeholder}
        </span>
        <ChevronDown className={cn("h-4 w-4 shrink-0 text-slate-400 transition", open && "rotate-180")} />
      </button>

      {open && position && typeof document !== "undefined"
        ? createPortal(
            <div
              id="product-search-select-portal"
              className="fixed z-70 overflow-hidden rounded-xl border border-white/15 bg-slate-950 shadow-2xl"
              style={{
                left: position.left,
                top: position.top,
                width: position.width,
              }}
            >
              <div className="border-b border-white/10 p-2">
                <div className="relative">
                  <Search className="pointer-events-none absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-slate-500" />
                  <Input
                    value={query}
                    onChange={(event) => setQuery(event.target.value)}
                    placeholder={isSingleLine ? "Cari nama" : "Cari SKU / nama produk"}
                    className="h-10 border-white/10 bg-slate-900 pl-9 text-slate-100 placeholder:text-slate-500"
                    autoFocus
                  />
                </div>
              </div>

              <div
                className="overflow-y-auto py-1"
                style={{
                  maxHeight: position.maxHeight,
                }}
              >
                {filteredItems.length ? (
                  filteredItems.map((item) => {
                    const isActive = String(item.id) === value;
                    return (
                      <button
                        key={item.id}
                        type="button"
                        onClick={() => {
                          onChange(String(item.id));
                          setOpen(false);
                          setQuery("");
                        }}
                        className={cn(
                          "flex w-full items-start justify-between gap-3 px-3 py-2 text-left text-sm transition hover:bg-white/8",
                          isActive ? "bg-cyan-500/12 text-cyan-100" : "text-slate-200",
                        )}
                      >
                        <span className="min-w-0">
                          {isSingleLine ? (
                            <span className="block truncate font-medium">
                              {item.nama} {item.aktif === false ? "(nonaktif)" : ""}
                            </span>
                          ) : (
                            <>
                              <span className="block truncate font-medium">{item.sku}</span>
                              <span className="block truncate text-xs text-slate-400">
                                {item.nama} {item.aktif === false ? "(nonaktif)" : ""}
                              </span>
                            </>
                          )}
                        </span>
                        {isActive ? <Check className="mt-0.5 h-4 w-4 shrink-0" /> : null}
                      </button>
                    );
                  })
                ) : (
                  <div className="px-3 py-3 text-sm text-slate-400">
                    {isSingleLine ? "Data tidak ditemukan." : "Produk tidak ditemukan."}
                  </div>
                )}
              </div>
            </div>,
            document.body,
          )
        : null}
    </div>
  );
}
