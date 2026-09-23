package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type ProviderAnomalyListArgs struct {
	Provider   string
	Q          string
	RefID      string
	KodeRespon string
	Status     string

	From    time.Time
	HasFrom bool

	To    time.Time
	HasTo bool

	Limit  int
	Offset int
}

type ProviderAnomalyRecord struct {
	ID               int64     `json:"id"`
	DibuatPada       time.Time `json:"dibuat_pada"`
	Provider         string    `json:"provider"`
	RefID            *string   `json:"ref_id,omitempty"`
	KodeRespon       *string   `json:"kode_respon,omitempty"`
	Pesan            *string   `json:"pesan,omitempty"`
	Harga            *int64    `json:"harga,omitempty"`
	Tujuan           *string   `json:"tujuan,omitempty"`
	Qty              *int64    `json:"qty,omitempty"`
	PayloadHash      *string   `json:"payload_hash,omitempty"`
	IsDuplicate      bool      `json:"is_duplicate"`
	IsSuspectedFraud bool      `json:"is_suspected_fraud"`
	FraudReason      *string   `json:"fraud_reason,omitempty"`
	RawQuery         *string   `json:"raw_query,omitempty"`
	RawBody          *string   `json:"raw_body,omitempty"`
}

func (r *ProviderReportingRepository) ListAnomalies(ctx context.Context, a ProviderAnomalyListArgs) ([]ProviderAnomalyRecord, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, errors.New("db not initialized")
	}

	if a.Limit <= 0 || a.Limit > 100 {
		a.Limit = 100
	}
	if a.Offset < 0 {
		a.Offset = 0
	}

	a.Provider = strings.TrimSpace(strings.ToLower(a.Provider))
	a.Q = strings.TrimSpace(a.Q)
	a.RefID = strings.TrimSpace(a.RefID)
	a.KodeRespon = strings.TrimSpace(a.KodeRespon)
	a.Status = strings.TrimSpace(strings.ToLower(a.Status))

	const qList = `
SELECT
  a.id,
  a.dibuat_pada,
  a.provider,
  a.ref_id,
  a.kode_respon,
  a.pesan,
  a.harga,
  a.tujuan,
  a.qty,
  a.payload_hash,
  a.is_duplicate,
  a.is_suspected_fraud,
  a.fraud_reason,
  a.raw_query,
  a.raw_body
FROM public.transaksi_anomasi_provider a
WHERE ($1 = '' OR a.provider = $1)
  AND ($2 = false OR a.dibuat_pada >= $3)
  AND ($4 = false OR a.dibuat_pada <  $5)
  AND (
    $6 = '' OR
    coalesce(a.ref_id,'') ILIKE '%'||$6||'%' OR
    coalesce(a.tujuan,'') ILIKE '%'||$6||'%' OR
    coalesce(a.pesan,'') ILIKE '%'||$6||'%'
  )
  AND ($7 = '' OR coalesce(a.ref_id,'') ILIKE '%'||$7||'%')
  AND ($8 = '' OR coalesce(a.kode_respon,'') = $8)
  AND (
    $9 = '' OR
    ($9 = 'duplicate' AND a.is_duplicate = true) OR
    ($9 = 'fraud' AND a.is_suspected_fraud = true)
  )
ORDER BY a.dibuat_pada DESC, a.id DESC
LIMIT $10 OFFSET $11
`
	rows, err := r.db.QueryContext(ctx, qList,
		a.Provider,
		a.HasFrom, a.From,
		a.HasTo, a.To,
		a.Q,
		a.RefID,
		a.KodeRespon,
		a.Status,
		a.Limit, a.Offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]ProviderAnomalyRecord, 0, a.Limit)
	for rows.Next() {
		var (
			x ProviderAnomalyRecord

			refID       sql.NullString
			kodeRespon  sql.NullString
			pesan       sql.NullString
			harga       sql.NullInt64
			tujuan      sql.NullString
			qty         sql.NullInt64
			payloadHash sql.NullString
			fraudReason sql.NullString
			rawQuery    sql.NullString
			rawBody     sql.NullString
		)
		if err := rows.Scan(
			&x.ID,
			&x.DibuatPada,
			&x.Provider,
			&refID,
			&kodeRespon,
			&pesan,
			&harga,
			&tujuan,
			&qty,
			&payloadHash,
			&x.IsDuplicate,
			&x.IsSuspectedFraud,
			&fraudReason,
			&rawQuery,
			&rawBody,
		); err != nil {
			return nil, 0, err
		}

		if refID.Valid {
			v := refID.String
			x.RefID = &v
		}
		if kodeRespon.Valid {
			v := kodeRespon.String
			x.KodeRespon = &v
		}
		if pesan.Valid {
			v := pesan.String
			x.Pesan = &v
		}
		if harga.Valid {
			v := harga.Int64
			x.Harga = &v
		}
		if tujuan.Valid {
			v := tujuan.String
			x.Tujuan = &v
		}
		if qty.Valid {
			v := qty.Int64
			x.Qty = &v
		}
		if payloadHash.Valid {
			v := payloadHash.String
			x.PayloadHash = &v
		}
		if fraudReason.Valid {
			v := fraudReason.String
			x.FraudReason = &v
		}
		if rawQuery.Valid {
			v := rawQuery.String
			x.RawQuery = &v
		}
		if rawBody.Valid {
			v := rawBody.String
			x.RawBody = &v
		}

		out = append(out, x)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	const qCount = `
SELECT count(*)
FROM public.transaksi_anomasi_provider a
WHERE ($1 = '' OR a.provider = $1)
  AND ($2 = false OR a.dibuat_pada >= $3)
  AND ($4 = false OR a.dibuat_pada <  $5)
  AND (
    $6 = '' OR
    coalesce(a.ref_id,'') ILIKE '%'||$6||'%' OR
    coalesce(a.tujuan,'') ILIKE '%'||$6||'%' OR
    coalesce(a.pesan,'') ILIKE '%'||$6||'%'
  )
  AND ($7 = '' OR coalesce(a.ref_id,'') ILIKE '%'||$7||'%')
  AND ($8 = '' OR coalesce(a.kode_respon,'') = $8)
  AND (
    $9 = '' OR
    ($9 = 'duplicate' AND a.is_duplicate = true) OR
    ($9 = 'fraud' AND a.is_suspected_fraud = true)
  )
`
	var total int64
	if err := r.db.QueryRowContext(ctx, qCount,
		a.Provider,
		a.HasFrom, a.From,
		a.HasTo, a.To,
		a.Q,
		a.RefID,
		a.KodeRespon,
		a.Status,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	return out, total, nil
}
