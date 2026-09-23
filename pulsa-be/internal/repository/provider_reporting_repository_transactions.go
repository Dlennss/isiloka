package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type ProviderTransactionListArgs struct {
	Provider   string
	Q          string
	MemberID   int64
	KodeProduk string
	RefID      string
	Tujuan     string
	KodeRespon string
	Status     string

	From    time.Time
	HasFrom bool

	To    time.Time
	HasTo bool

	Limit  int
	Offset int
}

type ProviderTransactionRecord struct {
	ID                int64     `json:"id"`
	DibuatPada        time.Time `json:"dibuat_pada"`
	Provider          string    `json:"provider"`
	StatusProvider    string    `json:"status_provider"`
	TransaksiMemberID int64     `json:"transaksi_member_id"`
	MemberID          *int64    `json:"member_id,omitempty"`
	MemberNama        *string   `json:"member_nama,omitempty"`
	RefID             string    `json:"ref_id"`
	Perintah          string    `json:"perintah"`
	KodeProduk        string    `json:"kode_produk"`
	Tujuan            string    `json:"tujuan"`
	Qty               int64     `json:"qty"`
	TrxIDJavapay      *string   `json:"trx_id_javapay,omitempty"`
	TrxIDProvider     *string   `json:"trx_id_provider,omitempty"`
	KodeRespon        *string   `json:"kode_respon,omitempty"`
	Pesan             *string   `json:"pesan,omitempty"`
	NoReferensi       *string   `json:"no_referensi,omitempty"`
	Harga             *int64    `json:"harga,omitempty"`
	SaldoTerakhir     *int64    `json:"saldo_terakhir,omitempty"`
	HTTPStatus        *int      `json:"http_status,omitempty"`
	Percobaan         int       `json:"percobaan"`
}

