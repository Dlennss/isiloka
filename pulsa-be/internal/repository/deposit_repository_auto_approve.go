package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func (r *DepositRepository) AutoApprovePendingFromBankMutations(ctx context.Context, limit int) (int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	approved := 0
	for approved < limit {
		ok, err := r.autoApproveOnePendingFromBankMutation(ctx)
		if err != nil {
			return approved, err
		}
		if !ok {
			break
		}
		approved++
	}
	return approved, nil
}

func (r *DepositRepository) autoApproveOnePendingFromBankMutation(ctx context.Context) (bool, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var (
		depositID int64
		memberID  int64
		bankID    int64
		amount    int64
		mutasiID  int64
		refID     string
	)
	err = tx.QueryRowContext(ctx, `
WITH candidate_deposit AS (
  SELECT d.id, d.member_id, d.bank_id, d.amount, d.dibuat_pada
  FROM public.deposit_request d
  WHERE d.status = 'pending'
    AND d.bank_id IS NOT NULL
    AND d.bank_id > 0
    AND d.amount > 0
    AND LOWER(TRIM(COALESCE(d.metode, ''))) <> 'qris'
    AND EXISTS (
      SELECT 1
      FROM public.mutasi_bank mb
      WHERE mb.bank_id = d.bank_id
        AND mb.jumlah = d.amount
        AND mb.arah = 'CREDIT'
        AND mb.alasan = 'BANK_MANUAL_IN'
        AND mb.member_id IS NULL
        AND mb.provider IS NULL
        AND mb.dibuat_pada >= d.dibuat_pada - INTERVAL '5 minutes'
        AND NOT (COALESCE(mb.meta, '{}'::jsonb) ? 'auto_deposit_request_id')
    )
  ORDER BY d.dibuat_pada ASC, d.id ASC
  LIMIT 1
  FOR UPDATE SKIP LOCKED
),
candidate_mutasi AS (
  SELECT mb.id, COALESCE(NULLIF(TRIM(mb.ref_id), ''), 'BMIN-' || mb.id::text) AS ref_id
  FROM public.mutasi_bank mb
  JOIN candidate_deposit d ON d.bank_id = mb.bank_id AND d.amount = mb.jumlah
  WHERE mb.arah = 'CREDIT'
    AND mb.alasan = 'BANK_MANUAL_IN'
    AND mb.member_id IS NULL
    AND mb.provider IS NULL
    AND mb.dibuat_pada >= d.dibuat_pada - INTERVAL '5 minutes'
    AND NOT (COALESCE(mb.meta, '{}'::jsonb) ? 'auto_deposit_request_id')
  ORDER BY mb.dibuat_pada ASC, mb.id ASC
  LIMIT 1
  FOR UPDATE SKIP LOCKED
)
SELECT d.id, d.member_id, d.bank_id, d.amount, m.id, m.ref_id
FROM candidate_deposit d
JOIN candidate_mutasi m ON true
`).Scan(&depositID, &memberID, &bankID, &amount, &mutasiID, &refID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	refID = strings.TrimSpace(refID)
	if refID == "" {
		return false, errors.New("mutasi bank ref_id kosong")
	}
	note := fmt.Sprintf("Auto approve dari mutasi bank %s", refID)

	if err := r.creditWalletTx(ctx, tx, memberID, amount, "deposit approve", note, refID, 0); err != nil {
		return false, err
	}

	res, err := tx.ExecContext(ctx, `
UPDATE public.mutasi_bank
SET member_id = $2::bigint,
    catatan = CASE
      WHEN NULLIF(TRIM(COALESCE(catatan, '')), '') IS NULL THEN $4
      ELSE catatan || ' | ' || $4
    END,
    meta = COALESCE(meta, '{}'::jsonb) || jsonb_build_object(
      'auto_approve', true,
      'auto_deposit_request_id', $1::bigint,
      'auto_deposit_member_id', $2::bigint,
      'auto_deposit_amount', $3::bigint,
      'auto_deposit_ref_id', $7::text,
      'auto_approved_at', now()
    )
WHERE id = $5::bigint
  AND bank_id = $6::bigint
  AND jumlah = $3::bigint
  AND arah = 'CREDIT'
  AND alasan = 'BANK_MANUAL_IN'
  AND member_id IS NULL
  AND provider IS NULL
`, depositID, memberID, amount, note, mutasiID, bankID, refID)
	if err != nil {
		return false, err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return false, errors.New("mutasi bank sudah diproses")
	}

	res, err = tx.ExecContext(ctx, `
UPDATE public.deposit_request
SET ref_id = $2::text,
    status = 'approved',
    note = CASE
      WHEN NULLIF(TRIM(COALESCE(note, '')), '') IS NULL THEN $6
      ELSE note || ' | ' || $6
    END,
    approved_amount = $3::bigint,
    diproses_pada = now(),
    diproses_oleh = NULL
WHERE id = $1::bigint
  AND member_id = $4::bigint
  AND bank_id = $5::bigint
  AND amount = $3::bigint
  AND status = 'pending'
`, depositID, refID, amount, memberID, bankID, note)
	if err != nil {
		return false, err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return false, errors.New("deposit pending sudah diproses")
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}
