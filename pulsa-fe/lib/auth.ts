export type RegisterBody = { email: string; password: string };
export type LoginBody = { email: string; password: string };
export type LoginResp = { token: string };
export type GoogleLoginBody = { email: string; nama?: string; google_id: string; id_token: string };
export type GoogleLoginResp = { token: string; role?: string; created?: boolean };

async function jsonFetch<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers || {}),
    },
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    const msg = (data && (data.error || data.message)) || `HTTP ${res.status}`;
    throw new Error(String(msg));
  }
  return data as T;
}

import { API_BASE } from './env';

export async function register(body: RegisterBody) {
  return jsonFetch<{ ok: boolean }>(`${API_BASE}/v1/auth/register`, {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

export async function login(body: LoginBody) {
  return jsonFetch<LoginResp>(`${API_BASE}/v1/auth/login`, {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

export async function loginGoogle(body: GoogleLoginBody) {
  return jsonFetch<GoogleLoginResp>(`${API_BASE}/v1/auth/google`, {
    method: 'POST',
    body: JSON.stringify(body),
  });
}
