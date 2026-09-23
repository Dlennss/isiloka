package repository

import (
	"context"
	"database/sql"
	"strings"
)

const pulsa24JamUnavailableStatus = "PULSA24JAM_OUT_OF_STOCK"

func (r *ProdukAppPricingRepository) Create(ctx context.Context, in ProdukAppPricingUpsertInput) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
INSERT INTO public.produk_app_pricing
  (produk_id, harga, aktif, fetched_at, created_at, updated_at)
VALUES
  ($1, $2, $3, $4, now(), now())
RETURNING id
`, in.ProdukID, in.Harga, in.Aktif, in.FetchedAt).Scan(&id)
	return id, err
}

func (r *ProdukAppPricingRepository) Update(ctx context.Context, in ProdukAppPricingUpsertInput) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE public.produk_app_pricing
SET produk_id = $2,
    harga = $3,
    aktif = $4,
    fetched_at = $5,
    updated_at = now()
WHERE id = $1
`, in.ID, in.ProdukID, in.Harga, in.Aktif, in.FetchedAt)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ProdukAppPricingRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `
DELETE FROM public.produk_app_pricing
WHERE id = $1
`, id)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ProdukAppPricingRepository) ExistsByProdukDefaultProvider(ctx context.Context, produkID int64) (bool, error) {
	var n int64
	err := r.db.QueryRowContext(ctx, `
SELECT COUNT(1)
FROM public.produk_app_pricing
WHERE produk_id = $1
  AND provider = 'yuscom'
`, produkID).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// MarkProviderProductUnavailable removes a rejected product from app and H2H
// listings until it is explicitly verified and enabled again.
func (r *ProdukAppPricingRepository) MarkProviderProductUnavailable(ctx context.Context, produkID int64, provider string) error {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if r == nil || r.db == nil || produkID <= 0 || provider == "" {
		return sql.ErrNoRows
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
UPDATE public.produk_app_pricing
SET aktif = false,
    yuscom_status = $3,
    updated_at = now(),
    diubah_pada = now()
WHERE produk_id = $1
  AND LOWER(TRIM(provider)) = $2
`, produkID, provider, pulsa24JamUnavailableStatus)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE public.produk_provider_map
SET aktif = false, diubah_pada = now()
WHERE produk_id = $1
  AND LOWER(TRIM(provider)) = $2
`, produkID, provider); err != nil {
		return err
	}
	return tx.Commit()
}
