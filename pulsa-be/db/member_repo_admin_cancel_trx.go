package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrTransaksiNotPending = errors.New("transaksi tidak dalam status pending")

type AdminCancelTrxResult struct {
	TrxID          int64  `json:"trx_id"`
	RefID          string `json:"ref_id"`
	StatusBefore   string `json:"status_before"`
	StatusAfter    string `json:"status_after"`
	RefundAmount   int64  `json:"refund_amount"`
	SaldoSebelum   int64  `json:"saldo_sebelum"`
	SaldoSesudah   int64  `json:"saldo_sesudah"`
	AlasanBatal    string `json:"alasan_batal_admin"`
	DibatalkanOleh int64  `json:"dibatalkan_oleh_admin_id"`
}

func (r *MemberRepo) AdminCancelPendingTransaksi(ctx context.Context, adminID, trxID int64, reason string) (*AdminCancelTrxResult, error) {
	if adminID <= 0 || trxID <= 0 {
		return nil, errors.New("admin_id/trx_id invalid")
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, errors.New("alasan wajib diisi")
	}

	tx, err := r.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var (
		memberID       int64
		refID          string
		statusBefore   string
		biayaPerkiraan int64
	)
	err = tx.QueryRowContext(ctx, `
SELECT member_id, ref_id, status, biaya_perkiraan
FROM public.transaksi_member
WHERE id = $1
FOR UPDATE
`, trxID).Scan(&memberID, &refID, &statusBefore, &biayaPerkiraan)
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

	var saldoBefore, saldoAfter int64
	refund := int64(0)
	if biayaPerkiraan > 0 {
		err = tx.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_member
WHERE member_id = $1
FOR UPDATE
`, memberID).Scan(&saldoBefore)
		if err != nil {
			return nil, err
		}

		refund = biayaPerkiraan
		saldoAfter = saldoBefore + refund

		if _, err = tx.ExecContext(ctx, `
UPDATE public.dompet_member
SET saldo = $2, diperbarui_pada = now()
WHERE member_id = $1
`, memberID, saldoAfter); err != nil {
			return nil, err
		}

		if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah)
VALUES
  ($1,$2,'CREDIT',$3,$4,NULLIF($5,''),$6,$7)
`, memberID, refID, refund, "ADMIN_CANCEL_REFUND", ket, saldoBefore, saldoAfter); err != nil {
			return nil, err
		}
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
