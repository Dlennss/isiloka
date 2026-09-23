ALTER TABLE public.provider
  ADD COLUMN IF NOT EXISTS keterangan TEXT NOT NULL DEFAULT '';

ALTER TABLE public.member_ip_whitelist
  ADD COLUMN IF NOT EXISTS webhook_url TEXT;

ALTER TABLE public.provider_saldo_snapshot
  ADD COLUMN IF NOT EXISTS saldo_provider BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS transaksi_member_id BIGINT,
  ADD COLUMN IF NOT EXISTS sumber TEXT;

UPDATE public.provider_saldo_snapshot
SET saldo_provider = saldo
WHERE COALESCE(saldo_provider, 0) = 0
  AND saldo IS NOT NULL;

UPDATE public.provider_saldo_snapshot
SET sumber = source
WHERE (sumber IS NULL OR BTRIM(sumber) = '')
  AND source IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS provider_saldo_snapshot_provider_trx_uidx
  ON public.provider_saldo_snapshot (provider, transaksi_provider_id)
  WHERE transaksi_provider_id IS NOT NULL;

ALTER TABLE public.mutasi_dompet_provider
  ADD COLUMN IF NOT EXISTS bank_nama TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS transaksi_member_id BIGINT,
  ADD COLUMN IF NOT EXISTS transaksi_provider_id BIGINT;
