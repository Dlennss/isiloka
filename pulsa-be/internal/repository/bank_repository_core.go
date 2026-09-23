package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type BankRepository struct {
	db *sql.DB
}

const (
	systemQrisBankName             = "QRIS"
	systemQRTPBankName             = "QRTP"
	pulsaKilatDepositBankName      = "BNI"
	pulsaKilatDepositAccountNumber = "1955637480"
	pulsaKilatDepositAccountHolder = "M Sansan Irfanda"
)

func NewBankRepository(db *sql.DB) *BankRepository {
	return &BankRepository{db: db}
}

// EnsurePulsaKilatDepositBank keeps the public deposit account available after
// deployment, including installations where SQL migrations are applied later.
func (r *BankRepository) EnsurePulsaKilatDepositBank(ctx context.Context) error {
	if r == nil || r.db == nil {
		return errors.New("bank repository belum dikonfigurasi")
	}

	result, err := r.db.ExecContext(ctx, `
UPDATE public.bank
SET nama = $2,
    nomor_rekening = $1,
    atas_nama = $3,
    aktif = true,
    admin_staff_only = false,
    diubah_pada = now()
WHERE regexp_replace(COALESCE(nomor_rekening, ''), '[^0-9]', '', 'g') = $1
`, pulsaKilatDepositAccountNumber, pulsaKilatDepositBankName, pulsaKilatDepositAccountHolder)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected > 0 {
		return nil
	}

	_, err = r.db.ExecContext(ctx, `
INSERT INTO public.bank
  (nama, nomor_rekening, atas_nama, saldo, aktif, admin_staff_only, dibuat_pada, diubah_pada)
VALUES
  ($1, $2, $3, 0, true, false, now(), now())
ON CONFLICT ((lower(trim(nama)))) DO UPDATE
SET nomor_rekening = EXCLUDED.nomor_rekening,
    atas_nama = EXCLUDED.atas_nama,
    aktif = true,
    admin_staff_only = false,
    diubah_pada = now()
`, pulsaKilatDepositBankName, pulsaKilatDepositAccountNumber, pulsaKilatDepositAccountHolder)
	return err
}

