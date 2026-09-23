"use client";

import Link from "next/link";
import { ArrowRightLeft, Boxes, ShieldAlert } from "lucide-react";
import DashboardProfileCard from "@/components/dashboard/DashboardProfileCard";

const cards = [
  {
    href: "/dashboard/operator/transaksi/member",
    title: "Transaksi Member",
    description: "Lihat, filter, sukseskan, dan batalkan transaksi member.",
    icon: ArrowRightLeft,
  },
  {
    href: "/dashboard/operator/transaksi/provider",
    title: "Transaksi Provider",
    description: "Pantau transaksi provider yang terkait proses member.",
    icon: Boxes,
  },
  {
    href: "/dashboard/operator/transaksi/status-mismatch",
    title: "Audit Status Mismatch",
    description: "Lihat transaksi member dan provider yang status akhirnya tidak sinkron.",
    icon: ShieldAlert,
  },
  {
    href: "/dashboard/operator/transaksi/member-status-logs",
    title: "Audit Status Saya",
    description: "Lihat perubahan status transaksi yang kamu lakukan bulan ini.",
    icon: ShieldAlert,
  },
];

export default function OperatorDashboardPage() {
  return (
    <div className="space-y-6 p-2">
      <DashboardProfileCard
        role="operator_trx"
        description="Akun operator transaksi aktif untuk memantau, membatalkan, dan menyukseskan transaksi member serta audit status terkait."
      />

      <div className="grid gap-4 md:grid-cols-2">
        {cards.map((card) => {
          const Icon = card.icon;
          return (
            <Link
              key={card.href}
              href={card.href}
              className="rounded-2xl border border-white/10 bg-linear-to-br from-slate-900/85 via-slate-900/70 to-cyan-950/25 p-5 shadow-[0_16px_40px_-24px_rgba(0,0,0,0.8)] transition hover:border-cyan-400/35 hover:shadow-[0_22px_54px_-24px_rgba(34,211,238,0.28)]"
            >
              <div className="flex items-start justify-between gap-3">
                <div>
                  <div className="text-lg font-semibold text-white">{card.title}</div>
                  <div className="mt-1 text-sm text-white/65">{card.description}</div>
                </div>
                <div className="grid h-11 w-11 place-items-center rounded-2xl border border-cyan-400/20 bg-cyan-400/10 text-cyan-300">
                  <Icon className="h-5 w-5" />
                </div>
              </div>
            </Link>
          );
        })}
      </div>
    </div>
  );
}
