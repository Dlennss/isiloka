package repository

import (
	"context"
	"database/sql"
	"strings"
)

func (r *ProdukAppPricingRepository) GetByProdukIDProviderActive(ctx context.Context, produkID int64, provider string) (*ProdukAppPricingRow, error) {
	provider = strings.TrimSpace(strings.ToLower(provider))
	if provider == "" {
		return nil, sql.ErrNoRows
	}

	var (
		x   ProdukAppPricingRow
		fAt sql.NullTime
		cAt sql.NullTime
		uAt sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
SELECT
  a.id,
  a.produk_id,
  p.sku,
  p.nama,
  p.kategori_id,
  k.nama AS kategori_nama,
  a.provider,
  a.harga,
  a.aktif,
  a.yuscom_group,
  a.yuscom_category,
  a.yuscom_subcategory,
  a.yuscom_sku,
  a.yuscom_name,
  a.yuscom_status,
  a.yuscom_display_brand,
  a.fetched_at,
  a.created_at,
  a.updated_at
FROM public.produk_app_pricing a
JOIN public.produk p ON p.id = a.produk_id
JOIN public.kategori k ON k.id = p.kategori_id
WHERE a.produk_id = $1
  AND LOWER(TRIM(a.provider)) = $2
  AND a.aktif = true
LIMIT 1
`, produkID, provider).Scan(
		&x.ID,
		&x.ProdukID,
		&x.ProdukSKU,
		&x.ProdukNama,
		&x.KategoriID,
		&x.KategoriNama,
		&x.Provider,
		&x.Harga,
		&x.Aktif,
		&x.YuscomGroup,
		&x.YuscomCategory,
		&x.YuscomSubcategory,
		&x.YuscomSKU,
		&x.YuscomName,
		&x.YuscomStatus,
		&x.YuscomDisplayBrand,
		&fAt,
		&cAt,
		&uAt,
	)
	if err != nil {
		return nil, err
	}
	if fAt.Valid {
		v := fAt.Time
		x.FetchedAt = &v
	}
	if cAt.Valid {
		v := cAt.Time
		x.CreatedAt = &v
	}
	if uAt.Valid {
		v := uAt.Time
		x.UpdatedAt = &v
	}
	return &x, nil
}

func (r *ProdukAppPricingRepository) GetEffectiveByProdukIDActive(ctx context.Context, produkID int64) (*ProdukAppPricingRow, error) {
	var (
		x   ProdukAppPricingRow
		fAt sql.NullTime
		cAt sql.NullTime
		uAt sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
SELECT
  a.id,
  a.produk_id,
  p.sku,
  p.nama,
  p.kategori_id,
  k.nama AS kategori_nama,
  a.provider,
  a.harga,
  a.aktif,
  a.yuscom_group,
  a.yuscom_category,
  a.yuscom_subcategory,
  a.yuscom_sku,
  a.yuscom_name,
  a.yuscom_status,
  a.yuscom_display_brand,
  a.fetched_at,
  a.created_at,
  a.updated_at
FROM public.produk_app_pricing a
JOIN public.produk p ON p.id = a.produk_id
JOIN public.kategori k ON k.id = p.kategori_id
WHERE a.produk_id = $1
  AND a.aktif = true
  AND LOWER(TRIM(a.provider)) = 'pulsa24jam'
ORDER BY
  a.harga ASC,
  a.id DESC
LIMIT 1
`, produkID).Scan(
		&x.ID,
		&x.ProdukID,
		&x.ProdukSKU,
		&x.ProdukNama,
		&x.KategoriID,
		&x.KategoriNama,
		&x.Provider,
		&x.Harga,
		&x.Aktif,
		&x.YuscomGroup,
		&x.YuscomCategory,
		&x.YuscomSubcategory,
		&x.YuscomSKU,
		&x.YuscomName,
		&x.YuscomStatus,
		&x.YuscomDisplayBrand,
		&fAt,
		&cAt,
		&uAt,
	)
	if err != nil {
		return nil, err
	}
	if fAt.Valid {
		v := fAt.Time
		x.FetchedAt = &v
	}
	if cAt.Valid {
		v := cAt.Time
		x.CreatedAt = &v
	}
	if uAt.Valid {
		v := uAt.Time
		x.UpdatedAt = &v
	}
	return &x, nil
}

func (r *ProdukAppPricingRepository) GetByProdukIDActive(ctx context.Context, produkID int64) (*ProdukAppPricingRow, error) {
	return r.GetByProdukIDProviderActive(ctx, produkID, "pulsa24jam")
}
