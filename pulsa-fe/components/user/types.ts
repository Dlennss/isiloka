import type { LucideIcon } from "lucide-react";

export type UserSession = {
  name?: string | null;
  email?: string | null;
  image?: string | null;
  role?: string;
};

export type UserProfile = {
  id: number;
  email: string;
  nama: string;
  phone?: string;
  store_name?: string;
  profile_photo_url?: string;
  role?: string;
  aktif: boolean;
  saldo: number;
  dibuat_pada?: string;
};

export type RetailCommissionSummary = {
  total_earned: number;
  total_pending_withdraw: number;
  total_approved_withdraw: number;
  total_rejected_withdraw?: number;
  available_saldo: number;
};

export type UserMenuItem = {
  title: string;
  href: string;
  icon: LucideIcon;
};

export type UserCategoryItem = {
  id: number;
  nama: string;
  aktif: boolean;
};

export type UserBrandItem = {
  id: number;
  nama: string;
  aktif: boolean;
};

export type UserProductItem = {
  id: number;
  sku: string;
  nama: string;
  kategori_id: number;
  kategori_nama: string;
  group_name?: string | null;
  brand_id: number;
  brand_nama: string;
  tipe_harga: string;
  harga_dasar_app: number;
  nominal?: number | null;
  maksimal_nominal?: number | null;
  fee_master: number;
  fee_agent: number;
  fee_user: number;
  fee_guest: number;
  harga_master_final?: number | null;
  harga_agent_final?: number | null;
  harga_user_final?: number | null;
  harga_guest_final?: number | null;
  payment_fee_master?: number;
  payment_fee_agent?: number;
  payment_fee_user?: number;
  payment_fee_guest?: number;
  harga_master_with_payment_fee_final?: number | null;
  harga_agent_with_payment_fee_final?: number | null;
  harga_user_with_payment_fee_final?: number | null;
  harga_guest_with_payment_fee_final?: number | null;
  aktif: boolean;
};

export type UserAppOrder = {
  id: number;
  invoice_id: string;
  member_id?: number | null;
  member_nama?: string | null;
  guest_nama?: string | null;
  guest_email?: string | null;
  guest_phone?: string | null;
  produk_id: number;
  produk_sku_snapshot: string;
  produk_nama_snapshot: string;
  dest: string;
  qty: number;
  nominal: number;
  buyer_type: "user" | "guest";
  harga_dasar: number;
  fee: number;
  harga_final: number;
  status: string;
  sn?: string | null;
  billing_inquiry?: UserAppBillingInquiry | null;
  dibuat_pada?: string | null;
  diubah_pada?: string | null;
};

export type UserAppBillingInquiry = {
  provider: string;
  provider_status: string;
  provider_message: string;
  display_message: string;
  customer_name?: string | null;
  usage_label?: string | null;
  meter_type?: string | null;
  period_label?: string | null;
  meter_range?: string | null;
  bill_amount: number;
  penalty_amount: number;
  admin_amount: number;
  total_amount: number;
  transaction_time?: string | null;
  can_pay: boolean;
};

export type UserAppBillingCheck = {
  id: number;
  ref_id: string;
  member_id?: number | null;
  member_nama?: string | null;
  guest_nama?: string | null;
  guest_email?: string | null;
  guest_phone?: string | null;
  produk_id: number;
  produk_sku_snapshot: string;
  produk_nama_snapshot: string;
  dest: string;
  buyer_type: "user" | "guest";
  buyer_role?: string;
  provider: string;
  harga_provider: number;
  status: string;
  kode_respon?: string | null;
  pesan?: string | null;
  sn?: string | null;
  billing_inquiry?: UserAppBillingInquiry | null;
  dibuat_pada?: string | null;
  diubah_pada?: string | null;
};

export type UserAppOrderPayment = {
  id: number;
  app_order_id: number;
  order_id: string;
  transaction_id?: string | null;
  gross_amount: number;
  payment_type?: string | null;
  transaction_status?: string | null;
  fraud_status?: string | null;
  acquirer?: string | null;
  qr_url?: string | null;
  raw_request?: string | null;
  raw_callback?: string | null;
  paid_at?: string | null;
  expired_at?: string | null;
  settlement_time?: string | null;
  dibuat_pada?: string | null;
  diubah_pada?: string | null;
};
