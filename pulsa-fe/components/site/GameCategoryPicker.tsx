"use client";

import Image from "next/image";
import Link from "next/link";
import { ChevronRight, Gamepad2, UserRound } from "lucide-react";

type GameProviderItem = {
  id: number;
  name: string;
  href: string;
  logo?: {
    src: string;
    alt: string;
  } | null;
};

type GameCategoryPickerProps = {
  title: string;
  items: GameProviderItem[];
};

function getTheme(name: string) {
  const value = name.toLowerCase();
  if (value.includes("mobile")) return "from-[#251405] via-[#13464b] to-[#16a34a]";
  if (value.includes("free fire")) return "from-[#09090b] via-[#13464b] to-[#f59e0b]";
  if (value.includes("magic")) return "from-[#243c5a] via-[#087e8b] to-[#f6b49d]";
  if (value.includes("call of duty")) return "from-[#020617] via-[#13464b] to-[#334155]";
  if (value.includes("genshin")) return "from-[#1b1030] via-[#13464b] to-[#16a34a]";
  if (value.includes("pubg")) return "from-[#111827] via-[#13464b] to-[#facc15]";
  if (value.includes("roblox")) return "from-[#020617] via-[#13464b] to-[#10b981]";
  if (value.includes("point")) return "from-[#020617] via-[#13464b] to-[#0f766e]";
  return "from-[#13464b] via-[#087e8b] to-[#22c55e]";
}

export function GameCategoryPicker({ title, items }: GameCategoryPickerProps) {
  return (
    <div className="space-y-3 pb-28">
      <section className="overflow-hidden rounded-[28px] border border-emerald-950/5 bg-linear-to-br from-white via-emerald-50/75 to-lime-50/70 p-3 shadow-[0_18px_42px_rgba(6,78,59,0.12)]">
        <div className="flex items-center justify-between gap-3 px-1 pb-3 pt-1">
          <div>
            <h2 className="text-base font-black tracking-tight text-slate-950">Pilih Game</h2>
            <p className="mt-0.5 text-xs font-semibold text-slate-500">{title} favorit kamu</p>
          </div>
          <span className="grid h-9 w-9 shrink-0 place-items-center rounded-2xl bg-[#13464b] text-lime-200 shadow-[0_10px_22px_rgba(5,46,38,0.18)]">
            <Gamepad2 className="h-4.5 w-4.5" />
          </span>
        </div>

        {items.length === 0 ? (
          <div className="mt-4 grid min-h-40 place-items-center rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 text-center text-sm text-slate-500">
            Belum ada penyedia game aktif.
          </div>
        ) : (
          <div className="grid grid-cols-2 gap-3">
            {items.map((item) => {
              const isPortraitArtwork = item.logo?.src.includes("banner_magic_chess");
              const isGenshinLogo = item.logo?.src.includes("genshin_impact_logo");
              const theme = getTheme(item.name);
              return (
                <Link
                  key={item.id}
                  href={item.href}
                  prefetch={false}
                  aria-label={item.name}
                  className={`group relative isolate min-h-[174px] overflow-hidden rounded-[24px] border border-white/65 bg-linear-to-br ${theme} p-3 text-left shadow-[0_16px_32px_rgba(6,78,59,0.16)] ring-1 ring-emerald-950/[0.03] transition duration-300 hover:-translate-y-1 hover:shadow-[0_22px_42px_rgba(6,78,59,0.23)] focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-lime-200`}
                >
                  <div className="absolute -right-10 -top-10 h-28 w-28 rounded-full bg-lime-300/30 blur-2xl transition duration-300 group-hover:bg-lime-200/45" />
                  <div className="absolute -bottom-12 left-0 h-24 w-28 rounded-full bg-white/15 blur-2xl" />
                  <div className="absolute inset-0 bg-[radial-gradient(circle_at_18%_10%,rgba(255,255,255,0.22),transparent_28%),linear-gradient(180deg,rgba(255,255,255,0.08),rgba(255,255,255,0.02))]" />
                  <div className="absolute inset-0 opacity-[0.16] bg-[repeating-radial-gradient(circle_at_0_100%,rgba(255,255,255,0.72)_0,rgba(255,255,255,0.72)_1px,transparent_1px,transparent_11px)] bg-size-[150%_120%]" />

                  <div className="relative grid h-[104px] place-items-center overflow-hidden rounded-[20px] bg-black/18 ring-1 ring-white/12">
                    {item.logo?.src ? (
                      <>
                        {isPortraitArtwork ? (
                          <Image
                            src={item.logo.src}
                            alt=""
                            width={220}
                            height={280}
                            aria-hidden="true"
                            className="absolute inset-0 h-full w-full scale-110 object-cover opacity-45 blur-lg"
                          />
                        ) : null}
                        <div className={`relative z-0 grid h-full w-full place-items-center overflow-hidden ${isPortraitArtwork ? "bg-white/8" : ""}`}>
                          <Image
                            src={item.logo.src}
                            alt={item.logo.alt || item.name}
                            width={220}
                            height={140}
                            className={`${isPortraitArtwork ? "h-full w-auto max-w-full object-contain group-hover:scale-105" : isGenshinLogo ? "max-h-[94%] max-w-[96%] object-contain brightness-110 contrast-110 group-hover:scale-105" : "max-h-[82%] max-w-[86%] object-contain group-hover:scale-105"} drop-shadow-[0_10px_18px_rgba(0,0,0,0.22)] transition duration-300`}
                          />
                        </div>
                      </>
                    ) : (
                      <div className="grid h-full w-full place-items-center bg-linear-to-br from-[#13464b] to-[#10b981] text-white">
                        <UserRound className="h-8 w-8" />
                      </div>
                    )}
                  </div>

                  <div className="relative mt-3 flex min-h-[38px] items-end justify-between gap-2">
                    <p className="line-clamp-2 text-[11px] font-black uppercase leading-tight text-white drop-shadow-sm">
                      {item.name}
                    </p>
                    <span className="grid h-7 w-7 shrink-0 place-items-center rounded-full bg-lime-300 text-[#13464b] shadow-[0_8px_16px_rgba(5,46,38,0.18)] transition duration-300 group-hover:translate-x-0.5">
                      <ChevronRight className="h-4 w-4" strokeWidth={2.4} />
                    </span>
                  </div>
                </Link>
              );
            })}
          </div>
        )}
      </section>
    </div>
  );
}
