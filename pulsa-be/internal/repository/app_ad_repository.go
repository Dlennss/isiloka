package repository

import (
	"context"
	"database/sql"
	"strings"
)

type AppAdRepository struct {
	db *sql.DB
}

func NewAppAdRepository(db *sql.DB) *AppAdRepository {
	return &AppAdRepository{db: db}
}

func (r *AppAdRepository) List(ctx context.Context, q string) ([]AppAdRow, error) {
	q = strings.TrimSpace(q)
	rows, err := r.db.QueryContext(ctx, `
SELECT
  id,
  judul,
  keterangan,
  image_url,
  COALESCE(link_url, ''),
  urutan,
  aktif,
  created_at,
  updated_at
FROM public.app_ads
WHERE ($1 = '' OR judul ILIKE '%'||$1||'%' OR keterangan ILIKE '%'||$1||'%')
ORDER BY urutan ASC, id DESC
`, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]AppAdRow, 0, 16)
	for rows.Next() {
		var (
			row AppAdRow
			cAt sql.NullTime
			uAt sql.NullTime
		)
		if err := rows.Scan(
			&row.ID,
			&row.Judul,
			&row.Keterangan,
			&row.ImageURL,
			&row.LinkURL,
			&row.Urutan,
			&row.Aktif,
			&cAt,
			&uAt,
		); err != nil {
			return nil, err
		}
		if cAt.Valid {
			v := cAt.Time
			row.CreatedAt = &v
		}
		if uAt.Valid {
			v := uAt.Time
			row.UpdatedAt = &v
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *AppAdRepository) ListActive(ctx context.Context) ([]AppAdRow, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT
  id,
  judul,
  keterangan,
  image_url,
  COALESCE(link_url, ''),
  urutan,
  aktif,
  created_at,
  updated_at
FROM public.app_ads
WHERE aktif = true
ORDER BY urutan ASC, id DESC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]AppAdRow, 0, 8)
	for rows.Next() {
		var (
			row AppAdRow
			cAt sql.NullTime
			uAt sql.NullTime
		)
		if err := rows.Scan(
			&row.ID,
			&row.Judul,
			&row.Keterangan,
			&row.ImageURL,
			&row.LinkURL,
			&row.Urutan,
			&row.Aktif,
			&cAt,
			&uAt,
		); err != nil {
			return nil, err
		}
		if cAt.Valid {
			v := cAt.Time
			row.CreatedAt = &v
		}
		if uAt.Valid {
			v := uAt.Time
			row.UpdatedAt = &v
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *AppAdRepository) Get(ctx context.Context, id int64) (*AppAdRow, error) {
	var (
		row AppAdRow
		cAt sql.NullTime
		uAt sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
SELECT
  id,
  judul,
  keterangan,
  image_url,
  COALESCE(link_url, ''),
  urutan,
  aktif,
  created_at,
  updated_at
FROM public.app_ads
WHERE id = $1
LIMIT 1
`, id).Scan(
		&row.ID,
		&row.Judul,
		&row.Keterangan,
		&row.ImageURL,
		&row.LinkURL,
		&row.Urutan,
		&row.Aktif,
		&cAt,
		&uAt,
	)
	if err != nil {
		return nil, err
	}
	if cAt.Valid {
		v := cAt.Time
		row.CreatedAt = &v
	}
	if uAt.Valid {
		v := uAt.Time
		row.UpdatedAt = &v
	}
	return &row, nil
}

func (r *AppAdRepository) Create(ctx context.Context, in AppAdUpsertInput) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
INSERT INTO public.app_ads
  (judul, keterangan, image_url, link_url, urutan, aktif, created_at, updated_at)
VALUES
  ($1, $2, $3, NULLIF($4, ''), $5, $6, now(), now())
RETURNING id
`, in.Judul, in.Keterangan, in.ImageURL, in.LinkURL, in.Urutan, in.Aktif).Scan(&id)
	return id, err
}

func (r *AppAdRepository) Update(ctx context.Context, in AppAdUpsertInput) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE public.app_ads
SET judul = $2,
    keterangan = $3,
    image_url = $4,
    link_url = NULLIF($5, ''),
    urutan = $6,
    aktif = $7,
    updated_at = now()
WHERE id = $1
`, in.ID, in.Judul, in.Keterangan, in.ImageURL, in.LinkURL, in.Urutan, in.Aktif)
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

func (r *AppAdRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM public.app_ads WHERE id = $1`, id)
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
