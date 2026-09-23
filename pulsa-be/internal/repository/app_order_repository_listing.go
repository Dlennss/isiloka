package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func (r *AppOrderRepository) ListByMemberID(ctx context.Context, f AppOrderListFilter) ([]AppOrderRow, error) {
	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	args := []any{f.MemberID}
	where := []string{"ao.member_id = $1", "ao.buyer_type = 'user'"}

	if status := strings.TrimSpace(strings.ToLower(f.Status)); status != "" {
		args = append(args, status)
		where = append(where, fmt.Sprintf("lower(ao.status) = $%d", len(args)))
	}
	if f.DateFrom != nil {
		args = append(args, *f.DateFrom)
		where = append(where, fmt.Sprintf("ao.dibuat_pada >= $%d", len(args)))
	}
	if f.DateTo != nil {
		args = append(args, *f.DateTo)
		where = append(where, fmt.Sprintf("ao.dibuat_pada < $%d", len(args)))
	}

	args = append(args, limit, offset)
	q := fmt.Sprintf(`
SELECT
  ao.id, ao.invoice_id, ao.member_id, m.nama, ao.guest_nama, ao.guest_email, ao.guest_phone,
  ao.produk_id, ao.produk_sku_snapshot, ao.produk_nama_snapshot,
  ao.dest, ao.qty, ao.nominal, ao.buyer_type, ao.harga_dasar, ao.fee, ao.harga_final,
  ao.status, ao.catatan, ao.alasan_gagal, ao.dibuat_pada, ao.diubah_pada
FROM public.app_order ao
LEFT JOIN public.member m ON m.id = ao.member_id
WHERE %s
ORDER BY ao.dibuat_pada DESC, ao.id DESC
LIMIT $%d OFFSET $%d
`, strings.Join(where, " AND "), len(args)-1, len(args))

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAppOrderRows(rows)
}

func (r *AppOrderRepository) List(ctx context.Context, f AppOrderListFilter) ([]AppOrderRow, error) {
	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	args := make([]any, 0, 8)
	where := make([]string, 0, 8)

	if !f.All && f.MemberID > 0 {
		args = append(args, f.MemberID)
		where = append(where, fmt.Sprintf("ao.member_id = $%d", len(args)))
		where = append(where, "ao.buyer_type = 'user'")
	}
	if buyerType := strings.TrimSpace(strings.ToLower(f.BuyerType)); buyerType != "" {
		args = append(args, buyerType)
		where = append(where, fmt.Sprintf("lower(ao.buyer_type) = $%d", len(args)))
	}
	if status := strings.TrimSpace(strings.ToLower(f.Status)); status != "" {
		args = append(args, status)
		where = append(where, fmt.Sprintf("lower(ao.status) = $%d", len(args)))
	}
	if q := strings.TrimSpace(strings.ToLower(f.Q)); q != "" {
		args = append(args, "%"+q+"%")
		where = append(where, fmt.Sprintf("(lower(ao.invoice_id) LIKE $%d OR lower(ao.dest) LIKE $%d OR lower(ao.produk_nama_snapshot) LIKE $%d)", len(args), len(args), len(args)))
	}
	if f.DateFrom != nil {
		args = append(args, *f.DateFrom)
		where = append(where, fmt.Sprintf("ao.dibuat_pada >= $%d", len(args)))
	}
	if f.DateTo != nil {
		args = append(args, *f.DateTo)
		where = append(where, fmt.Sprintf("ao.dibuat_pada < $%d", len(args)))
	}

	q := `
SELECT
  ao.id, ao.invoice_id, ao.member_id, m.nama, ao.guest_nama, ao.guest_email, ao.guest_phone,
  ao.produk_id, ao.produk_sku_snapshot, ao.produk_nama_snapshot,
  ao.dest, ao.qty, ao.nominal, ao.buyer_type, ao.harga_dasar, ao.fee, ao.harga_final,
  ao.status, ao.catatan, ao.alasan_gagal, ao.dibuat_pada, ao.diubah_pada
FROM public.app_order ao
LEFT JOIN public.member m ON m.id = ao.member_id`
	if len(where) > 0 {
		q += "\nWHERE " + strings.Join(where, " AND ")
	}
	args = append(args, limit, offset)
	q += fmt.Sprintf("\nORDER BY ao.dibuat_pada DESC, ao.id DESC\nLIMIT $%d OFFSET $%d", len(args)-1, len(args))

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAppOrderRows(rows)
}

func scanAppOrderRows(rows *sql.Rows) ([]AppOrderRow, error) {
	var out []AppOrderRow
	for rows.Next() {
		var (
			row         AppOrderRow
			memberID    sql.NullInt64
			memberNama  sql.NullString
			guestNama   sql.NullString
			guestEmail  sql.NullString
			guestPhone  sql.NullString
			catatan     sql.NullString
			alasanGagal sql.NullString
			dibuatPada  sql.NullTime
			diubahPada  sql.NullTime
		)
		if err := rows.Scan(
			&row.ID, &row.InvoiceID, &memberID, &memberNama, &guestNama, &guestEmail, &guestPhone,
			&row.ProdukID, &row.ProdukSKUSnapshot, &row.ProdukNamaSnapshot,
			&row.Dest, &row.Qty, &row.Nominal, &row.BuyerType, &row.HargaDasar, &row.Fee, &row.HargaFinal,
			&row.Status, &catatan, &alasanGagal, &dibuatPada, &diubahPada,
		); err != nil {
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
		if dibuatPada.Valid {
			v := dibuatPada.Time
			row.DibuatPada = &v
		}
		if diubahPada.Valid {
			v := diubahPada.Time
			row.DiubahPada = &v
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
