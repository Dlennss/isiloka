package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

func (r *HistoryRepository) GetAdminTrxCallbackTarget(ctx context.Context, trxID int64) (*AdminTrxCallbackTarget, error) {
	if trxID <= 0 {
		return nil, errors.New("trx_id invalid")
	}
	var row AdminTrxCallbackTarget
	err := r.db.QueryRowContext(ctx, `
SELECT id, member_id, ref_id, perintah, kode_produk, tujuan, qty,
       COALESCE(qty_provider, 0),
       COALESCE(charge_receiver_applied, false),
       COALESCE(fee_member_rp, 0),
       status,
       COALESCE(keterangan, ''),
       COALESCE(biaya_perkiraan, 0),
       COALESCE(biaya_aktual, 0),
       COALESCE(harga_member, 0)
FROM public.transaksi_member
WHERE id = $1
`, trxID).Scan(
		&row.ID,
		&row.MemberID,
		&row.RefID,
		&row.Perintah,
		&row.KodeProduk,
		&row.Tujuan,
		&row.Qty,
		&row.QtyProvider,
		&row.ChargeReceiverApplied,
		&row.FeeMemberRp,
		&row.Status,
		&row.Keterangan,
		&row.BiayaPerkiraan,
		&row.BiayaAktual,
		&row.HargaMember,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("transaksi tidak ditemukan")
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *HistoryRepository) GetMemberWebhookURL(ctx context.Context, memberID int64) (string, error) {
	if memberID <= 0 {
		return "", errors.New("member_id invalid")
	}
	var webhook sql.NullString
	err := r.db.QueryRowContext(ctx, `
SELECT webhook_url
FROM public.member_ip_whitelist
WHERE member_id = $1
  AND aktif = true
  AND webhook_url IS NOT NULL
  AND TRIM(webhook_url) <> ''
ORDER BY id DESC
LIMIT 1
`, memberID).Scan(&webhook)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(webhook.String), nil
}
