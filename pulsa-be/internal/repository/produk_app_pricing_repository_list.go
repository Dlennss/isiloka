package repository

import (
	"context"
	"database/sql"
	"strings"
)

func (r *ProdukAppPricingRepository) List(ctx context.Context, q string, aktif *bool, limit, offset int) ([]ProdukAppPricingRow, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	q = strings.TrimSpace(q)
	var (
		args   []any
		wheres []string
	)

	wheres = append(wheres, "1=1")
	if q != "" {
		args = append(args, q)
		p := len(args)
		wheres = append(wheres, "(p.sku ILIKE '%'||$"+itoa(p)+"||'%' OR p.nama ILIKE '%'||$"+itoa(p)+"||'%' OR k.nama ILIKE '%'||$"+itoa(p)+"||'%')")
	}
	if aktif != nil {
		args = append(args, *aktif)
		p := len(args)
		wheres = append(wheres, "a.aktif = $"+itoa(p))
	}

	baseFrom := `
FROM public.produk_app_pricing a
JOIN public.produk p ON p.id = a.produk_id
JOIN public.kategori k ON k.id = p.kategori_id
`
	whereSQL := strings.Join(wheres, " AND ")

	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, limit, offset)
	limitPos := len(args) + 1
	offsetPos := len(args) + 2

	qList := `
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
` + baseFrom + `
WHERE ` + whereSQL + `
ORDER BY a.id DESC
LIMIT $` + itoa(limitPos) + ` OFFSET $` + itoa(offsetPos)

	rows, err := r.db.QueryContext(ctx, qList, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]ProdukAppPricingRow, 0, limit)
	for rows.Next() {
		var (
			x   ProdukAppPricingRow
			fAt sql.NullTime
			cAt sql.NullTime
			uAt sql.NullTime
		)
		if err := rows.Scan(
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
		); err != nil {
			return nil, 0, err
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
		out = append(out, x)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	qCount := `SELECT COUNT(1) ` + baseFrom + ` WHERE ` + whereSQL
	var total int64
	if err := r.db.QueryRowContext(ctx, qCount, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	return out, total, nil
}

func (r *ProdukAppPricingRepository) Get(ctx context.Context, id int64) (*ProdukAppPricingRow, error) {
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
WHERE a.id = $1
LIMIT 1
`, id).Scan(
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
