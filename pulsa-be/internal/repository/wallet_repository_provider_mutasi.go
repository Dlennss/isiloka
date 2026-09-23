package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (r *WalletRepository) ListProviderDepositMutasi(ctx context.Context, provider, refID, fromStr, toStr string, limit, offset int) ([]ProviderWalletMutasiRow, int64, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	p := strings.TrimSpace(strings.ToLower(provider))
	if p != "" {
		np, err := r.normalizeWalletProvider(ctx, p)
		if err != nil {
			return nil, 0, err
		}
		p = np
	}
	refID = strings.TrimSpace(refID)
	loc, _ := time.LoadLocation("Asia/Jakarta")

	var (
		args   []any
		wheres []string
	)
	args = append(args, p)
	wheres = append(wheres, fmt.Sprintf("($%d = '' OR mdp.provider = $%d)", len(args), len(args)))
	wheres = append(wheres, "(mdp.alasan = 'PROVIDER_DEPOSIT' OR mdp.alasan = 'BANK_TRANSFER_IN')")
	wheres = append(wheres, "LOWER(TRIM(COALESCE(mdp.arah, ''))) = 'credit'")
	if refID != "" {
		args = append(args, refID)
		wheres = append(wheres, fmt.Sprintf("mdp.ref_id ILIKE '%%' || $%d || '%%'", len(args)))
	}
	if strings.TrimSpace(fromStr) != "" {
		t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(fromStr), loc)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("mdp.dibuat_pada >= $%d", len(args)))
	}
	if strings.TrimSpace(toStr) != "" {
		t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(toStr), loc)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t.AddDate(0, 0, 1))
		wheres = append(wheres, fmt.Sprintf("mdp.dibuat_pada < $%d", len(args)))
	}
	args = append(args, limit)
	limitPos := len(args)
	args = append(args, offset)
	offsetPos := len(args)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
SELECT
  mdp.id, mdp.provider, mdp.bank_id, NULLIF(mdp.bank_nama,''), mdp.ref_id, mdp.arah, mdp.jumlah, mdp.alasan, COALESCE(mdp.catatan,''), mdp.saldo_sebelum, mdp.saldo_sesudah,
  mdp.transaksi_member_id, mdp.transaksi_provider_id, mdp.diubah_oleh, actor.nama, mdp.dibuat_pada
