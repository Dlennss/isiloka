import { NextResponse } from "next/server";
import { requireApiBase } from "@/lib/adminApi";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

const PASSTHROUGH_HEADERS = [
  "content-type",
  "x-callback-token",
  "x-pulsa24jam-token",
  "cf-connecting-ip",
  "x-forwarded-for",
  "x-real-ip",
  "user-agent",
] as const;

function buildHeaders(req: Request): Headers {
  const headers = new Headers();

  for (const key of PASSTHROUGH_HEADERS) {
    const value = req.headers.get(key);
    if (value) {
      headers.set(key, value);
    }
  }

  return headers;
}

async function proxyPulsa24JamWebhook(req: Request, method: "GET" | "POST") {
  const base = requireApiBase();
  const incomingUrl = new URL(req.url);
  const qs = incomingUrl.searchParams.toString();
  const target = `${base}/v1/webhook/pulsa24jam${qs ? `?${qs}` : ""}`;

  const res = await fetch(target, {
    method,
    headers: buildHeaders(req),
    body: method === "POST" ? await req.text() : undefined,
    cache: "no-store",
  });

  const text = await res.text();
  return new NextResponse(text, {
    status: res.status,
    headers: {
      "Content-Type": res.headers.get("content-type") || "application/json",
    },
  });
}

export async function GET(req: Request) {
  return proxyPulsa24JamWebhook(req, "GET");
}

export async function POST(req: Request) {
  return proxyPulsa24JamWebhook(req, "POST");
}
