package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type MutasiRow struct {
	ID           int64     `json:"id"`
	RefID        string    `json:"ref_id"`
	Arah         string    `json:"arah"`
	Jumlah       int64     `json:"jumlah"`
	Alasan       string    `json:"alasan"`
	Catatan      *string   `json:"catatan,omitempty"`
	SaldoSebelum *int64    `json:"saldo_sebelum,omitempty"`
	SaldoSesudah *int64    `json:"saldo_sesudah,omitempty"`
	DibuatPada   time.Time `json:"dibuat_pada"`
}

type MutasiFilter struct {
	RefID  string
	Arah   string
	Date   string // YYYY-MM-DD (shortcut exact-day)
	From   string // YYYY-MM-DD
	To     string // YYYY-MM-DD (inclusive)
	Limit  int
	Offset int
}

func (r *MemberRepo) GetSaldo(ctx context.Context, memberID int64) (int64, error) {
	var saldo int64
	err := r.DB.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_member
WHERE member_id=$1
`, memberID).Scan(&saldo)
	return saldo, err
}

func (r *MemberRepo) ListMutasi(ctx context.Context, memberID int64, f MutasiFilter) ([]MutasiRow, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := f.Offset
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

	refID := strings.TrimSpace(f.RefID)
	if refID != "" {
		args = append(args, refID)
		p := fmt.Sprintf("$%d", len(args))
		wheres = append(wheres, fmt.Sprintf("ref_id ILIKE '%%' || %s || '%%'", p))
	}

	arah := strings.TrimSpace(strings.ToLower(f.Arah))
	if arah != "" {
		args = append(args, arah)
		wheres = append(wheres, fmt.Sprintf("LOWER(arah) = $%d", len(args)))
	}

	// date (exact-day) has priority over from/to if provided.
	dateStr := strings.TrimSpace(f.Date)
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
		fromStr := strings.TrimSpace(f.From)
		if fromStr != "" {
			t, err := time.ParseInLocation("2006-01-02", fromStr, loc)
			if err != nil {
				return nil, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
			}
			args = append(args, t)
			wheres = append(wheres, fmt.Sprintf("dibuat_pada >= $%d", len(args)))
		}

		toStr := strings.TrimSpace(f.To)
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
SELECT id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada
FROM public.mutasi_dompet
WHERE %s
ORDER BY dibuat_pada DESC, id DESC
LIMIT $%d OFFSET $%d
`, strings.Join(wheres, " AND "), limitPos, offsetPos)

	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []MutasiRow
	for rows.Next() {
		var m MutasiRow
		var cat sql.NullString
		var ss sql.NullInt64
		var se sql.NullInt64

		if err := rows.Scan(&m.ID, &m.RefID, &m.Arah, &m.Jumlah, &m.Alasan, &cat, &ss, &se, &m.DibuatPada); err != nil {
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
