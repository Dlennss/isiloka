import { NextResponse } from "next/server";
import { requireApiBase } from "@/lib/adminApi";
import { getBackendAuthorization } from "@/lib/server-auth";

export async function GET(req: Request) {
  const base = requireApiBase();
  const auth = await getBackendAuthorization(req);
  const query = new URL(req.url).searchParams.toString();
  try {
    const response = await fetch(`${base}/v1/admin/agent-credit/team-activity?${query}`, {
      headers: auth ? { Authorization: auth } : {},
      cache: "no-store",
      signal: AbortSignal.timeout(15000),
    });
    return new NextResponse(await response.text(), {
      status: response.status,
      headers: { "Content-Type": "application/json" },
    });
  } catch {
    return NextResponse.json(
      { ok: false, error: "Backend pemantauan tim tidak dapat dihubungi" },
      { status: 502 },
    );
  }
}
