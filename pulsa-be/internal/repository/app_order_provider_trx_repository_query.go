package repository

import (
	"context"
	"database/sql"
	"strings"
)

func (r *AppOrderProviderTrxRepository) Create(ctx context.Context, in AppOrderProviderTrxCreateInput) error {
	return r.db.QueryRowContext(ctx, `
INSERT INTO public.app_order_provider_trx
  (app_order_id, provider, ref_id, harga_provider, status, kode_respon, pesan, sn, raw_request, raw_callback, dibuat_pada, diubah_pada)
VALUES
  ($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10::jsonb,now(),now())
RETURNING id
`, in.AppOrderID, strings.TrimSpace(strings.ToLower(in.Provider)), strings.TrimSpace(in.RefID), in.HargaProvider, strings.TrimSpace(strings.ToLower(in.Status)),
		nullableStringValue(in.KodeRespon), nullableStringValue(in.Pesan), nullableStringValue(in.SN), nullableJSON(in.RawRequest), nullableJSON(in.RawCallback)).Scan(&in.ID)
}

func (r *AppOrderProviderTrxRepository) GetByRefID(ctx context.Context, refID, provider string) (*AppOrderProviderTrxRow, error) {
	var (
		row         AppOrderProviderTrxRow
		kodeRespon  sql.NullString
		pesan       sql.NullString
		sn          sql.NullString
		rawRequest  sql.NullString
		rawCallback sql.NullString
		dibuatPada  sql.NullTime
		diubahPada  sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
SELECT
  id, app_order_id, provider, ref_id, harga_provider, status, kode_respon, pesan,
  sn, raw_request::text, raw_callback::text, dibuat_pada, diubah_pada
FROM public.app_order_provider_trx
WHERE TRIM(ref_id) = $1
  AND LOWER(TRIM(provider)) = $2
ORDER BY id DESC
LIMIT 1
`, strings.TrimSpace(refID), strings.TrimSpace(strings.ToLower(provider))).Scan(
		&row.ID,
		&row.AppOrderID,
		&row.Provider,
		&row.RefID,
		&row.HargaProvider,
		&row.Status,
		&kodeRespon,
		&pesan,
		&sn,
		&rawRequest,
		&rawCallback,
		&dibuatPada,
		&diubahPada,
	)
	if err != nil {
		return nil, err
	}
	if kodeRespon.Valid {
		v := kodeRespon.String
		row.KodeRespon = &v
	}
	if pesan.Valid {
		v := pesan.String
		row.Pesan = &v
	}
	if sn.Valid {
		v := sn.String
		row.SN = &v
	}
	if rawRequest.Valid {
		v := rawRequest.String
		row.RawRequest = &v
	}
	if rawCallback.Valid {
		v := rawCallback.String
		row.RawCallback = &v
	}
	if dibuatPada.Valid {
		v := dibuatPada.Time
		row.DibuatPada = &v
	}
	if diubahPada.Valid {
		v := diubahPada.Time
		row.DiubahPada = &v
	}
	return &row, nil
}

func (r *AppOrderProviderTrxRepository) GetLatestByAppOrderID(ctx context.Context, appOrderID int64) (*AppOrderProviderTrxRow, error) {
	var (
		row         AppOrderProviderTrxRow
		kodeRespon  sql.NullString
		pesan       sql.NullString
		sn          sql.NullString
		rawRequest  sql.NullString
		rawCallback sql.NullString
		dibuatPada  sql.NullTime
		diubahPada  sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
SELECT
  id, app_order_id, provider, ref_id, harga_provider, status, kode_respon, pesan,
  sn, raw_request::text, raw_callback::text, dibuat_pada, diubah_pada
FROM public.app_order_provider_trx
WHERE app_order_id = $1
ORDER BY id DESC
LIMIT 1
`, appOrderID).Scan(
		&row.ID,
		&row.AppOrderID,
		&row.Provider,
		&row.RefID,
		&row.HargaProvider,
		&row.Status,
		&kodeRespon,
		&pesan,
		&sn,
		&rawRequest,
		&rawCallback,
		&dibuatPada,
		&diubahPada,
	)
	if err != nil {
		return nil, err
	}
	if kodeRespon.Valid {
		v := kodeRespon.String
		row.KodeRespon = &v
	}
	if pesan.Valid {
		v := pesan.String
		row.Pesan = &v
	}
	if sn.Valid {
		v := sn.String
		row.SN = &v
	}
	if rawRequest.Valid {
		v := rawRequest.String
		row.RawRequest = &v
	}
	if rawCallback.Valid {
		v := rawCallback.String
		row.RawCallback = &v
	}
	if dibuatPada.Valid {
		v := dibuatPada.Time
		row.DibuatPada = &v
	}
	if diubahPada.Valid {
		v := diubahPada.Time
		row.DiubahPada = &v
	}
	return &row, nil
}

func (r *AppOrderProviderTrxRepository) UpdateResult(ctx context.Context, in AppOrderProviderTrxUpdateInput) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE public.app_order_provider_trx
SET harga_provider = COALESCE($2, harga_provider),
    status = COALESCE(NULLIF($3, ''), status),
    kode_respon = COALESCE($4, kode_respon),
    pesan = COALESCE($5, pesan),
    sn = CASE
      WHEN COALESCE(NULLIF(BTRIM($6), ''), '') = '' THEN sn
      WHEN COALESCE(NULLIF(BTRIM(sn), ''), '') = '' THEN BTRIM($6)
      WHEN LENGTH(BTRIM(sn)) < 8 AND LENGTH(BTRIM($6)) >= 8 THEN BTRIM($6)
      WHEN POSITION('-' IN BTRIM($6)) > 0 AND POSITION('-' IN COALESCE(BTRIM(sn), '')) = 0 THEN BTRIM($6)
      ELSE sn
    END,
    raw_callback = COALESCE($7::jsonb, raw_callback),
    diubah_pada = now()
WHERE id = $1
`, in.ID, in.HargaProvider, nullableStringValue(strings.TrimSpace(strings.ToLower(in.Status))),
		nullableStringValue(in.KodeRespon), nullableStringValue(in.Pesan), nullableStringValue(in.SN), nullableJSON(in.RawCallback))
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}
