package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func (r *AuditRepository) AdminListProviderWalletMissingDebit(
	ctx context.Context,
	limit, offset int,
	provider, refID, fromStr, toStr string,
) ([]AdminProviderWalletMissingDebitRow, int64, error) {
	if err := r.ensureProviderWalletMissingDebitIgnoreTable(ctx); err != nil {
		return nil, 0, err
	}
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

	wheres = append(wheres, "LOWER(TRIM(COALESCE(tm.status, ''))) = 'success'")
	wheres = append(wheres, "COALESCE(tp.harga, 0) > 0")
	wheres = append(wheres, "LOWER(TRIM(COALESCE(tp.status, ''))) = 'success'")

	provider = strings.TrimSpace(strings.ToLower(provider))
	if provider != "" {
		args = append(args, provider)
		wheres = append(wheres, fmt.Sprintf("LOWER(TRIM(tp.provider)) = $%d", len(args)))
	}
	refID = strings.TrimSpace(refID)
	if refID != "" {
		args = append(args, refID)
		wheres = append(wheres, fmt.Sprintf("(tp.ref_id ILIKE '%%' || $%d || '%%' OR tm.ref_id ILIKE '%%' || $%d || '%%')", len(args), len(args)))
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
JOIN public.transaksi_member tm ON tm.id = tp.transaksi_member_id
WHERE ` + whereSQL + `
  AND NOT EXISTS (
    SELECT 1
    FROM public.mutasi_dompet_provider mdp
    WHERE mdp.transaksi_provider_id = tp.id
      AND LOWER(TRIM(COALESCE(mdp.arah, ''))) = 'debit'
  )
  AND NOT EXISTS (
    SELECT 1
    FROM public.audit_provider_wallet_missing_debit_ignore ign
    WHERE ign.transaksi_provider_id = tp.id
  )
`

	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, limit, offset)
	limitPos := len(args) + 1
	offsetPos := len(args) + 2

	qList := fmt.Sprintf(`
SELECT
  tm.id AS transaksi_member_id,
  tm.member_id,
  tm.status AS status_member,
  tm.ref_id AS ref_id_member,
  tm.kode_produk AS produk_member,
  tm.tujuan AS tujuan_member,
  tm.qty AS qty_member,
  tm.qty_provider,
  tp.id AS transaksi_provider_id,
  tp.provider,
  tp.ref_id AS ref_id_provider,
  tp.perintah,
  tp.kode_produk,
  tp.tujuan,
  tp.qty,
  tp.harga,
  tp.kode_respon,
  tp.pesan,
  tp.no_referensi,
  tp.dibuat_pada AS provider_dibuat_pada
%s
ORDER BY tp.dibuat_pada DESC, tp.id DESC
LIMIT $%d OFFSET $%d
`, baseFrom, limitPos, offsetPos)

	rows, err := r.db.QueryContext(ctx, qList, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]AdminProviderWalletMissingDebitRow, 0, limit)
	for rows.Next() {
		var (
			x           AdminProviderWalletMissingDebitRow
			qtyProvider sql.NullInt64
			kodeRespon  sql.NullString
			pesan       sql.NullString
			noReferensi sql.NullString
		)
		if err := rows.Scan(
			&x.TransaksiMemberID, &x.MemberID, &x.StatusMember, &x.RefIDMember, &x.ProdukMember, &x.TujuanMember,
			&x.QtyMember, &qtyProvider, &x.TransaksiProviderID, &x.Provider, &x.RefIDProvider, &x.Perintah,
			&x.KodeProduk, &x.Tujuan, &x.Qty, &x.Harga, &kodeRespon, &pesan, &noReferensi, &x.ProviderDibuatPada,
		); err != nil {
			return nil, 0, err
		}
		x.KodeProduk = normalizeProviderProductCodeDisplay(x.Provider, x.KodeProduk)
		if qtyProvider.Valid {
			v := qtyProvider.Int64
			x.QtyProvider = &v
		}
		if kodeRespon.Valid {
			v := strings.TrimSpace(kodeRespon.String)
			x.KodeRespon = &v
		}
		if pesan.Valid {
			v := strings.TrimSpace(pesan.String)
			x.Pesan = &v
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

	qCount := fmt.Sprintf(`SELECT COUNT(1) %s`, baseFrom)
	var total int64
	if err := r.db.QueryRowContext(ctx, qCount, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	return out, total, nil
}
