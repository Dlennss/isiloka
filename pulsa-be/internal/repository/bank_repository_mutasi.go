package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (r *BankRepository) ListMutasi(ctx context.Context, bankID int64, arah, refID, fromStr, toStr, q string, prioritizeUnassigned bool, limit, offset int) ([]BankMutasiRow, int64, error) {
	if bankID <= 0 {
		return nil, 0, errors.New("bank_id required")
	}
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	arah = strings.TrimSpace(strings.ToLower(arah))
	if arah != "" && arah != "credit" && arah != "debit" {
		return nil, 0, errors.New("arah must be credit|debit")
	}
	refID = strings.TrimSpace(refID)
	q = strings.TrimSpace(q)
	loc, _ := time.LoadLocation("Asia/Jakarta")

	var (
		args   []any
		wheres []string
	)
	args = append(args, bankID)
	wheres = append(wheres, fmt.Sprintf("mb.bank_id = $%d", len(args)))
	if arah != "" {
		args = append(args, strings.ToUpper(arah))
		wheres = append(wheres, fmt.Sprintf("upper(mb.arah) = $%d", len(args)))
	}
	if refID != "" {
		args = append(args, refID)
		wheres = append(wheres, fmt.Sprintf("mb.ref_id ILIKE '%%' || $%d || '%%'", len(args)))
	}
	if q != "" {
		args = append(args, q)
		wheres = append(wheres, fmt.Sprintf(`(
			mb.ref_id ILIKE '%%' || $%d || '%%'
			OR mb.alasan ILIKE '%%' || $%d || '%%'
			OR COALESCE(mb.catatan, '') ILIKE '%%' || $%d || '%%'
			OR COALESCE(mb.provider, '') ILIKE '%%' || $%d || '%%'
			OR COALESCE(suspect_provider.ref_id, target_trx.ref_id, '') ILIKE '%%' || $%d || '%%'
			OR COALESCE(target_member.nama, '') ILIKE '%%' || $%d || '%%'
			OR COALESCE(actor.nama, '') ILIKE '%%' || $%d || '%%'
			OR (
				regexp_replace($%d, '[^0-9]', '', 'g') <> ''
				AND mb.jumlah::text ILIKE '%%' || regexp_replace($%d, '[^0-9]', '', 'g') || '%%'
			)
		)`, len(args), len(args), len(args), len(args), len(args), len(args), len(args), len(args), len(args)))
	}
	adminFeeExpr := bankMutasiAdminFeeExpr("mb")
	if strings.TrimSpace(fromStr) != "" {
		t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(fromStr), loc)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) >= $%d", len(args)))
	}
	if strings.TrimSpace(toStr) != "" {
		t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(toStr), loc)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t.AddDate(0, 0, 1))
		wheres = append(wheres, fmt.Sprintf("COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) < $%d", len(args)))
	}
	args = append(args, limit)
	limitPos := len(args)
	args = append(args, offset)
	offsetPos := len(args)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
SELECT
  mb.id, mb.bank_id, b.nama, mb.ref_id, mb.arah, mb.jumlah, mb.alasan, COALESCE(mb.catatan,''), mb.saldo_sebelum, mb.saldo_sesudah,
  mb.provider, COALESCE(suspect_provider.ref_id, target_trx.ref_id), mb.member_id, target_member.nama, mb.diubah_oleh, actor.nama, mb.waktu_mutasi_bank, mb.pengirim, mb.penerima, mb.dibuat_pada
FROM public.mutasi_bank mb
JOIN public.bank b ON b.id = mb.bank_id
LEFT JOIN public.member target_member ON target_member.id = mb.member_id
LEFT JOIN public.transaksi_provider suspect_provider
  ON suspect_provider.id = CASE
    WHEN mb.ref_id ~ '^TRX-SUSPECT-[0-9]+$' THEN substring(mb.ref_id from 13)::bigint
    ELSE NULL
  END
LEFT JOIN public.transaksi_member target_trx ON target_trx.member_id = mb.member_id AND target_trx.ref_id = mb.ref_id
LEFT JOIN public.member actor ON actor.id = mb.diubah_oleh
WHERE %s
ORDER BY
  %s
