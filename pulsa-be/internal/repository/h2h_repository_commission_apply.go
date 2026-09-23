package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"pulsa2/internal/helper"
)

func (r *H2HRepository) ApplyCommissionForSuccessTrx(ctx context.Context, trxID int64) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var (
		memberID    int64
		refID       string
		kodeProduk  string
		status      string
		buyerRole   string
		h2hAgentID  sql.NullInt64
		h2hMasterID sql.NullInt64
	)
	if err := tx.QueryRowContext(ctx, `
SELECT
  tm.member_id, tm.ref_id, tm.kode_produk, COALESCE(tm.status, ''), COALESCE(m.role, ''),
  m.h2h_agent_member_id, m.h2h_master_member_id
FROM public.transaksi_member tm
JOIN public.member m ON m.id = tm.member_id
WHERE tm.id = $1
LIMIT 1
`, trxID).Scan(&memberID, &refID, &kodeProduk, &status, &buyerRole, &h2hAgentID, &h2hMasterID); err != nil {
		return err
	}

	if strings.TrimSpace(strings.ToLower(status)) != "success" {
		return tx.Commit()
	}
	if p24SkipH2HForSMPAYSourceEnabled() {
		skip, err := lookupSMPAYSkipH2HCommission(ctx, tx, trxID, memberID, refID)
		if err != nil {
			return err
		}
		if skip {
			return tx.Commit()
		}
	}

	categoryName := helper.ClassifyH2HFeeCategory(kodeProduk)

	type commissionTarget struct {
		memberID int64
		level    string
		amount   int64
	}
	targets := make([]commissionTarget, 0, 2)

	getCommission := func(targetID int64, level string) (int64, error) {
		var agentCommission, masterCommission int64
		err := tx.QueryRowContext(ctx, `
SELECT COALESCE(h2h_agent_commission_rp, 0), COALESCE(h2h_master_commission_rp, 0)
FROM public.member
WHERE id = $1
LIMIT 1
`, targetID).Scan(&agentCommission, &masterCommission)
		if err != nil {
			return 0, err
		}
		switch level {
		case helper.RoleH2HAgent:
			return agentCommission, nil
		case helper.RoleH2HMaster:
			return masterCommission, nil
		default:
			return 0, nil
		}
	}

	switch strings.TrimSpace(strings.ToLower(buyerRole)) {
	case helper.RoleMember:
		if h2hAgentID.Valid {
			agentFee, feeErr := getCommission(h2hAgentID.Int64, helper.RoleH2HAgent)
			if feeErr != nil {
				return feeErr
			}
			if agentFee > 0 {
				targets = append(targets, commissionTarget{memberID: h2hAgentID.Int64, level: "agent_member", amount: agentFee})
			}
			if h2hMasterID.Valid {
				masterFee, feeErr := getCommission(h2hMasterID.Int64, helper.RoleH2HMaster)
				if feeErr != nil {
					return feeErr
				}
				if masterFee > 0 {
					targets = append(targets, commissionTarget{memberID: h2hMasterID.Int64, level: "master_member", amount: masterFee})
				}
			}
		} else if h2hMasterID.Valid {
			masterFee, feeErr := getCommission(h2hMasterID.Int64, helper.RoleH2HMaster)
			if feeErr != nil {
				return feeErr
			}
			if masterFee > 0 {
				targets = append(targets, commissionTarget{memberID: h2hMasterID.Int64, level: "master_member", amount: masterFee})
			}
		}
	case helper.RoleH2HAgent:
		if h2hMasterID.Valid {
			masterFee, feeErr := getCommission(h2hMasterID.Int64, helper.RoleH2HMaster)
			if feeErr != nil {
				return feeErr
			}
			if masterFee > 0 {
				targets = append(targets, commissionTarget{memberID: h2hMasterID.Int64, level: "master_member", amount: masterFee})
			}
		}
	}

	for _, target := range targets {
		if target.memberID <= 0 || target.amount <= 0 {
			continue
		}
		res, err := tx.ExecContext(ctx, `
INSERT INTO public.h2h_commission_ledger
  (member_id, source_member_id, source_trx_member_id, ref_id, level_name, kategori_name, amount, note, created_at)
VALUES
  ($1,$2,$3,$4,$5,$6,$7,$8,now())
ON CONFLICT (member_id, source_trx_member_id) DO NOTHING
`, target.memberID, memberID, trxID, refID, target.level, categoryName, target.amount, "h2h commission auto credit")
		if err != nil {
			return err
		}
		aff, err := res.RowsAffected()
		if err != nil || aff == 0 {
			continue
		}

		var before int64
		if err := tx.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_member WHERE member_id = $1 FOR UPDATE`, target.memberID).Scan(&before); err != nil {
			return err
		}
		after := before + target.amount
		if _, err := tx.ExecContext(ctx, `UPDATE public.dompet_member SET saldo = $2, diperbarui_pada = now() WHERE member_id = $1`, target.memberID, after); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
VALUES
  ($1,$2,'CREDIT',$3,'H2H_COMMISSION',NULLIF($4,''),$5,$6,now())
`, target.memberID, refID, target.amount, "komisi h2h "+target.level, before, after); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func p24SkipH2HForSMPAYSourceEnabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("P24_SKIP_H2H_FOR_SMPAY_SOURCE_ENABLED")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func (r *H2HRepository) ApplyHistoricalCommission(ctx context.Context, sourceMemberID int64, buyerRole string, h2hAgentID, h2hMasterID *int64) (agentCount int64, agentTotal int64, masterCount int64, masterTotal int64, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer func() { _ = tx.Rollback() }()

	applyTarget := func(targetID int64, level string, amount int64) (int64, int64, error) {
		if targetID <= 0 || amount <= 0 {
			return 0, 0, nil
		}

		note := "h2h commission historical backfill"
		var count, total int64
		if err := tx.QueryRowContext(ctx, `
WITH inserted AS (
  INSERT INTO public.h2h_commission_ledger
    (member_id, source_member_id, source_trx_member_id, ref_id, level_name, kategori_name, amount, note, created_at)
  SELECT
    $1, tm.member_id, tm.id, tm.ref_id, $2, helper_class, $3, $4, now()
  FROM (
    SELECT
      tm.id, tm.member_id, tm.ref_id,
      CASE
        WHEN upper(COALESCE(tm.kode_produk, '')) = 'DANA' THEN 'DANA'
        WHEN upper(COALESCE(tm.kode_produk, '')) IN ('GOPAY', 'GOJEK', 'GPAY') THEN 'GOPAY'
        WHEN upper(COALESCE(tm.kode_produk, '')) = 'OVO' THEN 'OVO'
        WHEN upper(COALESCE(tm.kode_produk, '')) IN ('LINKAJA', 'LAJA') THEN 'LINKAJA'
        WHEN upper(COALESCE(tm.kode_produk, '')) IN ('SHOPEE', 'SHOPEEPAY', 'SHPAY') THEN 'SHOPEEPAY'
        ELSE 'LAINNYA'
      END AS helper_class
    FROM public.transaksi_member tm
    WHERE tm.member_id = $5
      AND lower(COALESCE(tm.status, '')) = 'success'
  ) tm
  ON CONFLICT (member_id, source_trx_member_id) DO NOTHING
  RETURNING amount
)
SELECT COUNT(*), COALESCE(SUM(amount), 0)
FROM inserted
`, targetID, level, amount, note, sourceMemberID).Scan(&count, &total); err != nil {
			return 0, 0, err
		}
		if total <= 0 {
			return count, total, nil
		}

		var before int64
		if err := tx.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_member WHERE member_id = $1 FOR UPDATE`, targetID).Scan(&before); err != nil {
			return 0, 0, err
		}
		after := before + total
		if _, err := tx.ExecContext(ctx, `UPDATE public.dompet_member SET saldo = $2, diperbarui_pada = now() WHERE member_id = $1`, targetID, after); err != nil {
			return 0, 0, err
		}

		refID := fmt.Sprintf("HCB-%s-%s", time.Now().Format("20060102150405"), strings.ToUpper(helper.RandHex(4)))
		if _, err := tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
VALUES
  ($1,$2,'CREDIT',$3,'H2H_COMMISSION_BACKFILL',NULLIF($4,''),$5,$6,now())
`, targetID, refID, total, fmt.Sprintf("backfill komisi h2h %s dari member %d", level, sourceMemberID), before, after); err != nil {
			return 0, 0, err
		}
		return count, total, nil
	}

	getTargetCommission := func(targetID int64, level string) (int64, error) {
		var agentCommission, masterCommission int64
		if err := tx.QueryRowContext(ctx, `
SELECT COALESCE(h2h_agent_commission_rp, 0), COALESCE(h2h_master_commission_rp, 0)
FROM public.member
WHERE id = $1
LIMIT 1
`, targetID).Scan(&agentCommission, &masterCommission); err != nil {
			return 0, err
		}
		if level == "agent_member" {
			return agentCommission, nil
		}
		return masterCommission, nil
	}

	switch strings.TrimSpace(strings.ToLower(buyerRole)) {
	case helper.RoleMember:
		if h2hAgentID != nil && *h2hAgentID > 0 {
			amount, feeErr := getTargetCommission(*h2hAgentID, "agent_member")
			if feeErr != nil {
				return 0, 0, 0, 0, feeErr
			}
			agentCount, agentTotal, err = applyTarget(*h2hAgentID, "agent_member", amount)
			if err != nil {
				return 0, 0, 0, 0, err
			}
		}
		if h2hMasterID != nil && *h2hMasterID > 0 {
			amount, feeErr := getTargetCommission(*h2hMasterID, "master_member")
			if feeErr != nil {
				return 0, 0, 0, 0, feeErr
			}
			masterCount, masterTotal, err = applyTarget(*h2hMasterID, "master_member", amount)
			if err != nil {
				return 0, 0, 0, 0, err
			}
		}
	case helper.RoleH2HAgent:
		if h2hMasterID != nil && *h2hMasterID > 0 {
			amount, feeErr := getTargetCommission(*h2hMasterID, "master_member")
			if feeErr != nil {
				return 0, 0, 0, 0, feeErr
			}
			masterCount, masterTotal, err = applyTarget(*h2hMasterID, "master_member", amount)
			if err != nil {
				return 0, 0, 0, 0, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, 0, 0, err
	}
	return agentCount, agentTotal, masterCount, masterTotal, nil
}
