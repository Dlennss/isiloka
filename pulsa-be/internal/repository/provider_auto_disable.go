package repository

import (
	"context"
	"database/sql"
	"strings"
)

func DisableProviderForOutOfBalance(ctx context.Context, db *sql.DB, providerName, reason string) (int64, error) {
	if db == nil {
		return 0, nil
	}
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	if providerName == "" {
		return 0, nil
	}
	reason = strings.Join(strings.Fields(strings.TrimSpace(reason)), " ")
	if reason == "" {
		reason = "auto nonaktif: saldo provider tidak cukup"
	}
	if len(reason) > 220 {
		reason = reason[:220]
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	var changed int64
	res, err := tx.ExecContext(ctx, `
UPDATE public.provider
SET aktif = false,
    keterangan = $2,
    diubah_pada = now()
WHERE lower(trim(nama)) = $1
  AND aktif = true`, providerName, reason)
	if err != nil {
		return 0, err
	}
	if n, err := res.RowsAffected(); err == nil {
		changed += n
	}

	res, err = tx.ExecContext(ctx, `
UPDATE public.produk_provider_map
SET aktif = false
WHERE lower(trim(provider)) = $1
  AND aktif = true`, providerName)
	if err != nil {
		return 0, err
	}
	if n, err := res.RowsAffected(); err == nil {
		changed += n
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return changed, nil
}
