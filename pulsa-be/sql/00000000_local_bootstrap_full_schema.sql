-- Local bootstrap schema for a fresh PostgreSQL database.
-- This file is inferred from backend repositories and SQL migrations in this repo.
-- Use it for local development when you do not have the original production dump.
--
-- Example:
--   psql "postgres://postgres:postgres@127.0.0.1:5432/pulsa?sslmode=disable" -f sql/00000000_local_bootstrap_full_schema.sql

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS public.member (
  id BIGSERIAL PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  nama TEXT,
  store_name TEXT NOT NULL DEFAULT '',
  profile_photo_url TEXT NOT NULL DEFAULT '',
  password_hash TEXT,
  pin_hash TEXT,
  role TEXT NOT NULL DEFAULT 'user',
  aktif BOOLEAN NOT NULL DEFAULT true,
  google_sub TEXT,
  apple_sub TEXT,
  charge_receiver BOOLEAN NOT NULL DEFAULT false,
  fee_member_rp BIGINT NOT NULL DEFAULT 0,
  retail_agent_commission_rp BIGINT NOT NULL DEFAULT 0,
  retail_master_commission_rp BIGINT NOT NULL DEFAULT 0,
  h2h_agent_commission_rp BIGINT NOT NULL DEFAULT 0,
  h2h_master_commission_rp BIGINT NOT NULL DEFAULT 0,
  retail_agent_id BIGINT,
  retail_master_id BIGINT,
  h2h_agent_member_id BIGINT,
  h2h_master_member_id BIGINT,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.member_api_key (
  id BIGSERIAL PRIMARY KEY,
  member_id BIGINT NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  api_key TEXT NOT NULL UNIQUE,
  label TEXT,
  aktif BOOLEAN NOT NULL DEFAULT true,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.member_ip_whitelist (
  id BIGSERIAL PRIMARY KEY,
  member_id BIGINT NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  ip TEXT NOT NULL,
  label TEXT,
  aktif BOOLEAN NOT NULL DEFAULT true,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(member_id, ip)
);

CREATE TABLE IF NOT EXISTS public.dompet_member (
  member_id BIGINT PRIMARY KEY REFERENCES public.member(id) ON DELETE CASCADE,
  saldo BIGINT NOT NULL DEFAULT 0,
  diperbarui_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.mutasi_dompet (
  id BIGSERIAL PRIMARY KEY,
  member_id BIGINT NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  arah TEXT NOT NULL,
  jumlah BIGINT NOT NULL DEFAULT 0,
  saldo_sebelum BIGINT NOT NULL DEFAULT 0,
  saldo_sesudah BIGINT NOT NULL DEFAULT 0,
  ref_id TEXT,
  tipe TEXT,
  alasan TEXT,
  catatan TEXT,
  diubah_oleh BIGINT,
  meta JSONB NOT NULL DEFAULT '{}'::jsonb,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.kategori (
  id BIGSERIAL PRIMARY KEY,
  nama TEXT NOT NULL UNIQUE,
  aktif BOOLEAN NOT NULL DEFAULT true,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.brand (
  id BIGSERIAL PRIMARY KEY,
  nama TEXT NOT NULL UNIQUE,
  aktif BOOLEAN NOT NULL DEFAULT true,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.produk (
  id BIGSERIAL PRIMARY KEY,
  sku TEXT NOT NULL UNIQUE,
  nama TEXT NOT NULL,
  group_name TEXT NOT NULL DEFAULT '',
  kategori_id BIGINT REFERENCES public.kategori(id),
  brand_id BIGINT REFERENCES public.brand(id),
  tipe_harga TEXT NOT NULL DEFAULT 'FIXED',
  nominal BIGINT,
  maksimal_nominal BIGINT,
  jam_buka TIME NOT NULL DEFAULT '00:31',
  jam_tutup TIME NOT NULL DEFAULT '23:29',
  aktif BOOLEAN NOT NULL DEFAULT true,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.provider (
  id BIGSERIAL PRIMARY KEY,
  nama TEXT NOT NULL UNIQUE,
  aktif BOOLEAN NOT NULL DEFAULT true,
  auto_disable_failed_then_success BOOLEAN NOT NULL DEFAULT false,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.produk_fee_provider (
  id BIGSERIAL PRIMARY KEY,
  produk_id BIGINT REFERENCES public.produk(id) ON DELETE CASCADE,
  provider TEXT NOT NULL,
  fee_rp BIGINT NOT NULL DEFAULT 0,
  aktif BOOLEAN NOT NULL DEFAULT true,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(produk_id, provider)
);

CREATE TABLE IF NOT EXISTS public.produk_provider_map (
  id BIGSERIAL PRIMARY KEY,
  produk_id BIGINT NOT NULL REFERENCES public.produk(id) ON DELETE CASCADE,
  provider TEXT NOT NULL,
  kode_provider TEXT NOT NULL,
  mode TEXT NOT NULL DEFAULT 'normal',
  aktif BOOLEAN NOT NULL DEFAULT true,
  prioritas INT NOT NULL DEFAULT 100,
  minimal_nominal BIGINT,
  maksimal_nominal BIGINT,
  fee_rp BIGINT NOT NULL DEFAULT 0,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(produk_id, provider, kode_provider)
);

CREATE TABLE IF NOT EXISTS public.kategori_fee_app (
  id BIGSERIAL PRIMARY KEY,
  kategori_id BIGINT REFERENCES public.kategori(id) ON DELETE CASCADE,
  fee_master BIGINT NOT NULL DEFAULT 0,
  fee_agent BIGINT NOT NULL DEFAULT 0,
  fee_user BIGINT NOT NULL DEFAULT 0,
  fee_non_user BIGINT NOT NULL DEFAULT 0,
  aktif BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(kategori_id)
);

CREATE TABLE IF NOT EXISTS public.produk_app_pricing (
  id BIGSERIAL PRIMARY KEY,
  produk_id BIGINT NOT NULL REFERENCES public.produk(id) ON DELETE CASCADE,
  provider TEXT NOT NULL DEFAULT 'yuscom',
  harga BIGINT NOT NULL DEFAULT 0,
  harga_dasar BIGINT NOT NULL DEFAULT 0,
  fee_user BIGINT NOT NULL DEFAULT 0,
  fee_agent BIGINT NOT NULL DEFAULT 0,
  fee_master BIGINT NOT NULL DEFAULT 0,
  harga_user BIGINT NOT NULL DEFAULT 0,
  harga_agent BIGINT NOT NULL DEFAULT 0,
  harga_master BIGINT NOT NULL DEFAULT 0,
  display_brand TEXT,
  yuscom_group TEXT NOT NULL DEFAULT '',
  yuscom_category TEXT NOT NULL DEFAULT '',
  yuscom_subcategory TEXT NOT NULL DEFAULT '',
  yuscom_sku TEXT NOT NULL DEFAULT '',
  yuscom_name TEXT NOT NULL DEFAULT '',
  yuscom_status TEXT NOT NULL DEFAULT '',
  yuscom_display_brand TEXT NOT NULL DEFAULT '',
  aktif BOOLEAN NOT NULL DEFAULT true,
  fetched_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(produk_id)
);

CREATE TABLE IF NOT EXISTS public.member_fee_kategori (
  id BIGSERIAL PRIMARY KEY,
  member_id BIGINT NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  kategori_id BIGINT NOT NULL REFERENCES public.kategori(id) ON DELETE CASCADE,
  fee_rp BIGINT NOT NULL DEFAULT 0,
  aktif BOOLEAN NOT NULL DEFAULT true,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(member_id, kategori_id)
);

CREATE TABLE IF NOT EXISTS public.member_fee_produk (
  id BIGSERIAL PRIMARY KEY,
  member_id BIGINT NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  produk_id BIGINT NOT NULL REFERENCES public.produk(id) ON DELETE CASCADE,
  fee_rp BIGINT NOT NULL DEFAULT 0,
  aktif BOOLEAN NOT NULL DEFAULT true,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(member_id, produk_id)
);

CREATE TABLE IF NOT EXISTS public.member_h2h_fee (
  id BIGSERIAL PRIMARY KEY,
  member_id BIGINT NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  fee_code TEXT NOT NULL,
  fee_rp BIGINT NOT NULL DEFAULT 0,
  aktif BOOLEAN NOT NULL DEFAULT true,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(member_id, fee_code)
);

CREATE TABLE IF NOT EXISTS public.transaksi_member (
  id BIGSERIAL PRIMARY KEY,
  member_id BIGINT NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  ref_id TEXT NOT NULL,
  perintah TEXT NOT NULL,
  kode_produk TEXT NOT NULL,
  tujuan TEXT NOT NULL,
  qty BIGINT NOT NULL DEFAULT 1,
  qty_provider BIGINT,
  charge_receiver_applied BOOLEAN NOT NULL DEFAULT false,
  status TEXT NOT NULL DEFAULT 'pending',
  keterangan TEXT,
  biaya_perkiraan BIGINT NOT NULL DEFAULT 0,
  biaya_aktual BIGINT NOT NULL DEFAULT 0,
  fee_member_rp BIGINT NOT NULL DEFAULT 0,
  harga_javapay BIGINT NOT NULL DEFAULT 0,
  harga_member BIGINT NOT NULL DEFAULT 0,
  callback_url TEXT,
  callback_status INT,
  callback_body TEXT,
  callback_sent_at TIMESTAMPTZ,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diperbarui_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(member_id, ref_id)
);

CREATE TABLE IF NOT EXISTS public.transaksi_provider (
  id BIGSERIAL PRIMARY KEY,
  provider TEXT NOT NULL DEFAULT 'javapay',
  transaksi_member_id BIGINT REFERENCES public.transaksi_member(id) ON DELETE SET NULL,
  ref_id TEXT NOT NULL DEFAULT ('PRV-' || replace(gen_random_uuid()::text, '-', '')),
  perintah TEXT,
  produk_sku_snapshot TEXT,
  produk_provider_map_id BIGINT,
  kode_produk TEXT,
  tujuan TEXT,
  qty BIGINT NOT NULL DEFAULT 1,
  percobaan INT NOT NULL DEFAULT 1,
  status TEXT NOT NULL DEFAULT 'pending',
  route_candidate JSONB,
  trx_id_javapay TEXT,
  kode_respon TEXT,
  pesan TEXT,
  no_referensi TEXT,
  sn TEXT,
  harga BIGINT,
  saldo_terakhir BIGINT,
  http_status INT,
  request_mentah JSONB,
  respon_mentah JSONB,
  provider_status TEXT,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diperbarui_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.transaksi_member_status_log (
  id BIGSERIAL PRIMARY KEY,
  transaksi_member_id BIGINT NOT NULL REFERENCES public.transaksi_member(id) ON DELETE CASCADE,
  status_lama TEXT,
  status_baru TEXT NOT NULL,
  keterangan TEXT,
  source TEXT,
  actor_id BIGINT,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.transaksi_anomasi_provider (
  id BIGSERIAL PRIMARY KEY,
  transaksi_provider_id BIGINT REFERENCES public.transaksi_provider(id) ON DELETE CASCADE,
  tipe TEXT,
  status TEXT NOT NULL DEFAULT 'open',
  catatan TEXT,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.marking_trx_resolve (
  id BIGSERIAL PRIMARY KEY,
  transaksi_provider_id BIGINT REFERENCES public.transaksi_provider(id) ON DELETE CASCADE,
  user_id BIGINT REFERENCES public.member(id),
  status TEXT NOT NULL DEFAULT 'resolved',
  note TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(transaksi_provider_id)
);

CREATE TABLE IF NOT EXISTS public.provider_message_pattern (
  id BIGSERIAL PRIMARY KEY,
  provider TEXT NOT NULL,
  pattern TEXT NOT NULL,
  status TEXT NOT NULL,
  aktif BOOLEAN NOT NULL DEFAULT true,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.dompet_provider (
  provider TEXT PRIMARY KEY,
  saldo BIGINT NOT NULL DEFAULT 0,
  diperbarui_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.mutasi_dompet_provider (
  id BIGSERIAL PRIMARY KEY,
  provider TEXT NOT NULL,
  arah TEXT NOT NULL,
  jumlah BIGINT NOT NULL DEFAULT 0,
  saldo_sebelum BIGINT NOT NULL DEFAULT 0,
  saldo_sesudah BIGINT NOT NULL DEFAULT 0,
  ref_id TEXT,
  tipe TEXT,
  alasan TEXT,
  catatan TEXT,
  bank_id BIGINT,
  diubah_oleh BIGINT,
  meta JSONB NOT NULL DEFAULT '{}'::jsonb,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.provider_saldo_snapshot (
  id BIGSERIAL PRIMARY KEY,
  provider TEXT NOT NULL,
  transaksi_provider_id BIGINT,
  ref_id TEXT,
  saldo BIGINT NOT NULL DEFAULT 0,
  source TEXT,
  raw JSONB,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.bank (
  id BIGSERIAL PRIMARY KEY,
  nama TEXT NOT NULL,
  nomor_rekening TEXT NOT NULL DEFAULT '',
  atas_nama TEXT NOT NULL DEFAULT '',
  saldo BIGINT NOT NULL DEFAULT 0,
  aktif BOOLEAN NOT NULL DEFAULT true,
  admin_staff_only BOOLEAN NOT NULL DEFAULT false,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.mutasi_bank (
  id BIGSERIAL PRIMARY KEY,
  bank_id BIGINT REFERENCES public.bank(id),
  arah TEXT NOT NULL,
  jumlah BIGINT NOT NULL DEFAULT 0,
  saldo_sebelum BIGINT NOT NULL DEFAULT 0,
  saldo_sesudah BIGINT NOT NULL DEFAULT 0,
  ref_id TEXT,
  alasan TEXT,
  catatan TEXT,
  member_id BIGINT,
  provider TEXT,
  diubah_oleh BIGINT,
  identity_key TEXT,
  external_id TEXT,
  meta JSONB NOT NULL DEFAULT '{}'::jsonb,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.deposit_request (
  id BIGSERIAL PRIMARY KEY,
  member_id BIGINT NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  bank_id BIGINT REFERENCES public.bank(id),
  bank_nama TEXT NOT NULL DEFAULT '',
  bank_nomor_rekening TEXT NOT NULL DEFAULT '',
  bank_atas_nama TEXT NOT NULL DEFAULT '',
  amount BIGINT NOT NULL DEFAULT 0,
  requested_amount BIGINT NOT NULL DEFAULT 0,
  unique_code BIGINT NOT NULL DEFAULT 0,
  approved_amount BIGINT NOT NULL DEFAULT 0,
  metode TEXT NOT NULL DEFAULT 'transfer',
  bukti_url TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'pending',
  note TEXT NOT NULL DEFAULT '',
  ref_id TEXT,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diproses_pada TIMESTAMPTZ,
  diproses_oleh BIGINT
);

CREATE TABLE IF NOT EXISTS public.app_ads (
  id BIGSERIAL PRIMARY KEY,
  judul TEXT NOT NULL DEFAULT '',
  keterangan TEXT NOT NULL DEFAULT '',
  image_url TEXT NOT NULL DEFAULT '',
  link_url TEXT,
  urutan INT NOT NULL DEFAULT 0,
  aktif BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.app_billing_check (
  id BIGSERIAL PRIMARY KEY,
  ref_id TEXT NOT NULL UNIQUE,
  member_id BIGINT REFERENCES public.member(id),
  guest_nama TEXT,
  guest_email TEXT,
  guest_phone TEXT,
  produk_id BIGINT NOT NULL REFERENCES public.produk(id),
  produk_sku_snapshot TEXT NOT NULL,
  produk_nama_snapshot TEXT NOT NULL,
  dest TEXT NOT NULL,
  buyer_type TEXT NOT NULL DEFAULT 'guest',
  buyer_role TEXT NOT NULL DEFAULT 'guest',
  provider TEXT NOT NULL DEFAULT '',
  harga_provider BIGINT NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'pending',
  kode_respon TEXT,
  pesan TEXT,
  sn TEXT,
  billing_inquiry JSONB,
  raw_request JSONB,
  raw_response JSONB,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.app_order (
  id BIGSERIAL PRIMARY KEY,
  invoice_id TEXT NOT NULL UNIQUE,
  member_id BIGINT REFERENCES public.member(id),
  guest_nama TEXT,
  guest_email TEXT,
  guest_phone TEXT,
  produk_id BIGINT NOT NULL REFERENCES public.produk(id),
  produk_sku_snapshot TEXT NOT NULL,
  produk_nama_snapshot TEXT NOT NULL,
  dest TEXT NOT NULL,
  qty BIGINT NOT NULL DEFAULT 1,
  nominal BIGINT NOT NULL DEFAULT 0,
  buyer_type TEXT NOT NULL DEFAULT 'guest',
  buyer_role TEXT NOT NULL DEFAULT 'guest',
  harga_dasar BIGINT NOT NULL DEFAULT 0,
  fee BIGINT NOT NULL DEFAULT 0,
  harga_final BIGINT NOT NULL DEFAULT 0,
  fee_user_snapshot BIGINT NOT NULL DEFAULT 0,
  fee_agent_snapshot BIGINT NOT NULL DEFAULT 0,
  fee_master_snapshot BIGINT NOT NULL DEFAULT 0,
  retail_agent_id_snapshot BIGINT,
  retail_master_id_snapshot BIGINT,
  status TEXT NOT NULL DEFAULT 'pending_payment',
  catatan TEXT,
  alasan_gagal TEXT,
  sn TEXT,
  billing_inquiry JSONB,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.app_order_payment (
  id BIGSERIAL PRIMARY KEY,
  app_order_id BIGINT NOT NULL REFERENCES public.app_order(id) ON DELETE CASCADE,
  order_id TEXT NOT NULL UNIQUE,
  transaction_id TEXT,
  gross_amount BIGINT NOT NULL DEFAULT 0,
  payment_type TEXT,
  transaction_status TEXT,
  fraud_status TEXT,
  acquirer TEXT,
  qr_url TEXT,
  raw_request JSONB,
  raw_callback JSONB,
  paid_at TIMESTAMPTZ,
  expired_at TIMESTAMPTZ,
  settlement_time TIMESTAMPTZ,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.app_order_provider_trx (
  id BIGSERIAL PRIMARY KEY,
  app_order_id BIGINT NOT NULL REFERENCES public.app_order(id) ON DELETE CASCADE,
  provider TEXT NOT NULL,
  ref_id TEXT NOT NULL,
  kode_provider TEXT NOT NULL DEFAULT '',
  harga_provider BIGINT NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'pending',
  kode_respon TEXT,
  pesan TEXT,
  sn TEXT,
  raw_request JSONB,
  raw_callback JSONB,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.app_order_guest_refund (
  id BIGSERIAL PRIMARY KEY,
  app_order_id BIGINT NOT NULL REFERENCES public.app_order(id) ON DELETE CASCADE,
  invoice_id TEXT NOT NULL,
  guest_nama TEXT,
  guest_email TEXT NOT NULL DEFAULT '',
  guest_phone TEXT NOT NULL DEFAULT '',
  amount_refund BIGINT NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'pending',
  reason TEXT,
  notes TEXT,
  claimed_member_id BIGINT,
  claimed_at TIMESTAMPTZ,
  processed_at TIMESTAMPTZ,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(app_order_id)
);

CREATE TABLE IF NOT EXISTS public.retail_commission_ledger (
  id BIGSERIAL PRIMARY KEY,
  member_id BIGINT NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  source_member_id BIGINT,
  source_app_order_id BIGINT REFERENCES public.app_order(id) ON DELETE SET NULL,
  invoice_id TEXT,
  level_name TEXT NOT NULL DEFAULT '',
  amount BIGINT NOT NULL DEFAULT 0,
  note TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS retail_commission_ledger_member_order_uidx
  ON public.retail_commission_ledger (member_id, source_app_order_id);

CREATE INDEX IF NOT EXISTS retail_commission_ledger_member_created_idx
  ON public.retail_commission_ledger (member_id, created_at DESC);

CREATE TABLE IF NOT EXISTS public.retail_withdraw_request (
  id BIGSERIAL PRIMARY KEY,
  member_id BIGINT NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  amount BIGINT NOT NULL DEFAULT 0,
  bank_name TEXT NOT NULL DEFAULT '',
  account_name TEXT NOT NULL DEFAULT '',
  account_number TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'pending',
  note TEXT,
  reject_reason TEXT NOT NULL DEFAULT '',
  ref_id TEXT NOT NULL DEFAULT '',
  processed_by BIGINT,
  processed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS retail_withdraw_request_ref_id_uidx
  ON public.retail_withdraw_request (ref_id)
  WHERE TRIM(ref_id) <> '';

CREATE INDEX IF NOT EXISTS retail_withdraw_request_member_created_idx
  ON public.retail_withdraw_request (member_id, created_at DESC);

CREATE INDEX IF NOT EXISTS retail_withdraw_request_status_created_idx
  ON public.retail_withdraw_request (status, created_at DESC);

CREATE TABLE IF NOT EXISTS public.h2h_commission_ledger (
  id BIGSERIAL PRIMARY KEY,
  member_id BIGINT NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  source_member_id BIGINT,
  source_trx_member_id BIGINT REFERENCES public.transaksi_member(id) ON DELETE SET NULL,
  ref_id TEXT NOT NULL,
  level TEXT NOT NULL,
  amount BIGINT NOT NULL DEFAULT 0,
  kategori TEXT NOT NULL DEFAULT '',
  note TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.h2h_withdraw_request (
  id BIGSERIAL PRIMARY KEY,
  member_id BIGINT NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  amount BIGINT NOT NULL DEFAULT 0,
  bank_name TEXT NOT NULL DEFAULT '',
  bank_account_no TEXT NOT NULL DEFAULT '',
  bank_account_name TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'pending',
  note TEXT,
  processed_by BIGINT,
  processed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.internal_finance_entry (
  id BIGSERIAL PRIMARY KEY,
  bank_id BIGINT NOT NULL REFERENCES public.bank(id),
  arah TEXT NOT NULL,
  amount BIGINT NOT NULL DEFAULT 0,
  category TEXT NOT NULL DEFAULT '',
  note TEXT,
  ref_id TEXT,
  created_by BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.accounting_opening_balance (
  id BIGSERIAL PRIMARY KEY,
  scope TEXT NOT NULL,
  scope_id BIGINT NOT NULL DEFAULT 0,
  balance BIGINT NOT NULL DEFAULT 0,
  as_of_date DATE NOT NULL DEFAULT CURRENT_DATE,
  note TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(scope, scope_id, as_of_date)
);

CREATE TABLE IF NOT EXISTS public.yuscom_produk_snapshot (
  id BIGSERIAL PRIMARY KEY,
  kode_produk TEXT NOT NULL,
  nama_produk TEXT NOT NULL DEFAULT '',
  kategori TEXT NOT NULL DEFAULT '',
  brand TEXT NOT NULL DEFAULT '',
  display_brand TEXT,
  harga BIGINT NOT NULL DEFAULT 0,
  aktif BOOLEAN NOT NULL DEFAULT true,
  raw JSONB,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.provider_rekening (
  id BIGSERIAL PRIMARY KEY,
  provider TEXT NOT NULL,
  bank_name TEXT NOT NULL DEFAULT '',
  account_no TEXT NOT NULL DEFAULT '',
  account_name TEXT NOT NULL DEFAULT '',
  aktif BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.provider_merchant_id (
  id BIGSERIAL PRIMARY KEY,
  provider TEXT NOT NULL,
  merchant_id TEXT NOT NULL,
  label TEXT,
  aktif BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(provider, merchant_id)
);

CREATE TABLE IF NOT EXISTS public.provider_failed_then_success_guard (
  id BIGSERIAL PRIMARY KEY,
  provider TEXT NOT NULL,
  ref_id TEXT NOT NULL,
  transaksi_provider_id BIGINT,
  status TEXT NOT NULL DEFAULT 'open',
  raw JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(provider, ref_id)
);

CREATE TABLE IF NOT EXISTS public.smpay_ref_sources (
  id BIGSERIAL PRIMARY KEY,
  member_id BIGINT NOT NULL,
  ref_id TEXT NOT NULL,
  product TEXT NOT NULL DEFAULT '',
  source TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(member_id, ref_id)
);

CREATE TABLE IF NOT EXISTS public.smpay_transaction_sources (
  id BIGSERIAL PRIMARY KEY,
  transaksi_member_id BIGINT NOT NULL REFERENCES public.transaksi_member(id) ON DELETE CASCADE,
  smpay_ref_source_id BIGINT REFERENCES public.smpay_ref_sources(id) ON DELETE SET NULL,
  source TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(transaksi_member_id)
);

CREATE TABLE IF NOT EXISTS public.qrtp_provider_transfers (
  id BIGSERIAL PRIMARY KEY,
  provider TEXT NOT NULL,
  bank_id BIGINT REFERENCES public.bank(id),
  amount BIGINT NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'pending',
  ref_id TEXT,
  note TEXT,
  created_by BIGINT,
  processed_by BIGINT,
  processed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.loketbayar_provider_transfers (
  id BIGSERIAL PRIMARY KEY,
  provider TEXT NOT NULL DEFAULT 'loketbayar',
  bank_id BIGINT REFERENCES public.bank(id),
  product_code TEXT NOT NULL DEFAULT '',
  amount BIGINT NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'pending',
  ref_id TEXT,
  inquiry_raw JSONB,
  process_raw JSONB,
  note TEXT,
  created_by BIGINT,
  processed_by BIGINT,
  processed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.provider_success_suspect_cache (
  transaksi_provider_id BIGINT PRIMARY KEY,
  provider TEXT NOT NULL,
  ref_id TEXT NOT NULL,
  status TEXT NOT NULL,
  harga BIGINT NOT NULL DEFAULT 0,
  saldo_terakhir BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  refreshed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.admin_daily_business_cache (
  scope TEXT NOT NULL,
  day DATE NOT NULL,
  month_key TEXT NOT NULL,
  transaction_count BIGINT NOT NULL DEFAULT 0,
  transaction_amount BIGINT NOT NULL DEFAULT 0,
  provider_payment_amount BIGINT NOT NULL DEFAULT 0,
  margin_amount BIGINT NOT NULL DEFAULT 0,
  commission_amount BIGINT NOT NULL DEFAULT 0,
  transaction_expense_amount BIGINT NOT NULL DEFAULT 0,
  member_deposit_amount BIGINT NOT NULL DEFAULT 0,
  provider_deposit_amount BIGINT NOT NULL DEFAULT 0,
  deposit_gap_amount BIGINT NOT NULL DEFAULT 0,
  profit_amount BIGINT NOT NULL DEFAULT 0,
  refreshed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY(scope, day)
);

CREATE TABLE IF NOT EXISTS public.admin_provider_analytics_daily_cache (
  provider TEXT NOT NULL,
  day DATE NOT NULL,
  trx_count BIGINT NOT NULL DEFAULT 0,
  success_count BIGINT NOT NULL DEFAULT 0,
  failed_count BIGINT NOT NULL DEFAULT 0,
  pending_count BIGINT NOT NULL DEFAULT 0,
  amount BIGINT NOT NULL DEFAULT 0,
  provider_cost BIGINT NOT NULL DEFAULT 0,
  refreshed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY(provider, day)
);

CREATE TABLE IF NOT EXISTS public.audit_provider_wallet_missing_debit_ignore (
  id BIGSERIAL PRIMARY KEY,
  transaksi_provider_id BIGINT NOT NULL UNIQUE,
  reason TEXT,
  created_by BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.maintenance_statuspay_recovery_refund_20260606 (
  id BIGSERIAL PRIMARY KEY,
  transaksi_member_id BIGINT,
  ref_id TEXT,
  amount BIGINT NOT NULL DEFAULT 0,
  note TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_member_role ON public.member(role);
CREATE UNIQUE INDEX IF NOT EXISTS member_google_sub_uidx
  ON public.member(google_sub)
  WHERE google_sub IS NOT NULL AND TRIM(google_sub) <> '';
CREATE UNIQUE INDEX IF NOT EXISTS member_apple_sub_uidx
  ON public.member(apple_sub)
  WHERE apple_sub IS NOT NULL AND TRIM(apple_sub) <> '';
CREATE INDEX IF NOT EXISTS idx_member_api_key_member ON public.member_api_key(member_id);
CREATE INDEX IF NOT EXISTS idx_produk_kategori_brand ON public.produk(kategori_id, brand_id);
CREATE INDEX IF NOT EXISTS idx_produk_provider_map_lookup ON public.produk_provider_map(produk_id, provider, aktif);
CREATE INDEX IF NOT EXISTS idx_transaksi_member_ref ON public.transaksi_member(member_id, ref_id);
CREATE INDEX IF NOT EXISTS idx_transaksi_member_status ON public.transaksi_member(status, dibuat_pada);
CREATE INDEX IF NOT EXISTS idx_transaksi_provider_ref_provider ON public.transaksi_provider(ref_id, provider);
CREATE INDEX IF NOT EXISTS idx_transaksi_provider_status ON public.transaksi_provider(status, dibuat_pada);
CREATE INDEX IF NOT EXISTS idx_mutasi_dompet_member ON public.mutasi_dompet(member_id, dibuat_pada DESC);
CREATE INDEX IF NOT EXISTS idx_mutasi_dompet_provider_provider ON public.mutasi_dompet_provider(provider, dibuat_pada DESC);
CREATE INDEX IF NOT EXISTS idx_mutasi_bank_bank ON public.mutasi_bank(bank_id, dibuat_pada DESC);
CREATE INDEX IF NOT EXISTS idx_deposit_request_status ON public.deposit_request(status, dibuat_pada);
CREATE INDEX IF NOT EXISTS idx_deposit_request_ref_id ON public.deposit_request(ref_id);
CREATE INDEX IF NOT EXISTS idx_app_order_invoice ON public.app_order(invoice_id);
CREATE INDEX IF NOT EXISTS idx_app_order_member ON public.app_order(member_id, dibuat_pada DESC);
CREATE INDEX IF NOT EXISTS idx_app_order_payment_order_id ON public.app_order_payment(order_id);
CREATE INDEX IF NOT EXISTS idx_app_order_provider_ref ON public.app_order_provider_trx(provider, ref_id);
CREATE INDEX IF NOT EXISTS idx_bank_active ON public.bank(aktif);
CREATE INDEX IF NOT EXISTS idx_provider_active ON public.provider(aktif);

INSERT INTO public.kategori (nama, aktif)
VALUES ('Pulsa', true), ('Paket Data', true), ('E-Wallet', true), ('PLN', true), ('Game', true)
ON CONFLICT (nama) DO NOTHING;

INSERT INTO public.brand (nama, aktif)
VALUES ('Telkomsel', true), ('Indosat', true), ('XL', true), ('Smartfren', true), ('Tri', true)
ON CONFLICT (nama) DO NOTHING;

INSERT INTO public.provider (nama, aktif)
VALUES
  ('yuscom', true),
  ('javapay', true),
  ('talenta', true),
  ('multikom', true),
  ('loketbayar', true),
  ('sagaramobile', true),
  ('minions', true),
  ('trionik', true),
  ('ajs', true),
  ('gemilang', true),
  ('smb', true),
  ('chytron', true),
  ('rajabiller', true)
ON CONFLICT (nama) DO NOTHING;
