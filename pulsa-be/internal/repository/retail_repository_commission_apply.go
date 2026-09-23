package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"pulsa2/internal/helper"
)

func (r *RetailRepository) ApplyCommissionForOrder(ctx context.Context, orderID int64) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var (
		invoiceID      string
		sourceMemberID sql.NullInt64
		buyerType      string
		buyerRole      string
		status         string
		retailAgentID  sql.NullInt64
		retailMasterID sql.NullInt64
	)
	if err := tx.QueryRowContext(ctx, `
SELECT
  invoice_id, member_id, buyer_type, COALESCE(buyer_role, ''), status,
  retail_agent_id_snapshot, retail_master_id_snapshot
FROM public.app_order
WHERE id = $1
FOR UPDATE
`, orderID).Scan(&invoiceID, &sourceMemberID, &buyerType, &buyerRole, &status, &retailAgentID, &retailMasterID); err != nil {
		return err
	}
	if strings.TrimSpace(strings.ToLower(status)) != "success" || strings.TrimSpace(strings.ToLower(buyerType)) != "user" || !sourceMemberID.Valid {
		return tx.Commit()
	}

	type commissionTarget struct {
		memberID int64
		level    string
		amount   int64
	}
	targets := make([]commissionTarget, 0, 2)

	getCommission := func(targetID int64, level string) (int64, error) {
		var agentCommission, masterCommission int64
		err := tx.QueryRowContext(ctx, `
SELECT COALESCE(retail_agent_commission_rp, 0), COALESCE(retail_master_commission_rp, 0)
FROM public.member
WHERE id = $1
LIMIT 1
`, targetID).Scan(&agentCommission, &masterCommission)
		if err != nil {
			return 0, err
		}
		switch level {
		case "agent":
			return agentCommission, nil
		case "master":
			return masterCommission, nil
		default:
			return 0, nil
		}
	}

	switch strings.TrimSpace(strings.ToLower(buyerRole)) {
	case "user":
		if retailAgentID.Valid {
			amt, feeErr := getCommission(retailAgentID.Int64, "agent")
			if feeErr != nil {
				return feeErr
			}
			if amt > 0 {
				targets = append(targets, commissionTarget{memberID: retailAgentID.Int64, level: "agent", amount: amt})
			}
			if retailMasterID.Valid {
				amt, feeErr := getCommission(retailMasterID.Int64, "master")
				if feeErr != nil {
					return feeErr
				}
				if amt > 0 {
					targets = append(targets, commissionTarget{memberID: retailMasterID.Int64, level: "master", amount: amt})
				}
			}
		} else if retailMasterID.Valid {
			amt, feeErr := getCommission(retailMasterID.Int64, "master")
			if feeErr != nil {
				return feeErr
			}
			if amt > 0 {
				targets = append(targets, commissionTarget{memberID: retailMasterID.Int64, level: "master", amount: amt})
			}
		}
	case "agent":
		if retailMasterID.Valid {
			amt, feeErr := getCommission(retailMasterID.Int64, "master")
			if feeErr != nil {
				return feeErr
			}
			if amt > 0 {
				targets = append(targets, commissionTarget{memberID: retailMasterID.Int64, level: "master", amount: amt})
			}
		}
	}

	for _, target := range targets {
		if target.memberID <= 0 || target.amount <= 0 {
			continue
		}
		res, err := tx.ExecContext(ctx, `
INSERT INTO public.retail_commission_ledger
  (member_id, source_member_id, source_app_order_id, invoice_id, level_name, amount, note, created_at)
VALUES
  ($1,$2,$3,$4,$5,$6,$7,now())
ON CONFLICT (member_id, source_app_order_id) DO NOTHING
`, target.memberID, sourceMemberID.Int64, orderID, invoiceID, target.level, target.amount, "retail commission auto credit")
		if err != nil {
			return err
		}
		aff, err := res.RowsAffected()
		if err != nil || aff == 0 {
			continue
		}

		var before int64
		if err := tx.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_member
WHERE member_id = $1
FOR UPDATE
`, target.memberID).Scan(&before); err != nil {
			return err
		}
		after := before + target.amount
		if _, err := tx.ExecContext(ctx, `
UPDATE public.dompet_member
SET saldo = $2, diperbarui_pada = now()
WHERE member_id = $1
`, target.memberID, after); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
VALUES
  ($1,$2,'CREDIT',$3,'RETAIL_COMMISSION',NULLIF($4,''),$5,$6,now())
`, target.memberID, invoiceID, target.amount, "komisi retail "+target.level, before, after); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *RetailRepository) ReverseCommissionForOrder(ctx context.Context, orderID int64, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "retail commission reversed because app order was refunded"
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var invoiceID string
	if err := tx.QueryRowContext(ctx, `
SELECT invoice_id
FROM public.app_order
WHERE id = $1
FOR UPDATE
`, orderID).Scan(&invoiceID); err != nil {
		return err
	}

	type commissionRow struct {
		id       int64
		memberID int64
		amount   int64
		level    string
	}
	rows, err := tx.QueryContext(ctx, `
SELECT id, member_id, amount, COALESCE(level_name, '')
FROM public.retail_commission_ledger
WHERE source_app_order_id = $1
  AND amount > 0
FOR UPDATE
`, orderID)
	if err != nil {
		return err
	}

	var items []commissionRow
	for rows.Next() {
		var item commissionRow
		if err := rows.Scan(&item.id, &item.memberID, &item.amount, &item.level); err != nil {
			_ = rows.Close()
			return err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}

	for _, item := range items {
		if item.memberID <= 0 || item.amount <= 0 {
			continue
		}

		var alreadyReversed bool
		if err := tx.QueryRowContext(ctx, `
SELECT EXISTS (
  SELECT 1
  FROM public.mutasi_dompet
  WHERE member_id = $1
    AND ref_id = $2
    AND LOWER(COALESCE(arah, '')) = 'debit'
    AND LOWER(COALESCE(alasan, '')) = 'retail_commission_reversal'
)
`, item.memberID, invoiceID).Scan(&alreadyReversed); err != nil {
			return err
		}

		if !alreadyReversed {
			var before int64
			if err := tx.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_member
WHERE member_id = $1
FOR UPDATE
`, item.memberID).Scan(&before); err != nil {
				return err
			}
			if before < item.amount {
				return fmt.Errorf("saldo komisi retail tidak cukup untuk reversal member_id=%d invoice=%s amount=%d saldo=%d", item.memberID, invoiceID, item.amount, before)
			}
			after := before - item.amount
			if _, err := tx.ExecContext(ctx, `
UPDATE public.dompet_member
SET saldo = $2, diperbarui_pada = now()
WHERE member_id = $1
`, item.memberID, after); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
VALUES
  ($1,$2,'DEBIT',$3,'RETAIL_COMMISSION_REVERSAL',NULLIF($4,''),$5,$6,now())
`, item.memberID, invoiceID, item.amount, "reversal komisi retail "+strings.TrimSpace(item.level)+": "+reason, before, after); err != nil {
				return err
			}
		}

		if _, err := tx.ExecContext(ctx, `
UPDATE public.retail_commission_ledger
SET amount = 0,
    note = CASE
      WHEN BTRIM(COALESCE(note, '')) = '' THEN $2
      ELSE BTRIM(note) || '; ' || $2
    END
WHERE id = $1
`, item.id, "reversed: "+reason); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *RetailRepository) ApplyHistoricalCommission(ctx context.Context, sourceMemberID int64, buyerRole string, retailAgentID, retailMasterID *int64) (agentCount int64, agentTotal int64, masterCount int64, masterTotal int64, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer func() { _ = tx.Rollback() }()

	agentCount, agentTotal, masterCount, masterTotal, err = applyHistoricalCommissionTx(ctx, tx, sourceMemberID, buyerRole, retailAgentID, retailMasterID)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, 0, 0, 0, err
	}
	return agentCount, agentTotal, masterCount, masterTotal, nil
}

