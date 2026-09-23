"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { usePathname } from "next/navigation";
import { Bell } from "lucide-react";
import { Button } from "@/components/ui/button";
import { AppModal } from "@/components/ui/app-modal";
import { alertError, alertSuccess } from "@/components/ui/alerts";

const DEPOSIT_AUDIO_UNLOCK_KEY = "deposit_audio_unlocked";
const POLL_MS = 10_000;

type DepositRow = {
  id: number;
  member_id: number;
  member_nama?: string;
  status: string;
};

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  if (!t) return {};
  return { Authorization: `Bearer ${t}` };
}

function pendingIds(rows: DepositRow[]) {
  return rows
    .filter((row) => String(row.status || "").toLowerCase() === "pending")
    .map((row) => Number(row.id || 0))
    .filter((id) => id > 0);
}

export default function GlobalDepositNotifier() {
  const pathname = usePathname();
  const enabledArea =
    pathname.startsWith("/dashboard/admin") ||
    pathname.startsWith("/dashboard/operator") ||
    pathname.startsWith("/dashboard/wallet");

  const [audioUnlocked, setAudioUnlocked] = useState(false);
  const [promptOpen, setPromptOpen] = useState(false);
  const [pendingCount, setPendingCount] = useState(0);
  const pollInFlightRef = useRef(false);
  const intervalRef = useRef<number | null>(null);
  const pendingIdsRef = useRef<Set<number>>(new Set());
  const hasPolledRef = useRef(false);
  const audioRef = useRef<HTMLAudioElement | null>(null);

  const playAudio = useCallback(async () => {
    if (typeof window === "undefined") return false;
    if (!audioRef.current) {
      audioRef.current = new Audio("/audio/cash.mp3");
    }
    const audio = audioRef.current;
    audio.preload = "auto";
    audio.currentTime = 0;
    await audio.play();
    return true;
  }, []);

  const poll = useCallback(async () => {
    if (!enabledArea) return;
    if (pollInFlightRef.current) return;
    pollInFlightRef.current = true;

    try {
      const qs = new URLSearchParams();
      qs.set("status", "pending");
      qs.set("limit", "2");
      qs.set("offset", "0");

      const r = await fetch(`/api/admin/deposit/requests?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      const rows: DepositRow[] = Array.isArray(j.items) ? j.items : [];
      const nextIds = pendingIds(rows);
      const nextCount = nextIds.length;
      setPendingCount(nextCount);

      if (nextCount === 0) {
        setPromptOpen(false);
        pendingIdsRef.current = new Set();
        hasPolledRef.current = true;
        return;
      }

      if (!audioUnlocked) {
        setPromptOpen(true);
        pendingIdsRef.current = new Set(nextIds);
        hasPolledRef.current = true;
        return;
      }

      setPromptOpen(false);
      pendingIdsRef.current = new Set(nextIds);
      if (document.visibilityState === "visible") {
        void playAudio().catch(() => {
          setAudioUnlocked(false);
          setPromptOpen(true);
          if (typeof window !== "undefined") {
            window.localStorage.removeItem(DEPOSIT_AUDIO_UNLOCK_KEY);
          }
        });
      }
      hasPolledRef.current = true;
    } finally {
      pollInFlightRef.current = false;
    }
  }, [audioUnlocked, enabledArea, playAudio]);

  const unlockAudio = useCallback(async () => {
    try {
      await playAudio();
      setAudioUnlocked(true);
      setPromptOpen(false);
      if (typeof window !== "undefined") {
        window.localStorage.setItem(DEPOSIT_AUDIO_UNLOCK_KEY, "1");
      }
      await alertSuccess("Notifikasi deposit aktif.");
    } catch {
      setAudioUnlocked(false);
      if (typeof window !== "undefined") {
        window.localStorage.removeItem(DEPOSIT_AUDIO_UNLOCK_KEY);
      }
      setPromptOpen(true);
      await alertError("Browser memblokir suara. Tekan Aktifkan Suara lagi.");
    }
  }, [playAudio]);

  useEffect(() => {
    if (typeof window === "undefined") return;
    if (!enabledArea) {
      setAudioUnlocked(false);
      setPromptOpen(false);
      setPendingCount(0);
      pendingIdsRef.current = new Set();
      hasPolledRef.current = false;
      return;
    }

    setAudioUnlocked(window.localStorage.getItem(DEPOSIT_AUDIO_UNLOCK_KEY) === "1");
    setPromptOpen(false);
  }, [enabledArea]);

  useEffect(() => {
    if (!enabledArea) return;
    void poll();

    if (intervalRef.current) {
      window.clearInterval(intervalRef.current);
    }

    intervalRef.current = window.setInterval(() => {
      if (document.visibilityState !== "visible") return;
      void poll();
    }, POLL_MS);

    return () => {
      if (intervalRef.current) {
        window.clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [enabledArea, poll]);

  if (!enabledArea) return null;

  return (
    <AppModal
      open={promptOpen && pendingCount > 0 && !audioUnlocked}
      onClose={() => {
        if (audioUnlocked || pendingCount === 0) {
          setPromptOpen(false);
        }
      }}
      title="Aktifkan Notifikasi Suara"
      maxWidthClassName="max-w-md"
      hideCloseButton
      footer={
        <div className="flex items-center justify-end gap-2">
          <Button variant="primary" onClick={() => void unlockAudio()}>
            <Bell className="h-4 w-4" />
            Aktifkan Suara
          </Button>
        </div>
      }
    >
      <div className="space-y-3 text-sm text-slate-200">
        <p>Ada {pendingCount} request deposit pending.</p>
        <div className="rounded-xl border border-cyan-400/15 bg-cyan-400/5 px-3 py-3 text-xs text-cyan-100">
          Status ini akan disimpan di browser.
        </div>
      </div>
    </AppModal>
  );
}
