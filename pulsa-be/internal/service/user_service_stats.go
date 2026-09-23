package service

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

func (s *UserService) StatsLast3Months(ctx context.Context, userID int64) ([]map[string]any, error) {
	if userID <= 0 {
		return nil, errors.New("user_id invalid")
	}
	items, err := s.repo.StatsLast3Months(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(items))
	for _, it := range items {
		out = append(out, map[string]any{
			"month":               it.Month.In(time.Local).Format("2006-01"),
			"trx_success_count":   it.TrxSuccessCount,
			"trx_failed_count":    it.TrxFailedCount,
			"trx_success_amount":  it.TrxSuccessAmount,
			"trx_failed_amount":   it.TrxFailedAmount,
			"dep_approved_count":  it.DepApprovedCount,
			"dep_rejected_count":  it.DepRejectedCount,
			"dep_approved_amount": it.DepApprovedAmount,
			"dep_rejected_amount": it.DepRejectedAmount,
			"wallet_adjust_net":   it.WalletAdjustNet,
		})
	}
	return out, nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
