package repository

import (
	"context"
	"database/sql"
)

type AppBrandRepository struct {
	db *sql.DB
}

func NewAppBrandRepository(db *sql.DB) *AppBrandRepository {
	return &AppBrandRepository{db: db}
}

func (r *AppBrandRepository) List(ctx context.Context, kategoriID int64) ([]MasterSimpleRow, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT DISTINCT
  b.id,
  b.nama,
  b.aktif,
  b.dibuat_pada,
  b.diubah_pada
FROM public.brand b
JOIN public.produk p
  ON p.brand_id = b.id
 AND p.aktif = true
JOIN LATERAL (
  SELECT a.id
  FROM public.produk_app_pricing a
  WHERE a.produk_id = p.id
    AND a.aktif = true
    AND LOWER(TRIM(a.provider)) = 'pulsa24jam'
  ORDER BY
    a.harga ASC,
    a.id DESC
  LIMIT 1
) app ON true
JOIN public.kategori_fee_app kfa
  ON kfa.kategori_id = p.kategori_id
 AND kfa.aktif = true
WHERE b.aktif = true
  AND ($1 <= 0 OR p.kategori_id = $1)
ORDER BY b.id ASC
`, kategoriID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]MasterSimpleRow, 0, 64)
	for rows.Next() {
		var row MasterSimpleRow
		if err := rows.Scan(&row.ID, &row.Nama, &row.Aktif, &row.DibuatPada, &row.DiubahPada); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
