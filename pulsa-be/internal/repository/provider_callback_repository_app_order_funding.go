package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func (r *ProviderCallbackRepository) RefundAppOrderFunding(ctx context.Context, memberID int64, refID, note string) error {
	if memberID <= 0 || strings.TrimSpace(refID) == "" {
		return fmt.Errorf("data refund order tidak valid")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var walletDebit, creditDebit int64
	if err := tx.QueryRowContext(ctx, `
SELECT
  COALESCE((ap.raw_request->>'wallet_debit')::bigint,0),
  COALESCE((ap.raw_request->>'credit_debit')::bigint,0)
FROM public.app_order_payment ap
JOIN public.app_order ao ON ao.id=ap.app_order_id
WHERE ap.order_id=$1 AND ao.member_id=$2
FOR UPDATE
`, refID, memberID).Scan(&walletDebit, &creditDebit); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("data pembayaran order tidak ditemukan")
		}
		return err
	}

	const refundReason = "APP_ORDER_BALANCE_REFUND"
	var walletRefunded int64
	if err := tx.QueryRowContext(ctx, `
SELECT COALESCE(SUM(jumlah),0)
FROM public.mutasi_dompet
WHERE member_id=$1 AND ref_id=$2 AND arah='CREDIT' AND alasan=$3
`, memberID, refID, refundReason).Scan(&walletRefunded); err != nil {
		return err
	}
	walletRefund := walletDebit - walletRefunded
	if walletRefund > 0 {
		var before int64
		if err := tx.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_member WHERE member_id=$1 FOR UPDATE`, memberID).Scan(&before); err != nil {
			return err
		}
		after := before + walletRefund
		if _, err := tx.ExecContext(ctx, `UPDATE public.dompet_member SET saldo=$2,diperbarui_pada=now() WHERE member_id=$1`, memberID, after); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id,ref_id,arah,jumlah,alasan,catatan,saldo_sebelum,saldo_sesudah,dibuat_pada)
VALUES ($1,$2,'CREDIT',$3,$4,NULLIF($5,''),$6,$7,now())
`, memberID, refID, walletRefund, refundReason, note, before, after); err != nil {
			return err
		}
	}
	if creditDebit > 0 {
		var refunded int64
		if err := tx.QueryRowContext(ctx, `
SELECT public.fn_agent_credit_refund_by_ref($1,$2,$3,$4,$5)
`, memberID, refID, creditDebit, refundReason, note).Scan(&refunded); err != nil {
			return err
		}
	}
	return tx.Commit()
}
