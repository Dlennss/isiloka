package repository

import "context"

func clampWalletRefundAmount(requested int64, outstanding int64) int64 {
	if requested <= 0 || outstanding <= 0 {
		return 0
	}
	if requested > outstanding {
		return outstanding
	}
	return requested
}

func ComputeWalletHoldDebitAmount(target int64, outstanding int64) int64 {
	if target <= 0 {
		return 0
	}
	if outstanding < 0 {
		outstanding = 0
	}
	if outstanding >= target {
		return 0
	}
	return target - outstanding
}

func memberWalletOutstandingHoldByRef(ctx context.Context, exec dbtx, memberID int64, refID string) (int64, error) {
	var outstanding int64
	err := exec.QueryRowContext(ctx, `
SELECT
  COALESCE(SUM(
    CASE
      WHEN LOWER(COALESCE(arah, '')) = 'debit'
       AND LOWER(COALESCE(alasan, '')) = 'trx_hold'
      THEN jumlah
      WHEN LOWER(COALESCE(arah, '')) = 'credit'
       AND LOWER(COALESCE(alasan, '')) = 'refund'
      THEN -jumlah
      ELSE 0
    END
  ), 0)
FROM public.mutasi_dompet
WHERE member_id = $1
  AND ref_id = $2
`, memberID, refID).Scan(&outstanding)
	return outstanding, err
}
