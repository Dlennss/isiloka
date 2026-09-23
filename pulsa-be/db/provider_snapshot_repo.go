package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// provider_saldo_snapshot menyimpan saldo dari provider (actual) yang dikirim via callback/response.
type ProviderSnapshotRepo struct {
	DB *sql.DB
}

func NewProviderSnapshotRepo(db *sql.DB) *ProviderSnapshotRepo { return &ProviderSnapshotRepo{DB: db} }

func normProviderSnapshot(p string) (string, error) {
	p = strings.TrimSpace(strings.ToLower(p))
	switch p {
	case "yuscom", "javapay":
		return p, nil
	default:
		return "", fmt.Errorf("unknown provider: %s", p)
	}
}

type ProviderSnapshotIn struct {
	Provider            string
	SaldoProvider       int64
	RefID               string
	TransaksiMemberID   *int64
	TransaksiProviderID *int64
	Sumber              string
	RawJSON             []byte
}

// Insert: UPSERT idempotent berdasar (provider, transaksi_provider_id) bila transaksi_provider_id tidak NULL.
// Wajib ada index unik partial:
//
// CREATE UNIQUE INDEX uq_provider_snapshot_once_per_provider_trx
// ON public.provider_saldo_snapshot (provider, transaksi_provider_id)
// WHERE transaksi_provider_id IS NOT NULL;
func (r *ProviderSnapshotRepo) Insert(ctx context.Context, in ProviderSnapshotIn) error {
	p, err := normProviderSnapshot(in.Provider)
	if err != nil {
		return err
	}
	if strings.TrimSpace(in.Sumber) == "" {
		return fmt.Errorf("sumber required")
	}
	if in.SaldoProvider <= 0 {
		return fmt.Errorf("saldo_provider must be > 0")
	}

	var tm any = nil
	var tp any = nil
	if in.TransaksiMemberID != nil {
		tm = *in.TransaksiMemberID
	}
	if in.TransaksiProviderID != nil {
		tp = *in.TransaksiProviderID
	}

	var raw any = nil
	if len(in.RawJSON) > 0 {
		raw = string(in.RawJSON) // cast ke jsonb via query
	}

	now := time.Now().UTC()

	// Penting:
	// - ON CONFLICT harus punya predicate yang sama dengan partial unique index:
	//   WHERE transaksi_provider_id IS NOT NULL
	//
	// - Jika transaksi_provider_id NULL: tidak kena index, dan ON CONFLICT tidak berlaku.
	//   Insert tetap jalan (boleh, tapi bisa dobel). Dalam sistem kita biasanya selalu ada id, jadi aman.
	const q = `
INSERT INTO public.provider_saldo_snapshot
  (provider, saldo_provider, ref_id, transaksi_provider_id, transaksi_member_id, sumber, dibuat_pada, raw)
VALUES
  ($1,$2,NULLIF($3,''),$4,$5,$6,$7,COALESCE($8,'{}'::jsonb))
ON CONFLICT (provider, transaksi_provider_id)
  WHERE transaksi_provider_id IS NOT NULL
DO UPDATE
SET
  saldo_provider = EXCLUDED.saldo_provider,
  ref_id = COALESCE(EXCLUDED.ref_id, public.provider_saldo_snapshot.ref_id),
  transaksi_member_id = COALESCE(EXCLUDED.transaksi_member_id, public.provider_saldo_snapshot.transaksi_member_id),
  sumber      = EXCLUDED.sumber,
  dibuat_pada = EXCLUDED.dibuat_pada,
  raw         = EXCLUDED.raw
`

	_, err = r.DB.ExecContext(
		ctx,
		q,
		p,
		in.SaldoProvider,
		strings.TrimSpace(in.RefID),
		tp,
		tm,
		strings.TrimSpace(in.Sumber),
		now,
		raw,
	)
	return err
}

func (r *ProviderSnapshotRepo) LatestByProvider(ctx context.Context, provider string) (saldo int64, ts time.Time, ok bool, err error) {
	p, err := normProviderSnapshot(provider)
	if err != nil {
		return 0, time.Time{}, false, err
	}
	const q = `
SELECT saldo_provider, dibuat_pada
FROM public.provider_saldo_snapshot
WHERE provider=$1
ORDER BY transaksi_provider_id DESC NULLS LAST, dibuat_pada DESC, id DESC
LIMIT 1`
	var t time.Time
	if err := r.DB.QueryRowContext(ctx, q, p).Scan(&saldo, &t); err != nil {
		if err == sql.ErrNoRows {
			return 0, time.Time{}, false, nil
		}
		return 0, time.Time{}, false, err
	}
	return saldo, t, true, nil
}
