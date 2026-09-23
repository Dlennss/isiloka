package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"pulsa2/db"
	"pulsa2/internal/helper"
	"pulsa2/model"
)

func (r *MemberTrxProviderTrxRepository) Create(ctx context.Context, in model.JavapayTrxCreateIn, requestMentah any) (*model.JavapayTrxRow, error) {
	b, _ := json.Marshal(requestMentah)

	in.Provider = strings.TrimSpace(in.Provider)
	if in.Provider == "" {
		in.Provider = "javapay"
	}

	// Kalau sudah ada row pending dengan refid+provider sama, pakai row itu — jangan buat baru
	if in.RefID != "" {
		var existID int64
		err := r.db.QueryRowContext(ctx, `
SELECT id FROM public.transaksi_provider
WHERE ref_id = $1 AND provider = $2 AND status = 'pending'
ORDER BY id DESC LIMIT 1`,
			in.RefID, in.Provider).Scan(&existID)
		if err == nil && existID > 0 {
			row, gErr := r.GetByID(ctx, existID)
			if gErr == nil && row != nil {
				return row, nil
			}
		}
	}

	var (
		id     int64
		refID  string
		dibuat time.Time
	)

	if in.RefID == "" {
		const q = `
INSERT INTO public.transaksi_provider
  (provider, transaksi_member_id, perintah, produk_sku_snapshot, produk_provider_map_id, kode_produk, tujuan, qty, request_mentah, percobaan, status)
VALUES
  ($1,$2,$3,NULLIF($4,''),$5,$6,$7,$8, $9::jsonb, 1, 'pending')
RETURNING id, ref_id, dibuat_pada`
		err := r.db.QueryRowContext(ctx, q,
			in.Provider, in.TransaksiMemberID, in.Perintah, in.ProdukSKUSnapshot, in.ProdukProviderMapID, in.KodeProduk, in.Tujuan, in.Qty, string(b),
		).Scan(&id, &refID, &dibuat)
		if err != nil {
			return nil, err
		}
	} else {
		const q = `
INSERT INTO public.transaksi_provider
  (provider, transaksi_member_id, ref_id, perintah, produk_sku_snapshot, produk_provider_map_id, kode_produk, tujuan, qty, request_mentah, percobaan, status)
VALUES
  ($1,$2,$3,$4,NULLIF($5,''),$6,$7,$8,$9, $10::jsonb, 1, 'pending')
RETURNING id, ref_id, dibuat_pada`
		err := r.db.QueryRowContext(ctx, q,
			in.Provider, in.TransaksiMemberID, in.RefID, in.Perintah, in.ProdukSKUSnapshot, in.ProdukProviderMapID, in.KodeProduk, in.Tujuan, in.Qty, string(b),
		).Scan(&id, &refID, &dibuat)
		if err != nil {
			return nil, err
		}
	}

	return &model.JavapayTrxRow{
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
		RequestMentah:       requestMentah,
	}, nil
}

func (r *MemberTrxProviderTrxRepository) UpdateRequestMentah(ctx context.Context, id int64, req any) error {
	b, _ := json.Marshal(req)
	_, err := r.db.ExecContext(ctx,
		`UPDATE public.transaksi_provider SET request_mentah = $2::jsonb WHERE id = $1`,
		id, string(b),
	)
	if err != nil {
		return fmt.Errorf("update public.transaksi_provider request_mentah: %w", err)
	}
	return nil
}

func (r *MemberTrxProviderTrxRepository) UpdateResult(ctx context.Context, id int64, u db.UpdateResult) error {
	respJSON, _ := json.Marshal(u.ResponMentah)
	provider := r.getProviderByID(ctx, id)
	currentHTTPStatus := r.getHTTPStatusByID(ctx, id)
	status := providerUpdateResultStatus(provider, u.HTTPStatus, u.KodeRespon, u.Pesan)
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
	} else {
		httpStatus = nil
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
	if !keepExistingStatus && (status == "success" || status == "failed") {
		_ = r.closeOlderPendingByRefProvider(ctx, id, provider, status)
	}
	return nil
}

func keepLoketBayarHTTP200OnFailedCheck(provider string, currentHTTPStatus int, nextStatus string, kodeRespon, pesan *string) bool {
	if !strings.EqualFold(strings.TrimSpace(provider), "loketbayar") ||
		currentHTTPStatus != 200 ||
		!strings.EqualFold(strings.TrimSpace(nextStatus), "failed") {
		return false
	}
	return !isLoketBayarExplicitFailedStatus(kodeRespon, pesan)
}

func isLoketBayarExplicitFailedStatus(kodeRespon, pesan *string) bool {
	rc := ""
	if kodeRespon != nil {
		rc = strings.TrimSpace(strings.ToUpper(*kodeRespon))
	}
	msg := ""
	if pesan != nil {
		msg = strings.TrimSpace(strings.ToUpper(*pesan))
	}
	if rc == "GAGAL" || rc == "FAILED" {
		return true
	}
	return strings.Contains(msg, "STATUS GAGAL") ||
		strings.Contains(msg, "TRANSAKSI GAGAL DI PROVIDER")
}

func providerUpdateResultStatus(provider string, httpStatus *int, kodeRespon, pesan *string) string {
	// HTTP status 0 means no HTTP response was received. That is not a final
	// provider failure; keep using the provider message/RC classifier.
	if httpStatus != nil && *httpStatus > 0 && *httpStatus != 200 {
		return "failed"
	}
	return helper.ProviderResponseStatusString(provider, kodeRespon, pesan)
}

func (r *MemberTrxProviderTrxRepository) ForceFailIfPending(ctx context.Context, id int64, httpStatus int, pesan string, raw any) error {
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

func (r *MemberTrxProviderTrxRepository) getHTTPStatusByID(ctx context.Context, id int64) int {
	var httpStatus sql.NullInt64
	if err := r.db.QueryRowContext(ctx, `SELECT http_status FROM public.transaksi_provider WHERE id = $1`, id).Scan(&httpStatus); err != nil || !httpStatus.Valid {
		return 0
	}
	return int(httpStatus.Int64)
}

func (r *MemberTrxProviderTrxRepository) closeOlderPendingByRefProvider(ctx context.Context, currentID int64, provider, finalStatus string) error {
	if currentID <= 0 {
		return nil
	}
	provider = strings.TrimSpace(strings.ToLower(provider))
	if provider == "" {
		return nil
	}
	finalStatus = strings.TrimSpace(strings.ToLower(finalStatus))
	if finalStatus != "success" && finalStatus != "failed" {
		return nil
	}
	msg := "Provider stale pending cleanup; digantikan row final lain"
	if finalStatus == "success" {
		msg = "Provider stale pending cleanup; provider sudah sukses di row lain"
	}
	const q = `
UPDATE public.transaksi_provider older
SET
  status = 'failed',
  kode_respon = CASE WHEN COALESCE(TRIM(older.kode_respon), '') = '' THEN '91' ELSE older.kode_respon END,
  pesan = CASE WHEN COALESCE(TRIM(older.pesan), '') = '' THEN $3 ELSE $4 END
FROM public.transaksi_provider cur
WHERE cur.id = $1
  AND lower(trim(coalesce(cur.provider, ''))) = $2
  AND older.id < cur.id
  AND older.ref_id = cur.ref_id
  AND lower(trim(coalesce(older.provider, ''))) = $2
  AND lower(trim(coalesce(older.status, ''))) = 'pending'`
	_, err := r.db.ExecContext(ctx, q, currentID, provider, msg, "Provider stale pending cleanup; digantikan row final lain")
	return err
}
