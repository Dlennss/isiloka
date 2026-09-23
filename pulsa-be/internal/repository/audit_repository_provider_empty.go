package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func (r *AuditRepository) AdminListProviderEmptyResponse(
	ctx context.Context,
	limit, offset int,
	provider, refID, kodeProduk, tujuan, fromStr, toStr string,
) ([]AdminProviderEmptyResponseRow, int64, error) {
	if limit <= 0 {
		limit = 10
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

	wheres = append(wheres, "LOWER(TRIM(COALESCE(tm.status, ''))) = 'pending'")
	wheres = append(wheres, "COALESCE(BTRIM(tp.kode_respon), '') = ''")
	wheres = append(wheres, "COALESCE(BTRIM(tp.pesan), '') = ''")

	provider = strings.TrimSpace(strings.ToLower(provider))
	if provider != "" {
		args = append(args, provider)
		wheres = append(wheres, fmt.Sprintf("LOWER(TRIM(tp.provider)) = $%d", len(args)))
	}
	refID = strings.TrimSpace(refID)
	if refID != "" {
		args = append(args, refID)
		wheres = append(wheres, fmt.Sprintf("tp.ref_id ILIKE '%%' || $%d || '%%'", len(args)))
	}
	kodeProduk = strings.TrimSpace(kodeProduk)
	if kodeProduk != "" {
		args = append(args, kodeProduk)
		wheres = append(wheres, fmt.Sprintf("tp.kode_produk ILIKE '%%' || $%d || '%%'", len(args)))
	}
	tujuan = strings.TrimSpace(tujuan)
	if tujuan != "" {
		args = append(args, tujuan)
		wheres = append(wheres, fmt.Sprintf("tp.tujuan ILIKE '%%' || $%d || '%%'", len(args)))
	}
	fromStr = strings.TrimSpace(fromStr)
	if fromStr != "" {
		t, err := time.ParseInLocation("2006-01-02", fromStr, loc)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("tp.dibuat_pada >= $%d", len(args)))
	}
	toStr = strings.TrimSpace(toStr)
	if toStr != "" {
		t, err := time.ParseInLocation("2006-01-02", toStr, loc)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t.AddDate(0, 0, 1))
		wheres = append(wheres, fmt.Sprintf("tp.dibuat_pada < $%d", len(args)))
	}

	whereSQL := strings.Join(wheres, " AND ")
	baseFrom := `
FROM public.transaksi_provider tp
LEFT JOIN public.transaksi_member tm ON tm.id = tp.transaksi_member_id
LEFT JOIN public.member m ON m.id = tm.member_id
`

	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, limit, offset)
	limitPos := len(args) + 1
	offsetPos := len(args) + 2

	qList := fmt.Sprintf(`
SELECT
  tp.id,
  tp.transaksi_member_id,
  tm.member_id,
  m.nama,
  tp.provider,
  tp.ref_id,
  tp.perintah,
  tp.kode_produk,
  tp.tujuan,
  tp.qty,
  tp.kode_respon,
  tp.pesan,
  tp.harga,
  tp.http_status,
  tp.percobaan,
  tp.dibuat_pada,
  tp.no_referensi
%s
WHERE %s
ORDER BY tp.dibuat_pada DESC, tp.id DESC
LIMIT $%d OFFSET $%d
`, baseFrom, whereSQL, limitPos, offsetPos)

	rows, err := r.db.QueryContext(ctx, qList, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]AdminProviderEmptyResponseRow, 0, limit)
	for rows.Next() {
		var (
			x           AdminProviderEmptyResponseRow
			memberID    sql.NullInt64
			memberNama  sql.NullString
			kodeRespon  sql.NullString
			pesan       sql.NullString
			harga       sql.NullInt64
			httpStatus  sql.NullInt64
			noReferensi sql.NullString
		)
		if err := rows.Scan(
			&x.TransaksiProviderID, &x.TransaksiMemberID, &memberID, &memberNama, &x.Provider, &x.RefID, &x.Perintah,
			&x.KodeProduk, &x.Tujuan, &x.Qty, &kodeRespon, &pesan, &harga, &httpStatus, &x.Percobaan, &x.DibuatPada, &noReferensi,
		); err != nil {
			return nil, 0, err
		}
		x.KodeProduk = normalizeProviderProductCodeDisplay(x.Provider, x.KodeProduk)
		if memberID.Valid {
			v := memberID.Int64
			x.MemberID = &v
		}
		if memberNama.Valid {
			v := strings.TrimSpace(memberNama.String)
			x.MemberNama = &v
		}
		if kodeRespon.Valid {
			v := strings.TrimSpace(kodeRespon.String)
			x.KodeRespon = &v
		}
		if pesan.Valid {
			v := strings.TrimSpace(pesan.String)
			x.Pesan = &v
		}
		if harga.Valid {
			v := harga.Int64
			x.Harga = &v
		}
		if httpStatus.Valid {
			v := int(httpStatus.Int64)
			x.HTTPStatus = &v
		}
		if noReferensi.Valid {
			v := strings.TrimSpace(noReferensi.String)
			x.NoReferensi = &v
		}
		out = append(out, x)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	qCount := fmt.Sprintf(`SELECT COUNT(1) %s WHERE %s`, baseFrom, whereSQL)
	var total int64
	if err := r.db.QueryRowContext(ctx, qCount, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}
