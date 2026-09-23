package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type DuplicateBankMutationError struct {
	RefID string
	Saldo int64
}

type BankProviderAssignResult struct {
	Provider      string `json:"provider"`
	ProviderSaldo int64  `json:"saldo_internal"`
	Amount        int64  `json:"amount"`
	RefID         string `json:"ref_id"`
	BankID        int64  `json:"bank_id"`
	BankNama      string `json:"bank_nama"`
}

const bcaOperationalAccountNumber = "3432738881"

type internalBankDestination struct {
	id      int64
	name    string
	account string
	owner   string
	active  bool
	digits  string
}

type scrapedProviderRefundMatch struct {
	bankMutationID int64
	refID          string
	provider       string
}

type scrapedRefundCreditMatch struct {
	id    int64
	refID string
}

func (e *DuplicateBankMutationError) Error() string {
	if strings.TrimSpace(e.RefID) == "" {
		return "mutasi bank sudah pernah masuk"
	}
	return fmt.Sprintf("mutasi bank sudah pernah masuk (ref_id %s)", e.RefID)
}

func (r *BankRepository) AdjustSaldo(ctx context.Context, actorID, bankID, amount int64, direction, reason, note, refID string) (before int64, after int64, err error) {
	direction = strings.TrimSpace(strings.ToLower(direction))
	if direction != "credit" && direction != "debit" {
		return 0, 0, errors.New("direction must be credit or debit")
	}
	if amount <= 0 {
		return 0, 0, errors.New("amount must be > 0")
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = tx.QueryRowContext(ctx, `
SELECT saldo
FROM public.bank
WHERE id = $1
FOR UPDATE
`, bankID).Scan(&before); err != nil {
		return 0, 0, err
	}

	after = before
	arahMutasi := "CREDIT"
	if direction == "credit" {
		after = before + amount
	} else {
		if before < amount {
			return 0, 0, errors.New("saldo bank tidak cukup")
		}
		after = before - amount
		arahMutasi = "DEBIT"
	}

	if _, err = tx.ExecContext(ctx, `
UPDATE public.bank
SET saldo = $2, diubah_pada = now()
WHERE id = $1
`, bankID, after); err != nil {
		return 0, 0, err
	}

	if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_bank
  (bank_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, diubah_oleh, dibuat_pada)
VALUES
  ($1,$2,$3,$4,$5,NULLIF($6,''),$7,$8,$9,now())
`, bankID, refID, arahMutasi, amount, reason, note, before, after, actorID); err != nil {
		return 0, 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, 0, err
	}
	return before, after, nil
}

func (r *BankRepository) ManualIncomingMutation(ctx context.Context, actorID, bankID, amount int64, note, refID string) (before int64, after int64, err error) {
	return r.ManualMutation(ctx, actorID, bankID, amount, "credit", note, refID)
}

func (r *BankRepository) ManualMutation(ctx context.Context, actorID, bankID, amount int64, direction, note, refID string) (before int64, after int64, err error) {
	return r.ManualMutationWithBalance(ctx, actorID, bankID, amount, direction, note, refID, 0)
}

func (r *BankRepository) ManualMutationWithBalance(ctx context.Context, actorID, bankID, amount int64, direction, note, refID string, actualBalance int64) (before int64, after int64, err error) {
	return r.ManualMutationWithBalanceDetails(ctx, actorID, bankID, amount, direction, note, refID, actualBalance, nil, "", "")
}

func (r *BankRepository) ManualMutationWithBalanceDetails(ctx context.Context, actorID, bankID, amount int64, direction, note, refID string, actualBalance int64, bankMutationAt *time.Time, sender, receiver string) (before int64, after int64, err error) {
	if amount <= 0 {
		return 0, 0, errors.New("amount must be > 0")
	}
	direction = strings.TrimSpace(strings.ToLower(direction))
	if direction != "credit" && direction != "debit" {
		return 0, 0, errors.New("direction must be credit or debit")
	}
	arahMutasi := "CREDIT"
	reason := "BANK_MANUAL_IN"
	if direction == "debit" {
		arahMutasi = "DEBIT"
		reason = "BANK_MANUAL_OUT"
	}
	refID = strings.TrimSpace(refID)
	hasExternalRef := strings.HasPrefix(refID, "K24-")
	hasScrapedBalance := hasExternalRef && actualBalance > 0
	sender = strings.TrimSpace(sender)
	receiver = strings.TrimSpace(receiver)
	mutationAt := normalizedBankMutationTime(bankMutationAt)
	if mutationAt == nil {
		if parsed, ok := bankMutationTimeFromNote(note); ok {
			mutationAt = &parsed
		}
	}
	semanticKey, legacySemanticKey, hasSemanticKey := bankMutationSemanticKeys(bankID, amount, direction, note, mutationAt, sender, receiver)

	loc, locErr := time.LoadLocation("Asia/Jakarta")
	if locErr != nil {
		loc = time.FixedZone("WIB", 7*60*60)
	}
	now := time.Now().In(loc)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	lockKey := fmt.Sprintf("bank-manual-%s:%d:%d:%s", direction, bankID, amount, startOfDay.Format("2006-01-02"))
	if hasSemanticKey {
		lockKey = fmt.Sprintf("bank-manual-semantic:%s", semanticKey)
	} else if hasExternalRef {
		lockKey = fmt.Sprintf("bank-manual-ref:%d:%s", bankID, refID)
	}
	if _, err = tx.ExecContext(ctx, `
SELECT pg_advisory_xact_lock(hashtextextended($1::text, 2405202603))
`, lockKey); err != nil {
		return 0, 0, err
	}

	if refID != "" && hasExternalRef {
		var existingRef string
		var existingID int64
		var existingSaldo int64
		refErr := tx.QueryRowContext(ctx, `
SELECT id, ref_id, saldo_sesudah
FROM public.mutasi_bank
WHERE bank_id = $1
  AND ref_id = $2
ORDER BY id DESC
LIMIT 1
`, bankID, refID).Scan(&existingID, &existingRef, &existingSaldo)
		if refErr == nil {
			if hasScrapedBalance {
				syncedSaldo, syncErr := syncDuplicateScrapedBankBalance(ctx, tx, bankID, existingID, direction, amount, actualBalance, existingRef)
				if syncErr != nil {
					return 0, 0, syncErr
				}
				if syncedSaldo > 0 {
					existingSaldo = syncedSaldo
				}
				if err = tx.Commit(); err != nil {
					return 0, 0, err
				}
			}
			return 0, existingSaldo, &DuplicateBankMutationError{RefID: existingRef, Saldo: existingSaldo}
		}
		if refErr != sql.ErrNoRows {
			return 0, 0, refErr
		}
	} else if refID != "" {
		var existingRef string
		var existingSaldo int64
		refErr := tx.QueryRowContext(ctx, `
SELECT ref_id, saldo_sesudah
FROM public.mutasi_bank
WHERE bank_id = $1
  AND ref_id = $2
  AND arah = $3
  AND alasan = $4
ORDER BY id DESC
LIMIT 1
`, bankID, refID, arahMutasi, reason).Scan(&existingRef, &existingSaldo)
		if refErr == nil {
			return 0, existingSaldo, &DuplicateBankMutationError{RefID: existingRef, Saldo: existingSaldo}
		}
		if refErr != sql.ErrNoRows {
			return 0, 0, refErr
		}
	}
	if hasSemanticKey {
		existingID, existingRef, existingSaldo, duplicate, duplicateErr := r.findSemanticDuplicateMutation(ctx, tx, bankID, amount, arahMutasi, reason, semanticKey, legacySemanticKey, mutationAt, sender, receiver, actualBalance)
		if duplicateErr != nil {
			return 0, 0, duplicateErr
		}
		if duplicate {
			if hasScrapedBalance && existingID > 0 {
				syncedSaldo, syncErr := syncDuplicateScrapedBankBalance(ctx, tx, bankID, existingID, direction, amount, actualBalance, existingRef)
				if syncErr != nil {
					return 0, 0, syncErr
				}
				if syncedSaldo > 0 {
					existingSaldo = syncedSaldo
				}
				if err = tx.Commit(); err != nil {
					return 0, 0, err
				}
			}
			return 0, existingSaldo, &DuplicateBankMutationError{RefID: existingRef, Saldo: existingSaldo}
		}
	}

	var internalDestination *internalBankDestination
	if direction == "debit" && hasExternalRef {
		internalDestination, err = r.findInternalBankDestinationFromScrapedDebit(ctx, tx, bankID, note, sender, receiver)
		if err != nil {
			return 0, 0, err
		}
		if internalDestination == nil && isScrapedBCAOperationalTransfer(amount, direction, note, sender, receiver) {
			internalDestination, err = r.findInternalBankDestinationByAccount(ctx, tx, bankID, bcaOperationalAccountNumber)
			if err != nil {
				return 0, 0, err
			}
		}
	}

	scrapedAdminFee := hasExternalRef && (isScrapedProviderTransferAdminFee(amount, direction, reason, note, sender, receiver) ||
		isScrapedBCAOperationalTransferAdminFee(amount, direction, note, sender, receiver) ||
		isScrapedInternalBankTransferAdminFee(amount, direction, internalDestination, note, sender, receiver))
	if scrapedAdminFee {
		reason = "BANK_TRANSFER_ADMIN_FEE"
	}
	internalBankTransfer := hasExternalRef && !scrapedAdminFee && internalDestination != nil

	detectedProvider := ""
	detectedProviderAccount := ""
	detectedProviderAccountName := ""
	if direction == "debit" && hasExternalRef && !scrapedAdminFee && !internalBankTransfer {
		detectedProvider, detectedProviderAccount, detectedProviderAccountName, err = r.findProviderDestinationFromScrapedDebit(ctx, tx, note, sender, receiver)
		if err != nil {
			return 0, 0, err
		}
		if detectedProvider != "" {
			reason = "BANK_TRANSFER_TO_PROVIDER"
		}
	}

	var providerRefund *scrapedProviderRefundMatch
	if direction == "credit" && hasExternalRef && isScrapedProviderRefundCredit(amount, direction, note, sender, receiver) {
		providerRefund, err = r.findScrapedProviderRefundMatch(ctx, tx, bankID, amount, mutationAt, note, sender, receiver)
		if err != nil {
			return 0, 0, err
		}
		if providerRefund != nil {
			reason = "BANK_TRANSFER_PROVIDER_REFUND"
		}
	}

	var duplicateRef sql.NullString
	if !hasExternalRef {
		duplicateMutationAt, hasDuplicateMutationAt := bankMutationSemanticTime(note, mutationAt)
		if hasDuplicateMutationAt {
			duplicateErr := tx.QueryRowContext(ctx, `
SELECT ref_id
FROM public.mutasi_bank
WHERE bank_id = $1
  AND jumlah = $2
  AND waktu_mutasi_bank = $3
  AND arah = $4
  AND alasan = $5
ORDER BY id DESC
LIMIT 1
`, bankID, amount, *duplicateMutationAt, arahMutasi, reason).Scan(&duplicateRef)
			if duplicateErr == nil {
				if duplicateRef.Valid && strings.TrimSpace(duplicateRef.String) != "" {
					return 0, 0, fmt.Errorf("nominal dan waktu transaksi ini sudah diinput untuk rekening ini (ref_id %s)", strings.TrimSpace(duplicateRef.String))
				}
				return 0, 0, errors.New("nominal dan waktu transaksi ini sudah diinput untuk rekening ini")
			}
			if duplicateErr != sql.ErrNoRows {
				return 0, 0, duplicateErr
			}
		}
	}

	var bankName string
	var bankAccount string
	if err = tx.QueryRowContext(ctx, `
SELECT saldo, nama, COALESCE(nomor_rekening, '')
FROM public.bank
WHERE id = $1
FOR UPDATE
`, bankID).Scan(&before, &bankName, &bankAccount); err != nil {
		return 0, 0, err
	}

	ledgerBefore := before
	if direction == "credit" {
		after = ledgerBefore + amount
	} else {
		if ledgerBefore < amount && !hasExternalRef {
			return 0, 0, errors.New("saldo bank tidak cukup")
		}
		after = ledgerBefore - amount
	}
	ledgerAfter := after
	if hasScrapedBalance {
		if scrapedBefore, scrapedAfter, ok := bankScrapedBalanceBeforeAfter(direction, amount, actualBalance); ok {
			before = scrapedBefore
			after = scrapedAfter
		}
	}
	if _, err = tx.ExecContext(ctx, `
UPDATE public.bank
SET saldo = $2, diubah_pada = now()
WHERE id = $1
`, bankID, after); err != nil {
		return 0, 0, err
	}

	bankMutationProvider := ""
	bankMutationMeta := map[string]any{}
	if hasScrapedBalance {
		bankMutationMeta["scraped_balance_source"] = "bank_scrape_checkpoint"
		bankMutationMeta["scraped_balance_after"] = actualBalance
		bankMutationMeta["scraped_balance_adopted"] = true
		bankMutationMeta["scraped_balance_before"] = before
		bankMutationMeta["ledger_balance_before"] = ledgerBefore
		bankMutationMeta["ledger_balance_after"] = ledgerAfter
		bankMutationMeta["ledger_scraped_balance_delta"] = ledgerAfter - actualBalance
	}
	if mutationAt != nil {
		bankMutationMeta["waktu_mutasi_bank"] = mutationAt.Format(time.RFC3339)
	}
	if sender != "" {
		bankMutationMeta["pengirim"] = sender
	}
	if receiver != "" {
		bankMutationMeta["penerima"] = receiver
	}
	if hasSemanticKey {
		bankMutationMeta["semantic_key"] = semanticKey
		if legacySemanticKey != "" && legacySemanticKey != semanticKey {
			bankMutationMeta["semantic_key_legacy"] = legacySemanticKey
		}
		bankMutationMeta["semantic_key_source"] = "bank_id,transaction_time,amount,direction,pengirim,penerima"
	}
	skipDetectedProviderCredit := false
	matchedManualProviderTopupRef := ""
	providerCreditSkipReason := ""
	var matchedRefundCredit *scrapedRefundCreditMatch
	if detectedProvider != "" {
		matchedManualProviderTopupRef, err = r.findRecentManualProviderTopupForScrape(ctx, tx, bankID, detectedProvider, amount, mutationAt)
		if err != nil {
			return 0, 0, err
		}
		if matchedManualProviderTopupRef != "" {
			skipDetectedProviderCredit = true
			providerCreditSkipReason = "manual_provider_topup_already_recorded"
			bankMutationMeta["provider_credit_skipped"] = true
			bankMutationMeta["provider_credit_skip_reason"] = providerCreditSkipReason
			bankMutationMeta["manual_provider_topup_ref"] = matchedManualProviderTopupRef
		} else {
			matchedRefundCredit, err = r.findExistingScrapedRefundCreditForProviderDebit(ctx, tx, bankID, detectedProvider, amount, mutationAt, note, sender, receiver)
			if err != nil {
				return 0, 0, err
			}
			if matchedRefundCredit != nil {
				skipDetectedProviderCredit = true
				providerCreditSkipReason = "bank_refund_detected"
				bankMutationMeta["provider_credit_skipped"] = true
				bankMutationMeta["provider_credit_skip_reason"] = providerCreditSkipReason
				bankMutationMeta["matched_refund_bank_ref"] = matchedRefundCredit.refID
				bankMutationMeta["matched_refund_bank_mutasi_id"] = matchedRefundCredit.id
			}
		}
		bankMutationProvider = detectedProvider
		bankMutationMeta["type"] = "bank_transfer_to_provider"
		bankMutationMeta["provider"] = detectedProvider
		bankMutationMeta["provider_account"] = detectedProviderAccount
		if strings.TrimSpace(detectedProviderAccountName) != "" {
			bankMutationMeta["provider_account_name"] = strings.TrimSpace(detectedProviderAccountName)
		}
		bankMutationMeta["source"] = "bank_scrape"
	}
	if providerRefund != nil {
		bankMutationProvider = providerRefund.provider
		bankMutationMeta["type"] = "bank_transfer_provider_refund"
		bankMutationMeta["provider"] = providerRefund.provider
		bankMutationMeta["source"] = "bank_scrape"
		bankMutationMeta["refund_of_bank_ref"] = providerRefund.refID
		bankMutationMeta["refund_of_bank_mutasi_id"] = providerRefund.bankMutationID
	}
	if internalBankTransfer {
		reason = "BANK_TRANSFER_OUT"
		bankMutationProvider = internalDestination.account
		bankMutationMeta["type"] = "bank_internal_transfer"
		bankMutationMeta["source"] = "bank_scrape"
		bankMutationMeta["destination_bank_id"] = internalDestination.id
		bankMutationMeta["destination_bank_name"] = strings.TrimSpace(internalDestination.name)
		bankMutationMeta["destination_account"] = strings.TrimSpace(internalDestination.account)
		bankMutationMeta["destination_account_name"] = strings.TrimSpace(internalDestination.owner)
	}
	bankMutationMetaJSON, _ := json.Marshal(bankMutationMeta)
	var mutationAtValue any
	if mutationAt != nil {
		mutationAtValue = *mutationAt
	}

	var insertedBankMutationID int64
	if err = tx.QueryRowContext(ctx, `
INSERT INTO public.mutasi_bank
  (bank_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, diubah_oleh, provider, dibuat_pada, meta, waktu_mutasi_bank, pengirim, penerima)
VALUES
  ($1,$2,$3,$4,$5,NULLIF($6,''),$7,$8,$9,NULLIF($10,''),now(),$11::jsonb,$12,NULLIF($13,''),NULLIF($14,''))
RETURNING id
`, bankID, refID, arahMutasi, amount, reason, note, before, after, actorID, bankMutationProvider, string(bankMutationMetaJSON), mutationAtValue, sender, receiver).Scan(&insertedBankMutationID); err != nil {
		return 0, 0, err
	}
	if hasScrapedBalance && mutationAt != nil {
		if err = recomputeScrapedBankLedgerForward(ctx, tx, bankID, insertedBankMutationID, after, refID); err != nil {
			return 0, 0, err
		}
	}

	if internalBankTransfer {
		if err = r.creditInternalBankFromScrapedMutation(ctx, tx, actorID, bankID, bankName, bankAccount, internalDestination, refID, amount, note, mutationAtValue, sender, receiver); err != nil {
			return 0, 0, err
		}
	}

	if detectedProvider != "" && skipDetectedProviderCredit {
		if matchedManualProviderTopupRef != "" {
			if _, err = tx.ExecContext(ctx, `
UPDATE public.mutasi_dompet_provider
SET meta = COALESCE(meta, '{}'::jsonb) || jsonb_build_object(
  'matched_k24_ref', $4::text,
  'matched_k24_at', now(),
  'provider_credit_skip_reason', 'manual_provider_topup_already_recorded'
)
WHERE provider = $1
  AND ref_id = $2
  AND jumlah = $3
  AND arah = 'credit'
  AND alasan = 'BANK_TRANSFER_IN'
`, detectedProvider, matchedManualProviderTopupRef, amount, refID); err != nil {
				return 0, 0, err
			}
			if _, err = tx.ExecContext(ctx, `
UPDATE public.mutasi_bank
SET meta = COALESCE(meta, '{}'::jsonb) || jsonb_build_object(
  'matched_k24_ref', $4::text,
  'provider_credit_skip_reason', 'manual_provider_topup_already_recorded'
)
WHERE bank_id = $1
  AND ref_id = $2
  AND jumlah = $3
  AND arah = 'DEBIT'
  AND alasan = 'BANK_TRANSFER_TO_PROVIDER'
`, bankID, matchedManualProviderTopupRef, amount, refID); err != nil {
				return 0, 0, err
			}
		}
		if matchedRefundCredit != nil {
			if _, err = tx.ExecContext(ctx, `
UPDATE public.mutasi_bank
SET alasan = 'BANK_TRANSFER_PROVIDER_REFUND',
    provider = NULLIF($2::text, ''),
    meta = COALESCE(meta, '{}'::jsonb) || jsonb_build_object(
      'type', 'bank_transfer_provider_refund',
      'provider', $2::text,
      'refund_matched_debit_ref', $3::text,
      'refund_matched_at', now()
    )
WHERE id = $1
`, matchedRefundCredit.id, detectedProvider, refID); err != nil {
				return 0, 0, err
			}
		}
	}

	if detectedProvider != "" && !skipDetectedProviderCredit {
		if _, err = tx.ExecContext(ctx, `
INSERT INTO public.dompet_provider (provider, saldo)
VALUES ($1, 0)
ON CONFLICT (provider) DO NOTHING
`, detectedProvider); err != nil {
			return 0, 0, err
		}

		var providerBefore int64
		if err = tx.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_provider
WHERE provider = $1
FOR UPDATE
`, detectedProvider).Scan(&providerBefore); err != nil {
			return 0, 0, err
		}
		providerAfter := providerBefore + amount

		if _, err = tx.ExecContext(ctx, `
UPDATE public.dompet_provider
SET saldo = $2, diperbarui_pada = now()
WHERE provider = $1
`, detectedProvider, providerAfter); err != nil {
			return 0, 0, err
		}

		providerMeta := map[string]any{
			"type":             "bank_transfer_to_provider",
			"provider":         detectedProvider,
			"provider_account": detectedProviderAccount,
			"source":           "bank_scrape",
		}
		if strings.TrimSpace(detectedProviderAccountName) != "" {
			providerMeta["provider_account_name"] = strings.TrimSpace(detectedProviderAccountName)
		}
		providerMetaJSON, _ := json.Marshal(providerMeta)
		if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet_provider
  (provider, bank_id, bank_nama, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, diubah_oleh, dibuat_pada, meta)
VALUES
  ($1,$2,$3,$4,'credit',$5,'BANK_TRANSFER_IN',NULLIF($6,''),$7,$8,$9,now(),$10::jsonb)
	`, detectedProvider, bankID, strings.TrimSpace(bankName), refID, amount, note, providerBefore, providerAfter, actorID, string(providerMetaJSON)); err != nil {
			return 0, 0, err
		}
	}

	if providerRefund != nil {
		if err = r.reverseProviderCreditFromScrapedRefund(ctx, tx, actorID, providerRefund, refID, amount, note); err != nil {
			return 0, 0, err
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, 0, err
	}
	return before, after, nil
}

func bankScrapedBalanceBeforeAfter(direction string, amount int64, actualBalance int64) (before int64, after int64, ok bool) {
	if amount <= 0 || actualBalance <= 0 {
		return 0, 0, false
	}
	after = actualBalance
	switch strings.TrimSpace(strings.ToLower(direction)) {
	case "credit":
		before = after - amount
	case "debit":
		before = after + amount
	default:
		return 0, 0, false
	}
	return before, after, true
}

func syncDuplicateScrapedBankBalance(ctx context.Context, tx *sql.Tx, bankID, mutationID int64, direction string, amount, actualBalance int64, refID string) (int64, error) {
	if tx == nil || bankID <= 0 || mutationID <= 0 {
		return 0, nil
	}
	before, after, ok := bankScrapedBalanceBeforeAfter(direction, amount, actualBalance)
	if !ok {
		return 0, nil
	}
	var currentBankSaldo int64
	var semanticKey sql.NullString
	var note sql.NullString
	var mutationAt sql.NullTime
	if err := tx.QueryRowContext(ctx, `
SELECT b.saldo,
       COALESCE(mb.meta->>'semantic_key', ''),
       COALESCE(mb.catatan, ''),
       mb.waktu_mutasi_bank
FROM public.bank b
JOIN public.mutasi_bank mb ON mb.bank_id = b.id
WHERE b.id = $1
  AND mb.id = $2
FOR UPDATE
`, bankID, mutationID).Scan(&currentBankSaldo, &semanticKey, &note, &mutationAt); err != nil {
		return 0, err
	}
	_ = currentBankSaldo
	shouldRecomputeForward := duplicateScrapedBalanceShouldRecomputeForward(semanticKey.String, note.String, mutationAt)
	_, err := tx.ExecContext(ctx, `
UPDATE public.mutasi_bank
SET saldo_sebelum = $3,
    saldo_sesudah = $4,
    meta = COALESCE(meta, '{}'::jsonb) || jsonb_build_object(
      'scraped_duplicate_balance_sync', true,
      'scraped_duplicate_balance_sync_at', now(),
      'scraped_duplicate_balance_ref', $5::text,
      'scraped_balance_source', 'bank_scrape_checkpoint',
      'scraped_balance_after', $4,
      'scraped_balance_before', $3,
      'scraped_balance_adopted', true
    )
WHERE id = $1
  AND bank_id = $2
  AND (saldo_sebelum <> $3 OR saldo_sesudah <> $4 OR COALESCE(meta, '{}'::jsonb)->>'scraped_duplicate_balance_sync' IS DISTINCT FROM 'true')
`, mutationID, bankID, before, after, strings.TrimSpace(refID))
	if err != nil {
		return 0, err
	}
	if shouldRecomputeForward {
		if err := recomputeScrapedBankLedgerForward(ctx, tx, bankID, mutationID, after, refID); err != nil {
			return 0, err
		}
		return after, nil
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE public.bank
SET saldo = $2,
    diubah_pada = now()
WHERE id = $1
`, bankID, after); err != nil {
		return 0, err
	}
	return after, nil
}

