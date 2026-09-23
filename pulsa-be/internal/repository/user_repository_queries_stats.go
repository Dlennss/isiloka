package repository

import "context"

func (r *UserRepository) StatsLast3Months(ctx context.Context, userID int64) ([]UserMonthStatsRow, error) {
	rows, err := r.db.QueryContext(ctx, `
WITH months AS (
  SELECT date_trunc('month', now()) - (gs * interval '1 month') AS m
  FROM generate_series(0,2) gs
),
trx AS (
  SELECT date_trunc('month', dibuat_pada) AS m,
         count(*) FILTER (WHERE status = 'success') AS success_count,
         count(*) FILTER (WHERE status = 'failed')  AS failed_count,
         sum(CASE WHEN status = 'success'
                  THEN GREATEST(biaya_aktual, biaya_perkiraan)
                  ELSE 0 END) AS success_amount,
         sum(CASE WHEN status = 'failed'
                  THEN GREATEST(biaya_aktual, biaya_perkiraan)
                  ELSE 0 END) AS failed_amount
  FROM public.transaksi_member
  WHERE member_id = $1
    AND dibuat_pada >= date_trunc('month', now()) - interval '2 month'
    AND dibuat_pada <  date_trunc('month', now()) + interval '1 month'
  GROUP BY 1
),
dep AS (
  SELECT date_trunc('month', dibuat_pada) AS m,
         count(*) FILTER (WHERE status = 'approved') AS approved_count,
         count(*) FILTER (WHERE status = 'rejected') AS rejected_count,
         sum(CASE WHEN status = 'approved' THEN COALESCE(approved_amount, amount) ELSE 0 END) AS approved_amount,
         sum(CASE WHEN status = 'rejected' THEN amount ELSE 0 END) AS rejected_amount
  FROM public.deposit_request
  WHERE member_id = $1
    AND dibuat_pada >= date_trunc('month', now()) - interval '2 month'
    AND dibuat_pada <  date_trunc('month', now()) + interval '1 month'
  GROUP BY 1
),
wallet_adj AS (
  SELECT date_trunc('month', dibuat_pada) AS m,
         sum(
           CASE
             WHEN lower(trim(coalesce(arah, ''))) = 'credit'
              AND lower(trim(coalesce(alasan, ''))) IN (
                'refund',
                'admin_cancel_refund',
                'app_order_refund',
                'app_order_wallet_refund',
                'guest_refund_claim',
                'admin manual credit'
              )
               THEN jumlah
             WHEN lower(trim(coalesce(arah, ''))) = 'debit'
              AND lower(trim(coalesce(alasan, ''))) IN (
                'admin_complete_redebit',
                'admin manual debit'
              )
               THEN -jumlah
             ELSE 0
           END
         ) AS net_amount
  FROM public.mutasi_dompet
  WHERE member_id = $1
    AND dibuat_pada >= date_trunc('month', now()) - interval '2 month'
    AND dibuat_pada <  date_trunc('month', now()) + interval '1 month'
  GROUP BY 1
)
SELECT months.m,
       COALESCE(trx.success_count, 0),
       COALESCE(trx.failed_count, 0),
       COALESCE(trx.success_amount, 0),
       COALESCE(trx.failed_amount, 0),
       COALESCE(dep.approved_count, 0),
       COALESCE(dep.rejected_count, 0),
       COALESCE(dep.approved_amount, 0),
       COALESCE(dep.rejected_amount, 0),
       COALESCE(wallet_adj.net_amount, 0)
FROM months
LEFT JOIN trx ON trx.m = months.m
LEFT JOIN dep ON dep.m = months.m
LEFT JOIN wallet_adj ON wallet_adj.m = months.m
ORDER BY months.m DESC
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]UserMonthStatsRow, 0, 3)
	for rows.Next() {
		var s UserMonthStatsRow
		if err := rows.Scan(
			&s.Month,
			&s.TrxSuccessCount,
			&s.TrxFailedCount,
			&s.TrxSuccessAmount,
			&s.TrxFailedAmount,
			&s.DepApprovedCount,
			&s.DepRejectedCount,
			&s.DepApprovedAmount,
			&s.DepRejectedAmount,
			&s.WalletAdjustNet,
		); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
