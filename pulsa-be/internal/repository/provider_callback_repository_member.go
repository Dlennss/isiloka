package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

func (r *ProviderCallbackRepository) UpdateProviderTrxPerintah(ctx context.Context, id int64, perintah string) error {
	perintah = strings.TrimSpace(strings.ToUpper(perintah))
	if perintah == "" {
		return nil
	}
	_, err := r.db.ExecContext(ctx, `
UPDATE public.transaksi_provider
SET perintah = $2
WHERE id = $1
`, id, perintah)
	return err
}

type CallbackTrxMemberFull struct {
	ID                    int64
	MemberID              int64
	RefID                 string
	Perintah              string
	KodeProduk            string
	Tujuan                string
	Qty                   int64
	QtyProvider           int64
	Status                string
	Keterangan            string
	ChargeReceiverApplied bool
	BiayaPerkiraan        int64
	FeeMemberRp           int64
	HargaMember           int64
}

func scanCallbackTrxMember(row *sql.Row) (*CallbackTrxMemberFull, error) {
	var t CallbackTrxMemberFull
	err := row.Scan(
		&t.ID, &t.MemberID, &t.RefID, &t.Perintah, &t.KodeProduk, &t.Tujuan, &t.Qty, &t.Status, &t.QtyProvider,
		&t.Keterangan, &t.ChargeReceiverApplied, &t.BiayaPerkiraan, &t.FeeMemberRp, &t.HargaMember,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *ProviderCallbackRepository) GetTransaksiMemberByID(ctx context.Context, trxID int64) (*CallbackTrxMemberFull, error) {
	const q = `
SELECT id, member_id, ref_id, perintah, kode_produk, tujuan, qty, status,
       COALESCE(qty_provider, qty), COALESCE(keterangan, ''), COALESCE(charge_receiver_applied, false), biaya_perkiraan, COALESCE(fee_member_rp, 0), harga_member
FROM public.transaksi_member
WHERE id=$1
LIMIT 1`
	return scanCallbackTrxMember(r.db.QueryRowContext(ctx, q, trxID))
}

func (r *ProviderCallbackRepository) GetLatestTransaksiMemberByRefID(ctx context.Context, refID string) (*CallbackTrxMemberFull, error) {
	const q = `
SELECT id, member_id, ref_id, perintah, kode_produk, tujuan, qty, status,
       COALESCE(qty_provider, qty), COALESCE(keterangan, ''), COALESCE(charge_receiver_applied, false), biaya_perkiraan, COALESCE(fee_member_rp, 0), harga_member
FROM public.transaksi_member
WHERE ref_id = $1
ORDER BY id DESC
LIMIT 1`
	return scanCallbackTrxMember(r.db.QueryRowContext(ctx, q, strings.TrimSpace(refID)))
}

func (r *ProviderCallbackRepository) UpdateTransaksiMemberStatus(ctx context.Context, trxID int64, status string, keterangan string, biayaAktual int64) error {
	return updateTrxMemberStatusWithAudit(ctx, r.db, trxID, status, keterangan, biayaAktual, "provider_callback_update", nil)
}

func (r *ProviderCallbackRepository) ForceReopenFailedToPending(ctx context.Context, trxID int64, keterangan string) error {
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
		Aksi:               "provider_callback_reopen_pending",
		DiubahOleh:         nil,
	}); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *ProviderCallbackRepository) UpdateTransaksiMemberSettle(ctx context.Context, trxID int64, status, keterangan string, biayaAktual, hargaJavapay, hargaMember int64) error {
	err := updateTrxMemberSettleWithAudit(ctx, r.db, trxID, status, keterangan, biayaAktual, hargaJavapay, hargaMember, "provider_callback_settle")
	if err != nil {
		return err
	}
	// Sync provider rows saat member final
	if status == "success" || status == "failed" {
		var refID string
		_ = r.db.QueryRowContext(ctx, `SELECT ref_id FROM public.transaksi_member WHERE id = $1`, trxID).Scan(&refID)
		if refID != "" {
			if status == "success" {
				// Saat member success, pastikan provider row terakhir juga success.
				// Hanya update status — field lain (pesan, respon_mentah, harga, kode_respon)
				// tetap dari callback asli provider.
				_, _ = r.db.ExecContext(ctx, `
UPDATE public.transaksi_provider
SET status = 'success'
WHERE id = (
    SELECT id FROM public.transaksi_provider
    WHERE ref_id = $1
    ORDER BY id DESC LIMIT 1
) AND status != 'success'`, refID)
			}
			// Cleanup orphan pending provider rows. Keep the row closed, but
			// make it explicit that this is not a provider failure response.
			_, _ = r.db.ExecContext(ctx, `
	UPDATE public.transaksi_provider
	SET status = 'failed',
	    kode_respon = COALESCE(NULLIF(TRIM(kode_respon),''),'91'),
	    pesan = COALESCE(NULLIF(TRIM(pesan),''),
	      CASE
	        WHEN $2 = 'success' THEN 'cleanup: member sudah success via provider lain; pending provider ditutup, bukan response gagal provider'
	        ELSE 'cleanup: member sudah failed; pending provider ditutup'
	      END)
	WHERE ref_id = $1 AND status = 'pending'`, refID, status)
		}
	}
	return nil
}

