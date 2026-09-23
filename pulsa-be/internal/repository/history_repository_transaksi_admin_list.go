package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func (r *HistoryRepository) AdminListTransaksiByMember(ctx context.Context, memberID int64, limit, offset int, search, status, kodeProduk, fromStr, toStr string) ([]AdminTrxRow, error) {
	if limit <= 0 {
		limit = 50
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
	wheres = append(wheres, fmt.Sprintf("tm.member_id = $%d", len(args)))

	search = strings.TrimSpace(search)
	if search != "" {
		args = append(args, search)
		p := fmt.Sprintf("$%d", len(args))
		wheres = append(wheres, fmt.Sprintf(`(
				tm.ref_id ILIKE '%%' || %s || '%%'
				OR tm.tujuan ILIKE '%%' || %s || '%%'
				OR tm.kode_produk ILIKE '%%' || %s || '%%'
				OR tm.perintah ILIKE '%%' || %s || '%%'
				OR tm.status ILIKE '%%' || %s || '%%'
				OR COALESCE(tm.keterangan,'') ILIKE '%%' || %s || '%%'
			)`, p, p, p, p, p, p))
	}

	status = strings.ToLower(strings.TrimSpace(status))
	if status != "" {
		args = append(args, status)
		wheres = append(wheres, fmt.Sprintf("LOWER(tm.status) = $%d", len(args)))
	}

	kodeProduk = strings.TrimSpace(kodeProduk)
	if kodeProduk != "" {
		kodes := expandKodeProdukFilters(kodeProduk)
		holders := make([]string, 0, len(kodes))
		for _, k := range kodes {
			args = append(args, k)
			holders = append(holders, fmt.Sprintf("$%d", len(args)))
		}
		wheres = append(wheres, fmt.Sprintf("UPPER(tm.kode_produk) IN (%s)", strings.Join(holders, ",")))
	}

	fromStr = strings.TrimSpace(fromStr)
	if fromStr != "" {
		t, err := time.ParseInLocation("2006-01-02", fromStr, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("tm.dibuat_pada >= $%d", len(args)))
	}

	toStr = strings.TrimSpace(toStr)
	if toStr != "" {
		t, err := time.ParseInLocation("2006-01-02", toStr, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t.AddDate(0, 0, 1))
		wheres = append(wheres, fmt.Sprintf("tm.dibuat_pada < $%d", len(args)))
	}

	args = append(args, limit)
	limitPos := len(args)
	args = append(args, offset)
	offsetPos := len(args)

	q := fmt.Sprintf(`
	SELECT tm.id, tm.member_id, m.email AS member_email, tm.ref_id, tm.perintah, tm.kode_produk, tm.tujuan, tm.qty, tm.status, tm.keterangan,
	       EXISTS (
	         SELECT 1
	         FROM public.member_ip_whitelist miw
	         WHERE miw.member_id = tm.member_id
           AND miw.aktif = true
           AND miw.webhook_url IS NOT NULL
           AND TRIM(miw.webhook_url) <> ''
	       ) AS has_callback_url,
	       tm.biaya_perkiraan, tm.biaya_aktual, tm.dibuat_pada, tm.diperbarui_pada
	FROM transaksi_member tm
	LEFT JOIN public.member m ON m.id = tm.member_id
	WHERE %s
	ORDER BY tm.dibuat_pada DESC, tm.id DESC
	LIMIT $%d OFFSET $%d
`, strings.Join(wheres, " AND "), limitPos, offsetPos)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]AdminTrxRow, 0, 64)
	for rows.Next() {
		var t AdminTrxRow
		var ket sql.NullString
		var memberEmail sql.NullString
		if err := rows.Scan(
			&t.ID, &t.MemberID, &memberEmail, &t.RefID, &t.Perintah, &t.KodeProduk, &t.Tujuan, &t.Qty,
			&t.Status, &ket, &t.HasCallbackURL, &t.BiayaPerkiraan, &t.BiayaAktual, &t.DibuatPada, &t.DiperbaruiPada,
		); err != nil {
			return nil, err
		}
		if memberEmail.Valid {
			t.MemberEmail = &memberEmail.String
		}
		if ket.Valid {
			t.Keterangan = &ket.String
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *HistoryRepository) AdminListTransaksiAll(ctx context.Context, limit, offset int, search, status, kodeProduk, refID, dest, fromStr, toStr, memberRole string) ([]AdminTrxRow, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")

	var (
		args   []any
		wheres []string
	)

	wheres = append(wheres, "1=1")
	if role := strings.ToLower(strings.TrimSpace(memberRole)); role != "" {
		args = append(args, role)
		wheres = append(wheres, fmt.Sprintf("LOWER(COALESCE(m.role, '')) = $%d", len(args)))
	}

	search = strings.TrimSpace(search)
	if search != "" {
		args = append(args, search)
		p := fmt.Sprintf("$%d", len(args))
		wheres = append(wheres, fmt.Sprintf(`(
				tm.ref_id ILIKE '%%' || %s || '%%'
				OR tm.tujuan ILIKE '%%' || %s || '%%'
				OR tm.kode_produk ILIKE '%%' || %s || '%%'
				OR tm.perintah ILIKE '%%' || %s || '%%'
				OR tm.status ILIKE '%%' || %s || '%%'
				OR COALESCE(tm.keterangan,'') ILIKE '%%' || %s || '%%'
			)`, p, p, p, p, p, p))
	}

	status = strings.ToLower(strings.TrimSpace(status))
	if status != "" {
		args = append(args, status)
		wheres = append(wheres, fmt.Sprintf("LOWER(tm.status) = $%d", len(args)))
	}

	kodeProduk = strings.TrimSpace(kodeProduk)
	if kodeProduk != "" {
		kodes := expandKodeProdukFilters(kodeProduk)
		holders := make([]string, 0, len(kodes))
		for _, k := range kodes {
			args = append(args, k)
			holders = append(holders, fmt.Sprintf("$%d", len(args)))
		}
		wheres = append(wheres, fmt.Sprintf("UPPER(tm.kode_produk) IN (%s)", strings.Join(holders, ",")))
	}

	refID = strings.TrimSpace(refID)
	if refID != "" {
		args = append(args, refID)
		p := fmt.Sprintf("$%d", len(args))
		wheres = append(wheres, fmt.Sprintf("tm.ref_id ILIKE '%%' || %s || '%%'", p))
	}

	dest = strings.TrimSpace(dest)
	if dest != "" {
		args = append(args, dest)
		p := fmt.Sprintf("$%d", len(args))
		wheres = append(wheres, fmt.Sprintf("tm.tujuan ILIKE '%%' || %s || '%%'", p))
	}

	fromStr = strings.TrimSpace(fromStr)
	if fromStr != "" {
		t, err := time.ParseInLocation("2006-01-02", fromStr, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("tm.dibuat_pada >= $%d", len(args)))
	}

	toStr = strings.TrimSpace(toStr)
	if toStr != "" {
		t, err := time.ParseInLocation("2006-01-02", toStr, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t.AddDate(0, 0, 1))
		wheres = append(wheres, fmt.Sprintf("tm.dibuat_pada < $%d", len(args)))
	}

	args = append(args, limit)
	limitPos := len(args)
	args = append(args, offset)
	offsetPos := len(args)

	q := fmt.Sprintf(`
	SELECT tm.id, tm.member_id, m.email AS member_email, tm.ref_id, tm.perintah, tm.kode_produk, tm.tujuan, tm.qty, tm.status, tm.keterangan,
	       EXISTS (
	         SELECT 1
	         FROM public.member_ip_whitelist miw
	         WHERE miw.member_id = tm.member_id
           AND miw.aktif = true
           AND miw.webhook_url IS NOT NULL
           AND TRIM(miw.webhook_url) <> ''
	       ) AS has_callback_url,
	       tm.biaya_perkiraan, tm.biaya_aktual, tm.dibuat_pada, tm.diperbarui_pada
	FROM transaksi_member tm
	LEFT JOIN public.member m ON m.id = tm.member_id
	WHERE %s
	ORDER BY tm.dibuat_pada DESC, tm.id DESC
	LIMIT $%d OFFSET $%d
`, strings.Join(wheres, " AND "), limitPos, offsetPos)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]AdminTrxRow, 0, 64)
	for rows.Next() {
		var t AdminTrxRow
		var ket sql.NullString
		var memberEmail sql.NullString
		if err := rows.Scan(
			&t.ID, &t.MemberID, &memberEmail, &t.RefID, &t.Perintah, &t.KodeProduk, &t.Tujuan, &t.Qty,
			&t.Status, &ket, &t.HasCallbackURL, &t.BiayaPerkiraan, &t.BiayaAktual, &t.DibuatPada, &t.DiperbaruiPada,
		); err != nil {
			return nil, err
		}
		if memberEmail.Valid {
			t.MemberEmail = &memberEmail.String
		}
		if ket.Valid {
			t.Keterangan = &ket.String
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
