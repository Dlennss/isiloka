package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func (r *HistoryRepository) ListTransaksi(ctx context.Context, memberID int64, limit, offset int, search, status, fromStr, toStr string) ([]TrxMemberRow, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")

	var (
		args   []any
		wheres []string
	)

	args = append(args, memberID)
	wheres = append(wheres, fmt.Sprintf("member_id=$%d", len(args)))

	search = strings.TrimSpace(search)
	if search != "" {
		args = append(args, search)
		p := fmt.Sprintf("$%d", len(args))
		wheres = append(wheres, fmt.Sprintf(`(
			ref_id ILIKE '%%' || %s || '%%'
			OR tujuan ILIKE '%%' || %s || '%%'
			OR kode_produk ILIKE '%%' || %s || '%%'
			OR perintah ILIKE '%%' || %s || '%%'
			OR status ILIKE '%%' || %s || '%%'
			OR COALESCE(keterangan,'') ILIKE '%%' || %s || '%%'
		)`, p, p, p, p, p, p))
	}

	status = strings.TrimSpace(strings.ToLower(status))
	if status != "" {
		args = append(args, status)
		wheres = append(wheres, fmt.Sprintf("LOWER(status) = $%d", len(args)))
	}

	if strings.TrimSpace(fromStr) != "" {
		t, err := time.ParseInLocation("2006-01-02", fromStr, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("dibuat_pada >= $%d", len(args)))
	}

	if strings.TrimSpace(toStr) != "" {
		t, err := time.ParseInLocation("2006-01-02", toStr, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
		}
		t = t.AddDate(0, 0, 1)
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("dibuat_pada < $%d", len(args)))
	}

	args = append(args, limit)
	limitPos := len(args)
	args = append(args, offset)
	offsetPos := len(args)

	sqlQ := fmt.Sprintf(`
SELECT id, ref_id, perintah, kode_produk, tujuan, qty, qty_provider, charge_receiver_applied, fee_member_rp, status, keterangan,
       biaya_perkiraan, biaya_aktual, dibuat_pada, diperbarui_pada
FROM public.transaksi_member
WHERE %s
ORDER BY dibuat_pada DESC, id DESC
LIMIT $%d OFFSET $%d
`, strings.Join(wheres, " AND "), limitPos, offsetPos)

	rows, err := r.db.QueryContext(ctx, sqlQ, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TrxMemberRow
	for rows.Next() {
		var t TrxMemberRow
		var ket sql.NullString
		if err := rows.Scan(
			&t.ID, &t.RefID, &t.Perintah, &t.KodeProduk, &t.Tujuan, &t.Qty, &t.QtyProvider, &t.ChargeReceiverApplied, &t.FeeMemberRp,
			&t.Status, &ket, &t.BiayaPerkiraan, &t.BiayaAktual, &t.DibuatPada, &t.DiperbaruiPada,
		); err != nil {
			return nil, err
		}
		if ket.Valid {
			t.Keterangan = &ket.String
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *HistoryRepository) CountTransaksi(ctx context.Context, memberID int64, search, status, fromStr, toStr string) (int64, error) {
	loc, _ := time.LoadLocation("Asia/Jakarta")

	var (
		args   []any
		wheres []string
	)

	args = append(args, memberID)
	wheres = append(wheres, fmt.Sprintf("member_id=$%d", len(args)))

	search = strings.TrimSpace(search)
	if search != "" {
		args = append(args, search)
		p := fmt.Sprintf("$%d", len(args))
		wheres = append(wheres, fmt.Sprintf(`(
			ref_id ILIKE '%%' || %s || '%%'
			OR tujuan ILIKE '%%' || %s || '%%'
			OR kode_produk ILIKE '%%' || %s || '%%'
			OR perintah ILIKE '%%' || %s || '%%'
			OR status ILIKE '%%' || %s || '%%'
			OR COALESCE(keterangan,'') ILIKE '%%' || %s || '%%'
		)`, p, p, p, p, p, p))
	}

	status = strings.TrimSpace(strings.ToLower(status))
	if status != "" {
		args = append(args, status)
		wheres = append(wheres, fmt.Sprintf("LOWER(status) = $%d", len(args)))
	}

	if strings.TrimSpace(fromStr) != "" {
		t, err := time.ParseInLocation("2006-01-02", fromStr, loc)
		if err != nil {
			return 0, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("dibuat_pada >= $%d", len(args)))
	}

	if strings.TrimSpace(toStr) != "" {
		t, err := time.ParseInLocation("2006-01-02", toStr, loc)
		if err != nil {
			return 0, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
		}
		t = t.AddDate(0, 0, 1)
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("dibuat_pada < $%d", len(args)))
	}

	sqlQ := fmt.Sprintf(`SELECT COUNT(1) FROM public.transaksi_member WHERE %s`, strings.Join(wheres, " AND "))

	var total int64
	if err := r.db.QueryRowContext(ctx, sqlQ, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}
