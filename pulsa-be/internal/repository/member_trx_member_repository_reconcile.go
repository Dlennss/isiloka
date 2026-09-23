package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

func (r *MemberTrxMemberRepository) UpdateTransaksiMemberStatus(ctx context.Context, trxID int64, status string, keterangan string, biayaAktual int64) error {
	return updateTrxMemberStatusWithAudit(ctx, r.db, trxID, status, keterangan, biayaAktual, "system_update", nil)
}

func (r *MemberTrxMemberRepository) ForceReopenFailedToPending(ctx context.Context, trxID int64, keterangan string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var (
		statusBefore     string
		keteranganBefore sql.NullString
		biayaAktualPrev  int64
	)
	err = tx.QueryRowContext(ctx, `
SELECT status, keterangan, COALESCE(biaya_aktual, 0)
FROM public.transaksi_member
WHERE id = $1
FOR UPDATE
`, trxID).Scan(&statusBefore, &keteranganBefore, &biayaAktualPrev)
	if err != nil {
		return err
	}

	statusNow := strings.ToLower(strings.TrimSpace(statusBefore))
	if statusNow == "pending" {
		return tx.Commit()
	}
	if statusNow != "failed" {
		return sql.ErrNoRows
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE public.transaksi_member
SET status = 'pending',
    keterangan = NULLIF($2,''),
    biaya_aktual = 0,
    diperbarui_pada = now()
WHERE id = $1
`, trxID, keterangan); err != nil {
		return err
	}

	if err := insertTrxMemberStatusLog(ctx, tx, TrxMemberStatusLogInput{
		TrxID:              trxID,
		StatusSebelum:      statusBefore,
		StatusSesudah:      "pending",
		KeteranganSebelum:  keteranganBefore.String,
		KeteranganSesudah:  keterangan,
		BiayaAktualSebelum: biayaAktualPrev,
		BiayaAktualSesudah: 0,
		Aksi:               "retry_same_ref_reopen",
		DiubahOleh:         nil,
	}); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *MemberTrxMemberRepository) ForceReconcileFailedToSuccess(ctx context.Context, trxID int64, keterangan string, biayaAktual int64, hargaJavapay int64, hargaMember int64) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var (
		memberID         int64
		refID            string
		statusBefore     string
		keteranganBefore sql.NullString
		biayaAktualPrev  int64
	)
	err = tx.QueryRowContext(ctx, `
SELECT member_id, ref_id, status, keterangan, COALESCE(biaya_aktual, 0)
FROM public.transaksi_member
WHERE id = $1
FOR UPDATE
`, trxID).Scan(&memberID, &refID, &statusBefore, &keteranganBefore, &biayaAktualPrev)
	if err != nil {
		return err
	}

	statusNow := strings.ToLower(strings.TrimSpace(statusBefore))
	if statusNow == "success" {
		return tx.Commit()
	}
	if statusNow != "failed" {
		return sql.ErrNoRows
	}
	if biayaAktual <= 0 {
		return errors.New("biaya aktual invalid untuk reconcile success")
	}

	var netDebit int64
	if err := tx.QueryRowContext(ctx, `
SELECT COALESCE(SUM(
  CASE
    WHEN UPPER(COALESCE(arah, '')) = 'DEBIT' THEN jumlah
    WHEN UPPER(COALESCE(arah, '')) IN ('CREDIT', 'KREDIT') THEN -jumlah
    ELSE 0
  END
), 0)
FROM public.mutasi_dompet
WHERE member_id = $1
  AND ref_id = $2
`, memberID, refID).Scan(&netDebit); err != nil {
		return err
	}

	redebitAmount := biayaAktual - netDebit
	if redebitAmount > 0 {
		var saldoBefore int64
		if err := tx.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_member
WHERE member_id = $1
FOR UPDATE
`, memberID).Scan(&saldoBefore); err != nil {
			return err
		}
		if saldoBefore < redebitAmount {
			return ErrInsufficientBalance
		}
		saldoAfter := saldoBefore - redebitAmount

		if _, err := tx.ExecContext(ctx, `
UPDATE public.dompet_member
SET saldo = $2, diperbarui_pada = now()
WHERE member_id = $1
`, memberID, saldoAfter); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah)
VALUES
  ($1,$2,'DEBIT',$3,'TRX_SETTLE',NULLIF($4,''),$5,$6)
`, memberID, refID, redebitAmount, "debit balik otomatis saat STATUS-PAY reconcile provider success", saldoBefore, saldoAfter); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE public.transaksi_member
SET status = 'success',
    keterangan = NULLIF($2,''),
    biaya_aktual = $3,
    harga_javapay = $4,
    harga_member = $5,
    diperbarui_pada = now()
WHERE id = $1
`, trxID, keterangan, biayaAktual, hargaJavapay, hargaMember); err != nil {
		return err
	}

	if err := insertTrxMemberStatusLog(ctx, tx, TrxMemberStatusLogInput{
		TrxID:              trxID,
		StatusSebelum:      statusBefore,
		StatusSesudah:      "success",
		KeteranganSebelum:  keteranganBefore.String,
		KeteranganSesudah:  keterangan,
		BiayaAktualSebelum: biayaAktualPrev,
		BiayaAktualSesudah: biayaAktual,
		Aksi:               "status_pay_reconcile_success",
		DiubahOleh:         nil,
	}); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *MemberTrxMemberRepository) UpdateTransaksiMemberSettle(ctx context.Context, trxID int64, status string, ket string, biayaAktual int64, hargaJavapay int64, hargaMember int64) error {
	return updateTrxMemberSettleWithAudit(ctx, r.db, trxID, status, ket, biayaAktual, hargaJavapay, hargaMember, "settle_update")
}

func (r *MemberTrxMemberRepository) UpdateTransaksiMemberFee(ctx context.Context, trxID int64, feeMember int64) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE public.transaksi_member
SET fee_member_rp=$2
WHERE id=$1
`, trxID, feeMember)
	return err
}