FROM public.mutasi_dompet_provider mdp
LEFT JOIN public.member actor ON actor.id = mdp.diubah_oleh
WHERE %s
ORDER BY mdp.dibuat_pada DESC, mdp.id DESC
LIMIT $%d OFFSET $%d
`, strings.Join(wheres, " AND "), limitPos, offsetPos), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]ProviderWalletMutasiRow, 0, limit)
	for rows.Next() {
		var (
			x         ProviderWalletMutasiRow
			bankID    sql.NullInt64
			bankNama  sql.NullString
			tm        sql.NullInt64
			tp        sql.NullInt64
			actorID   sql.NullInt64
			actorName sql.NullString
		)
		if err := rows.Scan(
			&x.ID, &x.Provider, &bankID, &bankNama, &x.RefID, &x.Arah, &x.Jumlah, &x.Alasan, &x.Catatan, &x.SaldoSebelum, &x.SaldoSesudah,
			&tm, &tp, &actorID, &actorName, &x.DibuatPada,
		); err != nil {
			return nil, 0, err
		}
		if bankID.Valid {
			v := bankID.Int64
			x.BankID = &v
		}
		if bankNama.Valid {
			v := bankNama.String
			x.BankNama = &v
		}
		if tm.Valid {
			v := tm.Int64
			x.TransaksiMemberID = &v
		}
		if tp.Valid {
			v := tp.Int64
			x.TransaksiProviderID = &v
		}
		if actorID.Valid {
			v := actorID.Int64
			x.DiubahOleh = &v
		}
		if actorName.Valid {
			v := actorName.String
			x.DiubahOlehNama = &v
		}
		out = append(out, x)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	countArgs := args[:limitPos-1]
	var total int64
	if err := r.db.QueryRowContext(ctx, fmt.Sprintf(`SELECT COUNT(1) FROM public.mutasi_dompet_provider mdp WHERE %s`, strings.Join(wheres, " AND ")), countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *WalletRepository) ListProviderMutasi(ctx context.Context, provider, arah, refID, fromStr, toStr string, limit, offset int) ([]ProviderWalletMutasiRow, int64, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	p := strings.TrimSpace(strings.ToLower(provider))
	if p != "" {
		np, err := r.normalizeWalletProvider(ctx, p)
		if err != nil {
			return nil, 0, err
		}
		p = np
	}
	a := strings.TrimSpace(strings.ToLower(arah))
	if a != "" && a != "credit" && a != "debit" {
		return nil, 0, errors.New("arah must be credit|debit")
	}
	refID = strings.TrimSpace(refID)
	loc, _ := time.LoadLocation("Asia/Jakarta")

	var (
		args   []any
		wheres []string
	)
	args = append(args, p)
	wheres = append(wheres, fmt.Sprintf("($%d = '' OR mdp.provider = $%d)", len(args), len(args)))
	args = append(args, a)
	wheres = append(wheres, fmt.Sprintf("($%d = '' OR mdp.arah = $%d)", len(args), len(args)))
	wheres = append(wheres, "(mdp.alasan = 'PROVIDER_DEPOSIT' OR mdp.alasan = 'BANK_TRANSFER_IN' OR mdp.alasan LIKE 'PROVIDER_ADJUST_%' OR mdp.alasan = 'PROVIDER_SET_BALANCE')")
	if refID != "" {
		args = append(args, refID)
		wheres = append(wheres, fmt.Sprintf("mdp.ref_id ILIKE '%%' || $%d || '%%'", len(args)))
	}
	if strings.TrimSpace(fromStr) != "" {
		t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(fromStr), loc)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("mdp.dibuat_pada >= $%d", len(args)))
	}
	if strings.TrimSpace(toStr) != "" {
		t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(toStr), loc)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t.AddDate(0, 0, 1))
		wheres = append(wheres, fmt.Sprintf("mdp.dibuat_pada < $%d", len(args)))
	}
	args = append(args, limit)
	limitPos := len(args)
	args = append(args, offset)
	offsetPos := len(args)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
SELECT
  mdp.id, mdp.provider, mdp.bank_id, NULLIF(mdp.bank_nama,''), mdp.ref_id, mdp.arah, mdp.jumlah, mdp.alasan, coalesce(mdp.catatan,''), mdp.saldo_sebelum, mdp.saldo_sesudah,
  mdp.transaksi_member_id, mdp.transaksi_provider_id, mdp.diubah_oleh, actor.nama, mdp.dibuat_pada
FROM public.mutasi_dompet_provider mdp
LEFT JOIN public.member actor ON actor.id = mdp.diubah_oleh
WHERE %s
ORDER BY mdp.dibuat_pada DESC, mdp.id DESC
LIMIT $%d OFFSET $%d
`, strings.Join(wheres, " AND "), limitPos, offsetPos), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]ProviderWalletMutasiRow, 0, limit)
	for rows.Next() {
		var (
			x         ProviderWalletMutasiRow
			bankID    sql.NullInt64
			bankNama  sql.NullString
			tm        sql.NullInt64
			tp        sql.NullInt64
			actorID   sql.NullInt64
			actorName sql.NullString
		)
		if err := rows.Scan(
			&x.ID, &x.Provider, &bankID, &bankNama, &x.RefID, &x.Arah, &x.Jumlah, &x.Alasan, &x.Catatan, &x.SaldoSebelum, &x.SaldoSesudah,
			&tm, &tp, &actorID, &actorName, &x.DibuatPada,
		); err != nil {
			return nil, 0, err
		}
		if bankID.Valid {
			v := bankID.Int64
			x.BankID = &v
		}
		if bankNama.Valid {
			v := bankNama.String
			x.BankNama = &v
		}
		if tm.Valid {
			v := tm.Int64
			x.TransaksiMemberID = &v
		}
		if tp.Valid {
			v := tp.Int64
			x.TransaksiProviderID = &v
		}
		if actorID.Valid {
			v := actorID.Int64
			x.DiubahOleh = &v
		}
		if actorName.Valid {
			v := actorName.String
			x.DiubahOlehNama = &v
		}
		out = append(out, x)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int64
	countArgs := args[:limitPos-1]
	if err := r.db.QueryRowContext(ctx, fmt.Sprintf(`SELECT count(*) FROM public.mutasi_dompet_provider mdp WHERE %s`, strings.Join(wheres, " AND ")), countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}