func (r *BankRepository) List(ctx context.Context, includeAdminStaffOnly bool) ([]BankRow, error) {
	if _, err := r.EnsureSystemQrisBank(ctx); err != nil {
		return nil, err
	}
	if _, err := r.EnsureSystemQRTPBank(ctx); err != nil {
		return nil, err
	}
	where := ""
	if !includeAdminStaffOnly {
		where = "WHERE COALESCE(admin_staff_only, false) = false"
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT id, nama, nomor_rekening, atas_nama, saldo, aktif, COALESCE(admin_staff_only, false), dibuat_pada, diubah_pada
FROM public.bank
`+where+`
ORDER BY
  CASE
    WHEN lower(trim(nama)) = lower(trim($1)) THEN COALESCE((SELECT id::numeric + 0.1 FROM public.bank WHERE nomor_rekening = '031901002025568' LIMIT 1), id::numeric)
    ELSE id::numeric
  END,
  id ASC
`, systemQRTPBankName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]BankRow, 0, 32)
	for rows.Next() {
		var (
			item   BankRow
			dibuat sql.NullTime
			diubah sql.NullTime
		)
		if err := rows.Scan(
			&item.ID,
			&item.Nama,
			&item.NomorRekening,
			&item.AtasNama,
			&item.Saldo,
			&item.Aktif,
			&item.AdminStaffOnly,
			&dibuat,
			&diubah,
		); err != nil {
			return nil, err
		}
		if dibuat.Valid {
			v := dibuat.Time
			item.DibuatPada = &v
		}
		if diubah.Valid {
			v := diubah.Time
			item.DiubahPada = &v
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *BankRepository) ListActive(ctx context.Context) ([]BankRow, error) {
	if _, err := r.EnsureSystemQrisBank(ctx); err != nil {
		return nil, err
	}
	if _, err := r.EnsureSystemQRTPBank(ctx); err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT id, nama, nomor_rekening, atas_nama, saldo, aktif, COALESCE(admin_staff_only, false), dibuat_pada, diubah_pada
FROM public.bank
WHERE aktif = true
  AND COALESCE(admin_staff_only, false) = false
ORDER BY
  CASE
    WHEN lower(trim(nama)) = lower(trim($1)) THEN COALESCE((SELECT id::numeric + 0.1 FROM public.bank WHERE nomor_rekening = '031901002025568' LIMIT 1), id::numeric)
    ELSE id::numeric
  END,
  id ASC
`, systemQRTPBankName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]BankRow, 0, 16)
	for rows.Next() {
		var (
			item   BankRow
			dibuat sql.NullTime
			diubah sql.NullTime
		)
		if err := rows.Scan(
			&item.ID,
			&item.Nama,
			&item.NomorRekening,
			&item.AtasNama,
			&item.Saldo,
			&item.Aktif,
			&item.AdminStaffOnly,
			&dibuat,
			&diubah,
		); err != nil {
			return nil, err
		}
		if dibuat.Valid {
			v := dibuat.Time
			item.DibuatPada = &v
		}
		if diubah.Valid {
			v := diubah.Time
			item.DiubahPada = &v
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *BankRepository) Get(ctx context.Context, id int64) (*BankRow, error) {
	return r.GetVisible(ctx, id, true)
}

func (r *BankRepository) GetVisible(ctx context.Context, id int64, includeAdminStaffOnly bool) (*BankRow, error) {
	var (
		item   BankRow
		dibuat sql.NullTime
		diubah sql.NullTime
	)
	visibilityClause := ""
	if !includeAdminStaffOnly {
		visibilityClause = "AND COALESCE(admin_staff_only, false) = false"
	}
	err := r.db.QueryRowContext(ctx, `
SELECT id, nama, nomor_rekening, atas_nama, saldo, aktif, COALESCE(admin_staff_only, false), dibuat_pada, diubah_pada
FROM public.bank
WHERE id = $1
`+visibilityClause+`
LIMIT 1
`, id).Scan(
		&item.ID,
		&item.Nama,
		&item.NomorRekening,
		&item.AtasNama,
		&item.Saldo,
		&item.Aktif,
		&item.AdminStaffOnly,
		&dibuat,
		&diubah,
	)
	if err != nil {
		return nil, err
	}
	if dibuat.Valid {
		v := dibuat.Time
		item.DibuatPada = &v
	}
	if diubah.Valid {
		v := diubah.Time
		item.DiubahPada = &v
	}
	return &item, nil
}

func (r *BankRepository) GetActiveByAccountNumber(ctx context.Context, accountNumber string, includeAdminStaffOnly bool) (*BankRow, error) {
	accountNumber = strings.TrimSpace(accountNumber)
	if accountNumber == "" {
		return nil, errors.New("nomor_rekening required")
	}
	var (
		item   BankRow
		dibuat sql.NullTime
		diubah sql.NullTime
	)
	visibilityClause := ""
	if !includeAdminStaffOnly {
		visibilityClause = "AND COALESCE(admin_staff_only, false) = false"
	}
	err := r.db.QueryRowContext(ctx, `
SELECT id, nama, nomor_rekening, atas_nama, saldo, aktif, COALESCE(admin_staff_only, false), dibuat_pada, diubah_pada
FROM public.bank
WHERE regexp_replace(COALESCE(nomor_rekening, ''), '[^0-9]', '', 'g') = regexp_replace($1, '[^0-9]', '', 'g')
  AND aktif = true
`+visibilityClause+`
ORDER BY id ASC
LIMIT 1
`, accountNumber).Scan(
		&item.ID,
		&item.Nama,
		&item.NomorRekening,
		&item.AtasNama,
		&item.Saldo,
		&item.Aktif,
		&item.AdminStaffOnly,
		&dibuat,
		&diubah,
	)
	if err != nil {
		return nil, err
	}
	if dibuat.Valid {
		v := dibuat.Time
		item.DibuatPada = &v
	}
	if diubah.Valid {
		v := diubah.Time
		item.DiubahPada = &v
	}
	return &item, nil
}

func (r *BankRepository) EnsureSystemQrisBank(ctx context.Context) (int64, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	id, err := r.ensureSystemQrisBankTx(ctx, tx)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *BankRepository) EnsureSystemQRTPBank(ctx context.Context) (int64, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	id, err := r.ensureSystemQRTPBankTx(ctx, tx)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *BankRepository) CreditSystemQrisIfNeeded(ctx context.Context, refID string, amount int64, reason, note string, memberID *int64) (int64, bool, error) {
	if strings.TrimSpace(refID) == "" {
		return 0, false, errors.New("ref_id required")
	}
	if amount <= 0 {
		return 0, false, errors.New("amount must be > 0")
	}
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, false, err
	}
	defer func() { _ = tx.Rollback() }()

	bankID, err := r.ensureSystemQrisBankTx(ctx, tx)
	if err != nil {
		return 0, false, err
	}
	var existing int64
	err = tx.QueryRowContext(ctx, `
SELECT id
FROM public.mutasi_bank
WHERE bank_id = $1
  AND TRIM(COALESCE(ref_id, '')) = $2
  AND TRIM(COALESCE(alasan, '')) = $3
LIMIT 1
`, bankID, strings.TrimSpace(refID), strings.TrimSpace(reason)).Scan(&existing)
	if err == nil {
		var saldo int64
		if err := tx.QueryRowContext(ctx, `SELECT saldo FROM public.bank WHERE id = $1`, bankID).Scan(&saldo); err != nil {
			return 0, false, err
		}
		if err := tx.Commit(); err != nil {
			return 0, false, err
		}
		return saldo, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, false, err
	}

	var before int64
	if err := tx.QueryRowContext(ctx, `
SELECT saldo
FROM public.bank
WHERE id = $1
FOR UPDATE
`, bankID).Scan(&before); err != nil {
		return 0, false, err
	}
	after := before + amount
	if _, err := tx.ExecContext(ctx, `
UPDATE public.bank
SET saldo = $2, diubah_pada = now()
WHERE id = $1
`, bankID, after); err != nil {
		return 0, false, err
	}

	var targetMember any
	if memberID != nil && *memberID > 0 {
		targetMember = *memberID
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO public.mutasi_bank
  (bank_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, member_id, dibuat_pada)
VALUES
  ($1,$2,'CREDIT',$3,$4,NULLIF($5,''),$6,$7,$8,now())
`, bankID, strings.TrimSpace(refID), amount, strings.TrimSpace(reason), strings.TrimSpace(note), before, after, targetMember); err != nil {
		return 0, false, err
	}

	if err := tx.Commit(); err != nil {
		return 0, false, err
	}
	return after, true, nil
}

func (r *BankRepository) CreditMemberDepositAndSystemQRTPIfNeeded(ctx context.Context, memberID int64, refID string, amount int64, note string) (memberCredited bool, bankCredited bool, err error) {
	refID = strings.TrimSpace(refID)
	note = strings.TrimSpace(note)
	if memberID <= 0 {
		return false, false, errors.New("member_id required")
	}
	if refID == "" {
		return false, false, errors.New("ref_id required")
	}
	if amount <= 0 {
		return false, false, errors.New("amount must be > 0")
	}
	if note == "" {
		note = "deposit api"
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return false, false, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var memberBefore int64
	if err = tx.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_member
WHERE member_id = $1
FOR UPDATE
`, memberID).Scan(&memberBefore); err != nil {
		return false, false, err
	}

	var memberExisting int64
	memberErr := tx.QueryRowContext(ctx, `
SELECT COUNT(1)
FROM public.mutasi_dompet
WHERE member_id = $1
  AND ref_id = $2
  AND LOWER(COALESCE(arah, '')) = 'credit'
  AND LOWER(COALESCE(alasan, '')) = 'deposit_api'
`, memberID, refID).Scan(&memberExisting)
	if memberErr != nil {
		return false, false, memberErr
	}
	if memberExisting == 0 {
		memberAfter := memberBefore + amount
		if _, err = tx.ExecContext(ctx, `
UPDATE public.dompet_member
SET saldo = $2, diperbarui_pada = now()
WHERE member_id = $1
`, memberID, memberAfter); err != nil {
			return false, false, err
		}
		if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah)
VALUES
  ($1,$2,'CREDIT',$3,'DEPOSIT_API',NULLIF($4,''),$5,$6)
`, memberID, refID, amount, note, memberBefore, memberAfter); err != nil {
			return false, false, err
		}
		memberCredited = true
	}

	bankID, err := r.ensureSystemQRTPBankTx(ctx, tx)
	if err != nil {
		return false, false, err
	}
	var bankExisting int64
	if err = tx.QueryRowContext(ctx, `
SELECT COUNT(1)
FROM public.mutasi_bank
WHERE bank_id = $1
  AND TRIM(COALESCE(ref_id, '')) = $2
  AND TRIM(COALESCE(alasan, '')) = 'SMPAY_WEDE_TRANSFER_IN'
`, bankID, refID).Scan(&bankExisting); err != nil {
		return false, false, err
	}
	if bankExisting == 0 {
		var bankBefore int64
		if err = tx.QueryRowContext(ctx, `
SELECT saldo
FROM public.bank
WHERE id = $1
FOR UPDATE
`, bankID).Scan(&bankBefore); err != nil {
			return false, false, err
		}
		bankAfter := bankBefore + amount
		if _, err = tx.ExecContext(ctx, `
UPDATE public.bank
SET saldo = $2, diubah_pada = now()
WHERE id = $1
`, bankID, bankAfter); err != nil {
			return false, false, err
		}
		if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_bank
  (bank_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, member_id, dibuat_pada)
VALUES
  ($1,$2,'CREDIT',$3,'SMPAY_WEDE_TRANSFER_IN',NULLIF($4,''),$5,$6,$7,now())
`, bankID, refID, amount, note, bankBefore, bankAfter, memberID); err != nil {
			return false, false, err
		}
		bankCredited = true
	}

	if err = tx.Commit(); err != nil {
		return false, false, err
	}
	return memberCredited, bankCredited, nil
}

