package repository

import (
	"context"
	"database/sql"
	"strings"
)

func (r *AppOrderRepository) GetByID(ctx context.Context, id int64) (*AppOrderRow, error) {
	var (
		row         AppOrderRow
		memberID    sql.NullInt64
		memberNama  sql.NullString
		guestNama   sql.NullString
		guestEmail  sql.NullString
		guestPhone  sql.NullString
		catatan     sql.NullString
		alasanGagal sql.NullString
		sn          sql.NullString
		dibuatPada  sql.NullTime
		diubahPada  sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
SELECT
  ao.id, ao.invoice_id, ao.member_id, m.nama, ao.guest_nama, ao.guest_email, ao.guest_phone,
  ao.produk_id, ao.produk_sku_snapshot, ao.produk_nama_snapshot,
  ao.dest, ao.qty, ao.nominal, ao.buyer_type, ao.harga_dasar, ao.fee, ao.harga_final,
  ao.status, ao.catatan, ao.alasan_gagal, apt.sn, ao.dibuat_pada, ao.diubah_pada
FROM public.app_order ao
LEFT JOIN public.member m ON m.id = ao.member_id
LEFT JOIN LATERAL (
  SELECT sn
  FROM public.app_order_provider_trx
  WHERE app_order_id = ao.id
  ORDER BY diubah_pada DESC, dibuat_pada DESC, id DESC
  LIMIT 1
) apt ON true
WHERE ao.id = $1
LIMIT 1
`, id).Scan(
		&row.ID, &row.InvoiceID, &memberID, &memberNama, &guestNama, &guestEmail, &guestPhone,
		&row.ProdukID, &row.ProdukSKUSnapshot, &row.ProdukNamaSnapshot,
		&row.Dest, &row.Qty, &row.Nominal, &row.BuyerType, &row.HargaDasar, &row.Fee, &row.HargaFinal,
		&row.Status, &catatan, &alasanGagal, &sn, &dibuatPada, &diubahPada,
	)
	if err != nil {
		return nil, err
	}
	if memberID.Valid {
		v := memberID.Int64
		row.MemberID = &v
	}
	if memberNama.Valid {
		v := memberNama.String
		row.MemberNama = &v
	}
	if guestNama.Valid {
		v := guestNama.String
		row.GuestNama = &v
	}
	if guestEmail.Valid {
		v := guestEmail.String
		row.GuestEmail = &v
	}
	if guestPhone.Valid {
		v := guestPhone.String
		row.GuestPhone = &v
	}
	if catatan.Valid {
		v := catatan.String
		row.Catatan = &v
	}
	if alasanGagal.Valid {
		v := alasanGagal.String
		row.AlasanGagal = &v
	}
	if sn.Valid {
		v := sn.String
		row.SN = &v
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

func (r *AppOrderRepository) GetByInvoiceID(ctx context.Context, invoiceID string) (*AppOrderRow, error) {
	var (
		row         AppOrderRow
		memberID    sql.NullInt64
		memberNama  sql.NullString
		guestNama   sql.NullString
		guestEmail  sql.NullString
		guestPhone  sql.NullString
		catatan     sql.NullString
		alasanGagal sql.NullString
		sn          sql.NullString
		dibuatPada  sql.NullTime
		diubahPada  sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
SELECT
  ao.id, ao.invoice_id, ao.member_id, m.nama, ao.guest_nama, ao.guest_email, ao.guest_phone,
  ao.produk_id, ao.produk_sku_snapshot, ao.produk_nama_snapshot,
  ao.dest, ao.qty, ao.nominal, ao.buyer_type, ao.harga_dasar, ao.fee, ao.harga_final,
  ao.status, ao.catatan, ao.alasan_gagal, apt.sn, ao.dibuat_pada, ao.diubah_pada
FROM public.app_order ao
LEFT JOIN public.member m ON m.id = ao.member_id
LEFT JOIN LATERAL (
  SELECT sn
  FROM public.app_order_provider_trx
  WHERE app_order_id = ao.id
  ORDER BY diubah_pada DESC, dibuat_pada DESC, id DESC
  LIMIT 1
) apt ON true
WHERE ao.invoice_id = $1
LIMIT 1
`, strings.TrimSpace(invoiceID)).Scan(
		&row.ID, &row.InvoiceID, &memberID, &memberNama, &guestNama, &guestEmail, &guestPhone,
		&row.ProdukID, &row.ProdukSKUSnapshot, &row.ProdukNamaSnapshot,
		&row.Dest, &row.Qty, &row.Nominal, &row.BuyerType, &row.HargaDasar, &row.Fee, &row.HargaFinal,
		&row.Status, &catatan, &alasanGagal, &sn, &dibuatPada, &diubahPada,
	)
	if err != nil {
		return nil, err
	}
	if memberID.Valid {
		v := memberID.Int64
		row.MemberID = &v
	}
	if memberNama.Valid {
		v := memberNama.String
		row.MemberNama = &v
	}
	if guestNama.Valid {
		v := guestNama.String
		row.GuestNama = &v
	}
	if guestEmail.Valid {
		v := guestEmail.String
		row.GuestEmail = &v
	}
	if guestPhone.Valid {
		v := guestPhone.String
		row.GuestPhone = &v
	}
	if catatan.Valid {
		v := catatan.String
		row.Catatan = &v
	}
	if alasanGagal.Valid {
		v := alasanGagal.String
		row.AlasanGagal = &v
	}
	if sn.Valid {
		v := sn.String
		row.SN = &v
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
