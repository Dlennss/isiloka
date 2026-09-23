export const runtime = "nodejs";

type TurnstileVerifyOk = { success: true };

type TurnstileVerifyFail = {
  success: false;
  errorCodes: string[];
  error?: string;
};

export type TurnstileVerifyResult = TurnstileVerifyOk | TurnstileVerifyFail;

type TurnstileRaw = {
  success: boolean;
  "error-codes"?: string[];
};

export async function verifyTurnstileToken(token: string, ip?: string): Promise<TurnstileVerifyResult> {
  const secret = process.env.TURNSTILE_SECRET_KEY;
  if (!secret) return { success: false, error: "TURNSTILE_SECRET_KEY missing", errorCodes: ["missing-secret"] };

  const form = new FormData();
  form.append("secret", secret);
  form.append("response", token);
  if (ip) form.append("remoteip", ip);

  const res = await fetch("https://challenges.cloudflare.com/turnstile/v0/siteverify", {
    method: "POST",
    body: form,
    cache: "no-store",
  });

  if (!res.ok) return { success: false, error: `turnstile http ${res.status}`, errorCodes: ["http-error"] };

  const json = (await res.json().catch(() => null)) as TurnstileRaw | null;
  if (!json || typeof json.success !== "boolean") {
    return { success: false, error: "turnstile bad json", errorCodes: ["bad-json"] };
  }

  if (json.success) return { success: true };
  return { success: false, errorCodes: Array.isArray(json["error-codes"]) ? json["error-codes"] : [] };
}
