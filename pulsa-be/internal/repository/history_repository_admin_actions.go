package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (r *HistoryRepository) AdminCancelPendingTransaksi(ctx context.Context, adminID, trxID int64, reason string, allowSuccessCancel bool) (*AdminCancelTrxResult, error) {
	if adminID <= 0 || trxID <= 0 {
		return nil, errors.New("admin_id/trx_id invalid")
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, errors.New("alasan wajib diisi")
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var (
		memberID          int64
		refID             string
		statusBefore      string
		keteranganBefore  sql.NullString
		biayaPerkiraan    int64
		biayaAktualBefore int64
	)
	err = tx.QueryRowContext(ctx, `
SELECT member_id, ref_id, status, keterangan, biaya_perkiraan, COALESCE(biaya_aktual,0)
FROM public.transaksi_member
WHERE id = $1
FOR UPDATE
`, trxID).Scan(&memberID, &refID, &statusBefore, &keteranganBefore, &biayaPerkiraan, &biayaAktualBefore)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("transaksi tidak ditemukan")
	}
	if err != nil {
		return nil, err
	}

	statusNow := strings.ToLower(strings.TrimSpace(statusBefore))
	if statusNow != "pending" && statusNow != "success" {
		return nil, ErrTransaksiNotPending
	}
	if statusNow == "success" && !allowSuccessCancel {
		return nil, errors.New("transaksi sukses hanya bisa dibatalkan manual oleh admin")
	}

	if statusNow == "success" {
		if _, err := tx.ExecContext(ctx, `SELECT set_config('p24.admin_manual_cancel_success', '1', true)`); err != nil {
			return nil, err
		}
	}

	statusAfter := "failed"
	ket := fmt.Sprintf("dibatalkan admin: %s", reason)
	_, err = tx.ExecContext(ctx, `
UPDATE public.transaksi_member
SET status = $2,
    keterangan = $3,
    biaya_aktual = 0,
    dibatalkan_oleh_admin_id = $4,
    dibatalkan_pada = $5,
    alasan_batal_admin = $6,
    diperbarui_pada = now()
WHERE id = $1
`, trxID, statusAfter, ket, adminID, time.Now().UTC(), reason)
	if err != nil {
		return nil, err
	}
	if err := insertTrxMemberStatusLog(ctx, tx, TrxMemberStatusLogInput{
		TrxID:              trxID,
		StatusSebelum:      statusBefore,
		StatusSesudah:      statusAfter,
		KeteranganSebelum:  keteranganBefore.String,
		KeteranganSesudah:  ket,
		BiayaAktualSebelum: biayaAktualBefore,
		BiayaAktualSesudah: 0,
		Aksi:               "admin_cancel",
		DiubahOleh:         &adminID,
	}); err != nil {
		return nil, err
	}

	var saldoBefore, saldoAfter int64
	refund := int64(0)
	if biayaPerkiraan > 0 {
		// Refund saldo sudah ditangani trigger DB saat status berubah ke failed.
		refund = biayaPerkiraan
		err = tx.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_member
WHERE member_id = $1
`, memberID).Scan(&saldoAfter)
		if err != nil {
			return nil, err
		}
		saldoBefore = saldoAfter - refund
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &AdminCancelTrxResult{
		TrxID:          trxID,
		RefID:          refID,
		StatusBefore:   statusBefore,
		StatusAfter:    statusAfter,
		RefundAmount:   refund,
		SaldoSebelum:   saldoBefore,
		SaldoSesudah:   saldoAfter,
		AlasanBatal:    reason,
		DibatalkanOleh: adminID,
	}, nil
}

func (r *HistoryRepository) AdminCompletePendingTransaksi(ctx context.Context, adminID, trxID int64, reason string) (*AdminCompleteTrxResult, error) {
	if adminID <= 0 || trxID <= 0 {
		return nil, errors.New("admin_id/trx_id invalid")
	}
	reason = strings.TrimSpace(reason)

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var (
		refID            string
		statusBefore     string
		keteranganBefore sql.NullString
		biayaPerkiraan   int64
		biayaAktual      int64
	)
	err = tx.QueryRowContext(ctx, `
SELECT ref_id, status, keterangan, biaya_perkiraan, COALESCE(biaya_aktual,0)
FROM public.transaksi_member
WHERE id = $1
FOR UPDATE
`, trxID).Scan(&refID, &statusBefore, &keteranganBefore, &biayaPerkiraan, &biayaAktual)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("transaksi tidak ditemukan")
	}
	if err != nil {
		return nil, err
	}

	statusNow := strings.ToLower(strings.TrimSpace(statusBefore))
	if statusNow != "pending" && statusNow != "failed" {
		return nil, ErrTransaksiNotPending
	}

	if biayaAktual <= 0 {
		biayaAktual = biayaPerkiraan
	}

	// Jika transaksi sebelumnya failed/cancel lalu ingin disukseskan admin,
	// saldo member harus didebit kembali agar tidak terjadi kelebihan refund.
	if statusNow == "failed" && biayaAktual > 0 {
		var saldoBefore int64
		if err := tx.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_member
WHERE member_id = (
  SELECT member_id FROM public.transaksi_member WHERE id = $1
)
FOR UPDATE
`, trxID).Scan(&saldoBefore); err != nil {
			return nil, err
		}
		if saldoBefore < biayaAktual {
			return nil, errors.New("saldo member tidak cukup untuk sukseskan transaksi failed")
		}
		saldoAfter := saldoBefore - biayaAktual

		if _, err := tx.ExecContext(ctx, `
UPDATE public.dompet_member
SET saldo = $2, diperbarui_pada = now()
WHERE member_id = (
  SELECT member_id FROM public.transaksi_member WHERE id = $1
)
`, trxID, saldoAfter); err != nil {
			return nil, err
		}

		catatanDebit := "debit balik saat admin sukseskan transaksi failed"
		if reason != "" {
			catatanDebit = "debit balik admin complete: " + reason
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah)
SELECT member_id, ref_id, 'DEBIT', $2, 'ADMIN_COMPLETE_REDEBIT', NULLIF($3,''), $4, $5
FROM public.transaksi_member
WHERE id = $1
`, trxID, biayaAktual, catatanDebit, saldoBefore, saldoAfter); err != nil {
			return nil, err
		}
	}

	statusAfter := "success"
	catatan := fmt.Sprintf("diselesaikan admin (id:%d)", adminID)
	if reason != "" {
		catatan = fmt.Sprintf("diselesaikan admin: %s", reason)
	}

	_, err = tx.ExecContext(ctx, `
UPDATE public.transaksi_member
SET status = $2,
    keterangan = $3,
    biaya_aktual = $4,
    diperbarui_pada = now()
WHERE id = $1
`, trxID, statusAfter, catatan, biayaAktual)
	if err != nil {
		return nil, err
	}
	if err := insertTrxMemberStatusLog(ctx, tx, TrxMemberStatusLogInput{
		TrxID:              trxID,
		StatusSebelum:      statusBefore,
		StatusSesudah:      statusAfter,
		KeteranganSebelum:  keteranganBefore.String,
		KeteranganSesudah:  catatan,
		BiayaAktualSebelum: biayaAktual,
		BiayaAktualSesudah: biayaAktual,
		Aksi:               "admin_complete",
		DiubahOleh:         &adminID,
	}); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &AdminCompleteTrxResult{
		TrxID:          trxID,
		RefID:          refID,
		StatusBefore:   statusBefore,
		StatusAfter:    statusAfter,
		BiayaAktual:    biayaAktual,
		DiselesaikanBy: adminID,
		Catatan:        catatan,
	}, nil
}
