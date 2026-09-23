package db

import (
	"context"
	"errors"
	"time"
)

type MonthStat struct {
	Year         int   `json:"year"`
	Month        int   `json:"month"`
	DepositCount int64 `json:"deposit_count"`
	DepositSum   int64 `json:"deposit_sum"`
	TrxCount     int64 `json:"trx_count"`
	TrxSum       int64 `json:"trx_sum"`
}

func monthStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func addMonths(t time.Time, m int) time.Time {
	return t.AddDate(0, m, 0)
}

// Statistik 3 bulan:
// - Deposit: deposit_request.status='approved' sum(amount)
// - Transaksi: transaksi_member.status IN ('success','done','paid') sum(harga_member)
//
// Catatan: jika status sukses transaksi kamu beda, sesuaikan IN (...)
func (r *MemberRepo) GetMemberStats3Months(ctx context.Context, memberID int64, now time.Time) ([]MonthStat, error) {
	if memberID <= 0 {
		return nil, errors.New("invalid member_id")
	}

	loc := now.Location()
	m0 := monthStart(now.In(loc))       // bulan ini
	m1 := monthStart(addMonths(m0, -1)) // bulan lalu
	m2 := monthStart(addMonths(m0, -2)) // dua bulan lalu

	months := []time.Time{m2, m1, m0}
	out := make([]MonthStat, 0, 3)

	for _, ms := range months {
		me := monthStart(addMonths(ms, 1))

		var depCount, depSum int64
		{
			const qDep = `
SELECT
  COALESCE(COUNT(*),0) AS c,
  COALESCE(SUM(amount),0) AS s
FROM deposit_request
WHERE member_id = $1
  AND dibuat_pada >= $2 AND dibuat_pada < $3
  AND status = 'approved'
`
			if err := r.DB.QueryRowContext(ctx, qDep, memberID, ms, me).Scan(&depCount, &depSum); err != nil {
				return nil, err
			}
		}

		var trxCount, trxSum int64
		{
			const qTrx = `
SELECT
  COALESCE(COUNT(*),0) AS c,
  COALESCE(SUM(harga_member),0) AS s
FROM transaksi_member
WHERE member_id = $1
  AND dibuat_pada >= $2 AND dibuat_pada < $3
  AND status IN ('success','done','paid')
`
			if err := r.DB.QueryRowContext(ctx, qTrx, memberID, ms, me).Scan(&trxCount, &trxSum); err != nil {
				return nil, err
			}
		}

		out = append(out, MonthStat{
			Year:         ms.Year(),
			Month:        int(ms.Month()),
			DepositCount: depCount,
			DepositSum:   depSum,
			TrxCount:     trxCount,
			TrxSum:       trxSum,
		})
	}

	return out, nil
}
