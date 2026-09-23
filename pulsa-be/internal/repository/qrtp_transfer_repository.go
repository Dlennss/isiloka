package repository

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
)

const defaultQRTPAdminFee int64 = 2500

func DefaultQRTPAdminFee() int64 {
	return defaultQRTPAdminFee
}

type QRTPTransferRow struct {
	ID                    int64           `json:"id"`
	RefID                 string          `json:"ref_id"`
	InquiryOrderID        string          `json:"inquiry_order_id"`
	Provider              string          `json:"provider"`
	ProviderRekeningID    int64           `json:"provider_rekening_id"`
	BankID                int64           `json:"bank_id"`
	BankCode              string          `json:"bank_code"`
	BankName              string          `json:"bank_name"`
	AccountNo             string          `json:"account_no"`
	AccountName           string          `json:"account_name"`
	Amount                int64           `json:"amount"`
	AdminFee              int64           `json:"admin_fee"`
	Note                  string          `json:"note"`
	Status                string          `json:"status"`
	InquiryStatus         string          `json:"inquiry_status"`
	AccountStatus         string          `json:"account_status"`
	InquiryPublicID       string          `json:"inquiry_public_id"`
	InquiryError          string          `json:"inquiry_error"`
	InquiryRaw            json.RawMessage `json:"inquiry_raw,omitempty"`
	PayoutPublicID        string          `json:"payout_public_id"`
	ProviderTransactionID string          `json:"provider_transaction_id"`
	PayoutError           string          `json:"payout_error"`
	PayoutReason          string          `json:"payout_reason"`
	PayoutRaw             json.RawMessage `json:"payout_raw,omitempty"`
	BankSaldoAfter        *int64          `json:"bank_saldo_after,omitempty"`
	ProviderSaldoAfter    *int64          `json:"provider_saldo_after,omitempty"`
	CreatedBy             int64           `json:"created_by"`
	CreatedByName         string          `json:"created_by_name"`
	ProcessedAt           *time.Time      `json:"processed_at,omitempty"`
	CompletedAt           *time.Time      `json:"completed_at,omitempty"`
	ReversedAt            *time.Time      `json:"reversed_at,omitempty"`
	CallbackAt            *time.Time      `json:"callback_at,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

type QRTPTransferLedgerRow struct {
	ID            int64      `json:"id"`
	Direction     string     `json:"direction"`
	LedgerType    string     `json:"ledger_type"`
	RefID         string     `json:"ref_id"`
	Provider      string     `json:"provider"`
	BankName      string     `json:"bank_name"`
	AccountNo     string     `json:"account_no"`
	AccountName   string     `json:"account_name"`
	Amount        int64      `json:"amount"`
	AdminFee      int64      `json:"admin_fee"`
	Status        string     `json:"status"`
	Reason        string     `json:"reason"`
	Note          string     `json:"note"`
	CreatedByName string     `json:"created_by_name"`
	SaldoSebelum  *int64     `json:"saldo_sebelum,omitempty"`
	SaldoSesudah  *int64     `json:"saldo_sesudah,omitempty"`
	ProcessedAt   *time.Time `json:"processed_at,omitempty"`
	CallbackAt    *time.Time `json:"callback_at,omitempty"`
	ReversedAt    *time.Time `json:"reversed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type QRTPInquiryInsert struct {
	RefID              string
	InquiryOrderID     string
	Provider           string
	ProviderRekeningID int64
	BankID             int64
	BankCode           string
	BankName           string
	AccountNo          string
	AccountName        string
	Amount             int64
	AdminFee           int64
	Note               string
	Status             string
	InquiryStatus      string
	AccountStatus      string
	InquiryPublicID    string
	InquiryError       string
	InquiryRaw         []byte
	CreatedBy          int64
}

type QRTPTransferRepository struct {
	db       *sql.DB
	bankRepo *BankRepository
}

func qrtpNullableActorID(actorID int64) any {
	if actorID > 0 {
		return actorID
	}
	return nil
}

