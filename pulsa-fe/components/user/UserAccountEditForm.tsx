"use client";

import { useState, type FormEvent } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowLeft, Camera, CheckCircle2, Mail, Phone, Save, UserRound } from "lucide-react";
import { getInitials } from "@/components/user/helpers";

type UserAccountEditFormProps = {
  nama: string;
  email: string;
  phone: string;
  profilePhotoURL?: string;
};

type ProfileResponse = {
  ok?: boolean;
  error?: string;
};

function cleanPhone(value: string) {
  return value.replace(/[^\d+]/g, "").slice(0, 16);
}

function normalizePhone(value: string) {
  const cleaned = cleanPhone(value.trim());
  if (cleaned.startsWith("+62")) return `0${cleaned.slice(3)}`;
  if (cleaned.startsWith("62")) return `0${cleaned.slice(2)}`;
  return cleaned;
}

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

export function UserAccountEditForm({ nama, email, phone, profilePhotoURL = "" }: UserAccountEditFormProps) {
  const router = useRouter();
  const [nameValue, setNameValue] = useState(nama);
  const [phoneValue, setPhoneValue] = useState(phone);
  const [photoValue, setPhotoValue] = useState(profilePhotoURL);
  const [loading, setLoading] = useState(false);
  const [photoLoading, setPhotoLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const initials = getInitials(nameValue, email);

  async function onPhotoChange(file?: File) {
    if (!file) return;
    setError("");
    setPhotoLoading(true);
    try {
      const dataUrl = await resizeProfilePhoto(file);
      if (dataUrl.length > 800000) {
        setError("Foto masih terlalu besar. Coba pakai foto lain.");
        return;
      }
      setPhotoValue(dataUrl);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gagal memproses foto.");
    } finally {
      setPhotoLoading(false);
    }
  }

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setSuccess("");

    const nextNama = nameValue.trim();
    const nextPhone = normalizePhone(phoneValue);
    if (!nextNama) {
      setError("Nama wajib diisi.");
      return;
    }
    if (!/^08\d{8,12}$/.test(nextPhone)) {
      setError("Nomor handphone gunakan format 08 dan 10-14 digit.");
      return;
    }

    setLoading(true);
    try {
      const res = await fetch("/api/me/profile", {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ nama: nextNama, phone: nextPhone, profile_photo_url: photoValue }),
      });
      const body = (await res.json().catch(() => ({}))) as ProfileResponse;
      if (!res.ok || !body.ok) {
        setError(body.error || "Gagal menyimpan profil.");
        return;
      }
      setPhoneValue(nextPhone);
      setSuccess("Profil berhasil disimpan.");
      router.refresh();
      setTimeout(() => router.replace("/user/account"), 650);
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="min-h-screen bg-[#f3f7f5] px-4 pb-24 pt-4">
      <div className="mx-auto w-full max-w-md">
        <header className="mb-4 flex items-center gap-3">
          <Link href="/user/account" className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-white text-[#087e8b] shadow-[0_10px_24px_rgba(6,78,59,0.08)]">
            <ArrowLeft className="h-5 w-5" />
          </Link>
          <div className="min-w-0">
            <h1 className="text-xl font-black text-slate-950">Edit Akun</h1>
            <p className="mt-0.5 text-xs font-semibold text-slate-500">Perbarui data yang tampil di akun.</p>
          </div>
        </header>

        <section className="overflow-hidden rounded-[28px] border border-emerald-950/5 bg-white shadow-[0_18px_42px_rgba(6,78,59,0.10)]">
          <div className="bg-[linear-gradient(135deg,#052e26_0%,#047857_60%,#84cc16_150%)] px-5 py-5 text-white">
            <div className="flex items-center gap-3">
              <label className="relative grid h-16 w-16 shrink-0 cursor-pointer place-items-center overflow-hidden rounded-2xl bg-white/15 ring-1 ring-white/20">
                {photoValue ? (
                  <span
                    className="absolute inset-0 bg-cover bg-center"
                    style={{ backgroundImage: `url(${photoValue})` }}
                    aria-hidden="true"
                  />
                ) : (
                  <span className="text-xl font-black">{initials}</span>
                )}
                <span className="absolute inset-x-0 bottom-0 grid h-6 place-items-center bg-black/30">
                  {photoLoading ? <span className="h-3 w-3 animate-spin rounded-full border-2 border-white/30 border-t-white" /> : <Camera className="h-3.5 w-3.5" />}
                </span>
                <input
                  type="file"
                  accept="image/*"
                  className="hidden"
                  onChange={(event) => void onPhotoChange(event.target.files?.[0])}
                />
              </label>
              <div className="min-w-0">
                <p className="text-sm font-black">Informasi Pribadi</p>
                <p className="mt-1 truncate text-xs font-medium text-white/75">{email}</p>
              </div>
            </div>
          </div>

          <form onSubmit={onSubmit} className="space-y-4 px-5 py-5">
            {error ? (
              <div className="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-xs font-bold text-rose-700">{error}</div>
            ) : null}

            {success ? (
              <div className="flex items-center gap-2 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-xs font-bold text-emerald-700">
                <CheckCircle2 className="h-4 w-4" />
                {success}
              </div>
            ) : null}

            <label className="block">
              <span className="mb-2 block text-xs font-black text-slate-700">Nama Lengkap</span>
              <span className="relative block">
                <UserRound className="pointer-events-none absolute left-4 top-1/2 h-5 w-5 -translate-y-1/2 text-emerald-700/55" />
                <input
                  value={nameValue}
                  onChange={(event) => setNameValue(event.target.value)}
                  className="h-14 w-full rounded-2xl border border-slate-200 bg-white px-12 text-sm font-bold text-slate-950 outline-none shadow-[0_8px_20px_rgba(15,23,42,0.04)] transition placeholder:text-slate-400 focus:border-[#10b981] focus:ring-4 focus:ring-emerald-100"
                  placeholder="Nama lengkap"
                  autoComplete="name"
                />
              </span>
            </label>

            <label className="block">
              <span className="mb-2 block text-xs font-black text-slate-700">Nomor Handphone</span>
              <span className="relative block">
                <Phone className="pointer-events-none absolute left-4 top-1/2 h-5 w-5 -translate-y-1/2 text-emerald-700/55" />
                <input
                  value={phoneValue}
                  onChange={(event) => setPhoneValue(cleanPhone(event.target.value))}
                  className="h-14 w-full rounded-2xl border border-slate-200 bg-white px-12 text-sm font-bold text-slate-950 outline-none shadow-[0_8px_20px_rgba(15,23,42,0.04)] transition placeholder:text-slate-400 focus:border-[#10b981] focus:ring-4 focus:ring-emerald-100"
                  placeholder="08xxxxxxxxxx"
                  autoComplete="tel"
                  inputMode="tel"
                  type="tel"
                />
              </span>
            </label>

            <label className="block">
              <span className="mb-2 block text-xs font-black text-slate-700">Email</span>
              <span className="relative block">
                <Mail className="pointer-events-none absolute left-4 top-1/2 h-5 w-5 -translate-y-1/2 text-slate-400" />
                <input
                  value={email}
                  className="h-14 w-full rounded-2xl border border-slate-200 bg-slate-50 px-12 text-sm font-bold text-slate-500 outline-none"
                  disabled
                />
              </span>
              <span className="mt-2 block text-[10px] font-semibold text-slate-400">Email belum bisa diubah dari halaman ini.</span>
            </label>

            <button
              type="submit"
              disabled={loading}
              className="flex h-14 w-full items-center justify-center gap-2 rounded-2xl bg-linear-to-r from-[#009944] via-[#16b934] to-[#57d735] text-sm font-black text-white shadow-[0_14px_30px_rgba(22,185,52,0.32)] transition active:scale-[0.98] disabled:opacity-60"
            >
              <Save className="h-5 w-5" />
              {loading ? "Menyimpan..." : "Simpan Perubahan"}
            </button>
          </form>
        </section>
      </div>
    </main>
  );
}
