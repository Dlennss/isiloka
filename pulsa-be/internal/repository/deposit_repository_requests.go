package repository

import (
	"context"
	cryptorand "crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

func (r *DepositRepository) CreateRequest(ctx context.Context, memberID, bankID, amount int64, metode, buktiURL, bankNama, bankNomorRekening, bankAtasNama string) error {
	if memberID <= 0 || amount <= 0 {
		return errors.New("invalid member_id/amount")
	}
	if bankID <= 0 {
		return errors.New("bank_id required")
	}
	if strings.TrimSpace(metode) == "" {
		return errors.New("metode required")
	}
	_, err := r.db.ExecContext(ctx, `
INSERT INTO public.deposit_request (member_id, bank_id, bank_nama, bank_nomor_rekening, bank_atas_nama, amount, metode, bukti_url, status)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'pending')
`, memberID, bankID, strings.TrimSpace(bankNama), strings.TrimSpace(bankNomorRekening), strings.TrimSpace(bankAtasNama), amount, strings.TrimSpace(metode), strings.TrimSpace(buktiURL))
	return err
}

func (r *DepositRepository) CreateTicketRequest(ctx context.Context, memberID, bankID, requestedAmount int64, metode, bankNama, bankNomorRekening, bankAtasNama, refID string) (*DepositRequestRow, error) {
	if memberID <= 0 || requestedAmount <= 0 {
		return nil, errors.New("invalid member_id/amount")
	}
	if bankID <= 0 {
		return nil, errors.New("bank_id required")
	}
	if strings.TrimSpace(metode) == "" {
		return nil, errors.New("metode required")
	}
	if strings.TrimSpace(refID) == "" {
		return nil, errors.New("ref_id required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(2405202601)`); err != nil {
		return nil, err
	}

	var activeCount int
	if err := tx.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM public.deposit_request
WHERE member_id = $1
  AND status IN ('ticket', 'pending')
  AND LOWER(TRIM(COALESCE(metode, ''))) <> 'qris'
`, memberID).Scan(&activeCount); err != nil {
		return nil, err
	}
	if activeCount >= 5 {
		return nil, errors.New("maksimal 5 tiket aktif")
	}

	var (
		insertedID int64
	)
	for i := 0; i < 200; i++ {
		uniqueCode, err := randomDepositUniqueCode()
		if err != nil {
			return nil, err
		}
		ticketAmount := requestedAmount + uniqueCode

		var exists int
		err = tx.QueryRowContext(ctx, `
SELECT 1
FROM public.deposit_request
WHERE amount = $1
  AND status IN ('ticket', 'pending')
  AND LOWER(TRIM(COALESCE(metode, ''))) <> 'qris'
LIMIT 1
`, ticketAmount).Scan(&exists)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		if exists == 1 {
			continue
		}

		err = tx.QueryRowContext(ctx, `
INSERT INTO public.deposit_request (member_id, bank_id, bank_nama, bank_nomor_rekening, bank_atas_nama, amount, requested_amount, unique_code, metode, bukti_url, status, note, ref_id)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,'','pending','Menunggu persetujuan admin',$10)
RETURNING id
`, memberID, bankID, strings.TrimSpace(bankNama), strings.TrimSpace(bankNomorRekening), strings.TrimSpace(bankAtasNama), ticketAmount, requestedAmount, uniqueCode, strings.TrimSpace(metode), strings.TrimSpace(refID)).Scan(&insertedID)
		if err != nil {
			return nil, err
		}
		break
	}
	if insertedID <= 0 {
		return nil, errors.New("kode unik aktif penuh, coba lagi")
	}

	row, err := r.getByIDTx(ctx, tx, insertedID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return row, nil
}

func (r *DepositRepository) ActiveNonQrisTicketCount(ctx context.Context, memberID int64) (int, error) {
	if memberID <= 0 {
		return 0, errors.New("invalid member_id")
	}
	var activeCount int
	err := r.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM public.deposit_request
WHERE member_id = $1
  AND status = 'ticket'
  AND LOWER(TRIM(COALESCE(metode, ''))) <> 'qris'
`, memberID).Scan(&activeCount)
	return activeCount, err
}

