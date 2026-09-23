"use client";

import { useEffect, useRef, useCallback } from "react";
import { useRouter } from "next/navigation";

const EVENTS = ["mousedown", "mousemove", "keydown", "scroll", "touchstart"];

export function IdleLogout() {
  const router = useRouter();
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const seconds = Number(process.env.NEXT_PUBLIC_IDLE_LOGOUT_SECONDS || "300");

  const logout = useCallback(() => {
    localStorage.removeItem("auth_token");
    localStorage.removeItem("user_name");
    localStorage.removeItem("user_email");
    localStorage.removeItem("user_role");
    router.replace("/login");
  }, [router]);

  const resetTimer = useCallback(() => {
    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(logout, seconds * 1000);
  }, [logout, seconds]);

  useEffect(() => {
    if (!localStorage.getItem("auth_token")) return;
    resetTimer();
    EVENTS.forEach((e) => window.addEventListener(e, resetTimer, { passive: true }));
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
      EVENTS.forEach((e) => window.removeEventListener(e, resetTimer));
    };
  }, [resetTimer]);

  return null;
}
