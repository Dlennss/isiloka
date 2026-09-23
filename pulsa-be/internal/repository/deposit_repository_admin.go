package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

type depositBankMutationForApprove struct {
	ID     int64
	RefID  string
	Amount int64
}

func normalizeDepositBankRefIDs(refIDs []string) ([]string, error) {
	out := make([]string, 0, len(refIDs))
	seen := map[string]struct{}{}
	for _, refID := range refIDs {
		refID = strings.TrimSpace(refID)
		if refID == "" {
			continue
		}
		if _, exists := seen[refID]; exists {
			return nil, fmt.Errorf("refid mutasi duplikat: %s", refID)
		}
		seen[refID] = struct{}{}
		out = append(out, refID)
	}
	if len(out) > 4 {
		return nil, errors.New("maksimal 4 refid mutasi bank")
	}
	return out, nil
}

func selectDepositBankMutationsByRefIDs(ctx context.Context, tx *sql.Tx, bankID int64, refIDs []string) ([]depositBankMutationForApprove, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT id, TRIM(COALESCE(ref_id, '')) AS ref_id, jumlah
FROM public.mutasi_bank
WHERE bank_id = $1::bigint
  AND TRIM(COALESCE(ref_id, '')) = ANY($2::text[])
  AND arah = 'CREDIT'
  AND alasan = 'BANK_MANUAL_IN'
  AND member_id IS NULL
  AND provider IS NULL
  AND NOT (COALESCE(meta, '{}'::jsonb) ? 'auto_deposit_request_id')
  AND NOT (COALESCE(meta, '{}'::jsonb) ? 'manual_deposit_request_id')
ORDER BY array_position($2::text[], TRIM(COALESCE(ref_id, ''))), id ASC
FOR UPDATE SKIP LOCKED
`, bankID, pq.Array(refIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]depositBankMutationForApprove, 0, len(refIDs))
	found := map[string]struct{}{}
	for rows.Next() {
		var row depositBankMutationForApprove
		if err := rows.Scan(&row.ID, &row.RefID, &row.Amount); err != nil {
			return nil, err
		}
		if row.RefID == "" || row.Amount <= 0 {
			continue
		}
		out = append(out, row)
		found[row.RefID] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(out) != len(refIDs) {
		missing := make([]string, 0, len(refIDs))
		for _, refID := range refIDs {
			if _, ok := found[refID]; !ok {
				missing = append(missing, refID)
			}
		}
		return nil, fmt.Errorf("refid mutasi belum ada/belum bisa dipakai: %s", strings.Join(missing, ", "))
	}
	return out, nil
}

func (r *DepositRepository) Approve(ctx context.Context, reqID, adminID, approvedAmount int64, note, fallbackRefID string, bankRefIDs []string) (string, int64, error) {
	if reqID <= 0 || adminID <= 0 {
		return "", 0, errors.New("invalid req_id/admin_id")
	}
	fallbackRefID = strings.TrimSpace(fallbackRefID)
	if fallbackRefID == "" {
		return "", 0, errors.New("ref_id required")
	}
	selectedRefIDs, err := normalizeDepositBankRefIDs(bankRefIDs)
	if err != nil {
		return "", 0, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", 0, err
	}
	defer tx.Rollback()

	var memberID, amount, requestedAmount int64
	var bankID sql.NullInt64
	var status, metode, memberName string
	err = tx.QueryRowContext(ctx, `
SELECT d.member_id, d.bank_id, d.amount, COALESCE(d.requested_amount, 0), d.status, COALESCE(d.metode, ''), COALESCE(m.nama, '')
FROM public.deposit_request d
LEFT JOIN public.member m ON m.id = d.member_id
WHERE d.id = $1
FOR UPDATE OF d
`, reqID).Scan(&memberID, &bankID, &amount, &requestedAmount, &status, &metode, &memberName)
	if err != nil {
		return "", 0, err
	}
	if status != "pending" {
		return "", 0, errors.New("request already processed")
	}
	if amount <= 0 {
		return "", 0, errors.New("invalid amount")
	}
	creditAmount := amount
	if requestedAmount > 0 {
		creditAmount = requestedAmount
	}
	if approvedAmount > 0 {
		creditAmount = approvedAmount
	}
	if creditAmount <= 0 {
		return "", 0, errors.New("invalid approved amount")
	}
	if strings.EqualFold(strings.TrimSpace(metode), "qris") {
		return "", 0, errors.New("approve qris gunakan proses qris")
	}
	if !bankID.Valid || bankID.Int64 <= 0 {
		return "", 0, errors.New("bank tujuan tidak valid")
	}

	var mutasiRows []depositBankMutationForApprove
	var refID string
	if len(selectedRefIDs) > 0 {
		mutasiRows, err = selectDepositBankMutationsByRefIDs(ctx, tx, bankID.Int64, selectedRefIDs)
		if err != nil {
			return "", 0, err
		}
		creditAmount = 0
		refs := make([]string, 0, len(mutasiRows))
		for _, row := range mutasiRows {
			creditAmount += row.Amount
			refs = append(refs, row.RefID)
		}
		refID = strings.Join(refs, ",")
	} else {
		var mutasiID int64
		err = tx.QueryRowContext(ctx, `
SELECT id, COALESCE(NULLIF(TRIM(ref_id), ''), $3::text) AS ref_id
FROM public.mutasi_bank
WHERE bank_id = $1::bigint
  AND jumlah = $2::bigint
  AND arah = 'CREDIT'
  AND alasan = 'BANK_MANUAL_IN'
  AND member_id IS NULL
  AND provider IS NULL
  AND NOT (COALESCE(meta, '{}'::jsonb) ? 'auto_deposit_request_id')
  AND NOT (COALESCE(meta, '{}'::jsonb) ? 'manual_deposit_request_id')
ORDER BY dibuat_pada ASC, id ASC
LIMIT 1
FOR UPDATE SKIP LOCKED
		`, bankID.Int64, creditAmount, fallbackRefID).Scan(&mutasiID, &refID)
		if errors.Is(err, sql.ErrNoRows) {
			// Manual approval is the operator's confirmation that the transfer
			// has reached the bank. Keep the accounting trail complete by
			// creating the incoming bank mutation inside this same transaction.
			var bankBefore int64
			if err := tx.QueryRowContext(ctx, `
SELECT saldo
FROM public.bank
WHERE id = $1::bigint
  AND aktif = true
FOR UPDATE
`, bankID.Int64).Scan(&bankBefore); err != nil {
				return "", 0, err
			}
			bankAfter := bankBefore + creditAmount
			if _, err := tx.ExecContext(ctx, `
UPDATE public.bank
SET saldo = $2::bigint,
    diubah_pada = now()
WHERE id = $1::bigint
`, bankID.Int64, bankAfter); err != nil {
				return "", 0, err
			}
			refID = fallbackRefID
			manualNote := fmt.Sprintf("Mutasi masuk dibuat saat approve deposit #%d", reqID)
			if trimmedNote := strings.TrimSpace(note); trimmedNote != "" {
				manualNote += " | " + trimmedNote
			}
			if err := tx.QueryRowContext(ctx, `
INSERT INTO public.mutasi_bank
  (bank_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, diubah_oleh, dibuat_pada, meta)
VALUES
  ($1,$2,'CREDIT',$3,'BANK_MANUAL_IN',$4,$5,$6,$7,now(),jsonb_build_object(
    'created_from_deposit_approval', true,
    'deposit_request_id', $8::bigint
  ))
RETURNING id
`, bankID.Int64, refID, creditAmount, manualNote, bankBefore, bankAfter, adminID, reqID).Scan(&mutasiID); err != nil {
				return "", 0, err
			}
			mutasiRows = []depositBankMutationForApprove{{ID: mutasiID, RefID: refID, Amount: creditAmount}}
			err = nil
		}
		if err != nil {
			return "", 0, err
		}
		refID = strings.TrimSpace(refID)
		if refID == "" {
			refID = fallbackRefID
		}
		mutasiRows = []depositBankMutationForApprove{{ID: mutasiID, RefID: refID, Amount: creditAmount}}
	}
	if creditAmount <= 0 {
		return "", 0, errors.New("invalid approved amount")
	}

	memberLabel := strings.TrimSpace(memberName)
	if memberLabel == "" {
		memberLabel = fmt.Sprintf("Member #%d", memberID)
	}
	bankNote := fmt.Sprintf("Manual approve deposit #%d member %s", reqID, memberLabel)
	if trimmedNote := strings.TrimSpace(note); trimmedNote != "" {
		bankNote += " | " + trimmedNote
	}
	walletNote := fmt.Sprintf("Manual approve dari mutasi bank %s", refID)
	if trimmedNote := strings.TrimSpace(note); trimmedNote != "" {
		walletNote = trimmedNote + " | " + walletNote
	}

	if err := r.creditWalletTx(ctx, tx, memberID, creditAmount, "deposit approve", walletNote, refID, adminID); err != nil {
		return "", 0, err
	}

	for _, mutasi := range mutasiRows {
		res, err := tx.ExecContext(ctx, `
UPDATE public.mutasi_bank
SET ref_id = CASE
      WHEN NULLIF(TRIM(COALESCE(ref_id, '')), '') IS NULL THEN $7::text
      ELSE ref_id
    END,
    member_id = $2::bigint,
    catatan = CASE
      WHEN NULLIF(TRIM(COALESCE(catatan, '')), '') IS NULL THEN $4
      ELSE catatan || ' | ' || $4
    END,
    meta = COALESCE(meta, '{}'::jsonb) || jsonb_build_object(
      'manual_approve', true,
      'manual_deposit_request_id', $1::bigint,
      'manual_deposit_member_id', $2::bigint,
      'manual_deposit_member_name', $8::text,
      'manual_deposit_amount', $3::bigint,
      'manual_deposit_ref_id', $7::text,
      'manual_approved_by', $9::bigint,
      'manual_approved_at', now()
    )
WHERE id = $5::bigint
  AND bank_id = $6::bigint
  AND jumlah = $3::bigint
  AND arah = 'CREDIT'
  AND alasan = 'BANK_MANUAL_IN'
  AND member_id IS NULL
  AND provider IS NULL
  AND NOT (COALESCE(meta, '{}'::jsonb) ? 'auto_deposit_request_id')
  AND NOT (COALESCE(meta, '{}'::jsonb) ? 'manual_deposit_request_id')
`, reqID, memberID, mutasi.Amount, bankNote, mutasi.ID, bankID.Int64, mutasi.RefID, memberLabel, adminID)
		if err != nil {
			return "", 0, err
		}
		if affected, _ := res.RowsAffected(); affected == 0 {
			return "", 0, errors.New("mutasi bank sudah diproses")
		}
	}

	res, err := tx.ExecContext(ctx, `
UPDATE public.deposit_request
SET ref_id = $1,
    status = 'approved',
    note = $2,
    approved_amount = $5,
    diproses_pada = now(),
    diproses_oleh = $3
WHERE id = $4
  AND member_id = $6::bigint
  AND status = 'pending'
`, refID, walletNote, adminID, reqID, creditAmount, memberID)
	if err != nil {
		return "", 0, err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return "", 0, errors.New("deposit pending sudah diproses")
	}

	if err := tx.Commit(); err != nil {
		return "", 0, err
	}
	return refID, creditAmount, nil
}

func (r *DepositRepository) ApproveQrisByRefID(ctx context.Context, refID, note string) error {
	if strings.TrimSpace(refID) == "" {
		return errors.New("ref_id required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var (
		memberID int64
		amount   int64
		status   string
		metode   string
	)
	err = tx.QueryRowContext(ctx, `
SELECT member_id, amount, status, metode
FROM public.deposit_request
WHERE TRIM(COALESCE(ref_id, '')) = $1
FOR UPDATE
`, strings.TrimSpace(refID)).Scan(&memberID, &amount, &status, &metode)
	if err != nil {
		return err
	}
	if strings.TrimSpace(strings.ToLower(metode)) != "qris" {
		return errors.New("deposit qris tidak ditemukan")
	}
	switch strings.TrimSpace(strings.ToLower(status)) {
	case "approved":
		return tx.Commit()
	case "pending":
	default:
		return errors.New("request already processed")
	}
	if amount <= 0 {
		return errors.New("invalid amount")
	}

	if err := r.creditWalletTx(ctx, tx, memberID, amount, "deposit qris", note, strings.TrimSpace(refID), 0); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE public.deposit_request
SET status = 'approved',
    note = CASE WHEN NULLIF($2, '') IS NULL THEN note ELSE $2 END,
    diproses_pada = now(),
    diproses_oleh = NULL
WHERE TRIM(COALESCE(ref_id, '')) = $1
`, strings.TrimSpace(refID), strings.TrimSpace(note)); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *DepositRepository) ApplyVACallback(ctx context.Context, ticketID, finalStatus, callbackDest, note string) (*DepositRequestRow, error) {
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" {
		return nil, errors.New("ref_id required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var (
		reqID     int64
		memberID  int64
		amount    int64
		status    string
		bankNomor string
	)
	err = tx.QueryRowContext(ctx, `
SELECT id, member_id, amount, status, bank_nomor_rekening
FROM public.deposit_request
WHERE TRIM(COALESCE(ref_id, '')) = $1
  AND LOWER(TRIM(COALESCE(metode, ''))) = 'va'
FOR UPDATE
`, ticketID).Scan(&reqID, &memberID, &amount, &status, &bankNomor)
	if err != nil {
		return nil, err
	}

	statusLower := strings.ToLower(strings.TrimSpace(status))
	finalStatus = strings.ToLower(strings.TrimSpace(finalStatus))
	note = strings.TrimSpace(note)

	switch finalStatus {
	case "success":
		if statusLower != "approved" {
			if amount <= 0 {
				return nil, errors.New("invalid amount")
			}
			if !depositVADestMatches(callbackDest, bankNomor) {
				return nil, errors.New("rekening tujuan callback VA tidak sesuai")
			}
			if err := r.creditWalletTx(ctx, tx, memberID, amount, "deposit va", note, ticketID, 0); err != nil {
				return nil, err
			}
			if err := r.creditLoketBayarProviderFromVATx(ctx, tx, ticketID, amount, note); err != nil {
				return nil, err
			}
			if _, err := tx.ExecContext(ctx, `
UPDATE public.deposit_request
SET status = 'approved',
    approved_amount = $2::bigint,
    note = CASE
      WHEN NULLIF($3::text, '') IS NULL THEN note
      WHEN NULLIF(TRIM(COALESCE(note, '')), '') IS NULL THEN $3::text
      ELSE note || ' | ' || $3::text
    END,
    diproses_pada = now(),
    diproses_oleh = NULL
WHERE id = $1
`, reqID, amount, note); err != nil {
				return nil, err
			}
		}
	case "failed":
		if statusLower == "ticket" || statusLower == "pending" || statusLower == "rejected" {
			if _, err := tx.ExecContext(ctx, `
UPDATE public.deposit_request
SET status = 'rejected',
    note = CASE
      WHEN NULLIF($2::text, '') IS NULL THEN note
      WHEN NULLIF(TRIM(COALESCE(note, '')), '') IS NULL THEN $2::text
      ELSE note || ' | ' || $2::text
    END,
    diproses_pada = now(),
    diproses_oleh = NULL
WHERE id = $1
`, reqID, note); err != nil {
				return nil, err
			}
		}
	default:
		// Pending or unknown VA callback statuses do not change the deposit row.
	}

	row, err := r.getByIDTx(ctx, tx, reqID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return row, nil
}

func (r *DepositRepository) ApproveVA(ctx context.Context, reqID, adminID, approvedAmount int64, note string) (*DepositRequestRow, error) {
	if reqID <= 0 {
		return nil, errors.New("invalid req_id")
	}
	if approvedAmount <= 0 {
		return nil, errors.New("nominal transfer wajib diisi")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var (
		memberID int64
		amount   int64
		refID    string
		status   string
	)
	err = tx.QueryRowContext(ctx, `
SELECT member_id, amount, COALESCE(ref_id, ''), status
FROM public.deposit_request
WHERE id = $1
  AND LOWER(TRIM(COALESCE(metode, ''))) = 'va'
FOR UPDATE
`, reqID).Scan(&memberID, &amount, &refID, &status)
	if err != nil {
		return nil, err
	}
	refID = strings.TrimSpace(refID)
	if refID == "" {
		return nil, errors.New("tiket VA belum ada ref_id")
	}
	statusLower := strings.ToLower(strings.TrimSpace(status))
	switch statusLower {
	case "approved":
		row, err := r.getByIDTx(ctx, tx, reqID)
		if err != nil {
			return nil, err
		}
		return row, tx.Commit()
	case "ticket", "pending":
	default:
		return nil, errors.New("request already processed")
	}
	if amount <= 0 {
		return nil, errors.New("invalid amount")
	}
	creditAmount := approvedAmount

	note = strings.TrimSpace(note)
	amountNote := fmt.Sprintf("nominal_transfer=%d tiket_va=%d", creditAmount, amount)
	if note == "" {
		note = "Manual approve deposit VA " + amountNote
	} else {
		note = "Manual approve deposit VA: " + note + " | " + amountNote
	}
	if err := r.creditWalletTx(ctx, tx, memberID, creditAmount, "deposit va", note, refID, adminID); err != nil {
		return nil, err
	}
	if err := r.creditLoketBayarProviderFromVATx(ctx, tx, refID, creditAmount, note); err != nil {
		return nil, err
	}

	var actor any
	if adminID > 0 {
		actor = adminID
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE public.deposit_request
SET status = 'approved',
    approved_amount = $2::bigint,
    note = CASE
      WHEN NULLIF($3::text, '') IS NULL THEN note
      WHEN NULLIF(TRIM(COALESCE(note, '')), '') IS NULL THEN $3::text
      ELSE note || ' | ' || $3::text
    END,
    diproses_pada = now(),
    diproses_oleh = $4::bigint
WHERE id = $1
`, reqID, creditAmount, note, actor); err != nil {
		return nil, err
	}

	row, err := r.getByIDTx(ctx, tx, reqID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return row, nil
}

func (r *DepositRepository) RejectVA(ctx context.Context, reqID, adminID int64, note string) (*DepositRequestRow, error) {
	if reqID <= 0 {
		return nil, errors.New("invalid req_id")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var status string
	if err := tx.QueryRowContext(ctx, `
SELECT status
FROM public.deposit_request
WHERE id = $1
  AND LOWER(TRIM(COALESCE(metode, ''))) = 'va'
FOR UPDATE
`, reqID).Scan(&status); err != nil {
		return nil, err
	}

	statusLower := strings.ToLower(strings.TrimSpace(status))
	switch statusLower {
	case "rejected":
		row, err := r.getByIDTx(ctx, tx, reqID)
		if err != nil {
			return nil, err
		}
		return row, tx.Commit()
	case "ticket", "pending":
	default:
		return nil, errors.New("request already processed")
	}

	note = strings.TrimSpace(note)
	if note == "" {
		note = "Manual reject deposit VA"
	} else {
		note = "Manual reject deposit VA: " + note
	}
	var actor any
	if adminID > 0 {
		actor = adminID
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE public.deposit_request
SET status = 'rejected',
    note = CASE
      WHEN NULLIF($2::text, '') IS NULL THEN note
      WHEN NULLIF(TRIM(COALESCE(note, '')), '') IS NULL THEN $2::text
      ELSE note || ' | ' || $2::text
    END,
    diproses_pada = now(),
    diproses_oleh = $3::bigint
WHERE id = $1
`, reqID, note, actor); err != nil {
		return nil, err
	}

	row, err := r.getByIDTx(ctx, tx, reqID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return row, nil
}

func (r *DepositRepository) creditLoketBayarProviderFromVATx(ctx context.Context, tx *sql.Tx, ticketID string, amount int64, note string) error {
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" || amount <= 0 {
		return errors.New("invalid provider deposit")
	}
	const provider = "loketbayar"
	const alasan = "PROVIDER_DEPOSIT"

	if _, err := tx.ExecContext(ctx, `
INSERT INTO public.dompet_provider (provider, saldo)
VALUES ($1, 0)
ON CONFLICT (provider) DO NOTHING
`, provider); err != nil {
		return err
	}

	var before int64
	if err := tx.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_provider
WHERE provider = $1
FOR UPDATE
`, provider).Scan(&before); err != nil {
		return err
	}

	var existing int
	err := tx.QueryRowContext(ctx, `
SELECT 1
FROM public.mutasi_dompet_provider
WHERE provider = $1
  AND ref_id = $2
  AND arah = 'credit'
  AND alasan = $3
LIMIT 1
`, provider, ticketID, alasan).Scan(&existing)
	if err == nil && existing == 1 {
		return nil
	}
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	after := before + amount
	if _, err := tx.ExecContext(ctx, `
UPDATE public.dompet_provider
SET saldo = $1::bigint,
    diperbarui_pada = now()
WHERE provider = $2
`, after, provider); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet_provider
  (provider, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada, meta)
VALUES
  ($1,$2,'credit',$3,$4,NULLIF($5::text,''),$6,$7,now(),jsonb_build_object('source','loketbayar_va_deposit_callback'))
`, provider, ticketID, amount, alasan, strings.TrimSpace(note), before, after)
	return err
}

func depositVADestMatches(callbackDest, storedDest string) bool {
	callbackDest = onlyDepositVADigits(callbackDest)
	storedDest = onlyDepositVADigits(storedDest)
	if callbackDest == "" || storedDest == "" {
		return true
	}
	return callbackDest == storedDest
}

func onlyDepositVADigits(value string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(value) {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func (r *DepositRepository) Reject(ctx context.Context, reqID, adminID int64, note string) error {
	if reqID <= 0 || adminID <= 0 {
		return errors.New("invalid req_id/admin_id")
	}
	res, err := r.db.ExecContext(ctx, `
UPDATE public.deposit_request
SET status = 'rejected',
    note = $1,
    diproses_pada = now(),
    diproses_oleh = $2
WHERE id = $3 AND status = 'pending'
`, strings.TrimSpace(note), adminID, reqID)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return errors.New("request already processed")
	}
	return nil
}

func (r *DepositRepository) RejectQrisByRefID(ctx context.Context, refID, note string) error {
	if strings.TrimSpace(refID) == "" {
		return errors.New("ref_id required")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status, metode string
	if err := tx.QueryRowContext(ctx, `
SELECT status, metode
FROM public.deposit_request
WHERE TRIM(COALESCE(ref_id, '')) = $1
FOR UPDATE
`, strings.TrimSpace(refID)).Scan(&status, &metode); err != nil {
		return err
	}
	if strings.TrimSpace(strings.ToLower(metode)) != "qris" {
		return errors.New("deposit qris tidak ditemukan")
	}
	switch strings.TrimSpace(strings.ToLower(status)) {
	case "rejected", "approved":
		return tx.Commit()
	case "pending":
	default:
		return tx.Commit()
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE public.deposit_request
SET status = 'rejected',
    note = CASE WHEN NULLIF($2, '') IS NULL THEN note ELSE $2 END,
    diproses_pada = now(),
    diproses_oleh = NULL
WHERE TRIM(COALESCE(ref_id, '')) = $1
`, strings.TrimSpace(refID), strings.TrimSpace(note)); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *DepositRepository) CreditInternal(ctx context.Context, memberID, amount int64, note, refID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := r.creditWalletTx(ctx, tx, memberID, amount, "deposit", note, refID, 0); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *DepositRepository) creditWalletTx(ctx context.Context, tx *sql.Tx, memberID, jumlah int64, alasan, catatan, refID string, diubahOleh int64) error {
	var before int64
	err := tx.QueryRowContext(ctx, `
SELECT saldo FROM public.dompet_member
WHERE member_id = $1::bigint
FOR UPDATE
`, memberID).Scan(&before)
	if err != nil {
		return err
	}
	after := before + jumlah

	_, err = tx.ExecContext(ctx, `
UPDATE public.dompet_member
SET saldo = $1::bigint, diperbarui_pada = now()
WHERE member_id = $2::bigint
`, after, memberID)
	if err != nil {
		return err
	}

	var actor any
	if diubahOleh > 0 {
		actor = diubahOleh
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, diubah_oleh, dibuat_pada)
VALUES
  ($1::bigint,$2::text,'CREDIT',$3::bigint,$4::text,NULLIF($5::text,''),$6::bigint,$7::bigint,$8::bigint,now())
`, memberID, refID, jumlah, alasan, strings.TrimSpace(catatan), before, after, actor)
	return err
}
