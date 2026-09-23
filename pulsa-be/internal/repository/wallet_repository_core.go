package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

type WalletRepository struct {
	db *sql.DB
}

type WalletActorActivitySummary struct {
	MemberAdjustCount   int64 `json:"member_adjust_count"`
	MemberCreditCount   int64 `json:"member_credit_count"`
	MemberDebitCount    int64 `json:"member_debit_count"`
	ProviderAdjustCount int64 `json:"provider_adjust_count"`
	ProviderCreditCount int64 `json:"provider_credit_count"`
	ProviderDebitCount  int64 `json:"provider_debit_count"`
}

type WalletActorActivityRow struct {
	Scope      string    `json:"scope"`
	Target     string    `json:"target"`
	RefID      string    `json:"ref_id"`
	Arah       string    `json:"arah"`
	Jumlah     int64     `json:"jumlah"`
	Alasan     string    `json:"alasan"`
	Catatan    string    `json:"catatan"`
	DibuatPada time.Time `json:"dibuat_pada"`
}

func NewWalletRepository(sqlDB *sql.DB) *WalletRepository {
	return &WalletRepository{db: sqlDB}
}

func (r *WalletRepository) WalletActorActivitySummary(ctx context.Context, actorID int64, from, to time.Time) (*WalletActorActivitySummary, error) {
	out := &WalletActorActivitySummary{}
	if err := r.db.QueryRowContext(ctx, `
SELECT
  count(*)::bigint,
  count(*) FILTER (WHERE lower(md.arah) = 'credit')::bigint,
  count(*) FILTER (WHERE lower(md.arah) = 'debit')::bigint
FROM public.mutasi_dompet md
WHERE md.diubah_oleh = $1
  AND md.dibuat_pada >= $2
  AND md.dibuat_pada < $3
`, actorID, from, to).Scan(&out.MemberAdjustCount, &out.MemberCreditCount, &out.MemberDebitCount); err != nil {
		return nil, err
	}

	if err := r.db.QueryRowContext(ctx, `
SELECT
  count(*)::bigint,
  count(*) FILTER (WHERE lower(mdp.arah) = 'credit')::bigint,
  count(*) FILTER (WHERE lower(mdp.arah) = 'debit')::bigint
FROM public.mutasi_dompet_provider mdp
WHERE mdp.diubah_oleh = $1
  AND mdp.dibuat_pada >= $2
  AND mdp.dibuat_pada < $3
`, actorID, from, to).Scan(&out.ProviderAdjustCount, &out.ProviderCreditCount, &out.ProviderDebitCount); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *WalletRepository) WalletActorRecentActivity(ctx context.Context, actorID int64, from, to time.Time, limit int) ([]WalletActorActivityRow, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT scope, target, ref_id, arah, jumlah, alasan, catatan, dibuat_pada
FROM (
  SELECT
    'member'::text AS scope,
    coalesce(m.nama, 'member-' || md.member_id::text) AS target,
    md.ref_id,
    md.arah,
    md.jumlah,
    md.alasan,
    coalesce(md.catatan, '') AS catatan,
    md.dibuat_pada
  FROM public.mutasi_dompet md
  LEFT JOIN public.member m ON m.id = md.member_id
  WHERE md.diubah_oleh = $1
    AND md.dibuat_pada >= $2
    AND md.dibuat_pada < $3

  UNION ALL

  SELECT
    'provider'::text AS scope,
    mdp.provider AS target,
    mdp.ref_id,
    mdp.arah,
    mdp.jumlah,
    mdp.alasan,
    coalesce(mdp.catatan, '') AS catatan,
    mdp.dibuat_pada
  FROM public.mutasi_dompet_provider mdp
  WHERE mdp.diubah_oleh = $1
    AND mdp.dibuat_pada >= $2
    AND mdp.dibuat_pada < $3
) x
ORDER BY dibuat_pada DESC
LIMIT $4
`, actorID, from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]WalletActorActivityRow, 0, limit)
	for rows.Next() {
		var item WalletActorActivityRow
		if err := rows.Scan(&item.Scope, &item.Target, &item.RefID, &item.Arah, &item.Jumlah, &item.Alasan, &item.Catatan, &item.DibuatPada); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *WalletRepository) ListWalletProviderNames(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT lower(trim(nama)) AS provider
FROM public.provider
WHERE trim(coalesce(nama, '')) <> ''
  AND COALESCE(aktif, false) = true
ORDER BY provider ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]string, 0, 8)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		name = strings.TrimSpace(strings.ToLower(name))
		if name == "" {
			continue
		}
		out = append(out, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
