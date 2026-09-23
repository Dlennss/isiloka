package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

func (r *MemberSelfRepository) GetMemberProfile(ctx context.Context, memberID int64) (MemberProfile, error) {
	if memberID <= 0 {
		return MemberProfile{}, errors.New("invalid member_id")
	}
	const q = `
SELECT
  m.id,
  m.email,
  COALESCE(m.nama,''),
  COALESCE(m.phone,''),
  COALESCE(m.store_name,''),
  COALESCE(m.profile_photo_url,''),
  COALESCE(m.role,''),
  m.aktif,
  COALESCE(m.charge_receiver, false),
  COALESCE(d.saldo, 0) AS saldo,
  m.dibuat_pada
FROM public.member m
LEFT JOIN public.dompet_member d ON d.member_id = m.id
WHERE m.id = $1
LIMIT 1`
	var p MemberProfile
	var dibuat time.Time
	err := r.db.QueryRowContext(ctx, q, memberID).Scan(
		&p.ID,
		&p.Email,
		&p.Nama,
		&p.Phone,
		&p.StoreName,
		&p.ProfilePhoto,
		&p.Role,
		&p.Aktif,
		&p.ChargeReceiver,
		&p.Saldo,
		&dibuat,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return MemberProfile{}, errors.New("member not found")
		}
		return MemberProfile{}, err
	}
	p.DibuatPada = dibuat.Format(time.RFC3339)
	return p, nil
}

func (r *MemberSelfRepository) UpdateMemberProfile(ctx context.Context, memberID int64, nama, phone, profilePhotoURL string) (MemberProfile, error) {
	if memberID <= 0 {
		return MemberProfile{}, errors.New("invalid member_id")
	}
	_, err := r.db.ExecContext(ctx, `
UPDATE public.member
SET
  nama = NULLIF($2, ''),
  phone = CASE WHEN $3 <> '' THEN $3 ELSE phone END,
  profile_photo_url = CASE WHEN $4 <> '' THEN $4 ELSE profile_photo_url END,
  diubah_pada = now()
WHERE id = $1
`, memberID, nama, phone, profilePhotoURL)
	if err != nil {
		return MemberProfile{}, err
	}
	return r.GetMemberProfile(ctx, memberID)
}

func (r *MemberSelfRepository) UpdateChargeReceiver(ctx context.Context, memberID int64, chargeReceiver bool) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE public.member
SET charge_receiver = $2
WHERE id = $1
`, memberID, chargeReceiver)
	return err
}

func (r *MemberSelfRepository) ListAPIKeys(ctx context.Context, memberID int64) ([]MemberAPIKey, error) {
	if memberID <= 0 {
		return nil, errors.New("invalid member_id")
	}
	const q = `
SELECT id, member_id, api_key, aktif, dibuat_pada
FROM public.member_api_key
WHERE member_id = $1
ORDER BY dibuat_pada DESC`
	rows, err := r.db.QueryContext(ctx, q, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]MemberAPIKey, 0, 8)
	for rows.Next() {
		var k MemberAPIKey
		var dibuat time.Time
		if err := rows.Scan(&k.ID, &k.MemberID, &k.ApiKey, &k.Aktif, &dibuat); err != nil {
			return nil, err
		}
		k.DibuatPada = dibuat.Format(time.RFC3339)
		out = append(out, k)
	}
	return out, rows.Err()
}

func (r *MemberSelfRepository) GetMemberStats3Months(ctx context.Context, memberID int64, now time.Time) ([]MemberMonthStat, error) {
	if memberID <= 0 {
		return nil, errors.New("invalid member_id")
	}
	monthStart := func(t time.Time) time.Time {
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	}
	addMonths := func(t time.Time, m int) time.Time {
		return t.AddDate(0, m, 0)
	}

	loc := now.Location()
	m0 := monthStart(now.In(loc))
	m1 := monthStart(addMonths(m0, -1))
	m2 := monthStart(addMonths(m0, -2))
	months := []time.Time{m2, m1, m0}

	out := make([]MemberMonthStat, 0, 3)
	for _, ms := range months {
		me := monthStart(addMonths(ms, 1))

		var depCount, ledgerMonth int64
		if err := r.db.QueryRowContext(ctx, `
