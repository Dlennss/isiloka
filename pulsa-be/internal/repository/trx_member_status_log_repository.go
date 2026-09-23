package repository

import (
	"context"
	"database/sql"
)

type dbtx interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type TrxMemberStatusLogInput struct {
	TrxID              int64
	StatusSebelum      string
	StatusSesudah      string
	KeteranganSebelum  string
	KeteranganSesudah  string
	BiayaAktualSebelum int64
	BiayaAktualSesudah int64
	Aksi               string
	DiubahOleh         *int64
}

func updateTrxMemberStatusWithAudit(
	ctx context.Context,
	exec dbtx,
	trxID int64,
	status string,
	keterangan string,
	biayaAktual int64,
	aksi string,
	diubahOleh *int64,
) error {
	const q = `
WITH prev AS (
  SELECT
    id,
    status,
    COALESCE(keterangan, '') AS keterangan,
    COALESCE(biaya_aktual, 0) AS biaya_aktual
  FROM public.transaksi_member
  WHERE id = $1
),
upd AS (
  UPDATE public.transaksi_member
  SET status = $2,
      keterangan = NULLIF($3,''),
      biaya_aktual = $4
  WHERE id = $1
    AND (
      LOWER(COALESCE(status, '')) NOT IN ('success', 'failed')
      OR LOWER(COALESCE(status, '')) = LOWER($2)
    )
  RETURNING id
)
INSERT INTO public.transaksi_member_status_log
  (transaksi_member_id, status_sebelum, status_sesudah, keterangan_sebelum, keterangan_sesudah, biaya_aktual_sebelum, biaya_aktual_sesudah, aksi, diubah_oleh)
SELECT
  prev.id, prev.status, $2, prev.keterangan, COALESCE(NULLIF($3,''), ''), prev.biaya_aktual, $4, $5, $6
FROM prev
WHERE EXISTS (SELECT 1 FROM upd)
`
	res, err := exec.ExecContext(ctx, q, trxID, status, keterangan, biayaAktual, aksi, diubahOleh)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err == nil && aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func updateTrxMemberSettleWithAudit(
	ctx context.Context,
	exec dbtx,
	trxID int64,
	status string,
	keterangan string,
	biayaAktual int64,
	hargaJavapay int64,
	hargaMember int64,
	aksi string,
) error {
	const q = `
WITH prev AS (
  SELECT
    id,
    status,
    COALESCE(keterangan, '') AS keterangan,
    COALESCE(biaya_aktual, 0) AS biaya_aktual
  FROM public.transaksi_member
  WHERE id = $1
),
upd AS (
  UPDATE public.transaksi_member
  SET status = $2,
      keterangan = NULLIF($3,''),
      biaya_aktual = $4,
      harga_javapay = $5,
      harga_member = $6
  WHERE id = $1
    AND (
      LOWER(COALESCE(status, '')) NOT IN ('success', 'failed')
      OR LOWER(COALESCE(status, '')) = LOWER($2)
    )
  RETURNING id
)
INSERT INTO public.transaksi_member_status_log
  (transaksi_member_id, status_sebelum, status_sesudah, keterangan_sebelum, keterangan_sesudah, biaya_aktual_sebelum, biaya_aktual_sesudah, aksi, diubah_oleh)
SELECT
  prev.id, prev.status, $2, prev.keterangan, COALESCE(NULLIF($3,''), ''), prev.biaya_aktual, $4, $7, NULL
FROM prev
WHERE EXISTS (SELECT 1 FROM upd)
`
	res, err := exec.ExecContext(ctx, q, trxID, status, keterangan, biayaAktual, hargaJavapay, hargaMember, aksi)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err == nil && aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func insertTrxMemberStatusLog(ctx context.Context, exec dbtx, in TrxMemberStatusLogInput) error {
	const q = `
INSERT INTO public.transaksi_member_status_log
  (transaksi_member_id, status_sebelum, status_sesudah, keterangan_sebelum, keterangan_sesudah, biaya_aktual_sebelum, biaya_aktual_sesudah, aksi, diubah_oleh)
VALUES
  ($1, NULLIF($2,''), NULLIF($3,''), NULLIF($4,''), NULLIF($5,''), $6, $7, NULLIF($8,''), $9)
`
	_, err := exec.ExecContext(
		ctx,
		q,
		in.TrxID,
		in.StatusSebelum,
		in.StatusSesudah,
		in.KeteranganSebelum,
		in.KeteranganSesudah,
		in.BiayaAktualSebelum,
		in.BiayaAktualSesudah,
		in.Aksi,
		in.DiubahOleh,
	)
	return err
}