func duplicateScrapedBalanceShouldRecomputeForward(semanticKey, note string, mutationAt sql.NullTime) bool {
	if strings.Contains(strings.ToLower(strings.TrimSpace(semanticKey)), "|balance:") {
		return false
	}
	if mutationAt.Valid && bankMutationHasClock(mutationAt.Time) {
		return true
	}
	if parsed, ok := bankMutationTimeFromNote(note); ok && bankMutationHasClock(parsed) {
		return true
	}
	return false
}

type bankLedgerRecomputeRow struct {
	id                  int64
	arah                string
	jumlah              int64
	mutationTime        time.Time
	scrapedBalanceAfter int64
}

func recomputeScrapedBankLedgerForward(ctx context.Context, tx *sql.Tx, bankID, anchorMutationID, anchorAfter int64, anchorRef string) error {
	if tx == nil || bankID <= 0 || anchorMutationID <= 0 || anchorAfter <= 0 {
		return nil
	}
	var anchorTime time.Time
	if err := tx.QueryRowContext(ctx, `
SELECT COALESCE(waktu_mutasi_bank, dibuat_pada)
FROM public.mutasi_bank
WHERE id = $1
  AND bank_id = $2
FOR UPDATE
`, anchorMutationID, bankID).Scan(&anchorTime); err != nil {
		return err
	}

	rows, err := tx.QueryContext(ctx, `
SELECT
  id,
  arah,
  jumlah,
  COALESCE(waktu_mutasi_bank, dibuat_pada),
  CASE
    WHEN COALESCE(meta->>'scraped_balance_after', '') ~ '^-?[0-9]+$'
      THEN (meta->>'scraped_balance_after')::bigint
    ELSE 0
  END
FROM public.mutasi_bank
WHERE bank_id = $1
  AND (
	COALESCE(waktu_mutasi_bank, dibuat_pada) > $2
	OR (COALESCE(waktu_mutasi_bank, dibuat_pada) = $2 AND id <> $3)
  )
ORDER BY COALESCE(waktu_mutasi_bank, dibuat_pada), id
FOR UPDATE
`, bankID, anchorTime, anchorMutationID)
	if err != nil {
		return err
	}
	var later []bankLedgerRecomputeRow
	for rows.Next() {
		var item bankLedgerRecomputeRow
		if err := rows.Scan(&item.id, &item.arah, &item.jumlah, &item.mutationTime, &item.scrapedBalanceAfter); err != nil {
			_ = rows.Close()
			return err
		}
		later = append(later, item)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if err := rows.Err(); err != nil {
		return err
	}
	later = orderBankLedgerRecomputeRows(anchorAfter, later)

	current := anchorAfter
	for _, item := range later {
		before, next, ok := bankLedgerRecomputeBeforeAfter(current, item)
		if !ok {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
UPDATE public.mutasi_bank
SET saldo_sebelum = $2,
    saldo_sesudah = $3,
    meta = COALESCE(meta, '{}'::jsonb) || jsonb_build_object(
      'ledger_recomputed_from_ref', $4::text,
      'ledger_recomputed_at', now()
    )
WHERE id = $1
`, item.id, before, next, strings.TrimSpace(anchorRef)); err != nil {
			return err
		}
		current = next
	}

	_, err = tx.ExecContext(ctx, `
UPDATE public.bank
SET saldo = $2,
    diubah_pada = now()
WHERE id = $1
`, bankID, current)
	return err
}

func bankLedgerRecomputeBeforeAfter(current int64, row bankLedgerRecomputeRow) (before int64, after int64, ok bool) {
	if scrapedBefore, scrapedAfter, scrapedOK := row.scrapedBeforeAfter(); scrapedOK {
		return scrapedBefore, scrapedAfter, true
	}
	next, ok := bankLedgerApplyMutation(current, row.arah, row.jumlah)
	if !ok {
		return 0, 0, false
	}
	return current, next, true
}

func orderBankLedgerRecomputeRows(anchorAfter int64, rows []bankLedgerRecomputeRow) []bankLedgerRecomputeRow {
	if len(rows) < 2 || anchorAfter <= 0 {
		return rows
	}
	ordered := make([]bankLedgerRecomputeRow, 0, len(rows))
	current := anchorAfter
	for i := 0; i < len(rows); {
		j := i + 1
		for j < len(rows) && rows[j].mutationTime.Equal(rows[i].mutationTime) {
			j++
		}
		group := rows[i:j]
		if len(group) == 1 {
			item := group[0]
			if i == 0 {
				if before, _, ok := item.scrapedBeforeAfter(); ok && before != current {
					i = j
					continue
				}
			}
			ordered = append(ordered, item)
			if next, ok := bankLedgerApplyMutation(current, item.arah, item.jumlah); ok {
				current = next
			}
			i = j
			continue
		}
		chained := orderBankLedgerSameTimestampGroup(current, group, i == 0)
		for _, item := range chained {
			ordered = append(ordered, item)
			if next, ok := bankLedgerApplyMutation(current, item.arah, item.jumlah); ok {
				current = next
			}
		}
		i = j
	}
	return ordered
}

func orderBankLedgerSameTimestampGroup(current int64, rows []bankLedgerRecomputeRow, skipUnchained bool) []bankLedgerRecomputeRow {
	remaining := append([]bankLedgerRecomputeRow(nil), rows...)
	ordered := make([]bankLedgerRecomputeRow, 0, len(rows))
	for len(remaining) > 0 && current > 0 {
		match := -1
		var after int64
		for idx, item := range remaining {
			before, candidateAfter, ok := item.scrapedBeforeAfter()
			if ok && before == current {
				match = idx
				after = candidateAfter
				break
			}
		}
		if match < 0 {
			break
		}
		item := remaining[match]
		ordered = append(ordered, item)
		current = after
		remaining = append(remaining[:match], remaining[match+1:]...)
	}
	if !skipUnchained {
		ordered = append(ordered, remaining...)
	}
	return ordered
}

func (row bankLedgerRecomputeRow) scrapedBeforeAfter() (before int64, after int64, ok bool) {
	if row.scrapedBalanceAfter <= 0 {
		return 0, 0, false
	}
	return bankScrapedBalanceBeforeAfter(row.arah, row.jumlah, row.scrapedBalanceAfter)
}

func bankLedgerApplyMutation(before int64, direction string, amount int64) (int64, bool) {
	if amount <= 0 {
		return 0, false
	}
	switch strings.TrimSpace(strings.ToUpper(direction)) {
	case "CREDIT":
		return before + amount, true
	case "DEBIT":
		return before - amount, true
	default:
		return 0, false
	}
}

func (r *BankRepository) findScrapedProviderRefundMatch(ctx context.Context, tx *sql.Tx, bankID, amount int64, mutationAt *time.Time, note, sender, receiver string) (*scrapedProviderRefundMatch, error) {
	if bankID <= 0 || amount <= 0 || !isScrapedProviderRefundCredit(amount, "credit", note, sender, receiver) {
		return nil, nil
	}
	refundTokens := bankReferenceTokens(note, sender, receiver)
	if len(refundTokens) == 0 {
		return nil, nil
	}
	start, end, anchor := bankMutationWindow(mutationAt, 24*time.Hour, 6*time.Hour)
	rows, err := tx.QueryContext(ctx, `
SELECT
  id,
  COALESCE(ref_id, ''),
  lower(trim(COALESCE(provider, ''))),
  COALESCE(catatan, ''),
  COALESCE(pengirim, ''),
  COALESCE(penerima, '')
FROM public.mutasi_bank
WHERE bank_id = $1
  AND jumlah = $2
  AND arah = 'DEBIT'
  AND alasan = 'BANK_TRANSFER_TO_PROVIDER'
  AND COALESCE(NULLIF(trim(provider), ''), '') <> ''
  AND COALESCE(waktu_mutasi_bank, dibuat_pada) >= $3
  AND COALESCE(waktu_mutasi_bank, dibuat_pada) <= $4
  AND NOT EXISTS (
    SELECT 1
    FROM public.mutasi_dompet_provider rev
    WHERE lower(trim(rev.provider)) = lower(trim(public.mutasi_bank.provider))
      AND rev.arah = 'debit'
      AND rev.jumlah = public.mutasi_bank.jumlah
      AND COALESCE(rev.meta->>'reverses_bank_ref_id', '') = public.mutasi_bank.ref_id
  )
ORDER BY abs(extract(epoch FROM (COALESCE(waktu_mutasi_bank, dibuat_pada) - $5::timestamptz))) ASC, id DESC
LIMIT 50
`, bankID, amount, start, end, anchor)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var match *scrapedProviderRefundMatch
	for rows.Next() {
		var candidate scrapedProviderRefundMatch
		var candidateNote, candidateSender, candidateReceiver string
		if err := rows.Scan(&candidate.bankMutationID, &candidate.refID, &candidate.provider, &candidateNote, &candidateSender, &candidateReceiver); err != nil {
			return nil, err
		}
		if !bankReferenceTokensIntersect(refundTokens, bankReferenceTokens(candidate.refID, candidateNote, candidateSender, candidateReceiver)) {
			continue
		}
		if match != nil {
			return nil, nil
		}
		candidate.refID = strings.TrimSpace(candidate.refID)
		candidate.provider = strings.TrimSpace(strings.ToLower(candidate.provider))
		match = &candidate
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return match, nil
}

func (r *BankRepository) findExistingScrapedRefundCreditForProviderDebit(ctx context.Context, tx *sql.Tx, bankID int64, provider string, amount int64, mutationAt *time.Time, note, sender, receiver string) (*scrapedRefundCreditMatch, error) {
	provider = strings.TrimSpace(strings.ToLower(provider))
	if bankID <= 0 || provider == "" || amount <= 0 {
		return nil, nil
	}
	debitTokens := bankReferenceTokens(note, sender, receiver)
	if len(debitTokens) == 0 {
		return nil, nil
	}
	start, end, anchor := bankMutationWindow(mutationAt, 6*time.Hour, 24*time.Hour)
	rows, err := tx.QueryContext(ctx, `
SELECT
  id,
  COALESCE(ref_id, ''),
  COALESCE(catatan, ''),
  COALESCE(pengirim, ''),
  COALESCE(penerima, '')
FROM public.mutasi_bank
WHERE bank_id = $1
  AND jumlah = $2
  AND arah = 'CREDIT'
  AND COALESCE(waktu_mutasi_bank, dibuat_pada) >= $3
  AND COALESCE(waktu_mutasi_bank, dibuat_pada) <= $4
  AND (
    alasan = 'BANK_TRANSFER_PROVIDER_REFUND'
    OR COALESCE(catatan, '') ~* '(refund|retur|reversal|koreksi|dikembalikan|pembatalan|(^|[^A-Z0-9])KOR([^A-Z0-9]|$))'
    OR COALESCE(pengirim, '') ~* '(refund|retur|reversal|koreksi|dikembalikan|pembatalan|(^|[^A-Z0-9])KOR([^A-Z0-9]|$))'
    OR COALESCE(penerima, '') ~* '(refund|retur|reversal|koreksi|dikembalikan|pembatalan|(^|[^A-Z0-9])KOR([^A-Z0-9]|$))'
  )
ORDER BY abs(extract(epoch FROM (COALESCE(waktu_mutasi_bank, dibuat_pada) - $5::timestamptz))) ASC, id DESC
LIMIT 50
`, bankID, amount, start, end, anchor)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var match *scrapedRefundCreditMatch
	for rows.Next() {
		var candidate scrapedRefundCreditMatch
		var candidateNote, candidateSender, candidateReceiver string
		if err := rows.Scan(&candidate.id, &candidate.refID, &candidateNote, &candidateSender, &candidateReceiver); err != nil {
			return nil, err
		}
		if !isScrapedProviderRefundCredit(amount, "credit", candidateNote, candidateSender, candidateReceiver) {
			continue
		}
		if !bankReferenceTokensIntersect(debitTokens, bankReferenceTokens(candidate.refID, candidateNote, candidateSender, candidateReceiver)) {
			continue
		}
		if match != nil {
			return nil, nil
		}
		candidate.refID = strings.TrimSpace(candidate.refID)
		match = &candidate
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return match, nil
}

func (r *BankRepository) reverseProviderCreditFromScrapedRefund(ctx context.Context, tx *sql.Tx, actorID int64, refund *scrapedProviderRefundMatch, refundRefID string, amount int64, note string) error {
	if refund == nil || strings.TrimSpace(refund.provider) == "" || strings.TrimSpace(refund.refID) == "" || amount <= 0 {
		return nil
	}
	provider := strings.TrimSpace(strings.ToLower(refund.provider))
	var providerCreditID int64
	var providerCreditReason string
	err := tx.QueryRowContext(ctx, `
SELECT id, COALESCE(alasan, '')
FROM public.mutasi_dompet_provider
WHERE provider = $1
  AND ref_id = $2
  AND jumlah = $3
  AND arah = 'credit'
  AND alasan = 'BANK_TRANSFER_IN'
ORDER BY id DESC
LIMIT 1
FOR UPDATE
`, provider, refund.refID, amount).Scan(&providerCreditID, &providerCreditReason)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}

	var existingReversalID int64
	existingErr := tx.QueryRowContext(ctx, `
SELECT id
FROM public.mutasi_dompet_provider
WHERE provider = $1
  AND arah = 'debit'
  AND jumlah = $2
  AND (
    COALESCE(meta->>'reverses_provider_ledger_id', '') = $3
    OR COALESCE(meta->>'reverses_bank_ref_id', '') = $4
    OR ref_id = $5
  )
ORDER BY id DESC
LIMIT 1
FOR UPDATE
`, provider, amount, fmt.Sprint(providerCreditID), refund.refID, refundRefID).Scan(&existingReversalID)
	if existingErr == nil {
		return nil
	}
	if existingErr != sql.ErrNoRows {
		return existingErr
	}

	if _, err = tx.ExecContext(ctx, `
INSERT INTO public.dompet_provider (provider, saldo)
VALUES ($1, 0)
ON CONFLICT (provider) DO NOTHING
`, provider); err != nil {
		return err
	}
	var providerBefore int64
	if err = tx.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_provider
WHERE provider = $1
FOR UPDATE
`, provider).Scan(&providerBefore); err != nil {
		return err
	}
	providerAfter := providerBefore - amount
	if _, err = tx.ExecContext(ctx, `
UPDATE public.dompet_provider
SET saldo = $2, diperbarui_pada = now()
WHERE provider = $1
`, provider, providerAfter); err != nil {
		return err
	}

	reversalMeta, _ := json.Marshal(map[string]any{
		"type":                        "bank_transfer_provider_refund",
		"source":                      "bank_scrape",
		"provider":                    provider,
		"refund_bank_ref_id":          refundRefID,
		"reverses_bank_ref_id":        refund.refID,
		"reverses_bank_mutasi_id":     refund.bankMutationID,
		"reverses_provider_ledger_id": providerCreditID,
		"previous_provider_reason":    providerCreditReason,
	})
	if _, err = tx.ExecContext(ctx, `
UPDATE public.mutasi_dompet_provider
SET alasan = 'BANK_TRANSFER_IN_REFUNDED',
    meta = COALESCE(meta, '{}'::jsonb) || jsonb_build_object(
      'refunded_by_bank_ref_id', $4::text,
      'refunded_at', now(),
      'refund_reason', 'bank_scrape_refund_credit_detected'
    )
WHERE id = $1
  AND provider = $2
  AND ref_id = $3
`, providerCreditID, provider, refund.refID, refundRefID); err != nil {
		return err
	}
	reversalNote := "Auto reverse provider credit karena mutasi bank refund/koreksi"
	if trimmed := strings.TrimSpace(note); trimmed != "" {
		reversalNote += " | " + trimmed
	}
	if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet_provider
  (provider, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, diubah_oleh, dibuat_pada, meta)
VALUES
  ($1,$2,'debit',$3,'BANK_TRANSFER_REFUND_REVERSAL',NULLIF($4,''),$5,$6,$7,now(),$8::jsonb)
`, provider, refundRefID, amount, reversalNote, providerBefore, providerAfter, actorID, string(reversalMeta)); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `
UPDATE public.mutasi_bank
SET meta = COALESCE(meta, '{}'::jsonb) || jsonb_build_object(
  'provider_credit_refunded_by_bank_ref', $2::text,
  'provider_credit_refunded_at', now()
)
WHERE id = $1
`, refund.bankMutationID, refundRefID); err != nil {
		return err
	}
	return nil
}

func (r *BankRepository) findProviderDestinationFromScrapedDebit(ctx context.Context, tx *sql.Tx, note, sender, receiver string) (string, string, string, error) {
	searchText := strings.Join([]string{note, sender, receiver}, " ")
	searchDigits := bankDigitsOnly(searchText)
	if searchDigits != "" {
		var provider string
		var account string
		var accountName string
		err := tx.QueryRowContext(ctx, `
SELECT lower(trim(pr.provider)), COALESCE(pr.nomor_rekening_digits, ''), COALESCE(pr.nama, '')
FROM public.provider_rekening pr
JOIN public.provider p ON lower(trim(p.nama)) = lower(trim(pr.provider))
WHERE pr.aktif = true
  AND COALESCE(pr.nomor_rekening_digits, '') <> ''
  AND $1 LIKE '%' || pr.nomor_rekening_digits || '%'
ORDER BY length(pr.nomor_rekening_digits) DESC, pr.id
LIMIT 1
`, searchDigits).Scan(&provider, &account, &accountName)
		if err == nil {
			return strings.TrimSpace(strings.ToLower(provider)), strings.TrimSpace(account), strings.TrimSpace(accountName), nil
		}
		if err != sql.ErrNoRows {
			return "", "", "", err
		}
	}

	normalizedText := bankNormalizeProviderName(searchText)
	if normalizedText == "" {
		return "", "", "", nil
	}

	rows, err := tx.QueryContext(ctx, `
SELECT lower(trim(pr.provider)), COALESCE(pr.nomor_rekening_digits, ''), COALESCE(pr.nama, '')
FROM public.provider_rekening pr
JOIN public.provider p ON lower(trim(p.nama)) = lower(trim(pr.provider))
WHERE pr.aktif = true
  AND COALESCE(pr.nama, '') <> ''
ORDER BY pr.id
`)
	if err != nil {
		return "", "", "", err
	}
	defer rows.Close()

	bestScore := 0
	bestProvider := ""
	bestAccount := ""
	bestName := ""
	ambiguous := false
	for rows.Next() {
		var provider string
		var account string
		var name string
		if err := rows.Scan(&provider, &account, &name); err != nil {
			return "", "", "", err
		}
		provider = strings.TrimSpace(strings.ToLower(provider))
		if provider == "" {
			continue
		}
		score := bankProviderNameMatchScore(name, normalizedText)
		if score <= 0 {
			continue
		}
		if score > bestScore {
			bestScore = score
			bestProvider = provider
			bestAccount = strings.TrimSpace(account)
			bestName = strings.TrimSpace(name)
			ambiguous = false
			continue
		}
		if score == bestScore && provider != bestProvider {
			ambiguous = true
		}
	}
	if err := rows.Err(); err != nil {
		return "", "", "", err
	}
	if bestScore > 0 && !ambiguous {
		return bestProvider, bestAccount, bestName, nil
	}
	return "", "", "", nil
}

func (r *BankRepository) findRecentManualProviderTopupForScrape(ctx context.Context, tx *sql.Tx, bankID int64, provider string, amount int64, mutationAt *time.Time) (string, error) {
	provider = strings.TrimSpace(strings.ToLower(provider))
	if bankID <= 0 || provider == "" || amount <= 0 {
		return "", nil
	}
	loc, locErr := time.LoadLocation("Asia/Jakarta")
	if locErr != nil {
		loc = time.FixedZone("WIB", 7*60*60)
	}
	anchor := time.Now().In(loc)
	if mutationAt != nil && !mutationAt.IsZero() {
		anchor = mutationAt.In(loc)
	}
	start := anchor.Add(-72 * time.Hour)
	end := anchor.Add(72 * time.Hour)

	var refID string
	err := tx.QueryRowContext(ctx, `
SELECT mdp.ref_id
FROM public.mutasi_dompet_provider mdp
JOIN public.mutasi_bank mb
  ON mb.bank_id = mdp.bank_id
 AND mb.ref_id = mdp.ref_id
 AND mb.jumlah = mdp.jumlah
WHERE mdp.provider = $1
  AND mdp.bank_id = $2
  AND mdp.jumlah = $3
  AND mdp.arah = 'credit'
  AND mdp.alasan = 'BANK_TRANSFER_IN'
  AND mdp.ref_id LIKE 'PDEP-%'
  AND COALESCE(mdp.meta->>'matched_k24_ref', '') = ''
  AND mb.arah = 'DEBIT'
  AND mb.alasan = 'BANK_TRANSFER_TO_PROVIDER'
  AND lower(trim(COALESCE(mb.provider, ''))) = $1
  AND mb.dibuat_pada >= $4
  AND mb.dibuat_pada <= $5
ORDER BY abs(extract(epoch FROM (COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) - $6::timestamptz))) ASC, mdp.id DESC
LIMIT 1
`, provider, bankID, amount, start, end, anchor).Scan(&refID)
	if err == nil {
		return strings.TrimSpace(refID), nil
	}
	if err == sql.ErrNoRows {
		return "", nil
	}
	return "", err
}

func (r *BankRepository) creditInternalBankFromScrapedMutation(ctx context.Context, tx *sql.Tx, actorID, sourceBankID int64, sourceBankName, sourceBankAccount string, destination *internalBankDestination, refID string, amount int64, originalNote string, mutationAtValue any, sender, receiver string) error {
	if destination == nil {
		return nil
	}
	var (
		destinationID      int64
		destinationName    string
		destinationAccount string
		destinationOwner   string
		destinationBefore  int64
		destinationActive  bool
	)
	if err := tx.QueryRowContext(ctx, `
SELECT id, nama, nomor_rekening, COALESCE(atas_nama, ''), saldo, aktif
FROM public.bank
WHERE id = $1
LIMIT 1
FOR UPDATE
`, destination.id).Scan(&destinationID, &destinationName, &destinationAccount, &destinationOwner, &destinationBefore, &destinationActive); err != nil {
		return err
	}
	if sourceBankID == destinationID {
		return nil
	}

	var existingDestinationID int64
	existingErr := tx.QueryRowContext(ctx, `
SELECT id
FROM public.mutasi_bank
WHERE bank_id = $1
  AND ref_id = $2
  AND arah = 'CREDIT'
  AND alasan = 'BANK_TRANSFER_IN'
ORDER BY id DESC
LIMIT 1
FOR UPDATE
`, destinationID, refID).Scan(&existingDestinationID)
	if existingErr == nil {
		return nil
	}
	if existingErr != sql.ErrNoRows {
		return existingErr
	}

	destinationAfter := destinationBefore + amount
	if _, err := tx.ExecContext(ctx, `
UPDATE public.bank
SET saldo = $2, diubah_pada = now()
WHERE id = $1
`, destinationID, destinationAfter); err != nil {
		return err
	}

	metaJSON, _ := json.Marshal(map[string]any{
		"type":                     "bank_internal_transfer",
		"source":                   "bank_scrape",
		"source_bank_id":           sourceBankID,
		"source_bank_name":         strings.TrimSpace(sourceBankName),
		"source_account":           strings.TrimSpace(sourceBankAccount),
		"destination_bank_id":      destinationID,
		"destination_bank_name":    strings.TrimSpace(destinationName),
		"destination_account":      strings.TrimSpace(destinationAccount),
		"destination_account_name": strings.TrimSpace(destinationOwner),
		"destination_active":       destinationActive,
	})

	note := fmt.Sprintf("Auto transfer masuk dari %s %s ke %s %s", strings.TrimSpace(sourceBankName), strings.TrimSpace(sourceBankAccount), strings.TrimSpace(destinationName), strings.TrimSpace(destinationAccount))
	if trimmed := strings.TrimSpace(originalNote); trimmed != "" {
		note += " | mutasi sumber: " + trimmed
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO public.mutasi_bank
  (bank_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, provider, diubah_oleh, dibuat_pada, meta, waktu_mutasi_bank, pengirim, penerima)
VALUES
  ($1,$2,'CREDIT',$3,'BANK_TRANSFER_IN',NULLIF($4,''),$5,$6,$7,$8,now(),$9::jsonb,$10,NULLIF($11,''),NULLIF($12,''))
`, destinationID, refID, amount, note, destinationBefore, destinationAfter, strings.TrimSpace(sourceBankName), actorID, string(metaJSON), mutationAtValue, sender, receiver); err != nil {
		return err
	}

	return nil
}