func (r *RetailRepository) ApplyHierarchyAssignment(ctx context.Context, memberID int64, role string, retailAgentID, retailMasterID *int64, applyHistorical bool) (agentCount int64, agentTotal int64, masterCount int64, masterTotal int64, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
UPDATE public.member
SET role = $2,
    retail_agent_id = $3,
    retail_master_id = $4
WHERE id = $1
`, memberID, role, retailNullableInt64(retailAgentID), retailNullableInt64(retailMasterID))
	if err != nil {
		return 0, 0, 0, 0, err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return 0, 0, 0, 0, sql.ErrNoRows
	}

	if applyHistorical {
		agentCount, agentTotal, masterCount, masterTotal, err = applyHistoricalCommissionTx(ctx, tx, memberID, role, retailAgentID, retailMasterID)
		if err != nil {
			return 0, 0, 0, 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, 0, 0, err
	}
	return agentCount, agentTotal, masterCount, masterTotal, nil
}

type retailHistoricalCommissionTarget struct {
	memberID int64
	level    string
	amount   int64
}

func applyHistoricalCommissionTx(ctx context.Context, tx *sql.Tx, sourceMemberID int64, buyerRole string, retailAgentID, retailMasterID *int64) (agentCount int64, agentTotal int64, masterCount int64, masterTotal int64, err error) {
	targets := make([]retailHistoricalCommissionTarget, 0, 2)

	getTargetCommission := func(targetID int64, level string) (int64, error) {
		var agentCommission, masterCommission int64
		if err := tx.QueryRowContext(ctx, `
SELECT COALESCE(retail_agent_commission_rp, 0), COALESCE(retail_master_commission_rp, 0)
FROM public.member
WHERE id = $1
LIMIT 1
`, targetID).Scan(&agentCommission, &masterCommission); err != nil {
			return 0, err
		}
		if level == "agent" {
			return agentCommission, nil
		}
		return masterCommission, nil
	}

	switch strings.TrimSpace(strings.ToLower(buyerRole)) {
	case "user":
		if retailAgentID != nil && *retailAgentID > 0 {
			amount, feeErr := getTargetCommission(*retailAgentID, "agent")
			if feeErr != nil {
				return 0, 0, 0, 0, feeErr
			}
			if amount > 0 {
				targets = append(targets, retailHistoricalCommissionTarget{memberID: *retailAgentID, level: "agent", amount: amount})
			}
		}
		if retailMasterID != nil && *retailMasterID > 0 {
			amount, feeErr := getTargetCommission(*retailMasterID, "master")
			if feeErr != nil {
				return 0, 0, 0, 0, feeErr
			}
			if amount > 0 {
				targets = append(targets, retailHistoricalCommissionTarget{memberID: *retailMasterID, level: "master", amount: amount})
			}
		}
	case "agent":
		if retailMasterID != nil && *retailMasterID > 0 {
			amount, feeErr := getTargetCommission(*retailMasterID, "master")
			if feeErr != nil {
				return 0, 0, 0, 0, feeErr
			}
			if amount > 0 {
				targets = append(targets, retailHistoricalCommissionTarget{memberID: *retailMasterID, level: "master", amount: amount})
			}
		}
	}

	if err := reverseObsoleteHistoricalRetailCommissionsTx(ctx, tx, sourceMemberID, targets); err != nil {
		return 0, 0, 0, 0, err
	}

	applyTarget := func(targetID int64, level string, amount int64) (int64, int64, error) {
		if targetID <= 0 || amount <= 0 {
			return 0, 0, nil
		}

		note := "retail commission historical backfill"
		var count, total int64
		if err := tx.QueryRowContext(ctx, `
WITH inserted AS (
  INSERT INTO public.retail_commission_ledger
    (member_id, source_member_id, source_app_order_id, invoice_id, level_name, amount, note, created_at)
  SELECT
    $1, o.member_id, o.id, o.invoice_id, $2, $3, $4, now()
  FROM public.app_order o
  WHERE o.member_id = $5
    AND lower(COALESCE(o.buyer_type, '')) = 'user'
    AND lower(COALESCE(o.status, '')) = 'success'
  ON CONFLICT (member_id, source_app_order_id) DO UPDATE
    SET amount = EXCLUDED.amount,
        level_name = EXCLUDED.level_name,
        note = CASE
          WHEN BTRIM(COALESCE(public.retail_commission_ledger.note, '')) = '' THEN EXCLUDED.note
          ELSE BTRIM(public.retail_commission_ledger.note) || '; ' || EXCLUDED.note
        END
    WHERE public.retail_commission_ledger.amount = 0
  RETURNING amount
)
SELECT COUNT(*), COALESCE(SUM(amount), 0)
FROM inserted
`, targetID, level, amount, note, sourceMemberID).Scan(&count, &total); err != nil {
			return 0, 0, err
		}
		if total <= 0 {
			return count, total, nil
		}

		var before int64
		if err := tx.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_member WHERE member_id = $1 FOR UPDATE`, targetID).Scan(&before); err != nil {
			return 0, 0, err
		}
		after := before + total
		if _, err := tx.ExecContext(ctx, `UPDATE public.dompet_member SET saldo = $2, diperbarui_pada = now() WHERE member_id = $1`, targetID, after); err != nil {
			return 0, 0, err
		}

		refID := fmt.Sprintf("RCB-%s-%s", time.Now().Format("20060102150405"), strings.ToUpper(helper.RandHex(4)))
		if _, err := tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
