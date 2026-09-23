package repository

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func (r *HistoryRepository) AdminCountTransaksiByMember(ctx context.Context, memberID int64, search, status, kodeProduk, fromStr, toStr string) (int64, error) {
	loc, _ := time.LoadLocation("Asia/Jakarta")

	var (
		args   []any
		wheres []string
	)

	args = append(args, memberID)
	wheres = append(wheres, fmt.Sprintf("member_id = $%d", len(args)))

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

	status = strings.ToLower(strings.TrimSpace(status))
	if status != "" {
		args = append(args, status)
		wheres = append(wheres, fmt.Sprintf("LOWER(status) = $%d", len(args)))
	}

	kodeProduk = strings.TrimSpace(kodeProduk)
	if kodeProduk != "" {
		kodes := expandKodeProdukFilters(kodeProduk)
		holders := make([]string, 0, len(kodes))
		for _, k := range kodes {
			args = append(args, k)
			holders = append(holders, fmt.Sprintf("$%d", len(args)))
		}
		wheres = append(wheres, fmt.Sprintf("UPPER(kode_produk) IN (%s)", strings.Join(holders, ",")))
	}

	fromStr = strings.TrimSpace(fromStr)
	if fromStr != "" {
		t, err := time.ParseInLocation("2006-01-02", fromStr, loc)
		if err != nil {
			return 0, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("dibuat_pada >= $%d", len(args)))
	}

	toStr = strings.TrimSpace(toStr)
	if toStr != "" {
		t, err := time.ParseInLocation("2006-01-02", toStr, loc)
		if err != nil {
			return 0, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t.AddDate(0, 0, 1))
		wheres = append(wheres, fmt.Sprintf("dibuat_pada < $%d", len(args)))
	}

	q := fmt.Sprintf(`SELECT COUNT(1) FROM transaksi_member WHERE %s`, strings.Join(wheres, " AND "))
	var total int64
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *HistoryRepository) AdminCountTransaksiAll(ctx context.Context, search, status, kodeProduk, refID, dest, fromStr, toStr, memberRole string) (int64, error) {
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
			ref_id ILIKE '%%' || %s || '%%'
			OR tujuan ILIKE '%%' || %s || '%%'
			OR kode_produk ILIKE '%%' || %s || '%%'
			OR perintah ILIKE '%%' || %s || '%%'
			OR status ILIKE '%%' || %s || '%%'
			OR COALESCE(keterangan,'') ILIKE '%%' || %s || '%%'
		)`, p, p, p, p, p, p))
	}

	status = strings.ToLower(strings.TrimSpace(status))
	if status != "" {
		args = append(args, status)
		wheres = append(wheres, fmt.Sprintf("LOWER(status) = $%d", len(args)))
	}

	kodeProduk = strings.TrimSpace(kodeProduk)
	if kodeProduk != "" {
		kodes := expandKodeProdukFilters(kodeProduk)
		holders := make([]string, 0, len(kodes))
		for _, k := range kodes {
			args = append(args, k)
			holders = append(holders, fmt.Sprintf("$%d", len(args)))
		}
		wheres = append(wheres, fmt.Sprintf("UPPER(kode_produk) IN (%s)", strings.Join(holders, ",")))
	}

	refID = strings.TrimSpace(refID)
	if refID != "" {
		args = append(args, refID)
		p := fmt.Sprintf("$%d", len(args))
		wheres = append(wheres, fmt.Sprintf("ref_id ILIKE '%%' || %s || '%%'", p))
	}

	dest = strings.TrimSpace(dest)
	if dest != "" {
		args = append(args, dest)
		p := fmt.Sprintf("$%d", len(args))
		wheres = append(wheres, fmt.Sprintf("tujuan ILIKE '%%' || %s || '%%'", p))
	}

	fromStr = strings.TrimSpace(fromStr)
	if fromStr != "" {
		t, err := time.ParseInLocation("2006-01-02", fromStr, loc)
		if err != nil {
			return 0, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("tm.dibuat_pada >= $%d", len(args)))
	}

	toStr = strings.TrimSpace(toStr)
	if toStr != "" {
		t, err := time.ParseInLocation("2006-01-02", toStr, loc)
		if err != nil {
			return 0, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t.AddDate(0, 0, 1))
		wheres = append(wheres, fmt.Sprintf("tm.dibuat_pada < $%d", len(args)))
	}

	q := fmt.Sprintf(`SELECT COUNT(1) FROM transaksi_member tm LEFT JOIN public.member m ON m.id = tm.member_id WHERE %s`, strings.Join(wheres, " AND "))
	var total int64
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}
