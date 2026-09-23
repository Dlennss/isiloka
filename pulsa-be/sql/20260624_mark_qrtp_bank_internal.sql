UPDATE public.bank
SET admin_staff_only = true,
    diubah_pada = now()
WHERE lower(trim(nama)) = 'qrtp'
  AND COALESCE(admin_staff_only, false) = false;
