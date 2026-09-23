ALTER TABLE public.deposit_request
ADD COLUMN IF NOT EXISTS ref_id text;

CREATE INDEX IF NOT EXISTS idx_deposit_request_ref_id
ON public.deposit_request (ref_id);

UPDATE public.deposit_request dr
SET ref_id = (
  SELECT md.ref_id
  FROM public.mutasi_dompet md
  WHERE md.member_id = dr.member_id
    AND LOWER(TRIM(COALESCE(md.arah, ''))) = 'credit'
    AND LOWER(TRIM(COALESCE(md.alasan, ''))) = 'deposit approve'
    AND md.jumlah = dr.amount
    AND NULLIF(TRIM(COALESCE(md.ref_id, '')), '') IS NOT NULL
    AND md.dibuat_pada >= COALESCE(dr.diproses_pada, dr.dibuat_pada) - interval '1 day'
    AND md.dibuat_pada < COALESCE(dr.diproses_pada, dr.dibuat_pada) + interval '1 day'
  ORDER BY ABS(EXTRACT(EPOCH FROM (md.dibuat_pada - COALESCE(dr.diproses_pada, dr.dibuat_pada)))) ASC, md.id DESC
  LIMIT 1
)
WHERE (dr.ref_id IS NULL OR TRIM(dr.ref_id) = '')
  AND dr.status = 'approved';
