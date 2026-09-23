package repository

import (
	"context"
	"fmt"
	"strings"
)

type ProviderSnapshotIn struct {
	Provider            string
	SaldoProvider       int64
	RefID               string
	TransaksiMemberID   *int64
	TransaksiProviderID *int64
	Sumber              string
	RawJSON             []byte
}

func (r *ProviderCallbackRepository) InsertProviderSnapshot(ctx context.Context, in ProviderSnapshotIn) error {
	p, err := r.normalizeProviderForCallback(ctx, in.Provider)
	if err != nil {
		return err
	}
	if strings.TrimSpace(in.Sumber) == "" {
		return fmt.Errorf("sumber required")
	}
	if in.SaldoProvider <= 0 {
		return fmt.Errorf("saldo_provider must be > 0")
	}

	var tm any
	var tp any
	if in.TransaksiMemberID != nil {
		tm = *in.TransaksiMemberID
	}
	if in.TransaksiProviderID != nil {
		tp = *in.TransaksiProviderID
	}
	var raw any
	if len(in.RawJSON) > 0 {
		raw = string(in.RawJSON)
	}

	_, err = r.db.ExecContext(ctx, `
INSERT INTO public.provider_saldo_snapshot
  (provider, saldo_provider, ref_id, transaksi_provider_id, transaksi_member_id, sumber, dibuat_pada, raw)
VALUES
  ($1,$2,NULLIF($3,''),$4,$5,$6,$7,COALESCE($8,'{}'::jsonb))
ON CONFLICT (provider, transaksi_provider_id) WHERE transaksi_provider_id IS NOT NULL
DO UPDATE SET
  saldo_provider = EXCLUDED.saldo_provider,
  ref_id = COALESCE(EXCLUDED.ref_id, public.provider_saldo_snapshot.ref_id),
  transaksi_member_id = COALESCE(EXCLUDED.transaksi_member_id, public.provider_saldo_snapshot.transaksi_member_id),
  sumber = EXCLUDED.sumber,
  dibuat_pada = EXCLUDED.dibuat_pada,
  raw = EXCLUDED.raw
	`, p, in.SaldoProvider, strings.TrimSpace(in.RefID), tp, tm, strings.TrimSpace(in.Sumber), nowWIB(), raw)
	return err
}

func (r *ProviderCallbackRepository) GetProviderSaldo(ctx context.Context, provider string) (int64, error) {
	p, err := r.normalizeProviderForCallback(ctx, provider)
	if err != nil {
		return 0, err
	}
	var saldo int64
	err = r.db.QueryRowContext(ctx, `
SELECT COALESCE(saldo, 0)
FROM public.dompet_provider
WHERE provider = $1
`, p).Scan(&saldo)
	return saldo, err
}
