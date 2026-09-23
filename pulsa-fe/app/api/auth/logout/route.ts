import { NextResponse } from "next/server";
import { PK_AUTH_COOKIE } from "@/lib/server-auth";

export async function POST() {
  const response = NextResponse.json({ ok: true });
  response.cookies.set(PK_AUTH_COOKIE, "", {
    httpOnly: true,
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    path: "/",
    maxAge: 0,
  });
  return response;
}