func (r *BankRepository) findInternalBankDestinationByAccount(ctx context.Context, tx *sql.Tx, sourceBankID int64, account string) (*internalBankDestination, error) {
	accountDigits := bankDigitsOnly(account)
	if accountDigits == "" {
		return nil, nil
	}
	rows, err := tx.QueryContext(ctx, `
SELECT id, nama, nomor_rekening, COALESCE(atas_nama, ''), aktif
FROM public.bank
WHERE id <> $1
ORDER BY id ASC
`, sourceBankID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		item := internalBankDestination{}
		if err := rows.Scan(&item.id, &item.name, &item.account, &item.owner, &item.active); err != nil {
			return nil, err
		}
		item.digits = bankDigitsOnly(item.account)
		if item.digits == accountDigits {
			return &item, nil
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *BankRepository) findInternalBankDestinationFromScrapedDebit(ctx context.Context, tx *sql.Tx, sourceBankID int64, note, sender, receiver string) (*internalBankDestination, error) {
	searchText := strings.Join([]string{note, sender, receiver}, " ")
	preferredDigits := bankDigitsOnly(strings.Join([]string{
		receiver,
		bankMutationNoteField(note, "penerima"),
		bankMutationNoteField(note, "receiver"),
	}, " "))
	allDigits := bankDigitsOnly(searchText)
	normalizedText := bankNormalizeProviderName(searchText)
	if preferredDigits == "" && allDigits == "" && normalizedText == "" {
		return nil, nil
	}

	rows, err := tx.QueryContext(ctx, `
SELECT id, nama, nomor_rekening, COALESCE(atas_nama, ''), aktif
FROM public.bank
WHERE id <> $1
ORDER BY id ASC
`, sourceBankID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var best *internalBankDestination
	bestPriority := 99
	var bestByName *internalBankDestination
	bestNameScore := 0
	nameAmbiguous := false
	exactReceiverName := bankInternalExactReceiverNameFromScrapedDebit(note, sender, receiver)
	for rows.Next() {
		item := internalBankDestination{}
		if err := rows.Scan(&item.id, &item.name, &item.account, &item.owner, &item.active); err != nil {
			return nil, err
		}
		item.digits = bankDigitsOnly(item.account)
		if len(item.digits) >= 8 {
			priority := 99
			if preferredDigits != "" && strings.Contains(preferredDigits, item.digits) {
				priority = 0
			} else if allDigits != "" && strings.Contains(allDigits, item.digits) {
				priority = 1
			}
			if priority != 99 && (best == nil ||
				priority < bestPriority ||
				(priority == bestPriority && item.active && !best.active) ||
				(priority == bestPriority && item.active == best.active && len(item.digits) > len(best.digits)) ||
				(priority == bestPriority && item.active == best.active && len(item.digits) == len(best.digits) && item.id < best.id)) {
				cp := item
				best = &cp
				bestPriority = priority
			}
		}

		score := bankInternalBankOwnerNameMatchScore(item.owner, normalizedText)
		if exactScore := bankInternalBankOwnerExactReceiverMatchScore(item.owner, exactReceiverName); exactScore > score {
			score = exactScore
		}
		if score > 0 {
			if score > bestNameScore {
				cp := item
				bestByName = &cp
				bestNameScore = score
				nameAmbiguous = false
			} else if score == bestNameScore && bestByName != nil && item.id != bestByName.id {
				nameAmbiguous = true
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if best != nil {
		return best, nil
	}
	if bestByName != nil && !nameAmbiguous {
		return bestByName, nil
	}
	return nil, nil
}

func (r *BankRepository) findSemanticDuplicateMutation(ctx context.Context, tx *sql.Tx, bankID, amount int64, arahMutasi, reason, semanticKey, legacySemanticKey string, mutationAt *time.Time, _, _ string, actualBalance int64) (id int64, refID string, saldo int64, found bool, err error) {
	var mutationAtValue any
	if mutationAt != nil {
		mutationAtValue = *mutationAt
	}
	rows, err := tx.QueryContext(ctx, `
SELECT
  id,
  COALESCE(ref_id, ''),
  saldo_sesudah,
  COALESCE(catatan, ''),
  COALESCE(meta->>'semantic_key', ''),
  COALESCE(meta->>'semantic_key_legacy', ''),
  waktu_mutasi_bank,
  COALESCE(pengirim, ''),
  COALESCE(penerima, '')
FROM public.mutasi_bank
WHERE bank_id = $1
  AND jumlah = $2
  AND arah = $3
  AND (
    alasan = $4
    OR ($3 = 'DEBIT' AND alasan IN ('BANK_MANUAL_OUT', 'BANK_TRANSFER_OUT', 'BANK_TRANSFER_TO_PROVIDER', 'BANK_TRANSFER_ADMIN_FEE'))
  )
  AND (
    COALESCE(meta->>'semantic_key', '') = $5
    OR COALESCE(meta->>'semantic_key_legacy', '') = $6
    OR ($7::timestamptz IS NOT NULL AND waktu_mutasi_bank = $7)
    OR catatan ILIKE '%date:%'
  )
ORDER BY id DESC
LIMIT 2000
`, bankID, amount, arahMutasi, reason, semanticKey, legacySemanticKey, mutationAtValue)
	if err != nil {
		return 0, "", 0, false, err
	}
	defer rows.Close()

	direction := "credit"
	if arahMutasi == "DEBIT" {
		direction = "debit"
	}
	for rows.Next() {
		var candidateID int64
		var candidateRef string
		var candidateSaldo int64
		var candidateNote string
		var candidateKey string
		var candidateLegacyKey string
		var candidateMutationAt sql.NullTime
		var candidateSender string
		var candidateReceiver string
		if err := rows.Scan(&candidateID, &candidateRef, &candidateSaldo, &candidateNote, &candidateKey, &candidateLegacyKey, &candidateMutationAt, &candidateSender, &candidateReceiver); err != nil {
			return 0, "", 0, false, err
		}
		if semanticKey != "" && candidateKey == semanticKey {
			return candidateID, candidateRef, candidateSaldo, true, nil
		}
		var candidateAt *time.Time
		if candidateMutationAt.Valid {
			candidateAt = &candidateMutationAt.Time
		}
		parsedKey, parsedLegacyKey, ok := bankMutationSemanticKeys(bankID, amount, direction, candidateNote, candidateAt, candidateSender, candidateReceiver)
		if ok && semanticKey != "" && parsedKey == semanticKey {
			return candidateID, candidateRef, candidateSaldo, true, nil
		}
		if legacySemanticKey != "" {
			candidateHasParty := bankMutationHasParty(candidateNote, candidateSender, candidateReceiver)
			inputHasParty := semanticKey != "" && legacySemanticKey != "" && semanticKey != legacySemanticKey
			legacyMatches := candidateLegacyKey == legacySemanticKey ||
				candidateKey == legacySemanticKey ||
				parsedLegacyKey == legacySemanticKey
			if legacyMatches && bankMutationLegacyDuplicateAllowed(candidateHasParty, inputHasParty, candidateSaldo, actualBalance) {
				return candidateID, candidateRef, candidateSaldo, true, nil
			}
		}
	}
	if err := rows.Err(); err != nil {
		return 0, "", 0, false, err
	}
	return 0, "", 0, false, nil
}

func bankMutationLegacyDuplicateAllowed(candidateHasParty, inputHasParty bool, candidateSaldo, actualBalance int64) bool {
	if !candidateHasParty || !inputHasParty {
		return true
	}
	return actualBalance > 0 && candidateSaldo == actualBalance
}

var bankMutationNoteDatePattern = regexp.MustCompile(`(?i)(?:^|\|)\s*date:\s*([^|]+)`)

func bankMutationSemanticKey(bankID, amount int64, direction, note string) (string, bool) {
	_, legacyKey, ok := bankMutationSemanticKeys(bankID, amount, direction, note, nil, "", "")
	return legacyKey, ok
}

func bankMutationSemanticKeys(bankID, amount int64, direction, note string, mutationAt *time.Time, sender, receiver string) (string, string, bool) {
	if bankID <= 0 || amount <= 0 {
		return "", "", false
	}
	direction = strings.ToLower(strings.TrimSpace(direction))
	switch {
	case direction == "credit" || direction == "c" || direction == "cr" || strings.Contains(direction, "kredit"):
		direction = "credit"
	case direction == "debit" || direction == "debet" || direction == "d" || direction == "db":
		direction = "debit"
	default:
		return "", "", false
	}
	txTime, ok := bankMutationSemanticTime(note, mutationAt)
	if !ok {
		return "", "", false
	}
	legacyKey := fmt.Sprintf("%d|%s|%d|%s", bankID, direction, amount, txTime.Format("2006-01-02 15:04:05"))
	senderKey := normalizeBankMutationParty(sender)
	if senderKey == "" {
		senderKey = normalizeBankMutationParty(bankMutationNoteField(note, "pengirim"))
	}
	if senderKey == "" {
		senderKey = normalizeBankMutationParty(bankMutationNoteField(note, "desc"))
	}
	receiverKey := normalizeBankMutationParty(receiver)
	if receiverKey == "" {
		receiverKey = normalizeBankMutationParty(bankMutationNoteField(note, "penerima"))
	}
	if receiverKey == "" {
		receiverKey = normalizeBankMutationParty(bankMutationNoteField(note, "rekening_provider"))
	}
	if senderKey == "" && receiverKey == "" {
		return bankMutationSemanticKeyWithBalance(legacyKey, note, *txTime), legacyKey, true
	}
	return bankMutationSemanticKeyWithBalance(fmt.Sprintf("%s|%s|%s", legacyKey, senderKey, receiverKey), note, *txTime), legacyKey, true
}

func bankMutationSemanticKeyWithBalance(baseKey, note string, txTime time.Time) string {
	if bankMutationHasClock(txTime) {
		return baseKey
	}
	balanceKey := bankMutationNoteBalanceKey(note)
	if balanceKey == "" {
		return baseKey
	}
	return baseKey + "|balance:" + balanceKey
}

func bankMutationNoteBalanceKey(note string) string {
	value := bankMutationNoteField(note, "balance")
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = regexp.MustCompile(`[,.]\d{2}$`).ReplaceAllString(value, "")
	digits := regexp.MustCompile(`\D`).ReplaceAllString(value, "")
	digits = strings.TrimLeft(digits, "0")
	if digits == "" {
		return ""
	}
	return digits
}

func normalizedBankMutationTime(value *time.Time) *time.Time {
	if value == nil || value.IsZero() {
		return nil
	}
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*60*60)
	}
	normalized := value.In(loc).Truncate(time.Second)
	return &normalized
}

func bankMutationSemanticTime(note string, mutationAt *time.Time) (*time.Time, bool) {
	if txTime := normalizedBankMutationTime(mutationAt); txTime != nil && bankMutationHasClock(*txTime) {
		return txTime, true
	}
	if parsed, ok := bankMutationTimeFromNote(note); ok {
		return &parsed, true
	}
	return nil, false
}

func bankMutationHasClock(value time.Time) bool {
	return value.Hour() != 0 || value.Minute() != 0 || value.Second() != 0 || value.Nanosecond() != 0
}

func bankMutationHasParty(note, sender, receiver string) bool {
	return normalizeBankMutationParty(sender) != "" ||
		normalizeBankMutationParty(receiver) != "" ||
		normalizeBankMutationParty(bankMutationNoteField(note, "pengirim")) != "" ||
		normalizeBankMutationParty(bankMutationNoteField(note, "desc")) != "" ||
		normalizeBankMutationParty(bankMutationNoteField(note, "penerima")) != "" ||
		normalizeBankMutationParty(bankMutationNoteField(note, "rekening_provider")) != ""
}

func normalizeBankMutationParty(value string) string {
	value = strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), " "))
	value = strings.Trim(value, " .,-")
	return value
}

func bankMutationNoteField(note, field string) string {
	field = strings.ToLower(strings.TrimSpace(field))
	if field == "" {
		return ""
	}
	for _, part := range strings.Split(note, "|") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToLower(part), field+":") {
			return strings.TrimSpace(part[len(field)+1:])
		}
	}
	return ""
}

