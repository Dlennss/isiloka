export function requireApiBase(): string {
  const base = process.env.NEXT_PUBLIC_API_BASE || process.env.API_BASE || "http://127.0.0.1:8083";
  return base.replace(/\/+$/, "");
}

export function forwardAuth(headers: Headers): string | null {
  const auth = headers.get("authorization");
  if (!auth) return null;
  return auth;
}