LIMIT $%d OFFSET $%d
`, strings.Join(wheres, " AND "), bankMutasiListOrderBy(prioritizeUnassigned, adminFeeExpr), limitPos, offsetPos), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]BankMutasiRow, 0, limit)
	for rows.Next() {
		var (
			item       BankMutasiRow
			provider   sql.NullString
			targetRef  sql.NullString
			memberID   sql.NullInt64
			memberNm   sql.NullString
			actorID    sql.NullInt64
			actorName  sql.NullString
			mutationAt sql.NullTime
			pengirim   sql.NullString
			penerima   sql.NullString
		)
		if err := rows.Scan(
			&item.ID,
			&item.BankID,
			&item.BankNama,
			&item.RefID,
			&item.Arah,
			&item.Jumlah,
			&item.Alasan,
			&item.Catatan,
			&item.SaldoSebelum,
			&item.SaldoSesudah,
			&provider,
			&targetRef,
			&memberID,
			&memberNm,
			&actorID,
			&actorName,
			&mutationAt,
			&pengirim,
			&penerima,
			&item.DibuatPada,
		); err != nil {
			return nil, 0, err
		}
		if provider.Valid {
			v := provider.String
			item.Provider = &v
		}
		if targetRef.Valid && strings.TrimSpace(targetRef.String) != "" {
			v := targetRef.String
			item.TargetRefID = &v
		}
		if memberID.Valid {
			v := memberID.Int64
			item.MemberID = &v
		}
		if memberNm.Valid {
			v := memberNm.String
			item.MemberNama = &v
		}
		if actorID.Valid {
			v := actorID.Int64
			item.DiubahOleh = &v
		}
		if actorName.Valid {
			v := actorName.String
			item.DiubahNama = &v
		}
		if mutationAt.Valid {
			v := mutationAt.Time
			item.WaktuMutasiBank = &v
		}
		if pengirim.Valid {
			v := pengirim.String
			item.Pengirim = &v
		}
		if penerima.Valid {
			v := penerima.String
			item.Penerima = &v
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	countArgs := args[:limitPos-1]
	var total int64
	if err := r.db.QueryRowContext(ctx, fmt.Sprintf(`
SELECT count(*)
FROM public.mutasi_bank mb
LEFT JOIN public.member target_member ON target_member.id = mb.member_id
LEFT JOIN public.transaksi_provider suspect_provider
  ON suspect_provider.id = CASE
    WHEN mb.ref_id ~ '^TRX-SUSPECT-[0-9]+$' THEN substring(mb.ref_id from 13)::bigint
    ELSE NULL
  END
