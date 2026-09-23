export type JwtClaims = { sub?: number | string; role?: string; iat?: number; exp?: number };

function base64UrlToString(b64url: string): string {
  const pad = b64url.length % 4;
  const b64 = (b64url + (pad ? '='.repeat(4 - pad) : '')).replace(/-/g, '+').replace(/_/g, '/');
  if (typeof window === 'undefined') return Buffer.from(b64, 'base64').toString('utf-8');
  return decodeURIComponent(escape(atob(b64))); // aman untuk unicode sederhana
}

export function decodeJwt(token: string): JwtClaims | null {
  try {
    const [, payload] = token.split('.');
    if (!payload) return null;
    return JSON.parse(base64UrlToString(payload));
  } catch {
    return null;
  }
}

export function isJwtValid(token?: string | null): boolean {
  const claims = token ? decodeJwt(token) : null;
  if (!claims) return false;
  if (!claims.exp) return true;
  return claims.exp * 1000 > Date.now();
}