func bankMutationTimeFromNote(note string) (time.Time, bool) {
	match := bankMutationNoteDatePattern.FindStringSubmatch(note)
	if len(match) < 2 {
		return time.Time{}, false
	}
	value := strings.Join(strings.Fields(strings.TrimSpace(match[1])), " ")
	if value == "" {
		return time.Time{}, false
	}
	replacer := strings.NewReplacer(
		"Januari", "January", "januari", "January",
		"Februari", "February", "februari", "February",
		"Maret", "March", "maret", "March",
		"April", "April", "april", "April",
		"Mei", "May", "mei", "May",
		"Juni", "June", "juni", "June",
		"Juli", "July", "juli", "July",
		"Agustus", "August", "agustus", "August",
		"September", "September", "september", "September",
		"Oktober", "October", "oktober", "October",
		"November", "November", "november", "November",
		"Desember", "December", "desember", "December",
	)
	value = replacer.Replace(value)
	loc := time.FixedZone("WIB", 7*60*60)
	for _, layout := range []string{
		"2 January 2006 15:04:05",
		"02 January 2006 15:04:05",
		"2 Jan 2006 15:04:05",
		"02 Jan 2006 15:04:05",
		"2/1/2006 15:04:05",
		"02/01/2006 15:04:05",
		"2-Jan-2006 15:04:05",
		"02-Jan-2006 15:04:05",
		"2 January 2006",
		"02 January 2006",
		"2 Jan 2006",
		"02 Jan 2006",
		"2/1/2006",
		"02/01/2006",
		"2-Jan-2006",
		"02-Jan-2006",
	} {
		if t, err := time.ParseInLocation(layout, value, loc); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func (r *BankRepository) TransferOut(ctx context.Context, actorID, bankID, amount int64, tujuan, note, refID string) (after int64, err error) {
	if amount <= 0 {
		return 0, errors.New("amount must be > 0")
	}
	tujuan = strings.TrimSpace(tujuan)
	if tujuan == "" {
		return 0, errors.New("tujuan required")
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var before int64
	var bankName string
	if err = tx.QueryRowContext(ctx, `
SELECT saldo, nama
FROM public.bank
WHERE id = $1
FOR UPDATE
`, bankID).Scan(&before, &bankName); err != nil {
		return 0, err
	}
	if before < amount {
		return 0, errors.New("saldo bank tidak cukup")
	}
	after = before - amount

	if _, err = tx.ExecContext(ctx, `
UPDATE public.bank
SET saldo = $2, diubah_pada = now()
WHERE id = $1
`, bankID, after); err != nil {
		return 0, err
	}

	metaJSON, _ := json.Marshal(map[string]any{
		"bank_id":   bankID,
		"bank_name": bankName,
		"tujuan":    tujuan,
		"type":      "bank_transfer_out",
	})

	if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_bank
  (bank_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, provider, diubah_oleh, dibuat_pada, meta)
VALUES
  ($1,$2,'DEBIT',$3,'BANK_TRANSFER_OUT',NULLIF($4,''),$5,$6,$7,$8,now(),$9::jsonb)
`, bankID, refID, amount, note, before, after, tujuan, actorID, string(metaJSON)); err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return after, nil
}

func (r *BankRepository) TransferToBCAOperational(ctx context.Context, actorID, sourceBankID, amount int64, adminFee int64, note, refID string) (sourceAfter int64, destinationAfter int64, destinationName string, destinationAccount string, err error) {
	if sourceBankID <= 0 {
		return 0, 0, "", "", errors.New("bank_id invalid")
	}
	if amount <= 0 {
		return 0, 0, "", "", errors.New("amount must be > 0")
	}
	if adminFee < 0 {
		return 0, 0, "", "", errors.New("admin fee invalid")
	}
	totalDebit := amount + adminFee
	if totalDebit < amount {
		return 0, 0, "", "", errors.New("total debit invalid")
	}

	const destinationAccountNumber = "3432738881"

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, 0, "", "", err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	type lockedBank struct {
		id      int64
		name    string
		account string
		saldo   int64
		active  bool
	}

	var source, destination *lockedBank
	rows, err := tx.QueryContext(ctx, `
SELECT id, nama, nomor_rekening, saldo, aktif
FROM public.bank
WHERE id = $1 OR trim(nomor_rekening) = $2
ORDER BY id
FOR UPDATE
`, sourceBankID, destinationAccountNumber)
	if err != nil {
		return 0, 0, "", "", err
	}
	defer rows.Close()

	for rows.Next() {
		item := lockedBank{}
		if err = rows.Scan(&item.id, &item.name, &item.account, &item.saldo, &item.active); err != nil {
			return 0, 0, "", "", err
		}
		if item.id == sourceBankID {
			cp := item
			source = &cp
		}
		if strings.TrimSpace(item.account) == destinationAccountNumber {
			cp := item
			destination = &cp
		}
	}
	if err = rows.Err(); err != nil {
		return 0, 0, "", "", err
	}
	if err = rows.Close(); err != nil {
		return 0, 0, "", "", err
	}
	if source == nil {
		return 0, 0, "", "", errors.New("bank sumber tidak ditemukan")
	}
	if destination == nil {
		return 0, 0, "", "", errors.New("BCA OPERASIONAL 3432738881 tidak ditemukan")
	}
	if source.id == destination.id {
		return 0, 0, "", "", errors.New("bank sumber tidak boleh BCA OPERASIONAL")
	}
	if !source.active {
		return 0, 0, "", "", errors.New("bank sumber tidak aktif")
	}
	if !destination.active {
		return 0, 0, "", "", errors.New("BCA OPERASIONAL tidak aktif")
	}
	if source.saldo < totalDebit {
		return 0, 0, "", "", errors.New("saldo bank sumber tidak cukup")
	}

	sourceAfterTransfer := source.saldo - amount
	sourceAfter = sourceAfterTransfer - adminFee
	destinationAfter = destination.saldo + amount

	if _, err = tx.ExecContext(ctx, `
UPDATE public.bank
SET saldo = $2, diubah_pada = now()
WHERE id = $1
`, source.id, sourceAfter); err != nil {
		return 0, 0, "", "", err
	}
	if _, err = tx.ExecContext(ctx, `
UPDATE public.bank
SET saldo = $2, diubah_pada = now()
WHERE id = $1
`, destination.id, destinationAfter); err != nil {
		return 0, 0, "", "", err
	}

	metaJSON, _ := json.Marshal(map[string]any{
		"type":                  "bank_internal_transfer",
		"amount":                amount,
		"admin_fee":             adminFee,
		"source_bank_id":        source.id,
		"source_bank_name":      strings.TrimSpace(source.name),
		"source_account":        strings.TrimSpace(source.account),
		"destination_bank_id":   destination.id,
		"destination_bank_name": strings.TrimSpace(destination.name),
		"destination_account":   destinationAccountNumber,
	})

	if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_bank
  (bank_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, provider, diubah_oleh, dibuat_pada, meta)
VALUES
  ($1,$2,'DEBIT',$3,'BANK_TRANSFER_OUT',NULLIF($4,''),$5,$6,$7,$8,now(),($9::jsonb || '{"entry":"source_transfer"}'::jsonb))
`, source.id, refID, amount, note, source.saldo, sourceAfterTransfer, destinationAccountNumber, actorID, string(metaJSON)); err != nil {
		return 0, 0, "", "", err
	}

	if adminFee > 0 {
		if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_bank
  (bank_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, provider, diubah_oleh, dibuat_pada, meta)
VALUES
  ($1,$2,'DEBIT',$3,'BANK_TRANSFER_ADMIN_FEE','biaya admin transfer ke 3432738881',$4,$5,$6,$7,now(),($8::jsonb || '{"entry":"source_admin_fee"}'::jsonb))
`, source.id, refID, adminFee, sourceAfterTransfer, sourceAfter, destinationAccountNumber, actorID, string(metaJSON)); err != nil {
			return 0, 0, "", "", err
		}
	}

	if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_bank
  (bank_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, provider, diubah_oleh, dibuat_pada, meta)
VALUES
  ($1,$2,'CREDIT',$3,'BANK_TRANSFER_IN',$4,$5,$6,$7,$8,now(),($9::jsonb || '{"entry":"destination_receive"}'::jsonb))
`, destination.id, refID, amount, "Transfer masuk dari "+strings.TrimSpace(source.name)+" "+strings.TrimSpace(source.account), destination.saldo, destinationAfter, strings.TrimSpace(source.name), actorID, string(metaJSON)); err != nil {
		return 0, 0, "", "", err
	}

	if err = tx.Commit(); err != nil {
		return 0, 0, "", "", err
	}
	return sourceAfter, destinationAfter, strings.TrimSpace(destination.name), destinationAccountNumber, nil
}

func (r *BankRepository) TransferToProvider(ctx context.Context, actorID, bankID, amount int64, adminFee int64, provider, note, refID string) (bankAfter int64, providerAfter int64, err error) {
	if amount <= 0 {
		return 0, 0, errors.New("amount must be > 0")
	}
	if adminFee < 0 {
		return 0, 0, errors.New("admin fee invalid")
	}
	totalDebit := amount + adminFee
	if totalDebit < amount {
		return 0, 0, errors.New("total debit invalid")
	}
	provider, err = resolveProviderName(ctx, r.db, provider)
	if err != nil {
		return 0, 0, err
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var bankBefore int64
	var bankName string
	if err = tx.QueryRowContext(ctx, `
SELECT saldo, nama
FROM public.bank
WHERE id = $1
FOR UPDATE
`, bankID).Scan(&bankBefore, &bankName); err != nil {
		return 0, 0, err
	}
	if bankBefore < totalDebit {
		return 0, 0, errors.New("saldo bank tidak cukup")
	}
	bankAfterTransfer := bankBefore - amount
	bankAfter = bankBefore - totalDebit

	if _, err = tx.ExecContext(ctx, `
UPDATE public.bank
SET saldo = $2, diubah_pada = now()
WHERE id = $1
`, bankID, bankAfter); err != nil {
		return 0, 0, err
	}

	metaJSON, _ := json.Marshal(map[string]any{
		"admin_fee": adminFee,
		"provider":  provider,
		"type":      "bank_transfer_to_provider",
	})

	if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_bank
  (bank_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, provider, diubah_oleh, dibuat_pada, meta)
VALUES
  ($1,$2,'DEBIT',$3,'BANK_TRANSFER_TO_PROVIDER',NULLIF($4,''),$5,$6,$7,$8,now(),$9::jsonb)
`, bankID, refID, amount, note, bankBefore, bankAfterTransfer, provider, actorID, string(metaJSON)); err != nil {
		return 0, 0, err
	}

	if adminFee > 0 {
		adminMetaJSON, _ := json.Marshal(map[string]any{
			"provider": provider,
			"type":     "bank_transfer_admin_fee",
		})
		if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_bank
  (bank_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, provider, diubah_oleh, dibuat_pada, meta)
VALUES
  ($1,$2,'DEBIT',$3,'BANK_TRANSFER_ADMIN_FEE','biaya admin',$4,$5,$6,$7,now(),$8::jsonb)
`, bankID, refID, adminFee, bankAfterTransfer, bankAfter, provider, actorID, string(adminMetaJSON)); err != nil {
			return 0, 0, err
		}
	}

	if _, err = tx.ExecContext(ctx, `
INSERT INTO public.dompet_provider (provider, saldo)
VALUES ($1, 0)
ON CONFLICT (provider) DO NOTHING
`, provider); err != nil {
		return 0, 0, err
	}

	var providerBefore int64
	if err = tx.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_provider
WHERE provider = $1
FOR UPDATE
`, provider).Scan(&providerBefore); err != nil {
		return 0, 0, err
	}
	providerAfter = providerBefore + amount

	if _, err = tx.ExecContext(ctx, `
UPDATE public.dompet_provider
SET saldo = $2, diperbarui_pada = now()
WHERE provider = $1
`, provider, providerAfter); err != nil {
		return 0, 0, err
	}

	if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet_provider
  (provider, bank_id, bank_nama, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, diubah_oleh, dibuat_pada, meta)
VALUES
  ($1,$2,$3,$4,'credit',$5,'BANK_TRANSFER_IN',NULLIF($6,''),$7,$8,$9,now(),$10::jsonb)
`, provider, bankID, strings.TrimSpace(bankName), refID, amount, note, providerBefore, providerAfter, actorID, string(metaJSON)); err != nil {
		return 0, 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, 0, err
	}
	return bankAfter, providerAfter, nil
}

