package repository

import (
	"context"
	"database/sql"
	"strings"
)

type ProdukRepository struct {
	db *sql.DB
}

func NewProdukRepository(db *sql.DB) *ProdukRepository {
	return &ProdukRepository{db: db}
}

func (r *ProdukRepository) List(ctx context.Context, q, sku, groupName string, kategoriID, brandID int64, limit, offset int) ([]ProdukRow, int64, error) {
	q = strings.TrimSpace(q)
	sku = strings.TrimSpace(sku)
	groupName = strings.TrimSpace(groupName)
	if limit <= 0 {
		limit = 10
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, `
SELECT COUNT(1)
FROM public.produk p
WHERE ($1 = '' OR p.sku ILIKE '%'||$1||'%' OR p.nama ILIKE '%'||$1||'%')
  AND ($2 = '' OR UPPER(TRIM(p.sku)) = UPPER(TRIM($2)))
  AND ($3 = '' OR COALESCE(p.group_name, '') ILIKE '%'||$3||'%')
  AND ($4 <= 0 OR p.kategori_id = $4)
  AND ($5 <= 0 OR p.brand_id = $5)
`, q, sku, groupName, kategoriID, brandID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT
  p.id, p.sku, p.nama, COALESCE(p.group_name, ''),
  p.kategori_id, COALESCE(k.nama, ''),
  p.brand_id, COALESCE(b.nama, ''),
  p.tipe_harga::text,
  p.nominal, p.maksimal_nominal,
  p.jam_buka::text, p.jam_tutup::text,
  p.aktif, p.dibuat_pada, p.diubah_pada
FROM public.produk p
LEFT JOIN public.kategori k ON k.id = p.kategori_id
LEFT JOIN public.brand b ON b.id = p.brand_id
WHERE ($1 = '' OR p.sku ILIKE '%'||$1||'%' OR p.nama ILIKE '%'||$1||'%')
  AND ($2 = '' OR UPPER(TRIM(p.sku)) = UPPER(TRIM($2)))
  AND ($3 = '' OR COALESCE(p.group_name, '') ILIKE '%'||$3||'%')
  AND ($4 <= 0 OR p.kategori_id = $4)
  AND ($5 <= 0 OR p.brand_id = $5)
ORDER BY p.id DESC
LIMIT $6 OFFSET $7
`, q, sku, groupName, kategoriID, brandID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]ProdukRow, 0, limit)
	for rows.Next() {
		var (
			row     ProdukRow
			nominal sql.NullInt64
			maxNom  sql.NullInt64
			dibuat  sql.NullTime
			diubah  sql.NullTime
		)
		if err := rows.Scan(
			&row.ID, &row.SKU, &row.Nama, &row.GroupName,
			&row.KategoriID, &row.KategoriNama,
			&row.BrandID, &row.BrandNama,
			&row.TipeHarga,
			&nominal, &maxNom,
			&row.JamBuka, &row.JamTutup,
			&row.Aktif, &dibuat, &diubah,
		); err != nil {
			return nil, 0, err
		}
		if nominal.Valid {
			v := nominal.Int64
			row.Nominal = &v
		}
		if maxNom.Valid {
			v := maxNom.Int64
			row.MaksimalNominal = &v
		}
		if dibuat.Valid {
			v := dibuat.Time
			row.DibuatPada = &v
		}
		if diubah.Valid {
			v := diubah.Time
			row.DiubahPada = &v
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *ProdukRepository) Get(ctx context.Context, id int64) (*ProdukRow, error) {
	var (
		row     ProdukRow
		nominal sql.NullInt64
		maxNom  sql.NullInt64
		dibuat  sql.NullTime
		diubah  sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
SELECT
  p.id, p.sku, p.nama, COALESCE(p.group_name, ''),
  p.kategori_id, COALESCE(k.nama, ''),
  p.brand_id, COALESCE(b.nama, ''),
  p.tipe_harga::text,
  p.nominal, p.maksimal_nominal,
  p.jam_buka::text, p.jam_tutup::text,
  p.aktif, p.dibuat_pada, p.diubah_pada
FROM public.produk p
LEFT JOIN public.kategori k ON k.id = p.kategori_id
LEFT JOIN public.brand b ON b.id = p.brand_id
WHERE p.id = $1
LIMIT 1
`, id).Scan(
		&row.ID, &row.SKU, &row.Nama, &row.GroupName,
		&row.KategoriID, &row.KategoriNama,
		&row.BrandID, &row.BrandNama,
		&row.TipeHarga,
		&nominal, &maxNom,
		&row.JamBuka, &row.JamTutup,
		&row.Aktif, &dibuat, &diubah,
	)
	if err != nil {
		return nil, err
	}
	if nominal.Valid {
		v := nominal.Int64
		row.Nominal = &v
	}
	if maxNom.Valid {
		v := maxNom.Int64
		row.MaksimalNominal = &v
	}
	if dibuat.Valid {
		v := dibuat.Time
		row.DibuatPada = &v
	}
	if diubah.Valid {
		v := diubah.Time
		row.DiubahPada = &v
	}
	return &row, nil
}

