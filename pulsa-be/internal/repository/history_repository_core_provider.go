package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"pulsa2/model"
)

func (r *HistoryRepository) GetLatestProviderByRefID(ctx context.Context, refID string) (*model.JavapayTrxRow, error) {
	refID = strings.TrimSpace(refID)
	if refID == "" {
		return nil, errors.New("ref_id invalid")
	}
	var row model.JavapayTrxRow
	var (
		trxIDJavapay  sql.NullString
		kodeRespon    sql.NullString
		pesan         sql.NullString
		noReferensi   sql.NullString
		harga         sql.NullInt64
		saldoTerakhir sql.NullInt64
		httpStatus    sql.NullInt32
	)
	err := r.db.QueryRowContext(ctx, `
SELECT id, transaksi_member_id, provider, ref_id, perintah, produk_sku_snapshot, produk_provider_map_id,
       kode_produk, tujuan, qty, trx_id_javapay, kode_respon, pesan, no_referensi,
       harga, saldo_terakhir, http_status, percobaan, dibuat_pada
FROM public.transaksi_provider
WHERE ref_id = $1
ORDER BY id DESC
LIMIT 1
`, refID).Scan(
		&row.ID,
		&row.TransaksiMemberID,
		&row.Provider,
		&row.RefID,
		&row.Perintah,
		&row.ProdukSKUSnapshot,
		&row.ProdukProviderMapID,
		&row.KodeProduk,
		&row.Tujuan,
		&row.Qty,
		&trxIDJavapay,
		&kodeRespon,
		&pesan,
		&noReferensi,
		&harga,
		&saldoTerakhir,
		&httpStatus,
		&row.Percobaan,
		&row.DibuatPada,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if trxIDJavapay.Valid {
		row.TrxIDJavapay = &trxIDJavapay.String
	}
	if kodeRespon.Valid {
		row.KodeRespon = &kodeRespon.String
	}
	if pesan.Valid {
		row.Pesan = &pesan.String
	}
	if noReferensi.Valid {
		row.NoReferensi = &noReferensi.String
	}
	if harga.Valid {
		v := harga.Int64
		row.Harga = &v
	}
	if saldoTerakhir.Valid {
		v := saldoTerakhir.Int64
		row.SaldoTerakhir = &v
	}
	if httpStatus.Valid {
		v := int(httpStatus.Int32)
		row.HTTPStatus = &v
	}
	return &row, nil
}

func (r *HistoryRepository) GetSaldo(ctx context.Context, memberID int64) (int64, error) {
	var saldo int64
	err := r.db.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_member
WHERE member_id=$1
`, memberID).Scan(&saldo)
	return saldo, err
}
