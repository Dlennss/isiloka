"use client";

import Image from "next/image";
import Link from "next/link";
import { ArrowLeft, BadgeCheck, CheckCircle2, ShieldCheck, TrendingUp, WalletCards } from "lucide-react";
import { useRef, useState } from "react";

const kilatLevels = [
  {
    code: "start",
    name: "Kilat Start",
    short: "Start",
    image: "/agent-levels/kilat-start-badge.png",
    bg: "from-emerald-50 via-white to-lime-100",
    accent: "bg-emerald-600",
    text: "text-emerald-700",
    desc: "Level awal untuk agent baru. Limit dasar Rp 500.000 dan menjadi titik awal membangun riwayat pembayaran.",
    benefits: [
      ["Limit sampai Rp 500.000", "Agent baru mulai dari limit dasar Isiloka", WalletCards],
      ["3 pengajuan lunas untuk naik", "Setelah 3 pinjaman lunas tepat waktu, agent naik ke Kilat Plus", CheckCircle2],
      ["Bayar tepat waktu", "Riwayat lunas tanpa telat menjadi syarat naik level", BadgeCheck],
      ["Target berikutnya Rp 1.000.000", "Limit berikutnya terbuka setelah pembayaran konsisten", TrendingUp],
    ],
  },
  {
    code: "plus",
    name: "Kilat Plus",
    short: "Plus",
    image: "/agent-levels/kilat-plus-badge.png",
    bg: "from-sky-50 via-white to-blue-100",
    accent: "bg-blue-600",
    text: "text-blue-700",
    desc: "Level untuk agent yang sudah menyelesaikan 3 pengajuan tepat waktu. Limit naik menjadi Rp 1.000.000.",
    benefits: [
      ["Limit sampai Rp 1.000.000", "Terbuka setelah 3 pinjaman lunas tepat waktu", WalletCards],
      ["5 pengajuan lunas untuk maksimal", "Setelah total 5 pinjaman lunas tepat waktu, agent naik ke Kilat Elite", CheckCircle2],
      ["Tidak pernah jatuh tempo", "Pembayaran telat lebih dari 3 hari perlu perbaikan akun", BadgeCheck],
      ["Target berikutnya Rp 2.000.000", "Limit maksimal aktif setelah 5 pengajuan lunas", TrendingUp],
    ],
  },
  {
    code: "elite",
    name: "Kilat Elite",
    short: "Elite",
    image: "/agent-levels/kilat-elite-badge.png",
    bg: "from-purple-50 via-white to-yellow-100",
    accent: "bg-purple-600",
    text: "text-purple-700",
    desc: "Level tertinggi Isiloka setelah agent menyelesaikan 5 pengajuan tepat waktu. Limit maksimal Rp 2.000.000.",
    benefits: [
      ["Limit maksimal Rp 2.000.000", "Elite adalah batas tertinggi pinjaman saldo agent", WalletCards],
      ["Syarat 5 pengajuan lunas", "Semua dihitung dari pinjaman yang selesai tepat waktu", CheckCircle2],
      ["Akun paling sehat", "Level ini menandai agent yang stabil dan disiplin membayar", BadgeCheck],
      ["Layanan prioritas", "Cocok untuk agent dengan loyalitas dan omzet kuat", TrendingUp],
    ],
  },
];

function clampLevel(index: number) {
  return Math.max(0, Math.min(kilatLevels.length - 1, index));
}

