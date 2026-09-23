WITH invalid_pending AS (
  SELECT
    id,
    requested_amount,
    1 + MOD(ABS(HASHTEXT(COALESCE(NULLIF(TRIM(ref_id), ''), id::text))::bigint), 999) AS corrected_unique_code
  FROM public.deposit_request
  WHERE status = 'pending'
    AND requested_amount > 0
    AND (unique_code < 1 OR unique_code > 999 OR amount <> requested_amount + unique_code)
)
UPDATE public.deposit_request AS request
SET unique_code = invalid_pending.corrected_unique_code,
    amount = invalid_pending.requested_amount + invalid_pending.corrected_unique_code,
    note = 'Menunggu persetujuan admin'
FROM invalid_pending
WHERE request.id = invalid_pending.id;