SELECT
  COALESCE(COUNT(*),0),
  COALESCE(SUM(
    CASE
      WHEN UPPER(COALESCE(arah,'')) = 'CREDIT' THEN jumlah
      ELSE -jumlah
    END
  ),0)
FROM public.mutasi_dompet
WHERE member_id = $1
  AND dibuat_pada >= $2 AND dibuat_pada < $3
`, memberID, ms, me).Scan(&depCount, &ledgerMonth); err != nil {
			return nil, err
		}

		var successCount, successSum int64
		if err := r.db.QueryRowContext(ctx, `
SELECT COALESCE(COUNT(*),0), COALESCE(SUM(GREATEST(COALESCE(biaya_aktual,0), COALESCE(biaya_perkiraan,0))),0)
FROM public.transaksi_member
WHERE member_id = $1
  AND dibuat_pada >= $2 AND dibuat_pada < $3
  AND status IN ('success','done','paid')
`, memberID, ms, me).Scan(&successCount, &successSum); err != nil {
			return nil, err
		}

		var failedCount, failedSum int64
		if err := r.db.QueryRowContext(ctx, `
SELECT COALESCE(COUNT(*),0), COALESCE(SUM(COALESCE(biaya_perkiraan,0)),0)
FROM public.transaksi_member
WHERE member_id = $1
  AND dibuat_pada >= $2 AND dibuat_pada < $3
  AND status IN ('failed','error','cancelled')
`, memberID, ms, me).Scan(&failedCount, &failedSum); err != nil {
			return nil, err
		}

		out = append(out, MemberMonthStat{
			Year:         ms.Year(),
			Month:        int(ms.Month()),
			DepositCount: depCount,
			DepositSum:   ledgerMonth + successSum,
			TrxCount:     successCount,
			TrxSum:       successSum,
			SuccessCount: successCount,
			SuccessSum:   successSum,
			FailedCount:  failedCount,
			FailedSum:    failedSum,
		})
	}

	return out, nil
}

func (r *MemberSelfRepository) GetMemberOverallStats(ctx context.Context, memberID int64) (MemberOverallStat, error) {
	if memberID <= 0 {
		return MemberOverallStat{}, errors.New("invalid member_id")
	}

	var out MemberOverallStat
	if err := r.db.QueryRowContext(ctx, `
SELECT
  COALESCE(COUNT(*) FILTER (WHERE status IN ('success','done','paid')),0),
  COALESCE(SUM(GREATEST(COALESCE(biaya_aktual,0), COALESCE(biaya_perkiraan,0))) FILTER (WHERE status IN ('success','done','paid')),0),
  COALESCE(COUNT(*) FILTER (WHERE status IN ('failed','error','cancelled')),0),
  COALESCE(SUM(COALESCE(biaya_perkiraan,0)) FILTER (WHERE status IN ('failed','error','cancelled')),0)
FROM public.transaksi_member
WHERE member_id = $1
`, memberID).Scan(&out.SuccessCount, &out.SuccessSum, &out.FailedCount, &out.FailedSum); err != nil {
		return MemberOverallStat{}, err
	}

	if err := r.db.QueryRowContext(ctx, `
SELECT COALESCE(COUNT(*),0)
FROM public.mutasi_dompet
WHERE member_id = $1
  AND UPPER(COALESCE(arah,'')) = 'CREDIT'
`, memberID).Scan(&out.DepositCount); err != nil {
		return MemberOverallStat{}, err
	}

	var saldoNow int64
	if err := r.db.QueryRowContext(ctx, `
SELECT COALESCE(saldo,0)
FROM public.dompet_member
WHERE member_id = $1
`, memberID).Scan(&saldoNow); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return MemberOverallStat{}, err
		}
	}

	out.DepositSum = saldoNow + out.SuccessSum
	out.LedgerBalance = saldoNow
	out.SaldoReconciled = true
	out.OtherMutationBreakdown = make([]MemberOtherMutationStat, 0)
	return out, nil
}
