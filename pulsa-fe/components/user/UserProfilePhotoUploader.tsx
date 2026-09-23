"use client";

import { useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { Camera } from "lucide-react";

type UserProfilePhotoUploaderProps = {
  name: string;
  email: string;
  phone: string;
  initials: string;
  profilePhotoURL?: string;
};

function resizeProfilePhoto(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    if (!file.type.startsWith("image/")) {
      reject(new Error("File harus berupa gambar."));
      return;
    }

    const reader = new FileReader();
    reader.onerror = () => reject(new Error("Gagal membaca foto profil."));
    reader.onload = () => {
      const img = new Image();
      img.onerror = () => reject(new Error("Foto profil tidak valid."));
      img.onload = () => {
        const maxSize = 512;
        const scale = Math.min(1, maxSize / Math.max(img.width, img.height));
        const width = Math.max(1, Math.round(img.width * scale));
        const height = Math.max(1, Math.round(img.height * scale));
        const canvas = document.createElement("canvas");
        canvas.width = width;
        canvas.height = height;
        const ctx = canvas.getContext("2d");
        if (!ctx) {
          reject(new Error("Browser tidak bisa memproses foto."));
          return;
        }
        ctx.drawImage(img, 0, 0, width, height);
        resolve(canvas.toDataURL("image/jpeg", 0.84));
      };
      img.src = String(reader.result || "");
    };
    reader.readAsDataURL(file);
  });
}

export function UserProfilePhotoUploader({
  name,
  phone,
  initials,
  profilePhotoURL = "",
}: UserProfilePhotoUploaderProps) {
  const router = useRouter();
  const inputRef = useRef<HTMLInputElement | null>(null);
  const [photo, setPhoto] = useState(profilePhotoURL);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function savePhoto(file?: File) {
    if (!file || loading) return;
    setError("");
    setLoading(true);
    try {
      const dataUrl = await resizeProfilePhoto(file);
      if (dataUrl.length > 800000) {
        throw new Error("Foto terlalu besar. Coba pakai foto lain.");
      }
      const res = await fetch("/api/me/profile", {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          nama: name,
          phone,
          profile_photo_url: dataUrl,
        }),
      });
      const body = (await res.json().catch(() => ({}))) as { ok?: boolean; error?: string };
      if (!res.ok || !body.ok) {
        throw new Error(body.error || "Gagal menyimpan foto profil.");
      }
      setPhoto(dataUrl);
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gagal menyimpan foto profil.");
    } finally {
      setLoading(false);
      if (inputRef.current) inputRef.current.value = "";
    }
  }

  return (
    <div className="flex flex-col items-center">
      <button
        type="button"
        onClick={() => inputRef.current?.click()}
        className="relative block cursor-pointer rounded-[28px] outline-none transition active:scale-95 focus-visible:ring-4 focus-visible:ring-white/30"
        aria-label="Pilih foto profil"
      >
        <span className="relative grid h-20 w-20 overflow-hidden rounded-[26px] bg-white text-2xl font-black text-[#087e8b] shadow-[0_16px_34px_rgba(6,78,59,0.18)]">
          {photo ? (
            <span className="absolute inset-0 bg-cover bg-center" style={{ backgroundImage: `url(${photo})` }} />
          ) : (
            <span className="grid h-full w-full place-items-center">{initials}</span>
          )}
          {loading ? <span className="absolute inset-0 grid place-items-center bg-black/30"><span className="h-5 w-5 animate-spin rounded-full border-2 border-white/40 border-t-white" /></span> : null}
        </span>
        <span className="absolute -bottom-1 -right-1 grid h-8 w-8 place-items-center rounded-full border-3 border-[#087e8b] bg-white text-[#087e8b] shadow-[0_8px_18px_rgba(6,78,59,0.18)]">
          <Camera className="h-4 w-4" strokeWidth={2.4} />
        </span>
      </button>
      <input
        ref={inputRef}
        type="file"
        accept="image/*"
        className="sr-only"
        onChange={(event) => void savePhoto(event.target.files?.[0])}
      />
      {error ? <p className="mt-2 max-w-[240px] text-center text-[10px] font-bold leading-4 text-rose-100">{error}</p> : null}
    </div>
  );
}