func (r *BankRepository) CreditProviderFromBankMutation(ctx context.Context, actorID, mutasiBankID int64, provider, note string, includeAdminStaffOnly bool) (result *BankProviderAssignResult, err error) {
	if actorID <= 0 {
		return nil, errors.New("actor invalid")
	}
	if mutasiBankID <= 0 {
		return nil, errors.New("mutasi_bank_id invalid")
	}
	provider, err = resolveProviderName(ctx, r.db, provider)
	if err != nil {
		return nil, err
	}
	note = strings.TrimSpace(note)

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var (
		bankID        int64
		bankNama      string
		refID         string
		arah          string
		amount        int64
		alasan        string
		catatan       string
		currentTarget sql.NullString
		memberID      sql.NullInt64
	)
	if err = tx.QueryRowContext(ctx, `
SELECT mb.bank_id, b.nama, mb.ref_id, mb.arah, mb.jumlah, mb.alasan, COALESCE(mb.catatan,''), mb.provider, mb.member_id
FROM public.mutasi_bank mb
JOIN public.bank b ON b.id = mb.bank_id
WHERE mb.id = $1
  AND ($2::boolean OR COALESCE(b.admin_staff_only, false) = false)
FOR UPDATE OF mb
`, mutasiBankID, includeAdminStaffOnly).Scan(&bankID, &bankNama, &refID, &arah, &amount, &alasan, &catatan, &currentTarget, &memberID); err != nil {
		return nil, err
	}
	if strings.ToUpper(strings.TrimSpace(arah)) != "DEBIT" {
		return nil, errors.New("hanya mutasi debit yang bisa masuk saldo provider")
	}
	refID = strings.TrimSpace(refID)
	if refID == "" {
		return nil, errors.New("ref_id mutasi bank kosong")
	}
	if amount <= 0 {
		return nil, errors.New("jumlah mutasi invalid")
	}
	if isBankAdminFeeMutation(alasan, catatan) {
		return nil, errors.New("biaya admin tidak perlu target provider")
	}
	if memberID.Valid {
		return nil, errors.New("mutasi bank sudah terkait member")
	}
	if currentTarget.Valid && strings.TrimSpace(currentTarget.String) != "" {
		return nil, errors.New("mutasi bank sudah punya target provider/tujuan")
	}

	var existingProvider string
	existingErr := tx.QueryRowContext(ctx, `
SELECT provider
FROM public.mutasi_dompet_provider
WHERE ref_id = $1
  AND arah = 'credit'
  AND alasan = 'BANK_TRANSFER_IN'
ORDER BY id DESC
LIMIT 1
FOR UPDATE
`, refID).Scan(&existingProvider)
	if existingErr == nil {
		return nil, fmt.Errorf("ref_id %s sudah pernah masuk ledger provider %s", refID, existingProvider)
	}
	if existingErr != sql.ErrNoRows {
		return nil, existingErr
	}

	if _, err = tx.ExecContext(ctx, `
INSERT INTO public.dompet_provider (provider, saldo)
VALUES ($1, 0)
ON CONFLICT (provider) DO NOTHING
`, provider); err != nil {
		return nil, err
	}

	var providerBefore int64
	if err = tx.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_provider
WHERE provider = $1
FOR UPDATE
`, provider).Scan(&providerBefore); err != nil {
		return nil, err
	}
	providerAfter := providerBefore + amount
	if _, err = tx.ExecContext(ctx, `
UPDATE public.dompet_provider
SET saldo = $2, diperbarui_pada = now()
WHERE provider = $1
`, provider, providerAfter); err != nil {
		return nil, err
	}

	metaJSON, _ := json.Marshal(map[string]any{
		"type":                       "manual_bank_mutation_provider_assign",
		"source":                     "dashboard_mutasi_bank",
		"mutasi_bank_id":             mutasiBankID,
		"provider":                   provider,
		"previous_bank_reason":       alasan,
		"manual_note":                note,
		"original_bank_note_excerpt": catatan,
	})

	ledgerNote := strings.TrimSpace(note)
	assignNote := fmt.Sprintf("Manual assign mutasi bank %s ke provider %s", refID, provider)
	if ledgerNote != "" {
		ledgerNote += " | " + assignNote
	} else {
		ledgerNote = assignNote
	}
	if strings.TrimSpace(catatan) != "" {
		ledgerNote += " | mutasi: " + strings.TrimSpace(catatan)
	}

	if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet_provider
  (provider, bank_id, bank_nama, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, diubah_oleh, dibuat_pada, meta)
VALUES
  ($1,$2,$3,$4,'credit',$5,'BANK_TRANSFER_IN',NULLIF($6,''),$7,$8,$9,now(),$10::jsonb)
`, provider, bankID, strings.TrimSpace(bankNama), refID, amount, ledgerNote, providerBefore, providerAfter, actorID, string(metaJSON)); err != nil {
		return nil, err
	}

	if _, err = tx.ExecContext(ctx, `
UPDATE public.mutasi_bank
SET provider = $2::text,
    alasan = 'BANK_TRANSFER_TO_PROVIDER',
    diubah_oleh = $3::bigint,
    meta = COALESCE(meta, '{}'::jsonb) || jsonb_build_object(
      'manual_provider_assign',
      jsonb_build_object(
        'provider', $2::text,
        'assigned_by', $3::bigint,
        'assigned_at', now(),
        'previous_reason', $4::text,
        'manual_note', NULLIF($5::text, '')
      )
    )
WHERE id = $1
`, mutasiBankID, provider, actorID, alasan, note); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &BankProviderAssignResult{
		Provider:      provider,
		ProviderSaldo: providerAfter,
		Amount:        amount,
		RefID:         refID,
		BankID:        bankID,
		BankNama:      strings.TrimSpace(bankNama),
	}, nil
}

