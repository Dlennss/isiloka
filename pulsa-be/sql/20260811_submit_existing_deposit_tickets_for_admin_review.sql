UPDATE public.deposit_request
SET status = 'pending',
    note = CASE
      WHEN NULLIF(TRIM(COALESCE(note, '')), '') IS NULL THEN 'Menunggu persetujuan admin'
      ELSE note
    END
WHERE status = 'ticket'
  AND LOWER(TRIM(COALESCE(metode, ''))) <> 'qris';