func qrtpJSONText(raw []byte) string {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return "{}"
	}
	if json.Valid(trimmed) {
		return string(trimmed)
	}
	wrapped, err := json.Marshal(map[string]string{"raw": string(trimmed)})
	if err != nil {
		return "{}"
	}
	return string(wrapped)
}

func NewQRTPTransferRepository(db *sql.DB) *QRTPTransferRepository {
	return &QRTPTransferRepository{db: db, bankRepo: NewBankRepository(db)}
}

func (r *QRTPTransferRepository) QRTPBank(ctx context.Context) (*BankRow, error) {
	id, err := r.bankRepo.EnsureSystemQRTPBank(ctx)
	if err != nil {
		return nil, err
	}
	return r.bankRepo.Get(ctx, id)
}

func (r *QRTPTransferRepository) GetProviderRekening(ctx context.Context, id int64) (*ProviderRekeningRow, error) {
	return NewProviderRekeningRepository(r.db).Get(ctx, id)
}

func (r *QRTPTransferRepository) InsertInquiry(ctx context.Context, in QRTPInquiryInsert) (*QRTPTransferRow, error) {
	in.RefID = strings.TrimSpace(in.RefID)
	in.InquiryOrderID = strings.TrimSpace(in.InquiryOrderID)
	in.Provider = strings.TrimSpace(strings.ToLower(in.Provider))
	if in.AdminFee < 0 {
		in.AdminFee = defaultQRTPAdminFee
	}
	if in.Status == "" {
		in.Status = "inquiry_success"
	}
	raw := qrtpJSONText(in.InquiryRaw)
	var id int64
	err := r.db.QueryRowContext(ctx, `
INSERT INTO public.qrtp_provider_transfers
  (ref_id, inquiry_order_id, provider, provider_rekening_id, bank_id, bank_code, bank_name, account_no, account_name,
   amount, admin_fee, note, status, inquiry_status, account_status, inquiry_public_id, inquiry_error, inquiry_raw, created_by,
   created_at, updated_at)
VALUES
  ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18::jsonb,$19,now(),now())
RETURNING id
`, in.RefID, in.InquiryOrderID, in.Provider, in.ProviderRekeningID, in.BankID, in.BankCode, in.BankName, in.AccountNo, in.AccountName,
		in.Amount, in.AdminFee, in.Note, in.Status, in.InquiryStatus, in.AccountStatus, in.InquiryPublicID, in.InquiryError, raw, qrtpNullableActorID(in.CreatedBy)).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
}

