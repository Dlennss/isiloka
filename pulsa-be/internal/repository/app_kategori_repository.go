package repository

import (
	"context"
	"database/sql"
)

type AppKategoriRepository struct {
	db *sql.DB
}

func NewAppKategoriRepository(db *sql.DB) *AppKategoriRepository {
	return &AppKategoriRepository{db: db}
}

func (r *AppKategoriRepository) List(ctx context.Context) ([]MasterSimpleRow, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT
  id,
  nama,
  aktif,
  dibuat_pada,
  diubah_pada
FROM public.kategori k
WHERE aktif = true
  AND EXISTS (
    SELECT 1
    FROM public.produk p
    JOIN public.produk_app_pricing app
      ON app.produk_id = p.id
     AND app.aktif = true
     AND LOWER(TRIM(app.provider)) = 'pulsa24jam'
    WHERE p.kategori_id = k.id
      AND p.aktif = true
  )
ORDER BY id ASC
`)
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
