package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type ProviderLatestSnapshot struct {
	Saldo               int64
	SnapshotAt          time.Time
	HasAnchor           bool
	AnchorInternalAfter int64
	AnchorAt            time.Time
	Source              string
}

func (r *WalletRepository) GetProviderSaldo(ctx context.Context, provider string) (int64, error) {
	p, err := r.normalizeWalletProvider(ctx, provider)
	if err != nil {
		return 0, err
	}
	if err := r.ensureProviderWallet(ctx, p); err != nil {
		return 0, err
	}
	var saldo int64
	if err := r.db.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_provider WHERE provider = $1 LIMIT 1`, p).Scan(&saldo); err != nil {
		return 0, err
	}
	return saldo, nil
}

func (r *WalletRepository) GetProviderLatestSnapshot(ctx context.Context, provider string) (ProviderLatestSnapshot, bool, error) {
	p, err := r.normalizeWalletProvider(ctx, provider)
	if err != nil {
		return ProviderLatestSnapshot{}, false, err
	}
	var out ProviderLatestSnapshot
	if err := r.db.QueryRowContext(ctx, `
SELECT saldo_provider, dibuat_pada
FROM public.provider_saldo_snapshot
WHERE provider = $1
ORDER BY transaksi_provider_id DESC NULLS LAST, dibuat_pada DESC, id DESC
LIMIT 1
`, p).Scan(&out.Saldo, &out.SnapshotAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ProviderLatestSnapshot{}, false, nil
		}
		return ProviderLatestSnapshot{}, false, err
	}
	out.Source = "raw_snapshot"
	return out, true, nil
}

func (r *WalletRepository) ApplyProviderTx(ctx context.Context, in ProviderWalletTxIn) (before int64, after int64, err error) {
	p, err := r.normalizeWalletProvider(ctx, in.Provider)
	if err != nil {
		return 0, 0, err
	}
	if strings.TrimSpace(in.RefID) == "" {
		return 0, 0, errors.New("ref_id required")
	}
	in.Arah = strings.TrimSpace(strings.ToLower(in.Arah))
	if in.Arah != "credit" && in.Arah != "debit" {
		return 0, 0, errors.New("arah must be credit|debit")
	}
	if in.Jumlah <= 0 {
		return 0, 0, errors.New("jumlah must be > 0")
	}
	if strings.TrimSpace(in.Alasan) == "" {
		return 0, 0, errors.New("alasan required")
	}

	if err := r.ensureProviderWallet(ctx, p); err != nil {
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

	after = before
	if in.Arah == "credit" {
		after = before + in.Jumlah
	} else {
		if before < in.Jumlah {
			return 0, 0, errors.New("provider saldo tidak cukup")
		}
		after = before - in.Jumlah
	}

	if _, err = tx.ExecContext(ctx, `UPDATE public.dompet_provider SET saldo=$1, diperbarui_pada=now() WHERE provider=$2`, after, p); err != nil {
		return 0, 0, err
	}

	var tm, tp, actor any
	if in.TransaksiMemberID != nil {
		tm = *in.TransaksiMemberID
	}
	if in.TransaksiProviderID != nil {
		tp = *in.TransaksiProviderID
	}
	if in.DiubahOleh != nil {
		actor = *in.DiubahOleh
	}
	var meta any
	if len(in.MetaJSON) > 0 {
		meta = string(in.MetaJSON)
	}

	if _, err = tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet_provider
  (provider, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah,
   transaksi_member_id, transaksi_provider_id, diubah_oleh, dibuat_pada, meta)
VALUES
  ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,COALESCE($13,'{}'::jsonb))
`, p, in.RefID, in.Arah, in.Jumlah, in.Alasan, in.Catatan, before, after, tm, tp, actor, time.Now().UTC(), meta); err != nil {
		return 0, 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, 0, err
	}
	return before, after, nil
}
func (r *WalletRepository) ensureProviderWallet(ctx context.Context, provider string) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO public.dompet_provider (provider, saldo)
VALUES ($1, 0)
ON CONFLICT (provider) DO NOTHING
`, provider)
	return err
}

func (r *WalletRepository) normalizeWalletProvider(ctx context.Context, p string) (string, error) {
	return resolveProviderName(ctx, r.db, p)
}