func (r *QRTPTransferRepository) Get(ctx context.Context, id int64) (*QRTPTransferRow, error) {
	var item QRTPTransferRow
	if err := r.scanRow(r.db.QueryRowContext(ctx, qrtpTransferSelectSQL()+` WHERE t.id = $1 LIMIT 1`, id), &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *QRTPTransferRepository) GetByRefID(ctx context.Context, refID string) (*QRTPTransferRow, error) {
	var item QRTPTransferRow
	if err := r.scanRow(r.db.QueryRowContext(ctx, qrtpTransferSelectSQL()+` WHERE t.ref_id = $1 LIMIT 1`, strings.TrimSpace(refID)), &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *QRTPTransferRepository) List(ctx context.Context, provider, status, direction, search string, limit, offset int) ([]QRTPTransferLedgerRow, int64, error) {
	provider = strings.TrimSpace(strings.ToLower(provider))
	status = strings.TrimSpace(strings.ToLower(status))
	direction = strings.TrimSpace(strings.ToLower(direction))
	search = strings.TrimSpace(search)
	if direction != "" && direction != "in" && direction != "out" {
		return nil, 0, errors.New("direction must be in|out")
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	bankID, err := r.bankRepo.EnsureSystemQRTPBank(ctx)
	if err != nil {
		return nil, 0, err
	}
	args := []any{bankID}
	where := []string{"1=1"}
	if provider != "" {
		args = append(args, provider)
		where = append(where, "lower(trim(provider)) = $"+strconv.Itoa(len(args)))
	}
	if status != "" {
		args = append(args, status)
		where = append(where, "lower(trim(status)) = $"+strconv.Itoa(len(args)))
	}
	if direction != "" {
		args = append(args, direction)
		where = append(where, "direction = $"+strconv.Itoa(len(args)))
	}
	if search != "" {
		args = append(args, search)
		pos := strconv.Itoa(len(args))
		where = append(where, `(search_blob ILIKE '%' || $`+pos+` || '%' OR (
			regexp_replace($`+pos+`, '[^0-9]', '', 'g') <> ''
			AND (amount::text ILIKE '%' || regexp_replace($`+pos+`, '[^0-9]', '', 'g') || '%'
			     OR admin_fee::text ILIKE '%' || regexp_replace($`+pos+`, '[^0-9]', '', 'g') || '%')
		))`)
	}
	whereSQL := strings.Join(where, " AND ")

	var total int64
	if err := r.db.QueryRowContext(ctx, qrtpTransferLedgerCTE()+`SELECT count(*)::bigint FROM ledger WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, qrtpTransferLedgerCTE()+`
SELECT id, direction, ledger_type, ref_id, provider, bank_name, account_no, account_name,
       amount, admin_fee, status, reason, note, created_by_name, saldo_sebelum, saldo_sesudah,
       processed_at, callback_at, reversed_at, created_at
FROM ledger
WHERE `+whereSQL+`
ORDER BY created_at DESC, id DESC
LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]QRTPTransferLedgerRow, 0, limit)
	for rows.Next() {
		var item QRTPTransferLedgerRow
		if err := scanQRTPTransferLedgerRow(rows, &item); err != nil {
			return nil, 0, err
		}
		out = append(out, item)
	}
	return out, total, rows.Err()
}

func (r *QRTPTransferRepository) ClaimInternalTransfer(ctx context.Context, id int64, actorID int64) (*QRTPTransferRow, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	actorParam := qrtpNullableActorID(actorID)

	var row QRTPTransferRow
	if err = scanQRTPTransferBase(tx.QueryRowContext(ctx, `
SELECT id, ref_id, inquiry_order_id, provider, provider_rekening_id, bank_id, bank_code, bank_name, account_no, account_name,
       amount, admin_fee, COALESCE(note,''), status, inquiry_status, account_status, inquiry_public_id, inquiry_error,
       COALESCE(inquiry_raw,'{}'::jsonb)::text, payout_public_id, provider_transaction_id, payout_error, payout_reason,
       COALESCE(payout_raw,'{}'::jsonb)::text, bank_saldo_after, provider_saldo_after, created_by, processed_at,
       completed_at, reversed_at, callback_at, created_at, updated_at
FROM public.qrtp_provider_transfers
WHERE id = $1
FOR UPDATE
`, id), &row); err != nil {
		return nil, err
	}
	if row.Status != "inquiry_success" {
		return nil, errors.New("transfer belum siap diproses")
	}
	if row.Amount <= 0 || row.Amount > 50000000 {
		return nil, errors.New("nominal transfer tidak valid")
	}
	if row.AdminFee < 0 {
		return nil, errors.New("admin fee tidak valid")
	}
	if row.BankID <= 0 {
		row.BankID, err = r.bankRepo.ensureSystemQRTPBankTx(ctx, tx)
		if err != nil {
			return nil, err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE public.qrtp_provider_transfers SET bank_id=$2, updated_at=now() WHERE id=$1`, row.ID, row.BankID); err != nil {
			return nil, err
		}
	}

	totalDebit := row.Amount + row.AdminFee
	var bankBefore int64
	var bankName string
	if err = tx.QueryRowContext(ctx, `
SELECT saldo, nama
FROM public.bank
WHERE id = $1
FOR UPDATE
`, row.BankID).Scan(&bankBefore, &bankName); err != nil {
		return nil, err
	}
	if bankBefore < totalDebit {
		return nil, errors.New("saldo QRTP tidak cukup")
	}
	bankAfterTransfer := bankBefore - row.Amount
	bankAfter := bankBefore - totalDebit
	if _, err = tx.ExecContext(ctx, `UPDATE public.bank SET saldo=$2, diubah_pada=now() WHERE id=$1`, row.BankID, bankAfter); err != nil {
		return nil, err
	}
	metaJSON, _ := json.Marshal(map[string]any{
		"type":                 "qrtp_transfer_to_provider",
		"admin_fee":            row.AdminFee,
		"provider":             row.Provider,
		"provider_rekening_id": row.ProviderRekeningID,
		"bank_code":            row.BankCode,
		"bank_name":            row.BankName,
		"account_no":           row.AccountNo,
		"account_name":         row.AccountName,
	})
	if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_bank
  (bank_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, provider, diubah_oleh, dibuat_pada, meta)
VALUES
  ($1,$2,'DEBIT',$3,'QRTP_TRANSFER_TO_PROVIDER',NULLIF($4,''),$5,$6,$7,$8,now(),$9::jsonb)
`, row.BankID, row.RefID, row.Amount, row.Note, bankBefore, bankAfterTransfer, row.Provider, actorParam, string(metaJSON)); err != nil {
		return nil, err
	}
	if row.AdminFee > 0 {
		if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_bank
  (bank_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, provider, diubah_oleh, dibuat_pada, meta)
VALUES
  ($1,$2,'DEBIT',$3,'QRTP_TRANSFER_ADMIN_FEE','biaya admin QRTP', $4,$5,$6,$7,now(),$8::jsonb)
`, row.BankID, row.RefID, row.AdminFee, bankAfterTransfer, bankAfter, row.Provider, actorParam, string(metaJSON)); err != nil {
			return nil, err
		}
	}

	if _, err = tx.ExecContext(ctx, `
INSERT INTO public.dompet_provider (provider, saldo)
VALUES ($1, 0)
ON CONFLICT (provider) DO NOTHING
`, row.Provider); err != nil {
		return nil, err
	}
	var providerBefore int64
	if err = tx.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_provider WHERE provider=$1 FOR UPDATE`, row.Provider).Scan(&providerBefore); err != nil {
		return nil, err
	}
	providerAfter := providerBefore + row.Amount
	if _, err = tx.ExecContext(ctx, `UPDATE public.dompet_provider SET saldo=$1, diperbarui_pada=now() WHERE provider=$2`, providerAfter, row.Provider); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet_provider
  (provider, bank_id, bank_nama, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, diubah_oleh, dibuat_pada, meta)
VALUES
  ($1,$2,$3,$4,'credit',$5,'QRTP_TRANSFER_IN',COALESCE(NULLIF($6,''),'transfer QRTP ke provider'),$7,$8,$9,now(),$10::jsonb)
`, row.Provider, row.BankID, strings.TrimSpace(bankName), row.RefID, row.Amount, row.Note, providerBefore, providerAfter, actorParam, string(metaJSON)); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `
UPDATE public.qrtp_provider_transfers
SET status='processing',
    bank_saldo_after=$2,
    provider_saldo_after=$3,
    processed_at=now(),
    updated_at=now()
WHERE id=$1
`, row.ID, bankAfter, providerAfter); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
}

func (r *QRTPTransferRepository) MarkPayoutResponse(ctx context.Context, id int64, status, publicID, providerTxID, reason, errText string, raw []byte) (*QRTPTransferRow, error) {
	status = normalizeQRTPStatus(status)
	if status == "" {
		status = "requested"
	}
	rawText := qrtpJSONText(raw)
	_, err := r.db.ExecContext(ctx, `
UPDATE public.qrtp_provider_transfers
SET status=$2,
    payout_public_id=COALESCE(NULLIF($3,''), ''),
    provider_transaction_id=COALESCE(NULLIF($4,''), ''),
    payout_reason=COALESCE(NULLIF($5,''), ''),
    payout_error=COALESCE(NULLIF($6,''), ''),
    payout_raw=$7::jsonb,
    completed_at=CASE WHEN $2 IN ('success','failed','create_failed') THEN COALESCE(completed_at, now()) ELSE completed_at END,
    updated_at=now()
WHERE id=$1
`, id, status, strings.TrimSpace(publicID), strings.TrimSpace(providerTxID), strings.TrimSpace(reason), strings.TrimSpace(errText), rawText)
	if err != nil {
		return nil, err
	}
	if status == "failed" || status == "create_failed" {
		if err := r.ReverseInternalTransfer(ctx, id, "QRTP_TRANSFER_FAILED_REFUND"); err != nil {
			return nil, err
		}
	}
	return r.Get(ctx, id)
}

func (r *QRTPTransferRepository) ApplyCallback(ctx context.Context, refID, status, publicID, providerTxID, reason string, raw []byte) (*QRTPTransferRow, error) {
	item, err := r.GetByRefID(ctx, refID)
	if err != nil {
		return nil, err
	}
	status = normalizeQRTPStatus(status)
	if status == "" {
		return item, nil
	}
	rawText := qrtpJSONText(raw)
	_, err = r.db.ExecContext(ctx, `
UPDATE public.qrtp_provider_transfers
SET status=$2,
    payout_public_id=COALESCE(NULLIF($3,''), payout_public_id),
    provider_transaction_id=COALESCE(NULLIF($4,''), provider_transaction_id),
    payout_reason=COALESCE(NULLIF($5,''), payout_reason),
    payout_raw=$6::jsonb,
    callback_at=now(),
    completed_at=CASE WHEN $2 IN ('success','failed','create_failed') THEN COALESCE(completed_at, now()) ELSE completed_at END,
    updated_at=now()
WHERE id=$1
`, item.ID, status, strings.TrimSpace(publicID), strings.TrimSpace(providerTxID), strings.TrimSpace(reason), rawText)
	if err != nil {
		return nil, err
	}
	if status == "failed" || status == "create_failed" {
		if err := r.ReverseInternalTransfer(ctx, item.ID, "QRTP_TRANSFER_FAILED_CALLBACK_REFUND"); err != nil {
			return nil, err
		}
	}
	return r.Get(ctx, item.ID)
}

func (r *QRTPTransferRepository) ReverseInternalTransfer(ctx context.Context, id int64, reason string) (err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var row QRTPTransferRow
	if err = scanQRTPTransferBase(tx.QueryRowContext(ctx, `
SELECT id, ref_id, inquiry_order_id, provider, provider_rekening_id, bank_id, bank_code, bank_name, account_no, account_name,
       amount, admin_fee, COALESCE(note,''), status, inquiry_status, account_status, inquiry_public_id, inquiry_error,
       COALESCE(inquiry_raw,'{}'::jsonb)::text, payout_public_id, provider_transaction_id, payout_error, payout_reason,
       COALESCE(payout_raw,'{}'::jsonb)::text, bank_saldo_after, provider_saldo_after, created_by, processed_at,
       completed_at, reversed_at, callback_at, created_at, updated_at
FROM public.qrtp_provider_transfers
WHERE id = $1
FOR UPDATE
`, id), &row); err != nil {
		return err
	}
	if row.ProcessedAt == nil || row.ReversedAt != nil {
		return tx.Commit()
	}
	totalCredit := row.Amount + row.AdminFee
	var bankBefore int64
	if err = tx.QueryRowContext(ctx, `SELECT saldo FROM public.bank WHERE id=$1 FOR UPDATE`, row.BankID).Scan(&bankBefore); err != nil {
		return err
	}
	bankAfterTransfer := bankBefore + row.Amount
	bankAfter := bankBefore + totalCredit
	if _, err = tx.ExecContext(ctx, `UPDATE public.bank SET saldo=$2, diubah_pada=now() WHERE id=$1`, row.BankID, bankAfter); err != nil {
		return err
	}
	metaJSON, _ := json.Marshal(map[string]any{
		"type":                 "qrtp_transfer_failed_refund",
		"provider":             row.Provider,
		"provider_rekening_id": row.ProviderRekeningID,
		"reason":               reason,
	})
	if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_bank
  (bank_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, provider, dibuat_pada, meta)
VALUES
  ($1,$2,'CREDIT',$3,$4,'refund transfer QRTP gagal',$5,$6,$7,now(),$8::jsonb)
`, row.BankID, row.RefID, row.Amount, reason, bankBefore, bankAfterTransfer, row.Provider, string(metaJSON)); err != nil {
		return err
	}
	if row.AdminFee > 0 {
		if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_bank
  (bank_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, provider, dibuat_pada, meta)
VALUES
  ($1,$2,'CREDIT',$3,'QRTP_TRANSFER_ADMIN_FEE_REFUND','refund admin QRTP gagal',$4,$5,$6,now(),$7::jsonb)
`, row.BankID, row.RefID, row.AdminFee, bankAfterTransfer, bankAfter, row.Provider, string(metaJSON)); err != nil {
			return err
		}
	}
	var providerBefore int64
	if err = tx.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_provider WHERE provider=$1 FOR UPDATE`, row.Provider).Scan(&providerBefore); err != nil {
		return err
	}
	providerAfter := providerBefore - row.Amount
	if _, err = tx.ExecContext(ctx, `UPDATE public.dompet_provider SET saldo=$1, diperbarui_pada=now() WHERE provider=$2`, providerAfter, row.Provider); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet_provider
  (provider, bank_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada, meta)
VALUES
  ($1,$2,$3,'debit',$4,'QRTP_TRANSFER_REVERSAL','reversal transfer QRTP gagal',$5,$6,now(),$7::jsonb)
`, row.Provider, row.BankID, row.RefID, row.Amount, providerBefore, providerAfter, string(metaJSON)); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `
UPDATE public.qrtp_provider_transfers
SET bank_saldo_after=$2,
    provider_saldo_after=$3,
    reversed_at=now(),
    updated_at=now()
WHERE id=$1
`, row.ID, bankAfter, providerAfter); err != nil {
		return err
	}
	return tx.Commit()
}

func qrtpTransferSelectSQL() string {
	return `
SELECT t.id, t.ref_id, t.inquiry_order_id, t.provider, t.provider_rekening_id, t.bank_id, t.bank_code, t.bank_name,
       t.account_no, t.account_name, t.amount, t.admin_fee, COALESCE(t.note,''), t.status, t.inquiry_status,
       t.account_status, t.inquiry_public_id, t.inquiry_error, COALESCE(t.inquiry_raw,'{}'::jsonb)::text,
       t.payout_public_id, t.provider_transaction_id, t.payout_error, t.payout_reason, COALESCE(t.payout_raw,'{}'::jsonb)::text,
       t.bank_saldo_after, t.provider_saldo_after, t.created_by, COALESCE(actor.nama,''), t.processed_at,
       t.completed_at, t.reversed_at, t.callback_at, t.created_at, t.updated_at
FROM public.qrtp_provider_transfers t
LEFT JOIN public.member actor ON actor.id = t.created_by
`
}

func qrtpTransferLedgerCTE() string {
	return `
WITH ledger AS (
  SELECT
    t.id,
    'out'::text AS direction,
    'transfer_provider'::text AS ledger_type,
    t.ref_id,
    t.provider,
    t.bank_name,
    t.account_no,
    t.account_name,
    t.amount,
    t.admin_fee,
    t.status,
    COALESCE(NULLIF(t.payout_reason,''), NULLIF(t.payout_error,''), NULLIF(t.note,''), '') AS reason,
    COALESCE(t.note,'') AS note,
    COALESCE(actor.nama,'') AS created_by_name,
    NULL::bigint AS saldo_sebelum,
    t.bank_saldo_after AS saldo_sesudah,
    t.processed_at,
    t.callback_at,
    t.reversed_at,
    t.created_at,
    concat_ws(' ', t.ref_id, t.provider, t.bank_name, t.account_no, t.account_name, t.status,
      t.payout_reason, t.payout_error, t.note, COALESCE(actor.nama,'')) AS search_blob
  FROM public.qrtp_provider_transfers t
  LEFT JOIN public.member actor ON actor.id = t.created_by

  UNION ALL

  SELECT
    mb.id,
    'in'::text AS direction,
    CASE
      WHEN trim(COALESCE(mb.alasan,'')) = 'SMPAY_WEDE_TRANSFER_IN' THEN 'smpay_wede'
      WHEN trim(COALESCE(mb.alasan,'')) ILIKE '%REFUND%' THEN 'refund'
      ELSE lower(trim(COALESCE(mb.alasan,'')))
    END AS ledger_type,
    mb.ref_id,
    COALESCE(mb.provider,'') AS provider,
    'QRTP'::text AS bank_name,
    ''::text AS account_no,
    COALESCE(target_member.nama,'') AS account_name,
    mb.jumlah AS amount,
    0::bigint AS admin_fee,
    CASE
      WHEN trim(COALESCE(mb.alasan,'')) ILIKE '%REFUND%' THEN 'refund'
      ELSE 'success'
    END AS status,
    COALESCE(mb.alasan,'') AS reason,
    COALESCE(mb.catatan,'') AS note,
    COALESCE(actor.nama, target_member.nama, '') AS created_by_name,
    mb.saldo_sebelum,
    mb.saldo_sesudah,
    NULL::timestamptz AS processed_at,
    NULL::timestamptz AS callback_at,
    NULL::timestamptz AS reversed_at,
    COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) AS created_at,
    concat_ws(' ', mb.ref_id, mb.provider, mb.alasan, mb.catatan, target_member.nama,
      actor.nama, mb.jumlah::text) AS search_blob
  FROM public.mutasi_bank mb
  LEFT JOIN public.member target_member ON target_member.id = mb.member_id
  LEFT JOIN public.member actor ON actor.id = mb.diubah_oleh
  WHERE mb.bank_id = $1
    AND upper(trim(COALESCE(mb.arah,''))) = 'CREDIT'
)
`
}

type rowScanner interface {
	Scan(dest ...any) error
}

func (r *QRTPTransferRepository) scanRow(scanner rowScanner, item *QRTPTransferRow) error {
	var actorName string
	if err := scanQRTPTransferBaseWithActor(scanner, item, &actorName); err != nil {
		return err
	}
	item.CreatedByName = actorName
	return nil
}

func scanQRTPTransferBase(scanner rowScanner, item *QRTPTransferRow) error {
	return scanQRTPTransferBaseWithActor(scanner, item, nil)
}

func scanQRTPTransferBaseWithActor(scanner rowScanner, item *QRTPTransferRow, actorName *string) error {
	var inquiryRaw string
	var payoutRaw string
	var bankSaldo sql.NullInt64
	var providerSaldo sql.NullInt64
	var createdBy sql.NullInt64
	var processedAt sql.NullTime
	var completedAt sql.NullTime
	var reversedAt sql.NullTime
	var callbackAt sql.NullTime
	var err error
	if actorName == nil {
		actorName = new(string)
		err = scanner.Scan(
			&item.ID, &item.RefID, &item.InquiryOrderID, &item.Provider, &item.ProviderRekeningID, &item.BankID,
			&item.BankCode, &item.BankName, &item.AccountNo, &item.AccountName, &item.Amount, &item.AdminFee,
			&item.Note, &item.Status, &item.InquiryStatus, &item.AccountStatus, &item.InquiryPublicID, &item.InquiryError,
			&inquiryRaw, &item.PayoutPublicID, &item.ProviderTransactionID, &item.PayoutError, &item.PayoutReason,
			&payoutRaw, &bankSaldo, &providerSaldo, &createdBy, &processedAt, &completedAt, &reversedAt,
			&callbackAt, &item.CreatedAt, &item.UpdatedAt,
		)
	} else {
		err = scanner.Scan(
			&item.ID, &item.RefID, &item.InquiryOrderID, &item.Provider, &item.ProviderRekeningID, &item.BankID,
			&item.BankCode, &item.BankName, &item.AccountNo, &item.AccountName, &item.Amount, &item.AdminFee,
			&item.Note, &item.Status, &item.InquiryStatus, &item.AccountStatus, &item.InquiryPublicID, &item.InquiryError,
			&inquiryRaw, &item.PayoutPublicID, &item.ProviderTransactionID, &item.PayoutError, &item.PayoutReason,
			&payoutRaw, &bankSaldo, &providerSaldo, &createdBy, actorName, &processedAt, &completedAt, &reversedAt,
			&callbackAt, &item.CreatedAt, &item.UpdatedAt,
		)
	}
	if err != nil {
		return err
	}
	item.InquiryRaw = json.RawMessage(inquiryRaw)
	item.PayoutRaw = json.RawMessage(payoutRaw)
	if createdBy.Valid {
		item.CreatedBy = createdBy.Int64
	} else {
		item.CreatedBy = 0
	}
	if bankSaldo.Valid {
		v := bankSaldo.Int64
		item.BankSaldoAfter = &v
	}
	if providerSaldo.Valid {
		v := providerSaldo.Int64
		item.ProviderSaldoAfter = &v
	}
	if processedAt.Valid {
		v := processedAt.Time
		item.ProcessedAt = &v
	}
	if completedAt.Valid {
		v := completedAt.Time
		item.CompletedAt = &v
	}
	if reversedAt.Valid {
		v := reversedAt.Time
		item.ReversedAt = &v
	}
	if callbackAt.Valid {
		v := callbackAt.Time
		item.CallbackAt = &v
	}
	return nil
}

func scanQRTPTransferLedgerRow(scanner rowScanner, item *QRTPTransferLedgerRow) error {
	var saldoSebelum sql.NullInt64
	var saldoSesudah sql.NullInt64
	var processedAt sql.NullTime
	var callbackAt sql.NullTime
	var reversedAt sql.NullTime
	if err := scanner.Scan(
		&item.ID,
		&item.Direction,
		&item.LedgerType,
		&item.RefID,
		&item.Provider,
		&item.BankName,
		&item.AccountNo,
		&item.AccountName,
		&item.Amount,
		&item.AdminFee,
		&item.Status,
		&item.Reason,
		&item.Note,
		&item.CreatedByName,
		&saldoSebelum,
		&saldoSesudah,
		&processedAt,
		&callbackAt,
		&reversedAt,
		&item.CreatedAt,
	); err != nil {
		return err
	}
	if saldoSebelum.Valid {
		v := saldoSebelum.Int64
		item.SaldoSebelum = &v
	}
	if saldoSesudah.Valid {
		v := saldoSesudah.Int64
		item.SaldoSesudah = &v
	}
	if processedAt.Valid {
		v := processedAt.Time
		item.ProcessedAt = &v
	}
	if callbackAt.Valid {
		v := callbackAt.Time
		item.CallbackAt = &v
	}
	if reversedAt.Valid {
		v := reversedAt.Time
		item.ReversedAt = &v
	}
	return nil
}

func normalizeQRTPStatus(status string) string {
	status = strings.TrimSpace(strings.ToLower(status))
	switch status {
	case "success", "requested", "processing", "failed", "create_failed":
		return status
	case "payout_success":
		return "success"
	case "payout_failed":
		return "failed"
	default:
		return status
	}
}