func (r *ProviderReportingRepository) ListTransactions(ctx context.Context, a ProviderTransactionListArgs) ([]ProviderTransactionRecord, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, errors.New("db not initialized")
	}

	if a.Limit <= 0 || a.Limit > 500 {
		a.Limit = 100
	}
	if a.Offset < 0 {
		a.Offset = 0
	}

	a.Provider = strings.TrimSpace(a.Provider)
	a.Q = strings.TrimSpace(a.Q)
	a.KodeProduk = strings.TrimSpace(a.KodeProduk)
	a.RefID = strings.TrimSpace(a.RefID)
	a.Tujuan = strings.TrimSpace(a.Tujuan)
	a.KodeRespon = strings.TrimSpace(a.KodeRespon)
	a.Status = strings.ToLower(strings.TrimSpace(a.Status))

	statusExpr := `lower(trim(coalesce(t.status, 'unknown')))`

	qList := `
SELECT
  t.id,
  t.dibuat_pada,
  t.provider,
  ` + statusExpr + ` AS status_provider,
  t.transaksi_member_id,
  tm.member_id AS member_id,
  m.nama      AS member_nama,
  t.ref_id,
  t.perintah,
  t.kode_produk,
  t.tujuan,
  t.qty,
  t.trx_id_javapay,
  t.kode_respon,
  t.pesan,
  t.no_referensi,
  t.harga,
  t.saldo_terakhir,
  t.http_status,
  t.percobaan
FROM public.transaksi_provider t
LEFT JOIN public.transaksi_member tm ON tm.id = t.transaksi_member_id
LEFT JOIN public.member m            ON m.id  = tm.member_id
WHERE ($1 = '' OR t.provider = $1)
  AND ($2 = false OR t.dibuat_pada >= $3)
  AND ($4 = false OR t.dibuat_pada <  $5)
  AND (
    $6 = '' OR t.ref_id ILIKE '%'||$6||'%' OR t.tujuan ILIKE '%'||$6||'%' OR t.kode_produk ILIKE '%'||$6||'%'
  )
  AND ($9 = 0 OR tm.member_id = $9)
  AND ($10 = '' OR t.kode_produk ILIKE '%'||$10||'%')
  AND ($11 = '' OR t.ref_id ILIKE '%'||$11||'%')
  AND ($12 = '' OR t.tujuan ILIKE '%'||$12||'%')
  AND ($13 = '' OR t.kode_respon = $13)
  AND ($14 = '' OR ` + statusExpr + ` = $14)
ORDER BY t.dibuat_pada DESC, t.ref_id DESC, t.id DESC
LIMIT $7 OFFSET $8
`
	rows, err := r.db.QueryContext(ctx, qList,
		a.Provider,
		a.HasFrom, a.From,
		a.HasTo, a.To,
		a.Q,
		a.Limit, a.Offset,
		a.MemberID,
		a.KodeProduk,
		a.RefID,
		a.Tujuan,
		a.KodeRespon,
		a.Status,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]ProviderTransactionRecord, 0, a.Limit)
	for rows.Next() {
		var (
			x ProviderTransactionRecord

			memberID   sql.NullInt64
			memberNama sql.NullString

			trxIDJavapay sql.NullString
			kodeRespon   sql.NullString
			pesan        sql.NullString
			noReferensi  sql.NullString

			harga         sql.NullInt64
			saldoTerakhir sql.NullInt64
			httpStatus    sql.NullInt64
			statusValue   sql.NullString
		)

		err := rows.Scan(
			&x.ID,
			&x.DibuatPada,
			&x.Provider,
			&statusValue,
			&x.TransaksiMemberID,
			&memberID,
			&memberNama,
			&x.RefID,
			&x.Perintah,
			&x.KodeProduk,
			&x.Tujuan,
			&x.Qty,
			&trxIDJavapay,
			&kodeRespon,
			&pesan,
			&noReferensi,
			&harga,
			&saldoTerakhir,
			&httpStatus,
			&x.Percobaan,
		)
		if err != nil {
			return nil, 0, err
		}
		x.KodeProduk = normalizeProviderProductCodeDisplay(x.Provider, x.KodeProduk)

		if memberID.Valid {
			v := memberID.Int64
			x.MemberID = &v
		}
		if memberNama.Valid {
			v := memberNama.String
			x.MemberNama = &v
		}
		if trxIDJavapay.Valid {
			v := trxIDJavapay.String
			x.TrxIDJavapay = &v
			x.TrxIDProvider = &v
		}
		if kodeRespon.Valid {
			v := kodeRespon.String
			x.KodeRespon = &v
		}
		if pesan.Valid {
			v := pesan.String
			x.Pesan = &v
		}
		if noReferensi.Valid {
			v := noReferensi.String
			x.NoReferensi = &v
		}
		if harga.Valid {
			v := harga.Int64
			x.Harga = &v
		}
		if saldoTerakhir.Valid {
			v := saldoTerakhir.Int64
			x.SaldoTerakhir = &v
		}
		if httpStatus.Valid {
			v := int(httpStatus.Int64)
			x.HTTPStatus = &v
		}
		if statusValue.Valid {
			x.StatusProvider = strings.TrimSpace(statusValue.String)
		}

		out = append(out, x)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	qCount := `
SELECT count(*)
FROM public.transaksi_provider t
LEFT JOIN public.transaksi_member tm ON tm.id = t.transaksi_member_id
WHERE ($1 = '' OR t.provider = $1)
  AND ($2 = false OR t.dibuat_pada >= $3)
  AND ($4 = false OR t.dibuat_pada <  $5)
  AND (
    $6 = '' OR t.ref_id ILIKE '%'||$6||'%' OR t.tujuan ILIKE '%'||$6||'%' OR t.kode_produk ILIKE '%'||$6||'%'
  )
  AND ($7 = 0 OR tm.member_id = $7)
  AND ($8 = '' OR t.kode_produk ILIKE '%'||$8||'%')
  AND ($9 = '' OR t.ref_id ILIKE '%'||$9||'%')
  AND ($10 = '' OR t.tujuan ILIKE '%'||$10||'%')
  AND ($11 = '' OR t.kode_respon = $11)
  AND ($12 = '' OR ` + statusExpr + ` = $12)
`
	var total int64
	err = r.db.QueryRowContext(ctx, qCount,
		a.Provider,
		a.HasFrom, a.From,
		a.HasTo, a.To,
		a.Q,
		a.MemberID,
		a.KodeProduk,
		a.RefID,
		a.Tujuan,
		a.KodeRespon,
		a.Status,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return out, total, nil
}
