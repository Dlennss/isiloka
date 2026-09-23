package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
)

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

func (r *ProviderCallbackRepository) InsertAnomali(ctx context.Context, in ProviderAnomasiIn) error {
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
		err := r.db.QueryRowContext(ctx, `
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

	_, err := r.db.ExecContext(ctx, `
INSERT INTO public.transaksi_anomasi_provider
  (provider, ref_id, kode_respon, pesan, harga, tujuan, qty, raw_query, raw_body, payload, payload_hash, is_duplicate, is_suspected_fraud, fraud_reason, dibuat_pada)
VALUES
  ($1, NULLIF($2,''), $3, $4, $5, NULLIF($6,''), $7, NULLIF($8,''), NULLIF($9,''), $10::jsonb, $11, $12, $13, NULLIF($14,''), now())
`, in.Provider, in.RefID, in.KodeRespon, in.Pesan, in.Harga, in.Tujuan, in.Qty, in.RawQuery, in.RawBody, string(payloadJSON), payloadHash, isDuplicate, isSuspectedFraud, fraudReason)
	return err
}

type ProviderTrxRefRow struct {
	ID                  int64
	TransaksiMemberID   int64
	ProdukSKUSnapshot   string
	ProdukProviderMapID *int64
	KodeProduk          string
	RequestMode         *string
	KodeRespon          *string
	Pesan               *string
	NoReferensi         *string
	Harga               *int64
	Status              string
	DibuatPada          time.Time
}

type ProviderAttemptRow struct {
	ID                  int64
	RefID               string
	Provider            string
	Perintah            string
	ProdukSKUSnapshot   string
	ProdukProviderMapID *int64
	KodeProduk          string
	KodeRespon          *string
	Pesan               *string
	NoReferensi         *string
	Harga               *int64
	Status              string
	HTTPStatus          int
	DibuatPada          time.Time
}
