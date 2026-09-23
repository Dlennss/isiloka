package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"pulsa2/internal/helper"
)

type AuditorRepository struct {
	db *sql.DB
}

func NewAuditorRepository(db *sql.DB) *AuditorRepository {
	return &AuditorRepository{db: db}
}

func normalizeAuditorScope(scope string) string {
	switch strings.TrimSpace(strings.ToLower(scope)) {
	case "", "all", "retail", "h2h":
		return strings.TrimSpace(strings.ToLower(scope))
	default:
		return ""
	}
}

func normalizeAuditorPeriod(period string) string {
	switch strings.TrimSpace(strings.ToLower(period)) {
	case "daily", "day", "":
		return "day"
	case "monthly", "month":
		return "month"
	case "quarter", "3month", "3months":
		return "quarter"
	default:
		return "day"
	}
}

func normalizeInternalEntryType(entryType string) (string, string, error) {
	switch strings.TrimSpace(strings.ToLower(entryType)) {
	case "purchase":
		return "purchase", "debit", nil
	case "salary":
		return "salary", "debit", nil
	case "other_expense":
		return "other_expense", "debit", nil
	case "other_income":
		return "other_income", "credit", nil
	case "bank_admin":
		return "bank_admin", "debit", nil
	case "bank_interest":
		return "bank_interest", "credit", nil
	case "bank_interest_tax":
		return "bank_interest_tax", "debit", nil
	case "rent_expense":
		return "rent_expense", "debit", nil
	case "audit_expense":
		return "audit_expense", "debit", nil
	case "printing_expense":
		return "printing_expense", "debit", nil
	case "event_expense":
		return "event_expense", "debit", nil
	case "tax_income_expense":
		return "tax_income_expense", "debit", nil
	case "operational_expense":
		return "operational_expense", "debit", nil
	default:
		return "", "", errors.New("entry_type tidak valid")
	}
}

func (r *AuditorRepository) CreateInternalFinanceEntry(ctx context.Context, actorID int64, in InternalFinanceCreateInput) (*InternalFinanceEntryRow, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("db not initialized")
	}
	if actorID <= 0 {
		return nil, errors.New("actor invalid")
	}
	entryType, direction, err := normalizeInternalEntryType(in.EntryType)
	if err != nil {
		return nil, err
	}
	if in.BankID <= 0 {
		return nil, errors.New("bank_id required")
	}
	if in.Amount <= 0 {
		return nil, errors.New("amount must be > 0")
	}
	if in.Fee < 0 {
		return nil, errors.New("fee must be >= 0")
	}
	occurredAt := in.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now()
	}
	totalAmount := in.Amount
	if direction == "debit" {
		totalAmount += in.Fee
	}
	if totalAmount <= 0 {
		return nil, errors.New("total amount invalid")
	}

	refID := "FIN-" + time.Now().Format("20060102150405") + "-" + strings.ToUpper(helper.RandHex(4))
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var bankName string
	var bankBefore int64
	if err := tx.QueryRowContext(ctx, `
SELECT nama, saldo
FROM public.bank
WHERE id = $1
FOR UPDATE
`, in.BankID).Scan(&bankName, &bankBefore); err != nil {
		return nil, err
	}

	var bankAfter int64
	arahMutasi := "CREDIT"
	if direction == "debit" {
		if bankBefore < totalAmount {
			return nil, errors.New("saldo bank tidak cukup")
		}
		bankAfter = bankBefore - totalAmount
		arahMutasi = "DEBIT"
	} else {
		bankAfter = bankBefore + totalAmount
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE public.bank
SET saldo = $2, diubah_pada = now()
WHERE id = $1
`, in.BankID, bankAfter); err != nil {
		return nil, err
	}

	metaJSON, _ := json.Marshal(map[string]any{
		"type":         "internal_finance",
		"entry_type":   entryType,
		"category":     strings.TrimSpace(in.Category),
		"amount":       in.Amount,
		"fee":          in.Fee,
		"total_amount": totalAmount,
		"counterparty": strings.TrimSpace(in.Counterparty),
	})
	reason := "INTERNAL_FINANCE_" + strings.ToUpper(direction)
	note := strings.TrimSpace(in.Note)
	if _, err := tx.ExecContext(ctx, `
INSERT INTO public.mutasi_bank
  (bank_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, diubah_oleh, dibuat_pada, meta)
VALUES
  ($1,$2,$3,$4,$5,NULLIF($6,''),$7,$8,$9,$10,$11::jsonb)
`, in.BankID, refID, arahMutasi, totalAmount, reason, note, bankBefore, bankAfter, actorID, occurredAt, string(metaJSON)); err != nil {
		return nil, err
	}

	var category = strings.TrimSpace(in.Category)
	if category == "" {
		category = entryType
	}

	var out InternalFinanceEntryRow
	var createdBy sql.NullInt64
	var createdName sql.NullString
	if err := tx.QueryRowContext(ctx, `
INSERT INTO public.internal_finance_entry
  (ref_id, entry_type, category, direction, bank_id, amount, fee, total_amount, counterparty, note, occurred_at, created_by, created_at, updated_at, meta)
VALUES
  ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,now(),now(),$13::jsonb)
