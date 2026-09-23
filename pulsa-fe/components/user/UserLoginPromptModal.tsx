'use client';

import * as React from 'react';
import Link from 'next/link';
import { X } from 'lucide-react';

type UserLoginPromptModalProps = {
  isOpen: boolean;
  onClose: () => void;
};

export function UserLoginPromptModal({ isOpen, onClose }: UserLoginPromptModalProps) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-slate-950/45 p-4">
      <div className="max-h-[92vh] w-full max-w-sm overflow-y-auto rounded-3xl bg-white p-5 shadow-[0_18px_46px_rgba(15,23,42,0.24)]">
        {/* Header */}
        <div className="flex items-start justify-between gap-3">
          <div>
            <p className="text-xs font-semibold tracking-wide text-slate-500 uppercase">Info Harga</p>
            <h3 className="mt-1 text-lg font-bold text-slate-900">Masuk atau Daftar</h3>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="inline-flex h-8 w-8 items-center justify-center rounded-xl bg-slate-100 text-slate-500"
            aria-label="Tutup informasi login"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        {/* Content - tanpa SVG */}
        <p className="mt-3 text-sm text-slate-600">
          Masuk atau daftar akun untuk dapat harga lebih hemat dan akses transaksi lebih lengkap.
        </p>

        {/* Buttons */}
        <div className="mt-4 grid grid-cols-2 gap-2">
          <Link
            href="/login?callbackUrl=/user"
            className="inline-flex h-10 items-center justify-center rounded-2xl bg-linear-to-r from-[#087e8b] to-[#087e8b] px-4 text-sm font-semibold text-white! visited:text-white! hover:text-white!"
            onClick={onClose}
          >
            Masuk
          </Link>
          <Link
            href="/register"
            className="inline-flex h-10 items-center justify-center rounded-2xl border border-sky-200 bg-white px-4 text-sm font-semibold text-sky-700"
            onClick={onClose}
          >
            Daftar
          </Link>
        </div>
      </div>
    </div>
  );
}
