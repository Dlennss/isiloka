"use client";

import Swal, { type SweetAlertOptions } from "sweetalert2";

type ConfirmOptions = {
  title?: string;
  text?: string;
  confirmButtonText?: string;
  cancelButtonText?: string;
};

function baseOptions(): SweetAlertOptions {
  return {
    background: "#0b1220",
    color: "#e5e7eb",
    confirmButtonColor: "#2563eb",
    cancelButtonColor: "#6b7280",
    customClass: {
      popup: "rounded-2xl border border-white/10 shadow-2xl",
      title: "text-base font-semibold",
      htmlContainer: "text-sm text-slate-300",
      confirmButton: "rounded-lg px-4 py-2 font-semibold",
      cancelButton: "rounded-lg px-4 py-2 font-semibold",
    },
  };
}

export async function alertSuccess(text: string, title = "Berhasil") {
  await Swal.fire({
    ...baseOptions(),
    icon: "success",
    title,
    text,
    confirmButtonText: "OK",
  });
}

export async function alertError(text: string, title = "Terjadi Error") {
  await Swal.fire({
    ...baseOptions(),
    icon: "error",
    title,
    text,
    confirmButtonText: "OK",
  });
}

export async function alertWarning(text: string, title = "Perhatian") {
  await Swal.fire({
    ...baseOptions(),
    icon: "warning",
    title,
    text,
    confirmButtonText: "OK",
  });
}

export async function alertConfirm(options: ConfirmOptions = {}): Promise<boolean> {
  const res = await Swal.fire({
    ...baseOptions(),
    icon: "question",
    title: options.title || "Konfirmasi",
    text: options.text || "Apakah Anda yakin?",
    showCancelButton: true,
    reverseButtons: true,
    focusCancel: true,
    confirmButtonText: options.confirmButtonText || "Ya, lanjut",
    cancelButtonText: options.cancelButtonText || "Batal",
  });
  return res.isConfirmed;
}

// Alias (ikut naming yang diminta user)
export const alertSucces = alertSuccess;