RETURNING id, ref_id, entry_type, category, direction, bank_id, $14, amount, fee, total_amount, counterparty, note, occurred_at, created_by, created_at
`, refID, entryType, category, direction, in.BankID, in.Amount, in.Fee, totalAmount, strings.TrimSpace(in.Counterparty), note, occurredAt, actorID, string(metaJSON), bankName).Scan(
		&out.ID, &out.RefID, &out.EntryType, &out.Category, &out.Direction, &out.BankID, &out.BankNama,
		&out.Amount, &out.Fee, &out.TotalAmount, &out.Counterparty, &out.Note, &out.OccurredAt, &createdBy, &out.CreatedAt,
	); err != nil {
		return nil, err
	}
	if createdBy.Valid {
		v := createdBy.Int64
		out.CreatedBy = &v
		if err := tx.QueryRowContext(ctx, `SELECT COALESCE(nama,'') FROM public.member WHERE id = $1`, v).Scan(&createdName); err == nil && createdName.Valid {
			s := createdName.String
			out.CreatedNama = &s
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *AuditorRepository) ListInternalFinanceEntries(ctx context.Context, bankID int64, entryType, fromStr, toStr string, limit, offset int, includeAdminStaffOnly bool) ([]InternalFinanceEntryRow, int64, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	loc, _ := time.LoadLocation("Asia/Jakarta")

	args := make([]any, 0, 8)
	wheres := make([]string, 0, 8)
	if !includeAdminStaffOnly {
		wheres = append(wheres, "COALESCE(b.admin_staff_only, false) = false")
	}
	if bankID > 0 {
		args = append(args, bankID)
		wheres = append(wheres, fmt.Sprintf("ife.bank_id = $%d", len(args)))
	}
	if strings.TrimSpace(entryType) != "" {
		args = append(args, strings.TrimSpace(strings.ToLower(entryType)))
		wheres = append(wheres, fmt.Sprintf("lower(ife.entry_type) = $%d", len(args)))
	}
	if strings.TrimSpace(fromStr) != "" {
		t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(fromStr), loc)
		if err != nil {
			return nil, 0, err
		}
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("ife.occurred_at >= $%d", len(args)))
	}
	if strings.TrimSpace(toStr) != "" {
		t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(toStr), loc)
		if err != nil {
			return nil, 0, err
		}
		args = append(args, t.AddDate(0, 0, 1))
		wheres = append(wheres, fmt.Sprintf("ife.occurred_at < $%d", len(args)))
	}
	whereSQL := "1=1"
	if len(wheres) > 0 {
		whereSQL = strings.Join(wheres, " AND ")
	}
	args = append(args, limit, offset)
	limitPos := len(args) - 1
	offsetPos := len(args)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
SELECT
  ife.id, ife.ref_id, ife.entry_type, ife.category, ife.direction, ife.bank_id, b.nama,
  ife.amount, ife.fee, ife.total_amount, ife.counterparty, ife.note, ife.occurred_at,
  ife.created_by, m.nama, ife.created_at
FROM public.internal_finance_entry ife
JOIN public.bank b ON b.id = ife.bank_id
LEFT JOIN public.member m ON m.id = ife.created_by
WHERE %s
ORDER BY ife.occurred_at DESC, ife.id DESC
LIMIT $%d OFFSET $%d
`, whereSQL, limitPos, offsetPos), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]InternalFinanceEntryRow, 0, limit)
	for rows.Next() {
		var item InternalFinanceEntryRow
		var createdBy sql.NullInt64
		var createdName sql.NullString
		if err := rows.Scan(
			&item.ID, &item.RefID, &item.EntryType, &item.Category, &item.Direction, &item.BankID, &item.BankNama,
			&item.Amount, &item.Fee, &item.TotalAmount, &item.Counterparty, &item.Note, &item.OccurredAt,
			&createdBy, &createdName, &item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		if createdBy.Valid {
			v := createdBy.Int64
			item.CreatedBy = &v
		}
		if createdName.Valid {
			v := createdName.String
			item.CreatedNama = &v
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	countArgs := args[:limitPos-1]
	var total int64
	if err := r.db.QueryRowContext(ctx, fmt.Sprintf(`
SELECT COUNT(1)
FROM public.internal_finance_entry ife
JOIN public.bank b ON b.id = ife.bank_id
WHERE %s
`, whereSQL), countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *AuditorRepository) UpsertOpeningBalances(ctx context.Context, actorID int64, periodMonth time.Time, items []OpeningBalanceUpsertInput) error {
	if r == nil || r.db == nil {
		return errors.New("db not initialized")
	}
	if periodMonth.IsZero() {
		return errors.New("period_month required")
	}
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	month := time.Date(periodMonth.Year(), periodMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	for _, item := range items {
		code := strings.TrimSpace(strings.ToUpper(item.AccountCode))
		if code == "" {
			return errors.New("account_code required")
		}
		name := strings.TrimSpace(item.AccountName)
		group := strings.TrimSpace(strings.ToLower(item.AccountGroup))
		if name == "" || group == "" {
			return errors.New("account_name/account_group required")
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO public.accounting_opening_balance
  (period_month, account_code, account_name, account_group, amount, note, created_by, created_at, updated_at)
VALUES
  ($1,$2,$3,$4,$5,$6,$7,now(),now())
ON CONFLICT (period_month, account_code)
DO UPDATE SET
  account_name = EXCLUDED.account_name,
  account_group = EXCLUDED.account_group,
  amount = EXCLUDED.amount,
  note = EXCLUDED.note,
  created_by = EXCLUDED.created_by,
  updated_at = now()
`, month, code, name, group, item.Amount, strings.TrimSpace(item.Note), actorID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *AuditorRepository) ListOpeningBalances(ctx context.Context, periodMonth time.Time) ([]OpeningBalanceRow, error) {
	month := time.Date(periodMonth.Year(), periodMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	rows, err := r.db.QueryContext(ctx, `
SELECT id, period_month, account_code, account_name, account_group, amount, note, created_by, created_at, updated_at
FROM public.accounting_opening_balance
WHERE period_month = $1
ORDER BY account_group ASC, account_code ASC
`, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]OpeningBalanceRow, 0, 32)
	for rows.Next() {
		var item OpeningBalanceRow
		var createdBy sql.NullInt64
		if err := rows.Scan(&item.ID, &item.PeriodMonth, &item.AccountCode, &item.AccountName, &item.AccountGroup, &item.Amount, &item.Note, &createdBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		if createdBy.Valid {
			v := createdBy.Int64
			item.CreatedBy = &v
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *AuditorRepository) ListTradingSummary(ctx context.Context, in AuditorTradingSummaryArgs) ([]AuditorTradingSummaryRow, error) {
	scope := normalizeAuditorScope(in.Scope)
	period := normalizeAuditorPeriod(in.Period)
	rows, err := r.db.QueryContext(ctx, `
WITH retail_rows AS (
  SELECT
    'retail'::text AS scope,
    o.dibuat_pada,
    COALESCE(o.harga_final, 0)::bigint AS harga_jual,
    COALESCE(rpc.harga_provider, 0)::bigint AS harga_beli,
    COALESCE(rc.amount, 0)::bigint AS komisi
  FROM public.app_order o
  LEFT JOIN (
    SELECT apt.app_order_id,
      COALESCE(SUM(CASE WHEN COALESCE(apt.harga_provider, 0) > 0 THEN apt.harga_provider ELSE 0 END), 0)::bigint AS harga_provider
    FROM public.app_order_provider_trx apt
    WHERE lower(COALESCE(apt.status, '')) = 'success'
    GROUP BY apt.app_order_id
  ) rpc ON rpc.app_order_id = o.id
  LEFT JOIN (
    SELECT source_app_order_id, SUM(amount)::bigint AS amount
    FROM public.retail_commission_ledger
    GROUP BY source_app_order_id
  ) rc ON rc.source_app_order_id = o.id
  WHERE lower(COALESCE(o.status,'')) = 'success'
    AND ($2 = false OR o.dibuat_pada >= $3)
    AND ($4 = false OR o.dibuat_pada < $5)
),
h2h_rows AS (
  SELECT
    'h2h'::text AS scope,
    tm.dibuat_pada,
    COALESCE(NULLIF(tm.biaya_aktual, 0), tm.harga_member, 0)::bigint AS harga_jual,
    COALESCE(hpc.harga_provider, 0)::bigint AS harga_beli,
    COALESCE(hc.amount, 0)::bigint AS komisi
  FROM public.transaksi_member tm
  LEFT JOIN (
    SELECT tp.transaksi_member_id,
      COALESCE(SUM(CASE WHEN COALESCE(tp.harga, 0) > 0 THEN tp.harga ELSE 0 END), 0)::bigint AS harga_provider
    FROM public.transaksi_provider tp
    WHERE lower(COALESCE(tp.status, '')) = 'success'
    GROUP BY tp.transaksi_member_id
  ) hpc ON hpc.transaksi_member_id = tm.id
  LEFT JOIN (
    SELECT source_trx_member_id, SUM(amount)::bigint AS amount
    FROM public.h2h_commission_ledger
    GROUP BY source_trx_member_id
  ) hc ON hc.source_trx_member_id = tm.id
  WHERE lower(COALESCE(tm.status,'')) = 'success'
    AND ($2 = false OR tm.dibuat_pada >= $3)
    AND ($4 = false OR tm.dibuat_pada < $5)
),
all_rows AS (
  SELECT * FROM retail_rows
  UNION ALL
  SELECT * FROM h2h_rows
)
SELECT
  CASE WHEN $1 = '' OR $1 = 'all' THEN 'all' ELSE scope END AS scope,
  CASE
    WHEN $6 = 'quarter' THEN to_char(date_trunc('month', dibuat_pada), 'YYYY-MM')
    WHEN $6 = 'month' THEN to_char(date_trunc('month', dibuat_pada), 'YYYY-MM')
    ELSE to_char(date_trunc('day', dibuat_pada), 'YYYY-MM-DD')
  END AS period_key,
  COUNT(*)::bigint AS transaction_count,
  COALESCE(SUM(harga_jual), 0)::bigint AS sales_amount,
  COALESCE(SUM(harga_beli), 0)::bigint AS provider_amount,
  COALESCE(SUM(komisi), 0)::bigint AS commission_amount,
  COALESCE(SUM(harga_jual - harga_beli - komisi), 0)::bigint AS margin_amount
FROM all_rows
WHERE ($1 = '' OR $1 = 'all' OR scope = $1)
GROUP BY 1, 2
ORDER BY period_key DESC
`, scope, in.HasFrom, in.From, in.HasTo, in.To, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]AuditorTradingSummaryRow, 0, 128)
	for rows.Next() {
		var item AuditorTradingSummaryRow
		if err := rows.Scan(&item.Scope, &item.PeriodKey, &item.TransactionCount, &item.SalesAmount, &item.ProviderAmount, &item.CommissionAmount, &item.MarginAmount); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *AuditorRepository) ListTradingDetails(ctx context.Context, in AuditorTradingDetailArgs) ([]AuditorTradingDetailRow, int64, error) {
	scope := normalizeAuditorScope(in.Scope)
	limit := in.Limit
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	if in.Offset < 0 {
		in.Offset = 0
	}
	rows, err := r.db.QueryContext(ctx, `
WITH retail_rows AS (
  SELECT
    'retail'::text AS scope,
    o.invoice_id AS ref_id,
    o.dibuat_pada AS occurred_at,
    COALESCE(o.status, '') AS status,
    o.member_id,
    m.nama AS member_nama,
    m.email AS member_email,
    COALESCE(o.produk_sku_snapshot, '') AS product_code,
    COALESCE(o.produk_nama_snapshot, COALESCE(o.produk_sku_snapshot, '')) AS product_name,
    COALESCE(o.dest, '') AS destination,
    picked.provider,
    COALESCE(picked.harga_provider, 0)::bigint AS harga_beli,
    COALESCE(o.harga_final, 0)::bigint AS harga_jual,
    COALESCE(rc.amount, 0)::bigint AS komisi,
    COALESCE(o.harga_final, 0)::bigint - COALESCE(picked.harga_provider, 0)::bigint - COALESCE(rc.amount, 0)::bigint AS margin,
    picked.sn AS provider_ref,
    COALESCE(picked.pesan, '') AS status_note
  FROM public.app_order o
  LEFT JOIN public.member m ON m.id = o.member_id
  LEFT JOIN (
    SELECT chosen.app_order_id, chosen.provider, chosen.harga_provider, chosen.sn, chosen.pesan
    FROM (
      SELECT apt.app_order_id, COALESCE(apt.provider, '') AS provider, COALESCE(apt.harga_provider, 0)::bigint AS harga_provider,
             NULLIF(apt.sn, '') AS sn, NULLIF(apt.pesan, '') AS pesan,
             ROW_NUMBER() OVER (
               PARTITION BY apt.app_order_id
               ORDER BY CASE WHEN lower(COALESCE(apt.status,''))='success' THEN 0 ELSE 1 END,
                        apt.diubah_pada DESC NULLS LAST, apt.dibuat_pada DESC NULLS LAST, apt.id DESC
             ) AS rn
      FROM public.app_order_provider_trx apt
    ) chosen
    WHERE chosen.rn = 1
  ) picked ON picked.app_order_id = o.id
  LEFT JOIN (
    SELECT source_app_order_id, SUM(amount)::bigint AS amount
    FROM public.retail_commission_ledger
    GROUP BY source_app_order_id
  ) rc ON rc.source_app_order_id = o.id
  WHERE ($2 = false OR o.dibuat_pada >= $3)
    AND ($4 = false OR o.dibuat_pada < $5)
    AND lower(COALESCE(o.status,'')) = 'success'
    AND ($1 = '' OR $1 = 'all' OR $1 = 'retail')
    AND ($6 = '' OR o.invoice_id ILIKE '%%' || $6 || '%%')
),
h2h_rows AS (
  SELECT
    'h2h'::text AS scope,
    tm.ref_id,
    tm.dibuat_pada AS occurred_at,
    COALESCE(tm.status, '') AS status,
    tm.member_id,
    m.nama AS member_nama,
    m.email AS member_email,
    COALESCE(tm.kode_produk, '') AS product_code,
    COALESCE(tm.kode_produk, '') AS product_name,
    COALESCE(tm.tujuan, '') AS destination,
    picked.provider,
    COALESCE(picked.harga_beli, 0)::bigint AS harga_beli,
    COALESCE(NULLIF(tm.biaya_aktual, 0), tm.harga_member, 0)::bigint AS harga_jual,
    COALESCE(hc.amount, 0)::bigint AS komisi,
    COALESCE(NULLIF(tm.biaya_aktual, 0), tm.harga_member, 0)::bigint - COALESCE(picked.harga_beli, 0)::bigint - COALESCE(hc.amount, 0)::bigint AS margin,
    picked.no_referensi AS provider_ref,
    COALESCE(picked.pesan, '') AS status_note
  FROM public.transaksi_member tm
  LEFT JOIN public.member m ON m.id = tm.member_id
  LEFT JOIN (
    SELECT chosen.transaksi_member_id, chosen.provider, chosen.harga_beli, chosen.no_referensi, chosen.pesan
    FROM (
      SELECT tp.transaksi_member_id, COALESCE(tp.provider, '') AS provider, COALESCE(tp.harga, 0)::bigint AS harga_beli,
             NULLIF(tp.no_referensi, '') AS no_referensi, NULLIF(tp.pesan, '') AS pesan,
             ROW_NUMBER() OVER (
               PARTITION BY tp.transaksi_member_id
               ORDER BY CASE WHEN lower(COALESCE(tp.status,''))='success' THEN 0 ELSE 1 END,
                        tp.id DESC
             ) AS rn
      FROM public.transaksi_provider tp
    ) chosen
    WHERE chosen.rn = 1
  ) picked ON picked.transaksi_member_id = tm.id
  LEFT JOIN (
    SELECT source_trx_member_id, SUM(amount)::bigint AS amount
    FROM public.h2h_commission_ledger
    GROUP BY source_trx_member_id
  ) hc ON hc.source_trx_member_id = tm.id
  WHERE ($2 = false OR tm.dibuat_pada >= $3)
    AND ($4 = false OR tm.dibuat_pada < $5)
    AND lower(COALESCE(tm.status,'')) = 'success'
    AND ($1 = '' OR $1 = 'all' OR $1 = 'h2h')
    AND ($6 = '' OR tm.ref_id ILIKE '%%' || $6 || '%%')
),
all_rows AS (
  SELECT * FROM retail_rows
  UNION ALL
  SELECT * FROM h2h_rows
)
SELECT
  scope, ref_id, occurred_at, status, member_id, member_nama, member_email,
  product_code, product_name, destination, provider, harga_beli, harga_jual, komisi, margin, provider_ref, status_note
FROM all_rows
ORDER BY occurred_at DESC, ref_id DESC
LIMIT $7 OFFSET $8
`, scope, in.HasFrom, in.From, in.HasTo, in.To, strings.TrimSpace(in.RefID), limit, in.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]AuditorTradingDetailRow, 0, limit)
	for rows.Next() {
		var item AuditorTradingDetailRow
		var memberID sql.NullInt64
		var memberNama, memberEmail, provider, providerRef, statusNote sql.NullString
		if err := rows.Scan(
			&item.Scope, &item.RefID, &item.OccurredAt, &item.Status, &memberID, &memberNama, &memberEmail,
			&item.ProductCode, &item.ProductName, &item.Destination, &provider, &item.HargaBeli, &item.HargaJual, &item.Komisi, &item.Margin, &providerRef, &statusNote,
		); err != nil {
			return nil, 0, err
		}
		if memberID.Valid {
			v := memberID.Int64
			item.MemberID = &v
		}
		if memberNama.Valid {
			v := memberNama.String
			item.MemberNama = &v
		}
		if memberEmail.Valid {
			v := memberEmail.String
			item.MemberEmail = &v
		}
		if provider.Valid && provider.String != "" {
			v := provider.String
			item.Provider = &v
		}
		if providerRef.Valid && providerRef.String != "" {
			v := providerRef.String
			item.ProviderRef = &v
		}
		if statusNote.Valid && statusNote.String != "" {
			v := statusNote.String
			item.ManualStatusNote = &v
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, `
WITH retail_rows AS (
  SELECT o.invoice_id AS ref_id, o.dibuat_pada
  FROM public.app_order o
  WHERE ($2 = false OR o.dibuat_pada >= $3)
    AND ($4 = false OR o.dibuat_pada < $5)
    AND lower(COALESCE(o.status,'')) = 'success'
    AND ($1 = '' OR $1 = 'all' OR $1 = 'retail')
    AND ($6 = '' OR o.invoice_id ILIKE '%%' || $6 || '%%')
),
h2h_rows AS (
  SELECT tm.ref_id, tm.dibuat_pada
  FROM public.transaksi_member tm
  WHERE ($2 = false OR tm.dibuat_pada >= $3)
    AND ($4 = false OR tm.dibuat_pada < $5)
    AND lower(COALESCE(tm.status,'')) = 'success'
    AND ($1 = '' OR $1 = 'all' OR $1 = 'h2h')
    AND ($6 = '' OR tm.ref_id ILIKE '%%' || $6 || '%%')
)
SELECT COUNT(1) FROM (
  SELECT ref_id FROM retail_rows
  UNION ALL
  SELECT ref_id FROM h2h_rows
) x
`, scope, in.HasFrom, in.From, in.HasTo, in.To, strings.TrimSpace(in.RefID)).Scan(&total); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func classifyFinanceType(reason string, meta string) string {
	reason = strings.TrimSpace(strings.ToUpper(reason))
	switch reason {
	case "MEMBER_DEPOSIT_APPROVE":
		return "member_deposit"
	case "BANK_TRANSFER_TO_PROVIDER":
		return "provider_deposit"
	case "H2H_WITHDRAW_APPROVE":
		return "h2h_withdraw"
	case "RETAIL_WITHDRAW_APPROVE":
		return "retail_withdraw"
	case "BANK_TRANSFER_OUT":
		return "bank_transfer_out"
	case "BANK_OPENING_BALANCE":
		return "opening_balance"
	case "BANK_ADJUST_CREDIT", "BANK_ADJUST_DEBIT":
		return "bank_adjustment"
	}
	if strings.HasPrefix(reason, "INTERNAL_FINANCE_") {
		var metaObj map[string]any
		if json.Unmarshal([]byte(meta), &metaObj) == nil {
			if entryType, ok := metaObj["entry_type"].(string); ok && strings.TrimSpace(entryType) != "" {
				return strings.TrimSpace(strings.ToLower(entryType))
			}
		}
		return "internal_finance"
	}
	return strings.ToLower(reason)
}

func (r *AuditorRepository) ListFinance(ctx context.Context, in AuditorFinanceArgs, includeAdminStaffOnly bool) ([]AuditorFinanceRow, int64, error) {
	if in.Limit <= 0 || in.Limit > 1000 {
		in.Limit = 100
	}
	if in.Offset < 0 {
		in.Offset = 0
	}
	args := make([]any, 0, 8)
	wheres := make([]string, 0, 8)
	if !includeAdminStaffOnly {
		wheres = append(wheres, "COALESCE(b.admin_staff_only, false) = false")
	}
	if in.BankID > 0 {
		args = append(args, in.BankID)
		wheres = append(wheres, fmt.Sprintf("mb.bank_id = $%d", len(args)))
	}
	if in.HasFrom {
		args = append(args, in.From)
		wheres = append(wheres, fmt.Sprintf("mb.dibuat_pada >= $%d", len(args)))
	}
	if in.HasTo {
		args = append(args, in.To)
		wheres = append(wheres, fmt.Sprintf("mb.dibuat_pada < $%d", len(args)))
	}
	whereSQL := "1=1"
	if len(wheres) > 0 {
		whereSQL = strings.Join(wheres, " AND ")
	}
	args = append(args, in.Limit, in.Offset)
	limitPos := len(args) - 1
	offsetPos := len(args)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
SELECT
  mb.id, mb.bank_id, b.nama, mb.ref_id, mb.arah, mb.jumlah, mb.alasan, COALESCE(mb.catatan,''), mb.member_id,
  COALESCE(target_member.nama,''), NULLIF(mb.provider,''), mb.dibuat_pada, COALESCE(ife.fee, 0), COALESCE(ife.total_amount, mb.jumlah), mb.saldo_sesudah, COALESCE(ife.counterparty,''), COALESCE(mb.meta::text, '{}')
FROM public.mutasi_bank mb
JOIN public.bank b ON b.id = mb.bank_id
LEFT JOIN public.member target_member ON target_member.id = mb.member_id
LEFT JOIN public.internal_finance_entry ife ON ife.ref_id = mb.ref_id
WHERE %s
ORDER BY mb.dibuat_pada DESC, mb.id DESC
LIMIT $%d OFFSET $%d
`, whereSQL, limitPos, offsetPos), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]AuditorFinanceRow, 0, in.Limit)
	for rows.Next() {
		var item AuditorFinanceRow
		var memberID sql.NullInt64
		var memberNama, provider, counterparty, metaText sql.NullString
		var reason string
		if err := rows.Scan(
			&item.ID, &item.BankID, &item.BankNama, &item.RefID, &item.Arah, &item.Amount, &reason, &item.Note, &memberID,
			&memberNama, &provider, &item.OccurredAt, &item.Fee, &item.TotalAmount, &item.SaldoBank, &counterparty, &metaText,
		); err != nil {
			return nil, 0, err
		}
		item.Reason = reason
		item.Type = classifyFinanceType(reason, coalesceNullString(metaText))
		if memberID.Valid {
			v := memberID.Int64
			item.MemberID = &v
		}
		if memberNama.Valid && memberNama.String != "" {
			v := memberNama.String
			item.MemberNama = &v
		}
		if provider.Valid && provider.String != "" {
			v := provider.String
			item.Provider = &v
		}
		if counterparty.Valid && counterparty.String != "" {
			v := counterparty.String
			item.Counterparty = &v
		} else if item.Provider != nil {
			item.Counterparty = item.Provider
		} else if item.MemberNama != nil {
			item.Counterparty = item.MemberNama
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	filtered := items
	if strings.TrimSpace(in.Type) != "" {
		needle := strings.TrimSpace(strings.ToLower(in.Type))
		filtered = make([]AuditorFinanceRow, 0, len(items))
		for _, item := range items {
			if item.Type == needle {
				filtered = append(filtered, item)
			}
		}
	}

	countArgs := args[:limitPos-1]
	var total int64
	if err := r.db.QueryRowContext(ctx, fmt.Sprintf(`
SELECT COUNT(1)
FROM public.mutasi_bank mb
JOIN public.bank b ON b.id = mb.bank_id
WHERE %s
`, whereSQL), countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if strings.TrimSpace(in.Type) != "" {
		total = int64(len(filtered))
	}
	return filtered, total, nil
}

func coalesceNullString(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

func (r *AuditorRepository) GetProfitLoss(ctx context.Context, in AuditorProfitLossArgs, includeAdminStaffOnly bool) (*AuditorProfitLossReport, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("db not initialized")
	}
	loc, _ := time.LoadLocation("Asia/Jakarta")
	monthStart := time.Date(in.Month.Year(), in.Month.Month(), 1, 0, 0, 0, 0, loc)
	if monthStart.IsZero() {
		now := time.Now().In(loc)
		monthStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	}
	monthEnd := monthStart.AddDate(0, 1, 0)
	yearStart := time.Date(monthStart.Year(), 1, 1, 0, 0, 0, 0, loc)

	summaryRows, err := r.ListTradingSummary(ctx, AuditorTradingSummaryArgs{
		Scope:   "all",
		Period:  "month",
		From:    monthStart,
		HasFrom: true,
		To:      monthEnd,
		HasTo:   true,
	})
	if err != nil {
		return nil, err
	}
	var revenue, hpp, commission, margin int64
	for _, row := range summaryRows {
		revenue += row.SalesAmount
		hpp += row.ProviderAmount
		commission += row.CommissionAmount
		margin += row.MarginAmount
	}

	var currentYearProfit int64
	yearRows, err := r.ListTradingSummary(ctx, AuditorTradingSummaryArgs{
		Scope:   "all",
		Period:  "month",
		From:    yearStart,
		HasFrom: true,
		To:      monthEnd,
		HasTo:   true,
	})
	if err != nil {
		return nil, err
	}
	for _, row := range yearRows {
		currentYearProfit += row.MarginAmount
	}

	internalEntries, _, err := r.ListInternalFinanceEntries(ctx, 0, "", monthStart.Format("2006-01-02"), monthEnd.AddDate(0, 0, -1).Format("2006-01-02"), 1000, 0, includeAdminStaffOnly)
	if err != nil {
		return nil, err
	}
	expenseMap := map[string]int64{}
	otherIncome := int64(0)
	otherExpense := int64(0)
	for _, item := range internalEntries {
		switch item.EntryType {
		case "other_income", "bank_interest":
			otherIncome += item.TotalAmount
		case "bank_interest_tax", "bank_admin":
			otherExpense += item.TotalAmount
		default:
			expenseMap[item.EntryType] += item.TotalAmount
		}
	}

	openingBalances, err := r.ListOpeningBalances(ctx, monthStart)
	if err != nil {
		return nil, err
	}

	assetLines, totalAsset, err := r.assetLinesAt(ctx, monthEnd, includeAdminStaffOnly)
	if err != nil {
		return nil, err
	}
	liabilityLines, totalLiability, err := r.liabilityLinesAt(ctx, monthEnd)
	if err != nil {
		return nil, err
	}
	equityLines := make([]AuditorBalanceLine, 0, 8)
	totalEquity := int64(0)
	for _, item := range openingBalances {
		if item.AccountGroup != "equity" {
			continue
		}
		equityLines = append(equityLines, AuditorBalanceLine{
			Code:   item.AccountCode,
			Label:  item.AccountName,
			Amount: item.Amount,
		})
		totalEquity += item.Amount
	}
	equityLines = append(equityLines, AuditorBalanceLine{Code: "PROFIT_CURRENT_YEAR", Label: "Profit (Loss) Current Year", Amount: currentYearProfit})
	totalEquity += currentYearProfit

	expenseLines := []AuditorProfitLossLine{
		{Code: "COMMISSION", Label: "Commission Expense", Amount: commission},
	}
	expenseOrder := []struct {
		key   string
		label string
	}{
		{"salary", "Salary Expense"},
		{"operational_expense", "Operasional Expense"},
		{"rent_expense", "Rent Expense"},
		{"audit_expense", "Audit Expense"},
		{"printing_expense", "Printing Expense"},
		{"event_expense", "Event Organizer Expense"},
		{"tax_income_expense", "Tax Income Expense"},
		{"purchase", "Purchase Expense"},
		{"other_expense", "Other Expense"},
	}
	for _, item := range expenseOrder {
		if expenseMap[item.key] == 0 {
			continue
		}
		expenseLines = append(expenseLines, AuditorProfitLossLine{Code: strings.ToUpper(item.key), Label: item.label, Amount: expenseMap[item.key]})
	}
	var totalExpense int64
	for _, item := range expenseLines {
		totalExpense += item.Amount
	}
	netProfit := revenue - hpp - totalExpense + otherIncome - otherExpense

	return &AuditorProfitLossReport{
		Month:          monthStart.Format("2006-01"),
		GeneratedAtWIB: time.Now().In(loc),
		RevenueLines: []AuditorProfitLossLine{
			{Code: "OPERATING_REVENUE", Label: "Penghasilan Usaha", Amount: revenue},
		},
		CostLines: []AuditorProfitLossLine{
			{Code: "COGS", Label: "Harga Pokok Penjualan", Amount: hpp},
			{Code: "GROSS_PROFIT", Label: "Laba Kotor", Amount: margin},
		},
		ExpenseLines: expenseLines,
		OtherIncomeLines: []AuditorProfitLossLine{
			{Code: "BANK_INTEREST", Label: "Bank Interest Income", Amount: otherIncome},
		},
		OtherExpenseLines: []AuditorProfitLossLine{
			{Code: "BANK_ADMIN_OTHER", Label: "Other Expense", Amount: otherExpense},
		},
		NetProfit:            netProfit,
		CurrentYearProfit:    currentYearProfit,
		AssetLines:           assetLines,
		LiabilityLines:       liabilityLines,
		EquityLines:          equityLines,
		TotalAsset:           totalAsset,
		TotalLiability:       totalLiability,
		TotalEquity:          totalEquity,
		TotalLiabilityEquity: totalLiability + totalEquity,
		OpeningBalances:      openingBalances,
	}, nil
}

func (r *AuditorRepository) assetLinesAt(ctx context.Context, cutoff time.Time, includeAdminStaffOnly bool) ([]AuditorBalanceLine, int64, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT b.id, b.nama,
  COALESCE((
    SELECT mb.saldo_sesudah
    FROM public.mutasi_bank mb
    WHERE mb.bank_id = b.id
      AND mb.dibuat_pada < $1
    ORDER BY mb.dibuat_pada DESC, mb.id DESC
    LIMIT 1
  ), b.saldo, 0)::bigint AS saldo_akhir
FROM public.bank b
WHERE b.aktif = true
  AND ($2 = true OR COALESCE(b.admin_staff_only, false) = false)
ORDER BY b.id ASC
`, cutoff, includeAdminStaffOnly)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	lines := make([]AuditorBalanceLine, 0, 16)
	total := int64(0)
	for rows.Next() {
		var id int64
		var name string
		var amount int64
		if err := rows.Scan(&id, &name, &amount); err != nil {
			return nil, 0, err
		}
		lines = append(lines, AuditorBalanceLine{Code: fmt.Sprintf("BANK_%d", id), Label: name, Amount: amount})
		total += amount
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var providerTotal int64
	if err := r.db.QueryRowContext(ctx, `
SELECT COALESCE(SUM(COALESCE((
  SELECT mdp.saldo_sesudah
  FROM public.mutasi_dompet_provider mdp
  WHERE mdp.provider = dp.provider
    AND mdp.dibuat_pada < $1
  ORDER BY mdp.dibuat_pada DESC, mdp.id DESC
  LIMIT 1
), dp.saldo, 0)), 0)::bigint
FROM public.dompet_provider dp
`, cutoff).Scan(&providerTotal); err != nil {
		return nil, 0, err
	}
	lines = append(lines, AuditorBalanceLine{Code: "TRADE_RECEIVABLE", Label: "Trade Receivable", Amount: providerTotal})
	total += providerTotal
	return lines, total, nil
}

func (r *AuditorRepository) liabilityLinesAt(ctx context.Context, cutoff time.Time) ([]AuditorBalanceLine, int64, error) {
	var memberLiability int64
	if err := r.db.QueryRowContext(ctx, `
SELECT COALESCE(SUM(COALESCE((
  SELECT md.saldo_sesudah
  FROM public.mutasi_dompet md
  WHERE md.member_id = dm.member_id
    AND md.dibuat_pada < $1
  ORDER BY md.dibuat_pada DESC, md.id DESC
  LIMIT 1
), dm.saldo, 0)), 0)::bigint
FROM public.dompet_member dm
`, cutoff).Scan(&memberLiability); err != nil {
		return nil, 0, err
	}
	lines := []AuditorBalanceLine{
		{Code: "MEMBER_LIABILITY", Label: "Saldo Member / Liabilitas", Amount: memberLiability},
	}
	return lines, memberLiability, nil
}