export function UserAgentLevelPageContent({ initialLevelCode = "start" }: { initialLevelCode?: string }) {
  const normalizedInitialLevel = ["pro", "max"].includes(initialLevelCode) ? "elite" : initialLevelCode;
  const initialIndex = Math.max(0, kilatLevels.findIndex((level) => level.code === normalizedInitialLevel));
  const [activeIndex, setActiveIndex] = useState(initialIndex);
  const touchStartX = useRef<number | null>(null);
  const activeLevel = kilatLevels[activeIndex];
  const prevLevel = kilatLevels[clampLevel(activeIndex - 1)];
  const nextLevel = kilatLevels[clampLevel(activeIndex + 1)];

  function setLevel(index: number) {
    setActiveIndex(clampLevel(index));
  }

  function handleTouchEnd(clientX: number) {
    if (touchStartX.current === null) return;
    const delta = clientX - touchStartX.current;
    touchStartX.current = null;
    if (Math.abs(delta) < 42) return;
    setLevel(activeIndex + (delta < 0 ? 1 : -1));
  }

  return (
    <div className="mx-auto w-full max-w-md pb-24">
      <section className={`relative min-h-[520px] overflow-hidden bg-linear-to-br ${activeLevel.bg} px-4 pt-5 text-slate-700`}>
        <div className="pointer-events-none absolute -left-20 top-20 h-40 w-40 rounded-full bg-white/65 blur-2xl" />
        <div className="pointer-events-none absolute -right-16 top-20 h-44 w-44 rounded-full bg-white/70 blur-2xl" />

        <div className="relative z-10 flex items-center gap-3">
          <Link href="/user/saldo/kredit-agent" className={`grid h-10 w-10 shrink-0 place-items-center rounded-2xl bg-white/70 ${activeLevel.text} ring-1 ring-white`}>
            <ArrowLeft className="h-5 w-5" strokeWidth={2.4} />
          </Link>
          <div className="min-w-0 flex-1">
            <h1 className="truncate text-lg font-semibold text-slate-700">Isiloka Level</h1>
          </div>
        </div>

        <div
          className="relative z-10 mt-5 h-[190px]"
          onTouchStart={(event) => {
            touchStartX.current = event.touches[0]?.clientX ?? null;
          }}
          onTouchEnd={(event) => handleTouchEnd(event.changedTouches[0]?.clientX ?? 0)}
        >
          {activeIndex > 0 ? (
            <button type="button" onClick={() => setLevel(activeIndex - 1)} className="absolute -left-20 top-8 h-32 w-40 opacity-60" aria-label="Level sebelumnya">
              <Image src={prevLevel.image} alt={prevLevel.name} fill sizes="160px" className="object-contain" />
            </button>
          ) : null}
          {activeIndex < kilatLevels.length - 1 ? (
            <button type="button" onClick={() => setLevel(activeIndex + 1)} className="absolute -right-20 top-8 h-32 w-40 opacity-60" aria-label="Level berikutnya">
              <Image src={nextLevel.image} alt={nextLevel.name} fill sizes="160px" className="object-contain" />
            </button>
          ) : null}

          <div className="absolute left-1/2 top-0 h-48 w-64 -translate-x-1/2">
            <Image src={activeLevel.image} alt={activeLevel.name} fill priority sizes="256px" className="object-contain drop-shadow-[0_18px_18px_rgba(15,23,42,0.16)]" />
          </div>
        </div>

        <div className="relative z-10 grid grid-cols-3 gap-1 text-center">
          {kilatLevels.map((level, index) => (
            <button key={level.name} type="button" onClick={() => setLevel(index)} className={index === activeIndex ? "text-sm font-black text-slate-800" : "text-sm font-semibold text-slate-500"}>
              {level.short}
            </button>
          ))}
        </div>

        <div className="relative z-10 mt-4 px-5">
          <div className="absolute left-10 right-10 top-5 h-1 rounded-full bg-slate-300/70" />
          <div className="relative grid grid-cols-3">
            {kilatLevels.map((level, index) => (
              <button key={level.name} type="button" onClick={() => setLevel(index)} className="grid place-items-center">
                <span className={index === activeIndex ? "relative grid h-14 w-14 place-items-center rounded-full bg-white shadow-[0_10px_26px_rgba(15,23,42,0.18)] ring-4 ring-white" : "relative grid h-9 w-9 place-items-center rounded-full bg-white/70 opacity-80"}>
                  <Image src={level.image} alt="" fill sizes="56px" className="object-contain p-0.5" />
                </span>
              </button>
            ))}
          </div>
        </div>
      </section>

      <div className="-mt-24 px-3">
        <section className="relative z-20 overflow-hidden rounded-[28px] bg-white shadow-[0_22px_46px_rgba(15,23,42,0.14)]">
          <div className="border-b border-slate-100 px-5 py-5 text-center">
            <div className="relative mx-auto mb-2 h-14 w-14 rounded-2xl bg-slate-100 shadow-sm">
              <Image src={activeLevel.image} alt="" fill sizes="56px" className="object-contain p-0.5" />
            </div>
            <h2 className="text-2xl font-semibold text-slate-700">{activeLevel.name.replace("Kilat ", "")}</h2>
            <p className="mx-auto mt-2 max-w-[280px] text-sm font-medium leading-5 text-slate-500">{activeLevel.desc}</p>
          </div>

          <div className="space-y-4 px-5 py-5">
            {activeLevel.benefits.map(([title, desc, Icon]) => (
              <div key={title as string} className="flex items-start gap-4">
                <span className={`grid h-11 w-11 shrink-0 place-items-center rounded-full text-white ${activeLevel.accent}`}>
                  <Icon className="h-5 w-5" strokeWidth={2.4} />
                </span>
                <span className="min-w-0">
                  <span className="block text-base font-semibold leading-5 text-slate-700">{title as string}</span>
                  <span className="mt-1 block text-sm font-medium leading-5 text-slate-400">{desc as string}</span>
                </span>
              </div>
            ))}
          </div>
        </section>

        <section className="mt-3 rounded-[24px] border border-emerald-100 bg-emerald-50 p-4 text-[#087e8b]">
          <div className="flex items-center gap-3">
            <ShieldCheck className="h-8 w-8 shrink-0" strokeWidth={2.5} />
            <div>
              <p className="text-sm font-black">Naik Level Bertahap</p>
              <p className="mt-1 text-[11px] font-semibold leading-5 text-[#087e8b]/75">Level naik dari jumlah pengajuan yang sudah lunas tepat waktu: 3 kali menjadi Plus, 5 kali menjadi Elite.</p>
            </div>
          </div>
        </section>
      </div>
    </div>
  );
}