LEFT JOIN public.transaksi_member target_trx ON target_trx.member_id = mb.member_id AND target_trx.ref_id = mb.ref_id
LEFT JOIN public.member actor ON actor.id = mb.diubah_oleh
WHERE %s
`, strings.Join(wheres, " AND ")), countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *BankRepository) ListUnpairedDebitMutasi(ctx context.Context, bankID int64, fromStr, toStr, q string, includeAdminStaffOnly bool, limit, offset int) ([]BankMutasiRow, int64, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	q = strings.TrimSpace(q)
	loc, _ := time.LoadLocation("Asia/Jakarta")
	adminFeeExpr := bankMutasiAdminFeeExpr("mb")
	providerPairExpr := bankMutasiProviderPairExistsExpr("mb")
	providerRefundPairExpr := bankMutasiProviderRefundPairExistsExpr("mb")
	excludedReasonExpr := bankMutasiUnpairedDebitExcludedReasonExpr("mb")

	args := []any{}
	wheres := []string{
		"upper(mb.arah) = 'DEBIT'",
		"mb.member_id IS NULL",
		fmt.Sprintf("NOT %s", adminFeeExpr),
		excludedReasonExpr,
		fmt.Sprintf("NOT %s", providerPairExpr),
		fmt.Sprintf("NOT %s", providerRefundPairExpr),
		`NOT EXISTS (
			SELECT 1
			FROM public.mutasi_bank dst
			WHERE dst.ref_id = mb.ref_id
			  AND dst.jumlah = mb.jumlah
			  AND dst.bank_id <> mb.bank_id
			  AND upper(dst.arah) = 'CREDIT'
			  AND COALESCE(dst.alasan, '') = 'BANK_TRANSFER_IN'
		)`,
	}
	if !includeAdminStaffOnly {
		wheres = append(wheres, "COALESCE(b.admin_staff_only, false) = false")
	}
	if bankID > 0 {
		args = append(args, bankID)
		wheres = append(wheres, fmt.Sprintf("mb.bank_id = $%d", len(args)))
	}
	if q != "" {
		args = append(args, q)
		wheres = append(wheres, fmt.Sprintf(`(
			mb.ref_id ILIKE '%%' || $%d || '%%'
			OR mb.alasan ILIKE '%%' || $%d || '%%'
			OR COALESCE(mb.catatan, '') ILIKE '%%' || $%d || '%%'
			OR COALESCE(mb.provider, '') ILIKE '%%' || $%d || '%%'
			OR COALESCE(b.nama, '') ILIKE '%%' || $%d || '%%'
			OR COALESCE(mb.pengirim, '') ILIKE '%%' || $%d || '%%'
			OR COALESCE(mb.penerima, '') ILIKE '%%' || $%d || '%%'
			OR (
				regexp_replace($%d, '[^0-9]', '', 'g') <> ''
				AND mb.jumlah::text ILIKE '%%' || regexp_replace($%d, '[^0-9]', '', 'g') || '%%'
			)
		)`, len(args), len(args), len(args), len(args), len(args), len(args), len(args), len(args), len(args)))
	}
	if strings.TrimSpace(fromStr) != "" {
		t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(fromStr), loc)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) >= $%d", len(args)))
	}
	if strings.TrimSpace(toStr) != "" {
		t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(toStr), loc)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t.AddDate(0, 0, 1))
		wheres = append(wheres, fmt.Sprintf("COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) < $%d", len(args)))
	}
	args = append(args, limit)
	limitPos := len(args)
	args = append(args, offset)
	offsetPos := len(args)
	whereSQL := strings.Join(wheres, " AND ")

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
SELECT
  mb.id, mb.bank_id, b.nama, mb.ref_id, mb.arah, mb.jumlah, mb.alasan, COALESCE(mb.catatan,''), mb.saldo_sebelum, mb.saldo_sesudah,
  mb.provider, NULL::text, mb.member_id, target_member.nama, mb.diubah_oleh, actor.nama, mb.waktu_mutasi_bank, mb.pengirim, mb.penerima, mb.dibuat_pada
FROM public.mutasi_bank mb
JOIN public.bank b ON b.id = mb.bank_id
LEFT JOIN public.member target_member ON target_member.id = mb.member_id
LEFT JOIN public.member actor ON actor.id = mb.diubah_oleh
WHERE %s
ORDER BY COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) DESC, mb.id DESC
LIMIT $%d OFFSET $%d
`, whereSQL, limitPos, offsetPos), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]BankMutasiRow, 0, limit)
	for rows.Next() {
		var (
			item       BankMutasiRow
			provider   sql.NullString
			targetRef  sql.NullString
			memberID   sql.NullInt64
			memberNm   sql.NullString
			actorID    sql.NullInt64
			actorName  sql.NullString
			mutationAt sql.NullTime
			pengirim   sql.NullString
			penerima   sql.NullString
		)
		if err := rows.Scan(
			&item.ID,
			&item.BankID,
			&item.BankNama,
			&item.RefID,
			&item.Arah,
			&item.Jumlah,
			&item.Alasan,
			&item.Catatan,
			&item.SaldoSebelum,
			&item.SaldoSesudah,
			&provider,
			&targetRef,
			&memberID,
			&memberNm,
			&actorID,
			&actorName,
			&mutationAt,
			&pengirim,
			&penerima,
			&item.DibuatPada,
		); err != nil {
			return nil, 0, err
		}
		if provider.Valid {
			v := provider.String
			item.Provider = &v
		}
		if targetRef.Valid && strings.TrimSpace(targetRef.String) != "" {
			v := targetRef.String
			item.TargetRefID = &v
		}
		if memberID.Valid {
			v := memberID.Int64
			item.MemberID = &v
		}
		if memberNm.Valid {
			v := memberNm.String
			item.MemberNama = &v
		}
		if actorID.Valid {
			v := actorID.Int64
			item.DiubahOleh = &v
		}
		if actorName.Valid {
			v := actorName.String
			item.DiubahNama = &v
		}
		if mutationAt.Valid {
			v := mutationAt.Time
			item.WaktuMutasiBank = &v
		}
		if pengirim.Valid {
			v := pengirim.String
			item.Pengirim = &v
		}
		if penerima.Valid {
			v := penerima.String
			item.Penerima = &v
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	countArgs := args[:limitPos-1]
	var total int64
	if err := r.db.QueryRowContext(ctx, fmt.Sprintf(`
SELECT count(*)
FROM public.mutasi_bank mb
JOIN public.bank b ON b.id = mb.bank_id
LEFT JOIN public.member target_member ON target_member.id = mb.member_id
LEFT JOIN public.member actor ON actor.id = mb.diubah_oleh
WHERE %s
`, whereSQL), countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func bankMutasiListOrderBy(prioritizeUnassigned bool, adminFeeExpr string) string {
	chronological := `COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) DESC,
  mb.id DESC`
	if !prioritizeUnassigned {
		return chronological
	}
	return fmt.Sprintf(`CASE
    WHEN upper(mb.arah) = 'CREDIT'
      AND mb.member_id IS NULL
      AND COALESCE(NULLIF(trim(mb.provider), ''), '') = ''
      THEN 0
    WHEN upper(mb.arah) = 'DEBIT'
      AND mb.member_id IS NULL
      AND COALESCE(NULLIF(trim(mb.provider), ''), '') = ''
      AND NOT %s
      THEN 1
    ELSE 2
  END ASC,
  %s`, adminFeeExpr, chronological)
}

func bankMutasiAdminFeeExpr(alias string) string {
	prefix := bankMutasiSQLPrefix(alias)
	return fmt.Sprintf(`(COALESCE(%salasan, '') = 'BANK_TRANSFER_ADMIN_FEE'
      OR COALESCE(%salasan, '') ILIKE '%%ADMIN_FEE%%'
      OR COALESCE(%scatatan, '') ILIKE '%%admin fee%%'
      OR COALESCE(%scatatan, '') ILIKE '%%biaya admin%%'
      OR COALESCE(%scatatan, '') ILIKE '%%biaya adm%%'
      OR COALESCE(%scatatan, '') ILIKE '%%adm bank%%'
      OR COALESCE(%scatatan, '') ILIKE '%%biaya transfer%%'
      OR COALESCE(%scatatan, '') ILIKE '%%biaya potongan%%'
      OR COALESCE(%scatatan, '') ILIKE '%%potongan rekening%%'
      OR %sjumlah IN (2500, 6500))`, prefix, prefix, prefix, prefix, prefix, prefix, prefix, prefix, prefix, prefix)
}

func bankMutasiUnpairedDebitExcludedReasonExpr(alias string) string {
	prefix := bankMutasiSQLPrefix(alias)
	return fmt.Sprintf("COALESCE(%salasan, '') NOT IN ('BANK_RECONCILE_ADJUSTMENT', 'INTERNAL_FINANCE_DEBIT', 'BANK_ADJUST_DEBIT')", prefix)
}

func bankMutasiProviderPairExistsExpr(alias string) string {
	prefix := bankMutasiSQLPrefix(alias)
	textExpr := fmt.Sprintf("concat_ws(' ', COALESCE(%scatatan, ''), COALESCE(%spengirim, ''), COALESCE(%spenerima, ''))", prefix, prefix, prefix)
	timeExpr := fmt.Sprintf("COALESCE(%swaktu_mutasi_bank, %sdibuat_pada)", prefix, prefix)
	return fmt.Sprintf(`EXISTS (
			SELECT 1
			FROM public.mutasi_dompet_provider mdp
			WHERE mdp.jumlah = %[1]sjumlah
			  AND lower(COALESCE(mdp.arah, '')) = 'credit'
			  AND COALESCE(mdp.alasan, '') = 'BANK_TRANSFER_IN'
			  AND (
			    COALESCE(NULLIF(trim(%[1]sprovider), ''), '') = ''
			    OR lower(trim(mdp.provider)) = lower(trim(%[1]sprovider))
			  )
			  AND (
			    mdp.ref_id = %[1]sref_id
			    OR (
			      COALESCE(%[1]smeta->>'manual_provider_topup_ref', '') <> ''
			      AND mdp.ref_id = %[1]smeta->>'manual_provider_topup_ref'
			    )
			    OR COALESCE(mdp.meta->>'matched_k24_ref', '') = %[1]sref_id
			    OR COALESCE(mdp.catatan, '') ILIKE '%%' || %[1]sref_id || '%%'
			    OR (
			      mdp.bank_id = %[1]sbank_id
			      AND mdp.dibuat_pada >= %[3]s - interval '72 hours'
			      AND mdp.dibuat_pada <= %[3]s + interval '72 hours'
			      AND EXISTS (
			        SELECT 1
			        FROM public.provider_rekening pr
			        WHERE pr.aktif = true
			          AND lower(trim(pr.provider)) = lower(trim(mdp.provider))
			          AND (
			            (
			              COALESCE(pr.nomor_rekening_digits, '') <> ''
			              AND regexp_replace(%[2]s, '[^0-9]', '', 'g') LIKE '%%' || pr.nomor_rekening_digits || '%%'
			            )
			            OR (
			              length(trim(regexp_replace(regexp_replace(pr.nama, '^PT[ .]*', '', 'i'), '\s+', ' ', 'g'))) >= 8
			              AND upper(%[2]s) LIKE '%%' || upper(left(trim(regexp_replace(regexp_replace(pr.nama, '^PT[ .]*', '', 'i'), '\s+', ' ', 'g')), 18)) || '%%'
			            )
			          )
			      )
			    )
			  )
		)`, prefix, textExpr, timeExpr)
}

func bankMutasiProviderRefundPairExistsExpr(alias string) string {
	prefix := bankMutasiSQLPrefix(alias)
	return fmt.Sprintf(`EXISTS (
			SELECT 1
			FROM public.mutasi_bank refund
			WHERE refund.bank_id = %[1]sbank_id
			  AND refund.jumlah = %[1]sjumlah
			  AND upper(refund.arah) = 'CREDIT'
			  AND COALESCE(refund.alasan, '') = 'BANK_TRANSFER_PROVIDER_REFUND'
			  AND (
			    refund.meta->>'refund_of_bank_ref' = %[1]sref_id
			    OR refund.meta->>'refund_of_bank_mutasi_id' = %[1]sid::text
			    OR %[1]smeta->>'provider_credit_refunded_by_bank_ref' = refund.ref_id
			  )
		)`, prefix)
}

func bankMutasiSQLPrefix(alias string) string {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return ""
	}
	return alias + "."
}

func (r *BankRepository) LatestMutasi(ctx context.Context, bankID int64) (*BankMutasiRow, error) {
	if bankID <= 0 {
		return nil, errors.New("bank_id required")
	}
	var (
		item       BankMutasiRow
		provider   sql.NullString
		memberID   sql.NullInt64
		memberNm   sql.NullString
		actorID    sql.NullInt64
		actorName  sql.NullString
		mutationAt sql.NullTime
		pengirim   sql.NullString
		penerima   sql.NullString
	)
	err := r.db.QueryRowContext(ctx, `
SELECT
  mb.id, mb.bank_id, b.nama, mb.ref_id, mb.arah, mb.jumlah, mb.alasan, COALESCE(mb.catatan,''), mb.saldo_sebelum, mb.saldo_sesudah,
  mb.provider, mb.member_id, target_member.nama, mb.diubah_oleh, actor.nama, mb.waktu_mutasi_bank, mb.pengirim, mb.penerima, mb.dibuat_pada
FROM public.mutasi_bank mb
JOIN public.bank b ON b.id = mb.bank_id
LEFT JOIN public.member target_member ON target_member.id = mb.member_id
LEFT JOIN public.member actor ON actor.id = mb.diubah_oleh
WHERE mb.bank_id = $1
ORDER BY
  CASE WHEN mb.saldo_sesudah = b.saldo THEN 0 ELSE 1 END,
  COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) DESC,
  mb.id DESC
LIMIT 1
`, bankID).Scan(
		&item.ID,
		&item.BankID,
		&item.BankNama,
		&item.RefID,
		&item.Arah,
		&item.Jumlah,
		&item.Alasan,
		&item.Catatan,
		&item.SaldoSebelum,
		&item.SaldoSesudah,
		&provider,
		&memberID,
		&memberNm,
		&actorID,
		&actorName,
		&mutationAt,
		&pengirim,
		&penerima,
		&item.DibuatPada,
	)
	if err != nil {
		return nil, err
	}
	if provider.Valid {
		v := provider.String
		item.Provider = &v
	}
	if memberID.Valid {
		v := memberID.Int64
		item.MemberID = &v
	}
	if memberNm.Valid {
		v := memberNm.String
		item.MemberNama = &v
	}
	if actorID.Valid {
		v := actorID.Int64
		item.DiubahOleh = &v
	}
	if actorName.Valid {
		v := actorName.String
		item.DiubahNama = &v
	}
	if mutationAt.Valid {
		v := mutationAt.Time
		item.WaktuMutasiBank = &v
	}
	if pengirim.Valid {
		v := pengirim.String
		item.Pengirim = &v
	}
	if penerima.Valid {
		v := penerima.String
		item.Penerima = &v
	}
	return &item, nil
}
