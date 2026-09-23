CREATE INDEX CONCURRENTLY IF NOT EXISTS mutasi_bank_debit_effective_time_idx
ON public.mutasi_bank ((COALESCE(waktu_mutasi_bank, dibuat_pada)) DESC, id DESC)
WHERE upper(arah) = 'DEBIT';

CREATE INDEX CONCURRENTLY IF NOT EXISTS mutasi_bank_internal_credit_pair_idx
ON public.mutasi_bank (ref_id, jumlah, bank_id)
WHERE upper(arah) = 'CREDIT'
  AND COALESCE(alasan, '') = 'BANK_TRANSFER_IN';

CREATE INDEX CONCURRENTLY IF NOT EXISTS mutasi_dompet_provider_bank_transfer_in_pair_idx
ON public.mutasi_dompet_provider (ref_id, jumlah, provider)
WHERE lower(COALESCE(arah, '')) = 'credit'
  AND COALESCE(alasan, '') = 'BANK_TRANSFER_IN';
