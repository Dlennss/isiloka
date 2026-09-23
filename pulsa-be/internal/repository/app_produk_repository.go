package repository

import (
	"context"
	"database/sql"
	"strings"
)

type AppProdukRepository struct {
	db *sql.DB
}

func NewAppProdukRepository(db *sql.DB) *AppProdukRepository {
	return &AppProdukRepository{db: db}
}

func (r *AppProdukRepository) List(ctx context.Context, q string, kategoriID, brandID int64) ([]AppProdukRow, error) {
	q = strings.TrimSpace(q)
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
  CASE
    WHEN p.tipe_harga::text = 'FIXED' AND COALESCE(app.harga, 0) <= 0 THEN 0
    WHEN p.tipe_harga::text = 'OPEN_AMOUNT' THEN COALESCE(kfa_open.fee_master, COALESCE(kfa.fee_master, 0))
    ELSE COALESCE(kfa.fee_master, 0)
  END AS fee_master,
  CASE
    WHEN p.tipe_harga::text = 'FIXED' AND COALESCE(app.harga, 0) <= 0 THEN 0
    WHEN p.tipe_harga::text = 'OPEN_AMOUNT' THEN COALESCE(kfa_open.fee_agent, COALESCE(kfa.fee_agent, 0))
    ELSE COALESCE(kfa.fee_agent, 0)
  END AS fee_agent,
  CASE
    WHEN p.tipe_harga::text = 'FIXED' AND COALESCE(app.harga, 0) <= 0 THEN 0
    WHEN p.tipe_harga::text = 'OPEN_AMOUNT' THEN COALESCE(kfa_open.fee_user, COALESCE(kfa.fee_user, 0))
    ELSE COALESCE(kfa.fee_user, 0)
  END AS fee_user,
  CASE
    WHEN p.tipe_harga::text = 'FIXED' AND COALESCE(app.harga, 0) <= 0 THEN 0
    WHEN p.tipe_harga::text = 'OPEN_AMOUNT' THEN COALESCE(kfa_open.fee_non_user, COALESCE(kfa.fee_non_user, 0))
    ELSE COALESCE(kfa.fee_non_user, 0)
  END AS fee_guest,
  CASE
    WHEN p.tipe_harga::text = 'FIXED' AND COALESCE(app.harga, 0) <= 0 THEN 0
    WHEN p.tipe_harga::text = 'OPEN_AMOUNT' THEN app.harga + COALESCE(kfa_open.fee_master, COALESCE(kfa.fee_master, 0))
    ELSE app.harga + COALESCE(kfa.fee_master, 0)
  END AS harga_master_final,
  CASE
    WHEN p.tipe_harga::text = 'FIXED' AND COALESCE(app.harga, 0) <= 0 THEN 0
    WHEN p.tipe_harga::text = 'OPEN_AMOUNT' THEN app.harga + COALESCE(kfa_open.fee_agent, COALESCE(kfa.fee_agent, 0))
    ELSE app.harga + COALESCE(kfa.fee_agent, 0)
  END AS harga_agent_final,
  CASE
    WHEN p.tipe_harga::text = 'FIXED' AND COALESCE(app.harga, 0) <= 0 THEN 0
    WHEN p.tipe_harga::text = 'OPEN_AMOUNT' THEN app.harga + COALESCE(kfa_open.fee_user, COALESCE(kfa.fee_user, 0))
    ELSE app.harga + COALESCE(kfa.fee_user, 0)
  END AS harga_user_final,
  CASE
    WHEN p.tipe_harga::text = 'FIXED' AND COALESCE(app.harga, 0) <= 0 THEN 0
    WHEN p.tipe_harga::text = 'OPEN_AMOUNT' THEN app.harga + COALESCE(kfa_open.fee_non_user, COALESCE(kfa.fee_non_user, 0))
    ELSE app.harga + COALESCE(kfa.fee_non_user, 0)
  END AS harga_guest_final,
  CASE
    WHEN p.tipe_harga::text = 'FIXED' AND COALESCE(app.harga, 0) <= 0 THEN 0
    WHEN p.tipe_harga::text = 'OPEN_AMOUNT' THEN ((app.harga + COALESCE(kfa_open.fee_master, COALESCE(kfa.fee_master, 0))) * 7 + 999) / 1000
    ELSE ((app.harga + COALESCE(kfa.fee_master, 0)) * 7 + 999) / 1000
  END AS payment_fee_master,
  CASE
    WHEN p.tipe_harga::text = 'FIXED' AND COALESCE(app.harga, 0) <= 0 THEN 0
    WHEN p.tipe_harga::text = 'OPEN_AMOUNT' THEN ((app.harga + COALESCE(kfa_open.fee_agent, COALESCE(kfa.fee_agent, 0))) * 7 + 999) / 1000
    ELSE ((app.harga + COALESCE(kfa.fee_agent, 0)) * 7 + 999) / 1000
  END AS payment_fee_agent,
  CASE
    WHEN p.tipe_harga::text = 'FIXED' AND COALESCE(app.harga, 0) <= 0 THEN 0
    WHEN p.tipe_harga::text = 'OPEN_AMOUNT' THEN ((app.harga + COALESCE(kfa_open.fee_user, COALESCE(kfa.fee_user, 0))) * 7 + 999) / 1000
    ELSE ((app.harga + COALESCE(kfa.fee_user, 0)) * 7 + 999) / 1000
  END AS payment_fee_user,
  CASE
    WHEN p.tipe_harga::text = 'FIXED' AND COALESCE(app.harga, 0) <= 0 THEN 0
    WHEN p.tipe_harga::text = 'OPEN_AMOUNT' THEN ((app.harga + COALESCE(kfa_open.fee_non_user, COALESCE(kfa.fee_non_user, 0))) * 7 + 999) / 1000
    ELSE ((app.harga + COALESCE(kfa.fee_non_user, 0)) * 7 + 999) / 1000
  END AS payment_fee_guest,
  CASE
    WHEN p.tipe_harga::text = 'FIXED' AND COALESCE(app.harga, 0) <= 0 THEN 0
    WHEN p.tipe_harga::text = 'OPEN_AMOUNT' THEN (app.harga + COALESCE(kfa_open.fee_master, COALESCE(kfa.fee_master, 0))) + (((app.harga + COALESCE(kfa_open.fee_master, COALESCE(kfa.fee_master, 0))) * 7 + 999) / 1000)
    ELSE (app.harga + COALESCE(kfa.fee_master, 0)) + (((app.harga + COALESCE(kfa.fee_master, 0)) * 7 + 999) / 1000)
  END AS harga_master_with_payment_fee_final,
  CASE
    WHEN p.tipe_harga::text = 'FIXED' AND COALESCE(app.harga, 0) <= 0 THEN 0
    WHEN p.tipe_harga::text = 'OPEN_AMOUNT' THEN (app.harga + COALESCE(kfa_open.fee_agent, COALESCE(kfa.fee_agent, 0))) + (((app.harga + COALESCE(kfa_open.fee_agent, COALESCE(kfa.fee_agent, 0))) * 7 + 999) / 1000)
    ELSE (app.harga + COALESCE(kfa.fee_agent, 0)) + (((app.harga + COALESCE(kfa.fee_agent, 0)) * 7 + 999) / 1000)
  END AS harga_agent_with_payment_fee_final,
  CASE
    WHEN p.tipe_harga::text = 'FIXED' AND COALESCE(app.harga, 0) <= 0 THEN 0
    WHEN p.tipe_harga::text = 'OPEN_AMOUNT' THEN (app.harga + COALESCE(kfa_open.fee_user, COALESCE(kfa.fee_user, 0))) + (((app.harga + COALESCE(kfa_open.fee_user, COALESCE(kfa.fee_user, 0))) * 7 + 999) / 1000)
    ELSE (app.harga + COALESCE(kfa.fee_user, 0)) + (((app.harga + COALESCE(kfa.fee_user, 0)) * 7 + 999) / 1000)
  END AS harga_user_with_payment_fee_final,
  CASE
    WHEN p.tipe_harga::text = 'FIXED' AND COALESCE(app.harga, 0) <= 0 THEN 0
    WHEN p.tipe_harga::text = 'OPEN_AMOUNT' THEN (app.harga + COALESCE(kfa_open.fee_non_user, COALESCE(kfa.fee_non_user, 0))) + (((app.harga + COALESCE(kfa_open.fee_non_user, COALESCE(kfa.fee_non_user, 0))) * 7 + 999) / 1000)
    ELSE (app.harga + COALESCE(kfa.fee_non_user, 0)) + (((app.harga + COALESCE(kfa.fee_non_user, 0)) * 7 + 999) / 1000)
  END AS harga_guest_with_payment_fee_final,
  p.aktif,
  p.dibuat_pada,
  p.diubah_pada
FROM public.produk p
JOIN LATERAL (
  SELECT a.provider, a.harga
  FROM public.produk_app_pricing a
  WHERE a.produk_id = p.id
    AND a.aktif = true
    AND LOWER(TRIM(a.provider)) = 'pulsa24jam'
  ORDER BY
    a.harga ASC,
    a.id DESC
  LIMIT 1
) app ON true
LEFT JOIN LATERAL (
  SELECT COUNT(*)::bigint AS success_count
  FROM public.app_order ao_rank
  WHERE ao_rank.produk_id = p.id
    AND ao_rank.status = 'success'
    AND ao_rank.dibuat_pada >= NOW() - INTERVAL '90 days'
) sales ON true
LEFT JOIN public.kategori_fee_app kfa
  ON kfa.kategori_id = p.kategori_id
 AND kfa.aktif = true
LEFT JOIN public.kategori k_open
  ON LOWER(k_open.nama) = LOWER('Bebas Nominal')
LEFT JOIN public.kategori_fee_app kfa_open
  ON kfa_open.kategori_id = k_open.id
 AND kfa_open.aktif = true
LEFT JOIN public.kategori k ON k.id = p.kategori_id
LEFT JOIN public.brand b ON b.id = p.brand_id
WHERE p.aktif = true
  AND ($1 = '' OR p.sku ILIKE '%'||$1||'%' OR p.nama ILIKE '%'||$1||'%')
  AND ($2 <= 0 OR p.kategori_id = $2)
  AND ($3 <= 0 OR p.brand_id = $3)
ORDER BY COALESCE(sales.success_count, 0) DESC,
         COALESCE(app.harga, 0) ASC,
         p.id DESC
`, q, kategoriID, brandID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]AppProdukRow, 0, 128)
	for rows.Next() {
		var (
			row                    AppProdukRow
			nominal                sql.NullInt64
			maxNominal             sql.NullInt64
			hargaMaster            sql.NullInt64
			hargaAgent             sql.NullInt64
			hargaUser              sql.NullInt64
			hargaGuest             sql.NullInt64
			hargaMasterWithPayment sql.NullInt64
			hargaAgentWithPayment  sql.NullInt64
			hargaUserWithPayment   sql.NullInt64
			hargaGuestWithPayment  sql.NullInt64
			dibuatPada             sql.NullTime
			diubahPada             sql.NullTime
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
			&row.HargaDasarApp,
			&nominal,
			&maxNominal,
			&row.FeeMaster,
			&row.FeeAgent,
			&row.FeeUser,
			&row.FeeGuest,
			&hargaMaster,
			&hargaAgent,
			&hargaUser,
			&hargaGuest,
			&row.PaymentFeeMaster,
			&row.PaymentFeeAgent,
			&row.PaymentFeeUser,
			&row.PaymentFeeGuest,
			&hargaMasterWithPayment,
			&hargaAgentWithPayment,
			&hargaUserWithPayment,
			&hargaGuestWithPayment,
			&row.Aktif,
			&dibuatPada,
			&diubahPada,
		); err != nil {
			return nil, err
		}
		if nominal.Valid {
			v := nominal.Int64
			row.Nominal = &v
		}
		if maxNominal.Valid {
			v := maxNominal.Int64
			row.MaksimalNominal = &v
		}
		if hargaGuest.Valid {
			v := hargaGuest.Int64
			row.HargaGuestFinal = &v
		}
		if hargaMaster.Valid {
			v := hargaMaster.Int64
			row.HargaMasterFinal = &v
		}
		if hargaAgent.Valid {
			v := hargaAgent.Int64
			row.HargaAgentFinal = &v
		}
		if hargaUser.Valid {
			v := hargaUser.Int64
			row.HargaUserFinal = &v
		}
		if hargaMasterWithPayment.Valid {
			v := hargaMasterWithPayment.Int64
			row.HargaMasterWithPaymentFeeFinal = &v
		}
		if hargaAgentWithPayment.Valid {
			v := hargaAgentWithPayment.Int64
			row.HargaAgentWithPaymentFeeFinal = &v
		}
		if hargaUserWithPayment.Valid {
			v := hargaUserWithPayment.Int64
			row.HargaUserWithPaymentFeeFinal = &v
		}
		if hargaGuestWithPayment.Valid {
			v := hargaGuestWithPayment.Int64
			row.HargaGuestWithPaymentFeeFinal = &v
		}
		if dibuatPada.Valid {
			v := dibuatPada.Time
			row.DibuatPada = &v
		}
		if diubahPada.Valid {
			v := diubahPada.Time
			row.DiubahPada = &v
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
