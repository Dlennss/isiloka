package repository

import (
	"context"
	"database/sql"
	"strings"
)

const OpenAmountFeeKategoriName = "Bebas Nominal"

type KategoriFeeAppRepository struct {
	db *sql.DB
}

func NewKategoriFeeAppRepository(db *sql.DB) *KategoriFeeAppRepository {
	return &KategoriFeeAppRepository{db: db}
}

func (r *KategoriFeeAppRepository) List(ctx context.Context, q string) ([]KategoriFeeAppRow, error) {
	q = strings.TrimSpace(q)
	rows, err := r.db.QueryContext(ctx, `
SELECT
  a.id,
  a.kategori_id,
  k.nama AS kategori_nama,
  a.fee_master,
  a.fee_agent,
  a.fee_user,
  a.fee_non_user,
  a.aktif,
  a.created_at,
  a.updated_at
FROM public.kategori_fee_app a
JOIN public.kategori k ON k.id = a.kategori_id
WHERE ($1 = '' OR k.nama ILIKE '%'||$1||'%')
ORDER BY a.id DESC
`, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]KategoriFeeAppRow, 0, 64)
	for rows.Next() {
		var (
			row KategoriFeeAppRow
			cAt sql.NullTime
			uAt sql.NullTime
		)
		if err := rows.Scan(
			&row.ID,
			&row.KategoriID,
			&row.KategoriNama,
			&row.FeeMaster,
			&row.FeeAgent,
			&row.FeeUser,
			&row.FeeNonUser,
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

func (r *KategoriFeeAppRepository) Get(ctx context.Context, id int64) (*KategoriFeeAppRow, error) {
	var (
		row KategoriFeeAppRow
		cAt sql.NullTime
		uAt sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
SELECT
  a.id,
  a.kategori_id,
  k.nama AS kategori_nama,
  a.fee_master,
  a.fee_agent,
  a.fee_user,
  a.fee_non_user,
  a.aktif,
  a.created_at,
  a.updated_at
FROM public.kategori_fee_app a
JOIN public.kategori k ON k.id = a.kategori_id
WHERE a.id = $1
LIMIT 1
`, id).Scan(
		&row.ID,
		&row.KategoriID,
		&row.KategoriNama,
		&row.FeeMaster,
		&row.FeeAgent,
		&row.FeeUser,
		&row.FeeNonUser,
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

func (r *KategoriFeeAppRepository) Create(ctx context.Context, in KategoriFeeAppUpsertInput) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
INSERT INTO public.kategori_fee_app
  (kategori_id, fee_master, fee_agent, fee_user, fee_non_user, aktif, created_at, updated_at)
VALUES
  ($1, $2, $3, $4, $5, $6, now(), now())
RETURNING id
`, in.KategoriID, in.FeeMaster, in.FeeAgent, in.FeeUser, in.FeeNonUser, in.Aktif).Scan(&id)
	return id, err
}

func (r *KategoriFeeAppRepository) Update(ctx context.Context, in KategoriFeeAppUpsertInput) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE public.kategori_fee_app
SET kategori_id = $2,
    fee_master = $3,
    fee_agent = $4,
    fee_user = $5,
    fee_non_user = $6,
    aktif = $7,
    updated_at = now()
WHERE id = $1
`, in.ID, in.KategoriID, in.FeeMaster, in.FeeAgent, in.FeeUser, in.FeeNonUser, in.Aktif)
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

func (r *KategoriFeeAppRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM public.kategori_fee_app WHERE id = $1`, id)
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

func (r *KategoriFeeAppRepository) GetByKategoriIDActive(ctx context.Context, kategoriID int64) (*KategoriFeeAppRow, error) {
	var (
		row KategoriFeeAppRow
		cAt sql.NullTime
		uAt sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
SELECT
  a.id,
  a.kategori_id,
  k.nama AS kategori_nama,
  a.fee_master,
  a.fee_agent,
  a.fee_user,
  a.fee_non_user,
  a.aktif,
  a.created_at,
  a.updated_at
FROM public.kategori_fee_app a
JOIN public.kategori k ON k.id = a.kategori_id
WHERE a.kategori_id = $1
  AND a.aktif = true
LIMIT 1
`, kategoriID).Scan(
		&row.ID,
		&row.KategoriID,
		&row.KategoriNama,
		&row.FeeMaster,
		&row.FeeAgent,
		&row.FeeUser,
		&row.FeeNonUser,
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

func (r *KategoriFeeAppRepository) GetByKategoriNameActive(ctx context.Context, kategoriName string) (*KategoriFeeAppRow, error) {
	var (
		row KategoriFeeAppRow
		cAt sql.NullTime
		uAt sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
SELECT
  a.id,
  a.kategori_id,
  k.nama AS kategori_nama,
  a.fee_master,
  a.fee_agent,
  a.fee_user,
  a.fee_non_user,
  a.aktif,
  a.created_at,
  a.updated_at
FROM public.kategori_fee_app a
JOIN public.kategori k ON k.id = a.kategori_id
WHERE LOWER(k.nama) = LOWER($1)
  AND a.aktif = true
LIMIT 1
`, strings.TrimSpace(kategoriName)).Scan(
		&row.ID,
		&row.KategoriID,
		&row.KategoriNama,
		&row.FeeMaster,
		&row.FeeAgent,
		&row.FeeUser,
		&row.FeeNonUser,
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
