import { NextResponse } from "next/server";
import { requireApiBase } from "@/lib/adminApi";

export async function GET(req: Request) {
  const base = requireApiBase();
  const url = new URL(req.url);
  const search = url.searchParams.toString();
  const target = `${base}/v1/app/produk${search ? `?${search}` : ""}`;

  const r = await fetch(target, {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
    cache: "no-store",
  });

  const text = await r.text();
  return new NextResponse(text, {
    status: r.status,
    headers: { "Content-Type": "application/json" },
  });
}
