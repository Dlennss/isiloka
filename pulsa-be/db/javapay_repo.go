package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"pulsa2/internal/helper"
	"pulsa2/model"
	"strings"
	"time"
)

type JavapayRepo struct {
	DB *sql.DB
}

func NewJavapayRepo(db *sql.DB) *JavapayRepo { return &JavapayRepo{DB: db} }

type rowScan struct {
	id                  int64
	transaksiMemberID   int64
	provider            string
	refID               string
	perintah            string
	produkSKUSnapshot   sql.NullString
	produkProviderMapID sql.NullInt64
	kodeProduk          string
	tujuan              string
	qty                 int64

	trxIDJavapay  sql.NullString
	kodeRespon    sql.NullString
	pesan         sql.NullString
	noReferensi   sql.NullString
	harga         sql.NullInt64
	saldoTerakhir sql.NullInt64

	requestMentah sql.NullString
	responMentah  sql.NullString
	httpStatus    sql.NullInt64
	percobaan     int
	dibuatPada    time.Time
	status        string
}

func toPtrStr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}
func toPtrI64(ni sql.NullInt64) *int64 {
	if ni.Valid {
		return &ni.Int64
	}
	return nil
}
func toPtrInt(ni sql.NullInt64) *int {
	if ni.Valid {
		v := int(ni.Int64)
		return &v
	}
	return nil
}

func parseJSONText(s sql.NullString) any {
	if !s.Valid || s.String == "" {
		return nil
	}
	var v any
	if err := json.Unmarshal([]byte(s.String), &v); err != nil {
		return map[string]any{"_raw": s.String}
	}
	return v
}

func scanToModel(s *rowScan) *model.JavapayTrxRow {
	return &model.JavapayTrxRow{
		ID:                  s.id,
		TransaksiMemberID:   s.transaksiMemberID,
		Provider:            s.provider,
		RefID:               s.refID,
		Perintah:            s.perintah,
		ProdukSKUSnapshot:   strings.TrimSpace(s.produkSKUSnapshot.String),
		ProdukProviderMapID: toPtrI64(s.produkProviderMapID),
		KodeProduk:          s.kodeProduk,
		Tujuan:              s.tujuan,
		Qty:                 s.qty,
		TrxIDJavapay:        toPtrStr(s.trxIDJavapay),
		KodeRespon:          toPtrStr(s.kodeRespon),
		Pesan:               toPtrStr(s.pesan),
		NoReferensi:         toPtrStr(s.noReferensi),
		Harga:               toPtrI64(s.harga),
		SaldoTerakhir:       toPtrI64(s.saldoTerakhir),
		HTTPStatus:          toPtrInt(s.httpStatus),
		Percobaan:           s.percobaan,
		DibuatPada:          s.dibuatPada,
		RequestMentah:       parseJSONText(s.requestMentah),
		ResponMentah:        parseJSONText(s.responMentah),
		Status:              s.status,
	}
}

func (r *JavapayRepo) getProviderByID(ctx context.Context, id int64) string {
	var provider sql.NullString
	if err := r.DB.QueryRowContext(ctx, `SELECT provider FROM public.transaksi_provider WHERE id = $1`, id).Scan(&provider); err != nil {
		return ""
	}
	return strings.TrimSpace(provider.String)
}

