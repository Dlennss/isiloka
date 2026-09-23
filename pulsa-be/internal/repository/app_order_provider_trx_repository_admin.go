package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func (r *AppOrderProviderTrxRepository) ListAdmin(ctx context.Context, f AppOrderProviderTrxListFilter) ([]AppOrderProviderTrxAdminRow, error) {
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

	if provider := strings.TrimSpace(strings.ToLower(f.Provider)); provider != "" {
		args = append(args, provider)
		where = append(where, fmt.Sprintf("LOWER(TRIM(apt.provider)) = $%d", len(args)))
	}
	if status := strings.TrimSpace(strings.ToLower(f.Status)); status != "" {
		args = append(args, status)
		where = append(where, fmt.Sprintf("LOWER(TRIM(apt.status)) = $%d", len(args)))
	}
	if refID := strings.TrimSpace(f.RefID); refID != "" {
		args = append(args, "%"+strings.ToLower(refID)+"%")
		where = append(where, fmt.Sprintf("LOWER(apt.ref_id) LIKE $%d", len(args)))
	}
	if invoiceID := strings.TrimSpace(f.InvoiceID); invoiceID != "" {
		args = append(args, "%"+strings.ToLower(invoiceID)+"%")
		where = append(where, fmt.Sprintf("LOWER(ao.invoice_id) LIKE $%d", len(args)))
	}
	if f.DateFrom != nil {
		args = append(args, *f.DateFrom)
		where = append(where, fmt.Sprintf("apt.dibuat_pada >= $%d", len(args)))
	}
	if f.DateTo != nil {
		args = append(args, *f.DateTo)
		where = append(where, fmt.Sprintf("apt.dibuat_pada < $%d", len(args)))
	}

	q := `
SELECT
  apt.id, apt.app_order_id, ao.invoice_id, apt.provider, apt.ref_id,
  COALESCE(apt.raw_request->>'product', ''), ao.produk_nama_snapshot, ao.dest, apt.harga_provider, apt.status,
  apt.kode_respon, apt.pesan, apt.sn, apt.dibuat_pada, apt.diubah_pada
FROM public.app_order_provider_trx apt
JOIN public.app_order ao ON ao.id = apt.app_order_id`
	if len(where) > 0 {
		q += "\nWHERE " + strings.Join(where, " AND ")
	}
	args = append(args, limit, offset)
	q += fmt.Sprintf("\nORDER BY apt.dibuat_pada DESC, apt.id DESC\nLIMIT $%d OFFSET $%d", len(args)-1, len(args))

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AppOrderProviderTrxAdminRow
	for rows.Next() {
		var (
			row        AppOrderProviderTrxAdminRow
			kodeRespon sql.NullString
			pesan      sql.NullString
			sn         sql.NullString
			dibuat     sql.NullTime
			diubah     sql.NullTime
		)
		if err := rows.Scan(
			&row.ID,
			&row.AppOrderID,
			&row.InvoiceID,
			&row.Provider,
			&row.RefID,
			&row.KodeProvider,
			&row.ProdukNamaSnapshot,
			&row.Dest,
			&row.HargaProvider,
			&row.Status,
			&kodeRespon,
			&pesan,
			&sn,
			&dibuat,
			&diubah,
		); err != nil {
			return nil, err
		}
		if kodeRespon.Valid {
			v := kodeRespon.String
			row.KodeRespon = &v
		}
		if pesan.Valid {
			v := pesan.String
			row.Pesan = &v
		}
		if sn.Valid {
			v := sn.String
			row.SN = &v
		}
		if dibuat.Valid {
			v := dibuat.Time
			row.DibuatPada = &v
		}
		if diubah.Valid {
			v := diubah.Time
			row.DiubahPada = &v
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
