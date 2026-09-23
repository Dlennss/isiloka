import { NextResponse } from "next/server";
import { requireApiBase, forwardAuth } from "@/lib/adminApi";

export async function GET(req: Request) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const url = new URL(req.url);
  const qs = url.searchParams.toString();

  const r = await fetch(`${base}/v1/admin/app-ads?${qs}`, {
    method: "GET",
    headers: auth ? { Authorization: auth } : {},
    cache: "no-store",
  });

  const text = await r.text();
  return new NextResponse(text, {
    status: r.status,
    headers: { "Content-Type": "application/json" },
  });
}

export async function POST(req: Request) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const contentType = req.headers.get("content-type") || "";

  let r: Response;
  if (contentType.includes("multipart/form-data")) {
    const body = await req.formData();
    r = await fetch(`${base}/v1/admin/app-ads`, {
      method: "POST",
      headers: {
        ...(auth ? { Authorization: auth } : {}),
      },
      body,
    });
  } else {
    const body = await req.text();
    r = await fetch(`${base}/v1/admin/app-ads`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...(auth ? { Authorization: auth } : {}),
      },
      body,
    });
  }

  const text = await r.text();
  return new NextResponse(text, {
    status: r.status,
    headers: { "Content-Type": "application/json" },
  });
}