VALUES
  ($1,$2,'CREDIT',$3,'RETAIL_COMMISSION_BACKFILL',NULLIF($4,''),$5,$6,now())
`, targetID, refID, total, fmt.Sprintf("backfill komisi retail %s dari member %d", level, sourceMemberID), before, after); err != nil {
			return 0, 0, err
		}
		return count, total, nil
	}

	for _, target := range targets {
		count, total, applyErr := applyTarget(target.memberID, target.level, target.amount)
		if applyErr != nil {
			return 0, 0, 0, 0, applyErr
		}
		switch target.level {
		case "agent":
			agentCount += count
			agentTotal += total
		case "master":
			masterCount += count
			masterTotal += total
		}
	}

	return agentCount, agentTotal, masterCount, masterTotal, nil
}

func reverseObsoleteHistoricalRetailCommissionsTx(ctx context.Context, tx *sql.Tx, sourceMemberID int64, targets []retailHistoricalCommissionTarget) error {
	isDesired := func(memberID int64, level string) bool {
		level = strings.TrimSpace(strings.ToLower(level))
		for _, target := range targets {
			if target.memberID == memberID && target.level == level {
				return true
			}
		}
		return false
	}

	rows, err := tx.QueryContext(ctx, `
SELECT
  rcl.id,
  rcl.member_id,
  rcl.invoice_id,
  COALESCE(rcl.level_name, ''),
  rcl.amount
