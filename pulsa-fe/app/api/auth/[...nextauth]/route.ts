import NextAuth from "next-auth";
import { NextResponse } from "next/server";
import { authOptions } from "@/lib/nextauth";

const handler = NextAuth(authOptions);

function isGoogleConfigured() {
  return Boolean(
    (process.env.GOOGLE_CLIENT_ID || process.env.AUTH_GOOGLE_ID || "").trim()
      && (process.env.GOOGLE_CLIENT_SECRET || process.env.AUTH_GOOGLE_SECRET || "").trim(),
  );
}

function redirectGoogleSetup(req: Request) {
  const url = new URL("/login", req.url);
  url.searchParams.set("google", "not_configured");
  return NextResponse.redirect(url);
}

export function GET(req: Request, ctx: unknown) {
  const url = new URL(req.url);
  if (url.pathname.endsWith("/api/auth/callback/google") && !isGoogleConfigured()) {
    return redirectGoogleSetup(req);
  }
  return handler(req, ctx);
}

export function POST(req: Request, ctx: unknown) {
  return handler(req, ctx);
}
