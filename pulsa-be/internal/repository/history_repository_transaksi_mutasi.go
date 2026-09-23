package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func (r *HistoryRepository) AdminListMutasiByMember(ctx context.Context, memberID int64, limit, offset int, refID, arah, dateStr, fromStr, toStr string, walletOnly bool) ([]AdminMutasiRow, error) {
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

	if memberID > 0 {
		args = append(args, memberID)
		wheres = append(wheres, fmt.Sprintf("md.member_id = $%d", len(args)))
	}

	refID = strings.TrimSpace(refID)
	if refID != "" {
		args = append(args, refID)
		p := fmt.Sprintf("$%d", len(args))
		wheres = append(wheres, fmt.Sprintf("md.ref_id ILIKE '%%' || %s || '%%'", p))
	}

	arah = strings.TrimSpace(strings.ToLower(arah))
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

	if walletOnly {
		wheres = append(wheres, "(LOWER(TRIM(COALESCE(md.alasan, ''))) = 'deposit' OR md.diubah_oleh IS NOT NULL)")
	}

	dateStr = strings.TrimSpace(dateStr)
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
		fromStr = strings.TrimSpace(fromStr)
		if fromStr != "" {
			t, err := time.ParseInLocation("2006-01-02", fromStr, loc)
			if err != nil {
				return nil, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
			}
			args = append(args, t)
			wheres = append(wheres, fmt.Sprintf("md.dibuat_pada >= $%d", len(args)))
		}

		toStr = strings.TrimSpace(toStr)
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
SELECT md.id, md.member_id, mb.nama, md.ref_id, md.arah, md.jumlah, md.alasan, md.catatan, md.saldo_sebelum, md.saldo_sesudah, md.diubah_oleh, actor.nama, md.dibuat_pada
FROM mutasi_dompet md
JOIN public.member mb ON mb.id = md.member_id
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

	out := make([]AdminMutasiRow, 0, 64)
	for rows.Next() {
		var m AdminMutasiRow
		var cat sql.NullString
		var ss sql.NullInt64
		var se sql.NullInt64
		var memberName sql.NullString
		var actorID sql.NullInt64
		var actorName sql.NullString

		if err := rows.Scan(
			&m.ID, &m.MemberID, &memberName, &m.RefID, &m.Arah, &m.Jumlah, &m.Alasan,
			&cat, &ss, &se, &actorID, &actorName, &m.DibuatPada,
		); err != nil {
			return nil, err
		}
		if memberName.Valid {
			v := memberName.String
			m.MemberNama = &v
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

func (r *HistoryRepository) AdminCountMutasiByMember(ctx context.Context, memberID int64, refID, arah, dateStr, fromStr, toStr string, walletOnly bool) (int64, error) {
	loc, _ := time.LoadLocation("Asia/Jakarta")

	var (
		args   []any
		wheres []string
	)

	if memberID > 0 {
		args = append(args, memberID)
		wheres = append(wheres, fmt.Sprintf("member_id = $%d", len(args)))
	}

	refID = strings.TrimSpace(refID)
	if refID != "" {
		args = append(args, refID)
		p := fmt.Sprintf("$%d", len(args))
		wheres = append(wheres, fmt.Sprintf("ref_id ILIKE '%%' || %s || '%%'", p))
	}

	arah = strings.TrimSpace(strings.ToLower(arah))
	if arah != "" {
		switch arah {
		case "credit":
			wheres = append(wheres, "LOWER(arah) = 'credit'")
		case "debit":
			wheres = append(wheres, "LOWER(arah) = 'debit'")
		default:
			args = append(args, arah)
			wheres = append(wheres, fmt.Sprintf("LOWER(arah) = $%d", len(args)))
		}
	}

	if walletOnly {
		wheres = append(wheres, "(LOWER(TRIM(COALESCE(alasan, ''))) = 'deposit' OR diubah_oleh IS NOT NULL)")
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
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}
