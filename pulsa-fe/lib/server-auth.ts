import { cookies } from "next/headers";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/nextauth";
import { decodeJwt } from "@/lib/jwt";
import type { UserSession } from "@/components/user/types";

export type AppServerSession = {
  user?: UserSession;
  backendToken?: string;
};

export const PK_AUTH_COOKIE = "pk_auth_token";

export async function getAppServerSession(): Promise<AppServerSession | null> {
  const token = (await cookies()).get(PK_AUTH_COOKIE)?.value || "";
  const claims = decodeJwt(token);
  const tokenIsActive = Boolean(token && claims && (!claims.exp || claims.exp * 1000 > Date.now()));
  if (tokenIsActive && claims) {
    const role = typeof claims.role === "string" ? claims.role : "user";
    return {
      backendToken: token,
      user: {
        name: "",
        email: "",
        role,
      },
    };
  }

  // Login tetap harus dapat dibuka walau konfigurasi NextAuth di server belum lengkap
  // atau cookie sesi lama sudah tidak lagi dapat dibaca.
  try {
    const session = (await getServerSession(authOptions)) as AppServerSession | null;
    return session?.backendToken ? session : null;
  } catch (error) {
    console.error("[auth] gagal membaca sesi NextAuth", error);
    return null;
  }
}

export async function getBackendAuthorization(req?: Request): Promise<string> {
  const session = await getAppServerSession();
  if (session?.backendToken) return `Bearer ${session.backendToken}`;

  return String(req?.headers.get("authorization") || "").trim();
}
