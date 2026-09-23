package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"pulsa2/internal/helper"
)

func (r *ProviderCallbackRepository) GetLatestByRefIDProvider(ctx context.Context, refID, provider string) (*ProviderTrxRefRow, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, transaksi_member_id, COALESCE(produk_sku_snapshot,''), produk_provider_map_id, kode_produk, NULLIF(TRIM(request_mentah->>'mode'),''), kode_respon, pesan, no_referensi, harga, COALESCE(status,'pending'), dibuat_pada
FROM public.transaksi_provider
WHERE ref_id = $1 AND provider = $2
ORDER BY id DESC
`, refID, provider)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		bestSuccess *ProviderTrxRefRow
		bestPending *ProviderTrxRefRow
		bestFailed  *ProviderTrxRefRow
		bestUnknown *ProviderTrxRefRow
	)

	for rows.Next() {
		var out ProviderTrxRefRow
		if err := rows.Scan(
			&out.ID,
			&out.TransaksiMemberID,
			&out.ProdukSKUSnapshot,
			&out.ProdukProviderMapID,
			&out.KodeProduk,
			&out.RequestMode,
			&out.KodeRespon,
			&out.Pesan,
			&out.NoReferensi,
			&out.Harga,
			&out.Status,
			&out.DibuatPada,
		); err != nil {
			return nil, err
		}

		rc := ""
		msg := ""
		if out.KodeRespon != nil {
			rc = strings.TrimSpace(*out.KodeRespon)
		}
		if out.Pesan != nil {
			msg = strings.TrimSpace(*out.Pesan)
		}

		switch helper.ProviderResponseStateOf(provider, rc, msg) {
		case helper.ProviderResponseSuccess:
			if bestSuccess == nil || out.ID > bestSuccess.ID {
				row := out
				bestSuccess = &row
			}
		case helper.ProviderResponsePending:
			if bestPending == nil || out.ID > bestPending.ID {
				row := out
				bestPending = &row
			}
		case helper.ProviderResponseFailed:
			if bestFailed == nil || out.ID > bestFailed.ID {
				row := out
				bestFailed = &row
			}
		default:
			if bestUnknown == nil || out.ID > bestUnknown.ID {
				row := out
				bestUnknown = &row
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	switch {
	case bestSuccess != nil:
		return bestSuccess, nil
	case bestPending != nil:
		return bestPending, nil
	case bestFailed != nil:
		return bestFailed, nil
	default:
		return bestUnknown, nil
	}
}

func (r *ProviderCallbackRepository) GetLatestByRefIDProviderCodeHint(ctx context.Context, refID, provider, codeHint string) (*ProviderTrxRefRow, error) {
	codeHint = strings.ToUpper(strings.TrimSpace(codeHint))
	if codeHint == "" {
		return r.GetLatestByRefIDProvider(ctx, refID, provider)
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT id, transaksi_member_id, COALESCE(produk_sku_snapshot,''), produk_provider_map_id, kode_produk, NULLIF(TRIM(request_mentah->>'mode'),''), kode_respon, pesan, no_referensi, harga, COALESCE(status,'pending'), dibuat_pada
FROM public.transaksi_provider
WHERE ref_id = $1 AND provider = $2
  AND (
    UPPER(TRIM(kode_produk)) = $3
    OR UPPER(TRIM(kode_produk)) LIKE ($3 || ':%')
  )
ORDER BY id DESC
`, refID, provider, codeHint)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		bestSuccess *ProviderTrxRefRow
		bestPending *ProviderTrxRefRow
		bestFailed  *ProviderTrxRefRow
		bestUnknown *ProviderTrxRefRow
	)

	for rows.Next() {
		var out ProviderTrxRefRow
		if err := rows.Scan(
			&out.ID,
			&out.TransaksiMemberID,
			&out.ProdukSKUSnapshot,
			&out.ProdukProviderMapID,
			&out.KodeProduk,
			&out.RequestMode,
			&out.KodeRespon,
			&out.Pesan,
			&out.NoReferensi,
			&out.Harga,
			&out.Status,
			&out.DibuatPada,
		); err != nil {
			return nil, err
		}

		rc := ""
		msg := ""
		if out.KodeRespon != nil {
			rc = strings.TrimSpace(*out.KodeRespon)
		}
		if out.Pesan != nil {
			msg = strings.TrimSpace(*out.Pesan)
		}

		switch helper.ProviderResponseStateOf(provider, rc, msg) {
		case helper.ProviderResponseSuccess:
			if bestSuccess == nil || out.ID > bestSuccess.ID {
				row := out
				bestSuccess = &row
			}
		case helper.ProviderResponsePending:
			if bestPending == nil || out.ID > bestPending.ID {
				row := out
				bestPending = &row
			}
		case helper.ProviderResponseFailed:
			if bestFailed == nil || out.ID > bestFailed.ID {
				row := out
				bestFailed = &row
			}
		default:
			if bestUnknown == nil || out.ID > bestUnknown.ID {
				row := out
				bestUnknown = &row
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	switch {
	case bestSuccess != nil:
		return bestSuccess, nil
	case bestPending != nil:
		return bestPending, nil
	case bestFailed != nil:
		return bestFailed, nil
	default:
		return bestUnknown, nil
	}
}

func (r *ProviderCallbackRepository) ListAttemptsByRefID(ctx context.Context, refID string) ([]ProviderAttemptRow, error) {
	refID = strings.TrimSpace(refID)
	if refID == "" {
		return nil, nil
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT id, ref_id, provider, perintah, COALESCE(produk_sku_snapshot,''), produk_provider_map_id, kode_produk, kode_respon, pesan, no_referensi, harga, COALESCE(status,''), COALESCE(http_status,0), dibuat_pada
FROM public.transaksi_provider
WHERE ref_id = $1
ORDER BY id DESC
`, refID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ProviderAttemptRow, 0, 8)
	for rows.Next() {
		var row ProviderAttemptRow
		if err := rows.Scan(&row.ID, &row.RefID, &row.Provider, &row.Perintah, &row.ProdukSKUSnapshot, &row.ProdukProviderMapID, &row.KodeProduk, &row.KodeRespon, &row.Pesan, &row.NoReferensi, &row.Harga, &row.Status, &row.HTTPStatus, &row.DibuatPada); err != nil {
			return nil, err
		}
		row.Provider = strings.ToLower(strings.TrimSpace(row.Provider))
		row.Perintah = strings.ToUpper(strings.TrimSpace(row.Perintah))
		row.ProdukSKUSnapshot = strings.ToUpper(strings.TrimSpace(row.ProdukSKUSnapshot))
		row.KodeProduk = strings.ToUpper(strings.TrimSpace(row.KodeProduk))
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

type UpdateResult struct {
	TrxIDJavapay  *string
	KodeRespon    *string
	Pesan         *string
	NoReferensi   *string
	Harga         *int64
	SaldoTerakhir *int64
	HTTPStatus    *int
	ResponMentah  any
}

func (r *ProviderCallbackRepository) UpdateResult(ctx context.Context, id int64, u UpdateResult) error {
	respJSON, _ := json.Marshal(u.ResponMentah)
	provider := r.getProviderByID(ctx, id)
	currentHTTPStatus := r.getHTTPStatusByID(ctx, id)
	// HTTP 200 = server terima → classify dari pesan
	// HTTP bukan 200 = server TIDAK terima → GAGAL, titik
	status := "failed"
	if u.HTTPStatus == nil || *u.HTTPStatus == 200 {
		status = helper.ProviderResponseStatusString(provider, u.KodeRespon, u.Pesan)
	}
	keepExistingStatus := keepLoketBayarHTTP200OnFailedCheck(provider, currentHTTPStatus, status, u.KodeRespon, u.Pesan)

	const q = `
UPDATE public.transaksi_provider SET
  trx_id_javapay = COALESCE($2, trx_id_javapay),
  kode_respon    = COALESCE($3, kode_respon),
  pesan          = COALESCE($4, pesan),
  no_referensi   = COALESCE($5, no_referensi),
  harga          = COALESCE($6, harga),
  saldo_terakhir = COALESCE($7, saldo_terakhir),
  http_status    = CASE WHEN $11 THEN http_status ELSE COALESCE($8, http_status) END,
  respon_mentah  = $9::jsonb,
  status         = CASE WHEN $11 THEN status ELSE COALESCE(NULLIF($10,''), status) END
WHERE id = $1`

	var httpStatus any
	if u.HTTPStatus != nil {
		httpStatus = *u.HTTPStatus
	}

	_, err := r.db.ExecContext(ctx, q,
		id,
		u.TrxIDJavapay,
		u.KodeRespon,
		u.Pesan,
		u.NoReferensi,
		u.Harga,
		u.SaldoTerakhir,
		httpStatus,
		string(respJSON),
		status,
		keepExistingStatus,
	)
	if err != nil {
		return fmt.Errorf("update public.transaksi_provider: %w", err)
	}
	return nil
}

func (r *ProviderCallbackRepository) ForceFailIfPending(ctx context.Context, id int64, httpStatus int, pesan string, raw any) error {
	respJSON, _ := json.Marshal(raw)
	provider := r.getProviderByID(ctx, id)
	currentHTTPStatus := r.getHTTPStatusByID(ctx, id)
	pesan = strings.TrimSpace(pesan)
	keepExistingStatus := keepLoketBayarHTTP200OnFailedCheck(provider, currentHTTPStatus, "failed", nil, &pesan)
	_, err := r.db.ExecContext(ctx, `
	UPDATE public.transaksi_provider SET
	  http_status   = CASE WHEN $5 THEN http_status ELSE COALESCE($2, http_status) END,
	  pesan         = COALESCE(NULLIF($3,''), pesan),
	  respon_mentah = $4::jsonb,
	  status        = CASE WHEN $5 THEN status ELSE 'failed' END
WHERE id = $1
  AND LOWER(TRIM(COALESCE(status,''))) = 'pending'`,
		id,
		httpStatus,
		pesan,
		string(respJSON),
		keepExistingStatus,
	)
	if err != nil {
		return fmt.Errorf("force fail public.transaksi_provider: %w", err)
	}
	return nil
}

func (r *ProviderCallbackRepository) getHTTPStatusByID(ctx context.Context, id int64) int {
	var httpStatus sql.NullInt64
	if err := r.db.QueryRowContext(ctx, `SELECT http_status FROM public.transaksi_provider WHERE id = $1`, id).Scan(&httpStatus); err != nil || !httpStatus.Valid {
		return 0
	}
	return int(httpStatus.Int64)
}

type ProviderTrxCreateIn struct {
	Provider            string
	TransaksiMemberID   int64
	RefID               string
	Perintah            string
	ProdukSKUSnapshot   string
	ProdukProviderMapID *int64
	KodeProduk          string
	Tujuan              string
	Qty                 int64
}

type ProviderTrxRow struct {
	ID                  int64
	TransaksiMemberID   int64
	RefID               string
	Perintah            string
	ProdukSKUSnapshot   string
	ProdukProviderMapID *int64
	KodeProduk          string
	Tujuan              string
	Qty                 int64
	Percobaan           int
	DibuatPada          time.Time
}

func (r *ProviderCallbackRepository) FindProviderTrxByRoute(ctx context.Context, in ProviderTrxCreateIn) (*ProviderTrxRow, error) {
	in.Provider = strings.TrimSpace(in.Provider)
	if in.Provider == "" {
		in.Provider = "javapay"
	}

	const q = `
SELECT id, transaksi_member_id, ref_id, perintah, produk_sku_snapshot, produk_provider_map_id,
       kode_produk, tujuan, qty, percobaan, dibuat_pada
FROM public.transaksi_provider
WHERE transaksi_member_id = $1
  AND ref_id = $2
  AND lower(trim(provider)) = lower(trim($3))
  AND perintah = $4
  AND (
    ($5::bigint IS NOT NULL AND produk_provider_map_id = $5)
    OR ($5::bigint IS NULL AND produk_provider_map_id IS NULL AND upper(trim(kode_produk)) = upper(trim($6)))
  )

  AND NOT (
    lower(trim($3)) = 'loketbayar'
    AND lower(trim(provider)) = 'loketbayar'
    AND lower(trim(COALESCE(status,''))) = 'pending'
    AND COALESCE(http_status,0) <> 200
    AND dibuat_pada <= (NOW() - INTERVAL '1 minute')
  )
ORDER BY id DESC
LIMIT 1`
	var row ProviderTrxRow
	err := r.db.QueryRowContext(ctx, q,
		in.TransaksiMemberID,
		strings.TrimSpace(in.RefID),
		in.Provider,
		in.Perintah,
		in.ProdukProviderMapID,
		in.KodeProduk,
	).Scan(
		&row.ID,
		&row.TransaksiMemberID,
		&row.RefID,
		&row.Perintah,
		&row.ProdukSKUSnapshot,
		&row.ProdukProviderMapID,
		&row.KodeProduk,
		&row.Tujuan,
		&row.Qty,
		&row.Percobaan,
		&row.DibuatPada,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ProviderCallbackRepository) getProviderByID(ctx context.Context, id int64) string {
	var provider sql.NullString
	if err := r.db.QueryRowContext(ctx, `SELECT provider FROM public.transaksi_provider WHERE id = $1`, id).Scan(&provider); err != nil {
		return ""
	}
	return strings.TrimSpace(provider.String)
}

func (r *ProviderCallbackRepository) CreateProviderTrx(ctx context.Context, in ProviderTrxCreateIn, requestMentah any) (*ProviderTrxRow, error) {
	b, _ := json.Marshal(requestMentah)
	in.Provider = strings.TrimSpace(in.Provider)
	if in.Provider == "" {
		in.Provider = "javapay"
	}
	if strings.TrimSpace(in.RefID) != "" {
		if existing, err := r.FindProviderTrxByRoute(ctx, in); err != nil {
			return nil, err
		} else if existing != nil {
			return existing, nil
		}
	}

	var (
		id     int64
		refID  string
		dibuat time.Time
	)
	if strings.TrimSpace(in.RefID) == "" {
		err := r.db.QueryRowContext(ctx, `
INSERT INTO public.transaksi_provider
  (provider, transaksi_member_id, perintah, produk_sku_snapshot, produk_provider_map_id, kode_produk, tujuan, qty, request_mentah, percobaan, status)
VALUES
  ($1,$2,$3,NULLIF($4,''),$5,$6,$7,$8,$9::jsonb,1,'pending')
RETURNING id, ref_id, dibuat_pada
`, in.Provider, in.TransaksiMemberID, in.Perintah, in.ProdukSKUSnapshot, in.ProdukProviderMapID, in.KodeProduk, in.Tujuan, in.Qty, string(b)).Scan(&id, &refID, &dibuat)
		if err != nil {
			return nil, err
		}
	} else {
		err := r.db.QueryRowContext(ctx, `
INSERT INTO public.transaksi_provider
  (provider, transaksi_member_id, ref_id, perintah, produk_sku_snapshot, produk_provider_map_id, kode_produk, tujuan, qty, request_mentah, percobaan, status)
VALUES
  ($1,$2,$3,$4,NULLIF($5,''),$6,$7,$8,$9,$10::jsonb,1,'pending')
RETURNING id, ref_id, dibuat_pada
`, in.Provider, in.TransaksiMemberID, in.RefID, in.Perintah, in.ProdukSKUSnapshot, in.ProdukProviderMapID, in.KodeProduk, in.Tujuan, in.Qty, string(b)).Scan(&id, &refID, &dibuat)
		if err != nil {
			return nil, err
		}
	}

	return &ProviderTrxRow{
		ID:                  id,
		TransaksiMemberID:   in.TransaksiMemberID,
		RefID:               refID,
		Perintah:            in.Perintah,
		ProdukSKUSnapshot:   strings.TrimSpace(in.ProdukSKUSnapshot),
		ProdukProviderMapID: in.ProdukProviderMapID,
		KodeProduk:          in.KodeProduk,
		Tujuan:              in.Tujuan,
		Qty:                 in.Qty,
		Percobaan:           1,
		DibuatPada:          dibuat,
	}, nil
}

func (r *ProviderCallbackRepository) UpdateRequestMentah(ctx context.Context, id int64, req any) error {
	b, _ := json.Marshal(req)
	_, err := r.db.ExecContext(ctx, `
UPDATE public.transaksi_provider
SET request_mentah = $2::jsonb
WHERE id = $1
`, id, string(b))
	return err
}