func (r *ProviderCallbackRepository) CreditDompet(ctx context.Context, memberID int64, refID string, jumlah int64, alasan, catatan string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var before int64
	err = tx.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_member WHERE member_id=$1 FOR UPDATE`, memberID).Scan(&before)
	if err != nil {
		return err
	}

	if strings.EqualFold(strings.TrimSpace(alasan), "REFUND") {
		outstanding, err := memberWalletOutstandingHoldByRef(ctx, tx, memberID, refID)
		if err != nil {
			return err
		}
		jumlah = clampWalletRefundAmount(jumlah, outstanding)
		if jumlah <= 0 {
			return tx.Commit()
		}
	} else {
		var existingCount int64
		if err := tx.QueryRowContext(ctx, `
SELECT COUNT(1)
FROM public.mutasi_dompet
WHERE member_id = $1
  AND ref_id = $2
  AND LOWER(COALESCE(arah, '')) = 'credit'
  AND LOWER(COALESCE(alasan, '')) = LOWER($3)
`, memberID, refID, alasan).Scan(&existingCount); err != nil {
			return err
		}
		if existingCount > 0 {
			return tx.Commit()
		}
	}

	after := before + jumlah

	_, err = tx.ExecContext(ctx, `UPDATE public.dompet_member SET saldo=$2, diperbarui_pada=now() WHERE member_id=$1`, memberID, after)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah)
VALUES
  ($1,$2,'CREDIT',$3,$4,NULLIF($5,''),$6,$7)
`, memberID, refID, jumlah, alasan, catatan, before, after)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *ProviderCallbackRepository) RecoverMemberRefund(ctx context.Context, memberID int64, refID string, jumlah int64, alasan, catatan string) error {
	if jumlah <= 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var before int64
	if err := tx.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_member WHERE member_id=$1 FOR UPDATE`, memberID).Scan(&before); err != nil {
		return err
	}

	var refunded int64
	if err := tx.QueryRowContext(ctx, `
SELECT COALESCE(SUM(jumlah), 0)
FROM public.mutasi_dompet
WHERE member_id = $1
  AND ref_id = $2
  AND LOWER(COALESCE(arah, '')) = 'credit'
  AND LOWER(COALESCE(alasan, '')) = 'refund'
`, memberID, refID).Scan(&refunded); err != nil {
		return err
	}

	var recovered int64
	if err := tx.QueryRowContext(ctx, `
SELECT COALESCE(SUM(jumlah), 0)
FROM public.mutasi_dompet
WHERE member_id = $1
  AND ref_id = $2
  AND LOWER(COALESCE(arah, '')) = 'debit'
  AND LOWER(COALESCE(alasan, '')) = LOWER($3)
`, memberID, refID, alasan).Scan(&recovered); err != nil {
		return err
	}

	outstanding := refunded - recovered
	if outstanding > jumlah {
		outstanding = jumlah
	}
	if outstanding <= 0 {
		return tx.Commit()
	}
	if before < outstanding {
		return ErrInsufficientBalance
	}

	after := before - outstanding
	if _, err := tx.ExecContext(ctx, `UPDATE public.dompet_member SET saldo=$2, diperbarui_pada=now() WHERE member_id=$1`, memberID, after); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah)
VALUES
  ($1,$2,'DEBIT',$3,$4,NULLIF($5,''),$6,$7)
`, memberID, refID, outstanding, alasan, catatan, before, after); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *ProviderCallbackRepository) GetMemberWebhookURL(ctx context.Context, memberID int64) (string, error) {
	if memberID <= 0 {
		return "", errors.New("invalid member")
	}
	var wh sql.NullString
	err := r.db.QueryRowContext(ctx, `
SELECT webhook_url
FROM public.member_ip_whitelist
WHERE member_id=$1 AND aktif=true AND webhook_url IS NOT NULL AND webhook_url <> ''
ORDER BY dibuat_pada DESC, id DESC
LIMIT 1
`, memberID).Scan(&wh)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	if !wh.Valid {
		return "", nil
	}
	return strings.TrimSpace(wh.String), nil
}

func (r *ProviderCallbackRepository) GetSaldo(ctx context.Context, memberID int64) (int64, error) {
	var saldo int64
	err := r.db.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_member
WHERE member_id=$1
`, memberID).Scan(&saldo)
	return saldo, err
}

func (r *ProviderCallbackRepository) DowngradeTransaksiMember(ctx context.Context, trxID int64, keterangan string) error {
	_, _ = r.db.ExecContext(ctx, `
INSERT INTO public.transaksi_member_status_log
  (transaksi_member_id, status_sebelum, status_sesudah, keterangan_sebelum, keterangan_sesudah, biaya_aktual_sebelum, biaya_aktual_sesudah, aksi)
SELECT id, status, 'failed', COALESCE(keterangan,''), $2, COALESCE(biaya_aktual,0), 0, 'provider_callback_downgrade'
FROM public.transaksi_member WHERE id = $1
`, trxID, keterangan)

	_, err := r.db.ExecContext(ctx, `
UPDATE public.transaksi_member
SET status = 'failed', keterangan = $2, biaya_aktual = 0, diperbarui_pada = NOW()
WHERE id = $1
`, trxID, keterangan)

	var refID string
	_ = r.db.QueryRowContext(ctx, `SELECT ref_id FROM public.transaksi_member WHERE id = $1`, trxID).Scan(&refID)
	if refID != "" {
		_, _ = r.db.ExecContext(ctx, `UPDATE public.transaksi_provider SET status = 'failed' WHERE id = (SELECT id FROM public.transaksi_provider WHERE ref_id = $1 ORDER BY id DESC LIMIT 1) AND status = 'success'`, refID)
	}

	return err
}
