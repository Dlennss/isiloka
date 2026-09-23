package repository

import (
	"context"
	"database/sql"
	"strings"
)

type BrandRepository struct {
	db *sql.DB
}

func NewBrandRepository(db *sql.DB) *BrandRepository {
	return &BrandRepository{db: db}
}

func (r *BrandRepository) List(ctx context.Context) ([]MasterSimpleRow, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, nama, aktif, dibuat_pada, diubah_pada
FROM public.brand
ORDER BY id ASC`)
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

func (r *BrandRepository) Get(ctx context.Context, id int64) (*MasterSimpleRow, error) {
	var row MasterSimpleRow
	err := r.db.QueryRowContext(ctx, `
SELECT id, nama, aktif, dibuat_pada, diubah_pada
FROM public.brand
WHERE id = $1
LIMIT 1
`, id).Scan(&row.ID, &row.Nama, &row.Aktif, &row.DibuatPada, &row.DiubahPada)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *BrandRepository) Create(ctx context.Context, nama string, aktif bool) (int64, error) {
	nama = strings.TrimSpace(nama)
	var id int64
	err := r.db.QueryRowContext(ctx, `
INSERT INTO public.brand (nama, aktif, dibuat_pada, diubah_pada)
VALUES ($1, $2, now(), now())
RETURNING id
`, nama, aktif).Scan(&id)
	return id, err
}

func (r *BrandRepository) Update(ctx context.Context, id int64, nama string, aktif bool) error {
	nama = strings.TrimSpace(nama)
	res, err := r.db.ExecContext(ctx, `
UPDATE public.brand
SET nama = $2, aktif = $3, diubah_pada = now()
WHERE id = $1
`, id, nama, aktif)
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

func (r *BrandRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM public.brand WHERE id = $1`, id)
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