func (r *BankRepository) ensureSystemQrisBankTx(ctx context.Context, tx *sql.Tx) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
SELECT id
FROM public.bank
WHERE lower(trim(nama)) = lower(trim($1))
LIMIT 1
FOR UPDATE
`, systemQrisBankName).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	if err := tx.QueryRowContext(ctx, `
INSERT INTO public.bank (nama, nomor_rekening, atas_nama, saldo, aktif, admin_staff_only, dibuat_pada, diubah_pada)
VALUES ($1, '', 'System QRIS', 0, true, false, now(), now())
RETURNING id
`, systemQrisBankName).Scan(&id); err != nil {
		if retryErr := tx.QueryRowContext(ctx, `
SELECT id
FROM public.bank
WHERE lower(trim(nama)) = lower(trim($1))
LIMIT 1
`, systemQrisBankName).Scan(&id); retryErr == nil {
			return id, nil
		}
		return 0, err
	}
	return id, nil
}

func (r *BankRepository) ensureSystemQRTPBankTx(ctx context.Context, tx *sql.Tx) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
SELECT id
FROM public.bank
WHERE lower(trim(nama)) = lower(trim($1))
LIMIT 1
FOR UPDATE
`, systemQRTPBankName).Scan(&id)
	if err == nil {
		if _, updateErr := tx.ExecContext(ctx, `
UPDATE public.bank
SET admin_staff_only = true,
    diubah_pada = now()
WHERE id = $1
  AND COALESCE(admin_staff_only, false) = false
`, id); updateErr != nil {
			return 0, updateErr
		}
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	if err := tx.QueryRowContext(ctx, `
INSERT INTO public.bank (nama, nomor_rekening, atas_nama, saldo, aktif, admin_staff_only, dibuat_pada, diubah_pada)
VALUES ($1, '', 'System QRTP', 0, true, true, now(), now())
RETURNING id
`, systemQRTPBankName).Scan(&id); err != nil {
		if retryErr := tx.QueryRowContext(ctx, `
SELECT id
FROM public.bank
WHERE lower(trim(nama)) = lower(trim($1))
LIMIT 1
`, systemQRTPBankName).Scan(&id); retryErr == nil {
			return id, nil
		}
		return 0, err
	}
	return id, nil
}
