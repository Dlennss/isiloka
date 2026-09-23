package db

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// ===============================
// CREDIT DOMPET (TX)
// ===============================
func (r *MemberRepo) creditDompetTx(
	ctx context.Context,
	tx *sql.Tx,
	memberID int64,
	jumlah int64,
	alasan string,
	catatan string,
	refID string,
) error {
	if jumlah <= 0 {
		return errors.New("jumlah harus > 0")
	}

	// pastikan dompet ada
	_, err := tx.ExecContext(ctx, `
INSERT INTO dompet_member (member_id, saldo)
VALUES ($1, 0)
ON CONFLICT (member_id) DO NOTHING
`, memberID)
	if err != nil {
		return err
	}

	// lock row + ambil saldo sebelum
	var saldoSebelum int64
	err = tx.QueryRowContext(ctx, `
SELECT saldo
FROM dompet_member
WHERE member_id = $1
FOR UPDATE
`, memberID).Scan(&saldoSebelum)
	if err != nil {
		return err
	}

	saldoSesudah := saldoSebelum + jumlah

	// update saldo
	_, err = tx.ExecContext(ctx, `
UPDATE dompet_member
SET saldo = $1, diperbarui_pada = now()
WHERE member_id = $2
`, saldoSesudah, memberID)
	if err != nil {
		return err
	}

	// insert mutasi
	_, err = tx.ExecContext(ctx, `
INSERT INTO mutasi_dompet
(member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
VALUES ($1,$2,'CREDIT',$3,$4,$5,$6,$7,$8)
`, memberID, refID, jumlah, alasan, catatan, saldoSebelum, saldoSesudah, time.Now())

	return err
}

// ===============================
// DEBIT DOMPET (TX)
// ===============================
func (r *MemberRepo) debitDompetTx(
	ctx context.Context,
	tx *sql.Tx,
	memberID int64,
	jumlah int64,
	alasan string,
	catatan string,
	refID string,
) error {
	if jumlah <= 0 {
		return errors.New("jumlah harus > 0")
	}

	// pastikan dompet ada
	_, err := tx.ExecContext(ctx, `
INSERT INTO dompet_member (member_id, saldo)
VALUES ($1, 0)
ON CONFLICT (member_id) DO NOTHING
`, memberID)
	if err != nil {
		return err
	}

	// lock row + ambil saldo sebelum
	var saldoSebelum int64
	err = tx.QueryRowContext(ctx, `
SELECT saldo
FROM dompet_member
WHERE member_id = $1
FOR UPDATE
`, memberID).Scan(&saldoSebelum)
	if err != nil {
		return err
	}

	if saldoSebelum < jumlah {
		return errors.New("saldo tidak cukup")
	}

	saldoSesudah := saldoSebelum - jumlah

	// update saldo
	_, err = tx.ExecContext(ctx, `
UPDATE dompet_member
SET saldo = $1, diperbarui_pada = now()
WHERE member_id = $2
`, saldoSesudah, memberID)
	if err != nil {
		return err
	}

	// insert mutasi
	_, err = tx.ExecContext(ctx, `
INSERT INTO mutasi_dompet
(member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
VALUES ($1,$2,'DEBIT',$3,$4,$5,$6,$7,$8)
`, memberID, refID, jumlah, alasan, catatan, saldoSebelum, saldoSesudah, time.Now())

	return err
}