func isBankAdminFeeMutation(alasan, catatan string) bool {
	text := strings.ToUpper(strings.TrimSpace(alasan) + " " + strings.TrimSpace(catatan))
	return strings.Contains(text, "BANK_TRANSFER_ADMIN_FEE") ||
		strings.Contains(text, "ADMIN_FEE") ||
		strings.Contains(text, "ADMIN FEE") ||
		strings.Contains(text, "BIAYA ADMIN") ||
		strings.Contains(text, "BIAYA TXN") ||
		strings.Contains(text, "BIAYA TRANSAKSI") ||
		strings.Contains(text, "BIAYA TRANSFER") ||
		strings.Contains(text, "BIAYA POTONGAN") ||
		strings.Contains(text, "POTONGAN REKENING")
}

func isScrapedProviderTransferAdminFee(amount int64, direction, alasan, catatan, pengirim, penerima string) bool {
	if amount != 2500 && amount != 6500 {
		return false
	}
	if !isDebitBankDirection(direction) {
		return false
	}
	if isBankAdminFeeMutation(alasan, catatan) {
		return true
	}
	text := strings.ToUpper(strings.Join([]string{
		strings.TrimSpace(alasan),
		strings.TrimSpace(catatan),
		strings.TrimSpace(pengirim),
		strings.TrimSpace(penerima),
	}, " "))
	return strings.Contains(text, "SWITCHING DB BIAYA") ||
		strings.Contains(text, "ATMSTRPRM") ||
		strings.Contains(text, "BY TRX BIFAST") ||
		strings.Contains(text, "BI-FAST") ||
		strings.Contains(text, "BIFAST")
}

