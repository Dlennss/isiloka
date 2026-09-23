package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (r *ProviderCallbackRepository) ApplyProviderWalletTx(ctx context.Context, in CallbackProviderWalletTxIn) (before int64, after int64, err error) {
	p, err := r.normalizeProviderForCallback(ctx, in.Provider)
	if err != nil {
		return 0, 0, err
	}
	if strings.TrimSpace(in.RefID) == "" {
		return 0, 0, fmt.Errorf("ref_id required")
	}
	in.Arah = strings.TrimSpace(strings.ToLower(in.Arah))
	if in.Arah != "credit" && in.Arah != "debit" {
		return 0, 0, fmt.Errorf("arah must be credit|debit")
	}
	if in.Jumlah <= 0 {
		return 0, 0, fmt.Errorf("jumlah must be > 0")
	}
	if strings.TrimSpace(in.Alasan) == "" {
		return 0, 0, fmt.Errorf("alasan required")
	}

	if _, err := r.db.ExecContext(ctx, `
INSERT INTO public.dompet_provider (provider, saldo)
VALUES ($1, 0)
ON CONFLICT (provider) DO NOTHING
`, p); err != nil {
		return 0, 0, err
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = tx.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_provider WHERE provider=$1 FOR UPDATE`, p).Scan(&before); err != nil {
		return 0, 0, err
	}

	var (
		existingBefore sql.NullInt64
		existingAfter  sql.NullInt64
	)
	var tm any
	var tp any
	var aop any
	if in.TransaksiMemberID != nil {
		tm = *in.TransaksiMemberID
	}
	if in.TransaksiProviderID != nil {
		tp = *in.TransaksiProviderID
	}
	if in.AppOrderProviderTrxID != nil {
		aop = *in.AppOrderProviderTrxID
	}
	err = tx.QueryRowContext(ctx, `
SELECT saldo_sebelum, saldo_sesudah
FROM public.mutasi_dompet_provider
WHERE provider = $1
  AND ref_id = $2
  AND arah = $3
  AND alasan = $4
  AND COALESCE(transaksi_member_id, 0) = COALESCE($5::bigint, 0)
  AND COALESCE(transaksi_provider_id, 0) = COALESCE($6::bigint, 0)
  AND COALESCE(app_order_provider_trx_id, 0) = COALESCE($7::bigint, 0)
ORDER BY id DESC
LIMIT 1
`, p, strings.TrimSpace(in.RefID), in.Arah, in.Alasan, tm, tp, aop).Scan(&existingBefore, &existingAfter)
	if err == nil {
		if existingBefore.Valid {
			before = existingBefore.Int64
		}
		if existingAfter.Valid {
			after = existingAfter.Int64
		} else {
			after = before
		}
		err = tx.Commit()
		return before, after, err
	}
	if err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}

	after = before
	if in.Arah == "credit" {
		after = before + in.Jumlah
	} else {
		if before < in.Jumlah && !in.AllowNegative {
			return 0, 0, errors.New("provider saldo tidak cukup")
		}
		after = before - in.Jumlah
	}

	if _, err = tx.ExecContext(ctx, `UPDATE public.dompet_provider SET saldo=$1, diperbarui_pada=now() WHERE provider=$2`, after, p); err != nil {
		return 0, 0, err
	}

	var meta any
	if len(in.MetaJSON) > 0 {
		meta = string(in.MetaJSON)
	}

	if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet_provider
  (provider, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah,
   transaksi_member_id, transaksi_provider_id, app_order_provider_trx_id, dibuat_pada, meta)
VALUES
  ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,COALESCE($13,'{}'::jsonb))
`, p, in.RefID, in.Arah, in.Jumlah, in.Alasan, in.Catatan, before, after, tm, tp, aop, time.Now().UTC(), meta); err != nil {
		return 0, 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, 0, err
	}
	return before, after, nil
}
