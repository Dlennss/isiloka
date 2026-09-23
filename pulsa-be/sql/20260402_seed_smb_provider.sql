BEGIN;

INSERT INTO public.provider (nama, aktif, dibuat_pada, diubah_pada)
SELECT 'smb', false, NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM public.provider WHERE lower(trim(nama)) = 'smb'
);

COMMIT;