func (r *DepositRepository) CreateVARequest(ctx context.Context, memberID, requestedAmount, ticketAmount int64, bankCode, bankNama, bankNomorRekening, bankAtasNama, ticketID, note string) (*DepositRequestRow, error) {
	if memberID <= 0 || requestedAmount <= 0 || ticketAmount <= 0 {
		return nil, errors.New("invalid member_id/amount")
	}
	if ticketAmount < requestedAmount {
		return nil, errors.New("invalid ticket amount")
	}
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" {
		return nil, errors.New("ref_id required")
	}
	if strings.TrimSpace(bankCode) == "" || strings.TrimSpace(bankNomorRekening) == "" {
		return nil, errors.New("data VA belum lengkap")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(2405202601)`); err != nil {
		return nil, err
	}

	var activeCount int
	if err := tx.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM public.deposit_request
WHERE member_id = $1
  AND status = 'ticket'
  AND LOWER(TRIM(COALESCE(metode, ''))) <> 'qris'
`, memberID).Scan(&activeCount); err != nil {
		return nil, err
	}
	if activeCount >= 5 {
		return nil, errors.New("maksimal 5 tiket aktif")
	}

	var exists int
	err = tx.QueryRowContext(ctx, `
SELECT 1
FROM public.deposit_request
WHERE TRIM(COALESCE(ref_id, '')) = $1
LIMIT 1
`, ticketID).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if exists == 1 {
		return nil, errors.New("tiket loketbayar sudah ada")
	}

	var insertedID int64
	err = tx.QueryRowContext(ctx, `
INSERT INTO public.deposit_request
  (member_id, bank_nama, bank_nomor_rekening, bank_atas_nama, amount, requested_amount, unique_code, metode, bukti_url, status, note, ref_id)
VALUES
  ($1,$2,$3,$4,$5,$6,$7,'va','','ticket',$8,$9)
RETURNING id
`, memberID, strings.TrimSpace(bankNama), strings.TrimSpace(bankNomorRekening), strings.TrimSpace(bankAtasNama), ticketAmount, requestedAmount, ticketAmount-requestedAmount, strings.TrimSpace(note), ticketID).Scan(&insertedID)
	if err != nil {
		return nil, err
	}

	row, err := r.getByIDTx(ctx, tx, insertedID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return row, nil
}

func randomDepositUniqueCode() (int64, error) {
	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(999))
	if err != nil {
		return 0, err
	}
	return 1 + n.Int64(), nil
}

func (r *DepositRepository) CreateQrisRequest(ctx context.Context, memberID, amount int64, metode, refID, note string) error {
	if memberID <= 0 || amount <= 0 {
		return errors.New("invalid member_id/amount")
	}
	if strings.TrimSpace(metode) == "" {
		return errors.New("metode required")
	}
	if strings.TrimSpace(refID) == "" {
		return errors.New("ref_id required")
	}
	_, err := r.db.ExecContext(ctx, `
INSERT INTO public.deposit_request (member_id, amount, metode, bukti_url, status, note, ref_id)
VALUES ($1,$2,$3,'','pending',$4,$5)
`, memberID, amount, strings.TrimSpace(metode), strings.TrimSpace(note), strings.TrimSpace(refID))
	return err
}

