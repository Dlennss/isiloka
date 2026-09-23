package repository

import (
	"context"
	"time"
)

// ListPendingOver — ambil transaksi member yang masih pending lebih dari durasi tertentu.
func (r *MemberTrxMemberRepository) ListPendingOver(ctx context.Context, age time.Duration, limit int) ([]*TrxMemberFull, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	cutoff := time.Now().Add(-age)
	rows, err := r.db.QueryContext(ctx, `
SELECT id, member_id, ref_id, perintah, kode_produk, tujuan, qty, COALESCE(qty_provider, qty), status,
       COALESCE(keterangan, ''), COALESCE(charge_receiver_applied, false),
       COALESCE(biaya_perkiraan, 0), COALESCE(biaya_aktual, 0), COALESCE(fee_member_rp, 0),
       COALESCE(harga_javapay, 0), COALESCE(harga_member, 0)
FROM public.transaksi_member
WHERE status = 'pending'
  AND perintah = 'PAY'
  AND dibuat_pada < $1
ORDER BY dibuat_pada ASC
LIMIT $2
`, cutoff, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*TrxMemberFull, 0, limit)
	for rows.Next() {
		var t TrxMemberFull
		if err := rows.Scan(
			&t.ID, &t.MemberID, &t.RefID, &t.Perintah, &t.KodeProduk, &t.Tujuan,
			&t.Qty, &t.QtyProvider, &t.Status, &t.Keterangan, &t.ChargeReceiverApplied,
			&t.BiayaPerkiraan, &t.BiayaAktual, &t.FeeMemberRp, &t.HargaJavapay, &t.HargaMember,
		); err != nil {
			return nil, err
		}
		out = append(out, &t)
	}
	return out, rows.Err()
}

// ListWithPendingLoketBayarProviderOver returns PAY member rows that still have
// a stale Loketbayar provider attempt, even when the member row is already final.
func (r *MemberTrxMemberRepository) ListWithPendingLoketBayarProviderOver(ctx context.Context, age time.Duration, limit int) ([]*TrxMemberFull, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	cutoff := time.Now().Add(-age)
	rows, err := r.db.QueryContext(ctx, `
	WITH pending_loket AS (
	  SELECT DISTINCT ON (tp.transaksi_member_id)
	         tp.transaksi_member_id, tp.dibuat_pada
	  FROM public.transaksi_provider tp
	  WHERE tp.provider = 'loketbayar'
	    AND tp.status = 'pending'
	    AND tp.dibuat_pada < $1
	  ORDER BY tp.transaksi_member_id, tp.dibuat_pada ASC
	)
	SELECT tm.id, tm.member_id, tm.ref_id, tm.perintah, tm.kode_produk, tm.tujuan, tm.qty, COALESCE(tm.qty_provider, tm.qty), tm.status,
	       COALESCE(tm.keterangan, ''), COALESCE(tm.charge_receiver_applied, false),
	       COALESCE(tm.biaya_perkiraan, 0), COALESCE(tm.biaya_aktual, 0), COALESCE(tm.fee_member_rp, 0),
	       COALESCE(tm.harga_javapay, 0), COALESCE(tm.harga_member, 0)
	FROM pending_loket pl
	JOIN public.transaksi_member tm ON tm.id = pl.transaksi_member_id
	WHERE tm.perintah = 'PAY'
	ORDER BY pl.dibuat_pada ASC
	LIMIT $2
	`, cutoff, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*TrxMemberFull, 0, limit)
	for rows.Next() {
		var t TrxMemberFull
		if err := rows.Scan(
			&t.ID, &t.MemberID, &t.RefID, &t.Perintah, &t.KodeProduk, &t.Tujuan,
			&t.Qty, &t.QtyProvider, &t.Status, &t.Keterangan, &t.ChargeReceiverApplied,
			&t.BiayaPerkiraan, &t.BiayaAktual, &t.FeeMemberRp, &t.HargaJavapay, &t.HargaMember,
		); err != nil {
			return nil, err
		}
		out = append(out, &t)
	}
	return out, rows.Err()
}

// ListWithPendingCallbackWaitProviderOver returns PAY member rows with pending
// SMB/LoketBayar attempts that should be retried by the same refid until a
// provider response or callback closes the provider row.
func (r *MemberTrxMemberRepository) ListWithPendingCallbackWaitProviderOver(ctx context.Context, age time.Duration, limit int) ([]*TrxMemberFull, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	cutoff := time.Now().Add(-age)
	rows, err := r.db.QueryContext(ctx, `
	WITH pending_provider AS (
	  SELECT DISTINCT ON (tp.transaksi_member_id)
	         tp.transaksi_member_id, tp.provider, tp.dibuat_pada
	  FROM public.transaksi_provider tp
	  JOIN public.transaksi_member tm ON tm.id = tp.transaksi_member_id
	  WHERE LOWER(TRIM(tp.provider)) IN ('smb', 'loketbayar')
	    AND tp.status = 'pending'
	    AND tp.dibuat_pada < $1
	    AND (
	      LOWER(TRIM(tp.provider)) = 'loketbayar'
	      OR tm.status = 'pending'
	    )
	  ORDER BY tp.transaksi_member_id, tp.dibuat_pada ASC
	)
	SELECT tm.id, tm.member_id, tm.ref_id, tm.perintah, tm.kode_produk, tm.tujuan, tm.qty, COALESCE(tm.qty_provider, tm.qty), tm.status,
	       COALESCE(tm.keterangan, ''), COALESCE(tm.charge_receiver_applied, false),
	       COALESCE(tm.biaya_perkiraan, 0), COALESCE(tm.biaya_aktual, 0), COALESCE(tm.fee_member_rp, 0),
	       COALESCE(tm.harga_javapay, 0), COALESCE(tm.harga_member, 0)
	FROM pending_provider pp
	JOIN public.transaksi_member tm ON tm.id = pp.transaksi_member_id
	WHERE tm.perintah = 'PAY'
	ORDER BY pp.dibuat_pada ASC
	LIMIT $2
	`, cutoff, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*TrxMemberFull, 0, limit)
	for rows.Next() {
		var t TrxMemberFull
		if err := rows.Scan(
			&t.ID, &t.MemberID, &t.RefID, &t.Perintah, &t.KodeProduk, &t.Tujuan,
			&t.Qty, &t.QtyProvider, &t.Status, &t.Keterangan, &t.ChargeReceiverApplied,
			&t.BiayaPerkiraan, &t.BiayaAktual, &t.FeeMemberRp, &t.HargaJavapay, &t.HargaMember,
		); err != nil {
			return nil, err
		}
		out = append(out, &t)
	}
	return out, rows.Err()
}
