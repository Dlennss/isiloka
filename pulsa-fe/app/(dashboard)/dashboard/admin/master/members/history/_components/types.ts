export type MutasiRow = {
  id: number;
  member_id: number;
  ref_id: string;
  arah: string;
  jumlah: number;
  alasan: string;
  catatan?: string | null;
  saldo_sebelum?: number | null;
  saldo_sesudah?: number | null;
  diubah_oleh?: number | null;
  diubah_oleh_nama?: string | null;
  dibuat_pada: string;
};

export type TrxRow = {
  id: number;
  member_id: number;
  ref_id: string;
  perintah: string;
  kode_produk: string;
  tujuan: string;
  qty: number;
  status: string;
  keterangan?: string | null;
  biaya_perkiraan: number;
  biaya_aktual: number;
  dibuat_pada: string;
  diperbarui_pada: string;
};

export type ExportKind = "" | "csv" | "excel" | "pdf";
