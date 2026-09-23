package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func (r *HistoryRepository) ListTrxMemberStatusLogs(ctx context.Context, f TrxMemberStatusLogFilter) ([]TrxMemberStatusLogRow, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 10
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

	if f.TrxID > 0 {
		args = append(args, f.TrxID)
		wheres = append(wheres, fmt.Sprintf("l.transaksi_member_id = $%d", len(args)))
	}
	if refID := strings.TrimSpace(f.RefID); refID != "" {
		args = append(args, refID)
		wheres = append(wheres, fmt.Sprintf("tm.ref_id = $%d", len(args)))
	}
	if f.DiubahOleh > 0 {
		args = append(args, f.DiubahOleh)
		wheres = append(wheres, fmt.Sprintf("l.diubah_oleh = $%d", len(args)))
	}
	if f.ManualOnly {
		wheres = append(wheres, "l.diubah_oleh IS NOT NULL")
	}
	if fromStr := strings.TrimSpace(f.From); fromStr != "" {
		t, err := time.ParseInLocation("2006-01-02", fromStr, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("l.dibuat_pada >= $%d", len(args)))
	}
	if toStr := strings.TrimSpace(f.To); toStr != "" {
		t, err := time.ParseInLocation("2006-01-02", toStr, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t.AddDate(0, 0, 1))
		wheres = append(wheres, fmt.Sprintf("l.dibuat_pada < $%d", len(args)))
	}
	if len(wheres) == 0 {
		wheres = append(wheres, "1=1")
	}

	args = append(args, limit)
	limitPos := len(args)
	args = append(args, offset)
	offsetPos := len(args)

	q := fmt.Sprintf(`
SELECT
  l.id,
  l.transaksi_member_id,
  tm.ref_id,
  l.status_sebelum,
  l.status_sesudah,
  l.keterangan_sebelum,
  l.keterangan_sesudah,
  l.biaya_aktual_sebelum,
  l.biaya_aktual_sesudah,
  l.aksi,
  l.diubah_oleh,
  actor.nama,
  l.dibuat_pada
FROM public.transaksi_member_status_log l
JOIN public.transaksi_member tm ON tm.id = l.transaksi_member_id
LEFT JOIN public.member actor ON actor.id = l.diubah_oleh
WHERE %s
ORDER BY l.dibuat_pada DESC, l.id DESC
LIMIT $%d OFFSET $%d
`, strings.Join(wheres, " AND "), limitPos, offsetPos)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]TrxMemberStatusLogRow, 0, limit)
	for rows.Next() {
		var (
			item           TrxMemberStatusLogRow
			statusBefore   sql.NullString
			statusAfter    sql.NullString
			ketBefore      sql.NullString
			ketAfter       sql.NullString
			diubahOleh     sql.NullInt64
			diubahOlehNama sql.NullString
		)
		if err := rows.Scan(
			&item.ID,
			&item.TransaksiMemberID,
			&item.RefID,
			&statusBefore,
			&statusAfter,
			&ketBefore,
			&ketAfter,
			&item.BiayaAktualSebelum,
			&item.BiayaAktualSesudah,
			&item.Aksi,
			&diubahOleh,
			&diubahOlehNama,
			&item.DibuatPada,
		); err != nil {
			return nil, err
		}
		if statusBefore.Valid {
			item.StatusSebelum = &statusBefore.String
		}
		if statusAfter.Valid {
			item.StatusSesudah = &statusAfter.String
		}
		if ketBefore.Valid {
			item.KeteranganSebelum = &ketBefore.String
		}
		if ketAfter.Valid {
			item.KeteranganSesudah = &ketAfter.String
		}
		if diubahOleh.Valid {
			v := diubahOleh.Int64
			item.DiubahOleh = &v
		}
		if diubahOlehNama.Valid {
			v := diubahOlehNama.String
			item.DiubahOlehNama = &v
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *HistoryRepository) CountTrxMemberStatusLogs(ctx context.Context, f TrxMemberStatusLogFilter) (int64, error) {
	loc, _ := time.LoadLocation("Asia/Jakarta")

	var (
		args   []any
		wheres []string
	)

	if f.TrxID > 0 {
		args = append(args, f.TrxID)
		wheres = append(wheres, fmt.Sprintf("l.transaksi_member_id = $%d", len(args)))
	}
	if refID := strings.TrimSpace(f.RefID); refID != "" {
		args = append(args, refID)
		wheres = append(wheres, fmt.Sprintf("tm.ref_id = $%d", len(args)))
	}
	if f.DiubahOleh > 0 {
		args = append(args, f.DiubahOleh)
		wheres = append(wheres, fmt.Sprintf("l.diubah_oleh = $%d", len(args)))
	}
	if f.ManualOnly {
		wheres = append(wheres, "l.diubah_oleh IS NOT NULL")
	}
	if fromStr := strings.TrimSpace(f.From); fromStr != "" {
		t, err := time.ParseInLocation("2006-01-02", fromStr, loc)
		if err != nil {
			return 0, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("l.dibuat_pada >= $%d", len(args)))
	}
	if toStr := strings.TrimSpace(f.To); toStr != "" {
		t, err := time.ParseInLocation("2006-01-02", toStr, loc)
		if err != nil {
			return 0, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t.AddDate(0, 0, 1))
		wheres = append(wheres, fmt.Sprintf("l.dibuat_pada < $%d", len(args)))
	}
	if len(wheres) == 0 {
		wheres = append(wheres, "1=1")
	}

	q := fmt.Sprintf(`
SELECT COUNT(*)
FROM public.transaksi_member_status_log l
JOIN public.transaksi_member tm ON tm.id = l.transaksi_member_id
WHERE %s
`, strings.Join(wheres, " AND "))

	var total int64
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}
