package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"pulsa2/db"
)

type MemberTrxProviderWalletRepository struct {
	db *sql.DB
}

type ProviderSnapshotTxIn struct {
	Provider            string
	SaldoProvider       int64
	RefID               string
	TransaksiMemberID   *int64
	TransaksiProviderID *int64
	Sumber              string
	RawJSON             []byte
}

func NewMemberTrxProviderWalletRepository(db *sql.DB) *MemberTrxProviderWalletRepository {
	return &MemberTrxProviderWalletRepository{db: db}
}

func (r *MemberTrxProviderWalletRepository) Ensure(ctx context.Context, provider string) error {
	p, err := resolveProviderName(ctx, r.db, provider)
	if err != nil {
		return err
	}
	const q = `
INSERT INTO public.dompet_provider (provider, saldo)
VALUES ($1, 0)
ON CONFLICT (provider) DO NOTHING`
	_, err = r.db.ExecContext(ctx, q, p)
	return err
}

func (r *MemberTrxProviderWalletRepository) GetSaldo(ctx context.Context, provider string) (int64, error) {
	p, err := resolveProviderName(ctx, r.db, provider)
	if err != nil {
		return 0, err
	}
	if err := r.Ensure(ctx, p); err != nil {
		return 0, err
	}
	const q = `SELECT saldo FROM public.dompet_provider WHERE provider = $1 LIMIT 1`
	var saldo int64
	if err := r.db.QueryRowContext(ctx, q, p).Scan(&saldo); err != nil {
		return 0, err
	}
	return saldo, nil
}

func (r *MemberTrxProviderWalletRepository) GetLatestSnapshot(ctx context.Context, provider string) (saldo int64, ts time.Time, ok bool, err error) {
	p, err := resolveProviderName(ctx, r.db, provider)
	if err != nil {
		return 0, time.Time{}, false, err
	}
	if err := r.db.QueryRowContext(ctx, `
SELECT saldo_provider, dibuat_pada
FROM public.provider_saldo_snapshot
WHERE provider = $1
ORDER BY transaksi_provider_id DESC NULLS LAST, dibuat_pada DESC, id DESC
LIMIT 1
`, p).Scan(&saldo, &ts); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, time.Time{}, false, nil
		}
		return 0, time.Time{}, false, err
	}
	return saldo, ts, true, nil
}

func (r *MemberTrxProviderWalletRepository) ApplyTx(ctx context.Context, in db.ProviderWalletTxIn) (before int64, after int64, err error) {
	return r.applyTx(ctx, in)
}

func (r *MemberTrxProviderWalletRepository) InsertSnapshot(ctx context.Context, in ProviderSnapshotTxIn) error {
	return r.insertSnapshot(ctx, in)
}
