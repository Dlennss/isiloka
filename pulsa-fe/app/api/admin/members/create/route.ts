import { NextResponse } from "next/server";
import { requireApiBase } from "@/lib/adminApi";
import { getBackendAuthorization } from "@/lib/server-auth";

type RegisterReq = {
  email: string;
  nama: string;
  phone?: string;
  password: string;
  pin?: string;
  role?: string;
  retail_agent_commission_rp?: number;
  retail_master_commission_rp?: number;
  h2h_agent_commission_rp?: number;
  h2h_master_commission_rp?: number;
  fee_dana?: number;
  fee_gopay?: number;
  fee_linkaja?: number;
  fee_ovo?: number;
  fee_shopee?: number;
  fee_bank?: number;
  fee_lainnya?: number;
};

function isH2HRole(role: string) {
  return role === "member" || role === "agent_member" || role === "master_member";
}

function readNumber(data: Record<string, unknown>, key: string) {
  const value = data[key];
  return typeof value === "number" && Number.isFinite(value) ? value : undefined;
}

export async function POST(req: Request) {
  const base = requireApiBase();
  const auth = await getBackendAuthorization(req);
  const bodyText = await req.text();

  let payload: RegisterReq;
  try {
    payload = JSON.parse(bodyText) as RegisterReq;
  } catch {
    return NextResponse.json({ ok: false, error: "invalid json" }, { status: 400 });
  }

  const role = String(payload.role || "member").trim().toLowerCase();

  if (isH2HRole(role)) {
    const adminToken = process.env.ADMIN_TOKEN || "";
    if (!adminToken) {
      return NextResponse.json(
        {
          ok: false,
          error:
            "ADMIN_TOKEN belum diisi. Akun H2H butuh ADMIN_TOKEN agar fee awal dipasang benar, jadi proses tidak dibuat sukses palsu.",
        },
        { status: 500 },
      );
    }

    const beBody = JSON.stringify({
      email: payload.email,
      nama: payload.nama,
      phone: payload.phone,
      password: payload.password,
      pin: payload.pin,
      role,
      retail_agent_commission_rp: payload.retail_agent_commission_rp,
      retail_master_commission_rp: payload.retail_master_commission_rp,
      h2h_agent_commission_rp: payload.h2h_agent_commission_rp,
      h2h_master_commission_rp: payload.h2h_master_commission_rp,
      fee_dana: payload.fee_dana,
      fee_gopay: payload.fee_gopay,
      fee_linkaja: payload.fee_linkaja,
      fee_ovo: payload.fee_ovo,
      fee_shopee: payload.fee_shopee,
      fee_bank: payload.fee_bank,
      fee_lainnya: payload.fee_lainnya,
    });

    const r = await fetch(`${base}/v1/auth/register`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Admin-Token": adminToken,
        ...(auth ? { Authorization: auth } : {}),
      },
      body: beBody,
      cache: "no-store",
    });

    const text = await r.text();
    return new NextResponse(text, {
      status: r.status,
      headers: { "Content-Type": "application/json" },
    });
  }

  if (!auth) {
    return NextResponse.json(
      { ok: false, error: "Sesi admin tidak ditemukan. Silakan login ulang." },
      { status: 401 },
    );
  }

  const beBody = JSON.stringify({
    email: payload.email,
    nama: payload.nama,
    password: payload.password,
    pin: payload.pin,
    role,
    aktif: true,
  });

  const r = await fetch(`${base}/v1/admin/users`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: auth,
    },
    body: beBody,
    cache: "no-store",
  });

  const text = await r.text();
  let data: Record<string, unknown> | null = null;
  try {
    data = text ? (JSON.parse(text) as Record<string, unknown>) : null;
  } catch {
    return new NextResponse(text, {
      status: r.status,
      headers: { "Content-Type": "application/json" },
    });
  }

  if (r.ok && data?.ok) {
    const memberID = readNumber(data, "member_id") ?? readNumber(data, "id");
    return NextResponse.json({ ...data, member_id: memberID }, { status: r.status });
  }

  return NextResponse.json(data ?? { ok: false, error: "empty response" }, { status: r.status });
}
