package repository

import (
	"context"
	"database/sql"
	"strings"
)

type H2HProdukRepository struct {
	db *sql.DB
}

func NewH2HProdukRepository(db *sql.DB) *H2HProdukRepository {
	return &H2HProdukRepository{db: db}
}

func (r *H2HProdukRepository) ListByMember(ctx context.Context, memberID int64, q, kategoriName, brandName string) ([]H2HProdukRow, error) {
	q = strings.TrimSpace(q)
	kategoriName = strings.TrimSpace(kategoriName)
	brandName = strings.TrimSpace(brandName)
	rows, err := r.db.QueryContext(ctx, `
SELECT
  p.id,
  p.sku,
  p.nama,
  COALESCE(p.group_name, ''),
  p.kategori_id,
  COALESCE(k.nama, ''),
  p.brand_id,
  COALESCE(b.nama, ''),
  p.tipe_harga::text,
  COALESCE(app.harga, 0),
  p.nominal,
  p.maksimal_nominal,
  COALESCE(mhf.fee_code, ''),
  COALESCE(mhf.fee_rp, 0),
  p.aktif,
  p.dibuat_pada,
  p.diubah_pada
FROM public.produk p
JOIN public.produk_app_pricing app
  ON app.produk_id = p.id
 AND LOWER(TRIM(app.provider)) = 'pulsa24jam'
 AND app.aktif = true
LEFT JOIN public.kategori k ON k.id = p.kategori_id
LEFT JOIN public.brand b ON b.id = p.brand_id
LEFT JOIN public.member_h2h_fee mhf
  ON mhf.member_id = $1
 AND mhf.aktif = true
 AND mhf.fee_code = CASE
   WHEN UPPER(TRIM(p.sku)) = 'DANA' THEN 'DANA'
   WHEN UPPER(TRIM(p.sku)) IN ('GOPAY', 'GOJEK', 'GPAY') THEN 'GOPAY'
   WHEN UPPER(TRIM(p.sku)) = 'OVO' THEN 'OVO'
   WHEN UPPER(TRIM(p.sku)) IN ('LINKAJA', 'LAJA') THEN 'LINKAJA'
   WHEN UPPER(TRIM(p.sku)) IN ('SHOPEE', 'SHOPEEPAY', 'SHPAY') THEN 'SHOPEEPAY'
   ELSE 'LAINNYA'
 END
WHERE p.aktif = true
  AND COALESCE(k.aktif, false) = true
  AND COALESCE(b.aktif, false) = true
  AND ($2 = '' OR p.sku ILIKE '%'||$2||'%' OR p.nama ILIKE '%'||$2||'%')
  AND ($3 = '' OR k.nama ILIKE $3)
  AND ($4 = '' OR b.nama ILIKE $4)
ORDER BY COALESCE(app.harga, 0) ASC, p.id DESC
`, memberID, q, kategoriName, brandName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]H2HProdukRow, 0, 128)
	for rows.Next() {
		var (
			row         H2HProdukRow
			rowHrgDasar int64
			nominal     sql.NullInt64
			maxNominal  sql.NullInt64
			feeMember   sql.NullInt64
			dibuatPada  sql.NullTime
			diubahPada  sql.NullTime
		)
		if err := rows.Scan(
			&row.ID,
			&row.SKU,
			&row.Nama,
			&row.GroupName,
			&row.KategoriID,
			&row.KategoriNama,
			&row.BrandID,
			&row.BrandNama,
			&row.TipeHarga,
			&rowHrgDasar,
			&nominal,
			&maxNominal,
			new(sql.NullString),
			&feeMember,
			&row.Aktif,
			&dibuatPada,
			&diubahPada,
		); err != nil {
			return nil, err
		}

		if maxNominal.Valid {
			v := maxNominal.Int64
			row.MaksimalNominal = &v
		}
		if dibuatPada.Valid {
			v := dibuatPada.Time
			row.DibuatPada = &v
		}
		if diubahPada.Valid {
			v := diubahPada.Time
			row.DiubahPada = &v
		}

		feeMemberValue := int64(0)
		if feeMember.Valid && feeMember.Int64 > 0 {
			feeMemberValue = feeMember.Int64
		}
		if strings.EqualFold(strings.TrimSpace(row.TipeHarga), "FIXED") {
			v := rowHrgDasar + feeMemberValue
			row.Harga = &v
		} else {
			row.FeeTambahan = &feeMemberValue
		}
		if !strings.EqualFold(strings.TrimSpace(row.TipeHarga), "FIXED") && nominal.Valid {
			v := nominal.Int64
			row.Nominal = &v
		}

		out = append(out, row)
	}
	return out, rows.Err()
}