func isScrapedBCAOperationalTransfer(amount int64, direction, catatan, pengirim, penerima string) bool {
	if amount <= 0 || !isDebitBankDirection(direction) {
		return false
	}
	if isScrapedBCAOperationalTransferAdminFee(amount, direction, catatan, pengirim, penerima) {
		return false
	}
	return bankTextContainsAccount(bcaOperationalAccountNumber, catatan, pengirim, penerima) ||
		bankTextLooksLikeBCAOperationalName(catatan, pengirim, penerima)
}

func isScrapedBCAOperationalTransferAdminFee(amount int64, direction, catatan, pengirim, penerima string) bool {
	if !isDebitBankDirection(direction) || !bankTextContainsAccount(bcaOperationalAccountNumber, catatan, pengirim, penerima) {
		return false
	}
	if amount != 2500 && amount != 6500 {
		return false
	}
	text := strings.ToUpper(strings.Join([]string{
		strings.TrimSpace(catatan),
		strings.TrimSpace(pengirim),
		strings.TrimSpace(penerima),
	}, " "))
	return strings.Contains(text, "BFST") ||
		strings.Contains(text, "BI-FAST") ||
		strings.Contains(text, "BIFAST") ||
		strings.Contains(text, "BIAYA") ||
		strings.Contains(text, "ADMIN")
}

