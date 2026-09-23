package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/model"
)

func (r *MemberTrxProviderTrxRepository) GetLatestByRefIDProvider(ctx context.Context, refID, provider string) (*model.JavapayTrxRow, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT
  id, transaksi_member_id, provider, ref_id, perintah, produk_sku_snapshot, produk_provider_map_id, kode_produk, tujuan, qty,
  trx_id_javapay, kode_respon, pesan, no_referensi, harga, saldo_terakhir,
  request_mentah::text, respon_mentah::text, http_status, percobaan, dibuat_pada, COALESCE(status, 'pending') AS status
FROM public.transaksi_provider
WHERE ref_id = $1 AND provider = $2
ORDER BY id DESC
`, refID, provider)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		bestSuccess *rowScan
		bestPending *rowScan
		bestFailed  *rowScan
		bestUnknown *rowScan
	)

	for rows.Next() {
		var s rowScan
		if err := rows.Scan(
			&s.id, &s.transaksiMemberID, &s.provider, &s.refID, &s.perintah, &s.produkSKUSnapshot, &s.produkProviderMapID, &s.kodeProduk, &s.tujuan, &s.qty,
			&s.trxIDJavapay, &s.kodeRespon, &s.pesan, &s.noReferensi, &s.harga, &s.saldoTerakhir,
			&s.requestMentah, &s.responMentah, &s.httpStatus, &s.percobaan, &s.dibuatPada, &s.status,
		); err != nil {
			return nil, err
		}

		rc := ""
		msg := ""
		if s.kodeRespon.Valid {
			rc = strings.TrimSpace(s.kodeRespon.String)
		}
		if s.pesan.Valid {
			msg = strings.TrimSpace(s.pesan.String)
		}

		switch helper.ProviderResponseStateOf(provider, rc, msg) {
		case helper.ProviderResponseSuccess:
			if bestSuccess == nil || s.id > bestSuccess.id {
				row := s
				bestSuccess = &row
			}
		case helper.ProviderResponsePending:
			if bestPending == nil || s.id > bestPending.id {
				row := s
				bestPending = &row
			}
		case helper.ProviderResponseFailed:
			if bestFailed == nil || s.id > bestFailed.id {
				row := s
				bestFailed = &row
			}
		default:
			if bestUnknown == nil || s.id > bestUnknown.id {
				row := s
				bestUnknown = &row
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	switch {
	case bestSuccess != nil:
		return scanToModel(bestSuccess), nil
	case bestPending != nil:
		return scanToModel(bestPending), nil
	case bestFailed != nil:
		return scanToModel(bestFailed), nil
	case bestUnknown != nil:
		return scanToModel(bestUnknown), nil
	default:
		return nil, nil
	}
}

func (r *MemberTrxProviderTrxRepository) GetByID(ctx context.Context, id int64) (*model.JavapayTrxRow, error) {
	if id <= 0 {
		return nil, nil
	}
	const q = `
SELECT
  id, transaksi_member_id, provider, ref_id, perintah, produk_sku_snapshot, produk_provider_map_id, kode_produk, tujuan, qty,
  trx_id_javapay, kode_respon, pesan, no_referensi, harga, saldo_terakhir,
  request_mentah::text, respon_mentah::text, http_status, percobaan, dibuat_pada, COALESCE(status, 'pending') AS status
FROM public.transaksi_provider
WHERE id = $1
LIMIT 1`
	var s rowScan
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&s.id, &s.transaksiMemberID, &s.provider, &s.refID, &s.perintah, &s.produkSKUSnapshot, &s.produkProviderMapID, &s.kodeProduk, &s.tujuan, &s.qty,
		&s.trxIDJavapay, &s.kodeRespon, &s.pesan, &s.noReferensi, &s.harga, &s.saldoTerakhir,
		&s.requestMentah, &s.responMentah, &s.httpStatus, &s.percobaan, &s.dibuatPada, &s.status,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return scanToModel(&s), nil
}

func (r *MemberTrxProviderTrxRepository) ExistsByRefID(ctx context.Context, refID string) (bool, error) {
	refID = strings.TrimSpace(refID)
	if refID == "" {
		return false, nil
	}
	const q = `
SELECT 1
FROM public.transaksi_provider
WHERE ref_id = $1
LIMIT 1`
	var one int
	err := r.db.QueryRowContext(ctx, q, refID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *MemberTrxProviderTrxRepository) ExistsByTransaksiMemberID(ctx context.Context, trxMemberID int64) (bool, error) {
	if trxMemberID <= 0 {
		return false, nil
	}
	const q = `
SELECT 1
FROM public.transaksi_provider
WHERE transaksi_member_id = $1
LIMIT 1`
	var one int
	err := r.db.QueryRowContext(ctx, q, trxMemberID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *MemberTrxProviderTrxRepository) ListProvidersByTransaksiMemberID(ctx context.Context, trxMemberID int64) ([]string, error) {
	if trxMemberID <= 0 {
		return nil, nil
	}
	const q = `
SELECT DISTINCT lower(trim(provider))
FROM public.transaksi_provider
WHERE transaksi_member_id = $1
  AND coalesce(trim(provider), '') <> ''
ORDER BY 1`
	rows, err := r.db.QueryContext(ctx, q, trxMemberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]string, 0, 8)
	for rows.Next() {
		var provider string
		if err := rows.Scan(&provider); err != nil {
			return nil, err
		}
		provider = strings.TrimSpace(strings.ToLower(provider))
		if provider == "" {
			continue
		}
		out = append(out, provider)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *MemberTrxProviderTrxRepository) ListByRefID(ctx context.Context, refID string) ([]*model.JavapayTrxRow, error) {
	refID = strings.TrimSpace(refID)
	if refID == "" {
		return nil, nil
	}
	const q = `
SELECT
  id, transaksi_member_id, provider, ref_id, perintah, produk_sku_snapshot, produk_provider_map_id, kode_produk, tujuan, qty,
  trx_id_javapay, kode_respon, pesan, no_referensi, harga, saldo_terakhir,
  request_mentah::text, respon_mentah::text, http_status, percobaan, dibuat_pada, COALESCE(status, 'pending') AS status
FROM public.transaksi_provider
WHERE ref_id = $1
ORDER BY id DESC`
	rows, err := r.db.QueryContext(ctx, q, refID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*model.JavapayTrxRow, 0, 8)
	for rows.Next() {
		var s rowScan
		if err := rows.Scan(
			&s.id, &s.transaksiMemberID, &s.provider, &s.refID, &s.perintah, &s.produkSKUSnapshot, &s.produkProviderMapID, &s.kodeProduk, &s.tujuan, &s.qty,
			&s.trxIDJavapay, &s.kodeRespon, &s.pesan, &s.noReferensi, &s.harga, &s.saldoTerakhir,
			&s.requestMentah, &s.responMentah, &s.httpStatus, &s.percobaan, &s.dibuatPada, &s.status,
		); err != nil {
			return nil, err
		}
		out = append(out, scanToModel(&s))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *MemberTrxProviderTrxRepository) MarkPendingRowsAsFailed(ctx context.Context, refID, reason string) (int64, error) {
	refID = strings.TrimSpace(refID)
	if refID == "" {
		return 0, nil
	}
	res, err := r.db.ExecContext(ctx, `
UPDATE public.transaksi_provider
SET status = 'failed', kode_respon = '91', pesan = $2
WHERE ref_id = $1 AND status = 'pending'`, refID, reason)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *MemberTrxProviderTrxRepository) MarkPendingRowsAsFailedByID(ctx context.Context, id int64, reason string) (int64, error) {
	if id <= 0 {
		return 0, nil
	}
	res, err := r.db.ExecContext(ctx, `
UPDATE public.transaksi_provider
SET status = 'failed', kode_respon = '91'
WHERE id = $1 AND status = 'pending'`, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