func (r *DepositRepository) ListByMember(ctx context.Context, memberID int64, limit int) ([]DepositRequestRow, error) {
	if memberID <= 0 {
		return nil, errors.New("invalid member_id")
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	const q = `
SELECT d.id, d.member_id, COALESCE(d.ref_id, ''), m.nama AS member_nama, d.bank_id, d.bank_nama, d.bank_nomor_rekening, d.bank_atas_nama, d.amount, COALESCE(d.requested_amount, 0), COALESCE(d.unique_code, 0), COALESCE(d.approved_amount, 0), d.metode, d.bukti_url, d.status, d.note, d.dibuat_pada, d.diproses_pada, d.diproses_oleh,
       p.nama as diproses_nama
FROM public.deposit_request d
LEFT JOIN public.member m ON m.id = d.member_id
LEFT JOIN public.member p ON p.id = d.diproses_oleh
WHERE d.member_id = $1
ORDER BY d.dibuat_pada DESC
LIMIT $2
`
	rows, err := r.db.QueryContext(ctx, q, memberID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDepositRows(rows)
}

func (r *DepositRepository) AdminList(ctx context.Context, status string, memberID int64, fromStr, toStr string, limit, offset int, order string) ([]DepositRequestRow, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	loc, _ := time.LoadLocation("Asia/Jakarta")

	var (
		args   []any
		wheres []string
	)

	status = strings.TrimSpace(status)
	if status == "" {
		wheres = append(wheres, "d.status NOT IN ('ticket', 'cancelled')")
	} else {
		args = append(args, status)
		wheres = append(wheres, fmt.Sprintf("d.status = $%d", len(args)))
	}

	if memberID > 0 {
		args = append(args, memberID)
		wheres = append(wheres, fmt.Sprintf("d.member_id = $%d", len(args)))
	}

	if s := strings.TrimSpace(fromStr); s != "" {
		t, err := time.ParseInLocation("2006-01-02", s, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("d.dibuat_pada >= $%d", len(args)))
	}
	if s := strings.TrimSpace(toStr); s != "" {
		t, err := time.ParseInLocation("2006-01-02", s, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t.AddDate(0, 0, 1))
		wheres = append(wheres, fmt.Sprintf("d.dibuat_pada < $%d", len(args)))
	}

	orderDir := "DESC"
	if strings.EqualFold(strings.TrimSpace(order), "asc") {
		orderDir = "ASC"
	}

	args = append(args, limit)
	limitPos := len(args)
	args = append(args, offset)
	offsetPos := len(args)

	q := fmt.Sprintf(`
SELECT d.id, d.member_id, COALESCE(d.ref_id, ''), m.nama AS member_nama, d.bank_id, d.bank_nama, d.bank_nomor_rekening, d.bank_atas_nama, d.amount, COALESCE(d.requested_amount, 0), COALESCE(d.unique_code, 0), COALESCE(d.approved_amount, 0), d.metode, d.bukti_url, d.status, d.note, d.dibuat_pada, d.diproses_pada, d.diproses_oleh,
       p.nama as diproses_nama
FROM public.deposit_request d
LEFT JOIN public.member m ON m.id = d.member_id
LEFT JOIN public.member p ON p.id = d.diproses_oleh
WHERE LOWER(TRIM(d.metode)) <> 'qris' AND %s
ORDER BY d.dibuat_pada %s, d.id %s
LIMIT $%d OFFSET $%d
`, strings.Join(wheres, " AND "), orderDir, orderDir, limitPos, offsetPos)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDepositRows(rows)
}

func (r *DepositRepository) AdminListVA(ctx context.Context, status string, memberID int64, refID, fromStr, toStr string, limit, offset int, order string) ([]DepositRequestRow, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	loc, _ := time.LoadLocation("Asia/Jakarta")

	var (
		args   []any
		wheres []string
	)

	wheres = append(wheres, "LOWER(TRIM(COALESCE(d.metode, ''))) = 'va'")
	refID = strings.TrimPrefix(strings.TrimSpace(refID), "#")
	if refID != "" {
		args = append(args, refID)
		wheres = append(wheres, fmt.Sprintf("TRIM(COALESCE(d.ref_id, '')) = $%d", len(args)))
	} else {
		status = strings.TrimSpace(status)
		if status != "" && !strings.EqualFold(status, "all") {
			args = append(args, status)
			wheres = append(wheres, fmt.Sprintf("d.status = $%d", len(args)))
		}

		if memberID > 0 {
			args = append(args, memberID)
			wheres = append(wheres, fmt.Sprintf("d.member_id = $%d", len(args)))
		}

		if s := strings.TrimSpace(fromStr); s != "" {
			t, err := time.ParseInLocation("2006-01-02", s, loc)
			if err != nil {
				return nil, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
			}
			args = append(args, t)
			wheres = append(wheres, fmt.Sprintf("d.dibuat_pada >= $%d", len(args)))
		}
		if s := strings.TrimSpace(toStr); s != "" {
			t, err := time.ParseInLocation("2006-01-02", s, loc)
			if err != nil {
				return nil, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
			}
			args = append(args, t.AddDate(0, 0, 1))
			wheres = append(wheres, fmt.Sprintf("d.dibuat_pada < $%d", len(args)))
		}
	}

	orderDir := "DESC"
	if strings.EqualFold(strings.TrimSpace(order), "asc") {
		orderDir = "ASC"
	}

	args = append(args, limit)
	limitPos := len(args)
	args = append(args, offset)
	offsetPos := len(args)

	q := fmt.Sprintf(`
SELECT d.id, d.member_id, COALESCE(d.ref_id, ''), m.nama AS member_nama, d.bank_id, d.bank_nama, d.bank_nomor_rekening, d.bank_atas_nama, d.amount, COALESCE(d.requested_amount, 0), COALESCE(d.unique_code, 0), COALESCE(d.approved_amount, 0), d.metode, d.bukti_url, d.status, d.note, d.dibuat_pada, d.diproses_pada, d.diproses_oleh,
       p.nama as diproses_nama
FROM public.deposit_request d
LEFT JOIN public.member m ON m.id = d.member_id
LEFT JOIN public.member p ON p.id = d.diproses_oleh
WHERE %s
ORDER BY d.dibuat_pada %s, d.id %s
LIMIT $%d OFFSET $%d
`, strings.Join(wheres, " AND "), orderDir, orderDir, limitPos, offsetPos)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDepositRows(rows)
}

func (r *DepositRepository) ConfirmTicketTransfer(ctx context.Context, memberID, reqID int64) (*DepositRequestRow, error) {
	if memberID <= 0 || reqID <= 0 {
		return nil, errors.New("invalid member_id/request_id")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
UPDATE public.deposit_request
SET status = 'pending'
WHERE id = $1
  AND member_id = $2
  AND status = 'ticket'
`, reqID, memberID)
	if err != nil {
		return nil, err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return nil, errors.New("tiket tidak ditemukan atau sudah dikirim")
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

func (r *DepositRepository) GetMemberTicket(ctx context.Context, memberID, reqID int64) (*DepositRequestRow, error) {
	if memberID <= 0 || reqID <= 0 {
		return nil, errors.New("invalid member_id/request_id")
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT d.id, d.member_id, COALESCE(d.ref_id, ''), m.nama AS member_nama, d.bank_id, d.bank_nama, d.bank_nomor_rekening, d.bank_atas_nama, d.amount, COALESCE(d.requested_amount, 0), COALESCE(d.unique_code, 0), COALESCE(d.approved_amount, 0), d.metode, d.bukti_url, d.status, d.note, d.dibuat_pada, d.diproses_pada, d.diproses_oleh,
       p.nama as diproses_nama
FROM public.deposit_request d
LEFT JOIN public.member m ON m.id = d.member_id
LEFT JOIN public.member p ON p.id = d.diproses_oleh
WHERE d.id = $1
  AND d.member_id = $2
  AND d.status = 'ticket'
LIMIT 1
`, reqID, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanDepositRows(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, sql.ErrNoRows
	}
	return &items[0], nil
}

func (r *DepositRepository) CancelTicket(ctx context.Context, memberID, reqID int64) (*DepositRequestRow, error) {
	return r.CancelTicketWithNote(ctx, memberID, reqID, "")
}

func (r *DepositRepository) CancelTicketWithNote(ctx context.Context, memberID, reqID int64, note string) (*DepositRequestRow, error) {
	if memberID <= 0 || reqID <= 0 {
		return nil, errors.New("invalid member_id/request_id")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
UPDATE public.deposit_request
SET status = 'cancelled',
    note = CASE
      WHEN NULLIF($3::text, '') IS NULL THEN note
      WHEN NULLIF(TRIM(COALESCE(note, '')), '') IS NULL THEN $3::text
      ELSE note || ' | ' || $3::text
    END
WHERE id = $1
  AND member_id = $2
  AND status = 'ticket'
`, reqID, memberID, strings.TrimSpace(note))
	if err != nil {
		return nil, err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return nil, errors.New("tiket tidak ditemukan atau tidak bisa dibatalkan")
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

func scanDepositRows(rows *sql.Rows) ([]DepositRequestRow, error) {
	out := make([]DepositRequestRow, 0, 32)
	for rows.Next() {
		d, err := scanDepositRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *d)
	}
	return out, rows.Err()
}

type depositRowScanner interface {
	Scan(dest ...any) error
}

func scanDepositRow(row depositRowScanner) (*DepositRequestRow, error) {
	var d DepositRequestRow
	var note sql.NullString
	var memberNama sql.NullString
	var diprosesNama sql.NullString
	var bankNama sql.NullString
	var bankNomor sql.NullString
	var bankAtasNama sql.NullString
	var metode sql.NullString
	var buktiURL sql.NullString
	var status sql.NullString
	var bankID sql.NullInt64
	var requested sql.NullInt64
	var uniqueCode sql.NullInt64
	var approved sql.NullInt64
	if err := row.Scan(
		&d.ID, &d.MemberID, &d.RefID, &memberNama, &bankID, &bankNama, &bankNomor, &bankAtasNama, &d.Amount, &requested, &uniqueCode, &approved, &metode, &buktiURL, &status,
		&note, &d.DibuatPada, &d.DiprosesPada, &d.DiprosesOleh, &diprosesNama,
	); err != nil {
		return nil, err
	}
	if bankID.Valid {
		v := bankID.Int64
		d.BankID = &v
	}
	if bankNama.Valid {
		d.BankNama = bankNama.String
	}
	if bankNomor.Valid {
		d.BankNomor = bankNomor.String
	}
	if bankAtasNama.Valid {
		d.BankAtasNama = bankAtasNama.String
	}
	if metode.Valid {
		d.Metode = metode.String
	}
	if buktiURL.Valid {
		d.BuktiURL = buktiURL.String
	}
	if status.Valid {
		d.Status = status.String
	}
	if note.Valid {
		d.Note = note.String
	}
	if requested.Valid {
		d.Requested = requested.Int64
	}
	if uniqueCode.Valid {
		d.UniqueCode = uniqueCode.Int64
	}
	if approved.Valid {
		d.Approved = approved.Int64
	}
	if memberNama.Valid {
		v := memberNama.String
		d.MemberNama = &v
	}
	if diprosesNama.Valid {
		v := diprosesNama.String
		d.DiprosesNama = &v
	}
	d.HydrateTicketFields()
	return &d, nil
}

func (r *DepositRepository) getByIDTx(ctx context.Context, tx *sql.Tx, id int64) (*DepositRequestRow, error) {
	return scanDepositRow(tx.QueryRowContext(ctx, `
SELECT d.id, d.member_id, COALESCE(d.ref_id, ''), m.nama AS member_nama, d.bank_id, d.bank_nama, d.bank_nomor_rekening, d.bank_atas_nama, d.amount, COALESCE(d.requested_amount, 0), COALESCE(d.unique_code, 0), COALESCE(d.approved_amount, 0), d.metode, d.bukti_url, d.status, d.note, d.dibuat_pada, d.diproses_pada, d.diproses_oleh,
       p.nama as diproses_nama
FROM public.deposit_request d
LEFT JOIN public.member m ON m.id = d.member_id
LEFT JOIN public.member p ON p.id = d.diproses_oleh
WHERE d.id = $1
`, id))
}

func (r *DepositRepository) GetByRefID(ctx context.Context, refID string) (*DepositRequestRow, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT d.id, d.member_id, COALESCE(d.ref_id, ''), m.nama AS member_nama, d.bank_id, d.bank_nama, d.bank_nomor_rekening, d.bank_atas_nama, d.amount, COALESCE(d.requested_amount, 0), COALESCE(d.unique_code, 0), COALESCE(d.approved_amount, 0), d.metode, d.bukti_url, d.status, d.note, d.dibuat_pada, d.diproses_pada, d.diproses_oleh,
       p.nama as diproses_nama
FROM public.deposit_request d
LEFT JOIN public.member m ON m.id = d.member_id
LEFT JOIN public.member p ON p.id = d.diproses_oleh
WHERE TRIM(COALESCE(d.ref_id, '')) = $1
LIMIT 1
`, strings.TrimSpace(refID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanDepositRows(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, sql.ErrNoRows
	}
	return &items[0], nil
}

func (r *DepositRepository) GetVARequestByID(ctx context.Context, reqID int64) (*DepositRequestRow, error) {
	if reqID <= 0 {
		return nil, errors.New("invalid req_id")
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT d.id, d.member_id, COALESCE(d.ref_id, ''), m.nama AS member_nama, d.bank_id, d.bank_nama, d.bank_nomor_rekening, d.bank_atas_nama, d.amount, COALESCE(d.requested_amount, 0), COALESCE(d.unique_code, 0), COALESCE(d.approved_amount, 0), d.metode, d.bukti_url, d.status, d.note, d.dibuat_pada, d.diproses_pada, d.diproses_oleh,
       p.nama as diproses_nama
FROM public.deposit_request d
LEFT JOIN public.member m ON m.id = d.member_id
LEFT JOIN public.member p ON p.id = d.diproses_oleh
WHERE d.id = $1
  AND LOWER(TRIM(COALESCE(d.metode, ''))) = 'va'
LIMIT 1
`, reqID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanDepositRows(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, sql.ErrNoRows
	}
	return &items[0], nil
}

func (r *DepositRepository) UpdateQrisPending(ctx context.Context, refID, qrURL, note string) error {
	if strings.TrimSpace(refID) == "" {
		return errors.New("ref_id required")
	}
	_, err := r.db.ExecContext(ctx, `
UPDATE public.deposit_request
SET bukti_url = CASE WHEN NULLIF($2, '') IS NULL THEN bukti_url ELSE $2 END,
    note = CASE WHEN NULLIF($3, '') IS NULL THEN note ELSE $3 END
WHERE TRIM(COALESCE(ref_id, '')) = $1
`, strings.TrimSpace(refID), strings.TrimSpace(qrURL), strings.TrimSpace(note))
	return err
}
