package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func (r *HistoryRepository) ListMutasi(ctx context.Context, memberID int64, f MutasiFilter) ([]MutasiRow, error) {
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
	wheres = append(wheres, fmt.Sprintf("md.member_id=$%d", len(args)))

	refID := strings.TrimSpace(f.RefID)
	if refID != "" {
		args = append(args, refID)
		p := fmt.Sprintf("$%d", len(args))
		wheres = append(wheres, fmt.Sprintf("md.ref_id ILIKE '%%' || %s || '%%'", p))
	}

	arah := strings.TrimSpace(strings.ToLower(f.Arah))
	if arah != "" {
		switch arah {
		case "credit":
			wheres = append(wheres, "LOWER(md.arah) = 'credit'")
		case "debit":
			wheres = append(wheres, "LOWER(md.arah) = 'debit'")
		default:
			args = append(args, arah)
			wheres = append(wheres, fmt.Sprintf("LOWER(md.arah) = $%d", len(args)))
		}
	}

	dateStr := strings.TrimSpace(f.Date)
	if dateStr != "" {
		d, err := time.ParseInLocation("2006-01-02", dateStr, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid date (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, d)
		wheres = append(wheres, fmt.Sprintf("md.dibuat_pada >= $%d", len(args)))

		args = append(args, d.AddDate(0, 0, 1))
		wheres = append(wheres, fmt.Sprintf("md.dibuat_pada < $%d", len(args)))
	} else {
		fromStr := strings.TrimSpace(f.From)
		if fromStr != "" {
			t, err := time.ParseInLocation("2006-01-02", fromStr, loc)
			if err != nil {
				return nil, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
			}
			args = append(args, t)
			wheres = append(wheres, fmt.Sprintf("md.dibuat_pada >= $%d", len(args)))
		}

		toStr := strings.TrimSpace(f.To)
		if toStr != "" {
			t, err := time.ParseInLocation("2006-01-02", toStr, loc)
			if err != nil {
				return nil, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
			}
			args = append(args, t.AddDate(0, 0, 1))
			wheres = append(wheres, fmt.Sprintf("md.dibuat_pada < $%d", len(args)))
		}
	}

	args = append(args, limit)
	limitPos := len(args)
	args = append(args, offset)
	offsetPos := len(args)

	q := fmt.Sprintf(`
SELECT md.id, md.ref_id, md.arah, md.jumlah, md.alasan, md.catatan, md.saldo_sebelum, md.saldo_sesudah, md.diubah_oleh, actor.nama, md.dibuat_pada
FROM public.mutasi_dompet md
LEFT JOIN public.member actor ON actor.id = md.diubah_oleh
WHERE %s
ORDER BY md.dibuat_pada DESC, md.id DESC
LIMIT $%d OFFSET $%d
`, strings.Join(wheres, " AND "), limitPos, offsetPos)

	rows, err := r.db.QueryContext(ctx, q, args...)
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
		var actorID sql.NullInt64
		var actorName sql.NullString

		if err := rows.Scan(&m.ID, &m.RefID, &m.Arah, &m.Jumlah, &m.Alasan, &cat, &ss, &se, &actorID, &actorName, &m.DibuatPada); err != nil {
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
		if actorID.Valid {
			v := actorID.Int64
			m.DiubahOleh = &v
		}
		if actorName.Valid {
			v := actorName.String
			m.DiubahOlehNama = &v
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *HistoryRepository) CountMutasi(ctx context.Context, memberID int64, f MutasiFilter) (int64, error) {
	loc, _ := time.LoadLocation("Asia/Jakarta")

	var (
		args   []any
		wheres []string
	)

	args = append(args, memberID)
	wheres = append(wheres, fmt.Sprintf("md.member_id=$%d", len(args)))

	refID := strings.TrimSpace(f.RefID)
	if refID != "" {
		args = append(args, refID)
		p := fmt.Sprintf("$%d", len(args))
		wheres = append(wheres, fmt.Sprintf("md.ref_id ILIKE '%%' || %s || '%%'", p))
	}

	arah := strings.TrimSpace(strings.ToLower(f.Arah))
	if arah != "" {
		switch arah {
		case "credit":
			wheres = append(wheres, "LOWER(md.arah) = 'credit'")
		case "debit":
			wheres = append(wheres, "LOWER(md.arah) = 'debit'")
		default:
			args = append(args, arah)
			wheres = append(wheres, fmt.Sprintf("LOWER(md.arah) = $%d", len(args)))
		}
	}

	dateStr := strings.TrimSpace(f.Date)
	if dateStr != "" {
		d, err := time.ParseInLocation("2006-01-02", dateStr, loc)
		if err != nil {
			return 0, fmt.Errorf("invalid date (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, d)
		wheres = append(wheres, fmt.Sprintf("md.dibuat_pada >= $%d", len(args)))

		args = append(args, d.AddDate(0, 0, 1))
		wheres = append(wheres, fmt.Sprintf("md.dibuat_pada < $%d", len(args)))
	} else {
		fromStr := strings.TrimSpace(f.From)
		if fromStr != "" {
			t, err := time.ParseInLocation("2006-01-02", fromStr, loc)
			if err != nil {
				return 0, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
			}
			args = append(args, t)
			wheres = append(wheres, fmt.Sprintf("md.dibuat_pada >= $%d", len(args)))
		}

		toStr := strings.TrimSpace(f.To)
		if toStr != "" {
			t, err := time.ParseInLocation("2006-01-02", toStr, loc)
			if err != nil {
				return 0, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
			}
			args = append(args, t.AddDate(0, 0, 1))
			wheres = append(wheres, fmt.Sprintf("md.dibuat_pada < $%d", len(args)))
		}
	}

	q := fmt.Sprintf(`SELECT COUNT(1) FROM public.mutasi_dompet md WHERE %s`, strings.Join(wheres, " AND "))
	var total int64
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *HistoryRepository) GetMutasiByID(ctx context.Context, memberID, id int64) (*MutasiRow, error) {
	var row MutasiRow
	var cat sql.NullString
	var ss sql.NullInt64
	var se sql.NullInt64
	var actorID sql.NullInt64
	var actorName sql.NullString

	err := r.db.QueryRowContext(ctx, `
SELECT md.id, md.member_id, md.ref_id, md.arah, md.jumlah, md.alasan, md.catatan, md.saldo_sebelum, md.saldo_sesudah, md.diubah_oleh, actor.nama, md.dibuat_pada
FROM public.mutasi_dompet md
LEFT JOIN public.member actor ON actor.id = md.diubah_oleh
WHERE md.member_id = $1 AND md.id = $2
LIMIT 1
`, memberID, id).Scan(
		&row.ID,
		&row.MemberID,
		&row.RefID,
		&row.Arah,
		&row.Jumlah,
		&row.Alasan,
		&cat,
		&ss,
		&se,
		&actorID,
		&actorName,
		&row.DibuatPada,
	)
	if err != nil {
		return nil, err
	}
	if cat.Valid {
		row.Catatan = &cat.String
	}
	if ss.Valid {
		v := ss.Int64
		row.SaldoSebelum = &v
	}
	if se.Valid {
		v := se.Int64
		row.SaldoSesudah = &v
	}
	if actorID.Valid {
		v := actorID.Int64
		row.DiubahOleh = &v
	}
	if actorName.Valid {
		v := actorName.String
		row.DiubahOlehNama = &v
	}
	return &row, nil
}
