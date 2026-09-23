package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"pulsa2/db"
)

func (r *MemberTrxProviderWalletRepository) applyTx(ctx context.Context, in db.ProviderWalletTxIn) (before int64, after int64, err error) {
	p, err := resolveProviderName(ctx, r.db, in.Provider)
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

	if err := r.Ensure(ctx, p); err != nil {
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

	const qLock = `SELECT saldo FROM public.dompet_provider WHERE provider=$1 FOR UPDATE`
	if err = tx.QueryRowContext(ctx, qLock, p).Scan(&before); err != nil {
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

	const qUpd = `UPDATE public.dompet_provider SET saldo=$1, diperbarui_pada=now() WHERE provider=$2`
	if _, err = tx.ExecContext(ctx, qUpd, after, p); err != nil {
		return 0, 0, err
	}

	const qIns = `
INSERT INTO public.mutasi_dompet_provider
  (provider, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah,
   transaksi_member_id, transaksi_provider_id, dibuat_pada, meta)
VALUES
  ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,COALESCE($12,'{}'::jsonb))`

	var tm any = nil
	var tp any = nil
	if in.TransaksiMemberID != nil {
		tm = *in.TransaksiMemberID
	}
	if in.TransaksiProviderID != nil {
		tp = *in.TransaksiProviderID
	}

	createdAt := time.Now().UTC()
	var meta any = nil
	if len(in.MetaJSON) > 0 {
		meta = string(in.MetaJSON)
	}

	if _, err = tx.ExecContext(ctx, qIns,
		p, in.RefID, in.Arah, in.Jumlah, in.Alasan, in.Catatan, before, after, tm, tp, createdAt, meta,
	); err != nil {
		return 0, 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, 0, err
	}
	return before, after, nil
}
