package repository

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const loketBayarTransferSourceProvider = "loketbayar"

type LoketBayarTransferSummary struct {
	Provider              string     `json:"provider"`
	SaldoInternal         int64      `json:"saldo_internal"`
	SnapshotSaldoProvider *int64     `json:"snapshot_saldo_provider,omitempty"`
	SnapshotAt            *time.Time `json:"snapshot_at,omitempty"`
	MaxAmount             int64      `json:"max_amount"`
}

type LoketBayarTransferRow struct {
	ID                    int64           `json:"id"`
	RefID                 string          `json:"ref_id"`
	SourceProvider        string          `json:"source_provider"`
	Provider              string          `json:"provider"`
	ProviderRekeningID    int64           `json:"provider_rekening_id"`
	BankCode              string          `json:"bank_code"`
	BankName              string          `json:"bank_name"`
	AccountNo             string          `json:"account_no"`
	AccountName           string          `json:"account_name"`
	Amount                int64           `json:"amount"`
	AdminFee              int64           `json:"admin_fee"`
	Note                  string          `json:"note"`
	Status                string          `json:"status"`
	RequestRaw            json.RawMessage `json:"request_raw,omitempty"`
	ResponseRaw           json.RawMessage `json:"response_raw,omitempty"`
	ProviderTransactionID string          `json:"provider_transaction_id"`
	ResponseError         string          `json:"response_error"`
	ResponseReason        string          `json:"response_reason"`
	SourceSaldoAfter      *int64          `json:"source_saldo_after,omitempty"`
	SourceSnapshotAfter   *int64          `json:"source_snapshot_after,omitempty"`
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

type LoketBayarTransferLedgerRow struct {
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

type LoketBayarTransferInsert struct {
	RefID              string
	Provider           string
	ProviderRekeningID int64
	BankCode           string
	BankName           string
	AccountNo          string
	AccountName        string
	Amount             int64
	Note               string
	CreatedBy          int64
}

type LoketBayarTransferRepository struct {
	db *sql.DB
}

func NewLoketBayarTransferRepository(db *sql.DB) *LoketBayarTransferRepository {
	return &LoketBayarTransferRepository{db: db}
}

func loketTransferNullableActorID(actorID int64) any {
	if actorID > 0 {
		return actorID
	}
	return nil
}

func loketTransferJSONText(raw []byte) string {
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

func (r *LoketBayarTransferRepository) Summary(ctx context.Context, maxAmount int64) (*LoketBayarTransferSummary, error) {
	if _, err := r.db.ExecContext(ctx, `
INSERT INTO public.dompet_provider (provider, saldo)
VALUES ($1, 0)
ON CONFLICT (provider) DO NOTHING
`, loketBayarTransferSourceProvider); err != nil {
		return nil, err
	}

	out := &LoketBayarTransferSummary{Provider: loketBayarTransferSourceProvider, MaxAmount: maxAmount}
	if err := r.db.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_provider
WHERE provider = $1
LIMIT 1
`, loketBayarTransferSourceProvider).Scan(&out.SaldoInternal); err != nil {
		return nil, err
	}

	var snapshot sql.NullInt64
	var snapshotAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
SELECT saldo_provider, dibuat_pada
FROM public.provider_saldo_snapshot
WHERE provider = $1
ORDER BY transaksi_provider_id DESC NULLS LAST, dibuat_pada DESC NULLS LAST, id DESC
LIMIT 1
`, loketBayarTransferSourceProvider).Scan(&snapshot, &snapshotAt)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if snapshot.Valid {
		v := snapshot.Int64
		out.SnapshotSaldoProvider = &v
	}
	if snapshotAt.Valid {
		v := snapshotAt.Time
		out.SnapshotAt = &v
	}
	return out, nil
}

func (r *LoketBayarTransferRepository) GetProviderRekening(ctx context.Context, id int64) (*ProviderRekeningRow, error) {
	return NewProviderRekeningRepository(r.db).Get(ctx, id)
}

func (r *LoketBayarTransferRepository) ResolveBankCode(ctx context.Context, bank string) (string, string, error) {
	sku := normalizeLoketBayarTransferBankSKU(bank)
	if sku == "" {
		return "", "", errors.New("bank rekening provider kosong")
	}
	const bankTransferProduct = "DBALLBANK"
	var code, name string
	err := r.db.QueryRowContext(ctx, `
SELECT ppm.kode_provider, p.nama
FROM public.produk_provider_map ppm
JOIN public.produk p ON p.id = ppm.produk_id
WHERE lower(trim(ppm.provider)) = 'loketbayar'
  AND upper(trim(COALESCE(ppm.special_code, ''))) = $2
  AND upper(trim(p.sku)) = $1
  AND ppm.aktif = true
ORDER BY ppm.id DESC
LIMIT 1
`, sku, bankTransferProduct).Scan(&code, &name)
	if err == nil {
		return strings.TrimSpace(code), strings.TrimSpace(name), nil
	}
	if err != sql.ErrNoRows {
		return "", "", err
	}
	return "", "", fmt.Errorf("bank %s belum aktif di LoketBayar %s", bank, bankTransferProduct)
}

func normalizeLoketBayarTransferBankSKU(bank string) string {
	up := strings.ToUpper(strings.TrimSpace(bank))
	up = strings.TrimPrefix(up, "BANK ")
	up = strings.ReplaceAll(up, ".", "")
	up = strings.ReplaceAll(up, "-", "")
	up = strings.ReplaceAll(up, "_", "")
	up = strings.Join(strings.Fields(up), "")
	switch up {
	case "BANKCENTRALASIA":
		return "BCA"
	case "BANKRAKYATINDONESIA":
		return "BRI"
	case "BANKNEGARAINDONESIA":
		return "BNI"
	case "BANKMANDIRI":
		return "MANDIRI"
	default:
		return up
	}
}

func (r *LoketBayarTransferRepository) InsertRequest(ctx context.Context, in LoketBayarTransferInsert) (*LoketBayarTransferRow, error) {
	in.RefID = strings.TrimSpace(in.RefID)
	in.Provider = strings.TrimSpace(strings.ToLower(in.Provider))
	if in.RefID == "" || in.Provider == "" || in.ProviderRekeningID <= 0 || in.Amount <= 0 {
		return nil, errors.New("payload transfer tidak valid")
	}
	var id int64
	err := r.db.QueryRowContext(ctx, `
INSERT INTO public.loketbayar_provider_transfers
  (ref_id, source_provider, provider, provider_rekening_id, bank_code, bank_name, account_no, account_name,
   amount, admin_fee, note, status, created_by, created_at, updated_at)
VALUES
  ($1,$2,$3,$4,$5,$6,$7,$8,$9,0,$10,'ready',$11,now(),now())
RETURNING id
`, in.RefID, loketBayarTransferSourceProvider, in.Provider, in.ProviderRekeningID, in.BankCode, in.BankName,
		in.AccountNo, in.AccountName, in.Amount, in.Note, loketTransferNullableActorID(in.CreatedBy)).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
}

func (r *LoketBayarTransferRepository) Get(ctx context.Context, id int64) (*LoketBayarTransferRow, error) {
	var item LoketBayarTransferRow
	if err := r.scanRow(r.db.QueryRowContext(ctx, loketTransferSelectSQL()+` WHERE t.id = $1 LIMIT 1`, id), &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *LoketBayarTransferRepository) GetByRefID(ctx context.Context, refID string) (*LoketBayarTransferRow, error) {
	var item LoketBayarTransferRow
	if err := r.scanRow(r.db.QueryRowContext(ctx, loketTransferSelectSQL()+` WHERE t.ref_id = $1 LIMIT 1`, strings.TrimSpace(refID)), &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *LoketBayarTransferRepository) List(ctx context.Context, provider, status, search string, limit, offset int) ([]LoketBayarTransferLedgerRow, int64, error) {
	provider = strings.TrimSpace(strings.ToLower(provider))
	status = strings.TrimSpace(strings.ToLower(status))
	search = strings.TrimSpace(search)
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	args := []any{}
	where := []string{"1=1"}
	if provider != "" {
		args = append(args, provider)
		where = append(where, "lower(trim(provider)) = $"+strconv.Itoa(len(args)))
	}
	if status != "" {
		args = append(args, status)
		where = append(where, "lower(trim(status)) = $"+strconv.Itoa(len(args)))
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
	if err := r.db.QueryRowContext(ctx, loketTransferLedgerCTE()+`SELECT count(*)::bigint FROM ledger WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, loketTransferLedgerCTE()+`
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
	out := make([]LoketBayarTransferLedgerRow, 0, limit)
	for rows.Next() {
		var item LoketBayarTransferLedgerRow
		if err := scanLoketTransferLedgerRow(rows, &item); err != nil {
			return nil, 0, err
		}
		out = append(out, item)
	}
	return out, total, rows.Err()
}

func (r *LoketBayarTransferRepository) ClaimInternalTransfer(ctx context.Context, id int64, actorID int64) (rowOut *LoketBayarTransferRow, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	actorParam := loketTransferNullableActorID(actorID)

	var row LoketBayarTransferRow
	if err = scanLoketTransferBase(tx.QueryRowContext(ctx, `
SELECT id, ref_id, source_provider, provider, provider_rekening_id, bank_code, bank_name, account_no, account_name,
       amount, admin_fee, COALESCE(note,''), status, COALESCE(request_raw,'{}'::jsonb)::text,
       COALESCE(response_raw,'{}'::jsonb)::text, provider_transaction_id, response_error, response_reason,
       source_saldo_after, source_snapshot_after, provider_saldo_after, created_by, processed_at, completed_at,
       reversed_at, callback_at, created_at, updated_at
FROM public.loketbayar_provider_transfers
WHERE id = $1
FOR UPDATE
`, id), &row); err != nil {
		return nil, err
	}
	if row.Status != "ready" {
		return nil, errors.New("transfer belum siap diproses")
	}
	if row.Amount <= 0 || row.Amount > 50000000 {
		return nil, errors.New("nominal transfer tidak valid")
	}

	if _, err = tx.ExecContext(ctx, `
INSERT INTO public.dompet_provider (provider, saldo)
VALUES ($1, 0), ($2, 0)
ON CONFLICT (provider) DO NOTHING
`, loketBayarTransferSourceProvider, row.Provider); err != nil {
		return nil, err
	}

	var sourceBefore int64
	if err = tx.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_provider WHERE provider=$1 FOR UPDATE`, loketBayarTransferSourceProvider).Scan(&sourceBefore); err != nil {
		return nil, err
	}
	if sourceBefore < row.Amount {
		return nil, errors.New("saldo internal LoketBayar tidak cukup")
	}
	sourceAfter := sourceBefore - row.Amount
	if _, err = tx.ExecContext(ctx, `UPDATE public.dompet_provider SET saldo=$1, diperbarui_pada=now() WHERE provider=$2`, sourceAfter, loketBayarTransferSourceProvider); err != nil {
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

	metaJSON, _ := json.Marshal(map[string]any{
		"type":                 "loketbayar_transfer_to_provider",
		"source_provider":      loketBayarTransferSourceProvider,
		"provider":             row.Provider,
		"provider_rekening_id": row.ProviderRekeningID,
		"bank_code":            row.BankCode,
		"bank_name":            row.BankName,
		"account_no":           row.AccountNo,
		"account_name":         row.AccountName,
	})
	if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet_provider
  (provider, bank_nama, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, diubah_oleh, dibuat_pada, meta)
VALUES
  ($1,$2,$3,'debit',$4,'LOKETBAYAR_TRANSFER_OUT',COALESCE(NULLIF($5,''),'transfer LoketBayar ke provider'),$6,$7,$8,now(),$9::jsonb)
`, loketBayarTransferSourceProvider, row.BankName, row.RefID, row.Amount, row.Note, sourceBefore, sourceAfter, actorParam, string(metaJSON)); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet_provider
  (provider, bank_nama, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, diubah_oleh, dibuat_pada, meta)
VALUES
  ($1,$2,$3,'credit',$4,'LOKETBAYAR_TRANSFER_IN',COALESCE(NULLIF($5,''),'transfer LoketBayar ke provider'),$6,$7,$8,now(),$9::jsonb)
`, row.Provider, row.BankName, row.RefID, row.Amount, row.Note, providerBefore, providerAfter, actorParam, string(metaJSON)); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `
UPDATE public.loketbayar_provider_transfers
SET status='processing',
    source_saldo_after=$2,
    provider_saldo_after=$3,
    processed_at=now(),
    updated_at=now()
WHERE id=$1
`, row.ID, sourceAfter, providerAfter); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
}

func (r *LoketBayarTransferRepository) MarkProviderResponse(ctx context.Context, id int64, status, providerTxID, reason, errText string, price int64, sourceSnapshot *int64, raw []byte) (*LoketBayarTransferRow, error) {
	status = normalizeLoketTransferStatus(status)
	if status == "" {
		status = "processing"
	}
	rawText := loketTransferJSONText(raw)
	err := r.updateTransferResponse(ctx, id, status, providerTxID, reason, errText, price, sourceSnapshot, rawText, false)
	if err != nil {
		return nil, err
	}
	if status == "failed" || status == "create_failed" {
		if err := r.ReverseInternalTransfer(ctx, id, "LOKETBAYAR_TRANSFER_FAILED_REFUND"); err != nil {
			return nil, err
		}
	}
	return r.Get(ctx, id)
}

func (r *LoketBayarTransferRepository) ApplyCallback(ctx context.Context, refID, status, providerTxID, reason string, price int64, sourceSnapshot *int64, raw []byte) (*LoketBayarTransferRow, error) {
	item, err := r.GetByRefID(ctx, refID)
	if err != nil {
		return nil, err
	}
	status = normalizeLoketTransferStatus(status)
	if status == "" {
		return item, nil
	}
	rawText := loketTransferJSONText(raw)
	if err := r.updateTransferResponse(ctx, item.ID, status, providerTxID, reason, "", price, sourceSnapshot, rawText, true); err != nil {
		return nil, err
	}
	if status == "failed" || status == "create_failed" {
		if err := r.ReverseInternalTransfer(ctx, item.ID, "LOKETBAYAR_TRANSFER_FAILED_CALLBACK_REFUND"); err != nil {
			return nil, err
		}
	}
	return r.Get(ctx, item.ID)
}

func (r *LoketBayarTransferRepository) InsertSourceSnapshot(ctx context.Context, refID string, saldo int64, raw []byte, source string) error {
	if saldo <= 0 {
		return nil
	}
	if strings.TrimSpace(source) == "" {
		source = "loketbayar_transfer"
	}
	rawText := loketTransferJSONText(raw)
	_, err := r.db.ExecContext(ctx, `
INSERT INTO public.provider_saldo_snapshot
  (provider, saldo_provider, ref_id, sumber, dibuat_pada, raw)
VALUES
  ($1,$2,NULLIF($3,''),$4,$5,$6::jsonb)
`, loketBayarTransferSourceProvider, saldo, strings.TrimSpace(refID), strings.TrimSpace(source), nowWIB(), rawText)
	return err
}

func (r *LoketBayarTransferRepository) updateTransferResponse(ctx context.Context, id int64, status, providerTxID, reason, errText string, price int64, sourceSnapshot *int64, rawText string, callback bool) (err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var row LoketBayarTransferRow
	if err = scanLoketTransferBase(tx.QueryRowContext(ctx, `
SELECT id, ref_id, source_provider, provider, provider_rekening_id, bank_code, bank_name, account_no, account_name,
       amount, admin_fee, COALESCE(note,''), status, COALESCE(request_raw,'{}'::jsonb)::text,
       COALESCE(response_raw,'{}'::jsonb)::text, provider_transaction_id, response_error, response_reason,
       source_saldo_after, source_snapshot_after, provider_saldo_after, created_by, processed_at, completed_at,
       reversed_at, callback_at, created_at, updated_at
FROM public.loketbayar_provider_transfers
WHERE id = $1
FOR UPDATE
`, id), &row); err != nil {
		return err
	}

	if row.ProcessedAt != nil && row.ReversedAt == nil && price > row.Amount {
		newFee := price - row.Amount
		if newFee > row.AdminFee {
			delta := newFee - row.AdminFee
			var sourceBefore int64
			if err = tx.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_provider WHERE provider=$1 FOR UPDATE`, loketBayarTransferSourceProvider).Scan(&sourceBefore); err != nil {
				return err
			}
			sourceAfter := sourceBefore - delta
			if _, err = tx.ExecContext(ctx, `UPDATE public.dompet_provider SET saldo=$1, diperbarui_pada=now() WHERE provider=$2`, sourceAfter, loketBayarTransferSourceProvider); err != nil {
				return err
			}
			metaJSON, _ := json.Marshal(map[string]any{
				"type":            "loketbayar_transfer_admin_fee",
				"provider":        row.Provider,
				"amount":          row.Amount,
				"new_admin_fee":   newFee,
				"admin_fee_delta": delta,
			})
			if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet_provider
  (provider, bank_nama, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada, meta)
VALUES
  ($1,$2,$3,'debit',$4,'LOKETBAYAR_TRANSFER_ADMIN_FEE','biaya admin transfer LoketBayar',$5,$6,now(),$7::jsonb)
`, loketBayarTransferSourceProvider, row.BankName, row.RefID, delta, sourceBefore, sourceAfter, string(metaJSON)); err != nil {
				return err
			}
			row.AdminFee = newFee
			row.SourceSaldoAfter = &sourceAfter
		}
	}

	var snapshot any
	if sourceSnapshot != nil && *sourceSnapshot > 0 {
		snapshot = *sourceSnapshot
	} else {
		snapshot = nil
	}
	var sourceAfter any
	if row.SourceSaldoAfter != nil {
		sourceAfter = *row.SourceSaldoAfter
	}
	if _, err = tx.ExecContext(ctx, `
UPDATE public.loketbayar_provider_transfers
SET status=$2,
    admin_fee=$3,
    provider_transaction_id=COALESCE(NULLIF($4,''), provider_transaction_id),
    response_reason=COALESCE(NULLIF($5,''), response_reason),
    response_error=COALESCE(NULLIF($6,''), response_error),
    response_raw=$7::jsonb,
    source_saldo_after=COALESCE($8::bigint, source_saldo_after),
    source_snapshot_after=COALESCE($9::bigint, source_snapshot_after),
    callback_at=CASE WHEN $10::boolean THEN now() ELSE callback_at END,
    completed_at=CASE WHEN $2 IN ('success','failed','create_failed') THEN COALESCE(completed_at, now()) ELSE completed_at END,
    updated_at=now()
WHERE id=$1
`, row.ID, status, row.AdminFee, strings.TrimSpace(providerTxID), strings.TrimSpace(reason), strings.TrimSpace(errText), rawText, sourceAfter, snapshot, callback); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *LoketBayarTransferRepository) ReverseInternalTransfer(ctx context.Context, id int64, reason string) (err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var row LoketBayarTransferRow
	if err = scanLoketTransferBase(tx.QueryRowContext(ctx, `
SELECT id, ref_id, source_provider, provider, provider_rekening_id, bank_code, bank_name, account_no, account_name,
       amount, admin_fee, COALESCE(note,''), status, COALESCE(request_raw,'{}'::jsonb)::text,
       COALESCE(response_raw,'{}'::jsonb)::text, provider_transaction_id, response_error, response_reason,
       source_saldo_after, source_snapshot_after, provider_saldo_after, created_by, processed_at, completed_at,
       reversed_at, callback_at, created_at, updated_at
FROM public.loketbayar_provider_transfers
WHERE id = $1
FOR UPDATE
`, id), &row); err != nil {
		return err
	}
	if row.ProcessedAt == nil || row.ReversedAt != nil {
		return tx.Commit()
	}
	totalCredit := row.Amount + row.AdminFee
	var sourceBefore int64
	if err = tx.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_provider WHERE provider=$1 FOR UPDATE`, loketBayarTransferSourceProvider).Scan(&sourceBefore); err != nil {
		return err
	}
	sourceAfterAmount := sourceBefore + row.Amount
	sourceAfter := sourceBefore + totalCredit
	if _, err = tx.ExecContext(ctx, `UPDATE public.dompet_provider SET saldo=$1, diperbarui_pada=now() WHERE provider=$2`, sourceAfter, loketBayarTransferSourceProvider); err != nil {
		return err
	}
	metaJSON, _ := json.Marshal(map[string]any{
		"type":            "loketbayar_transfer_failed_refund",
		"source_provider": loketBayarTransferSourceProvider,
		"provider":        row.Provider,
		"reason":          reason,
	})
	if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet_provider
  (provider, bank_nama, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada, meta)
VALUES
  ($1,$2,$3,'credit',$4,$5,'refund transfer LoketBayar gagal',$6,$7,now(),$8::jsonb)
`, loketBayarTransferSourceProvider, row.BankName, row.RefID, row.Amount, reason, sourceBefore, sourceAfterAmount, string(metaJSON)); err != nil {
		return err
	}
	if row.AdminFee > 0 {
		if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet_provider
  (provider, bank_nama, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada, meta)
VALUES
  ($1,$2,$3,'credit',$4,'LOKETBAYAR_TRANSFER_ADMIN_FEE_REFUND','refund admin transfer LoketBayar gagal',$5,$6,now(),$7::jsonb)
`, loketBayarTransferSourceProvider, row.BankName, row.RefID, row.AdminFee, sourceAfterAmount, sourceAfter, string(metaJSON)); err != nil {
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
  (provider, bank_nama, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada, meta)
VALUES
  ($1,$2,$3,'debit',$4,'LOKETBAYAR_TRANSFER_REVERSAL','reversal transfer LoketBayar gagal',$5,$6,now(),$7::jsonb)
`, row.Provider, row.BankName, row.RefID, row.Amount, providerBefore, providerAfter, string(metaJSON)); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `
UPDATE public.loketbayar_provider_transfers
SET source_saldo_after=$2,
    provider_saldo_after=$3,
    reversed_at=now(),
    updated_at=now()
WHERE id=$1
`, row.ID, sourceAfter, providerAfter); err != nil {
		return err
	}
	return tx.Commit()
}

func loketTransferSelectSQL() string {
	return `
SELECT t.id, t.ref_id, t.source_provider, t.provider, t.provider_rekening_id, t.bank_code, t.bank_name,
       t.account_no, t.account_name, t.amount, t.admin_fee, COALESCE(t.note,''), t.status,
       COALESCE(t.request_raw,'{}'::jsonb)::text, COALESCE(t.response_raw,'{}'::jsonb)::text,
       t.provider_transaction_id, t.response_error, t.response_reason, t.source_saldo_after,
       t.source_snapshot_after, t.provider_saldo_after, t.created_by, COALESCE(actor.nama,''), t.processed_at,
       t.completed_at, t.reversed_at, t.callback_at, t.created_at, t.updated_at
FROM public.loketbayar_provider_transfers t
LEFT JOIN public.member actor ON actor.id = t.created_by
`
}

func loketTransferLedgerCTE() string {
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
    COALESCE(NULLIF(t.response_reason,''), NULLIF(t.response_error,''), NULLIF(t.note,''), '') AS reason,
    COALESCE(t.note,'') AS note,
    COALESCE(actor.nama,'') AS created_by_name,
    NULL::bigint AS saldo_sebelum,
    t.source_saldo_after AS saldo_sesudah,
    t.processed_at,
    t.callback_at,
    t.reversed_at,
    t.created_at,
    concat_ws(' ', t.ref_id, t.provider, t.bank_name, t.account_no, t.account_name, t.status,
      t.response_reason, t.response_error, t.note, COALESCE(actor.nama,''), t.amount::text, t.admin_fee::text) AS search_blob
  FROM public.loketbayar_provider_transfers t
  LEFT JOIN public.member actor ON actor.id = t.created_by
)
`
}

type loketTransferRowScanner interface {
	Scan(dest ...any) error
}

func (r *LoketBayarTransferRepository) scanRow(scanner loketTransferRowScanner, item *LoketBayarTransferRow) error {
	var actorName string
	if err := scanLoketTransferBaseWithActor(scanner, item, &actorName); err != nil {
		return err
	}
	item.CreatedByName = actorName
	return nil
}

func scanLoketTransferBase(scanner loketTransferRowScanner, item *LoketBayarTransferRow) error {
	return scanLoketTransferBaseWithActor(scanner, item, nil)
}

func scanLoketTransferBaseWithActor(scanner loketTransferRowScanner, item *LoketBayarTransferRow, actorName *string) error {
	var requestRaw string
	var responseRaw string
	var sourceSaldo sql.NullInt64
	var sourceSnapshot sql.NullInt64
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
			&item.ID, &item.RefID, &item.SourceProvider, &item.Provider, &item.ProviderRekeningID, &item.BankCode,
			&item.BankName, &item.AccountNo, &item.AccountName, &item.Amount, &item.AdminFee, &item.Note,
			&item.Status, &requestRaw, &responseRaw, &item.ProviderTransactionID, &item.ResponseError, &item.ResponseReason,
			&sourceSaldo, &sourceSnapshot, &providerSaldo, &createdBy, &processedAt, &completedAt, &reversedAt,
			&callbackAt, &item.CreatedAt, &item.UpdatedAt,
		)
	} else {
		err = scanner.Scan(
			&item.ID, &item.RefID, &item.SourceProvider, &item.Provider, &item.ProviderRekeningID, &item.BankCode,
			&item.BankName, &item.AccountNo, &item.AccountName, &item.Amount, &item.AdminFee, &item.Note,
			&item.Status, &requestRaw, &responseRaw, &item.ProviderTransactionID, &item.ResponseError, &item.ResponseReason,
			&sourceSaldo, &sourceSnapshot, &providerSaldo, &createdBy, actorName, &processedAt, &completedAt, &reversedAt,
			&callbackAt, &item.CreatedAt, &item.UpdatedAt,
		)
	}
	if err != nil {
		return err
	}
	item.RequestRaw = json.RawMessage(requestRaw)
	item.ResponseRaw = json.RawMessage(responseRaw)
	if sourceSaldo.Valid {
		v := sourceSaldo.Int64
		item.SourceSaldoAfter = &v
	}
	if sourceSnapshot.Valid {
		v := sourceSnapshot.Int64
		item.SourceSnapshotAfter = &v
	}
	if providerSaldo.Valid {
		v := providerSaldo.Int64
		item.ProviderSaldoAfter = &v
	}
	if createdBy.Valid {
		item.CreatedBy = createdBy.Int64
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

func scanLoketTransferLedgerRow(scanner loketTransferRowScanner, item *LoketBayarTransferLedgerRow) error {
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

func normalizeLoketTransferStatus(status string) string {
	status = strings.TrimSpace(strings.ToLower(status))
	switch status {
	case "ready", "success", "processing", "requested", "failed", "create_failed":
		return status
	case "pending":
		return "processing"
	default:
		return status
	}
}