FROM public.retail_commission_ledger rcl
JOIN public.app_order o ON o.id = rcl.source_app_order_id
WHERE rcl.source_member_id = $1
  AND o.member_id = $1
  AND lower(COALESCE(o.buyer_type, '')) = 'user'
  AND lower(COALESCE(o.status, '')) = 'success'
  AND rcl.amount > 0
FOR UPDATE
`, sourceMemberID)
	if err != nil {
		return err
	}
	defer rows.Close()

	type obsoleteCommission struct {
		id       int64
		memberID int64
		invoice  string
		level    string
		amount   int64
	}
	items := make([]obsoleteCommission, 0)
	for rows.Next() {
		var item obsoleteCommission
		if err := rows.Scan(&item.id, &item.memberID, &item.invoice, &item.level, &item.amount); err != nil {
			return err
		}
		if isDesired(item.memberID, item.level) {
			continue
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}

	for _, item := range items {
		var before int64
		if err := tx.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_member
WHERE member_id = $1
FOR UPDATE
`, item.memberID).Scan(&before); err != nil {
			return err
		}
		if before < item.amount {
			return fmt.Errorf("saldo komisi retail tidak cukup untuk pindah jaringan member_id=%d invoice=%s amount=%d saldo=%d", item.memberID, item.invoice, item.amount, before)
		}
		after := before - item.amount
		if _, err := tx.ExecContext(ctx, `
UPDATE public.dompet_member
SET saldo = $2, diperbarui_pada = now()
WHERE member_id = $1
`, item.memberID, after); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
VALUES
  ($1,$2,'DEBIT',$3,'RETAIL_COMMISSION_REASSIGN_REVERSAL',NULLIF($4,''),$5,$6,now())
`, item.memberID, item.invoice, item.amount, "reversal komisi retail "+strings.TrimSpace(item.level)+" karena pindah jaringan", before, after); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
UPDATE public.retail_commission_ledger
SET amount = 0,
    note = CASE
      WHEN BTRIM(COALESCE(note, '')) = '' THEN $2
      ELSE BTRIM(note) || '; ' || $2
    END
WHERE id = $1
`, item.id, "reversed by retail hierarchy reassignment"); err != nil {
			return err
		}
	}
	return nil
}