func (r *ProdukRepository) ListGroupNames(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT DISTINCT COALESCE(group_name, '') AS group_name
FROM public.produk
WHERE TRIM(COALESCE(group_name, '')) <> ''
ORDER BY group_name ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]string, 0, 64)
	for rows.Next() {
		var groupName string
		if err := rows.Scan(&groupName); err != nil {
			return nil, err
		}
		out = append(out, groupName)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *ProdukRepository) Create(ctx context.Context, in ProdukUpsertInput) (int64, error) {
	in.SKU = strings.ToUpper(strings.TrimSpace(in.SKU))
	in.Nama = strings.TrimSpace(in.Nama)
	in.GroupName = strings.TrimSpace(in.GroupName)
	in.TipeHarga = strings.TrimSpace(in.TipeHarga)

	var id int64
	err := r.db.QueryRowContext(ctx, `
INSERT INTO public.produk
  (sku, nama, group_name, kategori_id, brand_id, tipe_harga, nominal, maksimal_nominal, jam_buka, jam_tutup, aktif, dibuat_pada, diubah_pada)
VALUES
  ($1,$2,$3,$4,$5,$6,$7,$8,COALESCE(NULLIF($9,''),'00:31')::time,COALESCE(NULLIF($10,''),'23:29')::time,$11,now(),now())
RETURNING id
`, in.SKU, in.Nama, in.GroupName, in.KategoriID, in.BrandID, in.TipeHarga, in.Nominal, in.MaksimalNominal, in.JamBuka, in.JamTutup, in.Aktif).Scan(&id)
	return id, err
}

func (r *ProdukRepository) Update(ctx context.Context, in ProdukUpsertInput) error {
	in.SKU = strings.ToUpper(strings.TrimSpace(in.SKU))
	in.Nama = strings.TrimSpace(in.Nama)
	in.GroupName = strings.TrimSpace(in.GroupName)
	in.TipeHarga = strings.TrimSpace(in.TipeHarga)

	res, err := r.db.ExecContext(ctx, `
UPDATE public.produk
SET sku = $2,
    nama = $3,
    group_name = $4,
    kategori_id = $5,
    brand_id = $6,
    tipe_harga = $7,
    nominal = $8,
    maksimal_nominal = $9,
    jam_buka = COALESCE(NULLIF($10,''),'00:31')::time,
    jam_tutup = COALESCE(NULLIF($11,''),'23:29')::time,
    aktif = $12,
    diubah_pada = now()
WHERE id = $1
`, in.ID, in.SKU, in.Nama, in.GroupName, in.KategoriID, in.BrandID, in.TipeHarga, in.Nominal, in.MaksimalNominal, in.JamBuka, in.JamTutup, in.Aktif)
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

func (r *ProdukRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM public.produk WHERE id = $1`, id)
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
