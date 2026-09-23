ALTER TABLE public.transaksi_provider
ADD COLUMN IF NOT EXISTS status VARCHAR(16);

UPDATE public.transaksi_provider tp
SET status = (
  CASE
    WHEN lower(trim(coalesce(tp.provider,''))) = 'smb' AND upper(coalesce(tp.pesan,'')) LIKE '%INQSUKSES%' THEN 'pending'
    WHEN ((lower(trim(coalesce(tp.provider,''))) = 'smb' AND trim(coalesce(tp.kode_respon,'')) = '1' AND upper(coalesce(tp.pesan,'')) NOT LIKE '%INQSUKSES%') OR trim(coalesce(tp.kode_respon,'')) = '20' OR upper(coalesce(tp.pesan,'')) LIKE '%SUKSES%') AND NOT (
      upper(coalesce(tp.pesan,'')) LIKE '%GAGAL%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%FAILED%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%ERROR%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%REJECT%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%TIDAK DIPROSES%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%CUTOFF%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%DIBATALKAN%' OR
      upper(coalesce(tp.pesan,'')) LIKE '% BATAL %' OR
      upper(coalesce(tp.pesan,'')) LIKE '%STATUS TIMEOUT%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%NOMOR TUJUAN SALAH%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%PRODUK SALAH%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%STOK TIDAK CUKUP%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%STOK SEDANG KOSONG/DITUTUP%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%PRODUK SEDANG GANGGUAN%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%STOK DIKEMBALIKAN%' OR
      (upper(coalesce(tp.pesan,'')) LIKE '%SALDO%' AND upper(coalesce(tp.pesan,'')) LIKE '%TIDAK%' AND upper(coalesce(tp.pesan,'')) LIKE '%CUKUP%') OR
      upper(coalesce(tp.pesan,'')) LIKE '%LIMIT 0%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%ALLOWED QTY%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%QTY TIDAK SESUAI%'
    ) THEN 'success'
    WHEN (trim(coalesce(tp.kode_respon,'')) IN ('0','1','2','3') OR upper(coalesce(tp.pesan,'')) LIKE '%PROSES%' OR upper(coalesce(tp.pesan,'')) LIKE '%PENDING%' OR upper(coalesce(tp.pesan,'')) LIKE '%AKAN DIPROSES%' OR upper(coalesce(tp.pesan,'')) LIKE '%SEDANG DIPROSES%') AND NOT ((lower(trim(coalesce(tp.provider,''))) = 'smb' AND trim(coalesce(tp.kode_respon,'')) = '1' AND upper(coalesce(tp.pesan,'')) NOT LIKE '%INQSUKSES%') OR trim(coalesce(tp.kode_respon,'')) = '20' OR upper(coalesce(tp.pesan,'')) LIKE '%SUKSES%') AND NOT (
      upper(coalesce(tp.pesan,'')) LIKE '%GAGAL%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%FAILED%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%ERROR%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%REJECT%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%TIDAK DIPROSES%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%CUTOFF%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%DIBATALKAN%' OR
      upper(coalesce(tp.pesan,'')) LIKE '% BATAL %' OR
      upper(coalesce(tp.pesan,'')) LIKE '%STATUS TIMEOUT%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%NOMOR TUJUAN SALAH%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%PRODUK SALAH%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%STOK TIDAK CUKUP%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%STOK SEDANG KOSONG/DITUTUP%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%PRODUK SEDANG GANGGUAN%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%STOK DIKEMBALIKAN%' OR
      (upper(coalesce(tp.pesan,'')) LIKE '%SALDO%' AND upper(coalesce(tp.pesan,'')) LIKE '%TIDAK%' AND upper(coalesce(tp.pesan,'')) LIKE '%CUKUP%') OR
      upper(coalesce(tp.pesan,'')) LIKE '%LIMIT 0%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%ALLOWED QTY%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%QTY TIDAK SESUAI%'
    ) THEN 'pending'
    WHEN (
      upper(coalesce(tp.pesan,'')) LIKE '%GAGAL%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%FAILED%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%ERROR%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%REJECT%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%TIDAK DIPROSES%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%CUTOFF%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%DIBATALKAN%' OR
      upper(coalesce(tp.pesan,'')) LIKE '% BATAL %' OR
      upper(coalesce(tp.pesan,'')) LIKE '%STATUS TIMEOUT%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%NOMOR TUJUAN SALAH%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%PRODUK SALAH%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%STOK TIDAK CUKUP%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%STOK SEDANG KOSONG/DITUTUP%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%PRODUK SEDANG GANGGUAN%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%STOK DIKEMBALIKAN%' OR
      (upper(coalesce(tp.pesan,'')) LIKE '%SALDO%' AND upper(coalesce(tp.pesan,'')) LIKE '%TIDAK%' AND upper(coalesce(tp.pesan,'')) LIKE '%CUKUP%') OR
      upper(coalesce(tp.pesan,'')) LIKE '%LIMIT 0%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%ALLOWED QTY%' OR
      upper(coalesce(tp.pesan,'')) LIKE '%QTY TIDAK SESUAI%' OR
      (trim(coalesce(tp.kode_respon,'')) NOT IN ('', '0', '1', '2', '3', '20') AND NOT (trim(coalesce(tp.kode_respon,'')) IN ('0','1','2','3') OR upper(coalesce(tp.pesan,'')) LIKE '%PROSES%' OR upper(coalesce(tp.pesan,'')) LIKE '%PENDING%' OR upper(coalesce(tp.pesan,'')) LIKE '%AKAN DIPROSES%' OR upper(coalesce(tp.pesan,'')) LIKE '%SEDANG DIPROSES%'))
    ) THEN 'failed'
    ELSE 'unknown'
  END
)
WHERE coalesce(trim(tp.status),'') = '';

ALTER TABLE public.transaksi_provider
ALTER COLUMN status SET DEFAULT 'pending';

CREATE INDEX IF NOT EXISTS transaksi_provider_status_dibuat_pada_idx
  ON public.transaksi_provider (status, dibuat_pada DESC);