func (r *JavapayRepo) GetLatestByRefID(ctx context.Context, refID string) (*model.JavapayTrxRow, error) {
	const q = `
SELECT
  id, transaksi_member_id, provider, ref_id, perintah, produk_sku_snapshot, produk_provider_map_id, kode_produk, tujuan, qty,
  trx_id_javapay, kode_respon, pesan, no_referensi, harga, saldo_terakhir,
  request_mentah::text, respon_mentah::text, http_status, percobaan, dibuat_pada, COALESCE(status, 'pending') AS status
FROM public.transaksi_provider
WHERE ref_id = $1
ORDER BY id DESC
LIMIT 1`

	var s rowScan
	err := r.DB.QueryRowContext(ctx, q, refID).Scan(
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

// KUNCI: idempotency = (ref_id + perintah)
func (r *JavapayRepo) GetLatestByRefIDAndPerintah(ctx context.Context, refID, perintah string) (*model.JavapayTrxRow, error) {
	const q = `
SELECT
  id, transaksi_member_id, provider, ref_id, perintah, produk_sku_snapshot, produk_provider_map_id, kode_produk, tujuan, qty,
  trx_id_javapay, kode_respon, pesan, no_referensi, harga, saldo_terakhir,
  request_mentah::text, respon_mentah::text, http_status, percobaan, dibuat_pada, COALESCE(status, 'pending') AS status
FROM public.transaksi_provider
WHERE ref_id = $1 AND perintah = $2
ORDER BY id DESC
LIMIT 1`

	var s rowScan
	err := r.DB.QueryRowContext(ctx, q, refID, perintah).Scan(
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

// Create: ref_id optional.
// - Jika in.RefID kosong -> DB generate via DEFAULT, kita ambil via RETURNING ref_id
// - Jika in.RefID terisi -> dipakai sesuai input (harus unik)
func (r *JavapayRepo) Create(ctx context.Context, in model.JavapayTrxCreateIn, requestMentah any) (*model.JavapayTrxRow, error) {
	b, _ := json.Marshal(requestMentah)

	in.Provider = strings.TrimSpace(in.Provider)
	if in.Provider == "" {
		in.Provider = "javapay"
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
		err := r.DB.QueryRowContext(ctx, q,
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
		err := r.DB.QueryRowContext(ctx, q,
			in.Provider, in.TransaksiMemberID, in.RefID, in.Perintah, in.ProdukSKUSnapshot, in.ProdukProviderMapID, in.KodeProduk, in.Tujuan, in.Qty, string(b),
		).Scan(&id, &refID, &dibuat)
		if err != nil {
			return nil, err
		}
	}

	return &model.JavapayTrxRow{
		ID:                  id,
		TransaksiMemberID:   in.TransaksiMemberID,
		Provider:            in.Provider,
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

func (r *JavapayRepo) UpdateRequestMentah(ctx context.Context, id int64, req any) error {
	b, _ := json.Marshal(req)
	_, err := r.DB.ExecContext(ctx,
		`UPDATE public.transaksi_provider SET request_mentah = $2::jsonb WHERE id = $1`,
		id, string(b),
	)
	if err != nil {
		return fmt.Errorf("update public.transaksi_provider request_mentah: %w", err)
	}
	return nil
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

func (r *JavapayRepo) UpdateResult(ctx context.Context, id int64, u UpdateResult) error {
	respJSON, _ := json.Marshal(u.ResponMentah)
	status := helper.ProviderResponseStatusString(r.getProviderByID(ctx, id), u.KodeRespon, u.Pesan)

	const q = `
UPDATE public.transaksi_provider SET
  trx_id_javapay = COALESCE($2, trx_id_javapay),
  kode_respon    = COALESCE($3, kode_respon),
  pesan          = COALESCE($4, pesan),
  no_referensi   = COALESCE($5, no_referensi),
  harga          = COALESCE($6, harga),
  saldo_terakhir = COALESCE($7, saldo_terakhir),
  http_status    = COALESCE($8, http_status),
  respon_mentah  = $9::jsonb,
  status         = COALESCE(NULLIF($10,''), status)
WHERE id = $1`

	var httpStatus any
	if u.HTTPStatus != nil {
		httpStatus = *u.HTTPStatus
	} else {
		httpStatus = nil
	}

	_, err := r.DB.ExecContext(ctx, q,
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
	)
	if err != nil {
		return fmt.Errorf("update public.transaksi_provider: %w", err)
	}
	return nil
}

func (r *JavapayRepo) GetLatestByRefIDProvider(ctx context.Context, refID, provider string) (*model.JavapayTrxRow, error) {
	rows, err := r.DB.QueryContext(ctx, `
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

type ProviderAnomasiIn struct {
	Provider   string
	RefID      string
	KodeRespon *string
	Pesan      *string
	Harga      *int64
	Tujuan     *string
	Qty        *int64
	RawQuery   string
	RawBody    string
	Payload    any
}

func (r *JavapayRepo) InsertTransaksiAnomasiProvider(ctx context.Context, in ProviderAnomasiIn) error {
	in.Provider = strings.TrimSpace(strings.ToLower(in.Provider))
	if in.Provider == "" {
		in.Provider = "unknown"
	}
	in.RefID = strings.TrimSpace(in.RefID)

	payloadJSON, _ := json.Marshal(in.Payload)
	sum := sha256.Sum256([]byte(strings.Join([]string{
		in.Provider,
		in.RefID,
		in.RawQuery,
		in.RawBody,
		string(payloadJSON),
	}, "|")))
	payloadHash := hex.EncodeToString(sum[:])

	isDuplicate := false
	isSuspectedFraud := false
	fraudReason := ""

	if in.RefID != "" {
		var (
			prevHash   sql.NullString
			prevHarga  sql.NullInt64
			prevTujuan sql.NullString
			prevQty    sql.NullInt64
		)
		err := r.DB.QueryRowContext(ctx, `
SELECT payload_hash, harga, tujuan, qty
FROM public.transaksi_anomasi_provider
WHERE provider = $1 AND ref_id = $2
ORDER BY id DESC
LIMIT 1
`, in.Provider, in.RefID).Scan(&prevHash, &prevHarga, &prevTujuan, &prevQty)
		if err == nil {
			if prevHash.Valid && prevHash.String == payloadHash {
				isDuplicate = true
			}

			var reasons []string
			if in.Harga != nil && prevHarga.Valid && prevHarga.Int64 != *in.Harga {
				isSuspectedFraud = true
				reasons = append(reasons, "harga_berbeda")
			}
			if in.Tujuan != nil && prevTujuan.Valid && strings.TrimSpace(prevTujuan.String) != strings.TrimSpace(*in.Tujuan) {
				isSuspectedFraud = true
				reasons = append(reasons, "tujuan_berbeda")
			}
			if in.Qty != nil && prevQty.Valid && prevQty.Int64 != *in.Qty {
				isSuspectedFraud = true
				reasons = append(reasons, "qty_berbeda")
			}
			fraudReason = strings.Join(reasons, ",")
		}
	}

	_, err := r.DB.ExecContext(ctx, `
INSERT INTO public.transaksi_anomasi_provider
  (provider, ref_id, kode_respon, pesan, harga, tujuan, qty, raw_query, raw_body, payload, payload_hash, is_duplicate, is_suspected_fraud, fraud_reason, dibuat_pada)
VALUES
  ($1, NULLIF($2,''), $3, $4, $5, NULLIF($6,''), $7, NULLIF($8,''), NULLIF($9,''), $10::jsonb, $11, $12, $13, NULLIF($14,''), now())
`, in.Provider, in.RefID, in.KodeRespon, in.Pesan, in.Harga, in.Tujuan, in.Qty, in.RawQuery, in.RawBody, string(payloadJSON), payloadHash, isDuplicate, isSuspectedFraud, fraudReason)
	return err
}