func isScrapedInternalBankTransferAdminFee(amount int64, direction string, destination *internalBankDestination, catatan, pengirim, penerima string) bool {
	if destination == nil || !isDebitBankDirection(direction) {
		return false
	}
	if amount != 2500 && amount != 6500 {
		return false
	}
	if !bankTextContainsAccount(destination.account, catatan, pengirim, penerima) {
		return false
	}
	text := strings.ToUpper(strings.Join([]string{
		strings.TrimSpace(catatan),
		strings.TrimSpace(pengirim),
		strings.TrimSpace(penerima),
	}, " "))
	return strings.Contains(text, "BFST") ||
		strings.Contains(text, "BI-FAST") ||
		strings.Contains(text, "BIFAST") ||
		strings.Contains(text, "BIAYA") ||
		strings.Contains(text, "ADMIN")
}

func isScrapedProviderRefundCredit(amount int64, direction, catatan, pengirim, penerima string) bool {
	if amount <= 0 {
		return false
	}
	direction = strings.ToLower(strings.TrimSpace(direction))
	if direction != "credit" && direction != "cr" && direction != "c" && !strings.Contains(direction, "kredit") {
		return false
	}
	text := bankNormalizeProviderName(strings.Join([]string{
		strings.TrimSpace(catatan),
		strings.TrimSpace(pengirim),
		strings.TrimSpace(penerima),
	}, " "))
	if text == "" {
		return false
	}
	return strings.Contains(" "+text+" ", " KOR ") ||
		strings.Contains(text, "KOREKSI") ||
		strings.Contains(text, "REFUND") ||
		strings.Contains(text, "RETUR") ||
		strings.Contains(text, "REVERSAL") ||
		strings.Contains(text, "DIKEMBALIKAN") ||
		strings.Contains(text, "PEMBATALAN")
}

func bankMutationWindow(mutationAt *time.Time, before, after time.Duration) (time.Time, time.Time, time.Time) {
	loc, locErr := time.LoadLocation("Asia/Jakarta")
	if locErr != nil {
		loc = time.FixedZone("WIB", 7*60*60)
	}
	anchor := time.Now().In(loc)
	if mutationAt != nil && !mutationAt.IsZero() {
		anchor = mutationAt.In(loc)
	}
	return anchor.Add(-before), anchor.Add(after), anchor
}

func bankTextLooksLikeBCAOperationalName(values ...string) bool {
	text := strings.ToUpper(strings.Join(values, " "))
	if !strings.Contains(text, "PULSA MITRA NASION") {
		return false
	}
	if strings.Contains(text, "CENAIDJA") {
		return true
	}
	if strings.Contains(text, "BANK RAKYAT") ||
		strings.Contains(text, " BRI ") ||
		strings.Contains(text, "MANDIRI") ||
		strings.Contains(text, " BNI ") ||
		strings.Contains(text, "OCBC") {
		return false
	}
	return strings.Contains(text, "TRSF E-BANKING DB") ||
		strings.Contains(text, "E-BANKING DB") ||
		strings.Contains(text, "FTSCY")
}

func isDebitBankDirection(direction string) bool {
	direction = strings.ToLower(strings.TrimSpace(direction))
	return direction == "debit" || direction == "debet" || direction == "d" || direction == "db"
}

var bankReferenceTokenPattern = regexp.MustCompile(`[A-Z0-9]{8,}`)

func bankReferenceTokens(values ...string) map[string]struct{} {
	tokens := make(map[string]struct{})
	for _, value := range values {
		normalized := strings.ToUpper(strings.TrimSpace(value))
		if normalized == "" {
			continue
		}
		for _, token := range bankReferenceTokenPattern.FindAllString(normalized, -1) {
			if !bankReferenceTokenHasDigit(token) {
				continue
			}
			tokens[token] = struct{}{}
		}
	}
	return tokens
}

func bankReferenceTokenHasDigit(token string) bool {
	hasDigit := false
	hasLetter := false
	for _, r := range token {
		if r >= '0' && r <= '9' {
			hasDigit = true
			continue
		}
		if r >= 'A' && r <= 'Z' {
			hasLetter = true
		}
	}
	return hasDigit && hasLetter
}

func bankReferenceTokensIntersect(left, right map[string]struct{}) bool {
	if len(left) == 0 || len(right) == 0 {
		return false
	}
	if len(left) > len(right) {
		left, right = right, left
	}
	for token := range left {
		if _, ok := right[token]; ok {
			return true
		}
	}
	return false
}

func bankProviderNameMatchScore(providerName, normalizedText string) int {
	normalizedName := bankNormalizeProviderName(providerName)
	if len(normalizedName) < 8 || normalizedText == "" {
		return 0
	}
	best := bankProviderNameVariantMatchScore(normalizedName, normalizedText)
	if trimmed := bankTrimProviderLegalPrefix(normalizedName); trimmed != normalizedName {
		if score := bankProviderNameVariantMatchScore(trimmed, normalizedText); score > best {
			best = score
		}
	}
	return best
}

func bankInternalBankOwnerNameMatchScore(ownerName, normalizedText string) int {
	normalizedOwner := bankNormalizeProviderName(ownerName)
	if len(normalizedOwner) < 8 || normalizedText == "" || bankInternalBankOwnerNameIsGeneric(normalizedOwner) {
		return 0
	}
	return bankProviderNameMatchScore(ownerName, normalizedText)
}

func bankInternalBankOwnerExactReceiverMatchScore(ownerName, exactReceiverName string) int {
	normalizedOwner := bankNormalizeProviderName(ownerName)
	if len(normalizedOwner) < 3 || exactReceiverName == "" || bankInternalBankOwnerNameIsReservedGeneric(normalizedOwner) {
		return 0
	}
	if normalizedOwner != exactReceiverName {
		return 0
	}
	return len(normalizedOwner) + 1000
}

func bankInternalBankOwnerNameIsGeneric(normalizedOwner string) bool {
	tokens := strings.Fields(normalizedOwner)
	if len(tokens) < 2 {
		return true
	}
	return bankInternalBankOwnerNameIsReservedGeneric(normalizedOwner)
}

func bankInternalBankOwnerNameIsReservedGeneric(normalizedOwner string) bool {
	switch normalizedOwner {
	case "PT PULSA", "PULSA MITRA NASIONAL", "PT PULSA MITRA NASIONAL", "PULSA MITRA NASIONAL PT":
		return true
	default:
		return false
	}
}

var bankBIFastReceiverPattern = regexp.MustCompile(`(?i)\b[A-Z0-9]{4,}IDJA/([^|]+?)(?:\s+ref\b|\s*\||$)`)

func bankInternalExactReceiverNameFromScrapedDebit(note, sender, receiver string) string {
	for _, value := range []string{
		receiver,
		bankMutationNoteField(note, "penerima"),
		bankMutationNoteField(note, "receiver"),
	} {
		if normalized := bankNormalizeProviderName(value); normalized != "" {
			return normalized
		}
	}
	for _, value := range []string{
		sender,
		bankMutationNoteField(note, "pengirim"),
		bankMutationNoteField(note, "desc"),
		note,
	} {
		if normalized := bankBIFastExactReceiverName(value); normalized != "" {
			return normalized
		}
	}
	return ""
}

func bankBIFastExactReceiverName(value string) string {
	match := bankBIFastReceiverPattern.FindStringSubmatch(value)
	if len(match) < 2 {
		return ""
	}
	return bankNormalizeProviderName(match[1])
}

func bankProviderNameVariantMatchScore(normalizedName, normalizedText string) int {
	if len(normalizedName) < 8 || normalizedText == "" {
		return 0
	}
	if strings.Contains(" "+normalizedText+" ", " "+normalizedName+" ") {
		return len(normalizedName)
	}
	compactName := strings.ReplaceAll(normalizedName, " ", "")
	compactText := strings.ReplaceAll(normalizedText, " ", "")
	if len(compactName) >= 8 && compactText != "" && strings.Contains(compactText, compactName) {
		return len(normalizedName)
	}
	nameTokens := strings.Fields(normalizedName)
	textTokens := strings.Fields(normalizedText)
	if len(nameTokens) < 2 {
		return 0
	}
	if len(textTokens) >= len(nameTokens) {
		for i := 0; i+len(nameTokens) <= len(textTokens); i++ {
			if bankProviderNameTokensMatch(nameTokens, textTokens[i:i+len(nameTokens)]) {
				return len(normalizedName)
			}
		}
	}
	if score := bankProviderNamePrefixMatchScore(nameTokens, textTokens); score > 0 {
		return score
	}
	return 0
}

func bankProviderNameTokensMatch(nameTokens, textTokens []string) bool {
	if len(nameTokens) != len(textTokens) || len(nameTokens) < 2 {
		return false
	}
	for i := 0; i < len(nameTokens)-1; i++ {
		if nameTokens[i] != textTokens[i] {
			return false
		}
	}
	nameLast := nameTokens[len(nameTokens)-1]
	textLast := textTokens[len(textTokens)-1]
	if nameLast == textLast {
		return true
	}
	shorter := nameLast
	longer := textLast
	if len(shorter) > len(longer) {
		shorter, longer = longer, shorter
	}
	return len(shorter) >= 5 && strings.HasPrefix(longer, shorter)
}

func bankProviderNamePrefixMatchScore(nameTokens, textTokens []string) int {
	if len(nameTokens) < 3 || len(textTokens) < 2 {
		return 0
	}
	maxPrefix := len(nameTokens) - 1
	if maxPrefix > len(textTokens) {
		maxPrefix = len(textTokens)
	}
	for prefixLen := maxPrefix; prefixLen >= 2; prefixLen-- {
		for i := 0; i+prefixLen <= len(textTokens); i++ {
			if bankProviderNameExactTokensMatch(nameTokens[:prefixLen], textTokens[i:i+prefixLen]) {
				return len(strings.Join(nameTokens[:prefixLen], " "))
			}
		}
	}
	return 0
}

func bankProviderNameExactTokensMatch(nameTokens, textTokens []string) bool {
	if len(nameTokens) != len(textTokens) || len(nameTokens) < 2 {
		return false
	}
	for i := range nameTokens {
		if nameTokens[i] != textTokens[i] {
			return false
		}
	}
	return len(nameTokens[len(nameTokens)-1]) >= 5
}

func bankTrimProviderLegalPrefix(normalizedName string) string {
	tokens := strings.Fields(normalizedName)
	for len(tokens) > 0 {
		switch tokens[0] {
		case "PT", "CV":
			tokens = tokens[1:]
			continue
		}
		break
	}
	return strings.Join(tokens, " ")
}

func bankNormalizeProviderName(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	var b strings.Builder
	lastSpace := true
	for _, r := range value {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastSpace = false
			continue
		}
		if !lastSpace {
			b.WriteByte(' ')
			lastSpace = true
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func bankTextContainsAccount(account string, values ...string) bool {
	account = bankDigitsOnly(account)
	if account == "" {
		return false
	}
	return strings.Contains(bankDigitsOnly(strings.Join(values, " ")), account)
}

func bankDigitsOnly(value string) string {
	var b strings.Builder
	for _, r := range value {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
