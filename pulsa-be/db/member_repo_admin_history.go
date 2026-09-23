package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type AdminMutasiRow struct {
	ID           int64     `json:"id"`
	MemberID     int64     `json:"member_id"`
	RefID        string    `json:"ref_id"`
	Arah         string    `json:"arah"`
	Jumlah       int64     `json:"jumlah"`
	Alasan       string    `json:"alasan"`
	Catatan      *string   `json:"catatan,omitempty"`
	SaldoSebelum *int64    `json:"saldo_sebelum,omitempty"`
	SaldoSesudah *int64    `json:"saldo_sesudah,omitempty"`
	DibuatPada   time.Time `json:"dibuat_pada"`
}

func (r *MemberRepo) AdminListMutasiByMember(
	ctx context.Context,
	memberID int64,
	limit, offset int,
	refID, arah, dateStr, fromStr, toStr string,
) ([]AdminMutasiRow, error) {
	if limit <= 0 || limit > 200 {
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
	wheres = append(wheres, fmt.Sprintf("member_id = $%d", len(args)))

	refID = strings.TrimSpace(refID)
	if refID != "" {
		args = append(args, refID)
		p := fmt.Sprintf("$%d", len(args))
		wheres = append(wheres, fmt.Sprintf("ref_id ILIKE '%%' || %s || '%%'", p))
	}

	arah = strings.TrimSpace(strings.ToLower(arah))
	if arah != "" {
		args = append(args, arah)
		wheres = append(wheres, fmt.Sprintf("LOWER(arah) = $%d", len(args)))
	}

	dateStr = strings.TrimSpace(dateStr)
	if dateStr != "" {
		d, err := time.ParseInLocation("2006-01-02", dateStr, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid date (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, d)
		wheres = append(wheres, fmt.Sprintf("dibuat_pada >= $%d", len(args)))

		args = append(args, d.AddDate(0, 0, 1))
		wheres = append(wheres, fmt.Sprintf("dibuat_pada < $%d", len(args)))
	} else {
		fromStr = strings.TrimSpace(fromStr)
		if fromStr != "" {
			t, err := time.ParseInLocation("2006-01-02", fromStr, loc)
			if err != nil {
				return nil, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
			}
			args = append(args, t)
			wheres = append(wheres, fmt.Sprintf("dibuat_pada >= $%d", len(args)))
		}

		toStr = strings.TrimSpace(toStr)
		if toStr != "" {
			t, err := time.ParseInLocation("2006-01-02", toStr, loc)
			if err != nil {
				return nil, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
			}
			args = append(args, t.AddDate(0, 0, 1))
			wheres = append(wheres, fmt.Sprintf("dibuat_pada < $%d", len(args)))
		}
	}

	args = append(args, limit)
	limitPos := len(args)
	args = append(args, offset)
	offsetPos := len(args)

	q := fmt.Sprintf(`
SELECT id, member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada
FROM mutasi_dompet
WHERE %s
ORDER BY dibuat_pada DESC, id DESC
LIMIT $%d OFFSET $%d
`, strings.Join(wheres, " AND "), limitPos, offsetPos)

	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]AdminMutasiRow, 0, 64)
	for rows.Next() {
		var m AdminMutasiRow
		var cat sql.NullString
		var ss sql.NullInt64
		var se sql.NullInt64

		if err := rows.Scan(
			&m.ID, &m.MemberID, &m.RefID, &m.Arah, &m.Jumlah, &m.Alasan,
			&cat, &ss, &se, &m.DibuatPada,
		); err != nil {
			return nil, err
		}
		if cat.Valid {
			m.Catatan = &cat.String
		}
		if ss.Valid {
			v := ss.Int64
			m.SaldoSebelum = &v
		}
		if se.Valid {
			v := se.Int64
			m.SaldoSesudah = &v
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MemberRepo) AdminCountMutasiByMember(
	ctx context.Context,
	memberID int64,
	refID, arah, dateStr, fromStr, toStr string,
) (int64, error) {
	loc, _ := time.LoadLocation("Asia/Jakarta")

	var (
		args   []any
		wheres []string
	)

	args = append(args, memberID)
	wheres = append(wheres, fmt.Sprintf("member_id = $%d", len(args)))

	refID = strings.TrimSpace(refID)
	if refID != "" {
		args = append(args, refID)
		p := fmt.Sprintf("$%d", len(args))
		wheres = append(wheres, fmt.Sprintf("ref_id ILIKE '%%' || %s || '%%'", p))
	}

	arah = strings.TrimSpace(strings.ToLower(arah))
	if arah != "" {
		args = append(args, arah)
		wheres = append(wheres, fmt.Sprintf("LOWER(arah) = $%d", len(args)))
	}

	dateStr = strings.TrimSpace(dateStr)
	if dateStr != "" {
		d, err := time.ParseInLocation("2006-01-02", dateStr, loc)
		if err != nil {
			return 0, fmt.Errorf("invalid date (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, d)
		wheres = append(wheres, fmt.Sprintf("dibuat_pada >= $%d", len(args)))

		args = append(args, d.AddDate(0, 0, 1))
		wheres = append(wheres, fmt.Sprintf("dibuat_pada < $%d", len(args)))
	} else {
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
	}

	q := fmt.Sprintf(`SELECT COUNT(1) FROM mutasi_dompet WHERE %s`, strings.Join(wheres, " AND "))
	var total int64
	if err := r.DB.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

type AdminTrxRow struct {
	ID             int64     `json:"id"`
	MemberID       int64     `json:"member_id"`
	RefID          string    `json:"ref_id"`
	Perintah       string    `json:"perintah"`
	KodeProduk     string    `json:"kode_produk"`
	Tujuan         string    `json:"tujuan"`
	Qty            int64     `json:"qty"`
	Status         string    `json:"status"`
	Keterangan     *string   `json:"keterangan,omitempty"`
	BiayaPerkiraan int64     `json:"biaya_perkiraan"`
	BiayaAktual    int64     `json:"biaya_aktual"`
	DibuatPada     time.Time `json:"dibuat_pada"`
	DiperbaruiPada time.Time `json:"diperbarui_pada"`
}

func (r *MemberRepo) AdminListTransaksiByMember(
	ctx context.Context,
	memberID int64,
	limit, offset int,
	search, fromStr, toStr string,
) ([]AdminTrxRow, error) {
	if limit <= 0 || limit > 200 {
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

	fromStr = strings.TrimSpace(fromStr)
	if fromStr != "" {
		t, err := time.ParseInLocation("2006-01-02", fromStr, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("dibuat_pada >= $%d", len(args)))
	}

	toStr = strings.TrimSpace(toStr)
	if toStr != "" {
		t, err := time.ParseInLocation("2006-01-02", toStr, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t.AddDate(0, 0, 1))
		wheres = append(wheres, fmt.Sprintf("dibuat_pada < $%d", len(args)))
	}

	args = append(args, limit)
	limitPos := len(args)
	args = append(args, offset)
	offsetPos := len(args)

	q := fmt.Sprintf(`
SELECT id, member_id, ref_id, perintah, kode_produk, tujuan, qty, status, keterangan,
       biaya_perkiraan, biaya_aktual, dibuat_pada, diperbarui_pada
FROM transaksi_member
WHERE %s
ORDER BY dibuat_pada DESC, id DESC
LIMIT $%d OFFSET $%d
`, strings.Join(wheres, " AND "), limitPos, offsetPos)

	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]AdminTrxRow, 0, 64)
	for rows.Next() {
		var t AdminTrxRow
		var ket sql.NullString
		if err := rows.Scan(
			&t.ID, &t.MemberID, &t.RefID, &t.Perintah, &t.KodeProduk, &t.Tujuan, &t.Qty,
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

func (r *MemberRepo) AdminCountTransaksiByMember(
	ctx context.Context,
	memberID int64,
	search, fromStr, toStr string,
) (int64, error) {
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
	if err := r.DB.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}
