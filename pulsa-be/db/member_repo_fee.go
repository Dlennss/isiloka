package db

import (
	"context"
	"database/sql"
	"strings"

	"pulsa2/internal/helper"
)

type ProdukPricingRule struct {
	TipeHarga       string
	Nominal         *int64
	MaksimalNominal *int64
	Aktif           bool
}

func (r *MemberRepo) GetMemberFee(ctx context.Context, memberID int64) (int64, error) {
	var fee int64
	err := r.DB.QueryRowContext(ctx, `SELECT fee_member_rp FROM public.member WHERE id=$1`, memberID).Scan(&fee)
	if err != nil {
		return 0, err
	}
	if fee < 0 {
		fee = 0
	}
	return fee, nil
}

// GetMemberFeeByH2HCategory mengambil fee H2H spesifik member per kategori H2H.
// Jalur H2H aktif sekarang hanya memakai fee_rp flat pada member_h2h_fee.
func (r *MemberRepo) GetMemberFeeByH2HCategory(ctx context.Context, memberID int64, kodeProduk string, qty int64) (int64, string, error) {
	kodeProduk = strings.TrimSpace(strings.ToUpper(kodeProduk))
	if memberID <= 0 || kodeProduk == "" {
		return 0, "", nil
	}
	if qty < 0 {
		qty = 0
	}
	categoryName := helper.ClassifyH2HFeeCategory(kodeProduk)

	var (
		feeRp sql.NullInt64
		err   error
	)

	err = r.DB.QueryRowContext(ctx, `
SELECT fee_rp
FROM public.member_h2h_fee mhf
WHERE mhf.member_id = $1
  AND mhf.aktif = true
  AND mhf.fee_code = $2
LIMIT 1
`, memberID, categoryName).Scan(&feeRp)
	if err == nil {
		if feeRp.Valid {
			if feeRp.Int64 < 0 {
				return 0, "category_flat", nil
			}
			return feeRp.Int64, "category_flat", nil
		}
		return 0, "category_flat", nil
	}
	if err != sql.ErrNoRows {
		return 0, "", err
	}
	return 0, "", nil
}

// GetProdukPricingRuleBySKU mengambil rule harga dari master produk aktif berdasarkan SKU.
// Return nil,nil jika produk tidak ditemukan/aktif.
func (r *MemberRepo) GetProdukPricingRuleBySKU(ctx context.Context, sku string) (*ProdukPricingRule, error) {
	sku = strings.TrimSpace(strings.ToUpper(sku))
	if sku == "" {
		return nil, nil
	}

	var (
		tipeHarga string
		nominal   sql.NullInt64
		maxNom    sql.NullInt64
		aktif     bool
	)

	err := r.DB.QueryRowContext(ctx, `
SELECT tipe_harga::text, nominal, maksimal_nominal, aktif
FROM public.produk
WHERE UPPER(sku) = $1
  AND aktif = true
LIMIT 1
`, sku).Scan(&tipeHarga, &nominal, &maxNom, &aktif)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	out := &ProdukPricingRule{
		TipeHarga: strings.ToUpper(strings.TrimSpace(tipeHarga)),
	}
	if nominal.Valid {
		v := nominal.Int64
		out.Nominal = &v
	}
	if maxNom.Valid {
		v := maxNom.Int64
		out.MaksimalNominal = &v
	}
	return out, nil
}
