import type { NextAuthOptions, User } from "next-auth";
import GoogleProvider from "next-auth/providers/google";
import CredentialsProvider from "next-auth/providers/credentials";

type BackendGoogleLoginResp = {
  ok?: boolean;
  token?: string;
  role?: string;
  error?: string;
};

type BackendPasswordLoginResp = {
  ok?: boolean;
  token?: string;
  error?: string;
};

type BackendMeResp = {
  ok?: boolean;
  role?: string;
};

type SessionWithBackend = {
  backendToken?: string;
  user?: Record<string, unknown>;
};

type UserWithBackend = User & {
  backendToken?: string;
  backendRole?: string;
};

function isGoogleLoginEnabled() {
  return String(process.env.GOOGLE_LOGIN_ENABLED ?? "true").toLowerCase() === "true";
}

function googleClientID() {
  return (process.env.GOOGLE_CLIENT_ID || process.env.AUTH_GOOGLE_ID || "").trim();
}

function googleClientSecret() {
  return (process.env.GOOGLE_CLIENT_SECRET || process.env.AUTH_GOOGLE_SECRET || "").trim();
}

function isGoogleProviderConfigured() {
  return Boolean(googleClientID() && googleClientSecret());
}

const apiBase = () => process.env.NEXT_PUBLIC_API_BASE || process.env.API_BASE || "http://127.0.0.1:8083";

async function syncGoogleToBackend(email: string, nama: string, googleID: string, idToken: string): Promise<BackendGoogleLoginResp> {
  const r = await fetch(`${apiBase()}/v1/auth/google`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      email: email.trim().toLowerCase(),
      nama: nama.trim(),
      google_id: googleID.trim(),
      id_token: idToken.trim(),
    }),
    cache: "no-store",
  });
  const body = (await r.json().catch(() => ({}))) as BackendGoogleLoginResp;
  if (!r.ok || !body?.ok || !body?.token) {
    console.error("[auth][google] backend sync failed", {
      status: r.status,
      email,
      hasGoogleID: Boolean(googleID),
      hasIDToken: Boolean(idToken),
      body,
    });
  }
  return body;
}

async function loginPasswordToBackend(email: string, password: string): Promise<BackendPasswordLoginResp> {
  const r = await fetch(`${apiBase()}/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email: email.trim().toLowerCase(), password }),
    cache: "no-store",
  });
  return (await r.json().catch(() => ({}))) as BackendPasswordLoginResp;
}

async function getRoleByToken(token: string): Promise<string> {
  const r = await fetch(`${apiBase()}/v1/auth/me`, {
    cache: "no-store",
    headers: { Authorization: `Bearer ${token}` },
  });
  const j = (await r.json().catch(() => ({}))) as BackendMeResp;
  return String(j?.role || "member").toLowerCase();
}

const providers: NextAuthOptions["providers"] = [
  CredentialsProvider({
    name: "credentials",
    credentials: {
      email: { label: "Email", type: "email" },
      password: { label: "Password", type: "password" },
    },
    async authorize(credentials) {
      const email = String(credentials?.email || "").trim().toLowerCase();
      const password = String(credentials?.password || "");
      if (!email || !password) return null;

      try {
        const login = await loginPasswordToBackend(email, password);
        if (!login.ok || !login.token) return null;
        const role = await getRoleByToken(login.token);
        return {
          id: email,
          email,
          name: email,
          backendToken: login.token,
          backendRole: role,
        };
      } catch {
        return null;
      }
    },
  }),
];

if (isGoogleLoginEnabled() && isGoogleProviderConfigured()) {
  providers.push(
    GoogleProvider({
      clientId: googleClientID(),
      clientSecret: googleClientSecret(),
      httpOptions: {
        timeout: 15000,
      },
    }),
  );
}

export const authOptions: NextAuthOptions = {
  secret: process.env.AUTH_SECRET,
  providers,
  pages: {
    signIn: "/login",
  },
  session: {
    strategy: "jwt",
  },
  callbacks: {
    async signIn({ user, account, profile }) {
      if (account?.provider !== "google") return true;
      if (!isGoogleLoginEnabled()) return false;

      const email = (user.email || "").trim().toLowerCase();
      const nama = (user.name || "").trim();
      const googleID = String((profile as { sub?: unknown } | undefined)?.sub || "").trim();
      const idToken = String((account as { id_token?: unknown } | undefined)?.id_token || "").trim();
      if (!email || !googleID || !idToken) {
        console.error("[auth][google] missing callback payload", {
          hasEmail: Boolean(email),
          hasGoogleID: Boolean(googleID),
          hasIDToken: Boolean(idToken),
          provider: account?.provider,
        });
        return false;
      }

      try {
        const backend = await syncGoogleToBackend(email, nama, googleID, idToken);
        if (!backend.ok || !backend.token) {
          console.error("[auth][google] access denied by backend", {
            email,
            role: backend?.role,
            error: backend?.error,
          });
          return false;
        }
        const u = user as unknown as UserWithBackend;
        u.backendToken = backend.token;
        u.backendRole = backend.role || "user";
        return true;
      } catch (error) {
        console.error("[auth][google] unexpected signIn error", error);
        return false;
      }
    },
    async jwt({ token, user }) {
      if (user) {
        const u = user as unknown as UserWithBackend;
        if (typeof u.backendToken === "string" && u.backendToken) token.backendToken = u.backendToken;
        if (typeof u.backendRole === "string" && u.backendRole) token.role = u.backendRole;
      }
      return token;
    },
    async session({ session, token }) {
      const s = session as unknown as SessionWithBackend;
      if (typeof token.backendToken === "string") s.backendToken = token.backendToken;
      if (s.user && typeof token.role === "string") {
        s.user.role = token.role;
      }
      return session;
    },
  },
};
