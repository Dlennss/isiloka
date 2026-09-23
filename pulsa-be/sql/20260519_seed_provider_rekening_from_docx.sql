INSERT INTO public.provider_rekening
  (provider, nama, bank, nomor_rekening, nomor_rekening_digits, catatan, aktif)
VALUES
  ('smb', 'PT SIMBA PRADANA SOLUSINDO', 'BRI', '017201002852561', '017201002852561', '', true),
  ('smb', 'PT SIMBA PRADANA SOLUSINDO', 'BNI', '1737856875', '1737856875', '', true),
  ('smb', 'PT SIMBA PRADANA SOLUSINDO', 'MANDIRI', '1710013580629', '1710013580629', '', true),
  ('smb', 'PT AGEN RETAIL DIGITAL', 'BCA', '3245078999', '3245078999', '', true),

  ('yuscom', 'PT YUS TELECOM INDONESIA', 'BCA', '3900945514', '3900945514', '', true),
  ('yuscom', 'PT YUS TELECOM INDONESIA', 'MANDIRI', '1740004662805', '1740004662805', '', true),
  ('yuscom', 'PT YUS TELECOM INDONESIA', 'BRI', '064201003041300', '064201003041300', '', true),
  ('yuscom', 'PT YUS TELECOM INDONESIA', 'BNI', '2199920221', '2199920221', '', true),

  ('trionik', 'PT MULYO TRONIK INDONESIA', 'BCA', '0469000089', '0469000089', '', true),
  ('trionik', 'LIEPTIONO GUNAWAN', 'BRI', '129401000022567', '129401000022567', '', true),

  ('chytron', 'AGUSTINUS SITUMORANG', 'BCA', '2970426003', '2970426003', 'Otomatis', true),
  ('chytron', 'AGUSTINUS SITUMORANG', 'BNI', '0410114933', '0410114933', 'Konfirmasi ke CS', true),
  ('chytron', 'AGUSTINUS SITUMORANG', 'MANDIRI', '1070007908850', '1070007908850', 'Konfirmasi ke CS', true),

  ('talentapay', 'DEWI RAHAYU', 'BCA', '3570553533', '3570553533', '', true),
  ('talentapay', 'DEWI RAHAYU', 'BRI', '000401001034567', '000401001034567', '', true),
  ('talentapay', 'CV.TALENTAPAY', 'MANDIRI', '1800034434433', '1800034434433', 'Manual', true),

  ('multikom', 'ESFINA GLOBAL MEDIA', 'BCA', '2833787870', '2833787870', '', true),
  ('multikom', 'ESFINA GLOBAL MEDIA', 'BRI', '035401002517305', '035401002517305', '', true),
  ('multikom', 'ESFINA GLOBAL MEDIA', 'MANDIRI', '1310026660664', '1310026660664', '', true),
  ('multikom', 'BISA BERSAMA BERKAH BAROKAH', 'BCA', '2837007778', '2837007778', '', true),

  ('ajs', 'PT. ARDI JAYA SOLUSINDO', 'BCA', '5875784444', '5875784444', '', true),
  ('ajs', 'PT. ARDI JAYA SOLUSINDO', 'BRI', '012001003140304', '012001003140304', '', true),
  ('ajs', 'PT. ARDI JAYA SOLUSINDO', 'BNI', '8698698965', '8698698965', '', true),
  ('ajs', 'PT. ARDI JAYA SOLUSINDO', 'MANDIRI', '1550077766692', '1550077766692', '', true),

  ('loketbayar', 'PT LOKET PEMBAYARAN', 'BCA', '3240789000', '3240789000', '', true),
  ('loketbayar', 'PT LOKET PEMBAYARAN', 'BNI', '2049012472', '2049012472', '', true),
  ('loketbayar', 'PT LOKET PEMBAYARAN', 'BRI', '017201002853567', '017201002853567', '', true),

  ('gemilang', 'ANNISA SAHIYA', 'BCA', '2891407931', '2891407931', '', true),
  ('gemilang', 'ANNISA SAHIYA', 'BRI', '007001003972562', '007001003972562', '', true),

  ('sagaramobile', 'HENDI HADIANSYAH', 'BCA', '2781478581', '2781478581', '', true),
  ('sagaramobile', 'HENDI HADIANSYAH', 'BRI', '065601000351569', '065601000351569', '', true),
  ('sagaramobile', 'HENDI HADIANSYAH', 'MANDIRI', '1320022853551', '1320022853551', '', true)
ON CONFLICT ((lower(trim(provider))), nomor_rekening_digits) DO UPDATE
SET nama = EXCLUDED.nama,
    bank = EXCLUDED.bank,
    nomor_rekening = EXCLUDED.nomor_rekening,
    catatan = EXCLUDED.catatan,
    aktif = EXCLUDED.aktif,
    diubah_pada = now();
