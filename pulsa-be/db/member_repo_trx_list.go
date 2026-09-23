package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type TrxMemberRow struct {
	ID            int64      `json:"id"`
	RefID         string     `json:"ref_id"`
	Perintah      string     `json:"perintah"`
	KodeProduk    string     `json:"kode_produk"`
	Tujuan        string     `json:"tujuan"`
	Qty           int64      `json:"qty"`
	Status        string     `json:"status"`
	Keterangan    *string    `json:"keterangan,omitempty"`
	BiayaPerkiraan int64     `json:"biaya_perkiraan"`
	BiayaAktual    int64     `json:"biaya_aktual"`
	DibuatPada    time.Time  `json:"dibuat_pada"`
	DiperbaruiPada time.Time `json:"diperbarui_pada"`
}

// fromStr/toStr:
// - "" berarti tidak dipakai
// - format disarankan: "YYYY-MM-DD"
//   toStr diperlakukan inclusive (sampai akhir hari), implementasi pakai < (to+1 hari)
func (r *MemberRepo) ListTransaksi(
	ctx context.Context,
	memberID int64,
	limit int,
	offset int,
	search string,
	fromStr string,
	toStr string,
) ([]TrxMemberRow, error) {
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

	// wajib: member_id
	args = append(args, memberID)
	wheres = append(wheres, fmt.Sprintf("member_id=$%d", len(args)))

	// search (ILIKE) di beberapa kolom
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

	// from (>= start of day)
	if strings.TrimSpace(fromStr) != "" {
		t, err := time.ParseInLocation("2006-01-02", fromStr, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("dibuat_pada >= $%d", len(args)))
	}

	// to inclusive: dibuat_pada < (to + 1 day)
	if strings.TrimSpace(toStr) != "" {
		t, err := time.ParseInLocation("2006-01-02", toStr, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
		}
		t = t.AddDate(0, 0, 1)
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("dibuat_pada < $%d", len(args)))
	}

	// limit & offset
	args = append(args, limit)
	limitPos := len(args)
	args = append(args, offset)
	offsetPos := len(args)

	sqlQ := fmt.Sprintf(`
SELECT id, ref_id, perintah, kode_produk, tujuan, qty, status, keterangan,
       biaya_perkiraan, biaya_aktual, dibuat_pada, diperbarui_pada
FROM public.transaksi_member
WHERE %s
ORDER BY dibuat_pada DESC, id DESC
LIMIT $%d OFFSET $%d
`, strings.Join(wheres, " AND "), limitPos, offsetPos)

	rows, err := r.DB.QueryContext(ctx, sqlQ, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TrxMemberRow
	for rows.Next() {
		var t TrxMemberRow
		var ket sql.NullString
		if err := rows.Scan(
			&t.ID, &t.RefID, &t.Perintah, &t.KodeProduk, &t.Tujuan, &t.Qty,
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
